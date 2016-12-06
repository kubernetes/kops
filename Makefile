# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

all: kops

.PHONY: channels examples

DOCKER_REGISTRY?=gcr.io/must-override
S3_BUCKET?=s3://must-override/
GCS_LOCATION?=gs://must-override
GCS_URL=$(GCS_LOCATION:gs://%=https://storage.googleapis.com/%)
LATEST_FILE?=latest-ci.txt
GOPATH_1ST=$(shell echo ${GOPATH} | cut -d : -f 1)
UNIQUE:=$(shell date +%s)
GOVERSION=1.7.1

# See http://stackoverflow.com/questions/18136918/how-to-get-current-relative-directory-of-your-makefile
MAKEDIR:=$(strip $(shell dirname "$(realpath $(lastword $(MAKEFILE_LIST)))"))

# Keep in sync with upup/models/cloudup/resources/addons/dns-controller/
DNS_CONTROLLER_TAG=1.4.1
PROTOKUBE_TAG=1.4.1

ifndef VERSION
  VERSION := git-$(shell git describe --always)
endif

# Go exports:

GO15VENDOREXPERIMENT=1
export GO15VENDOREXPERIMENT

ifdef STATIC_BUILD
  CGO_ENABLED=0
  export CGO_ENABLED
  EXTRA_BUILDFLAGS=-installsuffix cgo
  EXTRA_LDFLAGS=-s
endif

kops: kops-gobindata
	go install ${EXTRA_BUILDFLAGS} -ldflags "-X main.BuildVersion=${VERSION} ${EXTRA_LDFLAGS}" k8s.io/kops/cmd/kops/...

gobindata-tool:
	go build ${EXTRA_BUILDFLAGS} -ldflags "${EXTRA_LDFLAGS}" -o ${GOPATH_1ST}/bin/go-bindata k8s.io/kops/vendor/github.com/jteeuwen/go-bindata/go-bindata

kops-gobindata: gobindata-tool
	cd ${GOPATH_1ST}/src/k8s.io/kops; ${GOPATH_1ST}/bin/go-bindata -o upup/models/bindata.go -pkg models -ignore="\\.DS_Store" -ignore="bindata\\.go" -ignore="vfs\\.go" -prefix upup/models/ upup/models/...
	cd ${GOPATH_1ST}/src/k8s.io/kops; ${GOPATH_1ST}/bin/go-bindata -o federation/model/bindata.go -pkg model -ignore="\\.DS_Store" -ignore="bindata\\.go" -prefix federation/model/ federation/model/...

# Build in a docker container with golang 1.X
# Used to test we have not broken 1.X
check-builds-in-go15:
	docker run -v ${GOPATH_1ST}/src/k8s.io/kops:/go/src/k8s.io/kops golang:1.5 make -f /go/src/k8s.io/kops/Makefile kops

check-builds-in-go16:
	docker run -v ${GOPATH_1ST}/src/k8s.io/kops:/go/src/k8s.io/kops golang:1.6 make -f /go/src/k8s.io/kops/Makefile kops

check-builds-in-go17:
	docker run -v ${GOPATH_1ST}/src/k8s.io/kops:/go/src/k8s.io/kops golang:1.7 make -f /go/src/k8s.io/kops/Makefile kops

codegen: kops-gobindata
	go install k8s.io/kops/upup/tools/generators/...
	PATH=${GOPATH_1ST}/bin:${PATH} go generate k8s.io/kops/upup/pkg/fi/cloudup/awstasks
	PATH=${GOPATH_1ST}/bin:${PATH} go generate k8s.io/kops/upup/pkg/fi/cloudup/gcetasks
	PATH=${GOPATH_1ST}/bin:${PATH} go generate k8s.io/kops/upup/pkg/fi/fitasks

test:
	go test k8s.io/kops/upup/pkg/... -args -v=1 -logtostderr
	go test k8s.io/kops/pkg/... -args -v=1 -logtostderr
	go test k8s.io/kops/dns-controller/pkg/... -args -v=1 -logtostderr

crossbuild:
	mkdir -p .build/dist/
	GOOS=darwin GOARCH=amd64 go build -a ${EXTRA_BUILDFLAGS} -o .build/dist/darwin/amd64/kops -ldflags "${EXTRA_LDFLAGS} -X main.BuildVersion=${VERSION}" k8s.io/kops/cmd/kops
	GOOS=linux GOARCH=amd64 go build -a ${EXTRA_BUILDFLAGS} -o .build/dist/linux/amd64/kops -ldflags "${EXTRA_LDFLAGS} -X main.BuildVersion=${VERSION}" k8s.io/kops/cmd/kops
	#GOOS=windows GOARCH=amd64 go build -o .build/dist/windows/amd64/kops -ldflags "-X main.BuildVersion=${VERSION}" -v k8s.io/kops/cmd/kops/...

crossbuild-in-docker:
	docker pull golang:${GOVERSION} # Keep golang image up to date
	docker run --name=kops-build-${UNIQUE} -e STATIC_BUILD=yes -e VERSION=${VERSION} -v ${MAKEDIR}:/go/src/k8s.io/kops golang:${GOVERSION} make -f /go/src/k8s.io/kops/Makefile crossbuild
	docker cp kops-build-${UNIQUE}:/go/.build .

kops-dist: crossbuild-in-docker
	mkdir -p .build/dist/
	(sha1sum .build/dist/darwin/amd64/kops | cut -d' ' -f1) > .build/dist/darwin/amd64/kops.sha1
	(sha1sum .build/dist/linux/amd64/kops | cut -d' ' -f1) > .build/dist/linux/amd64/kops.sha1

version-dist: nodeup-dist kops-dist
	rm -rf .build/upload
	mkdir -p .build/upload/kops/${VERSION}/linux/amd64/
	mkdir -p .build/upload/kops/${VERSION}/darwin/amd64/
	cp .build/dist/nodeup .build/upload/kops/${VERSION}/linux/amd64/nodeup
	cp .build/dist/nodeup.sha1 .build/upload/kops/${VERSION}/linux/amd64/nodeup.sha1
	cp .build/dist/linux/amd64/kops .build/upload/kops/${VERSION}/linux/amd64/kops
	cp .build/dist/linux/amd64/kops.sha1 .build/upload/kops/${VERSION}/linux/amd64/kops.sha1
	cp .build/dist/darwin/amd64/kops .build/upload/kops/${VERSION}/darwin/amd64/kops
	cp .build/dist/darwin/amd64/kops.sha1 .build/upload/kops/${VERSION}/darwin/amd64/kops.sha1

upload: kops version-dist
	aws s3 sync --acl public-read .build/upload/ ${S3_BUCKET}

gcs-upload: version-dist
	@echo "== Logging gcloud info =="
	@gcloud info
	@echo "== Uploading kops =="
	gsutil -h "Cache-Control:private, max-age=0, no-transform" -m cp -n -r .build/upload/kops/* ${GCS_LOCATION}

gcs-publish-ci: gcs-upload
	echo "${GCS_URL}/${VERSION}" > .build/upload/${LATEST_FILE}
	gsutil -h "Cache-Control:private, max-age=0, no-transform" cp .build/upload/${LATEST_FILE} ${GCS_LOCATION}

gen-cli-docs:
	@kops genhelpdocs --out docs/cli

# Assumes running on linux for speed (todo: crossbuild on OSX?)
push: nodeup-gocode
	scp -C ${GOPATH_1ST}/bin/nodeup  ${TARGET}:/tmp/

push-gce-dry: push
	ssh ${TARGET} sudo SKIP_PACKAGE_UPDATE=1 /tmp/nodeup --conf=metadata://gce/config --dryrun --v=8

push-aws-dry: push
	ssh ${TARGET} sudo SKIP_PACKAGE_UPDATE=1 /tmp/nodeup --conf=/var/cache/kubernetes-install/kube_env.yaml --dryrun --v=8

push-gce-run: push
	ssh ${TARGET} sudo SKIP_PACKAGE_UPDATE=1 /tmp/nodeup --conf=metadata://gce/config --v=8

# -t is for CentOS http://unix.stackexchange.com/questions/122616/why-do-i-need-a-tty-to-run-sudo-if-i-can-sudo-without-a-password
push-aws-run: push
	ssh -t ${TARGET} sudo SKIP_PACKAGE_UPDATE=1 /tmp/nodeup --conf=/var/cache/kubernetes-install/kube_env.yaml --v=8



protokube-gocode:
	go install k8s.io/kops/protokube/cmd/protokube

protokube-builder-image:
	docker build -t protokube-builder images/protokube-builder

protokube-build-in-docker: protokube-builder-image
	docker run -it -e VERSION=${VERSION} -v `pwd`:/src protokube-builder /onbuild.sh

protokube-image: protokube-build-in-docker
	docker build -t ${DOCKER_REGISTRY}/protokube:${PROTOKUBE_TAG} -f images/protokube/Dockerfile .

protokube-push: protokube-image
	docker push ${DOCKER_REGISTRY}/protokube:${PROTOKUBE_TAG}

nodeup: nodeup-dist

nodeup-gocode: kops-gobindata
	go install ${EXTRA_BUILDFLAGS} -ldflags "${EXTRA_LDFLAGS} -X main.BuildVersion=${VERSION}" k8s.io/kops/cmd/nodeup

nodeup-dist:
	docker pull golang:${GOVERSION} # Keep golang image up to date
	docker run --name=nodeup-build-${UNIQUE} -e STATIC_BUILD=yes -e VERSION=${VERSION} -v ${MAKEDIR}:/go/src/k8s.io/kops golang:${GOVERSION} make -f /go/src/k8s.io/kops/Makefile nodeup-gocode
	mkdir -p .build/dist
	docker cp nodeup-build-${UNIQUE}:/go/bin/nodeup .build/dist/
	(sha1sum .build/dist/nodeup | cut -d' ' -f1) > .build/dist/nodeup.sha1

dns-controller-gocode:
	go install k8s.io/kops/dns-controller/cmd/dns-controller

dns-controller-builder-image:
	docker build -t dns-controller-builder images/dns-controller-builder

dns-controller-build-in-docker: dns-controller-builder-image
	docker run -it -e VERSION=${VERSION} -v `pwd`:/src dns-controller-builder /onbuild.sh

dns-controller-image: dns-controller-build-in-docker
	docker build -t ${DOCKER_REGISTRY}/dns-controller:${DNS_CONTROLLER_TAG}  -f images/dns-controller/Dockerfile .

dns-controller-push: dns-controller-image
	docker push ${DOCKER_REGISTRY}/dns-controller:${DNS_CONTROLLER_TAG}

# --------------------------------------------------
# development targets

# See docs/development/dependencies.md
copydeps:
	rsync -avz _vendor/ vendor/ --delete --exclude vendor/  --exclude .git

gofmt:
	gofmt -w -s channels/
	gofmt -w -s cloudmock/
	gofmt -w -s cmd/
	gofmt -w -s examples/
	gofmt -w -s federation/
	gofmt -w -s pkg/
	gofmt -w -s util/
	gofmt -w -s upup/pkg/
	gofmt -w -s pkg/
	gofmt -w -s protokube/cmd
	gofmt -w -s protokube/pkg
	gofmt -w -s dns-controller/cmd
	gofmt -w -s dns-controller/pkg


govet:
	go vet \
	  k8s.io/kops/cmd/... \
	  k8s.io/kops/pkg/... \
	  k8s.io/kops/channels/... \
	  k8s.io/kops/examples/... \
	  k8s.io/kops/federation/... \
	  k8s.io/kops/util/... \
	  k8s.io/kops/upup/... \
	  k8s.io/kops/protokube/... \
	  k8s.io/kops/dns-controller/...


# --------------------------------------------------
# Continuous integration targets

verify-boilerplate:
	sh -c hack/verify-boilerplate.sh

ci: kops nodeup-gocode examples test govet verify-boilerplate
	echo "Done!"

# --------------------------------------------------
# channel tool

channels: channels-gocode

channels-gocode:
	go install ${EXTRA_BUILDFLAGS} -ldflags "-X main.BuildVersion=${VERSION} ${EXTRA_LDFLAGS}" k8s.io/kops/channels/cmd/channels

# --------------------------------------------------
# API / embedding examples

examples:
	go install k8s.io/kops/examples/kops-api-example/...

# -----------------------------------------------------
# api machinery regenerate

apimachinery:
	#go install ./cmd/libs/go2idl/conversion-gen
	#~/k8s/bin/conversion-gen  --input-dirs k8s.io/kops/pkg/apis/kops/v1alpha1 --v=8  --output-file-base=zz_generated.conversion
	#go install github.com/ugorji/go/codec/codecgen
	# codecgen work only if invoked from directory where the file is located.
	cd pkg/apis/kops/v1alpha1/ && ~/k8s/bin/codecgen -d 1234 -o types.generated.go instancegroup.go cluster.go federation.go
	cd pkg/apis/kops/ && ~/k8s/bin/codecgen -d 1234 -o types.generated.go instancegroup.go cluster.go federation.go
