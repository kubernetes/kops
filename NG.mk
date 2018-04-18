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


KOPS_RELEASE_VERSION = 1.8.0-alpha.1
KOPS_CI_VERSION      = 1.8.0-alpha.2


HOST_OS:=$(shell uname -a | cut -d \  -f 1 | awk '{print tolower($$0)}')
UNIQUE:=$(shell date +%s)
GOVERSION=1.8.3
BUILD=.build
LOCAL=$(BUILD)/local
BINDATA_TARGETS=upup/models/bindata.go federation/model/bindata.go
PROTOKUBE_ARTIFACTS=$(BUILD)/protokube
DIST=$(BUILD)/kops
GOBINDATA=$(LOCAL)/go-bindata
CHANNELS=$(PROTOKUBE_ARTIFACTS)/channels
NODEUP=${DIST}/${KOPS_RELEASE_VERSION}/${HOST_OS}/amd64/nodeup
EXAMPLES=$(LOCAL)/examples
PROTOKUBE=$(PROTOKUBE_ARTIFACTS)/protokube
KUBECTL=$(PROTOKUBE_ARTIFACTS)/kubectl
TESTABLE_PACKAGES:=$(shell egrep -v "k8s.io/kops/cloudmock|k8s.io/kops/vendor" hack/.packages) 
GIT_BRANCH:=$(shell git rev-parse --abbrev-ref HEAD)
BUILD_IMAGE_NAME:=kops-build-${GIT_BRANCH}
SOURCES:=$(shell find . -name "*.go")
DOCKER_WORKING:=/go/src/k8s.io/kops
DOCKER_RUN:=docker run -w ${DOCKER_WORKING}

LINUX_DIST_BINARIES:=kops kops-server nodeup
DARWIN_DIST_BINARIES:=kops

# kops local location
KOPS := ${DIST}/${KOPS_RELEASE_VERSION}/${HOST_OS}/amd64/kops

GITSHA := $(shell git describe --always)

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

IMAGES=$(DIST)/$(KOPS_RELEASE_VERSION)/images

ifeq ($(shell which pigz 2>&1 > /dev/null; echo $$?),0)
	GZIP:=pigz
else
	GZIP:=gzip
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
kops-install: ${BINDATA_TARGETS}
	go install ${EXTRA_BUILDFLAGS} -ldflags "-X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA} ${EXTRA_LDFLAGS}" k8s.io/kops/cmd/kops/

.PHONY: clean
clean: # Remove build directory and bindata-generated files
	for t in ${BINDATA_TARGETS}; do if test -e $$t; then rm -fv $$t; fi; done 
	if test -e ${BUILD}; then rm -rfv ${BUILD}; fi

.PHONY: all
all: version-dist

${GOBINDATA}:
	mkdir -p ${LOCAL}
	go build ${EXTRA_BUILDFLAGS} -ldflags "${EXTRA_LDFLAGS}" -o $@ ./vendor/github.com/jteeuwen/go-bindata/go-bindata

UPUP_MODELS_BINDATA_SOURCES:=$(shell find upup/models | egrep -v "upup/models/bindata.go")
upup/models/bindata.go: ${GOBINDATA} ${UPUP_MODELS_BINDATA_SOURCES}
	${GOBINDATA} -o $@ -pkg models -ignore="\\.DS_Store" -ignore="bindata\\.go" -ignore="vfs\\.go" -prefix upup/models/ upup/models/...

FEDERATION_MODELS_BINDATA_SOURCES:=$(shell find federation/model | egrep -v "federation/model/bindata.go")
federation/model/bindata.go: ${GOBINDATA} ${FEDERATION_MODELS_BINDATA_SOURCES}
	${GOBINDATA} -o $@ -pkg model -ignore="\\.DS_Store" -ignore="bindata\\.go" -prefix federation/model/ federation/model/...


.PHONY: codegen
codegen: kops-gobindata
	go install k8s.io/kops/upup/tools/generators/...
	go generate k8s.io/kops/upup/pkg/fi/cloudup/awstasks
	go generate k8s.io/kops/upup/pkg/fi/cloudup/gcetasks
	go generate k8s.io/kops/upup/pkg/fi/cloudup/dotasks
	go generate k8s.io/kops/upup/pkg/fi/assettasks
	go generate k8s.io/kops/upup/pkg/fi/fitasks

.PHONY: test
test: ${BINDATA_TARGETS}  # Run tests locally
	go test -v ${TESTABLE_PACKAGES}

${KUBECTL}:
	mkdir -p $(@D)
	curl -L https://storage.googleapis.com/kubernetes-release/release/v1.6.6/bin/linux/amd64/kubectl -o $@
	chmod +x $@

.PHONY: protokube-container
protokube-container: ${IMAGES}/protokube.tar.gz ${IMAGES}/protokube.tar.gz.sha1

${IMAGES}/protokube.tar.gz.sha1: ${IMAGES}/protokube.tar.gz
	${SHASUMCMD} $< | cut -d' ' -f1  > $@

${DIST}/${KOPS_RELEASE_VERSION}/images/protokube.tar.gz: ${CHANNELS} ${PROTOKUBE} ${KUBECTL}
	docker build -t protokube:${PROTOKUBE_TAG} -f images/protokube-ng/Dockerfile .
	mkdir -p $(@D)
	docker save protokube:${PROTOKUBE_TAG} | ${GZIP} --force --best > $@

# --------------------------------------------------
# static utils

.PHONY: utils-dist
utils-dist: ${DIST}/${KOPS_RELEASE_VERSION}/linux/amd64/utils.tar.gz.sha1 ${DIST}/${KOPS_RELEASE_VERSION}/linux/amd64/utils.tar.gz

${DIST}/${KOPS_RELEASE_VERSION}/linux/amd64/utils.tar.gz:
	mkdir -p $(@D)
	docker build -t utils-builder images/utils-builder-ng
	docker run --name kops-utils-${UNIQUE} utils-builder
	docker cp kops-utils-${UNIQUE}:/utils.tar.gz $@
	docker rm kops-utils-${UNIQUE}

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
	if ! diff -r "$$TMP_DOCS" ./docs/cli; then \
	     echo "Please run make gen-cli-docs." 1>&2; \
	     exit 1; \
	fi
	@echo "cli docs up-to-date"
#
# verify-gendocs will call kops target
# verify-package has to be after verify-gendoc, because with .gitignore for federation bindata
# it bombs in travis. verify-gendoc generates the bindata file.
.PHONY: ci
ci: govet verify-gofmt verify-boilerplate nodeup  examples test | verify-gendocs verify-packages
	echo "Done!"

# --------------------------------------------------
# API / embedding examples

.PHONY: examples
examples: ${EXAMPLES} # Install kops API example

${EXAMPLES}: ${BINDATA_TARGETS}
	go build -o ${EXAMPLES} k8s.io/kops/examples/kops-api-example/...

# -----------------------------------------------------
# crossbuild targets

# wildcard target for protokube and channels
${PROTOKUBE_ARTIFACTS}/%: ${BINDATA_TARGETS}
	mkdir -p $(@D)
	${DOCKER_RUN} \
	-e "GOOS=linux" -e "GOARCH=amd64" -e "STATIC_BUILD=yes" \
	--name protokube-build-$(@F)-${UNIQUE} \
	-v `pwd`:/go/src/k8s.io/kops:ro \
	golang:${GOVERSION} \
	go build ${EXTRA_BUILDFLAGS} \
	-ldflags "-X k8s.io/kops.Version=${VERSION} \
	-X k8s.io/kops.GitVersion=${GITSHA} ${EXTRA_LDFLAGS}" \
	-o /tmp/$@ ./$(@F)/cmd/$(@F)/...
	docker cp protokube-build-$(@F)-${UNIQUE}:/tmp/$@ $@
	docker rm protokube-build-$(@F)-${UNIQUE}

# first position in BINARY_template is the binary name, second position is the
# OS, third is the architecture
define BINARY_template
${DIST}/${KOPS_RELEASE_VERSION}/$(2)/$(3)/$(1): ${BINDATA_TARGETS}
	mkdir -p $$(@D)
	${DOCKER_RUN} \
	-e "GOOS=$(2)" -e "GOARCH=$(3)" -e "STATIC_BUILD=yes"\
	--name kops-build-$(1)-$(2)-$(3)-${UNIQUE} \
	-v `pwd`:${DOCKER_WORKING}:ro \
	golang:${GOVERSION} \
	go build ${EXTRA_BUILDFLAGS} \
	-ldflags "-X k8s.io/kops.Version=${VERSION} \
	-X k8s.io/kops.GitVersion=${GITSHA} ${EXTRA_LDFLAGS}" \
	-o /tmp/$$@ ./cmd/$(1)/
	docker cp kops-build-$(1)-$(2)-$(3)-${UNIQUE}:/tmp/$$@ $$@ 
	docker rm kops-build-$(1)-$(2)-$(3)-${UNIQUE}
endef

# first position on SHA_template is OS, second is architecture
define SHA_BIN_template
${DIST}/${KOPS_RELEASE_VERSION}/$(1)/$(2)/%.sha1: .build/kops/${KOPS_RELEASE_VERSION}/$(1)/$(2)/%
	${SHASUMCMD} $$< | cut -d' ' -f1  > $$@
endef

# SHA targets for Linux and Darwin, amd64 only.
$(eval $(call SHA_BIN_template,linux,amd64))
$(eval $(call SHA_BIN_template,darwin,amd64))

# targets for every binary listed in LINUX_DIST_BINARIES
$(foreach bin,$(LINUX_DIST_BINARIES),$(eval $(call BINARY_template,$(bin),linux,amd64)))

# targets for every binary listed in DARWIN_DIST_BINARIES
$(foreach bin,$(DARWIN_DIST_BINARIES),$(eval $(call BINARY_template,$(bin),darwin,amd64)))

# build a list of targets for distribution to Linux
LINUX_DIST_TARGETS=
# binaries
$(foreach bin,$(LINUX_DIST_BINARIES),$(eval LINUX_DIST_TARGETS += ${DIST}/${KOPS_RELEASE_VERSION}/linux/amd64/$(bin)))
# hashes
$(foreach bin,$(LINUX_DIST_BINARIES),$(eval LINUX_DIST_TARGETS += ${DIST}/${KOPS_RELEASE_VERSION}/linux/amd64/$(bin).sha1))
.PHONY: linux-dist
linux-dist: ${LINUX_DIST_TARGETS}

# build a list of targets for distribution to Darwin
DARWIN_DIST_TARGETS=
# binaries
$(foreach bin,$(DARWIN_DIST_BINARIES),$(eval DARWIN_DIST_TARGETS += ${DIST}/${KOPS_RELEASE_VERSION}/darwin/amd64/$(bin)))
# hashes
$(foreach bin,$(DARWIN_DIST_BINARIES),$(eval DARWIN_DIST_TARGETS += ${DIST}/${KOPS_RELEASE_VERSION}/darwin/amd64/$(bin).sha1))
.PHONY: darwin-dist
darwin-dist: ${DARWIN_DIST_TARGETS}

.PHONY: version-dist
version-dist: linux-dist darwin-dist protokube-container utils-dist
