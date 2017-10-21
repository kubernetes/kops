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


DOCKER_REGISTRY?=gcr.io/must-override
S3_BUCKET?=s3://must-override/
GCS_LOCATION?=gs://must-override
GCS_URL=$(GCS_LOCATION:gs://%=https://storage.googleapis.com/%)
LATEST_FILE?=latest-ci.txt
GOPATH_1ST=$(shell go env | grep GOPATH | cut -f 2 -d \")
UNIQUE:=$(shell date +%s)
GOVERSION=1.8.3
BUILD=$(GOPATH_1ST)/src/k8s.io/kops/.build
LOCAL=$(BUILD)/local
BINDATA_TARGETS=upup/models/bindata.go federation/model/bindata.go
ARTIFACTS=$(BUILD)/artifacts
DIST=$(BUILD)/dist
IMAGES=$(DIST)/images
GOBINDATA=$(LOCAL)/go-bindata
CHANNELS=$(LOCAL)/channels
NODEUP=$(LOCAL)/nodeup
PROTOKUBE=$(LOCAL)/protokube
UPLOAD=$(BUILD)/upload
UID:=$(shell id -u)
GID:=$(shell id -g)
TESTABLE_PACKAGES:=$(shell egrep -v "k8s.io/kops/cloudmock|k8s.io/kops/vendor" hack/.packages) 

SOURCES:=$(shell find . -name "*.go")

# See http://stackoverflow.com/questions/18136918/how-to-get-current-relative-directory-of-your-makefile
MAKEDIR:=$(strip $(shell dirname "$(realpath $(lastword $(MAKEFILE_LIST)))"))

# Keep in sync with upup/models/cloudup/resources/addons/dns-controller/
DNS_CONTROLLER_TAG=1.7.1

KOPS_RELEASE_VERSION = 1.8.0-alpha.1
KOPS_CI_VERSION      = 1.8.0-alpha.2

# kops local location
KOPS                 = ${LOCAL}/kops

# kops source root directory (without trailing /)
KOPS_ROOT           ?= $(patsubst %/,%,$(abspath $(dir $(firstword $(MAKEFILE_LIST)))))

GITSHA := $(shell cd ${GOPATH_1ST}/src/k8s.io/kops; git describe --always)

ifndef VERSION
  # To keep both CI and end-users building from source happy,
  # we expect that CI sets CI=1.
  #
  # For end users, they need only build kops, and they can use the last
  # released version of nodeup/protokube.
  # For CI, we continue to build a synthetic version from the git SHA, so
  # we never cross versions.
  #
  # We expect that if you are uploading nodeup/protokube, you will set
  # VERSION (along with S3_BUCKET), either directly or by setting CI=1
  ifndef CI
    VERSION=${KOPS_RELEASE_VERSION}
  else
    VERSION := ${KOPS_CI_VERSION}+${GITSHA}
  endif
endif

# + is valid in semver, but not in docker tags. Fixup CI versions.
# Note that this mirrors the logic in DefaultProtokubeImageName
PROTOKUBE_TAG := $(subst +,-,${VERSION})
KOPS_SERVER_TAG := $(subst +,-,${VERSION})

# Go exports:

GO15VENDOREXPERIMENT=1
export GO15VENDOREXPERIMENT

ifdef STATIC_BUILD
  CGO_ENABLED=0
  export CGO_ENABLED
  EXTRA_BUILDFLAGS=-installsuffix cgo
  EXTRA_LDFLAGS=-s
endif

SHASUMCMD := $(shell command -v sha1sum || command -v shasum; 2> /dev/null)

ifndef SHASUMCMD
  $(error "Neither sha1sum nor shasum command is available")
endif

.PHONY: kops-install # Install kops to local $GOPATH/bin
kops-install: gobindata-tool ${BINDATA_TARGETS}
	go install ${EXTRA_BUILDFLAGS} -ldflags "-X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA} ${EXTRA_LDFLAGS}" k8s.io/kops/cmd/kops/

.PHONY: channels-install # Install channels to local $GOPATH/bin
channels-install: ${CHANNELS}
	cp ${CHANNELS} ${GOPATH_1ST}/bin

.PHONY: all-install # Install all kops project binaries
all-install: all kops-install channels-install
	cp ${NODEUP} ${GOPATH_1ST}/bin
	cp ${PROTOKUBE} ${GOPATH_1ST}/bin

.PHONY: all
all: ${KOPS} ${PROTOKUBE} ${NODEUP} ${CHANNELS}

.PHONY: help
help: # Show this help
	@{ \
	echo 'Targets:'; \
	echo ''; \
	grep '^[a-z/.-]*: .*# .*' Makefile \
	| sort \
	| sed 's/: \(.*\) # \(.*\)/ - \2 (deps: \1)/' `: fmt targets w/ deps` \
	| sed 's/:.*#/ -/'                            `: fmt targets w/o deps` \
	| sed 's/^/    /'                             `: indent`; \
	echo ''; \
	echo 'CLI options:'; \
	echo ''; \
	grep '^[^\s]*?=' Makefile \
	| sed 's/\?=\(.*\)/ (default: "\1")/' `: fmt default values`\
	| sed 's/^/    /'; \
	echo ''; \
	echo 'Undocumented targets:'; \
	echo ''; \
	grep '^[a-z/.-]*:\( [^#=]*\)*$$' Makefile \
	| sort \
	| sed 's/: \(.*\)/ (deps: \1)/' `: fmt targets w/ deps` \
	| sed 's/:$$//'                 `: fmt targets w/o deps` \
	| sed 's/^/    /'               `: indent`; \
	echo ''; \
	} 1>&2; \

.PHONY: clean
clean: # Remove build directory and bindata-generated files
	for t in ${BINDATA_TARGETS}; do if test -e $$t; then rm -fv $$t; fi; done 
	if test -e ${BUILD}; then rm -rfv ${BUILD}; fi

.PHONY: kops
kops: ${KOPS}

${KOPS}: ${BINDATA_TARGETS} ${SOURCES}
	go build ${EXTRA_BUILDFLAGS} -ldflags "-X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA} ${EXTRA_LDFLAGS}" -o $@ k8s.io/kops/cmd/kops/

${GOBINDATA}:
	mkdir -p ${LOCAL}
	go build ${EXTRA_BUILDFLAGS} -ldflags "${EXTRA_LDFLAGS}" -o $@ k8s.io/kops/vendor/github.com/jteeuwen/go-bindata/go-bindata

.PHONY: gobindata-tool
gobindata-tool: ${GOBINDATA}

.PHONY: kops-gobindata
kops-gobindata: gobindata-tool ${BINDATA_TARGETS}

UPUP_MODELS_BINDATA_SOURCES:=$(shell find upup/models/ | egrep -v "upup/models/bindata.go")
upup/models/bindata.go: ${GOBINDATA} ${UPUP_MODELS_BINDATA_SOURCES}
	cd ${GOPATH_1ST}/src/k8s.io/kops; ${GOBINDATA} -o $@ -pkg models -ignore="\\.DS_Store" -ignore="bindata\\.go" -ignore="vfs\\.go" -prefix upup/models/ upup/models/...

FEDERATION_MODELS_BINDATA_SOURCES:=$(shell find federation/model/ | egrep -v "federation/model/bindata.go")
federation/model/bindata.go: ${GOBINDATA} ${FEDERATION_MODELS_BINDATA_SOURCES}
	cd ${GOPATH_1ST}/src/k8s.io/kops; ${GOBINDATA} -o $@ -pkg model -ignore="\\.DS_Store" -ignore="bindata\\.go" -prefix federation/model/ federation/model/...

# Build in a docker container with golang 1.X
# Used to test we have not broken 1.X
# 1.8 is preferred, 1.9 is coming soon so we have a target for it
.PHONY: check-builds-in-go18
check-builds-in-go18:
	docker run -v ${GOPATH_1ST}/src/k8s.io/kops:/go/src/k8s.io/kops golang:1.8 make -C /go/src/k8s.io/kops ci

.PHONY: check-builds-in-go19
check-builds-in-go19:
	docker run -v ${GOPATH_1ST}/src/k8s.io/kops:/go/src/k8s.io/kops golang:1.9 make -C /go/src/k8s.io/kops ci


.PHONY: codegen
codegen: kops-gobindata
	go install k8s.io/kops/upup/tools/generators/...
	PATH=${GOPATH_1ST}/bin:${PATH} go generate k8s.io/kops/upup/pkg/fi/cloudup/awstasks
	PATH=${GOPATH_1ST}/bin:${PATH} go generate k8s.io/kops/upup/pkg/fi/cloudup/gcetasks
	PATH=${GOPATH_1ST}/bin:${PATH} go generate k8s.io/kops/upup/pkg/fi/cloudup/dotasks
	PATH=${GOPATH_1ST}/bin:${PATH} go generate k8s.io/kops/upup/pkg/fi/assettasks
	PATH=${GOPATH_1ST}/bin:${PATH} go generate k8s.io/kops/upup/pkg/fi/fitasks

.PHONY: protobuf
protobuf: protokube/pkg/gossip/mesh/mesh.pb.go

protokube/pkg/gossip/mesh/mesh.pb.go: protokube/pkg/gossip/mesh/mesh.proto
	cd ${GOPATH_1ST}/src; protoc --gofast_out=. k8s.io/kops/protokube/pkg/gossip/mesh/mesh.proto

.PHONY: hooks
hooks: # Install Git hooks
	cp hack/pre-commit.sh .git/hooks/pre-commit

.PHONY: test
test: ${BINDATA_TARGETS}  # Run tests locally
	go test -v ${TESTABLE_PACKAGES}

.PHONY: ${DIST}/linux/amd64/nodeup
${DIST}/linux/amd64/nodeup: ${BINDATA_TARGETS}
	mkdir -p ${DIST}
	GOOS=linux GOARCH=amd64 go build -a ${EXTRA_BUILDFLAGS} -o $@ -ldflags "${EXTRA_LDFLAGS} -X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA}" k8s.io/kops/cmd/nodeup

.PHONY: crossbuild-nodeup
crossbuild-nodeup: ${DIST}/linux/amd64/nodeup

.PHONY: crossbuild-nodeup-in-docker
crossbuild-nodeup-in-docker:
	docker pull golang:${GOVERSION} # Keep golang image up to date
	docker run --name=nodeup-build-${UNIQUE} -e STATIC_BUILD=yes -e VERSION=${VERSION} -v ${MAKEDIR}:/go/src/k8s.io/kops golang:${GOVERSION} make -C /go/src/k8s.io/kops/ crossbuild-nodeup
	docker cp nodeup-build-${UNIQUE}:/go/.build .

.PHONY: ${DIST}/darwin/amd64/kops
${DIST}/darwin/amd64/kops: ${BINDATA_TARGETS}
	mkdir -p ${DIST}
	GOOS=darwin GOARCH=amd64 go build -a ${EXTRA_BUILDFLAGS} -o $@ -ldflags "${EXTRA_LDFLAGS} -X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA}" k8s.io/kops/cmd/kops

.PHONY: ${DIST}/linux/amd64/kops
${DIST}/linux/amd64/kops: ${BINDATA_TARGETS}
	mkdir -p ${DIST}
	GOOS=linux GOARCH=amd64 go build -a ${EXTRA_BUILDFLAGS} -o $@ -ldflags "${EXTRA_LDFLAGS} -X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA}" k8s.io/kops/cmd/kops

.PHONY: crossbuild
crossbuild: ${DIST}/darwin/amd64/kops ${DIST}/linux/amd64/kops

.PHONY: crossbuild-in-docker
crossbuild-in-docker:
	docker pull golang:${GOVERSION} # Keep golang image up to date
	docker run --name=kops-build-${UNIQUE} -e STATIC_BUILD=yes -e VERSION=${VERSION} -v ${MAKEDIR}:/go/src/k8s.io/kops golang:${GOVERSION} make -C /go/src/k8s.io/kops/ crossbuild
	docker start kops-build-${UNIQUE}
	docker exec kops-build-${UNIQUE} chown -R ${UID}:${GID} /go/src/k8s.io/kops/.build
	docker cp kops-build-${UNIQUE}:/go/src/k8s.io/kops/.build .
	docker kill kops-build-${UNIQUE}
	docker rm kops-build-${UNIQUE}

.PHONY: kops-dist
kops-dist: crossbuild-in-docker
	mkdir -p ${DIST}
	(${SHASUMCMD} ${DIST}/darwin/amd64/kops | cut -d' ' -f1) > ${DIST}/darwin/amd64/kops.sha1
	(${SHASUMCMD} ${DIST}/linux/amd64/kops | cut -d' ' -f1) > ${DIST}/linux/amd64/kops.sha1

.PHONY: version-dist
version-dist: nodeup-dist kops-dist protokube-export utils-dist
	rm -rf ${UPLOAD}
	mkdir -p ${UPLOAD}/kops/${VERSION}/linux/amd64/
	mkdir -p ${UPLOAD}/kops/${VERSION}/darwin/amd64/
	mkdir -p ${UPLOAD}/kops/${VERSION}/images/
	mkdir -p ${UPLOAD}/utils/${VERSION}/linux/amd64/
	cp ${DIST}/nodeup ${UPLOAD}/kops/${VERSION}/linux/amd64/nodeup
	cp ${DIST}/nodeup.sha1 ${UPLOAD}/kops/${VERSION}/linux/amd64/nodeup.sha1
	cp ${IMAGES}/protokube.tar.gz ${UPLOAD}/kops/${VERSION}/images/protokube.tar.gz
	cp ${IMAGES}/protokube.tar.gz.sha1 ${UPLOAD}/kops/${VERSION}/images/protokube.tar.gz.sha1
	cp ${DIST}/linux/amd64/kops ${UPLOAD}/kops/${VERSION}/linux/amd64/kops
	cp ${DIST}/linux/amd64/kops.sha1 ${UPLOAD}/kops/${VERSION}/linux/amd64/kops.sha1
	cp ${DIST}/darwin/amd64/kops ${UPLOAD}/kops/${VERSION}/darwin/amd64/kops
	cp ${DIST}/darwin/amd64/kops.sha1 ${UPLOAD}/kops/${VERSION}/darwin/amd64/kops.sha1
	cp ${DIST}/linux/amd64/utils.tar.gz ${UPLOAD}/kops/${VERSION}/linux/amd64/utils.tar.gz
	cp ${DIST}/linux/amd64/utils.tar.gz.sha1 ${UPLOAD}/kops/${VERSION}/linux/amd64/utils.tar.gz.sha1

.PHONY: vsphere-version-dist
vsphere-version-dist: nodeup-dist protokube-export
	rm -rf ${UPLOAD}
	mkdir -p ${UPLOAD}/kops/${VERSION}/linux/amd64/
	mkdir -p ${UPLOAD}/kops/${VERSION}/darwin/amd64/
	mkdir -p ${UPLOAD}/kops/${VERSION}/images/
	mkdir -p ${UPLOAD}/utils/${VERSION}/linux/amd64/
	cp ${DIST}/nodeup ${UPLOAD}/kops/${VERSION}/linux/amd64/nodeup
	cp ${DIST}/nodeup.sha1 ${UPLOAD}/kops/${VERSION}/linux/amd64/nodeup.sha1
	cp ${IMAGES}/protokube.tar.gz ${UPLOAD}/kops/${VERSION}/images/protokube.tar.gz
	cp ${IMAGES}/protokube.tar.gz.sha1 ${UPLOAD}/kops/${VERSION}/images/protokube.tar.gz.sha1
	scp -r .build/dist/nodeup* ${TARGET}:${TARGET_PATH}/nodeup
	scp -r .build/dist/images/protokube.tar.gz* ${TARGET}:${TARGET_PATH}/protokube/
	make kops-dist
	cp ${DIST}/linux/amd64/kops ${UPLOAD}/kops/${VERSION}/linux/amd64/kops
	cp ${DIST}/linux/amd64/kops.sha1 ${UPLOAD}/kops/${VERSION}/linux/amd64/kops.sha1
	cp ${DIST}/darwin/amd64/kops ${UPLOAD}/kops/${VERSION}/darwin/amd64/kops
	cp ${DIST}/darwin/amd64/kops.sha1 ${UPLOAD}/kops/${VERSION}/darwin/amd64/kops.sha1

.PHONY: upload
upload: version-dist # Upload kops to S3
	aws s3 sync --acl public-read ${UPLOAD}/ ${S3_BUCKET}

.PHONY: gcs-upload
gcs-upload: version-dist # Upload kops to GCS
	@echo "== Uploading kops =="
	gsutil -h "Cache-Control:private, max-age=0, no-transform" -m cp -n -r ${UPLOAD}/kops/* ${GCS_LOCATION}

# In CI testing, always upload the CI version.
.PHONY: gcs-publish-ci
gcs-publish-ci: VERSION := ${KOPS_CI_VERSION}+${GITSHA}
gcs-publish-ci: PROTOKUBE_TAG := $(subst +,-,${VERSION})
gcs-publish-ci: gcs-upload
	echo "VERSION: ${VERSION}"
	echo "PROTOKUBE_TAG: ${PROTOKUBE_TAG}"
	echo "${GCS_URL}/${VERSION}" > ${UPLOAD}/${LATEST_FILE}
	gsutil -h "Cache-Control:private, max-age=0, no-transform" cp ${UPLOAD}/${LATEST_FILE} ${GCS_LOCATION}

.PHONY: gen-cli-docs
gen-cli-docs: ${KOPS} # Regenerate CLI docs
	KOPS_STATE_STORE= \
	KOPS_FEATURE_FLAGS= \
	${KOPS} genhelpdocs --out docs/cli

.PHONY: gen-api-docs
gen-api-docs:
	# Follow procedure in docs/apireference/README.md
	hack/make-gendocs.sh
	# Update the `pkg/openapi/openapi_generated.go`
	${GOPATH}/bin/apiserver-boot build generated --generator openapi --copyright hack/boilerplate/boilerplate.go.txt
	go install k8s.io/kops/cmd/kops-server
	${GOPATH}/bin/apiserver-boot build docs --disable-delegated-auth=false --output-dir docs/apireference --server kops-server

.PHONY: push
# Will always push a linux-based build up to the server
push: crossbuild-nodeup
	scp -C .build/dist/linux/amd64/nodeup  ${TARGET}:/tmp/

.PHONY: push-gce-dry
push-gce-dry: push
	ssh ${TARGET} sudo SKIP_PACKAGE_UPDATE=1 /tmp/nodeup --conf=metadata://gce/config --dryrun --v=8

.PHONY: push-gce-dry
push-aws-dry: push
	ssh ${TARGET} sudo SKIP_PACKAGE_UPDATE=1 /tmp/nodeup --conf=/var/cache/kubernetes-install/kube_env.yaml --dryrun --v=8

.PHONY: push-gce-run
push-gce-run: push
	ssh ${TARGET} sudo cp /tmp/nodeup /var/lib/toolbox/kubernetes-install/nodeup
	ssh ${TARGET} sudo SKIP_PACKAGE_UPDATE=1 /var/lib/toolbox/kubernetes-install/nodeup --conf=/var/lib/toolbox/kubernetes-install/kube_env.yaml --v=8

# -t is for CentOS http://unix.stackexchange.com/questions/122616/why-do-i-need-a-tty-to-run-sudo-if-i-can-sudo-without-a-password
.PHONY: push-aws-run
push-aws-run: push
	ssh -t ${TARGET} sudo SKIP_PACKAGE_UPDATE=1 /tmp/nodeup --conf=/var/cache/kubernetes-install/kube_env.yaml --v=8

${PROTOKUBE}: ${SOURCES}
	go build -o $@ -tags 'peer_name_alternative peer_name_hash' k8s.io/kops/protokube/cmd/protokube

.PHONY: protokube
protokube: ${PROTOKUBE}

.PHONY: protokube-builder-image
protokube-builder-image:
	docker build -t protokube-builder images/protokube-builder

.PHONY: protokube-build-in-docker
protokube-build-in-docker: protokube-builder-image
	mkdir -p ${IMAGES} # We have to create the directory first, so docker doesn't mess up the ownership of the dir
	docker run -t -e VERSION=${VERSION} -e HOST_UID=${UID} -e HOST_GID=${GID} -v `pwd`:/src protokube-builder /onbuild.sh

.PHONY: protokube-image
protokube-image: protokube-build-in-docker
	docker build -t protokube:${PROTOKUBE_TAG} -f images/protokube/Dockerfile .

.PHONY: protokube-export
protokube-export: protokube-image
	docker save protokube:${PROTOKUBE_TAG} > ${IMAGES}/protokube.tar
	gzip --force --best ${IMAGES}/protokube.tar
	(${SHASUMCMD} ${IMAGES}/protokube.tar.gz | cut -d' ' -f1) > ${IMAGES}/protokube.tar.gz.sha1

# protokube-push is no longer used (we upload a docker image tar file to S3 instead),
# but we're keeping it around in case it is useful for development etc
.PHONY: protokube-push
protokube-push: protokube-image
	docker tag protokube:${PROTOKUBE_TAG} ${DOCKER_REGISTRY}/protokube:${PROTOKUBE_TAG}
	docker push ${DOCKER_REGISTRY}/protokube:${PROTOKUBE_TAG}

.PHONY: nodeup
nodeup: ${NODEUP}

${NODEUP}: ${BINDATA_TARGETS} ${SOURCES}
	go build ${EXTRA_BUILDFLAGS} -ldflags "${EXTRA_LDFLAGS} -X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA}" -o $@ k8s.io/kops/cmd/nodeup

.PHONY: nodeup-dist
nodeup-dist:
	mkdir -p ${DIST}
	docker pull golang:${GOVERSION} # Keep golang image up to date
	docker run --name=nodeup-build-${UNIQUE} -e STATIC_BUILD=yes -e VERSION=${VERSION} -v ${MAKEDIR}:/go/src/k8s.io/kops golang:${GOVERSION} make -C /go/src/k8s.io/kops/ nodeup
	docker start nodeup-build-${UNIQUE}
	docker exec nodeup-build-${UNIQUE} chown -R ${UID}:${GID} /go/src/k8s.io/kops/.build
	docker cp nodeup-build-${UNIQUE}:/go/src/k8s.io/kops/.build/local/nodeup .build/dist/
	(${SHASUMCMD} .build/dist/nodeup | cut -d' ' -f1) > .build/dist/nodeup.sha1

.PHONY: dns-controller-gocode
dns-controller-gocode:
	go install -tags 'peer_name_alternative peer_name_hash' -ldflags "${EXTRA_LDFLAGS} -X main.BuildVersion=${DNS_CONTROLLER_TAG}" k8s.io/kops/dns-controller/cmd/dns-controller

.PHONY: dns-controller-builder-image
dns-controller-builder-image:
	docker build -t dns-controller-builder images/dns-controller-builder

.PHONY: dns-controller-build-in-docker
dns-controller-build-in-docker: dns-controller-builder-image
	docker run -t -e HOST_UID=${UID} -e HOST_GID=${GID} -v `pwd`:/src dns-controller-builder /onbuild.sh

.PHONY: dns-controller-image
dns-controller-image: dns-controller-build-in-docker
	docker build -t ${DOCKER_REGISTRY}/dns-controller:${DNS_CONTROLLER_TAG}  -f images/dns-controller/Dockerfile .

.PHONY: dns-controller-push
dns-controller-push: dns-controller-image
	docker push ${DOCKER_REGISTRY}/dns-controller:${DNS_CONTROLLER_TAG}

# --------------------------------------------------
# static utils

.PHONY: utils-dist
utils-dist:
	docker build -t utils-builder images/utils-builder
	mkdir -p ${DIST}/linux/amd64/
	docker run -v `pwd`/.build/dist/linux/amd64/:/dist utils-builder /extract.sh

# --------------------------------------------------
# development targets

# See docs/development/dependencies.md
.PHONY: copydeps
copydeps:
	rsync -avz _vendor/ vendor/ --delete --exclude vendor/  --exclude .git --exclude BUILD --exclude BUILD.bazel
	ln -sf kubernetes/staging/src/k8s.io/api vendor/k8s.io/api
	ln -sf kubernetes/staging/src/k8s.io/apiextensions-apiserver vendor/k8s.io/apiextensions-apiserver
	ln -sf kubernetes/staging/src/k8s.io/apimachinery vendor/k8s.io/apimachinery
	ln -sf kubernetes/staging/src/k8s.io/apiserver vendor/k8s.io/apiserver
	ln -sf kubernetes/staging/src/k8s.io/client-go vendor/k8s.io/client-go
	ln -sf kubernetes/staging/src/k8s.io/code-generator vendor/k8s.io/code-generator
	ln -sf kubernetes/staging/src/k8s.io/metrics vendor/k8s.io/metrics

.PHONY: gofmt
gofmt:
	gofmt -w -s channels/
	gofmt -w -s cloudmock/
	gofmt -w -s cmd/
	gofmt -w -s examples/
	gofmt -w -s federation/
	gofmt -w -s nodeup/
	gofmt -w -s util/
	gofmt -w -s upup/pkg/
	gofmt -w -s pkg/
	gofmt -w -s tests/
	gofmt -w -s protokube/cmd
	gofmt -w -s protokube/pkg
	gofmt -w -s protokube/tests
	gofmt -w -s dns-controller/cmd
	gofmt -w -s dns-controller/pkg

.PHONY: goimports
goimports:
	hack/update-goimports

.PHONY: verify-goimports
verify-goimports:
	hack/verify-goimports

.PHONY: govet
govet: ${BINDATA_TARGETS}
	go vet ${TESTABLE_PACKAGES}

# --------------------------------------------------
# Continuous integration targets

.PHONY: verify-boilerplate
verify-boilerplate:
	hack/verify-boilerplate.sh

.PHONY: verify-gofmt
verify-gofmt:
	hack/verify-gofmt.sh

.PHONY: verify-packages
verify-packages: ${BINDATA_TARGETS}
	hack/verify-packages.sh

.PHONY: verify-gendocs
verify-gendocs: ${KOPS}
	@TMP_DOCS="$$(mktemp -d)"; \
	'${KOPS}' genhelpdocs --out "$$TMP_DOCS"; \
	\
	if ! diff -r "$$TMP_DOCS" '${KOPS_ROOT}/docs/cli'; then \
	     echo "Please run make gen-cli-docs." 1>&2; \
	     exit 1; \
	fi
	@echo "cli docs up-to-date"
#
# verify-gendocs will call kops target
# verify-package has to be after verify-gendoc, because with .gitignore for federation bindata
# it bombs in travis. verify-gendoc generates the bindata file.
.PHONY: ci
ci: govet verify-gofmt verify-boilerplate nodeup examples test | verify-gendocs verify-packages
	echo "Done!"

# --------------------------------------------------
# channel tool

.PHONY: channels
channels: ${CHANNELS}

${CHANNELS}: ${SOURCES}
	go build ${EXTRA_BUILDFLAGS} -o $@ -ldflags "-X k8s.io/kops.Version=${VERSION} ${EXTRA_LDFLAGS}" k8s.io/kops/channels/cmd/channels

# --------------------------------------------------
# release tasks

.PHONY: release-tag
release-tag:
	git tag ${KOPS_RELEASE_VERSION}

.PHONY: release-github
release-github:
	shipbot -tag ${KOPS_RELEASE_VERSION} -config .shipbot.yaml

# --------------------------------------------------
# API / embedding examples

.PHONY: examples
examples: ${BINDATA_TARGETS} # Install kops API example
	go install k8s.io/kops/examples/kops-api-example/...

# -----------------------------------------------------
# api machinery regenerate

.PHONY: apimachinery
apimachinery:
	sh -c hack/make-apimachinery.sh
	${GOPATH}/bin/conversion-gen --skip-unsafe=true --input-dirs k8s.io/kops/pkg/apis/kops/v1alpha1 --v=0  --output-file-base=zz_generated.conversion
	${GOPATH}/bin/conversion-gen --skip-unsafe=true --input-dirs k8s.io/kops/pkg/apis/kops/v1alpha2 --v=0  --output-file-base=zz_generated.conversion
	${GOPATH}/bin/deepcopy-gen --input-dirs k8s.io/kops/pkg/apis/kops --v=0  --output-file-base=zz_generated.deepcopy
	${GOPATH}/bin/deepcopy-gen --input-dirs k8s.io/kops/pkg/apis/kops/v1alpha1 --v=0  --output-file-base=zz_generated.deepcopy
	${GOPATH}/bin/deepcopy-gen --input-dirs k8s.io/kops/pkg/apis/kops/v1alpha2 --v=0  --output-file-base=zz_generated.deepcopy
	${GOPATH}/bin/defaulter-gen --input-dirs k8s.io/kops/pkg/apis/kops/v1alpha1 --v=0  --output-file-base=zz_generated.defaults
	${GOPATH}/bin/defaulter-gen --input-dirs k8s.io/kops/pkg/apis/kops/v1alpha2 --v=0  --output-file-base=zz_generated.defaults
	#go install github.com/ugorji/go/codec/codecgen
	# codecgen works only if invoked from directory where the file is located.
	#cd pkg/apis/kops/v1alpha2/ && ~/k8s/bin/codecgen -d 1234 -o types.generated.go instancegroup.go cluster.go federation.go
	#cd pkg/apis/kops/v1alpha1/ && ~/k8s/bin/codecgen -d 1234 -o types.generated.go instancegroup.go cluster.go federation.go
	#cd pkg/apis/kops/ && ~/k8s/bin/codecgen -d 1234 -o types.generated.go instancegroup.go cluster.go federation.go
	${GOPATH}/bin/client-gen  --input-base k8s.io/kops/pkg/apis/ --input="kops/,kops/v1alpha1,kops/v1alpha2" --clientset-path k8s.io/kops/pkg/client/clientset_generated/
	${GOPATH}/bin/client-gen  --clientset-name="clientset" --input-base k8s.io/kops/pkg/apis/ --input="kops/,kops/v1alpha1,kops/v1alpha2" --clientset-path k8s.io/kops/pkg/client/clientset_generated/


# -----------------------------------------------------
# kops-server

.PHONY: kops-server-docker-compile
kops-server-docker-compile:
	GOOS=linux GOARCH=amd64 go build -a ${EXTRA_BUILDFLAGS} -o ${DIST}/linux/amd64/kops-server -ldflags "${EXTRA_LDFLAGS} -X k8s.io/kops-server.Version=${VERSION} -X k8s.io/kops-server.GitVersion=${GITSHA}" k8s.io/kops/cmd/kops-server

.PHONY: kops-server-build
kops-server-build:
	# Compile the API binary in linux, and copy to local filesystem
	docker pull golang:${GOVERSION}
	docker run --name=kops-server-build-${UNIQUE} -e STATIC_BUILD=yes -e VERSION=${VERSION} -v ${GOPATH}/src:/go/src -v ${MAKEDIR}:/go/src/k8s.io/kops golang:${GOVERSION} make -C /go/src/k8s.io/kops/ kops-server-docker-compile
	docker cp kops-server-build-${UNIQUE}:/go/.build .
	docker build -t ${DOCKER_REGISTRY}/kops-server:${KOPS_SERVER_TAG} -f images/kops-server/Dockerfile .

.PHONY: kops-server-push
kops-server-push: kops-server-build
	docker push ${DOCKER_REGISTRY}/kops-server:latest

# -----------------------------------------------------
# bazel targets

.PHONY: bazel-test
bazel-test:
	bazel test //cmd/... //pkg/... //channels/... //nodeup/... //channels/... //protokube/... //dns-controller/... //upup/... //util/... --test_output=errors

.PHONY: bazel-build
bazel-build:
	bazel build //cmd/... //pkg/... //channels/... //nodeup/... //channels/... //protokube/... //dns-controller/...

# TODO: Get working on a mac / windows machine!
# 	GOOS=linux GOARCH=amd64 go build -a ${EXTRA_BUILDFLAGS} -o $@ -ldflags "${EXTRA_LDFLAGS} -X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA}" k8s.io/kops/cmd/nodeup
.PHONY: bazel-crossbuild-nodeup
bazel-crossbuild-nodeup:
	bazel build //cmd/nodeup

.PHONY: bazel-push
# Will always push a linux-based build up to the server
bazel-push: bazel-crossbuild-nodeup
	scp -C bazel-bin/cmd/nodeup/nodeup  ${TARGET}:/tmp/

.PHONY: bazel-push-gce-run
bazel-push-gce-run: bazel-push
	ssh ${TARGET} sudo cp /tmp/nodeup /var/lib/toolbox/kubernetes-install/nodeup
	ssh ${TARGET} sudo SKIP_PACKAGE_UPDATE=1 /var/lib/toolbox/kubernetes-install/nodeup --conf=/var/lib/toolbox/kubernetes-install/kube_env.yaml --v=8

# -t is for CentOS http://unix.stackexchange.com/questions/122616/why-do-i-need-a-tty-to-run-sudo-if-i-can-sudo-without-a-password
.PHONY: bazel-push-aws-run
bazel-push-aws-run: bazel-push
	ssh -t ${TARGET} sudo SKIP_PACKAGE_UPDATE=1 /tmp/nodeup --conf=/var/cache/kubernetes-install/kube_env.yaml --v=8
