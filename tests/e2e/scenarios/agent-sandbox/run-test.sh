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

# Runs the kubernetes-sigs/agent-sandbox benchmark and stress test against a
# kOps cluster on GCE (the default), AWS or Azure, selected by CLOUD_PROVIDER.
# Upstream runs the same test on GCE in its benchmarks-kops-gcp-* jobs with a
# pinned kOps release; this scenario runs it with the kOps build under test, so
# kOps changes that affect pod launch throughput or control-plane behaviour
# show up in presubmits.
#
# The cluster tuning follows test/benchmarks/scenarios/benchmarks-kops-gcp/run
# in agent-sandbox, which documents the measurements behind each knob, and the
# STRESS_* variables keep that script's names.

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(git rev-parse --show-toplevel)
cd "${REPO_ROOT}"

export CLOUD_PROVIDER="${CLOUD_PROVIDER:-gce}"
# CNI is the upstream job's name for the --networking value (cilium or kindnet in its CI).
export NETWORKING="${CNI:-${NETWORKING:-cilium}}"

AGENT_SANDBOX_REPO="${AGENT_SANDBOX_REPO:-https://github.com/kubernetes-sigs/agent-sandbox}"
AGENT_SANDBOX_VERSION="${AGENT_SANDBOX_VERSION:-}"
# The controller images are not built here: the checkout and the images deployed
# to the cluster are pinned to the same published release tag.
export IMAGE_PREFIX="${IMAGE_PREFIX:-registry.k8s.io/agent-sandbox/}"

# 3 workers behind a control plane of the same 8 vCPU / 32 GiB shape on every
# cloud, the profile the kOps jobs run; AWS and Azure use each cloud's closest
# match to the GCE shapes. The control-plane size does not follow
# STRESS_NODE_COUNT, so a larger run sets STRESS_CONTROL_PLANE_SIZE as well.
STRESS_NODE_COUNT="${STRESS_NODE_COUNT:-3}"
STRESS_NODE_CIDR_MASK_SIZE="${STRESS_NODE_CIDR_MASK_SIZE:-23}"
STRESS_ETCD_QUOTA_BYTES="${STRESS_ETCD_QUOTA_BYTES:-8589934592}"
STRESS_VALIDATE_WAIT="${STRESS_VALIDATE_WAIT:-25m}"

OVERRIDES="${OVERRIDES-}"
case "${CLOUD_PROVIDER}" in
  gce)
    export ZONES="${ZONES:-us-central1-a}"
    STRESS_CONTROL_PLANE_SIZE="${STRESS_CONTROL_PLANE_SIZE:-c3-standard-8}"
    STRESS_NODE_SIZE="${STRESS_NODE_SIZE:-n2-standard-8}"
    # Boskos projects do not allow IAM changes, so the VMs run as the default compute service account.
    OVERRIDES="${OVERRIDES} --gce-service-account=default"
    # Root volumes on pd-ssd: pod launch is fsync-bound on pd-standard, and the c3
    # control-plane machine family does not support pd-standard boot disks.
    STRESS_NODE_VOLUME_TYPE="${STRESS_NODE_VOLUME_TYPE:-pd-ssd}"
    OVERRIDES="${OVERRIDES} --control-plane-volume-type=pd-ssd"
    ;;
  aws)
    # No default zone: kubetest2-kops picks a zone that offers both instance
    # types, in a random region per run (us-east-1a had no m6i.2xlarge
    # capacity). Set ZONES to pin one.
    STRESS_CONTROL_PLANE_SIZE="${STRESS_CONTROL_PLANE_SIZE:-m7i.2xlarge}"
    STRESS_NODE_SIZE="${STRESS_NODE_SIZE:-m6i.2xlarge}"
    # gp3, kOps' AWS default, is an SSD with a 3000 IOPS baseline at any size.
    STRESS_NODE_VOLUME_TYPE="${STRESS_NODE_VOLUME_TYPE:-gp3}"
    ;;
  azure)
    # No default zone: kubetest2-kops picks uksouth-1 or uksouth-2, where the
    # other kOps Azure jobs run. Ddsv5 is Dsv5 plus a local temp disk, and the
    # CI subscription's Ddsv5 quota there is 250 cores against the 65-core
    # default for Dsv5.
    STRESS_CONTROL_PLANE_SIZE="${STRESS_CONTROL_PLANE_SIZE:-Standard_D8ds_v5}"
    STRESS_NODE_SIZE="${STRESS_NODE_SIZE:-Standard_D8ds_v5}"
    # Worker roots are ephemeral OS disks on the VM's 300 GiB temp disk: pod
    # launch is fsync-bound, and a 128 GB Premium_LRS root (P10, 500 IOPS)
    # started pods three to four times slower than on gp3.
    STRESS_NODE_VOLUME_TYPE="${STRESS_NODE_VOLUME_TYPE:-ResourceDisk}"
    # kOps' Azure support is alpha and feature-gated, for every kops invocation.
    if [[ ",${KOPS_FEATURE_FLAGS:-}," != *,Azure,* ]]; then
      export KOPS_FEATURE_FLAGS="${KOPS_FEATURE_FLAGS:+${KOPS_FEATURE_FLAGS},}Azure"
    fi
    ;;
  *)
    echo "Unsupported CLOUD_PROVIDER=${CLOUD_PROVIDER}: expected gce, aws or azure" >&2
    exit 1
    ;;
esac
OVERRIDES="${OVERRIDES} --control-plane-size=${STRESS_CONTROL_PLANE_SIZE}"
OVERRIDES="${OVERRIDES} --node-count=${STRESS_NODE_COUNT}"
OVERRIDES="${OVERRIDES} --node-size=${STRESS_NODE_SIZE}"
OVERRIDES="${OVERRIDES} --node-volume-type=${STRESS_NODE_VOLUME_TYPE}"
# kubetest2-kops would otherwise shrink the root volumes to 48GB (100GB for GCE
# nodes); keep the kOps defaults the upstream job runs with, as pd-ssd IOPS
# scale with size.
OVERRIDES="${OVERRIDES} --node-volume-size=128"
OVERRIDES="${OVERRIDES} --control-plane-volume-size=64"
# Let the stress tool scrape kube-controller-manager and kube-scheduler /metrics
# anonymously through the apiserver pod proxy. The ports are not reachable from
# outside the cluster on any cloud: kOps' GCE firewall, AWS security groups and
# Azure NSG expose only SSH, the API and the NodePort range kubetest2-kops
# opens. The probe paths must be repeated because setting the field replaces
# the default list the liveness probes use.
for path in /healthz /readyz /livez /metrics; do
  OVERRIDES="${OVERRIDES} --set=cluster.spec.kubeControllerManager.authorizationAlwaysAllowPaths=${path}"
  OVERRIDES="${OVERRIDES} --set=cluster.spec.kubeScheduler.authorizationAlwaysAllowPaths=${path}"
done
# Client-side rate limits are the cluster-scale throughput cap: the garbage
# collector in kube-controller-manager is on the sandbox delete path, and the
# scheduler binds every sandbox pod.
OVERRIDES="${OVERRIDES} --set=cluster.spec.kubeControllerManager.kubeAPIQPS=500"
OVERRIDES="${OVERRIDES} --set=cluster.spec.kubeControllerManager.kubeAPIBurst=500"
OVERRIDES="${OVERRIDES} --set=cluster.spec.kubeScheduler.qps=500"
OVERRIDES="${OVERRIDES} --set=cluster.spec.kubeScheduler.burst=500"
# Plain-HTTP etcd metrics listeners on the ports hard-coded in the stress tool
# (test/stress/promscrape.go). etcd's conventional 2381 is already the
# etcd-events peer port on kOps. Index order matches kops create cluster:
# [0]=main, [1]=events.
OVERRIDES="${OVERRIDES} --set=cluster.spec.etcdClusters[0].manager.listenMetricsURLs=http://0.0.0.0:2391"
OVERRIDES="${OVERRIDES} --set=cluster.spec.etcdClusters[1].manager.listenMetricsURLs=http://0.0.0.0:2392"
# Each kubelet status sync is a serialized GET+PATCH, which caps churn at about
# 5 pods/s/node under the default QPS. Event emission has its own limiter and
# drops extra events, which is acceptable for a benchmark.
OVERRIDES="${OVERRIDES} --set=cluster.spec.kubelet.kubeAPIQPS=500"
OVERRIDES="${OVERRIDES} --set=cluster.spec.kubelet.eventQPS=5"
OVERRIDES="${OVERRIDES} --set=cluster.spec.kubelet.eventBurst=50"
# /23 per node allows up to 510 pods per node and 4096 nodes within the /11 pod
# CIDR kOps defaults to with cilium (512 within the /14 that GCE IP aliases use
# with kindnet).
OVERRIDES="${OVERRIDES} --set=cluster.spec.kubeControllerManager.nodeCIDRMaskSize=${STRESS_NODE_CIDR_MASK_SIZE}"
# Above etcd's 2GiB default quota every write fails with NOSPACE until the alarm is cleared.
OVERRIDES="${OVERRIDES} --set=cluster.spec.etcdClusters[*].manager.env=ETCD_QUOTA_BACKEND_BYTES=${STRESS_ETCD_QUOTA_BYTES}"
if [[ "${NETWORKING}" == "kindnet" && "${STRESS_KINDNET_NETWORK_POLICIES:-false}" != "true" ]]; then
  # kindnet's NetworkPolicy engine watches every pod from every node; the benchmark does not use NetworkPolicy.
  OVERRIDES="${OVERRIDES} --set=cluster.spec.networking.kindnet.networkPolicies=false"
fi
if [[ "${NETWORKING}" == "cilium" ]]; then
  # Serve cilium-agent metrics on :9090, where the stress tool reads endpoint regeneration latency.
  OVERRIDES="${OVERRIDES} --set=cluster.spec.networking.cilium.enablePrometheusMetrics=true"
fi
# Serve containerd's CRI and CNI timers on :1338, which the stress tool scrapes
# through the node-exporter pod proxy to split network setup out of sandbox
# creation; like the other metrics ports it is not reachable from outside the
# cluster. kOps writes containerd config version 4, where the listener is the
# io.containerd.server.v1.metrics plugin. The single quotes keep the inner
# double quotes intact through kubetest2's create-args split.
OVERRIDES="${OVERRIDES} --set='cluster.spec.containerd.configAdditions=plugins.\"io.containerd.server.v1.metrics\".address=0.0.0.0:1338'"
# CGROUP_DRIVER=cgroupfs runs the benchmark without the systemd cgroup driver,
# whose per-container systemd round trips are a cost on the pod launch path.
# kOps 1.34+ always configures systemd and ignores spec.kubelet.cgroupDriver,
# so the switch is made in containerd instead: Kubernetes 1.34+ kubelets take
# the cgroup driver from the runtime. The single quotes keep the inner double
# quotes intact through kubetest2's create-args split.
if [[ "${CGROUP_DRIVER:-systemd}" == "cgroupfs" ]]; then
  OVERRIDES="${OVERRIDES} --set='cluster.spec.containerd.configAdditions=plugins.\"io.containerd.cri.v1.runtime\".containerd.runtimes.runc.options.SystemdCgroup=false'"
fi

source "${REPO_ROOT}/tests/e2e/scenarios/lib/common.sh"

# GCE provisioning on shared boskos projects intermittently stalls past
# kubetest2-kops' 20m default; the budget applies on every cloud.
KUBETEST2="${KUBETEST2} --validation-wait=${STRESS_VALIDATE_WAIT}"

kops-acquire-latest

kops-up

KUBECONFIG=$(mktemp -t kops.XXXXXXXXX)
export KUBECONFIG
"${KOPS}" export kubeconfig --name "${CLUSTER_NAME}" --admin --kubeconfig "${KUBECONFIG}"

if [[ "${NETWORKING}" == "cilium" ]]; then
  # Raise cilium-agent's endpoint-create API rate limit (0.5/s with auto-adjust
  # by default) and its client-go budget (5 QPS, burst 10). Both otherwise cap
  # pod launch throughput below what the datapath can do.
  kubectl --namespace kube-system patch configmap cilium-config --type merge \
    --patch '{"data":{"api-rate-limit":"endpoint-create=rate-limit:100/s,rate-burst:100,parallel-requests:32,auto-adjust:false","k8s-client-qps":"50","k8s-client-burst":"100"}}'
  # kOps deploys the cilium DaemonSet with updateStrategy=OnDelete, so the
  # agent pods have to be deleted to pick up the new config.
  kubectl --namespace kube-system delete pods --selector=k8s-app=cilium
  "${KOPS}" validate cluster --name "${CLUSTER_NAME}" --wait "${STRESS_CILIUM_VALIDATE_WAIT:-5m}"
fi

if [[ -z "${AGENT_SANDBOX_ROOT:-}" ]]; then
  AGENT_SANDBOX_ROOT="${WORKSPACE}/agent-sandbox"
  git clone --quiet "${AGENT_SANDBOX_REPO}" "${AGENT_SANDBOX_ROOT}"
  cd "${AGENT_SANDBOX_ROOT}"
  if [[ -z "${AGENT_SANDBOX_VERSION}" ]]; then
    AGENT_SANDBOX_VERSION=$(python3 dev/tools/latest-published-tag)
  fi
  git checkout --quiet "${AGENT_SANDBOX_VERSION}"
fi
cd "${AGENT_SANDBOX_ROOT}"
# deploy-to-kube and test-e2e read IMAGE_PREFIX and IMAGE_TAG from the environment.
if [[ -z "${IMAGE_TAG:-}" ]]; then
  IMAGE_TAG="${AGENT_SANDBOX_VERSION:-$(python3 dev/tools/latest-published-tag)}"
fi
export IMAGE_TAG
echo "Using agent-sandbox checkout ${AGENT_SANDBOX_ROOT} ($(git describe --tags --always)) with images ${IMAGE_PREFIX}*:${IMAGE_TAG}"

# node-exporter gives the report node-level CPU (including iowait), memory and
# disk stats; the stress tool scrapes it through the apiserver pod proxy.
kubectl apply -f test/benchmarks/scenarios/benchmarks-kops-gcp/node-exporter.yaml
kubectl --namespace kube-system rollout status daemonset/node-exporter --timeout=3m
# The stress tool's scrapes are best-effort, so fail here if a node's containerd
# metrics listener is missing.
node_exporter_pod_names=$(kubectl --namespace kube-system get pods --selector=app.kubernetes.io/name=node-exporter -o jsonpath='{.items[*].metadata.name}')
read -r -a node_exporter_pods <<< "${node_exporter_pod_names}"
for pod in "${node_exporter_pods[@]:?No node-exporter pods found}"; do
  kubectl --request-timeout=10s get --raw "/api/v1/namespaces/kube-system/pods/http:${pod}:1338/proxy/v1/metrics" > /dev/null
done

# CONTROLLER_ARGS appends controller flags, e.g. --enable-pprof-debug so the
# stress tool can capture controller profiles during claims phases.
dev/tools/deploy-to-kube --image-prefix="${IMAGE_PREFIX}" --image-tag="${IMAGE_TAG}" --extensions \
  ${CONTROLLER_ARGS:+--controller-args="${CONTROLLER_ARGS}"}
kubectl rollout status deployment agent-sandbox-controller --namespace agent-sandbox-system --timeout=10m

dev/tools/test-e2e --suite=benchmarks --image-prefix="${IMAGE_PREFIX}"

STRESS_OUTPUT_DIR="${ARTIFACTS:-${WORKSPACE}}/stress-test"
mkdir -p "${STRESS_OUTPUT_DIR}"

# Phases are documented in test/stress/phase.go. The default list is sized for
# the 3-worker default: launch latency, then throughput at mif200, mif100 and
# mif50 on a mostly empty cluster. Throughput phases run for at least 300 s
# (the tool's default is 45 s): steady-state throughput does not depend on the
# phase length, and the extra pod churn per worker makes the node memory and
# cgroup trends readable. The longer phases need the 90m tool timeout. The
# scrape interval stays at the tool's 15 s default, exposed as a knob.
set +o errexit
# STRESS_EXTRA_ARGS is word-split on purpose, so callers can pass tool flags
# verbatim (e.g. --client-connections=4 for the claims profile).
# shellcheck disable=SC2086
go run ./test/stress \
  --phases="${STRESS_PHASES:-fill,probe,throughput-mif:200,throughput-mif:100,throughput-mif:50}" \
  --fill-per-node="${STRESS_FILL_PER_NODE:-10}" \
  --probe-count="${STRESS_PROBE_COUNT:-30}" \
  --throughput-count="${STRESS_THROUGHPUT_COUNT:-500}" \
  --throughput-min-seconds="${STRESS_THROUGHPUT_MIN_SECONDS:-300}" \
  --create-concurrency="${STRESS_CREATE_CONCURRENCY:-50}" \
  --claims-warm-count="${STRESS_CLAIMS_WARM:-300}" \
  --timeout="${STRESS_TIMEOUT:-90m}" \
  --metrics-interval="${STRESS_METRICS_INTERVAL:-15s}" \
  --output-dir="${STRESS_OUTPUT_DIR}" \
  ${STRESS_EXTRA_ARGS:-}
stress_status=$?
set -o errexit

kubectl logs deployment/agent-sandbox-controller --namespace agent-sandbox-system --tail=-1 > "${STRESS_OUTPUT_DIR}/controller.log" || true
kubectl get pods -A -o wide > "${STRESS_OUTPUT_DIR}/pods.txt" || true
kubectl get nodes -o wide > "${STRESS_OUTPUT_DIR}/nodes.txt" || true

# The HTML report is rendered next to the raw data it is built from.
REPORT_OUTPUT_DIR="${STRESS_OUTPUT_DIR}/report"
python3 -m venv "${WORKSPACE}/report-venv"
"${WORKSPACE}/report-venv/bin/pip" install --quiet 'duckdb>=1,<2' 'jinja2>=3,<4'
"${WORKSPACE}/report-venv/bin/python3" test/stress/generate-report/generate_report.py \
  --input-dir "${STRESS_OUTPUT_DIR}" \
  --output-dir "${REPORT_OUTPUT_DIR}"
echo "Stress test report written to ${REPORT_OUTPUT_DIR}"

# Spyglass renders artifacts/*.link.txt files as links on the prow job page. A
# gs:// URL opens directly in the GCS browser, while an https:// URL goes
# through a redirect interstitial.
if [[ -n "${ARTIFACTS:-}" && -n "${JOB_NAME:-}" && -n "${BUILD_ID:-}" ]]; then
  if [[ -n "${PULL_NUMBER:-}" ]]; then
    JOB_GCS_PATH="pr-logs/pull/${REPO_OWNER:-kubernetes}_${REPO_NAME:-kops}/${PULL_NUMBER}/${JOB_NAME}/${BUILD_ID}"
  else
    JOB_GCS_PATH="logs/${JOB_NAME}/${BUILD_ID}"
  fi
  REPORT_PATH="kubernetes-ci-logs/${JOB_GCS_PATH}/artifacts/stress-test/report/index.html"
  echo "gs://${REPORT_PATH}" > "${ARTIFACTS}/stress-test.link.txt"
  echo "Stress test report: https://storage.googleapis.com/${REPORT_PATH}"
fi

exit "${stress_status}"
