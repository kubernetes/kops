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

Requires a linux/amd64 host with `kubectl`, `go`, `python3` and an authenticated `gcloud`, the
cloud's credentials below and a cluster name. `gcloud` is needed on every cloud: by default the script builds kOps from the
current tree and stages it in the kOps staging GCS bucket, which needs write access to that
bucket; `JOB_TYPE=periodic` downloads the latest kOps CI build instead. DNS defaults to `none`,
so a `.k8s.local` name works on every cloud:

```
export CLUSTER_NAME=agent-sandbox.k8s.local
```

**GCE** (default): set the GCP project and leave `KOPS_STATE_STORE` unset: kubetest2-kops creates
a bucket for the run and deletes it on teardown, and it deletes a bucket you set as well.

```
export GCP_PROJECT=...
```

**AWS**: kubetest2-kops passes kops only `AWS_PROFILE`, `AWS_SHARED_CREDENTIALS_FILE` and the
container-credential variables, so keep the credentials in a shared credentials file or a
container credential provider. Set the kOps state store; a bucket kubetest2-kops creates for the
run is not deleted on teardown:

```
export CLOUD_PROVIDER=aws
export KOPS_STATE_STORE=s3://...
```

**Azure**: kOps' Azure support is alpha and feature-gated
([getting started](https://kops.sigs.k8s.io/getting_started/azure/)); the script adds `Azure` to
`KOPS_FEATURE_FLAGS` itself. kubetest2-kops passes kops only `AZURE_TENANT_ID`,
`AZURE_SUBSCRIPTION_ID`, `AZURE_CLIENT_ID` and `AZURE_FEDERATED_TOKEN_FILE`, so authenticate with
`az login` or workload identity; a client secret in `AZURE_CLIENT_SECRET` is not passed on.
`AZURE_SUBSCRIPTION_ID` is required. Set the kOps state store on a storage account where the
identity has Storage Blob Data Contributor; when unset, kubetest2-kops uses the kOps CI account's
container:

```
export CLOUD_PROVIDER=azure
export AZURE_SUBSCRIPTION_ID=...
export KOPS_STATE_STORE=azureblob://<storage-account>/<container>
```

The defaults run 3 workers behind a control plane of the same size (see the table below) with a
single mif300 throughput phase, the profile the kOps CI jobs run. The control-plane size and
the phase list do not follow the worker count, so a larger run sets them as well, for example on
GCE:

```
export STRESS_NODE_COUNT=20
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

Workers have 16 vCPUs / 64 GiB on every cloud: a pod lifecycle costs about 1 CPU-second on the
node, so an 8 vCPU worker tops out below 10 pods/s on CPU alone. The GCE control plane is
`c3-standard-22`, the c3 size above 8 vCPUs; AWS and Azure use 16 vCPU control planes, each
cloud's closest match.

| | GCE | AWS | Azure |
|---|---|---|---|
| `ZONES` | `us-central1-a` | chosen per run | `uksouth-1` or `uksouth-2` |
| `STRESS_CONTROL_PLANE_SIZE` | `c3-standard-22` | `m7i.4xlarge` | `Standard_D16ds_v5` |
| `STRESS_NODE_SIZE` | `n2-standard-16` | `m6i.4xlarge` | `Standard_D16ds_v5` |
| `STRESS_NODE_VOLUME_TYPE` | `pd-ssd` | `gp3`, provisioned IOPS and throughput | `ResourceDisk` (ephemeral OS disk) |
| `INSTANCE_IMAGE` | Ubuntu 26.10 `v20261014` | Ubuntu 26.10 `20261020` | Ubuntu 26.10 `26.10.202610200` |
| Control-plane root volume type | `pd-ssd` | `gp3` (kOps default) | `StandardSSD_LRS` (kOps default) |

On AWS no zone is pinned: kubetest2-kops picks a zone that offers both instance types, in a
random region per run, so a capacity shortage in one zone does not fail every run. Set `ZONES`
to pin one. On Azure kubetest2-kops picks one of the `uksouth` zones the other kOps Azure jobs
use. The Ddsv5 sizes are the Dsv5 ones plus a local temp disk, with a larger quota in the kOps CI
subscription, and the worker roots are
[ephemeral OS disks](https://kops.sigs.k8s.io/getting_started/azure/#ephemeral-os-disks) on that
temp disk: pod launch is fsync-bound, and a 128 GB `Premium_LRS` root (P10, 500 IOPS) starts pods
three to four times slower than `gp3`. On AWS the worker roots are `gp3` with provisioned IOPS and
throughput: gp3's default IOPS saturate under pod churn and inflate containerd's create
operations. On GCE the roots are `pd-ssd` because pod launch is
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
| `INSTANCE_IMAGE` | per cloud, see above | kOps `--image` for all instance groups; Ubuntu 26.10 on every cloud, because the 24.04 kernel keeps dying memory cgroups pinned under pod churn |
| `K8S_VERSION` | `https://dl.k8s.io/ci/latest.txt` (latest dev build) | Kubernetes version or marker URL passed to kubetest2-kops |
| `AGENT_SANDBOX_VERSION` | newest tag published on `registry.k8s.io` | agent-sandbox tag to check out and deploy; v1.0.8 or newer for AWS and Azure |
| `AGENT_SANDBOX_ROOT` | fresh clone in the workspace | Use an existing agent-sandbox checkout as is; `patches/` are applied only to the fresh clone |
| `IMAGE_PREFIX` / `IMAGE_TAG` | `registry.k8s.io/agent-sandbox/` / `AGENT_SANDBOX_VERSION` | Controller images to deploy |
| `CONTROLLER_ARGS` | unset | Extra controller flags, e.g. `--enable-pprof-debug` |
| `STRESS_NODE_COUNT` | `3` | Worker node count |
| `STRESS_CONTROL_PLANE_SIZE` | per cloud, see above | Control-plane machine type |
| `STRESS_NODE_SIZE` | per cloud, see above | Worker machine type |
| `STRESS_NODE_VOLUME_TYPE` | per cloud, see above | Worker root volume type |
| `STRESS_PHASES` | `fill,probe,throughput-mif:300` | Stress tool phases, see `test/stress/phase.go` upstream |
| `STRESS_THROUGHPUT_MIN_SECONDS` | `300` | Minimum length of each throughput phase (tool default `45`); the extra pod churn per worker makes the node memory and cgroup trends readable |
| `STRESS_TIMEOUT` | `90m` | Stress tool run timeout (tool default `30m`), for the longer throughput phase |
| `STRESS_METRICS_INTERVAL` | `15s` | Interval between the stress tool's Prometheus scrapes (the tool default) |
| `STRESS_EXTRA_ARGS` | unset | Extra stress tool flags, word-split and appended to the command |
| `STRESS_VALIDATE_WAIT` | `25m` | Cluster validation timeout |
| `STRESS_ETCD_UNSAFE_NO_FSYNC` | `true` | Sets `ETCD_UNSAFE_NO_FSYNC` in every etcd cluster's `manager.env`, which etcd-manager passes on to etcd, so the etcd volume's fsync latency stays out of the launch latency; the clusters are discarded after the run. `false` keeps fsync |

The other `STRESS_*` knobs (`STRESS_FILL_PER_NODE`, `STRESS_PROBE_COUNT`, `STRESS_THROUGHPUT_COUNT`,
`STRESS_CREATE_CONCURRENCY`, `STRESS_CLAIMS_WARM`, `STRESS_NODE_CIDR_MASK_SIZE`,
`STRESS_ETCD_QUOTA_BYTES`, `STRESS_KINDNET_NETWORK_POLICIES`, `STRESS_CILIUM_VALIDATE_WAIT`) match
the upstream script.

### Output

The stress tool's raw data, the controller log and the rendered HTML report are written to
`${ARTIFACTS}/stress-test/` (or the temporary workspace when `ARTIFACTS` is unset). In prow,
the job page links to the report through `stress-test.link.txt`.

`controller.log` follows the controller for the whole stress run, about 50 KB/s under churn. If
the stream ends early, on a dropped connection or a controller restart, `controller-tail.log`
holds the end-of-run snapshot as well.

With kindnet, the script sets `CNI_LOG_FILE=/var/log/cni-kindnet.log` in containerd's environment
(`spec.containerd.env`). The cni-kindnet plugin inherits it and appends its warnings and errors,
such as failed conntrack cleanups and IPAM errors, to that file, which the node dump copies to
`${ARTIFACTS}/<node>/cni-kindnet.log`. A CNI call without warnings or errors writes nothing.
kindnet runs from `docker.io/hakman/kindnet:v1.0.2-23-g32cbbf7` (`spec.networking.kindnet.image`),
upstream main with the fix for the pod interface creation retry, so each failed attempt shows up
in that log as a `failed attempt` line.

`patches/` holds upstream changes the agent-sandbox tag lacks; the fresh clone gets them with
`git apply`, and a patch a tag already carries is skipped. `node-exporter-collectors.patch` adds
the pressure, stat, schedstat, processes and vmstat collectors and the cpu collector's cpu.info
option to the node-exporter manifest, so the metrics file and the report's explorer carry PSI
(`node_pressure_*`), tasks blocked on I/O (`node_procs_blocked`), run-queue wait
(`node_schedstat_waiting_seconds_total`), thread counts per state
(`node_processes_threads_state`), page-cache dirtying and writeback (`node_vmstat_nr_*`) and the
CPU model (`node_cpu_info`).
