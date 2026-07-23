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

# Like upgrade-ab, but on a gossip cluster — exercises hybrid worker bootstrap (#18245).

REPO_ROOT=$(git rev-parse --show-toplevel);

# k8s.local suffix triggers gossip mode across all clouds.
export KOPS_DNS_DOMAIN="${KOPS_DNS_DOMAIN:-k8s.local}"
if [[ "${KOPS_DNS_DOMAIN}" != "k8s.local" && "${KOPS_DNS_DOMAIN}" != *.k8s.local ]]; then
    >&2 echo "KOPS_DNS_DOMAIN must be (or end in) k8s.local for gossip; got ${KOPS_DNS_DOMAIN}"
    exit 1
fi

source "${REPO_ROOT}"/tests/e2e/scenarios/lib/common.sh
source "${REPO_ROOT}"/tests/e2e/scenarios/lib/upgrade.sh

# Plain clusters only; the many-addons template is AWS-specific.
unset KOPS_TEMPLATE

# Gossip needs an NLB on AWS, any LB elsewhere.
case "${CLOUD_PROVIDER}" in
    aws)
        cloud_args="--api-loadbalancer-type=public --api-loadbalancer-class=network"
        ;;
    gce)
        cloud_args="--api-loadbalancer-type=public --gce-service-account=default"
        ;;
    *)
        cloud_args="--api-loadbalancer-type=public"
        ;;
esac

# Drop dns-controller's priorityClassName so kops `validate cluster` ignores the pod.
# The new image side-loads only on newly bootstrapped nodes, so until the rolling
# update replaces the old ones the pod stays Pending, and would fail validation.
override="--override=spec.externalDns.priorityClassName="

kops-upgrade "--networking cilium --dns=private ${cloud_args} ${override} ${KOPS_EXTRA_FLAGS:-}"
