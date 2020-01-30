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


DOCKER_REGISTRY?=gcr.io/must-override
S3_BUCKET?=s3://must-override/
UPLOAD_DEST?=$(S3_BUCKET)
GCS_LOCATION?=gs://must-override
GCS_URL=$(GCS_LOCATION:gs://%=https://storage.googleapis.com/%)
LATEST_FILE?=latest-ci.txt
GOPATH_1ST:=$(shell go env | grep GOPATH | cut -f 2 -d \")
UNIQUE:=$(shell date +%s)
GOVERSION=1.12.9
BUILD=$(GOPATH_1ST)/src/k8s.io/kops/.build
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
BAZELBUILD=$(GOPATH_1ST)/src/k8s.io/kops/.bazelbuild
BAZELDIST=$(BAZELBUILD)/dist
BAZELIMAGES=$(BAZELDIST)/images
BAZELUPLOAD=$(BAZELBUILD)/upload
UID:=$(shell id -u)
GID:=$(shell id -g)
BAZEL_OPTIONS?=
API_OPTIONS?=
GCFLAGS?=

# See http://stackoverflow.com/questions/18136918/how-to-get-current-relative-directory-of-your-makefile
MAKEDIR:=$(strip $(shell dirname "$(realpath $(lastword $(MAKEFILE_LIST)))"))

UPLOAD=$(MAKEDIR)/hack/upload


# Unexport environment variables that can affect tests and are not used in builds
unexport AWS_ACCESS_KEY_ID AWS_REGION AWS_SECRET_ACCESS_KEY AWS_SESSION_TOKEN CNI_VERSION_URL DNS_IGNORE_NS_CHECK DNSCONTROLLER_IMAGE DO_ACCESS_TOKEN GOOGLE_APPLICATION_CREDENTIALS
unexport KOPS_BASE_URL KOPS_CLUSTER_NAME KOPS_RUN_OBSOLETE_VERSION KOPS_STATE_STORE KOPS_STATE_S3_ACL KUBE_API_VERSIONS NODEUP_URL OPENSTACK_CREDENTIAL_FILE PROTOKUBE_IMAGE SKIP_PACKAGE_UPDATE
unexport SKIP_REGION_CHECK S3_ACCESS_KEY_ID S3_ENDPOINT S3_REGION S3_SECRET_ACCESS_KEY VSPHERE_USERNAME VSPHERE_PASSWORD

# Keep in sync with upup/models/cloudup/resources/addons/dns-controller/
DNS_CONTROLLER_TAG=1.15.1

# Keep in sync with logic in get_workspace_status
# TODO: just invoke tools/get_workspace_status.sh?
KOPS_RELEASE_VERSION:=$(shell grep 'KOPS_RELEASE_VERSION\s*=' version.go | awk '{print $$3}' | sed -e 's_"__g')
KOPS_CI_VERSION:=$(shell grep 'KOPS_CI_VERSION\s*=' version.go | awk '{print $$3}' | sed -e 's_"__g')

# kops local location
KOPS                 = ${LOCAL}/kops

# kops source root directory (without trailing /)
KOPS_ROOT           ?= $(patsubst %/,%,$(abspath $(dir $(firstword $(MAKEFILE_LIST)))))

GITSHA := $(shell cd ${GOPATH_1ST}/src/k8s.io/kops; git describe --always)

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
KOPS_SERVER_TAG := $(subst +,-,${VERSION})

# Go exports:

GO15VENDOREXPERIMENT=1
export GO15VENDOREXPERIMENT

COMPILERVERSION := $(shell go version | cut -d' ' -f3 | sed 's/go//g' | tr -d '\n')
COMPILER_VER_MAJOR := $(shell echo $(COMPILERVERSION) | cut -f1 -d.)
COMPILER_VER_MINOR := $(shell echo $(COMPILERVERSION) | cut -f2 -d.)
COMPILER_GT_1_10 := $(shell [ $(COMPILER_VER_MAJOR) -gt 1 -o \( $(COMPILER_VER_MAJOR) -eq 1 -a $(COMPILER_VER_MINOR) -ge 10 \) ] && echo true)

ifeq ($(COMPILER_GT_1_10), true)
LDFLAGS := -ldflags=all=
else
LDFLAGS := -ldflags=
endif

ifdef STATIC_BUILD
  CGO_ENABLED=0
  export CGO_ENABLED
  EXTRA_BUILDFLAGS=-installsuffix cgo
  EXTRA_LDFLAGS=-s -w
endif

SHASUMCMD := $(shell command -v sha1sum || command -v shasum; 2> /dev/null)
ifndef SHASUMCMD
  $(error "Neither sha1sum nor shasum command is available")
endif

SHA256SUMCMD := $(shell command -v sha256sum || command -v shasum; 2> /dev/null)
ifndef SHA256SUMCMD
  $(error "Neither sha256sum nor shasum command is available")
endif
ifeq ($(SHA256SUMCMD), "shasum")
SHA256SUMCMD = "shasum -a 256"
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
clean: # Remove build directory and bindata-generated files
	for t in ${BINDATA_TARGETS}; do if test -e $$t; then rm -fv $$t; fi; done
	if test -e ${BUILD}; then rm -rfv ${BUILD}; fi

.PHONY: kops
kops: ${KOPS}

.PHONY: ${KOPS}
${KOPS}: ${BINDATA_TARGETS}
	go build ${GCFLAGS} ${EXTRA_BUILDFLAGS} ${LDFLAGS}"-X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA} ${EXTRA_LDFLAGS}" -o $@ k8s.io/kops/cmd/kops/

${GOBINDATA}:
	mkdir -p ${LOCAL}
	go build ${GCFLAGS} ${EXTRA_BUILDFLAGS} ${LDFLAGS}"${EXTRA_LDFLAGS}" -o $@ k8s.io/kops/vendor/github.com/jteeuwen/go-bindata/go-bindata

.PHONY: gobindata-tool
gobindata-tool: ${GOBINDATA}

.PHONY: kops-gobindata
kops-gobindata: gobindata-tool ${BINDATA_TARGETS}

UPUP_MODELS_BINDATA_SOURCES:=$(shell find upup/models/ | egrep -v "upup/models/bindata.go")
upup/models/bindata.go: ${GOBINDATA} ${UPUP_MODELS_BINDATA_SOURCES}
	cd ${GOPATH_1ST}/src/k8s.io/kops; ${GOBINDATA} -o $@ -pkg models -ignore="\\.DS_Store" -ignore="bindata\\.go" -ignore="vfs\\.go" -prefix upup/models/ upup/models/...

# Build in a docker container with golang 1.X
# Used to test we have not broken 1.X
# 1.10 is the default for k8s 1.11.  Others are best-effort
.PHONY: check-builds-in-go18
check-builds-in-go18:
	# Note we only check that kops builds; we know the tests don't compile because of type aliasing in uber zap
	docker run -v ${GOPATH_1ST}/src/k8s.io/kops:/go/src/k8s.io/kops golang:1.8 make -C /go/src/k8s.io/kops kops

.PHONY: check-builds-in-go19
check-builds-in-go19:
	docker run -v ${GOPATH_1ST}/src/k8s.io/kops:/go/src/k8s.io/kops golang:1.9 make -C /go/src/k8s.io/kops ci

.PHONY: check-builds-in-go110
check-builds-in-go110:
	docker run -v ${GOPATH_1ST}/src/k8s.io/kops:/go/src/k8s.io/kops golang:1.10 make -C /go/src/k8s.io/kops ci

.PHONY: check-builds-in-go111
check-builds-in-go111:
	docker run -v ${GOPATH_1ST}/src/k8s.io/kops:/go/src/k8s.io/kops golang:1.11 make -C /go/src/k8s.io/kops ci

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
	docker run --name=nodeup-build-${UNIQUE} -e STATIC_BUILD=yes -e VERSION=${VERSION} -v ${MAKEDIR}:/go/src/k8s.io/kops golang:${GOVERSION} make -C /go/src/k8s.io/kops/ crossbuild-nodeup
	docker cp nodeup-build-${UNIQUE}:/go/.build .

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
	(${SHA256SUMCMD} ${DIST}/darwin/amd64/kops | cut -d' ' -f1) > ${DIST}/darwin/amd64/kops.sha256
	(${SHASUMCMD} ${DIST}/linux/amd64/kops | cut -d' ' -f1) > ${DIST}/linux/amd64/kops.sha1
	(${SHA256SUMCMD} ${DIST}/linux/amd64/kops | cut -d' ' -f1) > ${DIST}/linux/amd64/kops.sha256
	(${SHASUMCMD} ${DIST}/windows/amd64/kops.exe | cut -d' ' -f1) > ${DIST}/windows/amd64/kops.exe.sha1
	(${SHA256SUMCMD} ${DIST}/windows/amd64/kops.exe | cut -d' ' -f1) > ${DIST}/windows/amd64/kops.exe.sha256

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

.PHONY: ${PROTOKUBE}
${PROTOKUBE}:
	go build ${GCFLAGS} -o $@ -tags 'peer_name_alternative peer_name_hash' k8s.io/kops/protokube/cmd/protokube

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
	(${SHA256SUMCMD} ${IMAGES}/protokube.tar.gz | cut -d' ' -f1) > ${IMAGES}/protokube.tar.gz.sha256

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
	docker run --name=nodeup-build-${UNIQUE} -e STATIC_BUILD=yes -e VERSION=${VERSION} -v ${MAKEDIR}:/go/src/k8s.io/kops golang:${GOVERSION} make -C /go/src/k8s.io/kops/ nodeup
	docker start nodeup-build-${UNIQUE}
	docker exec nodeup-build-${UNIQUE} chown -R ${UID}:${GID} /go/src/k8s.io/kops/.build
	docker cp nodeup-build-${UNIQUE}:/go/src/k8s.io/kops/.build/local/nodeup .build/dist/
	(${SHASUMCMD} .build/dist/nodeup | cut -d' ' -f1) > .build/dist/nodeup.sha1
	(${SHA256SUMCMD} .build/dist/nodeup | cut -d' ' -f1) > .build/dist/nodeup.sha256

.PHONY: dns-controller-gocode
dns-controller-gocode:
	go install ${GCFLAGS} -tags 'peer_name_alternative peer_name_hash' ${LDFLAGS}"${EXTRA_LDFLAGS} -X main.BuildVersion=${DNS_CONTROLLER_TAG}" k8s.io/kops/dns-controller/cmd/dns-controller

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

.PHONY: bazel-utils-dist
bazel-utils-dist:
	bazel build //images/utils-builder:utils

# --------------------------------------------------
# development targets

.PHONY: dep-prereqs
dep-prereqs:
	(which hg > /dev/null) || (echo "dep requires that mercurial is installed"; exit 1)
	(which dep > /dev/null) || (echo "dep-ensure requires that dep is installed"; exit 1)
	(which bazel > /dev/null) || (echo "dep-ensure requires that bazel is installed"; exit 1)

.PHONY: dep-ensure
dep-ensure: dep-prereqs
	echo "`make dep-ensure` has been replaced by `make gomod`"
	exit 1
	dep ensure -v
	# Switch weavemesh to use peer_name_hash - bazel rule-go doesn't support build tags yet
	rm vendor/github.com/weaveworks/mesh/peer_name_mac.go
	sed -i -e 's/peer_name_hash/!peer_name_mac/g' vendor/github.com/weaveworks/mesh/peer_name_hash.go
	# Remove all bazel build files that were vendored and regenerate (we assume they are go-gettable)
	find vendor/ -name "BUILD" -delete
	find vendor/ -name "BUILD.bazel" -delete
	# Remove recursive symlinks that really confuse bazel
	rm -rf vendor/github.com/coreos/etcd/cmd/
	rm -rf vendor/github.com/jteeuwen/go-bindata/testdata/
	# Remove depenencies that dep just can't figure out
	rm -rf vendor/k8s.io/code-generator/cmd/set-gen/
	rm -rf vendor/k8s.io/code-generator/cmd/go-to-protobuf/
	rm -rf vendor/k8s.io/code-generator/cmd/import-boss/
	rm -rf vendor/github.com/docker/docker/contrib/
	make gazelle

.PHONY: gomod
gomod:
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
	find $(MAKEDIR) -name "*.go" | grep -v vendor | xargs bazel run //:gofmt -- -w -s

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

# ci target is for developers, it aims to cover all the CI jobs
# verify-gendocs will call kops target
# verify-package has to be after verify-gendoc, because with .gitignore for federation bindata
# it bombs in travis. verify-gendoc generates the bindata file.
.PHONY: ci
ci: govet verify-gofmt verify-gomod verify-goimports verify-boilerplate verify-bazel verify-misspelling nodeup examples test | verify-gendocs verify-packages verify-apimachinery
	echo "Done!"

# travis-ci is the target that travis-ci calls
# we skip tasks that rely on bazel and are covered by other jobs
#  verify-gofmt: uses bazel, covered by pull-kops-verify-gofmt
#  verify-bazel: uses bazel, covered by pull-kops-verify-bazel
#  govet: covered by pull-kops-verify-govet
#  verify-boilerplate: covered by pull-kops-verify-boilerplate
.PHONY: travis-ci
travis-ci: verify-misspelling nodeup examples test | verify-gendocs verify-packages verify-apimachinery
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
	git tag ${KOPS_RELEASE_VERSION}
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
	${GOPATH}/bin/conversion-gen ${API_OPTIONS} --skip-unsafe=true --input-dirs k8s.io/kops/pkg/apis/kops/v1alpha1 --v=0  --output-file-base=zz_generated.conversion \
		 --go-header-file "hack/boilerplate/boilerplate.go.txt"
	${GOPATH}/bin/conversion-gen ${API_OPTIONS} --skip-unsafe=true --input-dirs k8s.io/kops/pkg/apis/kops/v1alpha2 --v=0  --output-file-base=zz_generated.conversion \
		 --go-header-file "hack/boilerplate/boilerplate.go.txt"
	${GOPATH}/bin/deepcopy-gen ${API_OPTIONS} --input-dirs k8s.io/kops/pkg/apis/kops --v=0  --output-file-base=zz_generated.deepcopy \
		 --go-header-file "hack/boilerplate/boilerplate.go.txt"
	${GOPATH}/bin/deepcopy-gen ${API_OPTIONS} --input-dirs k8s.io/kops/pkg/apis/kops/v1alpha1 --v=0  --output-file-base=zz_generated.deepcopy \
	${GOPATH}/bin/deepcopy-gen ${API_OPTIONS} --input-dirs k8s.io/kops/pkg/apis/kops/v1alpha2 --v=0  --output-file-base=zz_generated.deepcopy \
		 --go-header-file "hack/boilerplate/boilerplate.go.txt"
	${GOPATH}/bin/defaulter-gen ${API_OPTIONS} --input-dirs k8s.io/kops/pkg/apis/kops/v1alpha1 --v=0  --output-file-base=zz_generated.defaults \
		 --go-header-file "hack/boilerplate/boilerplate.go.txt"
	${GOPATH}/bin/defaulter-gen ${API_OPTIONS} --input-dirs k8s.io/kops/pkg/apis/kops/v1alpha2 --v=0  --output-file-base=zz_generated.defaults \
		 --go-header-file "hack/boilerplate/boilerplate.go.txt"
	#go install github.com/ugorji/go/codec/codecgen
	# codecgen works only if invoked from directory where the file is located.
	#cd pkg/apis/kops/ && ~/k8s/bin/codecgen -d 1234 -o types.generated.go instancegroup.go cluster.go
	${GOPATH}/bin/client-gen  ${API_OPTIONS} --input-base k8s.io/kops/pkg/apis/ --input="kops/,kops/v1alpha1,kops/v1alpha2" --clientset-path k8s.io/kops/pkg/client/clientset_generated/ \
		 --go-header-file "hack/boilerplate/boilerplate.go.txt"
	${GOPATH}/bin/client-gen  ${API_OPTIONS} --clientset-name="clientset" --input-base k8s.io/kops/pkg/apis/ --input="kops/,kops/v1alpha1,kops/v1alpha2" --clientset-path k8s.io/kops/pkg/client/clientset_generated/ \
		 --go-header-file "hack/boilerplate/boilerplate.go.txt"

.PHONY: verify-apimachinery
verify-apimachinery:
	hack/verify-apimachinery.sh

# -----------------------------------------------------
# kops-server

.PHONY: kops-server-docker-compile
kops-server-docker-compile:
	GOOS=linux GOARCH=amd64 go build ${GCFLAGS} -a ${EXTRA_BUILDFLAGS} -o ${DIST}/linux/amd64/kops-server ${LDFLAGS}"${EXTRA_LDFLAGS} -X k8s.io/kops-server.Version=${VERSION} -X k8s.io/kops-server.GitVersion=${GITSHA}" k8s.io/kops/cmd/kops-server

.PHONY: kops-server-build
kops-server-build:
	# Compile the API binary in linux, and copy to local filesystem
	docker pull golang:${GOVERSION}
	docker run --name=kops-server-build-${UNIQUE} -e STATIC_BUILD=yes -e VERSION=${VERSION} -v ${GOPATH}/src:/go/src -v ${MAKEDIR}:/go/src/k8s.io/kops golang:${GOVERSION} make -C /go/src/k8s.io/kops/ kops-server-docker-compile
	docker cp kops-server-build-${UNIQUE}:/go/src/k8s.io/kops/.build .
	docker build -t ${DOCKER_REGISTRY}/kops-server:${KOPS_SERVER_TAG} -f images/kops-server/Dockerfile .

.PHONY: kops-server-push
kops-server-push: kops-server-build
	docker push ${DOCKER_REGISTRY}/kops-server:latest

# -----------------------------------------------------
# bazel targets

.PHONY: bazel-test
bazel-test:
	bazel ${BAZEL_OPTIONS} test  --test_output=errors -- //... -//vendor/...

.PHONY: bazel-build
bazel-build:
	bazel build --features=pure //cmd/... //pkg/... //channels/... //nodeup/... //protokube/... //dns-controller/... //util/...

.PHONY: bazel-build-cli
bazel-build-cli:
	bazel build --features=pure //cmd/kops/...

.PHONY: bazel-crossbuild-kops
bazel-crossbuild-kops:
	bazel build --features=pure --platforms=@io_bazel_rules_go//go/toolchain:darwin_amd64 //cmd/kops/...
	bazel build --features=pure --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //cmd/kops/...
	bazel build --features=pure --platforms=@io_bazel_rules_go//go/toolchain:windows_amd64 //cmd/kops/...

.PHONY: bazel-crossbuild-nodeup
bazel-crossbuild-nodeup:
	bazel build --features=pure --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //cmd/nodeup/...

.PHONY: bazel-crossbuild-protokube
bazel-crossbuild-protokube:
	bazel build --features=pure --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //protokube/...

.PHONY: bazel-crossbuild-dns-controller
bazel-crossbuild-dns-controller:
	bazel build --features=pure --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //dns-controller/...

.PHONY: bazel-crossbuild-dns-controller-image
bazel-crossbuild-dns-controller-image:
	bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //images:dns-controller.tar

.PHONY: bazel-crossbuild-protokube-image
bazel-crossbuild-protokube-image:
	bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //images:protokube.tar

.PHONY: bazel-crossbuild-kube-discovery-image
bazel-crossbuild-kube-discovery-image:
	bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //images:kube-discovery.tar

.PHONY: bazel-crossbuild-node-authorizer-image
bazel-crossbuild-node-authorizer-image:
	bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //images:node-authorizer.tar

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
	ssh -t ${TARGET} sudo SKIP_PACKAGE_UPDATE=1 /tmp/nodeup --conf=/var/cache/kubernetes-install/kube_env.yaml --v=8

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
	bazel run //kube-discovery/images:kube-discovery
	docker tag bazel/kube-discovery/images:kube-discovery ${DOCKER_REGISTRY}/kube-discovery:${DOCKER_TAG}
	docker push ${DOCKER_REGISTRY}/kube-discovery:${DOCKER_TAG}

.PHONY: push-node-authorizer
push-node-authorizer:
	bazel run //node-authorizer/images:node-authorizer
	docker tag bazel/node-authorizer/images:node-authorizer ${DOCKER_REGISTRY}/node-authorizer:${DOCKER_TAG}
	docker push ${DOCKER_REGISTRY}/node-authorizer:${DOCKER_TAG}

.PHONY: bazel-protokube-export
bazel-protokube-export:
	mkdir -p ${BAZELIMAGES}
	bazel build --action_env=PROTOKUBE_TAG=${PROTOKUBE_TAG} --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //images:protokube.tar
	cp -fp bazel-bin/images/protokube.tar ${BAZELIMAGES}/protokube.tar
	gzip --force --fast ${BAZELIMAGES}/protokube.tar
	(${SHASUMCMD} ${BAZELIMAGES}/protokube.tar.gz | cut -d' ' -f1) > ${BAZELIMAGES}/protokube.tar.gz.sha1
	(${SHA256SUMCMD} ${BAZELIMAGES}/protokube.tar.gz | cut -d' ' -f1) > ${BAZELIMAGES}/protokube.tar.gz.sha256

.PHONY: bazel-version-dist
bazel-version-dist: bazel-crossbuild-nodeup bazel-crossbuild-kops bazel-protokube-export bazel-utils-dist
	rm -rf ${BAZELUPLOAD}
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/darwin/amd64/
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/windows/amd64/
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/images/
	mkdir -p ${BAZELUPLOAD}/utils/${VERSION}/linux/amd64/
	cp bazel-bin/cmd/nodeup/linux_amd64_pure_stripped/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup
	(${SHASUMCMD} ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup | cut -d' ' -f1) > ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup.sha1
	(${SHA256SUMCMD} ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup | cut -d' ' -f1) > ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup.sha256
	cp ${BAZELIMAGES}/protokube.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/protokube.tar.gz
	cp ${BAZELIMAGES}/protokube.tar.gz.sha1 ${BAZELUPLOAD}/kops/${VERSION}/images/protokube.tar.gz.sha1
	cp ${BAZELIMAGES}/protokube.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/protokube.tar.gz.sha256
	cp bazel-bin/cmd/kops/linux_amd64_pure_stripped/kops ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/kops
	(${SHASUMCMD} ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/kops | cut -d' ' -f1) > ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/kops.sha1
	(${SHA256SUMCMD} ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/kops | cut -d' ' -f1) > ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/kops.sha256
	cp bazel-bin/cmd/kops/darwin_amd64_pure_stripped/kops ${BAZELUPLOAD}/kops/${VERSION}/darwin/amd64/kops
	(${SHASUMCMD} ${BAZELUPLOAD}/kops/${VERSION}/darwin/amd64/kops | cut -d' ' -f1) > ${BAZELUPLOAD}/kops/${VERSION}/darwin/amd64/kops.sha1
	(${SHA256SUMCMD} ${BAZELUPLOAD}/kops/${VERSION}/darwin/amd64/kops | cut -d' ' -f1) > ${BAZELUPLOAD}/kops/${VERSION}/darwin/amd64/kops.sha256
	cp bazel-bin/cmd/kops/windows_amd64_pure_stripped/kops.exe ${BAZELUPLOAD}/kops/${VERSION}/windows/amd64/kops.exe
	(${SHASUMCMD} ${BAZELUPLOAD}/kops/${VERSION}/windows/amd64/kops.exe | cut -d' ' -f1) > ${BAZELUPLOAD}/kops/${VERSION}/windows/amd64/kops.exe.sha1
	(${SHA256SUMCMD} ${BAZELUPLOAD}/kops/${VERSION}/windows/amd64/kops.exe | cut -d' ' -f1) > ${BAZELUPLOAD}/kops/${VERSION}/windows/amd64/kops.exe.sha256
	cp bazel-bin/images/utils-builder/utils.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/utils.tar.gz
	(${SHASUMCMD} ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/utils.tar.gz | cut -d' ' -f1) > ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/utils.tar.gz.sha1
	(${SHA256SUMCMD} ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/utils.tar.gz | cut -d' ' -f1) > ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/utils.tar.gz.sha256
	cp -fr ${BAZELUPLOAD}/kops/${VERSION}/* ${BAZELDIST}/

.PHONY: bazel-upload
bazel-upload: bazel-version-dist # Upload kops to S3
	aws s3 sync --acl public-read ${BAZELUPLOAD}/ ${S3_BUCKET}

# prow-postsubmit is run by the prow postsubmit job
# It uploads a build to a staging directory, which in theory we can publish as a release
.PHONY: prow-postsubmit
prow-postsubmit: bazel-version-dist
	${UPLOAD} ${BAZELUPLOAD}/kops/${VERSION}/ ${UPLOAD_DEST}/${KOPS_RELEASE_VERSION}-${GITSHA}/

#-----------------------------------------------------------
# static html documentation

.PHONY: live-docs
live-docs:
	@docker run --rm -it -p 3000:3000 -v ${PWD}:/docs aledbf/mkdocs:0.1

.PHONY: build-docs
build-docs:
	@docker run --rm -it -v ${PWD}:/docs aledbf/mkdocs:0.1 build

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
	(${SHASUMCMD} ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup | cut -d' ' -f1) > ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup.sha1
	(${SHA256SUMCMD} ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup | cut -d' ' -f1) > ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup.sha256
	${UPLOAD} ${BAZELUPLOAD}/ ${UPLOAD_DEST}

# dev-upload-protokube uploads protokube to GCS
.PHONY: dev-upload-protokube
dev-upload-protokube: bazel-protokube-export # Upload kops to GCS
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/images/
	cp -fp ${BAZELIMAGES}/protokube.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/protokube.tar.gz
	cp -fp ${BAZELIMAGES}/protokube.tar.gz.sha1 ${BAZELUPLOAD}/kops/${VERSION}/images/protokube.tar.gz.sha1
	cp -fp ${BAZELIMAGES}/protokube.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/protokube.tar.gz.sha256
	${UPLOAD} ${BAZELUPLOAD}/ ${UPLOAD_DEST}

# dev-copy-utils copies utils from a recent release
# We don't currently have a bazel build for them, and the build is pretty slow, but they change rarely.
.PHONE: dev-copy-utils
dev-copy-utils:
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/
	cd ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/; wget -N https://kubeupv2.s3.amazonaws.com/kops/1.11.0-alpha.1/linux/amd64/utils.tar.gz
	cd ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/; wget -N https://kubeupv2.s3.amazonaws.com/kops/1.11.0-alpha.1/linux/amd64/utils.tar.gz.sha1
	cd ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/; wget -N https://kubeupv2.s3.amazonaws.com/kops/1.11.0-alpha.1/linux/amd64/utils.tar.gz.sha256
	${UPLOAD} ${BAZELUPLOAD}/ ${UPLOAD_DEST}

# dev-upload does a faster build and uploads to GCS / S3
# It copies utils instead of building it
.PHONY: dev-upload
dev-upload: dev-upload-nodeup dev-upload-protokube dev-copy-utils
	echo "Done"

.PHONY: crds
crds:
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go crd --apis-path pkg/apis/kops/v1alpha2 --domain k8s.io --output-dir k8s/crds/
