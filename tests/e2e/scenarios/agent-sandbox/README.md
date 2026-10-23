# agent-sandbox benchmark

Runs the [kubernetes-sigs/agent-sandbox](https://github.com/kubernetes-sigs/agent-sandbox)
benchmark and stress test (`test/stress` in that repository) against a kOps cluster on GCE
(the default), AWS or Azure, selected by `CLOUD_PROVIDER`.

Upstream runs this test on GCE in its `benchmarks-kops-gcp-*` jobs with a pinned kOps release.
This scenario runs it with the kOps build under test, so kOps changes that affect pod
launch throughput or control-plane behaviour show up in presubmits.

The cluster tuning follows upstream's `test/benchmarks/scenarios/benchmarks-kops-gcp/run`, which
documents the measurements behind each knob, and the `STRESS_*` variables keep its names. The
controller is deployed from the published `registry.k8s.io/agent-sandbox/` images, and the
agent-sandbox checkout is pinned to the same release tag, so no images are built. AWS and Azure
need agent-sandbox v1.0.8 or newer: the e2e framework of older releases accepts only GCE node
provider ids, and their Chrome readiness probe never succeeds from a prow runner on another
continent.

### Running locally

Requires `kubectl`, `go`, `python3`, an authenticated `gcloud`, the cloud's credentials below and
a cluster name. `gcloud` is needed on every cloud: by default the script builds kOps from the
current tree and stages it in the kOps staging GCS bucket, which needs write access to that
bucket; `JOB_TYPE=periodic` downloads the latest kOps CI build instead. DNS defaults to `none`,
so a `.k8s.local` name works on every cloud:

```
export CLUSTER_NAME=agent-sandbox.k8s.local
```

**GCE** (default): set the GCP project and the kOps state store:

```
export GCP_PROJECT=...
export KOPS_STATE_STORE=gs://...
```

**AWS**: kubetest2-kops passes kops only `AWS_PROFILE`, `AWS_SHARED_CREDENTIALS_FILE` and the
container-credential variables, so keep the credentials in a shared credentials file or a
container credential provider. Set the kOps state store:

```
export CLOUD_PROVIDER=aws
export KOPS_STATE_STORE=s3://...
```

**Azure**: kOps' Azure support is alpha and feature-gated
([getting started](https://kops.sigs.k8s.io/getting_started/azure/)); the script adds `Azure` to
`KOPS_FEATURE_FLAGS` itself. kubetest2-kops passes kops only `AZURE_TENANT_ID`,
`AZURE_SUBSCRIPTION_ID`, `AZURE_CLIENT_ID` and `AZURE_FEDERATED_TOKEN_FILE`, so authenticate with
`az login` or workload identity; `AZURE_SUBSCRIPTION_ID` is required. Set the kOps state store:

```
export CLOUD_PROVIDER=azure
export AZURE_SUBSCRIPTION_ID=...
export KOPS_STATE_STORE=azureblob://<storage-account>/<container>
```

The defaults run 3 workers behind a control plane of the same size (see the table below) with a
phase list that tops out at mif200, the profile the kOps CI jobs run. The control-plane size and
the phase list do not follow the worker count, so a larger run sets them as well, for example on
GCE:

```
export STRESS_NODE_COUNT=20
export STRESS_CONTROL_PLANE_SIZE=c3-standard-22
export STRESS_PHASES=fill,probe,throughput-mif:600,throughput-mif:400,throughput-mif:200,throughput-mif:100,throughput-mif:50
```

To compare cgroup drivers, run a second time with `CGROUP_DRIVER=cgroupfs`, which sets
`spec.kubelet.cgroupDriver: cgroupfs`; kOps sets containerd's runc `SystemdCgroup` option to
match.

Then run:

```
tests/e2e/scenarios/agent-sandbox/run-test.sh
```

### Per-cloud defaults

Workers and control planes have 8 vCPUs / 32 GiB on every cloud; AWS and Azure use each cloud's
closest match to the GCE shapes.

| | GCE | AWS | Azure |
|---|---|---|---|
| `ZONES` | `us-central1-a` | chosen per run | `uksouth-1` or `uksouth-2` |
| `STRESS_CONTROL_PLANE_SIZE` | `c3-standard-8` | `m7i.2xlarge` | `Standard_D8ds_v5` |
| `STRESS_NODE_SIZE` | `n2-standard-8` | `m6i.2xlarge` | `Standard_D8ds_v5` |
| `STRESS_NODE_VOLUME_TYPE` | `pd-ssd` | `gp3` at 16000 IOPS and 1000 MB/s | `ResourceDisk` (ephemeral OS disk) |
| Control-plane root volume type | `pd-ssd` | `gp3` (kOps default) | `StandardSSD_LRS` (kOps default) |

On AWS no zone is pinned: kubetest2-kops picks a zone that offers both instance types, in a
random region per run, so a capacity shortage in one zone does not fail every run. Set `ZONES`
to pin one. On Azure kubetest2-kops picks one of the `uksouth` zones the other kOps Azure jobs
use. The Ddsv5 sizes are the Dsv5 ones plus a local temp disk, with a larger quota in the kOps CI
subscription, and the worker roots are
[ephemeral OS disks](https://kops.sigs.k8s.io/getting_started/azure/#ephemeral-os-disks) on that
temp disk: pod launch is fsync-bound, and a 128 GB `Premium_LRS` root (P10, 500 IOPS) starts pods
three to four times slower than `gp3`. On AWS the worker roots are `gp3` provisioned at 16000 IOPS
and 1000 MB/s: at the 3000 IOPS default the disks ran 89% busy at mif 200 with 3 ms average writes
and 12% iowait while the CPUs were 37% busy, and containerd's container create averaged 840 ms
against 16 ms on Azure and 51 ms on GCE. On GCE the roots are `pd-ssd` because pod launch is
fsync-bound on `pd-standard`, and the `c3` family does not support `pd-standard` boot disks.

etcd runs without fsync (`STRESS_ETCD_UNSAFE_NO_FSYNC`): the WAL fsync on the etcd volumes
(`StandardSSD_LRS` on Azure, `gp3` on AWS, `pd-ssd` on GCE) paces the control plane, p99 55 ms on
Azure against 5 ms on gp3, and the launch latency tails follow it.

### Variables

| Variable | Default | Purpose |
|---|---|---|
| `CLOUD_PROVIDER` | `gce` | `gce`, `aws` or `azure` |
| `CNI` | `kindnet` | kOps `--networking` value; `cilium` runs the Cilium variant with the agent's rate limits raised |
| `CGROUP_DRIVER` | `systemd` | `cgroupfs` switches containerd and the kubelet off the systemd cgroup driver |
| `ZONES` | per cloud, see above | Zone to create the cluster in |
| `K8S_VERSION` | `stable` | Kubernetes version passed to kubetest2 |
| `AGENT_SANDBOX_VERSION` | newest tag published on `registry.k8s.io` | agent-sandbox tag to check out and deploy; v1.0.8 or newer for AWS and Azure |
| `AGENT_SANDBOX_ROOT` | fresh clone in the workspace | Use an existing agent-sandbox checkout as is |
| `IMAGE_PREFIX` / `IMAGE_TAG` | `registry.k8s.io/agent-sandbox/` / `AGENT_SANDBOX_VERSION` | Controller images to deploy |
| `CONTROLLER_ARGS` | unset | Extra controller flags, e.g. `--enable-pprof-debug` |
| `STRESS_NODE_COUNT` | `3` | Worker node count |
| `STRESS_CONTROL_PLANE_SIZE` | per cloud, see above | Control-plane machine type |
| `STRESS_NODE_SIZE` | per cloud, see above | Worker machine type |
| `STRESS_NODE_VOLUME_TYPE` | per cloud, see above | Worker root volume type |
| `STRESS_PHASES` | `fill,probe,throughput-mif:200,throughput-mif:100,throughput-mif:50` | Stress tool phases, see `test/stress/phase.go` upstream |
| `STRESS_THROUGHPUT_MIN_SECONDS` | `300` | Minimum length of each throughput phase (tool default `45`); the extra pod churn per worker makes the node memory and cgroup trends readable |
| `STRESS_TIMEOUT` | `90m` | Stress tool run timeout (tool default `30m`), for the longer throughput phases |
| `STRESS_METRICS_INTERVAL` | `15s` | Interval between the stress tool's Prometheus scrapes (the tool default) |
| `STRESS_EXTRA_ARGS` | unset | Extra stress tool flags, passed verbatim |
| `STRESS_VALIDATE_WAIT` | `25m` | Cluster validation timeout |
| `STRESS_ETCD_UNSAFE_NO_FSYNC` | `true` | Sets `ETCD_UNSAFE_NO_FSYNC` in every etcd cluster's `manager.env`, which etcd-manager passes on to etcd, so the etcd volume's fsync latency stays out of the launch latency; the clusters are discarded after the run. `false` keeps fsync |

The other `STRESS_*` knobs (`STRESS_FILL_PER_NODE`, `STRESS_PROBE_COUNT`, `STRESS_THROUGHPUT_COUNT`,
`STRESS_CREATE_CONCURRENCY`, `STRESS_CLAIMS_WARM`, `STRESS_NODE_CIDR_MASK_SIZE`,
`STRESS_ETCD_QUOTA_BYTES`, `STRESS_KINDNET_NETWORK_POLICIES`) match the upstream script.

### Output

The stress tool's raw data, the controller log and the rendered HTML report are written to
`${ARTIFACTS}/stress-test/` (or the temporary workspace when `ARTIFACTS` is unset). In prow,
the job page links to the report through `stress-test.link.txt`.

`controller.log` follows the controller for the whole stress run, about 50 KB/s under churn. If
the stream ends early, on a dropped connection or a controller restart, `controller-tail.log`
holds the end-of-run snapshot as well.
