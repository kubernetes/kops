# Copyright 2019 The Kubernetes Authors.
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


# kops source root directory (without trailing /)
KOPS_ROOT?=$(patsubst %/,%,$(abspath $(dir $(lastword $(MAKEFILE_LIST)))))
DOCKER_REGISTRY?=gcr.io/must-override
S3_BUCKET?=s3://must-override/
UPLOAD_DEST?=$(S3_BUCKET)
GCS_LOCATION?=gs://must-override
GCS_URL=$(GCS_LOCATION:gs://%=https://storage.googleapis.com/%)
LATEST_FILE?=latest-ci.txt
GOPATH_1ST:=$(shell go env | grep GOPATH | cut -f 2 -d \")
UNIQUE:=$(shell date +%s)
GOVERSION=1.13.8
BUILD=$(KOPS_ROOT)/.build
LOCAL=$(BUILD)/local
BINDATA_TARGETS=upup/models/bindata.go
ARTIFACTS=$(BUILD)/artifacts
DIST=$(BUILD)/dist
IMAGES=$(DIST)/images
GOBINDATA=$(LOCAL)/go-bindata
CHANNELS=$(LOCAL)/channels
NODEUP=$(LOCAL)/nodeup
PROTOKUBE=$(LOCAL)/protokube
UPLOAD=$(BUILD)/upload
BAZELBUILD=$(KOPS_ROOT)/.bazelbuild
BAZELDIST=$(BAZELBUILD)/dist
BAZELIMAGES=$(BAZELDIST)/images
BAZELUPLOAD=$(BAZELBUILD)/upload
UID:=$(shell id -u)
GID:=$(shell id -g)
BAZEL_OPTIONS?=
BAZEL_CONFIG?=
API_OPTIONS?=
GCFLAGS?=

UPLOAD_CMD=$(KOPS_ROOT)/hack/upload

# Unexport environment variables that can affect tests and are not used in builds
unexport AWS_ACCESS_KEY_ID AWS_REGION AWS_SECRET_ACCESS_KEY AWS_SESSION_TOKEN CNI_VERSION_URL DNS_IGNORE_NS_CHECK DNSCONTROLLER_IMAGE DO_ACCESS_TOKEN GOOGLE_APPLICATION_CREDENTIALS
unexport KOPS_BASE_URL KOPS_CLUSTER_NAME KOPS_RUN_OBSOLETE_VERSION KOPS_STATE_STORE KOPS_STATE_S3_ACL KUBE_API_VERSIONS NODEUP_URL OPENSTACK_CREDENTIAL_FILE PROTOKUBE_IMAGE SKIP_PACKAGE_UPDATE
unexport SKIP_REGION_CHECK S3_ACCESS_KEY_ID S3_ENDPOINT S3_REGION S3_SECRET_ACCESS_KEY VSPHERE_USERNAME VSPHERE_PASSWORD

# Keep in sync with upup/models/cloudup/resources/addons/dns-controller/
DNS_CONTROLLER_TAG=1.18.0-alpha.2
# Keep in sync with upup/models/cloudup/resources/addons/kops-controller.addons.k8s.io/
KOPS_CONTROLLER_TAG=1.18.0-alpha.2

# Keep in sync with logic in get_workspace_status
# TODO: just invoke tools/get_workspace_status.sh?
KOPS_RELEASE_VERSION:=$(shell grep 'KOPS_RELEASE_VERSION\s*=' version.go | awk '{print $$3}' | sed -e 's_"__g')
KOPS_CI_VERSION:=$(shell grep 'KOPS_CI_VERSION\s*=' version.go | awk '{print $$3}' | sed -e 's_"__g')

# kops local location
KOPS                 = ${LOCAL}/kops

GITSHA := $(shell cd ${KOPS_ROOT}; git describe --always)

# Keep in sync with logic in get_workspace_status
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
  # VERSION (along with UPLOAD_DEST), either directly or by setting CI=1
  ifndef CI
    VERSION=${KOPS_RELEASE_VERSION}
  else
    VERSION := ${KOPS_CI_VERSION}+${GITSHA}
  endif
endif

# + is valid in semver, but not in docker tags. Fixup CI versions.
# Note that this mirrors the logic in DefaultProtokubeImageName
PROTOKUBE_TAG := $(subst +,-,${VERSION})

# Go exports:
LDFLAGS := -ldflags=all=

ifdef STATIC_BUILD
  CGO_ENABLED=0
  export CGO_ENABLED
  EXTRA_BUILDFLAGS=-installsuffix cgo
  EXTRA_LDFLAGS=-s -w
endif


# Set compiler flags to allow binary debugging
ifdef DEBUGGABLE
  GCFLAGS=-gcflags "all=-N -l"
endif

.PHONY: kops-install # Install kops to local $GOPATH/bin
kops-install: gobindata-tool ${BINDATA_TARGETS}
	go install ${GCFLAGS} ${EXTRA_BUILDFLAGS} ${LDFLAGS}"-X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA} ${EXTRA_LDFLAGS}" k8s.io/kops/cmd/kops/

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
clean:
	if test -e ${BUILD}; then rm -rfv ${BUILD}; fi
	bazel clean
	rm -rf tests/integration/update_cluster/*/.terraform

.PHONY: kops
kops: ${KOPS}

.PHONY: ${KOPS}
${KOPS}: ${BINDATA_TARGETS}
	go build ${GCFLAGS} ${EXTRA_BUILDFLAGS} ${LDFLAGS}"-X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA} ${EXTRA_LDFLAGS}" -o $@ k8s.io/kops/cmd/kops/

${GOBINDATA}:
	mkdir -p ${LOCAL}
	go build ${GCFLAGS} ${EXTRA_BUILDFLAGS} ${LDFLAGS}"${EXTRA_LDFLAGS}" -o $@ k8s.io/kops/vendor/github.com/go-bindata/go-bindata/go-bindata

.PHONY: gobindata-tool
gobindata-tool: ${GOBINDATA}

.PHONY: kops-gobindata
kops-gobindata: gobindata-tool ${BINDATA_TARGETS}

.PHONY: update-bindata
update-bindata:
	cd ${KOPS_ROOT}; ${GOBINDATA} -o ${BINDATA_TARGETS} -pkg models -nometadata -nocompress -ignore="\\.DS_Store" -ignore="bindata\\.go" -ignore="vfs\\.go" -prefix upup/models/ upup/models/...
	GO111MODULE=on go run golang.org/x/tools/cmd/goimports -w -v ${BINDATA_TARGETS}
	gofmt -w -s ${BINDATA_TARGETS}

UPUP_MODELS_BINDATA_SOURCES:=$(shell find upup/models/ | egrep -v "upup/models/bindata.go")
upup/models/bindata.go: ${GOBINDATA} ${UPUP_MODELS_BINDATA_SOURCES}
	make update-bindata

# Build in a docker container with golang 1.X
# Used to test we have not broken 1.X
.PHONY: check-builds-in-go112
check-builds-in-go112:
	docker run -e GO111MODULE=on -e EXTRA_BUILDFLAGS=-mod=vendor -v ${KOPS_ROOT}:/go/src/k8s.io/kops golang:1.12 make -C /go/src/k8s.io/kops all

.PHONY: check-builds-in-go113
check-builds-in-go113:
	docker run -e EXTRA_BUILDFLAGS=-mod=vendor -v ${KOPS_ROOT}:/go/src/k8s.io/kops golang:1.13 make -C /go/src/k8s.io/kops all

.PHONY: codegen
codegen: kops-gobindata
	go install k8s.io/kops/upup/tools/generators/...
	PATH="${GOPATH_1ST}/bin:${PATH}" go generate k8s.io/kops/upup/pkg/fi/cloudup/awstasks
	PATH="${GOPATH_1ST}/bin:${PATH}" go generate k8s.io/kops/upup/pkg/fi/cloudup/gcetasks
	PATH="${GOPATH_1ST}/bin:${PATH}" go generate k8s.io/kops/upup/pkg/fi/cloudup/dotasks
	PATH="${GOPATH_1ST}/bin:${PATH}" go generate k8s.io/kops/upup/pkg/fi/cloudup/openstacktasks
	PATH="${GOPATH_1ST}/bin:${PATH}" go generate k8s.io/kops/upup/pkg/fi/cloudup/alitasks
	PATH="${GOPATH_1ST}/bin:${PATH}" go generate k8s.io/kops/upup/pkg/fi/cloudup/spotinsttasks
	PATH="${GOPATH_1ST}/bin:${PATH}" go generate k8s.io/kops/upup/pkg/fi/assettasks
	PATH="${GOPATH_1ST}/bin:${PATH}" go generate k8s.io/kops/upup/pkg/fi/fitasks

.PHONY: protobuf
protobuf:
	cd ${GOPATH_1ST}/src; protoc --gofast_out=. k8s.io/kops/protokube/pkg/gossip/mesh/mesh.proto

.PHONY: hooks
hooks: # Install Git hooks
	cp hack/pre-commit.sh .git/hooks/pre-commit

.PHONY: test
test: ${BINDATA_TARGETS}  # Run tests locally
	go test -v ./...

.PHONY: ${DIST}/linux/amd64/nodeup
${DIST}/linux/amd64/nodeup: ${BINDATA_TARGETS}
	mkdir -p ${DIST}
	GOOS=linux GOARCH=amd64 go build ${GCFLAGS} -a ${EXTRA_BUILDFLAGS} -o $@ ${LDFLAGS}"${EXTRA_LDFLAGS} -X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA}" k8s.io/kops/cmd/nodeup

.PHONY: crossbuild-nodeup
crossbuild-nodeup: ${DIST}/linux/amd64/nodeup

.PHONY: crossbuild-nodeup-in-docker
crossbuild-nodeup-in-docker:
	docker pull golang:${GOVERSION} # Keep golang image up to date
	docker run --name=nodeup-build-${UNIQUE} -e STATIC_BUILD=yes -e VERSION=${VERSION} -v ${KOPS_ROOT}:/go/src/k8s.io/kops golang:${GOVERSION} make -C /go/src/k8s.io/kops/ crossbuild-nodeup
	docker start nodeup-build-${UNIQUE}
	docker exec nodeup-build-${UNIQUE} chown -R ${UID}:${GID} /go/src/k8s.io/kops/.build
	docker cp nodeup-build-${UNIQUE}:/go/src/k8s.io/kops/.build .
	docker kill nodeup-build-${UNIQUE}
	docker rm nodeup-build-${UNIQUE}

.PHONY: ${DIST}/darwin/amd64/kops
${DIST}/darwin/amd64/kops: ${BINDATA_TARGETS}
	mkdir -p ${DIST}
	GOOS=darwin GOARCH=amd64 go build ${GCFLAGS} -a ${EXTRA_BUILDFLAGS} -o $@ ${LDFLAGS}"${EXTRA_LDFLAGS} -X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA}" k8s.io/kops/cmd/kops

.PHONY: ${DIST}/linux/amd64/kops
${DIST}/linux/amd64/kops: ${BINDATA_TARGETS}
	mkdir -p ${DIST}
	GOOS=linux GOARCH=amd64 go build ${GCFLAGS} -a ${EXTRA_BUILDFLAGS} -o $@ ${LDFLAGS}"${EXTRA_LDFLAGS} -X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA}" k8s.io/kops/cmd/kops

.PHONY: ${DIST}/windows/amd64/kops.exe
${DIST}/windows/amd64/kops.exe: ${BINDATA_TARGETS}
	mkdir -p ${DIST}
	GOOS=windows GOARCH=amd64 go build ${GCFLAGS} -a ${EXTRA_BUILDFLAGS} -o $@ ${LDFLAGS}"${EXTRA_LDFLAGS} -X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA}" k8s.io/kops/cmd/kops


.PHONY: crossbuild
crossbuild: ${DIST}/windows/amd64/kops.exe ${DIST}/darwin/amd64/kops ${DIST}/linux/amd64/kops

.PHONY: crossbuild-in-docker
crossbuild-in-docker:
	docker pull golang:${GOVERSION} # Keep golang image up to date
	docker run --name=kops-build-${UNIQUE} -e STATIC_BUILD=yes -e VERSION=${VERSION} -v ${KOPS_ROOT}:/go/src/k8s.io/kops golang:${GOVERSION} make -C /go/src/k8s.io/kops/ crossbuild
	docker start kops-build-${UNIQUE}
	docker exec kops-build-${UNIQUE} chown -R ${UID}:${GID} /go/src/k8s.io/kops/.build
	docker cp kops-build-${UNIQUE}:/go/src/k8s.io/kops/.build .
	docker kill kops-build-${UNIQUE}
	docker rm kops-build-${UNIQUE}

.PHONY: kops-dist
kops-dist: crossbuild-in-docker
	mkdir -p ${DIST}
	tools/sha1 ${DIST}/darwin/amd64/kops ${DIST}/darwin/amd64/kops.sha1
	tools/sha256 ${DIST}/darwin/amd64/kops ${DIST}/darwin/amd64/kops.sha256
	tools/sha1 ${DIST}/linux/amd64/kops ${DIST}/linux/amd64/kops.sha1
	tools/sha256 ${DIST}/linux/amd64/kops ${DIST}/linux/amd64/kops.sha256
	tools/sha1 ${DIST}/windows/amd64/kops.exe ${DIST}/windows/amd64/kops.exe.sha1
	tools/sha256 ${DIST}/windows/amd64/kops.exe ${DIST}/windows/amd64/kops.exe.sha256

.PHONY: version-dist
version-dist: nodeup-dist kops-dist protokube-export utils-dist
	rm -rf ${UPLOAD}
	mkdir -p ${UPLOAD}/kops/${VERSION}/linux/amd64/
	mkdir -p ${UPLOAD}/kops/${VERSION}/darwin/amd64/
	mkdir -p ${UPLOAD}/kops/${VERSION}/images/
	mkdir -p ${UPLOAD}/utils/${VERSION}/linux/amd64/
	cp ${DIST}/nodeup ${UPLOAD}/kops/${VERSION}/linux/amd64/nodeup
	cp ${DIST}/nodeup.sha1 ${UPLOAD}/kops/${VERSION}/linux/amd64/nodeup.sha1
	cp ${DIST}/nodeup.sha256 ${UPLOAD}/kops/${VERSION}/linux/amd64/nodeup.sha256
	cp ${IMAGES}/protokube.tar.gz ${UPLOAD}/kops/${VERSION}/images/protokube.tar.gz
	cp ${IMAGES}/protokube.tar.gz.sha1 ${UPLOAD}/kops/${VERSION}/images/protokube.tar.gz.sha1
	cp ${IMAGES}/protokube.tar.gz.sha256 ${UPLOAD}/kops/${VERSION}/images/protokube.tar.gz.sha256
	cp ${DIST}/linux/amd64/kops ${UPLOAD}/kops/${VERSION}/linux/amd64/kops
	cp ${DIST}/linux/amd64/kops.sha1 ${UPLOAD}/kops/${VERSION}/linux/amd64/kops.sha1
	cp ${DIST}/linux/amd64/kops.sha256 ${UPLOAD}/kops/${VERSION}/linux/amd64/kops.sha256
	cp ${DIST}/darwin/amd64/kops ${UPLOAD}/kops/${VERSION}/darwin/amd64/kops
	cp ${DIST}/darwin/amd64/kops.sha1 ${UPLOAD}/kops/${VERSION}/darwin/amd64/kops.sha1
	cp ${DIST}/darwin/amd64/kops.sha256 ${UPLOAD}/kops/${VERSION}/darwin/amd64/kops.sha256
	cp ${DIST}/linux/amd64/utils.tar.gz ${UPLOAD}/kops/${VERSION}/linux/amd64/utils.tar.gz
	cp ${DIST}/linux/amd64/utils.tar.gz.sha1 ${UPLOAD}/kops/${VERSION}/linux/amd64/utils.tar.gz.sha1
	cp ${DIST}/linux/amd64/utils.tar.gz.sha256 ${UPLOAD}/kops/${VERSION}/linux/amd64/utils.tar.gz.sha256

.PHONY: vsphere-version-dist
vsphere-version-dist: nodeup-dist protokube-export
	rm -rf ${UPLOAD}
	mkdir -p ${UPLOAD}/kops/${VERSION}/linux/amd64/
	mkdir -p ${UPLOAD}/kops/${VERSION}/darwin/amd64/
	mkdir -p ${UPLOAD}/kops/${VERSION}/images/
	mkdir -p ${UPLOAD}/utils/${VERSION}/linux/amd64/
	cp ${DIST}/nodeup ${UPLOAD}/kops/${VERSION}/linux/amd64/nodeup
	cp ${DIST}/nodeup.sha1 ${UPLOAD}/kops/${VERSION}/linux/amd64/nodeup.sha1
	cp ${DIST}/nodeup.sha256 ${UPLOAD}/kops/${VERSION}/linux/amd64/nodeup.sha256
	cp ${IMAGES}/protokube.tar.gz ${UPLOAD}/kops/${VERSION}/images/protokube.tar.gz
	cp ${IMAGES}/protokube.tar.gz.sha1 ${UPLOAD}/kops/${VERSION}/images/protokube.tar.gz.sha1
	cp ${IMAGES}/protokube.tar.gz.sha256 ${UPLOAD}/kops/${VERSION}/images/protokube.tar.gz.sha256
	scp -r .build/dist/nodeup* ${TARGET}:${TARGET_PATH}/nodeup
	scp -r .build/dist/images/protokube.tar.gz* ${TARGET}:${TARGET_PATH}/protokube/
	make kops-dist
	cp ${DIST}/linux/amd64/kops ${UPLOAD}/kops/${VERSION}/linux/amd64/kops
	cp ${DIST}/linux/amd64/kops.sha1 ${UPLOAD}/kops/${VERSION}/linux/amd64/kops.sha1
	cp ${DIST}/linux/amd64/kops.sha256 ${UPLOAD}/kops/${VERSION}/linux/amd64/kops.sha256
	cp ${DIST}/darwin/amd64/kops ${UPLOAD}/kops/${VERSION}/darwin/amd64/kops
	cp ${DIST}/darwin/amd64/kops.sha1 ${UPLOAD}/kops/${VERSION}/darwin/amd64/kops.sha1
	cp ${DIST}/darwin/amd64/kops.sha256 ${UPLOAD}/kops/${VERSION}/darwin/amd64/kops.sha256
	cp ${DIST}/windows/amd64/kops.exe ${UPLOAD}/kops/${VERSION}/windows/amd64/kops.exe
	cp ${DIST}/windows/amd64/kops.exe.sha1 ${UPLOAD}/kops/${VERSION}/windows/amd64/kops.exe.sha1
	cp ${DIST}/windows/amd64/kops.exe.sha256 ${UPLOAD}/kops/${VERSION}/windows/amd64/kops.exe.sha256

.PHONY: upload
upload: version-dist # Upload kops to S3
	aws s3 sync --acl public-read ${UPLOAD}/ ${S3_BUCKET}

# oss-upload builds kops and uploads to OSS
.PHONY: oss-upload
oss-upload: version-dist
	@echo "== Uploading kops =="
	aliyun oss cp --acl public-read -r -f --include "*" ${UPLOAD}/ ${OSS_BUCKET}

# gcs-upload builds kops and uploads to GCS
.PHONY: gcs-upload
gcs-upload: bazel-version-dist
	@echo "== Uploading kops =="
	gsutil -h "Cache-Control:private, max-age=0, no-transform" -m cp -n -r ${BAZELUPLOAD}/kops/* ${GCS_LOCATION}

# gcs-publish-ci is the entry point for CI testing
# In CI testing, always upload the CI version.
.PHONY: gcs-publish-ci
gcs-publish-ci: VERSION := ${KOPS_CI_VERSION}+${GITSHA}
gcs-publish-ci: PROTOKUBE_TAG := $(subst +,-,${VERSION})
gcs-publish-ci: gcs-upload
	echo "VERSION: ${VERSION}"
	echo "PROTOKUBE_TAG: ${PROTOKUBE_TAG}"
	echo "${GCS_URL}/${VERSION}" > ${BAZELUPLOAD}/${LATEST_FILE}
	gsutil -h "Cache-Control:private, max-age=0, no-transform" cp ${BAZELUPLOAD}/${LATEST_FILE} ${GCS_LOCATION}

.PHONY: gen-cli-docs
gen-cli-docs: ${KOPS} # Regenerate CLI docs
	KOPS_STATE_STORE= \
	KOPS_FEATURE_FLAGS= \
	${KOPS} genhelpdocs --out docs/cli

.PHONY: push
# Will always push a linux-based build up to the server
push: crossbuild-nodeup
	scp -C .build/dist/linux/amd64/nodeup  ${TARGET}:/tmp/

.PHONY: push-gce-dry
push-gce-dry: push
	ssh ${TARGET} sudo /tmp/nodeup --conf=metadata://gce/config --dryrun --v=8

.PHONY: push-gce-dry
push-aws-dry: push
	ssh ${TARGET} sudo /tmp/nodeup --conf=/opt/kops/conf/kube_env.yaml --dryrun --v=8

.PHONY: push-gce-run
push-gce-run: push
	ssh ${TARGET} sudo cp /tmp/nodeup /var/lib/toolbox/kubernetes-install/nodeup
	ssh ${TARGET} sudo /var/lib/toolbox/kubernetes-install/nodeup --conf=/var/lib/toolbox/kubernetes-install/kube_env.yaml --v=8

# -t is for CentOS http://unix.stackexchange.com/questions/122616/why-do-i-need-a-tty-to-run-sudo-if-i-can-sudo-without-a-password
.PHONY: push-aws-run
push-aws-run: push
	ssh -t ${TARGET} sudo /tmp/nodeup --conf=/opt/kops/conf/kube_env.yaml --v=8

.PHONY: ${PROTOKUBE}
${PROTOKUBE}:
	go build ${GCFLAGS} ${EXTRA_BUILDFLAGS} -o $@ -tags 'peer_name_alternative peer_name_hash' k8s.io/kops/protokube/cmd/protokube

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
	gzip --no-name --force --best ${IMAGES}/protokube.tar
	tools/sha1 ${IMAGES}/protokube.tar.gz ${IMAGES}/protokube.tar.gz.sha1
	tools/sha256 ${IMAGES}/protokube.tar.gz ${IMAGES}/protokube.tar.gz.sha256

# protokube-push is no longer used (we upload a docker image tar file to S3 instead),
# but we're keeping it around in case it is useful for development etc
.PHONY: protokube-push
protokube-push: protokube-image
	docker tag protokube:${PROTOKUBE_TAG} ${DOCKER_REGISTRY}/protokube:${PROTOKUBE_TAG}
	docker push ${DOCKER_REGISTRY}/protokube:${PROTOKUBE_TAG}

.PHONY: nodeup
nodeup: ${NODEUP}

.PHONY: ${NODEUP}
${NODEUP}: ${BINDATA_TARGETS}
	go build ${GCFLAGS} ${EXTRA_BUILDFLAGS} ${LDFLAGS}"${EXTRA_LDFLAGS} -X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA}" -o $@ k8s.io/kops/cmd/nodeup

.PHONY: nodeup-dist
nodeup-dist:
	mkdir -p ${DIST}
	docker pull golang:${GOVERSION} # Keep golang image up to date
	docker run --name=nodeup-build-${UNIQUE} -e STATIC_BUILD=yes -e VERSION=${VERSION} -v ${KOPS_ROOT}:/go/src/k8s.io/kops golang:${GOVERSION} make -C /go/src/k8s.io/kops/ nodeup
	docker start nodeup-build-${UNIQUE}
	docker exec nodeup-build-${UNIQUE} chown -R ${UID}:${GID} /go/src/k8s.io/kops/.build
	docker cp nodeup-build-${UNIQUE}:/go/src/k8s.io/kops/.build/local/nodeup .build/dist/
	tools/sha1 .build/dist/nodeup .build/dist/nodeup.sha1
	tools/sha256 .build/dist/nodeup .build/dist/nodeup.sha256

.PHONY: bazel-crossbuild-dns-controller
bazel-crossbuild-dns-controller:
	bazel build ${BAZEL_CONFIG} --features=pure --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //dns-controller/...


.PHONY: dns-controller-push
dns-controller-push:
	DOCKER_REGISTRY=${DOCKER_REGISTRY} DOCKER_IMAGE_PREFIX=${DOCKER_IMAGE_PREFIX} DNS_CONTROLLER_TAG=${DNS_CONTROLLER_TAG} bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //dns-controller/cmd/dns-controller:push-image

# --------------------------------------------------
# static utils

.PHONY: utils-dist
utils-dist:
	docker build -t utils-builder images/utils-builder
	mkdir -p ${DIST}/linux/amd64/
	docker run -v `pwd`/.build/dist/linux/amd64/:/dist utils-builder /extract.sh

.PHONY: bazel-utils-dist
bazel-utils-dist:
	bazel build ${BAZEL_CONFIG} //images/utils-builder:utils

# --------------------------------------------------
# development targets

.PHONY: gomod-prereqs
gomod-prereqs:
	(which bazel > /dev/null) || (echo "gomod requires that bazel is installed"; exit 1)

.PHONY: dep-ensure
dep-ensure:
	echo "'make dep-ensure' has been replaced by 'make gomod'"
	exit 1

.PHONY: gomod
gomod: gomod-prereqs
	GO111MODULE=on go mod vendor
	# Switch weavemesh to use peer_name_hash - bazel rule-go doesn't support build tags yet
	rm vendor/github.com/weaveworks/mesh/peer_name_mac.go
	sed -i -e 's/peer_name_hash/!peer_name_mac/g' vendor/github.com/weaveworks/mesh/peer_name_hash.go
	# Remove all bazel build files that were vendored and regenerate (we assume they are go-gettable)
	find vendor/ -name "BUILD" -delete
	find vendor/ -name "BUILD.bazel" -delete
	make gazelle


.PHONY: gofmt
gofmt:
	find $(KOPS_ROOT) -name "*.go" | grep -v vendor | xargs bazel run //:gofmt -- -w -s

.PHONY: goimports
goimports:
	hack/update-goimports

.PHONY: verify-goimports
verify-goimports:
	hack/verify-goimports

.PHONY: govet
govet: ${BINDATA_TARGETS}
	go vet ./...

# --------------------------------------------------
# Continuous integration targets

# verify is ran by the pull-kops-verify prow job
.PHONY: verify
verify: travis-ci verify-gofmt

.PHONY: verify-boilerplate
verify-boilerplate:
	hack/verify-boilerplate.sh

.PHONY: verify-gofmt
verify-gofmt:
	hack/verify-gofmt.sh

.PHONY: verify-gomod
verify-gomod:
	hack/verify-gomod

.PHONY: verify-packages
verify-packages: ${BINDATA_TARGETS}
	hack/verify-packages.sh

# find release notes, remove PR titles and output the rest to .build, then run misspell on all files
.PHONY: verify-misspelling
verify-misspelling:
	hack/verify-spelling.sh

.PHONY: verify-gendocs
verify-gendocs: ${KOPS}
	@TMP_DOCS="$$(mktemp -d)"; \
	'${KOPS}' genhelpdocs --out "$$TMP_DOCS"; \
	\
	if ! diff -r "$$TMP_DOCS" '${KOPS_ROOT}/docs/cli'; then \
	     echo "FAIL: make verify-gendocs failed, as the generated markdown docs are out of date." 1>&2; \
	     echo "FAIL: Please run the following command: make gen-cli-docs." 1>&2; \
	     exit 1; \
	fi
	@echo "cli docs up-to-date"

.PHONY: verify-bazel
verify-bazel:
	hack/verify-bazel.sh

.PHONY: verify-staticcheck
verify-staticcheck: ${BINDATA_TARGETS}
	hack/verify-staticcheck.sh

.PHONY: verify-shellcheck
verify-shellcheck:
	${KOPS_ROOT}/hack/verify-shellcheck.sh

.PHONY: verify-terraform
verify-terraform:
	./hack/verify-terraform.sh

.PHONY: verify-bindata
verify-bindata:
	./hack/verify-bindata.sh

# ci target is for developers, it aims to cover all the CI jobs
# verify-gendocs will call kops target
# verify-package has to be after verify-gendocs, because with .gitignore for federation bindata
# it bombs in travis. verify-gendocs generates the bindata file.
.PHONY: ci
ci: govet verify-gofmt verify-generate verify-gomod verify-goimports verify-boilerplate verify-bazel verify-misspelling verify-shellcheck verify-staticcheck verify-terraform verify-bindata nodeup examples test | verify-gendocs verify-packages verify-apimachinery
	echo "Done!"

# travis-ci is the target that travis-ci calls
# we skip tasks that rely on bazel and are covered by other jobs
# verify-gofmt: uses bazel, covered by pull-kops-verify
# govet needs to be after verify-goimports because it generates bindata.go
.PHONY: travis-ci
travis-ci: verify-generate verify-gomod verify-goimports govet verify-boilerplate verify-bazel verify-misspelling verify-shellcheck verify-bindata | verify-gendocs verify-packages verify-apimachinery
	echo "Done!"

.PHONY: pr
pr:
	@echo "Test passed!"
	@echo "Feel free to open your pr at https://github.com/kubernetes/kops/compare"

# --------------------------------------------------
# channel tool

.PHONY: channels
channels: ${CHANNELS}

.PHONY: ${CHANNELS}
${CHANNELS}:
	go build ${GCFLAGS} ${EXTRA_BUILDFLAGS} -o $@ ${LDFLAGS}"-X k8s.io/kops.Version=${VERSION} ${EXTRA_LDFLAGS}" k8s.io/kops/channels/cmd/channels

# --------------------------------------------------
# release tasks

.PHONY: release-tag
release-tag:
	git tag v${KOPS_RELEASE_VERSION}

.PHONY: release-github
release-github:
	shipbot -tag v${KOPS_RELEASE_VERSION} -config .shipbot.yaml

# --------------------------------------------------
# API / embedding examples

.PHONY: examples
examples: ${BINDATA_TARGETS} # Install kops API example
	go install k8s.io/kops/examples/kops-api-example/...

# -----------------------------------------------------
# api machinery regenerate

.PHONY: apimachinery
apimachinery: apimachinery-codegen goimports

.PHONY: apimachinery-codegen
apimachinery-codegen:
	sh -c hack/make-apimachinery.sh
	${GOPATH}/bin/conversion-gen ${API_OPTIONS} --skip-unsafe=true --input-dirs k8s.io/kops/pkg/apis/kops/v1alpha2 --v=0  --output-file-base=zz_generated.conversion \
		 --go-header-file "hack/boilerplate/boilerplate.go.txt"
	${GOPATH}/bin/deepcopy-gen ${API_OPTIONS} --input-dirs k8s.io/kops/pkg/apis/kops --v=0  --output-file-base=zz_generated.deepcopy \
		 --go-header-file "hack/boilerplate/boilerplate.go.txt"
	${GOPATH}/bin/deepcopy-gen ${API_OPTIONS} --input-dirs k8s.io/kops/pkg/apis/kops/v1alpha2 --v=0  --output-file-base=zz_generated.deepcopy \
		 --go-header-file "hack/boilerplate/boilerplate.go.txt"
	${GOPATH}/bin/defaulter-gen ${API_OPTIONS} --input-dirs k8s.io/kops/pkg/apis/kops/v1alpha2 --v=0  --output-file-base=zz_generated.defaults \
		 --go-header-file "hack/boilerplate/boilerplate.go.txt"
	#go install github.com/ugorji/go/codec/codecgen
	# codecgen works only if invoked from directory where the file is located.
	#cd pkg/apis/kops/ && ~/k8s/bin/codecgen -d 1234 -o types.generated.go instancegroup.go cluster.go
	${GOPATH}/bin/client-gen  ${API_OPTIONS} --input-base k8s.io/kops/pkg/apis/ --input="kops/,kops/v1alpha2" --clientset-path k8s.io/kops/pkg/client/clientset_generated/ \
		 --go-header-file "hack/boilerplate/boilerplate.go.txt"
	${GOPATH}/bin/client-gen  ${API_OPTIONS} --clientset-name="clientset" --input-base k8s.io/kops/pkg/apis/ --input="kops/,kops/v1alpha2" --clientset-path k8s.io/kops/pkg/client/clientset_generated/ \
		 --go-header-file "hack/boilerplate/boilerplate.go.txt"

.PHONY: verify-apimachinery
verify-apimachinery:
	hack/verify-apimachinery.sh

.PHONY: verify-generate
verify-generate:
	hack/verify-generate.sh

# -----------------------------------------------------
# bazel targets

.PHONY: bazel-test
bazel-test:
	bazel ${BAZEL_OPTIONS} test ${BAZEL_CONFIG} --test_output=errors -- //... -//vendor/...

.PHONY: bazel-build
bazel-build:
	bazel build ${BAZEL_CONFIG} --features=pure //cmd/... //pkg/... //channels/... //nodeup/... //protokube/... //dns-controller/... //util/...

.PHONY: bazel-build-cli
bazel-build-cli:
	bazel build ${BAZEL_CONFIG} --features=pure //cmd/kops/...

.PHONY: bazel-crossbuild-kops
bazel-crossbuild-kops:
	bazel build ${BAZEL_CONFIG} --features=pure --platforms=@io_bazel_rules_go//go/toolchain:darwin_amd64 //cmd/kops/...
	bazel build ${BAZEL_CONFIG} --features=pure --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //cmd/kops/...
	bazel build ${BAZEL_CONFIG} --features=pure --platforms=@io_bazel_rules_go//go/toolchain:windows_amd64 //cmd/kops/...

.PHONY: bazel-crossbuild-nodeup
bazel-crossbuild-nodeup:
	bazel build ${BAZEL_CONFIG} --features=pure --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //cmd/nodeup/...

.PHONY: bazel-crossbuild-protokube
bazel-crossbuild-protokube:
	bazel build ${BAZEL_CONFIG} --features=pure --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //protokube/...

.PHONY: bazel-crossbuild-protokube-image
bazel-crossbuild-protokube-image:
	bazel build ${BAZEL_CONFIG} --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //images:protokube.tar

.PHONY: bazel-crossbuild-kube-discovery-image
bazel-crossbuild-kube-discovery-image:
	bazel build ${BAZEL_CONFIG} --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //images:kube-discovery.tar

.PHONY: bazel-crossbuild-node-authorizer-image
bazel-crossbuild-node-authorizer-image:
	bazel build ${BAZEL_CONFIG} --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //images:node-authorizer.tar

.PHONY: bazel-push
# Will always push a linux-based build up to the server
bazel-push: bazel-crossbuild-nodeup
	ssh ${TARGET} touch /tmp/nodeup
	ssh ${TARGET} chmod +w /tmp/nodeup
	scp -C bazel-bin/cmd/nodeup/linux_amd64_pure_stripped/nodeup  ${TARGET}:/tmp/

.PHONY: bazel-push-gce-run
bazel-push-gce-run: bazel-push
	ssh ${TARGET} sudo cp /tmp/nodeup /var/lib/toolbox/kubernetes-install/nodeup
	ssh ${TARGET} sudo SKIP_PACKAGE_UPDATE=1 /var/lib/toolbox/kubernetes-install/nodeup --conf=/var/lib/toolbox/kubernetes-install/kube_env.yaml --v=8

# -t is for CentOS http://unix.stackexchange.com/questions/122616/why-do-i-need-a-tty-to-run-sudo-if-i-can-sudo-without-a-password
.PHONY: bazel-push-aws-run
bazel-push-aws-run: bazel-push
	ssh ${TARGET} chmod +x /tmp/nodeup
	ssh -t ${TARGET} sudo SKIP_PACKAGE_UPDATE=1 /tmp/nodeup --conf=/opt/kops/conf/kube_env.yaml --v=8

.PHONY: gazelle
gazelle:
	hack/update-bazel.sh

.PHONY: bazel-gazelle
bazel-gazelle: gazelle
	echo "bazel-gazelle is deprecated; please just use 'make gazelle'"

.PHONY: check-markdown-links
check-markdown-links:
	docker run -t -v $$PWD:/tmp \
		-e LC_ALL=C.UTF-8 \
		-e LANG=en_US.UTF-8 \
		-e LANGUAGE=en_US.UTF-8 \
		rubygem/awesome_bot --allow-dupe --allow-redirect \
		$(shell find $$PWD -name "*.md" -mindepth 1 -printf '%P\n' | grep -v vendor | grep -v Changelog.md)

#-----------------------------------------------------------
# kube-discovery

.PHONY: push-kube-discovery
push-kube-discovery:
	bazel run ${BAZEL_CONFIG} //kube-discovery/images:kube-discovery
	docker tag bazel/kube-discovery/images:kube-discovery ${DOCKER_REGISTRY}/kube-discovery:${DOCKER_TAG}
	docker push ${DOCKER_REGISTRY}/kube-discovery:${DOCKER_TAG}

.PHONY: push-node-authorizer
push-node-authorizer:
	bazel run ${BAZEL_CONFIG} //node-authorizer/images:node-authorizer
	docker tag bazel/node-authorizer/images:node-authorizer ${DOCKER_REGISTRY}/node-authorizer:${DOCKER_TAG}
	docker push ${DOCKER_REGISTRY}/node-authorizer:${DOCKER_TAG}

.PHONY: bazel-protokube-export
bazel-protokube-export:
	mkdir -p ${BAZELIMAGES}
	bazel build ${BAZEL_CONFIG} --action_env=PROTOKUBE_TAG=${PROTOKUBE_TAG} --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //images:protokube.tar.gz //images:protokube.tar.gz.sha1  //images:protokube.tar.gz.sha256
	cp -fp bazel-bin/images/protokube.tar.gz ${BAZELIMAGES}/protokube.tar.gz
	cp -fp bazel-bin/images/protokube.tar.gz.sha1 ${BAZELIMAGES}/protokube.tar.gz.sha1
	cp -fp bazel-bin/images/protokube.tar.gz.sha256 ${BAZELIMAGES}/protokube.tar.gz.sha256

.PHONY: bazel-kops-controller-export
bazel-kops-controller-export:
	mkdir -p ${BAZELIMAGES}
	DOCKER_REGISTRY="" DOCKER_IMAGE_PREFIX="kope/" KOPS_CONTROLLER_TAG=${KOPS_CONTROLLER_TAG} bazel build ${BAZEL_CONFIG} --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //cmd/kops-controller:image-bundle.tar.gz //cmd/kops-controller:image-bundle.tar.gz.sha1 //cmd/kops-controller:image-bundle.tar.gz.sha256
	cp -fp bazel-bin/cmd/kops-controller/image-bundle.tar.gz ${BAZELIMAGES}/kops-controller.tar.gz
	cp -fp bazel-bin/cmd/kops-controller/image-bundle.tar.gz.sha1 ${BAZELIMAGES}/kops-controller.tar.gz.sha1
	cp -fp bazel-bin/cmd/kops-controller/image-bundle.tar.gz.sha256 ${BAZELIMAGES}/kops-controller.tar.gz.sha256

.PHONY: bazel-dns-controller-export
bazel-dns-controller-export:
	mkdir -p ${BAZELIMAGES}
	DOCKER_REGISTRY="" DOCKER_IMAGE_PREFIX="kope/" DNS_CONTROLLER_TAG=${DNS_CONTROLLER_TAG} bazel build ${BAZEL_CONFIG} --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //dns-controller/cmd/dns-controller:image-bundle.tar.gz //dns-controller/cmd/dns-controller:image-bundle.tar.gz.sha1 //dns-controller/cmd/dns-controller:image-bundle.tar.gz.sha256
	cp -fp bazel-bin/dns-controller/cmd/dns-controller/image-bundle.tar.gz ${BAZELIMAGES}/dns-controller.tar.gz
	cp -fp bazel-bin/dns-controller/cmd/dns-controller/image-bundle.tar.gz.sha1 ${BAZELIMAGES}/dns-controller.tar.gz.sha1
	cp -fp bazel-bin/dns-controller/cmd/dns-controller/image-bundle.tar.gz.sha256 ${BAZELIMAGES}/dns-controller.tar.gz.sha256

.PHONY: bazel-version-dist
bazel-version-dist: bazel-crossbuild-nodeup bazel-crossbuild-kops bazel-kops-controller-export bazel-dns-controller-export bazel-protokube-export bazel-utils-dist
	rm -rf ${BAZELUPLOAD}
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/darwin/amd64/
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/windows/amd64/
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/images/
	mkdir -p ${BAZELUPLOAD}/utils/${VERSION}/linux/amd64/
	cp bazel-bin/cmd/nodeup/linux_amd64_pure_stripped/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup
	tools/sha1 ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup.sha1
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup.sha256
	cp ${BAZELIMAGES}/protokube.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/protokube.tar.gz
	cp ${BAZELIMAGES}/protokube.tar.gz.sha1 ${BAZELUPLOAD}/kops/${VERSION}/images/protokube.tar.gz.sha1
	cp ${BAZELIMAGES}/protokube.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/protokube.tar.gz.sha256
	cp ${BAZELIMAGES}/kops-controller.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/kops-controller.tar.gz
	cp ${BAZELIMAGES}/kops-controller.tar.gz.sha1 ${BAZELUPLOAD}/kops/${VERSION}/images/kops-controller.tar.gz.sha1
	cp ${BAZELIMAGES}/kops-controller.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/kops-controller.tar.gz.sha256
	cp ${BAZELIMAGES}/dns-controller.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/dns-controller.tar.gz
	cp ${BAZELIMAGES}/dns-controller.tar.gz.sha1 ${BAZELUPLOAD}/kops/${VERSION}/images/dns-controller.tar.gz.sha1
	cp ${BAZELIMAGES}/dns-controller.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/dns-controller.tar.gz.sha256
	cp bazel-bin/cmd/kops/linux_amd64_pure_stripped/kops ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/kops
	tools/sha1 ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/kops ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/kops.sha1
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/kops ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/kops.sha256
	cp bazel-bin/cmd/kops/darwin_amd64_pure_stripped/kops ${BAZELUPLOAD}/kops/${VERSION}/darwin/amd64/kops
	tools/sha1 ${BAZELUPLOAD}/kops/${VERSION}/darwin/amd64/kops ${BAZELUPLOAD}/kops/${VERSION}/darwin/amd64/kops.sha1
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/darwin/amd64/kops ${BAZELUPLOAD}/kops/${VERSION}/darwin/amd64/kops.sha256
	cp bazel-bin/cmd/kops/windows_amd64_pure_stripped/kops.exe ${BAZELUPLOAD}/kops/${VERSION}/windows/amd64/kops.exe
	tools/sha1 ${BAZELUPLOAD}/kops/${VERSION}/windows/amd64/kops.exe ${BAZELUPLOAD}/kops/${VERSION}/windows/amd64/kops.exe.sha1
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/windows/amd64/kops.exe ${BAZELUPLOAD}/kops/${VERSION}/windows/amd64/kops.exe.sha256
	cp bazel-bin/images/utils-builder/utils.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/utils.tar.gz
	tools/sha1 ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/utils.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/utils.tar.gz.sha1
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/utils.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/utils.tar.gz.sha256
	cp -fr ${BAZELUPLOAD}/kops/${VERSION}/* ${BAZELDIST}/

.PHONY: bazel-upload
bazel-upload: bazel-version-dist # Upload kops to S3
	aws s3 sync --acl public-read ${BAZELUPLOAD}/ ${S3_BUCKET}

# prow-postsubmit is run by the prow postsubmit job
# It uploads a build to a staging directory, which in theory we can publish as a release
.PHONY: prow-postsubmit
prow-postsubmit: bazel-version-dist
	${UPLOAD_CMD} ${BAZELUPLOAD}/kops/${VERSION}/ ${UPLOAD_DEST}/${VERSION}/

#-----------------------------------------------------------
# static html documentation

.PHONY: live-docs
live-docs:
	docker build -t kops/mkdocs images/mkdocs
	docker run --rm -it -p 3000:3000 -v ${PWD}:/docs kops/mkdocs

.PHONY: build-docs
build-docs:
	docker build --pull -t kops/mkdocs images/mkdocs
	docker run --rm -v ${PWD}:/docs kops/mkdocs build

.PHONY: build-docs-netlify
build-docs-netlify:
	pip install -r ${KOPS_ROOT}/images/mkdocs/requirements.txt
	mkdocs build

# Update machine_types.go
.PHONY: update-machine-types
update-machine-types:
	go build -o hack/machine_types/machine_types  ${KOPS_ROOT}/hack/machine_types/
	hack/machine_types/machine_types --out upup/pkg/fi/cloudup/awsup/machine_types.go
	go fmt upup/pkg/fi/cloudup/awsup/machine_types.go

#-----------------------------------------------------------
# development targets

# dev-upload-nodeup uploads nodeup to GCS
.PHONY: dev-upload-nodeup
dev-upload-nodeup: bazel-crossbuild-nodeup
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/
	cp -fp bazel-bin/cmd/nodeup/linux_amd64_pure_stripped/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup
	tools/sha1 ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup.sha1
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup.sha256
	${UPLOAD_CMD} ${BAZELUPLOAD}/ ${UPLOAD_DEST}

# dev-upload-protokube uploads protokube to GCS
.PHONY: dev-upload-protokube
dev-upload-protokube: bazel-protokube-export # Upload kops to GCS
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/images/
	cp -fp ${BAZELIMAGES}/protokube.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/protokube.tar.gz
	cp -fp ${BAZELIMAGES}/protokube.tar.gz.sha1 ${BAZELUPLOAD}/kops/${VERSION}/images/protokube.tar.gz.sha1
	cp -fp ${BAZELIMAGES}/protokube.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/protokube.tar.gz.sha256
	${UPLOAD_CMD} ${BAZELUPLOAD}/ ${UPLOAD_DEST}

# dev-upload-kops-controller uploads kops-controller to GCS
.PHONY: dev-upload-kops-controller
dev-upload-kops-controller: bazel-kops-controller-export # Upload kops to GCS
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/images/
	cp -fp ${BAZELIMAGES}/kops-controller.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/kops-controller.tar.gz
	cp -fp ${BAZELIMAGES}/kops-controller.tar.gz.sha1 ${BAZELUPLOAD}/kops/${VERSION}/images/kops-controller.tar.gz.sha1
	cp -fp ${BAZELIMAGES}/kops-controller.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/kops-controller.tar.gz.sha256
	${UPLOAD_CMD} ${BAZELUPLOAD}/ ${UPLOAD_DEST}

# dev-upload-dns-controller uploads dns-controller to GCS
.PHONY: dev-upload-dns-controller
dev-upload-dns-controller: bazel-dns-controller-export # Upload kops to GCS
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/images/
	cp -fp ${BAZELIMAGES}/dns-controller.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/dns-controller.tar.gz
	cp -fp ${BAZELIMAGES}/dns-controller.tar.gz.sha1 ${BAZELUPLOAD}/kops/${VERSION}/images/dns-controller.tar.gz.sha1
	cp -fp ${BAZELIMAGES}/dns-controller.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/dns-controller.tar.gz.sha256
	${UPLOAD_CMD} ${BAZELUPLOAD}/ ${UPLOAD_DEST}

# dev-copy-utils copies utils from a recent release
# We don't currently have a bazel build for them, and the build is pretty slow, but they change rarely.
.PHONE: dev-copy-utils
dev-copy-utils:
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/
	cd ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/; wget -N https://kubeupv2.s3.amazonaws.com/kops/1.15.0-alpha.1/linux/amd64/utils.tar.gz
	cd ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/; wget -N https://kubeupv2.s3.amazonaws.com/kops/1.15.0-alpha.1/linux/amd64/utils.tar.gz.sha1
	cd ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/; wget -N https://kubeupv2.s3.amazonaws.com/kops/1.15.0-alpha.1/linux/amd64/utils.tar.gz.sha256
	${UPLOAD_CMD} ${BAZELUPLOAD}/ ${UPLOAD_DEST}

# dev-upload does a faster build and uploads to GCS / S3
# It copies utils instead of building it
.PHONY: dev-upload
dev-upload: dev-upload-nodeup dev-upload-protokube dev-upload-dns-controller dev-upload-kops-controller dev-copy-utils
	echo "Done"

.PHONY: crds
crds:
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go crd paths=k8s.io/kops/pkg/apis/kops/v1alpha2 output:dir=k8s/crds/ crd:crdVersions=v1

#------------------------------------------------------
# kops-controller

.PHONY: kops-controller-push
kops-controller-push:
	DOCKER_REGISTRY=${DOCKER_REGISTRY} DOCKER_IMAGE_PREFIX=${DOCKER_IMAGE_PREFIX} KOPS_CONTROLLER_TAG=${KOPS_CONTROLLER_TAG} bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //cmd/kops-controller:push-image
