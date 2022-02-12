#!/usr/bin/env bash

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

REPO_ROOT=$(git rev-parse --show-toplevel);
source "${REPO_ROOT}"/tests/e2e/scenarios/lib/common.sh

kops-acquire-latest

OVERRIDES="${OVERRIDES-} --override=cluster.spec.metricsServer.enabled=true"
OVERRIDES="$OVERRIDES --override=cluster.spec.certManager.enabled=true"

kops-up

# shellcheck disable=SC2164
cd "$(mktemp -dt kops.XXXXXXXXX)"

git clone --branch v0.6.0 https://github.com/kubernetes-sigs/metrics-server.git .

# some of the metrics only show when requested
kubectl get --raw "/apis/metrics.k8s.io/v1beta1/pods"
kubectl get --raw "/apis/metrics.k8s.io/v1beta1/nodes"
kubectl get --raw "/apis/metrics.k8s.io/v1beta1/pods"
kubectl get --raw "/apis/metrics.k8s.io/v1beta1/nodes"


# shellcheck disable=SC2164
go test -v ./test/e2e_test.go -count=1
