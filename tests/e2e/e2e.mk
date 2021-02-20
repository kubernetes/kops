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
		export GO111MODULE=on && \
		go install sigs.k8s.io/kubetest2 && \
		go install ./kubetest2-tester-kops && \
		go install ./kubetest2-kops

.PHONY: test-e2e-aws-simple-1-20
test-e2e-aws-simple-1-20: test-e2e-install
	kubetest2 kops \
		-v 2 \
		--up --down \
		--cloud-provider=aws \
		--create-args="--image='099720109477/ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-20210119.1' --networking=calico --container-runtime=containerd" \
		--env=KOPS_FEATURE_FLAGS= \
		--kops-version-marker=https://storage.googleapis.com/kops-ci/bin/latest-ci-updown-green.txt \
		--kubernetes-version=https://storage.googleapis.com/kubernetes-release/release/stable.txt \
		--test=kops \
		-- \
		--ginkgo-args="--debug" \
		--test-args="-test.timeout=60m -num-nodes=0" \
		--test-package-marker=stable.txt \
		--parallel 25 \
		--skip-regex="\[Slow\]|\[Serial\]|\[Disruptive\]|\[Flaky\]|\[Feature:.+\]|\[HPA\]|Dashboard|RuntimeClass|RuntimeHandler"
