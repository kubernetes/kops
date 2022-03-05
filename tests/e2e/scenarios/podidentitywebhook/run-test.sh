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
TEST_ROOT="${REPO_ROOT}/tests/e2e/scenarios/podidentitywebhook"

# shellcheck disable=SC2034
KOPS_TEMPLATE="${TEST_ROOT}/cluster.yaml.tmpl"

kops-acquire-latest

kops-up

kubectl apply -f "${TEST_ROOT}"/pod.yaml

kubectl -n default wait --for=condition=Ready pod/pod-identity-webhook-test

# This command will exit code 253 if there are no credentials
kubectl exec -it -n default pod-identity-webhook-test -- aws sts get-caller-identity



