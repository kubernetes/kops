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
KOPS_ROOT?=$(patsubst %/,%,$(abspath $(dir $(firstword $(MAKEFILE_LIST)))))
DOCKER_REGISTRY?=gcr.io/must-override
S3_BUCKET?=s3://must-override/
UPLOAD_DEST?=$(S3_BUCKET)
GCS_LOCATION?=gs://must-override
GCS_URL=$(GCS_LOCATION:gs://%=https://storage.googleapis.com/%)
LATEST_FILE?=latest-ci.txt
GOPATH_1ST:=$(shell go env | grep GOPATH | cut -f 2 -d '"' | sed 's/ /\\ /g')
UNIQUE:=$(shell date +%s)
BUILD=$(KOPS_ROOT)/.build
LOCAL=$(BUILD)/local
ARTIFACTS?=$(BUILD)/artifacts
DIST=$(BUILD)/dist
IMAGES=$(DIST)/images
CHANNELS=$(LOCAL)/channels
PROTOKUBE=$(LOCAL)/protokube
UPLOAD=$(BUILD)/upload
BAZELBUILD=$(KOPS_ROOT)/.bazelbuild
BAZELDIST=$(BAZELBUILD)/dist
BAZELIMAGES=$(BAZELDIST)/images
BAZELUPLOAD=$(BAZELBUILD)/upload
UID:=$(shell id -u)
GID:=$(shell id -g)
BAZEL?=bazelisk
BAZEL_OPTIONS?=
BAZEL_CONFIG?=
API_OPTIONS?=
GCFLAGS?=
BAZEL_BIN=.bazel-bin

# CODEGEN_VERSION is the version of k8s.io/code-generator to use
CODEGEN_VERSION=v0.22.2


UPLOAD_CMD=$(KOPS_ROOT)/hack/upload ${UPLOAD_ARGS}

# Unexport environment variables that can affect tests and are not used in builds
unexport AWS_ACCESS_KEY_ID AWS_REGION AWS_SECRET_ACCESS_KEY AWS_SESSION_TOKEN CNI_VERSION_URL DNS_IGNORE_NS_CHECK DNSCONTROLLER_IMAGE DO_ACCESS_TOKEN GOOGLE_APPLICATION_CREDENTIALS
unexport KOPS_BASE_URL KOPS_CLUSTER_NAME KOPS_RUN_OBSOLETE_VERSION KOPS_STATE_STORE KOPS_STATE_S3_ACL KUBE_API_VERSIONS NODEUP_URL OPENSTACK_CREDENTIAL_FILE SKIP_PACKAGE_UPDATE
unexport SKIP_REGION_CHECK S3_ACCESS_KEY_ID S3_ENDPOINT S3_REGION S3_SECRET_ACCESS_KEY


VERSION=$(shell tools/get_version.sh | grep VERSION | awk '{print $$2}')

KOPS_RELEASE_VERSION:=$(shell grep 'KOPS_RELEASE_VERSION\s*=' version.go | awk '{print $$3}' | sed -e 's_"__g')
KOPS_CI_VERSION:=$(shell grep 'KOPS_CI_VERSION\s*=' version.go | awk '{print $$3}' | sed -e 's_"__g')

# kops local location
KOPS                 = ${LOCAL}/kops

GITSHA := $(shell cd ${KOPS_ROOT}; git describe --always)

# We lock the versions of our controllers also
# We need to keep in sync with:
#   upup/models/cloudup/resources/addons/dns-controller/
DNS_CONTROLLER_TAG=1.24.0-alpha.3
DNS_CONTROLLER_PUSH_TAG=$(shell tools/get_workspace_status.sh | grep STABLE_DNS_CONTROLLER_TAG | awk '{print $$2}')
#   upup/models/cloudup/resources/addons/kops-controller.addons.k8s.io/
KOPS_CONTROLLER_TAG=1.24.0-alpha.3
KOPS_CONTROLLER_PUSH_TAG=$(shell tools/get_workspace_status.sh | grep STABLE_KOPS_CONTROLLER_TAG | awk '{print $$2}')
#   pkg/model/components/kubeapiserver/model.go
KUBE_APISERVER_HEALTHCHECK_TAG=1.24.0-alpha.3
KUBE_APISERVER_HEALTHCHECK_PUSH_TAG=$(shell tools/get_workspace_status.sh | grep STABLE_KUBE_APISERVER_HEALTHCHECK_TAG | awk '{print $$2}')


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
kops-install:
	go install ${GCFLAGS} ${EXTRA_BUILDFLAGS} ${LDFLAGS}"-X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA} ${EXTRA_LDFLAGS}" k8s.io/kops/cmd/kops/

.PHONY: channels-install # Install channels to local $GOPATH/bin
channels-install: ${CHANNELS}
	cp ${CHANNELS} ${GOPATH_1ST}/bin

.PHONY: all-install # Install all kops project binaries
all-install: all kops-install channels-install
	cp ${NODEUP} ${GOPATH_1ST}/bin
	cp ${PROTOKUBE} ${GOPATH_1ST}/bin

.PHONY: all
all: ${KOPS} ${PROTOKUBE} nodeup ${CHANNELS} ko-kops-controller-export ko-dns-controller-export ko-kube-apiserver-healthcheck-export

include tests/e2e/e2e.mk

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
	${BAZEL} clean
	rm -rf tests/integration/update_cluster/*/.terraform

.PHONY: kops
kops: ${KOPS}

.PHONY: ${KOPS}
${KOPS}:
	go build ${GCFLAGS} ${EXTRA_BUILDFLAGS} ${LDFLAGS}"-X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA} ${EXTRA_LDFLAGS}" -o $@ k8s.io/kops/cmd/kops/

.PHONY: codegen
codegen:
	go build -o ${KOPS_ROOT}/_output/bin k8s.io/kops/upup/tools/generators/...
	${KOPS_ROOT}/_output/bin/fitask \
		--input-dirs k8s.io/kops/upup/pkg/fi/... \
		--go-header-file hack/boilerplate/boilerplate.generatego.txt \
		--output-base ${KOPS_ROOT}

.PHONY: verify-codegen
verify-codegen:
	go build -o ${KOPS_ROOT}/_output/bin k8s.io/kops/upup/tools/generators/...
	${KOPS_ROOT}/_output/bin/fitask --verify-only \
		--input-dirs k8s.io/kops/upup/pkg/fi/... \
		--go-header-file hack/boilerplate/boilerplate.generatego.txt \
		--output-base ${KOPS_ROOT}

.PHONY: protobuf
protobuf:
	cd ${GOPATH_1ST}/src; protoc --gogo_out=. k8s.io/kops/protokube/pkg/gossip/mesh/mesh.proto

.PHONY: hooks
hooks: # Install Git hooks
	cp hack/pre-commit.sh .git/hooks/pre-commit

.PHONY: test
test:
	go test -v ./...

.PHONY: test-windows
test-windows:
	go test -v $(go list ./... | grep -v /nodeup/)

.PHONY: nodeup-amd64 nodeup-arm64
nodeup-amd64 nodeup-arm64: nodeup-%:
	mkdir -p ${DIST}/linux/$*
	GOOS=linux GOARCH=$* go build ${GCFLAGS} -a ${EXTRA_BUILDFLAGS} -o ${DIST}/linux/$*/nodeup ${LDFLAGS}"${EXTRA_LDFLAGS} -X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA}" k8s.io/kops/cmd/nodeup

.PHONY: nodeup
nodeup: nodeup-amd64

.PHONY: crossbuild-nodeup
crossbuild-nodeup: nodeup-amd64 nodeup-arm64

.PHONY: ${DIST}/darwin/amd64/kops ${DIST}/darwin/arm64/kops
${DIST}/darwin/amd64/kops ${DIST}/darwin/arm64/kops: ${DIST}/darwin/%/kops:
	mkdir -p ${DIST}
	GOOS=darwin GOARCH=$* go build ${GCFLAGS} -a ${EXTRA_BUILDFLAGS} -o $@ ${LDFLAGS}"${EXTRA_LDFLAGS} -X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA}" k8s.io/kops/cmd/kops

.PHONY: ${DIST}/linux/amd64/kops ${DIST}/linux/arm64/kops
${DIST}/linux/amd64/kops ${DIST}/linux/arm64/kops: ${DIST}/linux/%/kops:
	mkdir -p ${DIST}
	GOOS=linux GOARCH=$* go build ${GCFLAGS} -a ${EXTRA_BUILDFLAGS} -o $@ ${LDFLAGS}"${EXTRA_LDFLAGS} -X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA}" k8s.io/kops/cmd/kops

.PHONY: ${DIST}/windows/amd64/kops.exe
${DIST}/windows/amd64/kops.exe:
	mkdir -p ${DIST}
	GOOS=windows GOARCH=amd64 go build ${GCFLAGS} -a ${EXTRA_BUILDFLAGS} -o $@ ${LDFLAGS}"${EXTRA_LDFLAGS} -X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA}" k8s.io/kops/cmd/kops

.PHONY: crossbuild
crossbuild: ${DIST}/windows/amd64/kops.exe ${DIST}/darwin/amd64/kops ${DIST}/darwin/arm64/kops ${DIST}/linux/amd64/kops ${DIST}/linux/arm64/kops

.PHONY: upload
upload: bazel-version-dist # Upload kops to S3
	aws s3 sync --acl public-read ${UPLOAD}/ ${S3_BUCKET}

# gcs-upload builds kops and uploads to GCS
.PHONY: gcs-upload
gcs-upload: bazel-version-dist
	@echo "== Uploading kops =="
	gsutil -h "Cache-Control:private, max-age=0, no-transform" -m cp -n -r ${BAZELUPLOAD}/kops/* ${GCS_LOCATION}

# gcs-upload-tag runs gcs-upload to upload, then uploads a version-marker to LATEST_FILE
.PHONY: gcs-upload-and-tag
gcs-upload-and-tag: gcs-upload
	echo "${GCS_URL}${VERSION}" > ${BAZELUPLOAD}/latest.txt
	gsutil -h "Cache-Control:private, max-age=0, no-transform" cp ${BAZELUPLOAD}/latest.txt ${GCS_LOCATION}${LATEST_FILE}

.PHONY: bazel-version-ci
bazel-version-ci: bazel-version-dist-linux-amd64 bazel-version-dist-linux-arm64
	rm -rf ${BAZELUPLOAD}
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/images/
	cp ${BAZEL_BIN}/cmd/kops/linux-amd64/kops ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/kops
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/kops ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/kops.sha256
	cp ${BAZEL_BIN}/cmd/nodeup/linux-amd64/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup.sha256
	cp ${BAZEL_BIN}/cmd/nodeup/linux-arm64/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/nodeup
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/nodeup.sha256
	cp -fp ${BAZEL_BIN}/channels/cmd/channels/linux-amd64/channels ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/channels
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/channels ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/channels.sha256
	cp -fp ${BAZEL_BIN}/channels/cmd/channels/linux-arm64/channels ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/channels
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/channels ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/channels.sha256
	cp -fp ${BAZEL_BIN}/protokube/cmd/protokube/linux-amd64/protokube ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/protokube
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/protokube ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/protokube.sha256
	cp -fp ${BAZEL_BIN}/protokube/cmd/protokube/linux-arm64/protokube ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/protokube
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/protokube ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/protokube.sha256
	cp ${BAZELIMAGES}/kops-controller-amd64.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/kops-controller-amd64.tar.gz
	cp ${BAZELIMAGES}/kops-controller-amd64.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/kops-controller-amd64.tar.gz.sha256
	cp ${BAZELIMAGES}/kops-controller-arm64.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/kops-controller-arm64.tar.gz
	cp ${BAZELIMAGES}/kops-controller-arm64.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/kops-controller-arm64.tar.gz.sha256
	cp ${BAZELIMAGES}/kube-apiserver-healthcheck-amd64.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/kube-apiserver-healthcheck-amd64.tar.gz
	cp ${BAZELIMAGES}/kube-apiserver-healthcheck-amd64.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/kube-apiserver-healthcheck-amd64.tar.gz.sha256
	cp ${BAZELIMAGES}/kube-apiserver-healthcheck-arm64.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/kube-apiserver-healthcheck-arm64.tar.gz
	cp ${BAZELIMAGES}/kube-apiserver-healthcheck-arm64.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/kube-apiserver-healthcheck-arm64.tar.gz.sha256
	cp ${BAZELIMAGES}/dns-controller-amd64.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/dns-controller-amd64.tar.gz
	cp ${BAZELIMAGES}/dns-controller-amd64.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/dns-controller-amd64.tar.gz.sha256
	cp ${BAZELIMAGES}/dns-controller-arm64.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/dns-controller-arm64.tar.gz
	cp ${BAZELIMAGES}/dns-controller-arm64.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/dns-controller-arm64.tar.gz.sha256
	cp -fr ${BAZELUPLOAD}/kops/${VERSION}/* ${BAZELDIST}/

# gcs-publish-ci is the entry point for CI testing
# In CI testing, always upload the CI version.
.PHONY: gcs-publish-ci
gcs-publish-ci: VERSION := ${KOPS_CI_VERSION}+${GITSHA}
gcs-publish-ci: bazel-version-ci
	@echo "== Uploading kops =="
	gsutil -h "Cache-Control:private, max-age=0, no-transform" -m cp -n -r ${BAZELUPLOAD}/kops/* ${GCS_LOCATION}
	echo "VERSION: ${VERSION}"
	echo "${GCS_URL}/${VERSION}" > ${BAZELUPLOAD}/${LATEST_FILE}
	gsutil -h "Cache-Control:private, max-age=0, no-transform" cp ${BAZELUPLOAD}/${LATEST_FILE} ${GCS_LOCATION}

.PHONY: gen-cli-docs
gen-cli-docs: ${KOPS} # Regenerate CLI docs
	KOPS_STATE_STORE= \
	KOPS_FEATURE_FLAGS= \
	${KOPS} gen-cli-docs --out docs/cli

.PHONY: push-amd64 push-arm64
push-amd64 push-arm64: push-%: nodeup-%
	scp -C ${DIST}/linux/$*/nodeup  ${TARGET}:/tmp/

.PHONY: push-gce-dry-amd64 push-gce-dry-arm64
push-gce-dry-amd64 push-gce-dry-arm64: push-gce-dry-%: push-%
	ssh ${TARGET} sudo /tmp/nodeup --conf=metadata://gce/instance/attributes/config --dryrun --v=8

.PHONY: push-aws-dry-amd64 push-aws-dry-arm64
push-aws-dry-amd64 push-aws-dry-arm64: push-aws-dry-%: push-%
	ssh ${TARGET} sudo /tmp/nodeup --conf=/opt/kops/conf/kube_env.yaml --dryrun --v=8

.PHONY: push-gce-run-amd64 push-gce-run-arm64
push-gce-run-amd64 push-gce-run-arm64: push-gce-run-%: push-%
	ssh ${TARGET} sudo cp /tmp/nodeup /var/lib/toolbox/kubernetes-install/nodeup
	ssh ${TARGET} sudo /var/lib/toolbox/kubernetes-install/nodeup --conf=/var/lib/toolbox/kubernetes-install/kube_env.yaml --v=8

# -t is for CentOS http://unix.stackexchange.com/questions/122616/why-do-i-need-a-tty-to-run-sudo-if-i-can-sudo-without-a-password
.PHONY: push-aws-run-amd64 push-aws-run-arm64
push-aws-run-amd64 push-aws-run-arm64: push-aws-run-%: push-%
	ssh -t ${TARGET} sudo /tmp/nodeup --conf=/opt/kops/conf/kube_env.yaml --v=8

${PROTOKUBE}:
	go build ${GCFLAGS} ${EXTRA_BUILDFLAGS} -o $@ -tags 'peer_name_alternative peer_name_hash' k8s.io/kops/protokube/cmd/protokube

.PHONY: protokube
protokube: ${PROTOKUBE}


.PHONY: ${NODEUP}
${NODEUP}:
	go build ${GCFLAGS} ${EXTRA_BUILDFLAGS} ${LDFLAGS}"${EXTRA_LDFLAGS} -X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA}" -o $@ k8s.io/kops/cmd/nodeup

.PHONY: bazel-crossbuild-dns-controller
bazel-crossbuild-dns-controller:
	${BAZEL} ${BAZEL_OPTIONS} build ${BAZEL_CONFIG} --@io_bazel_rules_go//go/config:pure --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //dns-controller/...
	${BAZEL} ${BAZEL_OPTIONS} build ${BAZEL_CONFIG} --@io_bazel_rules_go//go/config:pure --platforms=@io_bazel_rules_go//go/toolchain:linux_arm64 //dns-controller/...


.PHONY: dns-controller-push
dns-controller-push: bazelisk
	DOCKER_REGISTRY=${DOCKER_REGISTRY} DOCKER_IMAGE_PREFIX=${DOCKER_IMAGE_PREFIX} DNS_CONTROLLER_TAG=${DNS_CONTROLLER_PUSH_TAG} ${BAZEL} run --@io_bazel_rules_go//go/config:pure --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //dns-controller/cmd/dns-controller:push-image-amd64
	DOCKER_REGISTRY=${DOCKER_REGISTRY} DOCKER_IMAGE_PREFIX=${DOCKER_IMAGE_PREFIX} DNS_CONTROLLER_TAG=${DNS_CONTROLLER_PUSH_TAG} ${BAZEL} run --@io_bazel_rules_go//go/config:pure --platforms=@io_bazel_rules_go//go/toolchain:linux_arm64 //dns-controller/cmd/dns-controller:push-image-arm64

.PHONY: dns-controller-manifest
dns-controller-manifest:
	docker manifest create --amend ${DOCKER_REGISTRY}/${DOCKER_IMAGE_PREFIX}dns-controller:${DNS_CONTROLLER_PUSH_TAG} ${DOCKER_REGISTRY}/${DOCKER_IMAGE_PREFIX}dns-controller:${DNS_CONTROLLER_PUSH_TAG}-amd64
	docker manifest create --amend ${DOCKER_REGISTRY}/${DOCKER_IMAGE_PREFIX}dns-controller:${DNS_CONTROLLER_PUSH_TAG} ${DOCKER_REGISTRY}/${DOCKER_IMAGE_PREFIX}dns-controller:${DNS_CONTROLLER_PUSH_TAG}-arm64
	docker manifest push --purge ${DOCKER_REGISTRY}/${DOCKER_IMAGE_PREFIX}dns-controller:${DNS_CONTROLLER_PUSH_TAG}

# --------------------------------------------------
# development targets

.PHONY: gomod-prereqs
gomod-prereqs:
	(which ${BAZEL} > /dev/null) || (echo "gomod requires that ${BAZEL} is installed"; exit 1)

.PHONY: gomod
gomod: bazelisk gomod-prereqs
	go mod tidy
	go mod vendor
	# Switch weavemesh to use peer_name_hash - bazel rule-go doesn't support build tags yet
	rm vendor/github.com/weaveworks/mesh/peer_name_mac.go
	sed -i.bak -e 's/peer_name_hash/!peer_name_mac/g' vendor/github.com/weaveworks/mesh/peer_name_hash.go
	# Remove all bazel build files that were vendored and regenerate (we assume they are go-gettable)
	find vendor/ -name "BUILD" -delete
	find vendor/ -name "BUILD.bazel" -delete
	make gazelle
	cd tests/e2e; go mod tidy
	cd hack; go mod tidy


.PHONY: gofmt
gofmt:
	find $(KOPS_ROOT) -name "*.go" | grep -v vendor | xargs ${BAZEL} run @go_sdk//:bin/gofmt -- -w -s

.PHONY: goimports
goimports:
	hack/update-goimports.sh

.PHONY: verify-goimports
verify-goimports:
	hack/verify-goimports.sh

.PHONY: govet
govet:
	go vet ./...

# --------------------------------------------------
# Continuous integration targets

# verify is ran by the pull-kops-verify prow job
.PHONY: verify
verify: quick-ci verify-gofmt

.PHONY: verify-boilerplate
verify-boilerplate:
	hack/verify-boilerplate.sh

.PHONY: verify-gofmt
verify-gofmt: bazelisk
	hack/verify-gofmt.sh

.PHONY: verify-gomod
verify-gomod:
	hack/verify-gomod.sh

# find release notes, remove PR titles and output the rest to .build, then run misspell on all files
.PHONY: verify-misspelling
verify-misspelling:
	hack/verify-spelling.sh

.PHONY: verify-gendocs
verify-gendocs: ${KOPS}
	@TMP_DOCS="$$(mktemp -d)"; \
	'${KOPS}' gen-cli-docs --out "$$TMP_DOCS"; \
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

.PHONY: verify-golangci-lint
verify-golangci-lint:
	hack/verify-golangci-lint.sh

.PHONY: verify-shellcheck
verify-shellcheck:
	hack/verify-shellcheck.sh

.PHONY: verify-terraform
verify-terraform:
	hack/verify-terraform.sh

.PHONE: verify-cloudformation
verify-cloudformation:
	hack/verify-cloudformation.sh

.PHONY: verify-hashes
verify-hashes:
	hack/verify-hashes.sh

# ci target is for developers, it aims to cover all the CI jobs
# verify-gendocs will call kops target
.PHONY: ci
ci: govet verify-gofmt verify-crds verify-gomod verify-goimports verify-boilerplate verify-bazel verify-versions verify-misspelling verify-shellcheck verify-golangci-lint verify-terraform nodeup examples test | verify-gendocs verify-apimachinery verify-codegen
	echo "Done!"

# we skip tasks that rely on bazel and are covered by other jobs
# verify-gofmt: uses bazel, covered by pull-kops-verify
.PHONY: quick-ci
quick-ci: verify-crds verify-goimports govet verify-boilerplate verify-bazel verify-versions verify-misspelling verify-shellcheck | verify-gendocs verify-apimachinery verify-codegen
	echo "Done!"

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
	shipbot -tag v${KOPS_RELEASE_VERSION} -config .shipbot.yaml -src .build/dist/

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
apimachinery-codegen: apimachinery-codegen-conversion apimachinery-codegen-deepcopy apimachinery-codegen-defaulter apimachinery-codegen-client

.PHONY: apimachinery-codegen-conversion
apimachinery-codegen-conversion:
	go run k8s.io/code-generator/cmd/conversion-gen@${CODEGEN_VERSION} ${API_OPTIONS} --skip-unsafe=true --input-dirs ./pkg/apis/kops/v1alpha2 --v=0  --output-file-base=zz_generated.conversion \
		 --go-header-file "hack/boilerplate/boilerplate.generatego.txt"
	grep 'requires manual conversion' ${KOPS_ROOT}/pkg/apis/kops/v1alpha2/zz_generated.conversion.go ; [ $$? -eq 1 ]
	go run k8s.io/code-generator/cmd/conversion-gen@${CODEGEN_VERSION} ${API_OPTIONS} --skip-unsafe=true --input-dirs ./pkg/apis/kops/v1alpha3 --v=0  --output-file-base=zz_generated.conversion \
		 --go-header-file "hack/boilerplate/boilerplate.generatego.txt"
	grep 'requires manual conversion' ${KOPS_ROOT}/pkg/apis/kops/v1alpha3/zz_generated.conversion.go ; [ $$? -eq 1 ]

.PHONY: apimachinery-codegen-deepcopy
apimachinery-codegen-deepcopy:
	go run k8s.io/code-generator/cmd/deepcopy-gen@${CODEGEN_VERSION} ${API_OPTIONS} --input-dirs ./pkg/apis/kops --v=0  --output-file-base=zz_generated.deepcopy \
		 --go-header-file "hack/boilerplate/boilerplate.generatego.txt"
	go run k8s.io/code-generator/cmd/deepcopy-gen@${CODEGEN_VERSION} ${API_OPTIONS} --input-dirs ./pkg/apis/kops/v1alpha2 --v=0  --output-file-base=zz_generated.deepcopy \
		 --go-header-file "hack/boilerplate/boilerplate.generatego.txt"
	go run k8s.io/code-generator/cmd/deepcopy-gen@${CODEGEN_VERSION} ${API_OPTIONS} --input-dirs ./pkg/apis/kops/v1alpha3 --v=0  --output-file-base=zz_generated.deepcopy \
		 --go-header-file "hack/boilerplate/boilerplate.generatego.txt"

.PHONY: apimachinery-codegen-defaulter
apimachinery-codegen-defaulter:
	go run k8s.io/code-generator/cmd/defaulter-gen@${CODEGEN_VERSION} ${API_OPTIONS} --input-dirs ./pkg/apis/kops/v1alpha2 --v=0  --output-file-base=zz_generated.defaults \
		 --go-header-file "hack/boilerplate/boilerplate.generatego.txt"
	go run k8s.io/code-generator/cmd/defaulter-gen@${CODEGEN_VERSION} ${API_OPTIONS} --input-dirs ./pkg/apis/kops/v1alpha3 --v=0  --output-file-base=zz_generated.defaults \
		 --go-header-file "hack/boilerplate/boilerplate.generatego.txt"

.PHONY: apimachinery-codegen-client
apimachinery-codegen-client:
	# Hack: doesn't seem to be a way to output to the current directory - wants to insert the full package path
	go run k8s.io/code-generator/cmd/client-gen@${CODEGEN_VERSION} ${API_OPTIONS} --input-base=k8s.io/kops/pkg/apis --input-dirs=. --input="kops/,kops/v1alpha2,kops/v1alpha3" --output-package=k8s.io/kops/pkg/client/clientset_generated/ --output-base=../../ \
		 --go-header-file "hack/boilerplate/boilerplate.generatego.txt"
	go run k8s.io/code-generator/cmd/client-gen@${CODEGEN_VERSION} ${API_OPTIONS} --clientset-name="clientset" --input-base=k8s.io/kops/pkg/apis --input-dirs=. --input="kops/,kops/v1alpha2,kops/v1alpha3" --output-package=k8s.io/kops/pkg/client/clientset_generated/  --output-base=../../ \
		 --go-header-file "hack/boilerplate/boilerplate.generatego.txt"

.PHONY: verify-apimachinery
verify-apimachinery:
	hack/verify-apimachinery.sh

.PHONY: verify-generate
verify-generate: verify-crds

.PHONY: verify-crds
verify-crds:
	hack/verify-crds.sh

.PHONY: verify-versions
verify-versions:
	hack/verify-versions.sh

# -----------------------------------------------------
# bazel targets

.PHONY: bazel-test
bazel-test: bazelisk
	${BAZEL} ${BAZEL_OPTIONS} test ${BAZEL_CONFIG} --@io_bazel_rules_go//go/config:pure --test_output=errors -- //... -//vendor/...

.PHONY: bazel-build
bazel-build: bazelisk
	${BAZEL} ${BAZEL_OPTIONS} build ${BAZEL_CONFIG} --@io_bazel_rules_go//go/config:pure //cmd/... //pkg/... //channels/... //nodeup/... //protokube/... //dns-controller/... //util/...

.PHONY: bazel-build-cli
bazel-build-cli: bazelisk
	${BAZEL} ${BAZEL_OPTIONS} build ${BAZEL_CONFIG} --@io_bazel_rules_go//go/config:pure //cmd/kops/...

.PHONY: bazel-build-kops-darwin-amd64 bazel-build-kops-darwin-arm64
bazel-build-kops-darwin-amd64 bazel-build-kops-darwin-arm64: bazel-build-kops-darwin-%:
	${BAZEL} ${BAZEL_OPTIONS} build ${BAZEL_CONFIG} --@io_bazel_rules_go//go/config:pure --platforms=@io_bazel_rules_go//go/toolchain:darwin_$* //cmd/kops/...

.PHONY: bazel-build-kops-linux-amd64 bazel-build-kops-linux-arm64
bazel-build-kops-linux-amd64 bazel-build-kops-linux-arm64: bazel-build-kops-linux-%:
	${BAZEL} ${BAZEL_OPTIONS} build ${BAZEL_CONFIG} --@io_bazel_rules_go//go/config:pure --platforms=@io_bazel_rules_go//go/toolchain:linux_$* //cmd/kops/...

.PHONY: bazel-build-kops-windows-amd64
bazel-build-kops-windows-amd64:
	${BAZEL} ${BAZEL_OPTIONS} build ${BAZEL_CONFIG} --@io_bazel_rules_go//go/config:pure --platforms=@io_bazel_rules_go//go/toolchain:windows_amd64 //cmd/kops/...

.PHONY: bazel-crossbuild-kops
bazel-crossbuild-kops: bazel-build-kops-darwin-amd64 bazel-build-kops-darwin-arm64 bazel-build-kops-linux-amd64 bazel-build-kops-linux-arm64 bazel-build-kops-windows-amd64
	echo "Done cross-building kops"

.PHONY: bazel-build-nodeup-linux-amd64 bazel-build-nodeup-linux-arm64
bazel-build-nodeup-linux-amd64 bazel-build-nodeup-linux-arm64: bazel-build-nodeup-linux-%:
	${BAZEL} ${BAZEL_OPTIONS} build ${BAZEL_CONFIG} --@io_bazel_rules_go//go/config:pure --platforms=@io_bazel_rules_go//go/toolchain:linux_$* //cmd/nodeup/...

.PHONY: bazel-crossbuild-nodeup
bazel-crossbuild-nodeup: bazel-build-nodeup-linux-amd64 bazel-build-nodeup-linux-arm64
	echo "Done cross-building nodeup"

.PHONY: bazel-build-protokube-linux-amd64 bazel-build-protokube-linux-arm64
bazel-build-protokube-linux-amd64 bazel-build-protokube-linux-arm64: bazel-build-protokube-linux-%:
	${BAZEL} ${BAZEL_OPTIONS} build ${BAZEL_CONFIG} --@io_bazel_rules_go//go/config:pure --platforms=@io_bazel_rules_go//go/toolchain:linux_$* //protokube/...

.PHONY: bazel-crossbuild-protokube
bazel-crossbuild-protokube: bazel-build-protokube-linux-amd64 bazel-build-protokube-linux-arm64
	echo "Done cross-building protokube"

.PHONY: bazel-build-channels-linux-amd64 bazel-build-channels-linux-arm64
bazel-build-channels-linux-amd64 bazel-build-channels-linux-arm64: bazel-build-channels-linux-%:
	${BAZEL} ${BAZEL_OPTIONS} build ${BAZEL_CONFIG} --@io_bazel_rules_go//go/config:pure --platforms=@io_bazel_rules_go//go/toolchain:linux_$* //channels/...

.PHONY: bazel-crossbuild-channels
bazel-crossbuild-channels: bazel-build-channels-linux-amd64 bazel-build-channels-linux-arm64
	echo "Done cross-building channels"

.PHONY: bazel-push
# Will always push a linux-based build up to the server
bazel-push: bazel-build-nodeup-linux-amd64
	ssh ${TARGET} touch /tmp/nodeup
	ssh ${TARGET} chmod +w /tmp/nodeup
	scp -C ${BAZEL_BIN}/cmd/nodeup/linux-amd64/nodeup  ${TARGET}:/tmp/

.PHONY: bazel-push-gce-run
bazel-push-gce-run: bazel-push
	ssh ${TARGET} sudo cp /tmp/nodeup /var/lib/toolbox/kubernetes-install/nodeup
	ssh ${TARGET} sudo SKIP_PACKAGE_UPDATE=1 /var/lib/toolbox/kubernetes-install/nodeup --conf=/var/lib/toolbox/kubernetes-install/kube_env.yaml --v=8

# -t is for CentOS http://unix.stackexchange.com/questions/122616/why-do-i-need-a-tty-to-run-sudo-if-i-can-sudo-without-a-password
.PHONY: bazel-push-aws-run
bazel-push-aws-run: bazel-push
	ssh ${TARGET} chmod +x /tmp/nodeup
	ssh -t ${TARGET} sudo SKIP_PACKAGE_UPDATE=1 /tmp/nodeup --conf=/opt/kops/conf/kube_env.yaml --v=8

.PHONY: ko
ko:
	hack/install-ko.sh

.PHONY: bazelisk
bazelisk:
	hack/install-bazelisk.sh

.PHONY: gazelle
gazelle:
	hack/update-bazel.sh

.PHONY: check-markdown-links
check-markdown-links:
	docker run -t -v $$PWD:/tmp \
		-e LC_ALL=C.UTF-8 \
		-e LANG=en_US.UTF-8 \
		-e LANGUAGE=en_US.UTF-8 \
		rubygem/awesome_bot --allow-dupe --allow-redirect \
		$(shell find $$PWD -name "*.md" -mindepth 1 -printf '%P\n' | grep -v vendor | grep -v Changelog.md)

#-----------------------------------------------------------

.PHONY: bazel-kube-apiserver-healthcheck-export-linux-amd64 bazel-kube-apiserver-healthcheck-export-linux-arm64
bazel-kube-apiserver-healthcheck-export-linux-amd64 bazel-kube-apiserver-healthcheck-export-linux-arm64: bazel-kube-apiserver-healthcheck-export-linux-%:
	mkdir -p ${BAZELIMAGES}
	DOCKER_REGISTRY="" DOCKER_IMAGE_PREFIX="registry.k8s.io/kops/" KUBE_APISERVER_HEALTHCHECK_TAG=${KUBE_APISERVER_HEALTHCHECK_TAG} ${BAZEL} ${BAZEL_OPTIONS} build ${BAZEL_CONFIG} --@io_bazel_rules_go//go/config:pure --platforms=@io_bazel_rules_go//go/toolchain:linux_$* //cmd/kube-apiserver-healthcheck:image-bundle-$*.tar.gz.sha256
	cp -fp ${BAZEL_BIN}/cmd/kube-apiserver-healthcheck/image-bundle-$*.tar.gz ${BAZELIMAGES}/kube-apiserver-healthcheck-$*.tar.gz
	cp -fp ${BAZEL_BIN}/cmd/kube-apiserver-healthcheck/image-bundle-$*.tar.gz.sha256 ${BAZELIMAGES}/kube-apiserver-healthcheck-$*.tar.gz.sha256

.PHONY: bazel-kube-apiserver-healthcheck-export
bazel-kube-apiserver-healthcheck-export: bazel-kube-apiserver-healthcheck-export-linux-amd64 bazel-kube-apiserver-healthcheck-export-linux-arm64
	echo "Done exporting kube-apiserver-healthcheck images"

.PHONY: bazel-kops-controller-export-linux-amd64 bazel-kops-controller-export-linux-arm64
bazel-kops-controller-export-linux-amd64 bazel-kops-controller-export-linux-arm64: bazel-kops-controller-export-linux-%:
	mkdir -p ${BAZELIMAGES}
	DOCKER_REGISTRY="" DOCKER_IMAGE_PREFIX="registry.k8s.io/kops/" KOPS_CONTROLLER_TAG=${KOPS_CONTROLLER_TAG} ${BAZEL} ${BAZEL_OPTIONS} build ${BAZEL_CONFIG} --@io_bazel_rules_go//go/config:pure --platforms=@io_bazel_rules_go//go/toolchain:linux_$* //cmd/kops-controller:image-bundle-$*.tar.gz.sha256
	cp -fp ${BAZEL_BIN}/cmd/kops-controller/image-bundle-$*.tar.gz ${BAZELIMAGES}/kops-controller-$*.tar.gz
	cp -fp ${BAZEL_BIN}/cmd/kops-controller/image-bundle-$*.tar.gz.sha256 ${BAZELIMAGES}/kops-controller-$*.tar.gz.sha256

.PHONY: bazel-kops-controller-export
bazel-kops-controller-export: bazel-kops-controller-export-linux-amd64 bazel-kops-controller-export-linux-arm64
	echo "Done exporting kops-controller images"

.PHONY: ko-kops-controller-export-linux-amd64 ko-kops-controller-export-linux-arm64
ko-kops-controller-export-linux-amd64 ko-kops-controller-export-linux-arm64: ko-kops-controller-export-linux-%: ko
	mkdir -p ${IMAGES}
	KO_DOCKER_REPO="registry.k8s.io/kops" ko build --tags ${KOPS_CONTROLLER_TAG} --platform=linux/$* -B --push=false --tarball=${IMAGES}/kops-controller-$*.tar ./cmd/kops-controller/
	gzip -f ${IMAGES}/kops-controller-$*.tar
	tools/sha256 ${IMAGES}/kops-controller-$*.tar.gz ${IMAGES}/kops-controller-$*.tar.gz.sha256

.PHONY: ko-kops-controller-export
ko-kops-controller-export: ko-kops-controller-export-linux-amd64 ko-kops-controller-export-linux-arm64
	echo "Done exporting kops-controller images"

.PHONY: ko-kube-apiserver-healthcheck-export-linux-amd64 ko-kube-apiserver-healthcheck-export-linux-arm64
ko-kube-apiserver-healthcheck-export-linux-amd64 ko-kube-apiserver-healthcheck-export-linux-arm64: ko-kube-apiserver-healthcheck-export-linux-%: ko
	mkdir -p ${IMAGES}
	KO_DOCKER_REPO="registry.k8s.io/kops" ko build --tags ${KUBE_APISERVER_HEALTHCHECK_TAG} --platform=linux/$* -B --push=false --tarball=${IMAGES}/kube-apiserver-healthcheck-$*.tar ./cmd/kube-apiserver-healthcheck
	gzip -f ${IMAGES}/kube-apiserver-healthcheck-$*.tar
	tools/sha256 ${IMAGES}/kube-apiserver-healthcheck-$*.tar.gz ${IMAGES}/kube-apiserver-healthcheck-$*.tar.gz.sha256

.PHONY: ko-kube-apiserver-healthcheck-export
ko-kube-apiserver-healthcheck-export: ko-kube-apiserver-healthcheck-export-linux-amd64 ko-kube-apiserver-healthcheck-export-linux-arm64
	echo "Done exporting kube-apiserver-healthcheck images"

.PHONY: ko-dns-controller-export-linux-amd64 ko-dns-controller-export-linux-arm64
ko-dns-controller-export-linux-amd64 ko-dns-controller-export-linux-arm64: ko-dns-controller-export-linux-%: ko
	mkdir -p ${IMAGES}
	KO_DOCKER_REPO="registry.k8s.io/kops" ko build --tags ${DNS_CONTROLLER_TAG} --platform=linux/$* -B --push=false --tarball=${IMAGES}/dns-controller-$*.tar ./dns-controller/cmd/dns-controller
	gzip -f ${IMAGES}/dns-controller-$*.tar
	tools/sha256 ${IMAGES}/dns-controller-$*.tar.gz ${IMAGES}/dns-controller-$*.tar.gz.sha256

.PHONY: ko-dns-controller-export
ko-dns-controller-export: ko-dns-controller-export-linux-amd64 ko-dns-controller-export-linux-arm64
	echo "Done exporting dns-controller images"

.PHONY: bazel-dns-controller-export-linux-amd64 bazel-dns-controller-export-linux-arm64
bazel-dns-controller-export-linux-amd64 bazel-dns-controller-export-linux-arm64: bazel-dns-controller-export-linux-%:
	mkdir -p ${BAZELIMAGES}
	DOCKER_REGISTRY="" DOCKER_IMAGE_PREFIX="registry.k8s.io/kops/" DNS_CONTROLLER_TAG=${DNS_CONTROLLER_TAG} ${BAZEL} ${BAZEL_OPTIONS} build ${BAZEL_CONFIG} --@io_bazel_rules_go//go/config:pure --platforms=@io_bazel_rules_go//go/toolchain:linux_$* //dns-controller/cmd/dns-controller:image-bundle-$*.tar.gz.sha256
	cp -fp ${BAZEL_BIN}/dns-controller/cmd/dns-controller/image-bundle-$*.tar.gz ${BAZELIMAGES}/dns-controller-$*.tar.gz
	cp -fp ${BAZEL_BIN}/dns-controller/cmd/dns-controller/image-bundle-$*.tar.gz.sha256 ${BAZELIMAGES}/dns-controller-$*.tar.gz.sha256

.PHONY: bazel-dns-controller-export
bazel-dns-controller-export:
	echo "Done exporting dns-controller images"

.PHONY: bazel-version-dist-linux-amd64 bazel-version-dist-linux-arm64
bazel-version-dist-linux-amd64 bazel-version-dist-linux-arm64: bazel-version-dist-linux-%: bazelisk bazel-build-kops-linux-% bazel-build-nodeup-linux-% bazel-kops-controller-export-linux-% bazel-kube-apiserver-healthcheck-export-linux-% bazel-dns-controller-export-linux-% bazel-build-protokube-linux-% bazel-build-channels-linux-%
	echo "Done building dist for $*"

.PHONY: bazel-version-dist
bazel-version-dist: bazel-version-dist-linux-amd64 bazel-version-dist-linux-arm64 bazel-build-kops-darwin-amd64 bazel-build-kops-darwin-arm64 bazel-build-kops-windows-amd64
	rm -rf ${BAZELUPLOAD}
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/darwin/amd64/
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/darwin/arm64/
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/windows/amd64/
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/images/
	cp ${BAZEL_BIN}/cmd/nodeup/linux-amd64/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup.sha256
	cp ${BAZEL_BIN}/cmd/nodeup/linux-arm64/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/nodeup
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/nodeup.sha256
	cp -fp ${BAZEL_BIN}/channels/cmd/channels/linux-amd64/channels ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/channels
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/channels ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/channels.sha256
	cp -fp ${BAZEL_BIN}/channels/cmd/channels/linux-arm64/channels ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/channels
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/channels ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/channels.sha256
	cp -fp ${BAZEL_BIN}/protokube/cmd/protokube/linux-amd64/protokube ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/protokube
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/protokube ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/protokube.sha256
	cp -fp ${BAZEL_BIN}/protokube/cmd/protokube/linux-arm64/protokube ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/protokube
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/protokube ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/protokube.sha256
	cp ${BAZELIMAGES}/kops-controller-amd64.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/kops-controller-amd64.tar.gz
	cp ${BAZELIMAGES}/kops-controller-amd64.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/kops-controller-amd64.tar.gz.sha256
	cp ${BAZELIMAGES}/kops-controller-arm64.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/kops-controller-arm64.tar.gz
	cp ${BAZELIMAGES}/kops-controller-arm64.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/kops-controller-arm64.tar.gz.sha256
	cp ${BAZELIMAGES}/kube-apiserver-healthcheck-amd64.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/kube-apiserver-healthcheck-amd64.tar.gz
	cp ${BAZELIMAGES}/kube-apiserver-healthcheck-amd64.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/kube-apiserver-healthcheck-amd64.tar.gz.sha256
	cp ${BAZELIMAGES}/kube-apiserver-healthcheck-arm64.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/kube-apiserver-healthcheck-arm64.tar.gz
	cp ${BAZELIMAGES}/kube-apiserver-healthcheck-arm64.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/kube-apiserver-healthcheck-arm64.tar.gz.sha256
	cp ${BAZELIMAGES}/dns-controller-amd64.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/dns-controller-amd64.tar.gz
	cp ${BAZELIMAGES}/dns-controller-amd64.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/dns-controller-amd64.tar.gz.sha256
	cp ${BAZELIMAGES}/dns-controller-arm64.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/dns-controller-arm64.tar.gz
	cp ${BAZELIMAGES}/dns-controller-arm64.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/dns-controller-arm64.tar.gz.sha256
	cp ${BAZEL_BIN}/cmd/kops/linux-amd64/kops ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/kops
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/kops ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/kops.sha256
	cp ${BAZEL_BIN}/cmd/kops/linux-arm64/kops ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/kops
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/kops ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/kops.sha256
	cp ${BAZEL_BIN}/cmd/kops/darwin-amd64/kops ${BAZELUPLOAD}/kops/${VERSION}/darwin/amd64/kops
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/darwin/amd64/kops ${BAZELUPLOAD}/kops/${VERSION}/darwin/amd64/kops.sha256
	cp ${BAZEL_BIN}/cmd/kops/darwin-arm64/kops ${BAZELUPLOAD}/kops/${VERSION}/darwin/arm64/kops
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/darwin/arm64/kops ${BAZELUPLOAD}/kops/${VERSION}/darwin/arm64/kops.sha256
	cp ${BAZEL_BIN}/cmd/kops/windows-amd64/kops ${BAZELUPLOAD}/kops/${VERSION}/windows/amd64/kops.exe
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/windows/amd64/kops.exe ${BAZELUPLOAD}/kops/${VERSION}/windows/amd64/kops.exe.sha256
	tar cfvz ${BAZELUPLOAD}/kops/${VERSION}/images/images.tar.gz -C ${BAZELIMAGES} --exclude "*.sha256" .
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/images.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/images.tar.gz.sha256
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

#-----------------------------------------------------------
# development targets

# dev-upload-nodeup uploads nodeup to GCS
.PHONY: dev-upload-nodeup
dev-upload-nodeup: bazel-crossbuild-nodeup
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/
	cp -fp ${BAZEL_BIN}/cmd/nodeup/linux-amd64/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup.sha256
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/
	cp -fp ${BAZEL_BIN}/cmd/nodeup/linux-arm64/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/nodeup
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/nodeup.sha256
	${UPLOAD_CMD} ${BAZELUPLOAD}/ ${UPLOAD_DEST}

.PHONY: dev-upload-nodeup-amd64
dev-upload-nodeup-amd64: bazel-build-nodeup-linux-amd64
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/
	cp -fp ${BAZEL_BIN}/cmd/nodeup/linux-amd64/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/nodeup.sha256
	${UPLOAD_CMD} ${BAZELUPLOAD}/ ${UPLOAD_DEST}

.PHONY: dev-upload-nodeup-arm64
dev-upload-nodeup-arm64: bazel-build-nodeup-linux-arm64
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/
	cp -fp ${BAZEL_BIN}/cmd/nodeup/linux-arm64/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/nodeup
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/nodeup.sha256
	${UPLOAD_CMD} ${BAZELUPLOAD}/ ${UPLOAD_DEST}

# dev-upload-protokube uploads protokube to GCS
.PHONY: dev-upload-protokube
dev-upload-protokube: bazel-crossbuild-protokube # Upload kops to GCS
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/
	cp -fp ${BAZEL_BIN}/protokube/cmd/protokube/linux-amd64/protokube ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/protokube
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/protokube ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/protokube.sha256
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/
	cp -fp ${BAZEL_BIN}/protokube/cmd/protokube/linux-arm64/protokube ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/protokube
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/protokube ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/protokube.sha256
	${UPLOAD_CMD} ${BAZELUPLOAD}/ ${UPLOAD_DEST}

# dev-upload-channels uploads channels to GCS
.PHONY: dev-upload-channels
dev-upload-channels: bazel-crossbuild-channels
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/
	cp -fp ${BAZEL_BIN}/channels/cmd/channels/linux-amd64/channels ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/channels
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/channels ${BAZELUPLOAD}/kops/${VERSION}/linux/amd64/channels.sha256
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/
	cp -fp ${BAZEL_BIN}/channels/cmd/channels/linux-arm64/channels ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/channels
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/channels ${BAZELUPLOAD}/kops/${VERSION}/linux/arm64/channels.sha256
	${UPLOAD_CMD} ${BAZELUPLOAD}/ ${UPLOAD_DEST}

# dev-upload-kops-controller uploads kops-controller to GCS
.PHONY: dev-upload-kops-controller
dev-upload-kops-controller: bazel-kops-controller-export # Upload kops to GCS
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/images/
	cp -fp ${BAZELIMAGES}/kops-controller-amd64.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/kops-controller-amd64.tar.gz
	cp -fp ${BAZELIMAGES}/kops-controller-amd64.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/kops-controller-amd64.tar.gz.sha256
	cp -fp ${BAZELIMAGES}/kops-controller-arm64.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/kops-controller-arm64.tar.gz
	cp -fp ${BAZELIMAGES}/kops-controller-arm64.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/kops-controller-arm64.tar.gz.sha256
	${UPLOAD_CMD} ${BAZELUPLOAD}/ ${UPLOAD_DEST}

# dev-upload-dns-controller uploads dns-controller to GCS
.PHONY: dev-upload-dns-controller
dev-upload-dns-controller: bazel-dns-controller-export # Upload kops to GCS
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/images/
	cp -fp ${BAZELIMAGES}/dns-controller-amd64.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/dns-controller-amd64.tar.gz
	cp -fp ${BAZELIMAGES}/dns-controller-amd64.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/dns-controller-amd64.tar.gz.sha256
	cp -fp ${BAZELIMAGES}/dns-controller-arm64.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/dns-controller-arm64.tar.gz
	cp -fp ${BAZELIMAGES}/dns-controller-arm64.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/dns-controller-arm64.tar.gz.sha256
	${UPLOAD_CMD} ${BAZELUPLOAD}/ ${UPLOAD_DEST}

# dev-upload-kube-apiserver-healthcheck uploads kube-apiserver-healthcheck to GCS
.PHONY: dev-upload-kube-apiserver-healthcheck
dev-upload-kube-apiserver-healthcheck: bazel-kube-apiserver-healthcheck-export # Upload kops to GCS
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/images/
	cp -fp ${BAZELIMAGES}/kube-apiserver-healthcheck-amd64.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/kube-apiserver-healthcheck-amd64.tar.gz
	cp -fp ${BAZELIMAGES}/kube-apiserver-healthcheck-amd64.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/kube-apiserver-healthcheck-amd64.tar.gz.sha256
	cp -fp ${BAZELIMAGES}/kube-apiserver-healthcheck-arm64.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/kube-apiserver-healthcheck-arm64.tar.gz
	cp -fp ${BAZELIMAGES}/kube-apiserver-healthcheck-arm64.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/kube-apiserver-healthcheck-arm64.tar.gz.sha256
	${UPLOAD_CMD} ${BAZELUPLOAD}/ ${UPLOAD_DEST}

# dev-upload-linux-amd64 does a faster build and uploads to GCS / S3
.PHONY: dev-upload-linux-amd64 dev-upload-linux-arm64
dev-upload-linux-amd64 dev-upload-linux-arm64: dev-upload-linux-%: bazel-build-nodeup-linux-% bazel-kops-controller-export-linux-% bazel-kube-apiserver-healthcheck-export-linux-% bazel-dns-controller-export-linux-% bazel-build-protokube-linux-% bazel-build-channels-linux-%
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/images/
	mkdir -p ${BAZELUPLOAD}/kops/${VERSION}/linux/$*/
	cp -fp ${BAZEL_BIN}/cmd/nodeup/linux-$*/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/$*/nodeup
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/$*/nodeup ${BAZELUPLOAD}/kops/${VERSION}/linux/$*/nodeup.sha256
	cp -fp ${BAZEL_BIN}/channels/cmd/channels/linux-$*/channels ${BAZELUPLOAD}/kops/${VERSION}/linux/$*/channels
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/$*/channels ${BAZELUPLOAD}/kops/${VERSION}/linux/$*/channels.sha256
	cp -fp ${BAZEL_BIN}/protokube/cmd/protokube/linux-$*/protokube ${BAZELUPLOAD}/kops/${VERSION}/linux/$*/protokube
	tools/sha256 ${BAZELUPLOAD}/kops/${VERSION}/linux/$*/protokube ${BAZELUPLOAD}/kops/${VERSION}/linux/$*/protokube.sha256
	cp -fp ${BAZELIMAGES}/kops-controller-$*.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/kops-controller-$*.tar.gz
	cp -fp ${BAZELIMAGES}/kops-controller-$*.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/kops-controller-$*.tar.gz.sha256
	cp -fp ${BAZELIMAGES}/dns-controller-$*.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/dns-controller-$*.tar.gz
	cp -fp ${BAZELIMAGES}/dns-controller-$*.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/dns-controller-$*.tar.gz.sha256
	cp -fp ${BAZELIMAGES}/kube-apiserver-healthcheck-$*.tar.gz ${BAZELUPLOAD}/kops/${VERSION}/images/kube-apiserver-healthcheck-$*.tar.gz
	cp -fp ${BAZELIMAGES}/kube-apiserver-healthcheck-$*.tar.gz.sha256 ${BAZELUPLOAD}/kops/${VERSION}/images/kube-apiserver-healthcheck-$*.tar.gz.sha256
	${UPLOAD_CMD} ${BAZELUPLOAD}/ ${UPLOAD_DEST}

# dev-upload does a faster build and uploads to GCS / S3
.PHONY: dev-upload
dev-upload: dev-upload-linux-amd64 dev-upload-linux-arm64
	echo "Done"

.PHONY: crds
crds:
	cd "${KOPS_ROOT}/hack" && go build -o "${KOPS_ROOT}/_output/bin/controller-gen" sigs.k8s.io/controller-tools/cmd/controller-gen
	"${KOPS_ROOT}/_output/bin/controller-gen" crd paths=k8s.io/kops/pkg/apis/kops/v1alpha2 output:dir=k8s/crds/ crd:crdVersions=v1

#------------------------------------------------------
# kops-controller

.PHONY: kops-controller-push
kops-controller-push: bazelisk
	DOCKER_REGISTRY=${DOCKER_REGISTRY} DOCKER_IMAGE_PREFIX=${DOCKER_IMAGE_PREFIX} KOPS_CONTROLLER_TAG=${KOPS_CONTROLLER_PUSH_TAG} ${BAZEL} run --@io_bazel_rules_go//go/config:pure --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //cmd/kops-controller:push-image-amd64
	DOCKER_REGISTRY=${DOCKER_REGISTRY} DOCKER_IMAGE_PREFIX=${DOCKER_IMAGE_PREFIX} KOPS_CONTROLLER_TAG=${KOPS_CONTROLLER_PUSH_TAG} ${BAZEL} run --@io_bazel_rules_go//go/config:pure --platforms=@io_bazel_rules_go//go/toolchain:linux_arm64 //cmd/kops-controller:push-image-arm64

.PHONY: kops-controller-manifest
kops-controller-manifest:
	docker manifest create --amend ${DOCKER_REGISTRY}/${DOCKER_IMAGE_PREFIX}kops-controller:${KOPS_CONTROLLER_PUSH_TAG} ${DOCKER_REGISTRY}/${DOCKER_IMAGE_PREFIX}kops-controller:${KOPS_CONTROLLER_PUSH_TAG}-amd64
	docker manifest create --amend ${DOCKER_REGISTRY}/${DOCKER_IMAGE_PREFIX}kops-controller:${KOPS_CONTROLLER_PUSH_TAG} ${DOCKER_REGISTRY}/${DOCKER_IMAGE_PREFIX}kops-controller:${KOPS_CONTROLLER_PUSH_TAG}-arm64
	docker manifest push --purge ${DOCKER_REGISTRY}/${DOCKER_IMAGE_PREFIX}kops-controller:${KOPS_CONTROLLER_PUSH_TAG}

#------------------------------------------------------
# kube-apiserver-healthcheck

.PHONY: kube-apiserver-healthcheck-push
kube-apiserver-healthcheck-push: bazelisk
	DOCKER_REGISTRY=${DOCKER_REGISTRY} DOCKER_IMAGE_PREFIX=${DOCKER_IMAGE_PREFIX} KUBE_APISERVER_HEALTHCHECK_TAG=${KUBE_APISERVER_HEALTHCHECK_PUSH_TAG} ${BAZEL} run --@io_bazel_rules_go//go/config:pure --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //cmd/kube-apiserver-healthcheck:push-image-amd64
	DOCKER_REGISTRY=${DOCKER_REGISTRY} DOCKER_IMAGE_PREFIX=${DOCKER_IMAGE_PREFIX} KUBE_APISERVER_HEALTHCHECK_TAG=${KUBE_APISERVER_HEALTHCHECK_PUSH_TAG} ${BAZEL} run --@io_bazel_rules_go//go/config:pure --platforms=@io_bazel_rules_go//go/toolchain:linux_arm64 //cmd/kube-apiserver-healthcheck:push-image-arm64

.PHONY: kube-apiserver-healthcheck-manifest
kube-apiserver-healthcheck-manifest:
	docker manifest create --amend ${DOCKER_REGISTRY}/${DOCKER_IMAGE_PREFIX}kube-apiserver-healthcheck:${KUBE_APISERVER_HEALTHCHECK_PUSH_TAG} ${DOCKER_REGISTRY}/${DOCKER_IMAGE_PREFIX}kube-apiserver-healthcheck:${KUBE_APISERVER_HEALTHCHECK_PUSH_TAG}-amd64
	docker manifest create --amend ${DOCKER_REGISTRY}/${DOCKER_IMAGE_PREFIX}kube-apiserver-healthcheck:${KUBE_APISERVER_HEALTHCHECK_PUSH_TAG} ${DOCKER_REGISTRY}/${DOCKER_IMAGE_PREFIX}kube-apiserver-healthcheck:${KUBE_APISERVER_HEALTHCHECK_PUSH_TAG}-arm64
	docker manifest push --purge ${DOCKER_REGISTRY}/${DOCKER_IMAGE_PREFIX}kube-apiserver-healthcheck:${KUBE_APISERVER_HEALTHCHECK_PUSH_TAG}

#------------------------------------------------------
# CloudBuild artifacts
#
# We hash some artifacts, so that we have can know that they were not modified after being built.

.PHONY: cloudbuild-artifacts
cloudbuild-artifacts:
	mkdir -p ${KOPS_ROOT}/cloudbuild/
	cd ${BAZELUPLOAD}/kops/; find . -type f | sort | xargs sha256sum > ${KOPS_ROOT}/cloudbuild/files.sha256
	cd ${KOPS_ROOT}/${BAZEL_BIN}/; find . -name '*.digest' -type f | sort | xargs grep . > ${KOPS_ROOT}/cloudbuild/image-digests
	# ${BUILDER_OUTPUT}/output is a special cloudbuild target; the first 4KB is captured securely
	cd ${KOPS_ROOT}/cloudbuild/; find -type f | sort | xargs sha256sum > ${BUILDER_OUTPUT}/output
