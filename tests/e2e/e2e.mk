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

.PHONY: test-e2e-aws-simple
test-e2e-install:
	cd $(KOPS_ROOT)/tests/e2e && \
		export GO111MODULE=on && \
		go get sigs.k8s.io/kubetest2@latest && \
		go install ./kubetest2-tester-kops && \
		go install ./kubetest2-kops

.PHONY: test-e2e-aws-simple-1-20
test-e2e-aws-simple-1-20: test-e2e-install
	kubetest2 kops \
		-v 2 \
		--up --down \
		--cloud-provider=aws \
		--kops-version-marker=https://storage.googleapis.com/kops-ci/bin/latest-ci-updown-green.txt \
		--kubernetes-version=v1.20.2 \
		--template-path=tests/e2e/templates/simple.yaml.tmpl \
		--test=kops \
		-- \
		--test-package-version=v1.20.2 \
		--parallel 25 \
		--skip-regex="\[Slow\]|\[Serial\]|\[Disruptive\]|\[Flaky\]|\[Feature:.+\]|\[HPA\]|Dashboard|Services.*functioning.*NodePort"
