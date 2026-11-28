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

set -e
set -x

make test-e2e-install

REPO_ROOT=$(git rev-parse --show-toplevel)
if [[ -z "${K8S_VERSION:-}" ]]; then
  K8S_VERSION=https://storage.googleapis.com/k8s-release-dev/ci/latest.txt
fi

# Default cloud provider to aws
if [[ -z "${CLOUD_PROVIDER:-}" ]]; then
  CLOUD_PROVIDER="aws"
fi
echo "CLOUD_PROVIDER=${CLOUD_PROVIDER}"
if [[ -z "${ADMIN_ACCESS:-}" ]]; then
  ADMIN_ACCESS="0.0.0.0/0" # Or use your IPv4 with /32
fi
echo "ADMIN_ACCESS=${ADMIN_ACCESS}"

KOPS_SCHEDULER_QPS="${KOPS_SCHEDULER_QPS:-500}"
KOPS_SCHEDULER_BURST="${KOPS_SCHEDULER_BURST:-500}"
KOPS_CONTROLLER_MANAGER_QPS="${KOPS_CONTROLLER_MANAGER_QPS:-500}"
KOPS_CONTROLLER_MANAGER_BURST="${KOPS_CONTROLLER_MANAGER_BURST:-500}"
KOPS_APISERVER_MAX_REQUESTS_INFLIGHT="${KOPS_APISERVER_MAX_REQUESTS_INFLIGHT:-640}"
ETCD_QUOTA_BACKEND_BYTES="${ETCD_QUOTA_BACKEND_BYTES:-8589934592}"
echo "KOPS_SCHEDULER_QPS=${KOPS_SCHEDULER_QPS} KOPS_SCHEDULER_BURST=${KOPS_SCHEDULER_BURST}"
echo "KOPS_CONTROLLER_MANAGER_QPS=${KOPS_CONTROLLER_MANAGER_QPS} KOPS_CONTROLLER_MANAGER_BURST=${KOPS_CONTROLLER_MANAGER_BURST}"
echo "KOPS_APISERVER_MAX_REQUESTS_INFLIGHT=${KOPS_APISERVER_MAX_REQUESTS_INFLIGHT}"

# cilium does not yet pass conformance tests (shared hostport test)
#create_args="--networking cilium"
create_args=()
if [[ "${CLOUD_PROVIDER}" == "aws" ]]; then
  create_args+=("--network-cidr=10.0.0.0/16,10.1.0.0/16,10.2.0.0/16,10.3.0.0/16,10.4.0.0/16,10.5.0.0/16,10.6.0.0/16,10.7.0.0/16,10.8.0.0/16,10.9.0.0/16,10.10.0.0/16,10.11.0.0/16,10.12.0.0/16")
  create_args+=("--node-size=${NODE_SIZE:-t3a.medium,t3.medium,t3a.large,c5a.large,t3.large,c5.large,m5a.large,m6a.large,m5.large,c7a.large,r5a.large,r6a.large,m7a.large}")
  create_args+=("--node-volume-size=20")
  create_args+=("--control-plane-volume-size=500")
  create_args+=("--zones=us-east-2a,us-east-2b,us-east-2c")
  create_args+=("--image=${INSTANCE_IMAGE:-ssm:/aws/service/canonical/ubuntu/server/24.04/stable/current/amd64/hvm/ebs-gp3/ami-id}")
  # TODO: track failures of tests (HostPort & OIDC) when using `--dns=none`
  create_args+=("--dns=none")
fi
if [[ "${CLOUD_PROVIDER}" == "gce" ]]; then
  create_args+=("--zones=us-east1-b,us-east1-c,us-east1-d")
  create_args+=("--node-size=${NODE_SIZE:-e2-medium}")
  create_args+=("--node-volume-size=30")
  create_args+=("--control-plane-volume-size=1000")
  create_args+=("--gce-service-account=default")
  create_args+=("--topology=private")
  create_args+=("--image=${INSTANCE_IMAGE:-ubuntu-os-cloud/ubuntu-2404-noble-amd64-v20251001}")
  create_args+=("--set spec.networking.podCIDR=10.64.0.0/10")
  create_args+=("--set spec.networking.subnets[0].cidr=10.128.0.0/15")
  create_args+=("--set spec.networking.serviceClusterIPRange=10.130.0.0/15")
  create_args+=("--set spec.networking.nonMasqueradeCIDR=10.64.0.0/10")
  create_args+=("--set spec.etcdClusters[*].etcdMembers[*].volumeIOPS=10000")
  create_args+=("--set spec.etcdClusters[*].etcdMembers[*].volumeThroughput=1000")
  create_args+=("--set spec.etcdClusters[*].etcdMembers[*].volumeSize=120")
  create_args+=("--set spec.etcdClusters[*].etcdMembers[*].volumeType=hyperdisk-balanced")
fi
create_args+=("--networking=${CNI_PLUGIN:-calico}")
if [[ "${CNI_PLUGIN}" == "amazonvpc" ]]; then
  create_args+=("--set spec.networking.amazonVPC.env=ENABLE_PREFIX_DELEGATION=true")
fi
create_args+=("--set spec.etcdClusters[0].manager.listenMetricsURLs=http://localhost:2382")
create_args+=("--set spec.etcdClusters[*].manager.env=ETCD_QUOTA_BACKEND_BYTES=${ETCD_QUOTA_BACKEND_BYTES}")
create_args+=("--set spec.etcdClusters[*].manager.env=ETCD_ENABLE_PPROF=true")
create_args+=("--set spec.cloudControllerManager.concurrentNodeSyncs=10")
create_args+=("--set spec.kubelet.maxPods=96")
create_args+=("--set spec.kubelet.kubeAPIQPS=100")
create_args+=("--set spec.kubeScheduler.authorizationAlwaysAllowPaths=/healthz")
create_args+=("--set spec.kubeScheduler.authorizationAlwaysAllowPaths=/livez")
create_args+=("--set spec.kubeScheduler.authorizationAlwaysAllowPaths=/readyz")
create_args+=("--set spec.kubeScheduler.authorizationAlwaysAllowPaths=/metrics")
create_args+=("--set spec.kubeScheduler.qps=${KOPS_SCHEDULER_QPS}")
create_args+=("--set spec.kubeScheduler.burst=${KOPS_SCHEDULER_BURST}")
create_args+=("--set spec.kubeScheduler.enableProfiling=true")
create_args+=("--set spec.kubeScheduler.enableContentionProfiling=true")
create_args+=("--set spec.kubeControllerManager.endpointUpdatesBatchPeriod=500ms")
create_args+=("--set spec.kubeControllerManager.endpointSliceUpdatesBatchPeriod=500ms")
create_args+=("--set spec.kubeControllerManager.kubeAPIQPS=${KOPS_CONTROLLER_MANAGER_QPS}")
create_args+=("--set spec.kubeControllerManager.kubeAPIBurst=${KOPS_CONTROLLER_MANAGER_BURST}")
create_args+=("--set spec.kubeControllerManager.enableProfiling=true")
create_args+=("--set spec.kubeControllerManager.enableContentionProfiling=true")
# inflight requests are bit higher than what currently upstream uses for GCE scale tests
create_args+=("--set spec.kubeAPIServer.maxRequestsInflight=${KOPS_APISERVER_MAX_REQUESTS_INFLIGHT}")
create_args+=("--set spec.kubeAPIServer.maxMutatingRequestsInflight=0")
create_args+=("--set spec.kubeAPIServer.enableProfiling=true")
create_args+=("--set spec.kubeAPIServer.enableContentionProfiling=true")
create_args+=("--set spec.kubeAPIServer.logLevel=3")
create_args+=("--set spec.kubeAPIServer.deleteCollectionWorkers=16")
create_args+=("--set spec.kubeAPIServer.compactionInterval=150s")

# this is required for Prometheus server to scrape metrics endpoint on APIServer
create_args+=("--set spec.kubeAPIServer.anonymousAuth=true")
# this is required for kindnet to use nftables
create_args+=("--set spec.kubeProxy.proxyMode=${KUBE_PROXY_MODE:-iptables}")
# this is required for prometheus to scrape kube-proxy metrics endpoint
create_args+=("--set spec.kubeProxy.metricsBindAddress=0.0.0.0:10249")
# bump coredns memory on large clusters
if [[ "${KUBE_NODE_COUNT:-100}" -ge 5000 ]]; then
  create_args+=("--set spec.kubeDNS.memoryRequest=340Mi")
  create_args+=("--set spec.kubeDNS.memoryLimit=340Mi")
fi
create_args+=("--node-count=${KUBE_NODE_COUNT:-100}")
create_args+=("--control-plane-count=${CONTROL_PLANE_COUNT:-1}")
create_args+=("--control-plane-size=${CONTROL_PLANE_SIZE:-c5.2xlarge}")

# Enable HTTP for events etcd to reduce TLS overhead in scale tests
KOPS_FEATURE_FLAGS="EtcdEventsHTTP,${KOPS_FEATURE_FLAGS:-}"

# AWS ONLY feature flags
if [[ "${CLOUD_PROVIDER}" == "aws" ]]; then
  # AWS doesn't run dedicated addons node
  export CL2_PROMETHEUS_TOLERATE_MASTER="true"
  # Enable creating a single nodes instance group
  KOPS_FEATURE_FLAGS="AWSSingleNodesInstanceGroup,${KOPS_FEATURE_FLAGS:-}"
  create_args+=("--set spec.etcdClusters[*].etcdMembers[*].volumeIOPS=10000")
  create_args+=("--set spec.etcdClusters[*].etcdMembers[*].volumeSize=120")
  create_args+=("--set spec.etcdClusters[*].etcdMembers[*].volumeType=io2")

fi
echo "KOPS_FEATURE_FLAGS=${KOPS_FEATURE_FLAGS}"

# Note that these arguments for kubetest2
KUBETEST2_ARGS=()
KUBETEST2_ARGS+=("-v=2")
KUBETEST2_ARGS+=("--max-nodes-to-dump=${MAX_NODES_TO_DUMP:-5}")
KUBETEST2_ARGS+=("--cloud-provider=${CLOUD_PROVIDER}")
KUBETEST2_ARGS+=("--cluster-name=${CLUSTER_NAME:-}")
KUBETEST2_ARGS+=("--admin-access=${ADMIN_ACCESS:-}")
KUBETEST2_ARGS+=("--env=KOPS_FEATURE_FLAGS=${KOPS_FEATURE_FLAGS}")
KUBETEST2_ARGS+=("--pre-test-cmd=${REPO_ROOT}/tests/e2e/scenarios/scalability/pre-test.sh")
if [[ -n "${KUBE_FEATURE_GATES:-}" ]]; then
  KUBETEST2_ARGS+=("--kubernetes-feature-gates=${KUBE_FEATURE_GATES}")
fi
if [[ "${JOB_TYPE}" == "presubmit" && "${REPO_OWNER}/${REPO_NAME}" == "kubernetes/kops" ]]; then
  KUBETEST2_ARGS+=("--build")
  KUBETEST2_ARGS+=("--kops-binary-path=${GOPATH}/src/k8s.io/kops/.build/dist/linux/$(go env GOARCH)/kops")
elif [[ "${JOB_TYPE}" == "presubmit" && "${REPO_OWNER}/${REPO_NAME}" == "kubernetes/kubernetes" ]]; then
  KUBETEST2_ARGS+=("--build")
  KUBETEST2_ARGS+=("--build-kubernetes=true")
  # Scale clusters run linux/amd64 nodes only, so build a single arch instead of
  # the full multi-arch release matrix.
  KUBETEST2_ARGS+=("--target-build-arch=linux/amd64")
  KUBETEST2_ARGS+=("--kops-version-marker=${KOPS_VERSION_MARKER:-https://storage.googleapis.com/k8s-staging-kops/kops/releases/markers/master/latest-ci.txt}")
  cd "${GOPATH}/src/k8s.io/kubernetes"
else
  KUBETEST2_ARGS+=("--kops-version-marker=${KOPS_VERSION_MARKER:-https://storage.googleapis.com/k8s-staging-kops/kops/releases/markers/master/latest-ci.txt}")
fi

if [[ "${CLOUD_PROVIDER}" == "gce" ]]; then
  if [[ -n "${GCP_PROJECT:-}" ]]; then
    KUBETEST2_ARGS+=("--gcp-project=${GCP_PROJECT}")
  else
    KUBETEST2_ARGS+=("--boskos-resource-type=${BOSKOS_RESOURCE_TYPE:-scalability-project}")
  fi
  KUBETEST2_ARGS+=("--control-plane-instance-group-overrides=spec.rootVolume.type=hyperdisk-balanced")
  KUBETEST2_ARGS+=("--control-plane-instance-group-overrides=spec.rootVolume.iops=10000")
  KUBETEST2_ARGS+=("--control-plane-instance-group-overrides=spec.rootVolume.throughput=1000")
  KUBETEST2_ARGS+=("--control-plane-instance-group-overrides=spec.associatePublicIP=true")
elif [[ "${CLOUD_PROVIDER}" == "aws" ]]; then
  KUBETEST2_ARGS+=("--control-plane-instance-group-overrides=spec.rootVolume.type=io2")
fi

# More time for bigger clusters
KUBETEST2_ARGS+=("--validation-wait=75m")
KUBETEST2_ARGS+=("--validation-count=3")
KUBETEST2_ARGS+=("--validation-interval=60s")

# The caller can set DELETE_CLUSTER=false to stop us deleting the cluster
if [[ -z "${DELETE_CLUSTER:-}" ]]; then
  DELETE_CLUSTER="true"
fi

if [[ "${DELETE_CLUSTER:-}" == "true" ]]; then
  KUBETEST2_ARGS+=("--down")
fi

export PROMETHEUS_KUBE_PROXY_SELECTOR_KEY="k8s-app"

# Download and install kube-burner if path not already specified
if [[ -z "${KUBE_BURNER_PATH:-}" ]]; then
  if [[ -z "${KUBE_BURNER_VERSION:-}" ]]; then
    KUBE_BURNER_VERSION=$(curl -fsSL https://api.github.com/repos/kube-burner/kube-burner/releases/latest | grep -o '"tag_name": "v[^"]*"' | sed 's/"tag_name": "v\(.*\)"/\1/')
  fi
  KUBE_BURNER_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
  KUBE_BURNER_ARCH=$(uname -m)
  if [[ "${KUBE_BURNER_ARCH}" == "aarch64" ]]; then
    KUBE_BURNER_ARCH="arm64"
  fi
  KUBE_BURNER_DIR=$(mktemp -d)
  KUBE_BURNER_TARBALL="kube-burner-V${KUBE_BURNER_VERSION}-${KUBE_BURNER_OS}-${KUBE_BURNER_ARCH}.tar.gz"
  KUBE_BURNER_URL="https://github.com/kube-burner/kube-burner/releases/download/v${KUBE_BURNER_VERSION}/${KUBE_BURNER_TARBALL}"
  echo "Downloading kube-burner ${KUBE_BURNER_VERSION} from ${KUBE_BURNER_URL}"
  curl -fsSL "${KUBE_BURNER_URL}" -o "${KUBE_BURNER_DIR}/${KUBE_BURNER_TARBALL}"
  tar -xzf "${KUBE_BURNER_DIR}/${KUBE_BURNER_TARBALL}" -C "${KUBE_BURNER_DIR}"
  KUBE_BURNER_PATH="${KUBE_BURNER_DIR}/kube-burner"
  chmod +x "${KUBE_BURNER_PATH}"
  echo "kube-burner ${KUBE_BURNER_VERSION} installed at ${KUBE_BURNER_PATH}"
fi

KUBE_BURNER_ARGS=()
KUBE_BURNER_ARGS+=("--workdir=${KUBE_BURNER_WORKDIR:-k8s.io/perf-tests}")
KUBE_BURNER_ARGS+=("--workload=${KUBE_BURNER_WORKLOAD}")
KUBE_BURNER_ARGS+=("--kube-burner-path=${KUBE_BURNER_PATH}")
if [[ -n "${KUBE_BURNER_UUID:-}" ]]; then
  KUBE_BURNER_ARGS+=("--uuid=${KUBE_BURNER_UUID}")
fi
if [[ "${KUBE_BURNER_SKIP_TLS_VERIFY:-}" == "true" ]]; then
  KUBE_BURNER_ARGS+=("--skip-tls-verify")
fi
if [[ -n "${KUBE_BURNER_KUBECONFIG:-}" ]]; then
  KUBE_BURNER_ARGS+=("--kubeconfig=${KUBE_BURNER_KUBECONFIG}")
fi
if [[ -n "${KUBE_BURNER_LOG_LEVEL:-}" ]]; then
  KUBE_BURNER_ARGS+=("--log-level=${KUBE_BURNER_LOG_LEVEL}")
fi
if [[ -n "${KUBE_BURNER_EXTRA_ARGS:-}" ]]; then
  KUBE_BURNER_ARGS+=("--extra-args=${KUBE_BURNER_EXTRA_ARGS}")
fi

kubetest2 kops "${KUBETEST2_ARGS[@]}" \
  --up \
  --kubernetes-version="${K8S_VERSION}" \
  --create-args="${create_args[*]}" \
  --test=kube-burner \
  -- \
  "${KUBE_BURNER_ARGS[@]}"

if [[ -n "${KUBE_BURNER_REPORT_DIR:-}" ]]; then
  mkdir -p "${ARTIFACTS}/${KUBE_BURNER_REPORT_DIR}"
  if ls collected-metrics* 1>/dev/null 2>&1; then
    mv collected-metrics* "${ARTIFACTS}/${KUBE_BURNER_REPORT_DIR}/"
    echo "Kube-burner locally indexed metrics moved to ${ARTIFACTS}/${KUBE_BURNER_REPORT_DIR}"
  fi
fi
