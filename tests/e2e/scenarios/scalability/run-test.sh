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

make test-e2e-install

# Default cluster name
SCRIPT_NAME=$(basename "$(dirname "$0")")
if [[ -z "${CLUSTER_NAME:-}" ]]; then
  CLUSTER_NAME="${SCRIPT_NAME}.k8s.local"
fi
echo "CLUSTER_NAME=${CLUSTER_NAME}"

if [[ -z "${K8S_VERSION:-}" ]]; then
  K8S_VERSION=https://storage.googleapis.com/k8s-release-dev/ci/latest.txt
fi

# Default Scale Scenario to performance
if [[ -z "${SCALE_SCENARIO:-}" ]]; then
  SCALE_SCENARIO="performance"
  export SCALE_SCENARIO
fi

# Default cloud provider to aws
if [[ -z "${CLOUD_PROVIDER:-}" ]]; then
  CLOUD_PROVIDER="aws"
fi
echo "CLOUD_PROVIDER=${CLOUD_PROVIDER}"
if [[ "${CLOUD_PROVIDER}" != "gce" ]]; then
  # KOPS_STATE_STORE holds metadata about the clusters we create
  if [[ -z "${KOPS_STATE_STORE:-}" ]]; then
    echo "Must specify KOPS_STATE_STORE"
    exit 1
  fi
  echo "KOPS_STATE_STORE=${KOPS_STATE_STORE}"
  export KOPS_STATE_STORE
fi

if [[ -z "${ADMIN_ACCESS:-}" ]]; then
  ADMIN_ACCESS="0.0.0.0/0" # Or use your IPv4 with /32
fi
echo "ADMIN_ACCESS=${ADMIN_ACCESS}"

KOPS_SCHEDULER_QPS="${KOPS_SCHEDULER_QPS:-500}"
KOPS_SCHEDULER_BURST="${KOPS_SCHEDULER_BURST:-500}"
KOPS_CONTROLLER_MANAGER_QPS="${KOPS_CONTROLLER_MANAGER_QPS:-500}"
KOPS_CONTROLLER_MANAGER_BURST="${KOPS_CONTROLLER_MANAGER_BURST:-500}"
KOPS_APISERVER_MAX_REQUESTS_INFLIGHT="${KOPS_APISERVER_MAX_REQUESTS_INFLIGHT:-800}"
echo "KOPS_SCHEDULER_QPS=${KOPS_SCHEDULER_QPS} KOPS_SCHEDULER_BURST=${KOPS_SCHEDULER_BURST}"
echo "KOPS_CONTROLLER_MANAGER_QPS=${KOPS_CONTROLLER_MANAGER_QPS} KOPS_CONTROLLER_MANAGER_BURST=${KOPS_CONTROLLER_MANAGER_BURST}"
echo "KOPS_APISERVER_MAX_REQUESTS_INFLIGHT=${KOPS_APISERVER_MAX_REQUESTS_INFLIGHT}"

# cilium does not yet pass conformance tests (shared hostport test)
#create_args="--networking cilium"
create_args=()
if [[ "${CLOUD_PROVIDER}" == "aws" ]]; then
  create_args+=("--network-cidr=10.0.0.0/16,10.1.0.0/16,10.2.0.0/16,10.3.0.0/16,10.4.0.0/16,10.5.0.0/16,10.6.0.0/16,10.7.0.0/16,10.8.0.0/16,10.9.0.0/16,10.10.0.0/16,10.11.0.0/16,10.12.0.0/16")
  create_args+=("--node-size=t3a.medium,t3.medium,t3a.large,c5a.large,t3.large,c5.large,m5a.large,m6a.large,m5.large,c7a.large,r5a.large,r6a.large,m7a.large")
  create_args+=("--node-volume-size=20")
  create_args+=("--zones=us-east-2a,us-east-2b,us-east-2c")
  create_args+=("--image=${INSTANCE_IMAGE:-ssm:/aws/service/canonical/ubuntu/server/22.04/stable/current/amd64/hvm/ebs-gp2/ami-id}")
  # TODO: track failures of tests (HostPort & OIDC) when using `--dns=none`
  create_args+=("--dns=none")
fi
if [[ "${CLOUD_PROVIDER}" == "gce" ]]; then
  create_args+=("--zones=us-east1-b,us-east1-c,us-east1-d")
  create_args+=("--node-size=e2-standard-2")
  create_args+=("--node-volume-size=30")
  create_args+=("--master-volume-size=1000")
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
create_args+=("--set spec.etcdClusters[*].manager.env=ETCD_QUOTA_BACKEND_BYTES=8589934592")
create_args+=("--set spec.etcdClusters[*].manager.env=ETCD_ENABLE_PPROF=true")
create_args+=("--set spec.cloudControllerManager.concurrentNodeSyncs=10")
create_args+=("--set spec.kubelet.maxPods=96")
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
create_args+=("--set spec.kubeAPIServer.logLevel=2")
# this is required for Prometheus server to scrape metrics endpoint on APIServer
create_args+=("--set spec.kubeAPIServer.anonymousAuth=true")
# this is required for kindnet to use nftables
create_args+=("--set spec.kubeProxy.proxyMode=${KUBE_PROXY_MODE:-iptables}")
# this is required for prometheus to scrape kube-proxy metrics endpoint
create_args+=("--set spec.kubeProxy.metricsBindAddress=0.0.0.0:10249")
create_args+=("--node-count=${KUBE_NODE_COUNT:-100}")
create_args+=("--control-plane-count=${CONTROL_PLANE_COUNT:-1}")
create_args+=("--master-size=c7i.24xlarge")


# AWS ONLY feature flags
if [[ "${CLOUD_PROVIDER}" == "aws" ]]; then
  # Enable creating a single nodes instance group
  KOPS_FEATURE_FLAGS="AWSSingleNodesInstanceGroup,${KOPS_FEATURE_FLAGS:-}"
  create_args+=("--set spec.etcdClusters[*].etcdMembers[*].volumeIOPS=20000")
  create_args+=("--set spec.etcdClusters[*].etcdMembers[*].volumeThroughput=4000")
  create_args+=("--set spec.etcdClusters[*].etcdMembers[*].volumeSize=200")
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
if [[ "${JOB_TYPE}" == "presubmit" && "${REPO_OWNER}/${REPO_NAME}" == "kubernetes/kops" ]]; then
  KUBETEST2_ARGS+=("--build")
  KUBETEST2_ARGS+=("--kops-binary-path=${GOPATH}/src/k8s.io/kops/.build/dist/linux/$(go env GOARCH)/kops")
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

# this is used as a label to select kube-proxy pods on kops for kube-proxy service 
# used by CL2 Prometheus here https://github.com/kubernetes/perf-tests/blob/master/clusterloader2/pkg/prometheus/manifests/default/kube-proxy-service.yaml#L2
export PROMETHEUS_KUBE_PROXY_SELECTOR_KEY="k8s-app"
export PROMETHEUS_SCRAPE_APISERVER_ONLY="true"
export CL2_PROMETHEUS_TOLERATE_MASTER="true"
if [[ "${CLOUD_PROVIDER}" == "aws" && "${SCALE_SCENARIO}" == "performance" ]]; then
  # CL2 uses KUBE_SSH_KEY_PATH path to ssh to instances for scraping metrics
  export KUBE_SSH_KEY_PATH="/tmp/kops/${CLUSTER_NAME}/id_ed25519"
  cat > "${GOPATH}"/src/k8s.io/perf-tests/clusterloader2/testing/load/overrides.yaml <<EOL
  # we are not testing PVS at this point
  CL2_ENABLE_PVS: false
  ENABLE_RESTART_COUNT_CHECK: false
EOL
  cat "${GOPATH}"/src/k8s.io/perf-tests/clusterloader2/testing/load/overrides.yaml
else
  cat > "${GOPATH}"/src/k8s.io/perf-tests/clusterloader2/testing/load/overrides.yaml <<EOL
  # setting a default value here to avoid an incorrect yaml file
  CL2_ENABLE_PVS: true
EOL
fi

CLUSTERLOADER2_ARGS=()
if [[ -n "${KOPS_CL2_TEST_CONFIG}" ]]; then
  CLUSTERLOADER2_ARGS+=("--test-configs=${GOPATH}/src/k8s.io/perf-tests/clusterloader2/${KOPS_CL2_TEST_CONFIG}")
else
  CLUSTERLOADER2_ARGS+=("--test-configs=${GOPATH}/src/k8s.io/perf-tests/clusterloader2/testing/load/config.yaml")
  CLUSTERLOADER2_ARGS+=("--test-configs=${GOPATH}/src/k8s.io/perf-tests/clusterloader2/testing/access-tokens/config.yaml")
  CLUSTERLOADER2_ARGS+=("--test-overrides=${GOPATH}/src/k8s.io/perf-tests/clusterloader2/testing/load/overrides.yaml")
  CLUSTERLOADER2_ARGS+=("--test-overrides=${GOPATH}/src/k8s.io/perf-tests/clusterloader2/testing/experiments/enable_restart_count_check.yaml")
  CLUSTERLOADER2_ARGS+=("--test-overrides=${GOPATH}/src/k8s.io/perf-tests/clusterloader2/testing/experiments/ignore_known_gce_container_restarts.yaml")
  CLUSTERLOADER2_ARGS+=("--test-overrides=${GOPATH}/src/k8s.io/perf-tests/clusterloader2/testing/overrides/5000_nodes.yaml")
  CLUSTERLOADER2_ARGS+=("--extra-args=--experimental-prometheus-snapshot-to-report-dir=true")
fi

# ToDo: remove this once we can run the huge-service test on AWS
if [[ -z "${KOPS_CL2_TEST_CONFIG}" && "${CLOUD_PROVIDER}" == "gce" ]]; then
  CLUSTERLOADER2_ARGS+=("--test-configs=${GOPATH}/src/k8s.io/perf-tests/clusterloader2/testing/huge-service/config.yaml")
fi

if [[ "${SCALE_SCENARIO:performance}" == "correctness" ]]; then
  kubetest2 kops "${KUBETEST2_ARGS[@]}" \
    --up \
    --kubernetes-version="${K8S_VERSION}" \
    --create-args="${create_args[*]}" \
    --test=kops \
    -- \
    --test-package-url=https://storage.googleapis.com/k8s-release-dev \
    --test-package-dir=ci \
    --test-package-marker=latest.txt \
    --skip-regex="\[Driver:.gcepd\]|\[Serial\]|\[Disruptive\]|\[Flaky\]|\[Feature:([^L].*|L[^o].*|Lo[^a].*|Loa[^d].*)\]\[KubeUp\]" \
    --parallel=25
else
  kubetest2 kops "${KUBETEST2_ARGS[@]}" \
    --up \
    --kubernetes-version="${K8S_VERSION}" \
    --create-args="${create_args[*]}" \
    --test=clusterloader2 \
    -- \
    --provider="${CLOUD_PROVIDER}" \
    --repo-root="${GOPATH}"/src/k8s.io/perf-tests \
    --kube-config="${HOME}/.kube/config" \
    "${CLUSTERLOADER2_ARGS[@]}"
fi
