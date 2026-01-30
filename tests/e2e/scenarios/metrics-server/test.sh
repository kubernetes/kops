#!/usr/bin/env bash

# Copyright 2026 The Kubernetes Authors.
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

MS_VERSION=$(kubectl get deployment -n kube-system metrics-server -o jsonpath='{.spec.template.spec.containers[?(@.name=="metrics-server")].image}' | cut -d':' -f2-)

# some of the metrics only show when requested
kubectl get --raw "/apis/metrics.k8s.io/v1beta1/pods"
kubectl get --raw "/apis/metrics.k8s.io/v1beta1/nodes"
kubectl get --raw "/apis/metrics.k8s.io/v1beta1/pods"
kubectl get --raw "/apis/metrics.k8s.io/v1beta1/nodes"

cd "${GOPATH}/src/sigs.k8s.io/metrics-server" || exit
git checkout "${MS_VERSION}"
go install github.com/onsi/ginkgo/v2/ginkgo
ginkgo --junit-report="${ARTIFACTS}/junit.xml" ./test/...
