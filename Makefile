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
UPLOAD=$(BUILD)/upload
UID:=$(shell id -u)
GID:=$(shell id -g)
API_OPTIONS?=
GCFLAGS?=
OSARCH=$(shell go env GOOS)/$(shell go env GOARCH)

GOBIN=$(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN := $(shell go env GOPATH)/bin
endif

# CODEGEN_VERSION is the version of k8s.io/code-generator to use
CODEGEN_VERSION=v0.24.0


UPLOAD_CMD=$(KOPS_ROOT)/hack/upload ${UPLOAD_ARGS}

# Unexport environment variables that can affect tests and are not used in builds
unexport AWS_ACCESS_KEY_ID AWS_REGION AWS_SECRET_ACCESS_KEY AWS_SESSION_TOKEN CNI_VERSION_URL DNS_IGNORE_NS_CHECK DNSCONTROLLER_IMAGE DO_ACCESS_TOKEN GOOGLE_APPLICATION_CREDENTIALS
unexport KOPS_BASE_URL KOPS_CLUSTER_NAME KOPS_RUN_OBSOLETE_VERSION KOPS_STATE_STORE KOPS_STATE_S3_ACL KUBE_API_VERSIONS NODEUP_URL OPENSTACK_CREDENTIAL_FILE SKIP_PACKAGE_UPDATE
unexport SKIP_REGION_CHECK S3_ACCESS_KEY_ID S3_ENDPOINT S3_REGION S3_SECRET_ACCESS_KEY HCLOUD_TOKEN SCW_ACCESS_KEY SCW_SECRET_KEY SCW_DEFAULT_PROJECT_ID SCW_DEFAULT_REGION SCW_DEFAULT_ZONE YANDEX_CLOUD_CREDENTIAL_FILE


VERSION=$(shell tools/get_version.sh | grep VERSION | awk '{print $$2}')

KOPS_RELEASE_VERSION:=$(shell grep 'KOPS_RELEASE_VERSION\s*=' version.go | awk '{print $$3}' | sed -e 's_"__g')
KOPS_CI_VERSION:=$(shell grep 'KOPS_CI_VERSION\s*=' version.go | awk '{print $$3}' | sed -e 's_"__g')

# kops local location
KOPS=${DIST}/$(shell go env GOOS)/$(shell go env GOARCH)/kops

GITSHA := $(shell cd ${KOPS_ROOT}; git describe --always)

# We lock the versions of our controllers also
# We need to keep in sync with:
#   upup/models/cloudup/resources/addons/dns-controller/
DNS_CONTROLLER_TAG=1.26.0-alpha.1
DNS_CONTROLLER_PUSH_TAG=$(shell tools/get_workspace_status.sh | grep STABLE_DNS_CONTROLLER_TAG | awk '{print $$2}')
#   upup/models/cloudup/resources/addons/kops-controller.addons.k8s.io/
KOPS_CONTROLLER_TAG=1.26.0-alpha.1
KOPS_CONTROLLER_PUSH_TAG=$(shell tools/get_workspace_status.sh | grep STABLE_KOPS_CONTROLLER_TAG | awk '{print $$2}')
#   pkg/model/components/kubeapiserver/model.go
KUBE_APISERVER_HEALTHCHECK_TAG=1.26.0-alpha.1
KUBE_APISERVER_HEALTHCHECK_PUSH_TAG=$(shell tools/get_workspace_status.sh | grep STABLE_KUBE_APISERVER_HEALTHCHECK_TAG | awk '{print $$2}')

CGO_ENABLED=0
export CGO_ENABLED
BUILDFLAGS="-trimpath"


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
kops-install: kops
	cp ${DIST}/$(shell go env GOOS)/$(shell go env GOARCH)/kops* ${GOBIN}

.phony: channels-install # install channels to local $gopath/bin
channels-install: channels
	cp ${DIST}/${OSARCH}/channels ${GOPATH_1ST}/bin

.phony: nodeup-install # install channels to local $gopath/bin
nodeup-install: nodeup
	cp ${DIST}/${OSARCH}/channels ${GOPATH_1ST}/bin

.PHONY: all-install # Install all kops project binaries
all-install: all kops-install channels-install nodeup-install

.PHONY: all
all: kops protokube nodeup channels ko-kops-controller-export ko-dns-controller-export ko-kube-apiserver-healthcheck-export

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
	rm -rf tests/integration/update_cluster/*/.terraform

.PHONY: codegen
codegen:
	go build -o ${KOPS_ROOT}/_output/bin/ k8s.io/kops/upup/tools/generators/...
	${KOPS_ROOT}/_output/bin/fitask \
		--input-dirs k8s.io/kops/upup/pkg/fi/... \
		--go-header-file hack/boilerplate/boilerplate.generatego.txt \
		--output-base ${KOPS_ROOT}

.PHONY: verify-codegen
verify-codegen:
	go build -o ${KOPS_ROOT}/_output/bin/ k8s.io/kops/upup/tools/generators/...
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


.PHONY: kops
kops: crossbuild-kops-$(shell go env GOOS)-$(shell go env GOARCH)

.PHONY: crossbuild-kops-linux-amd64 crossbuild-kops-linux-arm64
crossbuild-kops-linux-amd64 crossbuild-kops-linux-arm64: crossbuild-kops-linux-%:
	mkdir -p ${DIST}/linux/$*
	GOOS=linux GOARCH=$* go build ${GCFLAGS} ${BUILDFLAGS} ${EXTRA_BUILDFLAGS} -o ${DIST}/linux/$*/kops ${LDFLAGS}"${EXTRA_LDFLAGS} -X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA}" k8s.io/kops/cmd/kops

.PHONY: crossbuild-kops-darwin-amd64 crossbuild-kops-darwin-arm64
crossbuild-kops-darwin-amd64 crossbuild-kops-darwin-arm64: crossbuild-kops-darwin-%:
	mkdir -p ${DIST}/darwin/$*
	GOOS=darwin GOARCH=$* go build ${GCFLAGS} ${BUILDFLAGS} ${EXTRA_BUILDFLAGS} -o ${DIST}/darwin/$*/kops ${LDFLAGS}"${EXTRA_LDFLAGS} -X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA}" k8s.io/kops/cmd/kops


.PHONY: crossbuild-kops-windows-amd64
crossbuild-kops-windows-amd64:
	mkdir -p ${DIST}/windows/amd64
	GOOS=windows GOARCH=amd64 go build ${GCFLAGS} ${BUILDFLAGS} ${EXTRA_BUILDFLAGS} -o ${DIST}/windows/amd64/kops.exe ${LDFLAGS}"${EXTRA_LDFLAGS} -X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA}" k8s.io/kops/cmd/kops

.PHONY: crossbuild
crossbuild: crossbuild-kops

.PHONY: crossbuild-kops
crossbuild: crossbuild-kops-linux-amd64 crossbuild-kops-linux-arm64 crossbuild-kops-darwin-amd64 crossbuild-kops-darwin-arm64 crossbuild-kops-windows-amd64

.PHONY: nodeup-amd64 nodeup-arm64
nodeup-amd64 nodeup-arm64: nodeup-%:
	mkdir -p ${DIST}/linux/$*
	GOOS=linux GOARCH=$* go build ${GCFLAGS} ${BUILDFLAGS} ${EXTRA_BUILDFLAGS} -o ${DIST}/linux/$*/nodeup ${LDFLAGS}"${EXTRA_LDFLAGS} -X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA}" k8s.io/kops/cmd/nodeup

.PHONY: nodeup
nodeup: nodeup-amd64

.PHONY: crossbuild-nodeup
crossbuild-nodeup: nodeup-amd64 nodeup-arm64

.PHONY: protokube-amd64 protokube-arm64
protokube-amd64 protokube-arm64: protokube-%:
	mkdir -p ${DIST}/linux/$*
	GOOS=linux GOARCH=$* go build -tags=peer_name_alternative,peer_name_hash ${GCFLAGS} ${BUILDFLAGS} ${EXTRA_BUILDFLAGS} -o ${DIST}/linux/$*/protokube ${LDFLAGS}"${EXTRA_LDFLAGS} -X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA}" k8s.io/kops/protokube/cmd/protokube

.PHONY: protokube
protokube: protokube-amd64

.PHONY: crossbuild-protokube
crossbuild-protokube: protokube-amd64 protokube-arm64

.PHONY: channels-amd64 channels-arm64
channels-amd64 channels-arm64: channels-%:
	mkdir -p ${DIST}/linux/$*
	GOOS=linux GOARCH=$* go build ${GCFLAGS} ${BUILDFLAGS} ${EXTRA_BUILDFLAGS} -o ${DIST}/linux/$*/channels ${LDFLAGS}"${EXTRA_LDFLAGS} -X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA}" k8s.io/kops/channels/cmd/channels

.PHONY: channels
channels: channels-amd64

.PHONY: crossbuild-channels
crossbuild-channels: channels-amd64 channels-arm64

.PHONY: upload
upload: version-dist # Upload kops to S3
	aws s3 sync --acl public-read ${UPLOAD}/ ${S3_BUCKET}

# gcs-upload builds kops and uploads to GCS
.PHONY: gcs-upload
gcs-upload: gsutil version-dist
	@echo "== Uploading kops =="
	gsutil -h "Cache-Control:private, max-age=0, no-transform" -m cp -n -r ${UPLOAD}/kops/* ${GCS_LOCATION}

# gcs-upload-tag runs gcs-upload to upload, then uploads a version-marker to LATEST_FILE
.PHONY: gcs-upload-and-tag
gcs-upload-and-tag: gsutil gcs-upload
	echo "${GCS_URL}${VERSION}" > ${UPLOAD}/latest.txt
	gsutil -h "Cache-Control:private, max-age=0, no-transform" cp ${UPLOAD}/latest.txt ${GCS_LOCATION}${LATEST_FILE}

# gcs-publish-ci is the entry point for CI testing
# In CI testing, always upload the CI version.
.PHONY: gcs-publish-ci
gcs-publish-ci: VERSION := ${KOPS_CI_VERSION}+${GITSHA}
gcs-publish-ci: gsutil version-dist-ci
	@echo "== Uploading kops =="
	gsutil -h "Cache-Control:private, max-age=0, no-transform" -m cp -n -r ${UPLOAD}/kops/* ${GCS_LOCATION}
	echo "VERSION: ${VERSION}"
	echo "${GCS_URL}/${VERSION}" > ${UPLOAD}/${LATEST_FILE}
	gsutil -h "Cache-Control:private, max-age=0, no-transform" cp ${UPLOAD}/${LATEST_FILE} ${GCS_LOCATION}

.PHONY: gen-cli-docs
gen-cli-docs: kops # Regenerate CLI docs
	KOPS_STATE_STORE= \
	KOPS_FEATURE_FLAGS= \
	${DIST}/${OSARCH}/kops gen-cli-docs --out docs/cli

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

.PHONY: ${NODEUP}
${NODEUP}:
	go build ${GCFLAGS} ${EXTRA_BUILDFLAGS} ${LDFLAGS}"${EXTRA_LDFLAGS} -X k8s.io/kops.Version=${VERSION} -X k8s.io/kops.GitVersion=${GITSHA}" -o $@ k8s.io/kops/cmd/nodeup

.PHONY: dns-controller-push
dns-controller-push: ko-dns-controller-push

.PHONY: ko-dns-controller-push
ko-dns-controller-push: ko
	KO_DOCKER_REPO="${DOCKER_REGISTRY}/${DOCKER_IMAGE_PREFIX}dns-controller" GOFLAGS="-tags=peer_name_alternative,peer_name_hash" ko build --tags ${DNS_CONTROLLER_PUSH_TAG} --platform=linux/amd64,linux/arm64 --bare ./dns-controller/cmd/dns-controller/

# --------------------------------------------------
# development targets

.PHONY: gomod
gomod:
	go mod tidy
	go mod vendor
	cd tests/e2e; go mod tidy
	cd hack; go mod tidy

.PHONY: goget
goget:
	go get $(shell go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -mod=mod -m all)

.PHONY: depup
depup: goget gomod gen-cli-docs

.PHONY: gofmt
gofmt:
	find $(KOPS_ROOT) -name "*.go" | grep -v vendor | xargs gofmt -w -s

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
verify-gofmt:
	hack/verify-gofmt.sh

.PHONY: verify-gomod
verify-gomod:
	hack/verify-gomod.sh

# find release notes, remove PR titles and output the rest to .build, then run misspell on all files
.PHONY: verify-misspelling
verify-misspelling:
	hack/verify-spelling.sh

.PHONY: verify-gendocs
verify-gendocs: kops
	@TMP_DOCS="$$(mktemp -d)"; \
	'${KOPS}' gen-cli-docs --out "$$TMP_DOCS"; \
	\
	if ! diff -r "$$TMP_DOCS" '${KOPS_ROOT}/docs/cli'; then \
	     echo "FAIL: make verify-gendocs failed, as the generated markdown docs are out of date." 1>&2; \
	     echo "FAIL: Please run the following command: make gen-cli-docs." 1>&2; \
	     exit 1; \
	fi
	@echo "cli docs up-to-date"

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
ci: govet verify-gofmt verify-crds verify-gomod verify-goimports verify-boilerplate verify-versions verify-misspelling verify-shellcheck verify-golangci-lint verify-terraform nodeup examples test | verify-gendocs verify-apimachinery verify-codegen
	echo "Done!"

# we skip tasks that are covered by other jobs
.PHONY: quick-ci
quick-ci: verify-crds verify-goimports govet verify-boilerplate verify-versions verify-misspelling verify-shellcheck | verify-gendocs verify-apimachinery verify-codegen
	echo "Done!"

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
apimachinery-codegen-conversion: export GOPATH=
apimachinery-codegen-conversion:
	go run k8s.io/code-generator/cmd/conversion-gen@${CODEGEN_VERSION} --skip-unsafe=true --v=0 --input-dirs ./pkg/apis/kops/v1alpha2 \
		 --output-base=./ --output-file-base=zz_generated.conversion \
		 --go-header-file "hack/boilerplate/boilerplate.generatego.txt"
	grep 'requires manual conversion' ${KOPS_ROOT}/pkg/apis/kops/v1alpha2/zz_generated.conversion.go ; [ $$? -eq 1 ]
	go run k8s.io/code-generator/cmd/conversion-gen@${CODEGEN_VERSION} --skip-unsafe=true --v=0 --input-dirs ./pkg/apis/kops/v1alpha3 \
		 --output-base=./ --output-file-base=zz_generated.conversion \
		 --go-header-file "hack/boilerplate/boilerplate.generatego.txt"
	grep 'requires manual conversion' ${KOPS_ROOT}/pkg/apis/kops/v1alpha3/zz_generated.conversion.go ; [ $$? -eq 1 ]

.PHONY: apimachinery-codegen-deepcopy
apimachinery-codegen-deepcopy: export GOPATH=
apimachinery-codegen-deepcopy:
	go run k8s.io/code-generator/cmd/deepcopy-gen@${CODEGEN_VERSION} --v=0 --input-dirs ./pkg/apis/kops \
		 --output-base=./ --output-file-base=zz_generated.deepcopy \
		 --go-header-file "hack/boilerplate/boilerplate.generatego.txt"
	go run k8s.io/code-generator/cmd/deepcopy-gen@${CODEGEN_VERSION} --v=0 --input-dirs ./pkg/apis/kops/v1alpha2 \
		 --output-base=./ --output-file-base=zz_generated.deepcopy \
		 --go-header-file "hack/boilerplate/boilerplate.generatego.txt"
	go run k8s.io/code-generator/cmd/deepcopy-gen@${CODEGEN_VERSION} --v=0 --input-dirs ./pkg/apis/kops/v1alpha3 \
		 --output-base=./ --output-file-base=zz_generated.deepcopy \
		 --go-header-file "hack/boilerplate/boilerplate.generatego.txt"

.PHONY: apimachinery-codegen-defaulter
apimachinery-codegen-defaulter: export GOPATH=
apimachinery-codegen-defaulter:
	go run k8s.io/code-generator/cmd/defaulter-gen@${CODEGEN_VERSION} --v=0 --input-dirs ./pkg/apis/kops/v1alpha2 \
		 --output-base=./ --output-file-base=zz_generated.defaults \
		 --go-header-file "hack/boilerplate/boilerplate.generatego.txt"
	go run k8s.io/code-generator/cmd/defaulter-gen@${CODEGEN_VERSION} --v=0 --input-dirs ./pkg/apis/kops/v1alpha3 \
		 --output-base=./ --output-file-base=zz_generated.defaults \
		 --go-header-file "hack/boilerplate/boilerplate.generatego.txt"

.PHONY: apimachinery-codegen-client
apimachinery-codegen-client: export GOPATH=
apimachinery-codegen-client: TMPDIR := $(shell mktemp -d)
apimachinery-codegen-client:
	go run k8s.io/code-generator/cmd/client-gen@${CODEGEN_VERSION} --v=0 \
		 --input-base=k8s.io/kops/pkg/apis --input-dirs=. --input="kops/,kops/v1alpha2,kops/v1alpha3" \
		 --output-package=k8s.io/kops/pkg/client/clientset_generated/ --output-base=$(TMPDIR) \
		 --go-header-file "hack/boilerplate/boilerplate.generatego.txt"
	go run k8s.io/code-generator/cmd/client-gen@${CODEGEN_VERSION} --v=0 --clientset-name="clientset" \
		 --input-base=k8s.io/kops/pkg/apis --input-dirs=. --input="kops/,kops/v1alpha2,kops/v1alpha3" \
		 --output-package=k8s.io/kops/pkg/client/clientset_generated/ --output-base=$(TMPDIR) \
		 --go-header-file "hack/boilerplate/boilerplate.generatego.txt"
	cp -r $(TMPDIR)/k8s.io/kops/pkg .
	rm -rf $(TMPDIR)

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

.PHONY: gsutil
gsutil:
	hack/install-gsutil.sh

.PHONY: ko
ko:
	hack/install-ko.sh

.PHONY: check-markdown-links
check-markdown-links:
	docker run -t -v $$PWD:/tmp \
		-e LC_ALL=C.UTF-8 \
		-e LANG=en_US.UTF-8 \
		-e LANGUAGE=en_US.UTF-8 \
		rubygem/awesome_bot --allow-dupe --allow-redirect \
		$(shell find $$PWD -name "*.md" -mindepth 1 -printf '%P\n' | grep -v vendor | grep -v Changelog.md)

#-----------------------------------------------------------

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
	KO_DOCKER_REPO="registry.k8s.io/kops" GOFLAGS="-tags=peer_name_alternative,peer_name_hash" ko build --tags ${DNS_CONTROLLER_TAG} --platform=linux/$* -B --push=false --tarball=${IMAGES}/dns-controller-$*.tar ./dns-controller/cmd/dns-controller
	gzip -f ${IMAGES}/dns-controller-$*.tar
	tools/sha256 ${IMAGES}/dns-controller-$*.tar.gz ${IMAGES}/dns-controller-$*.tar.gz.sha256

.PHONY: ko-dns-controller-export
ko-dns-controller-export: ko-dns-controller-export-linux-amd64 ko-dns-controller-export-linux-arm64
	echo "Done exporting dns-controller images"

.PHONY: version-dist
version-dist: dev-version-dist-amd64 dev-version-dist-arm64 crossbuild
	mkdir -p ${UPLOAD}/kops/${VERSION}/linux/amd64/
	mkdir -p ${UPLOAD}/kops/${VERSION}/linux/arm64/
	mkdir -p ${UPLOAD}/kops/${VERSION}/darwin/amd64/
	mkdir -p ${UPLOAD}/kops/${VERSION}/darwin/arm64/
	mkdir -p ${UPLOAD}/kops/${VERSION}/windows/amd64/
	cp ${DIST}/linux/amd64/kops ${UPLOAD}/kops/${VERSION}/linux/amd64/kops
	tools/sha256 ${UPLOAD}/kops/${VERSION}/linux/amd64/kops ${UPLOAD}/kops/${VERSION}/linux/amd64/kops.sha256
	cp ${DIST}/linux/arm64/kops ${UPLOAD}/kops/${VERSION}/linux/arm64/kops
	tools/sha256 ${UPLOAD}/kops/${VERSION}/linux/arm64/kops ${UPLOAD}/kops/${VERSION}/linux/arm64/kops.sha256
	cp ${DIST}/darwin/amd64/kops ${UPLOAD}/kops/${VERSION}/darwin/amd64/kops
	tools/sha256 ${UPLOAD}/kops/${VERSION}/darwin/amd64/kops ${UPLOAD}/kops/${VERSION}/darwin/amd64/kops.sha256
	cp ${DIST}/darwin/arm64/kops ${UPLOAD}/kops/${VERSION}/darwin/arm64/kops
	tools/sha256 ${UPLOAD}/kops/${VERSION}/darwin/arm64/kops ${UPLOAD}/kops/${VERSION}/darwin/arm64/kops.sha256
	cp ${DIST}/windows/amd64/kops.exe ${UPLOAD}/kops/${VERSION}/windows/amd64/kops.exe
	tools/sha256 ${UPLOAD}/kops/${VERSION}/windows/amd64/kops.exe ${UPLOAD}/kops/${VERSION}/windows/amd64/kops.exe.sha256

# This target skips arm64 and windows kops binary since CI does not use them
.PHONY: version-dist-ci
version-dist-ci: dev-version-dist-amd64 dev-version-dist-arm64 crossbuild-kops-linux-amd64
	mkdir -p ${UPLOAD}/kops/${VERSION}/linux/amd64/
	cp ${DIST}/linux/amd64/kops ${UPLOAD}/kops/${VERSION}/linux/amd64/kops
	tools/sha256 ${UPLOAD}/kops/${VERSION}/linux/amd64/kops ${UPLOAD}/kops/${VERSION}/linux/amd64/kops.sha256


# prow-postsubmit is run by the prow postsubmit job
# It uploads a build to a staging directory, which in theory we can publish as a release
.PHONY: prow-postsubmit
prow-postsubmit: version-dist
	${UPLOAD_CMD} ${UPLOAD}/kops/${VERSION}/ ${UPLOAD_DEST}/${VERSION}/

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

# dev-upload-nodeup uploads nodeup
.PHONY: version-dist-nodeup version-dist-nodeup-amd64 version-dist-nodeup-arm64
version-dist-nodeup: version-dist-nodeup-amd64 version-dist-nodeup-arm64

version-dist-nodeup-amd64 version-dist-nodeup-arm64: version-dist-nodeup-%: nodeup-%
	mkdir -p ${UPLOAD}/kops/${VERSION}/linux/$*/
	cp -fp ${DIST}/linux/$*/nodeup ${UPLOAD}/kops/${VERSION}/linux/$*/nodeup
	tools/sha256 ${UPLOAD}/kops/${VERSION}/linux/$*/nodeup ${UPLOAD}/kops/${VERSION}/linux/$*/nodeup.sha256

.PHONY: dev-upload-nodeup
dev-upload-nodeup: version-dist-nodeup
	${UPLOAD_CMD} ${UPLOAD}/ ${UPLOAD_DEST}

.PHONY: dev-upload-nodeup-amd64 dev-upload-nodeup-arm64
dev-upload-nodeup-amd64 dev-upload-nodeup-arm64: dev-upload-nodeup-%: version-dist-nodeup-%
	${UPLOAD_CMD} ${UPLOAD}/ ${UPLOAD_DEST}

# dev-upload-protokube uploads protokube
.PHONY: version-dist-protokube version-dist-protokube-amd64 version-dist-protokube-arm64
version-dist-protokube: version-dist-protokube-amd64 version-dist-protokube-arm64

version-dist-protokube-amd64 version-dist-protokube-arm64: version-dist-protokube-%: protokube-%
	mkdir -p ${UPLOAD}/kops/${VERSION}/linux/$*/
	cp -fp ${DIST}/linux/$*/protokube ${UPLOAD}/kops/${VERSION}/linux/$*/protokube
	tools/sha256 ${UPLOAD}/kops/${VERSION}/linux/$*/protokube ${UPLOAD}/kops/${VERSION}/linux/$*/protokube.sha256

.PHONY: dev-upload-protokube
dev-upload-protokube: version-dist-protokube
	${UPLOAD_CMD} ${UPLOAD}/ ${UPLOAD_DEST}

.PHONY: dev-upload-protokube-amd64 dev-upload-protokube-arm64
dev-upload-protokube-amd64 dev-upload-protokube-arm64: dev-upload-protokube-%: version-dist-protokube-%
	${UPLOAD_CMD} ${UPLOAD}/ ${UPLOAD_DEST}

# dev-upload-channels uploads channels
.PHONY: version-dist-channels version-dist-channels-amd64 version-dist-channels-arm64
version-dist-channels: version-dist-channels-amd64 version-dist-channels-arm64

version-dist-channels-amd64 version-dist-channels-arm64: version-dist-channels-%: channels-%
	mkdir -p ${UPLOAD}/kops/${VERSION}/linux/$*/
	cp -fp ${DIST}/linux/$*/channels ${UPLOAD}/kops/${VERSION}/linux/$*/channels
	tools/sha256 ${UPLOAD}/kops/${VERSION}/linux/$*/channels ${UPLOAD}/kops/${VERSION}/linux/$*/channels.sha256

.PHONY: dev-upload-channels
dev-upload-channels: version-dist-channels
	${UPLOAD_CMD} ${PLOAD}/ ${UPLOAD_DEST}

.PHONY: dev-upload-channels-amd64 dev-upload-channels-arm64
dev-upload-channels-amd64 dev-upload-channels-arm64: dev-upload-channels-%: version-dist-channels-%
	${UPLOAD_CMD} ${UPLOAD}/ ${UPLOAD_DEST}

# dev-upload-kops-controller uploads kops-controller
.PHONY: version-dist-kops-controller version-dist-kops-controller-amd64 version-dist-kops-controller-arm64
version-dist-kops-controller: version-dist-kops-controller-amd64 version-dist-kops-controller-arm64

version-dist-kops-controller-amd64 version-dist-kops-controller-arm64: version-dist-kops-controller-%: ko-kops-controller-export-linux-%
	mkdir -p ${UPLOAD}/kops/${VERSION}/images/
	cp -fp ${IMAGES}/kops-controller-$*.tar.gz ${UPLOAD}/kops/${VERSION}/images/kops-controller-$*.tar.gz
	cp -fp ${IMAGES}/kops-controller-$*.tar.gz.sha256 ${UPLOAD}/kops/${VERSION}/images/kops-controller-$*.tar.gz.sha256

.PHONY: dev-upload-kops-controller
dev-upload-kops-controller: version-dist-kops-controller
	${UPLOAD_CMD} ${UPLOAD}/ ${UPLOAD_DEST}

.PHONY: dev-upload-kops-controller-amd64 dev-upload-kops-controller-arm64
dev-upload-kops-controller-amd64 dev-upload-kops-controller-arm64: dev-upload-kops-controller-%: version-dist-kops-controller-%
	${UPLOAD_CMD} ${UPLOAD}/ ${UPLOAD_DEST}

# dev-upload-kube-apiserver-healthcheck uploads kube-apiserver-healthcheck
.PHONY: version-dist-kube-apiserver-healthcheck version-dist-kube-apiserver-healthcheck-amd64 version-dist-kube-apiserver-healthcheck-arm64
version-dist-kube-apiserver-healthcheck: version-dist-kube-apiserver-healthcheck-amd64 version-dist-kube-apiserver-healthcheck-arm64

version-dist-kube-apiserver-healthcheck-amd64 version-dist-kube-apiserver-healthcheck-arm64: version-dist-kube-apiserver-healthcheck-%: ko-kube-apiserver-healthcheck-export-linux-%
	mkdir -p ${UPLOAD}/kops/${VERSION}/images/
	cp -fp ${IMAGES}/kube-apiserver-healthcheck-$*.tar.gz ${UPLOAD}/kops/${VERSION}/images/kube-apiserver-healthcheck-$*.tar.gz
	cp -fp ${IMAGES}/kube-apiserver-healthcheck-$*.tar.gz.sha256 ${UPLOAD}/kops/${VERSION}/images/kube-apiserver-healthcheck-$*.tar.gz.sha256

.PHONY: dev-upload-kube-apiserver-healthcheck
dev-upload-kube-apiserver-healthcheck: version-dist-kube-apiserver-healthcheck
	${UPLOAD_CMD} ${UPLOAD}/ ${UPLOAD_DEST}

.PHONY: dev-upload-kube-apiserver-healthcheck-amd64 dev-upload-kube-apiserver-healthcheck-arm64
dev-upload-kube-apiserver-healthcheck-amd64 dev-upload-kube-apiserver-healthcheck-arm64: dev-upload-kube-apiserver-healthcheck-%: version-dist-kube-apiserver-healthcheck-%
	${UPLOAD_CMD} ${UPLOAD}/ ${UPLOAD_DEST}

# dev-upload-dns-controller uploads dns-controller
.PHONY: version-dist-dns-controller version-dist-dns-controller-amd64 version-dist-dns-controller-arm64
version-dist-dns-controller: version-dist-dns-controller-amd64 version-dist-dns-controller-arm64

version-dist-dns-controller-amd64 version-dist-dns-controller-arm64: version-dist-dns-controller-%: ko-dns-controller-export-linux-%
	mkdir -p ${UPLOAD}/kops/${VERSION}/images/
	cp -fp ${IMAGES}/dns-controller-$*.tar.gz ${UPLOAD}/kops/${VERSION}/images/dns-controller-$*.tar.gz
	cp -fp ${IMAGES}/dns-controller-$*.tar.gz.sha256 ${UPLOAD}/kops/${VERSION}/images/dns-controller-$*.tar.gz.sha256

.PHONY: dev-upload-dns-controller
dev-upload-dns-controller: version-dist-dns-controller
	${UPLOAD_CMD} ${UPLOAD}/ ${UPLOAD_DEST}

.PHONY: dev-upload-dns-controller-amd64 dev-upload-dns-controller-arm64
dev-upload-dns-controller-amd64 dev-upload-dns-controller-arm64: dev-upload-dns-controller-%: version-dist-dns-controller-%
	${UPLOAD_CMD} ${UPLOAD}/ ${UPLOAD_DEST}

# dev-upload-linux-amd64 does a faster build and uploads to GCS / S3
.PHONY: dev-version-dist dev-version-dist-amd64 dev-version-dist-arm64
dev-version-dist: dev-version-dist-amd64 dev-version-dist-arm64

dev-version-dist-amd64 dev-version-dist-arm64: dev-version-dist-%: version-dist-nodeup-% version-dist-channels-% version-dist-protokube-% version-dist-kops-controller-% version-dist-kube-apiserver-healthcheck-% version-dist-dns-controller-%

.PHONY: dev-upload-linux-amd64 dev-upload-linux-arm64
dev-upload-linux-amd64 dev-upload-linux-arm64: dev-upload-linux-%: dev-version-dist-%
	${UPLOAD_CMD} ${UPLOAD}/ ${UPLOAD_DEST}

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
kops-controller-push: ko-kops-controller-push

.PHONY: ko-kops-controller-push
ko-kops-controller-push: ko
	KO_DOCKER_REPO="${DOCKER_REGISTRY}/${DOCKER_IMAGE_PREFIX}kops-controller" ko build --tags ${KOPS_CONTROLLER_PUSH_TAG} --platform=linux/amd64,linux/arm64 --bare ./cmd/kops-controller/

#------------------------------------------------------
# kube-apiserver-healthcheck

.PHONY: kube-apiserver-healthcheck-push
kube-apiserver-healthcheck-push: ko-kube-apiserver-healthcheck-push

.PHONY: ko-kube-apiserver-healthcheck-push
ko-kube-apiserver-healthcheck-push: ko
	KO_DOCKER_REPO="${DOCKER_REGISTRY}/${DOCKER_IMAGE_PREFIX}kube-apiserver-healthcheck" ko build --tags ${KUBE_APISERVER_HEALTHCHECK_PUSH_TAG} --platform=linux/amd64,linux/arm64 --bare ./cmd/kube-apiserver-healthcheck/

#------------------------------------------------------
# CloudBuild artifacts
#
# We hash some artifacts, so that we have can know that they were not modified after being built.

.PHONY: cloudbuild-artifacts
cloudbuild-artifacts:
	mkdir -p ${KOPS_ROOT}/cloudbuild/
	cd ${UPLOAD}/kops/; find . -type f | sort | xargs sha256sum > ${KOPS_ROOT}/cloudbuild/files.sha256
	# cd ${KOPS_ROOT}/${BAZEL_BIN}/; find . -name '*.digest' -type f | sort | xargs grep . > ${KOPS_ROOT}/cloudbuild/image-digests
	# ${BUILDER_OUTPUT}/output is a special cloudbuild target; the first 4KB is captured securely
	cd ${KOPS_ROOT}/cloudbuild/; find -type f | sort | xargs sha256sum > ${BUILDER_OUTPUT}/output
