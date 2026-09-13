# Copyright 2021 The Kubernetes Authors.
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

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

.PHONY: test-e2e-install
test-e2e-install:
	cd $(KOPS_ROOT)/tests/e2e && \
		go install ./kubetest2-tester-kops && \
		go install ./kubetest2-kops

.PHONY: build-e2e-binaries
build-e2e-binaries: build-e2e-binaries-amd64 build-e2e-binaries-arm64 build-e2e-binaries-s390x build-e2e-binaries-ppc64le build-e2e-binaries-riscv64 build-e2e-binaries-darwin-arm64

.PHONY: build-e2e-binaries-amd64 build-e2e-binaries-arm64 build-e2e-binaries-s390x build-e2e-binaries-ppc64le build-e2e-binaries-riscv64
build-e2e-binaries-amd64 build-e2e-binaries-arm64 build-e2e-binaries-s390x build-e2e-binaries-ppc64le build-e2e-binaries-riscv64: build-e2e-binaries-%:
	mkdir -p "$(KOPS_ROOT)/.build/dist/kubetest2/linux/$*"
	cd "$(KOPS_ROOT)/tests/e2e" && \
		CGO_ENABLED=0 GOOS=linux GOARCH=$* go build -o "$(KOPS_ROOT)/.build/dist/kubetest2/linux/$*/" ./kubetest2-kops ./kubetest2-tester-kops

.PHONY: build-e2e-binaries-darwin-arm64
build-e2e-binaries-darwin-arm64:
	mkdir -p "$(KOPS_ROOT)/.build/dist/kubetest2/darwin/arm64"
	cd "$(KOPS_ROOT)/tests/e2e" && \
		CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o "$(KOPS_ROOT)/.build/dist/kubetest2/darwin/arm64/" ./kubetest2-kops ./kubetest2-tester-kops

.PHONY: upload-e2e-binaries
upload-e2e-binaries: gcloud build-e2e-binaries
	gcloud storage cp --custom-metadata="Surrogate-Key=kops-kubetest2" --cache-control="private, max-age=0, no-transform" \
		--recursive "$(KOPS_ROOT)/.build/dist/kubetest2/"* "$(GCS_LOCATION)kubetest2/"
