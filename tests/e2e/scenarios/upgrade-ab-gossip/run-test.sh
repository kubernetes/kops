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

REPO_ROOT=$(git rev-parse --show-toplevel);

case "${CLOUD_PROVIDER:-aws}" in
    aws)
        # AWS deployer composes <jobname>.${KOPS_DNS_DOMAIN}; force a gossip suffix.
        export KOPS_DNS_DOMAIN="${KOPS_DNS_DOMAIN:-tests-kops-aws.k8s.local}"
        if [[ "${KOPS_DNS_DOMAIN}" != *.k8s.local ]]; then
            >&2 echo "KOPS_DNS_DOMAIN must end in .k8s.local for gossip; got ${KOPS_DNS_DOMAIN}"
            exit 1
        fi
        ;;
    azure)
        # Azure deployer's defaultClusterName produces no domain suffix; we must
        # set CLUSTER_NAME explicitly to a gossip-style name.
        if [[ -z "${CLUSTER_NAME-}" ]]; then
            if [[ -z "${JOB_NAME-}" || -z "${BUILD_ID-}" ]]; then
                >&2 echo "set CLUSTER_NAME to a *.k8s.local name, or run inside a Prow job (JOB_NAME and BUILD_ID required)"
                exit 1
            fi
            if [[ "${JOB_TYPE-}" == "presubmit" ]]; then
                if [[ -z "${PULL_NUMBER-}" ]]; then
                    >&2 echo "PULL_NUMBER must be set when JOB_TYPE=presubmit"
                    exit 1
                fi
                export CLUSTER_NAME="e2e-pr${PULL_NUMBER}.${JOB_NAME}.k8s.local"
            else
                export CLUSTER_NAME="e2e-${JOB_NAME}.k8s.local"
            fi
        fi
        if [[ "${CLUSTER_NAME}" != *.k8s.local ]]; then
            >&2 echo "CLUSTER_NAME must end in .k8s.local for gossip; got ${CLUSTER_NAME}"
            exit 1
        fi
        ;;
    gce)
        # GCE deployer's defaultClusterName already produces a .k8s.local suffix
        # when KOPS_DNS_DOMAIN is unset; nothing extra to do.
        ;;
    *)
        >&2 echo "unsupported CLOUD_PROVIDER for upgrade-ab-gossip: ${CLOUD_PROVIDER}"
        exit 1
        ;;
esac

source "${REPO_ROOT}"/tests/e2e/scenarios/lib/common.sh
source "${REPO_ROOT}"/tests/e2e/scenarios/lib/upgrade.sh

# Hybrid worker bootstrap (#18245) is gated on the API load balancer:
# AWS needs an NLB; Azure / GCE accept any LB.
case "${CLOUD_PROVIDER}" in
    aws)
        lb_args="--api-loadbalancer-type=public --api-loadbalancer-class=network"
        ;;
    *)
        lb_args="--api-loadbalancer-type=public"
        ;;
esac

kops-upgrade "--networking cilium ${lb_args} ${KOPS_EXTRA_FLAGS:-}"
