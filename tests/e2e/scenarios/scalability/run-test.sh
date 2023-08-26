#!/usr/bin/env bash

# Copyright 2023 The Kubernetes Authors.
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

set -e
set -x

REPO_ROOT=$(git rev-parse --show-toplevel)
cd "${REPO_ROOT}"
cd ..
WORKSPACE=$(pwd)
cd "${WORKSPACE}/kops"

# Create bindir
BINDIR="${WORKSPACE}/bin"
export PATH="${BINDIR}:${PATH}"
mkdir -p "${BINDIR}"

# Build kubetest-2 kOps support
pushd "${WORKSPACE}/kops"
GOBIN=${BINDIR} make test-e2e-install
popd

# Setup our cleanup function; as we allocate resources we set a variable to indicate they should be cleaned up
function cleanup {
  # shellcheck disable=SC2153
  if [[ "${DELETE_CLUSTER:-}" == "true" ]]; then
      kubetest2 kops "${KUBETEST2_ARGS[@]}" --down || echo "kubetest2 down failed"
  fi
}
trap cleanup EXIT

# Default cluster name
SCRIPT_NAME=$(basename "$(dirname "$0")")
if [[ -z "${CLUSTER_NAME:-}" ]]; then
  CLUSTER_NAME="${SCRIPT_NAME}.k8s.local"
fi
echo "CLUSTER_NAME=${CLUSTER_NAME}"

if [[ -z "${K8S_VERSION:-}" ]]; then
  K8S_VERSION="$(curl -s -L https://dl.k8s.io/release/stable.txt)"
fi

# Download latest prebuilt kOps
if [[ -z "${KOPS_BASE_URL:-}" ]]; then
  KOPS_BASE_URL="$(curl -s https://storage.googleapis.com/kops-ci/bin/latest-ci-updown-green.txt)"
fi
export KOPS_BASE_URL

KOPS_BIN=${BINDIR}/kops
wget -qO "${KOPS_BIN}" "$KOPS_BASE_URL/$(go env GOOS)/$(go env GOARCH)/kops"
chmod +x "${KOPS_BIN}"

# Default cloud provider to aws
if [[ -z "${CLOUD_PROVIDER:-}" ]]; then
  CLOUD_PROVIDER="aws"
fi
echo "CLOUD_PROVIDER=${CLOUD_PROVIDER}"

# KOPS_STATE_STORE holds metadata about the clusters we create
if [[ -z "${KOPS_STATE_STORE:-}" ]]; then
  echo "Must specify KOPS_STATE_STORE"
  exit 1
fi
echo "KOPS_STATE_STORE=${KOPS_STATE_STORE}"
export KOPS_STATE_STORE


if [[ -z "${ADMIN_ACCESS:-}" ]]; then
  ADMIN_ACCESS="0.0.0.0/0" # Or use your IPv4 with /32
fi
echo "ADMIN_ACCESS=${ADMIN_ACCESS}"

# cilium does not yet pass conformance tests (shared hostport test)
#create_args="--networking cilium"
create_args=()
create_args=("--network-cidr=10.0.0.0/16")
create_args=("--network-cidr=10.1.0.0/16")
create_args=("--network-cidr=10.2.0.0/16")
create_args=("--network-cidr=10.3.0.0/16")
create_args=("--network-cidr=10.4.0.0/16")
create_args+=("--networking=${CNI_PLUGIN:-calico}")
if [[ "${CNI_PLUGIN}" == "amazonvpc" ]]; then
create_args+=("--set spec.networking.amazonVPC.env=ENABLE_PREFIX_DELEGATION=true")
fi
# INSTANCE_IMAGE configures the image used for control plane and kube nodes
create_args+=("--image=${INSTANCE_IMAGE:-ssm:/aws/service/canonical/ubuntu/server/20.04/stable/current/arm64/hvm/ebs-gp2/ami-id}")
create_args+=("--etcd-clusters=main")
create_args+=("--set spec.etcdClusters[0].manager.listenMetricsURLs=http://localhost:2382")
create_args+=("--set spec.kubeScheduler.authorizationAlwaysAllowPaths=/healthz")
create_args+=("--set spec.kubeScheduler.authorizationAlwaysAllowPaths=/metrics")
create_args+=("--node-count=${KUBE_NODE_COUNT:-101}")
# TODO: track failures of tests (HostPort & OIDC) when using `--dns=none`
create_args+=("--dns none")
create_args+=("--node-size=c6g.medium")
create_args+=("--control-plane-count=${CONTROL_PLANE_COUNT:-1}")
create_args+=("--master-size=${CONTROL_PLANE_SIZE:-c6g.2xlarge}")
create_args+=("--zones=us-east-2a,us-east-2b,us-east-2c")


# Enable cluster addons, this enables us to replace the built-in manifest
KOPS_FEATURE_FLAGS="ClusterAddons,${KOPS_FEATURE_FLAGS:-}"
echo "KOPS_FEATURE_FLAGS=${KOPS_FEATURE_FLAGS}"


# Note that these arguments for kubetest2
KUBETEST2_ARGS=()
KUBETEST2_ARGS+=("-v=2")
KUBETEST2_ARGS+=("--cloud-provider=${CLOUD_PROVIDER}")
KUBETEST2_ARGS+=("--cluster-name=${CLUSTER_NAME:-}")
KUBETEST2_ARGS+=("--kops-binary-path=${KOPS_BIN}")
KUBETEST2_ARGS+=("--admin-access=${ADMIN_ACCESS:-}")
KUBETEST2_ARGS+=("--env=KOPS_FEATURE_FLAGS=${KOPS_FEATURE_FLAGS}")

# More time for bigger clusters
KUBETEST2_ARGS+=("--validation-wait=30m")

# The caller can set DELETE_CLUSTER=false to stop us deleting the cluster
if [[ -z "${DELETE_CLUSTER:-}" ]]; then
  DELETE_CLUSTER="true"
fi

kubetest2 kops "${KUBETEST2_ARGS[@]}" \
  --up \
  --kubernetes-version="${K8S_VERSION}" \
  --create-args="${create_args[*]}"

KUBECONFIG=$(mktemp -t kubeconfig.XXXXXXXXX)
kops export kubecfg --admin --kubeconfig="${KUBECONFIG}"

if [[ "${RUN_CL2_TEST:-}" == "true" ]]; then
  # CL2 uses KUBE_SSH_KEY_PATH path to ssh to instances for scraping metrics
  export KUBE_SSH_KEY_PATH="/tmp/kops/${CLUSTER_NAME}/id_ed25519"

  kubetest2 kops "${KUBETEST2_ARGS[@]}" \
  --test=clusterloader2 \
  --kubernetes-version="${K8S_VERSION}" \
  -- \
  --provider="${CLOUD_PROVIDER}" \
  --repo-root="${GOPATH}"/src/k8s.io/perf-tests \
  --test-configs="${GOPATH}"/src/k8s.io/perf-tests/clusterloader2/testing/load/config.yaml \
  --kube-config="${KUBECONFIG}"
else
  kubetest2 kops "${KUBETEST2_ARGS[@]}" \
  --test=kops \
  --kubernetes-version="${K8S_VERSION}" \
  -- \
  --test-package-version="${K8S_VERSION}" \
  --parallel=30 \
  --skip-regex="\[Serial\]" \
  --focus-regex="\[Conformance\]"
fi

if [[ "${DELETE_CLUSTER:-}" == "true" ]]; then
  kubetest2 kops "${KUBETEST2_ARGS[@]}" --down
  DELETE_CLUSTER=false # Don't delete again in trap
fi
