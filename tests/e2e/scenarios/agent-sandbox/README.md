# agent-sandbox benchmark

Runs the [kubernetes-sigs/agent-sandbox](https://github.com/kubernetes-sigs/agent-sandbox)
benchmark and stress test (`test/stress` in that repository) against a kOps cluster on GCE
(the default), AWS or Azure, selected by `CLOUD_PROVIDER`.

Upstream runs this test in its `benchmarks-kops-gcp-*` jobs with a pinned kOps release.
This scenario runs it with the kOps build under test, so kOps changes that affect pod
launch throughput or control-plane behaviour show up in presubmits.

The cluster tuning is the same as upstream's `test/benchmarks/scenarios/benchmarks-kops-gcp/run`,
and the `STRESS_*` variables keep the same names and defaults, so results are comparable with
the upstream jobs. The controller is deployed from the published `registry.k8s.io/agent-sandbox/`
images, and the agent-sandbox checkout is pinned to the same release tag, so no images are built
and no container registry is needed on any cloud.

### Running locally

Requires `kubectl`, `go` and `python3`, plus the cloud's credentials below. kubetest2-kops
creates the cluster (and tears it down on exit), resolves an SSH key pair, and reads the
cloud credentials from the environment. Set a cluster name on every cloud; DNS defaults to
`none`, so a `.k8s.local` name works everywhere:

```
export CLUSTER_NAME=agent-sandbox.k8s.local
```

**GCE** (default): requires `gcloud` (authenticated). Set the project and a state store:

```
export GCP_PROJECT=...
export KOPS_STATE_STORE=gs://...
```

**AWS**: credentials come from the standard AWS credential chain (`AWS_PROFILE`,
`AWS_SHARED_CREDENTIALS_FILE` or the `AWS_*` key variables). `KOPS_STATE_STORE` is optional:
when unset, kubetest2-kops creates an S3 bucket for the run and deletes it on teardown.

```
export CLOUD_PROVIDER=aws
export KOPS_STATE_STORE=s3://...
```

**Azure**: kOps' Azure support is alpha and feature-gated
([getting started](https://kops.sigs.k8s.io/getting_started/azure/)); the script exports
`KOPS_FEATURE_FLAGS=Azure` itself. kubetest2-kops forwards only `AZURE_TENANT_ID`,
`AZURE_SUBSCRIPTION_ID`, `AZURE_CLIENT_ID` and `AZURE_FEDERATED_TOKEN_FILE` to kops, so
authenticate with `az login` (the Azure CLI credential) or with workload identity; a client
secret in `AZURE_CLIENT_SECRET` is not forwarded. `AZURE_SUBSCRIPTION_ID` is required. Set a
state store on a storage account where the identity has Storage Blob Data Contributor;
when unset, kubetest2-kops uses the kOps CI account's container.

```
export CLOUD_PROVIDER=azure
export AZURE_SUBSCRIPTION_ID=...
export KOPS_STATE_STORE=azureblob://<storage-account>/<container>
```

The subscription needs quota for the `Standard_D8s_v5` workers and the `Standard_D32s_v5`
control plane in the zone: about 192 vCPUs for the full profile. The default zone is
`eastus-1`; kubetest2-kops' own Azure e2e jobs run in `uksouth`, so set `ZONES=uksouth-1` if
that is where the subscription's quota is.

By default the script builds kOps from the current tree and stages it in GCS. To test the
latest kOps CI build instead, set `JOB_TYPE=periodic`.

The full upstream periodic profile uses 20 workers. For a smaller run, use the upstream
cilium presubmit profile, with a control plane of the same family as the full profile:

```
export STRESS_NODE_COUNT=3
export STRESS_CONTROL_PLANE_SIZE=c3-standard-8      # gce
export STRESS_CONTROL_PLANE_SIZE=c7i.2xlarge        # aws
export STRESS_CONTROL_PLANE_SIZE=Standard_D8s_v5    # azure
export STRESS_PHASES=fill,probe,throughput-mif:200,throughput-mif:100,throughput-mif:50,fill-pct:80,throughput-mif:50-label:pct80
```

To compare cgroup drivers, run a second time with `CGROUP_DRIVER=cgroupfs`. kOps 1.34+
always configures the systemd cgroup driver, so this switches containerd's runc
`SystemdCgroup` option off through `spec.containerd.configAdditions`, and the kubelet
follows the runtime's driver.

Then run:

```
tests/e2e/scenarios/agent-sandbox/run-test.sh
```

### Per-cloud defaults

The GCE shapes are the measured upstream ones. AWS and Azure use each cloud's closest
shapes: 8 vCPU / 32 GiB workers like `n2-standard-8`, and a 32 vCPU control plane, as
neither cloud has a 22 vCPU size like `c3-standard-22`.

| | GCE | AWS | Azure |
|---|---|---|---|
| `ZONES` | `us-central1-a` | `us-east-1a` | `eastus-1` |
| `STRESS_CONTROL_PLANE_SIZE` | `c3-standard-22` | `c7i.8xlarge` | `Standard_D32s_v5` |
| `STRESS_NODE_SIZE` | `n2-standard-8` | `m6i.2xlarge` | `Standard_D8s_v5` |
| `STRESS_NODE_VOLUME_TYPE` | `pd-ssd` | `gp3` | `Premium_LRS` |
| Control-plane root volume type | `pd-ssd` | `gp3` (kOps default) | `StandardSSD_LRS` (kOps default) |
| Small profile `STRESS_CONTROL_PLANE_SIZE` | `c3-standard-8` | `c7i.2xlarge` | `Standard_D8s_v5` |

On GCE the worker roots are `pd-ssd` because pod launch is fsync-bound on `pd-standard`
(kOps' GCE default), and the `c3` control-plane family does not support `pd-standard` boot
disks. `gp3` is kOps' AWS default and has a 3000 IOPS baseline at any size. On Azure disk
IOPS scale with size, and both `StandardSSD_LRS` and `Premium_LRS` cap at 500 IOPS on the
128 GB root, so raise `STRESS_NODE_VOLUME_SIZE` if runs come out disk-bound.

### Variables

| Variable | Default | Purpose |
|---|---|---|
| `CLOUD_PROVIDER` | `gce` | `gce`, `aws` or `azure` |
| `CNI` | `cilium` | kOps `--networking` value; upstream CI runs `cilium` and `kindnet` |
| `CGROUP_DRIVER` | `systemd` | `cgroupfs` switches containerd and the kubelet off the systemd cgroup driver |
| `ZONES` | per cloud, see above | Zone to create the cluster in |
| `K8S_VERSION` | `stable` | Kubernetes version passed to kubetest2 |
| `AGENT_SANDBOX_VERSION` | newest tag published on `registry.k8s.io` | agent-sandbox tag to check out and deploy |
| `AGENT_SANDBOX_ROOT` | fresh clone in the workspace | Use an existing agent-sandbox checkout as is |
| `IMAGE_PREFIX` / `IMAGE_TAG` | `registry.k8s.io/agent-sandbox/` / `AGENT_SANDBOX_VERSION` | Controller images to deploy |
| `CONTROLLER_ARGS` | unset | Extra controller flags, e.g. `--enable-pprof-debug` |
| `STRESS_NODE_COUNT` | `20` | Worker node count |
| `STRESS_CONTROL_PLANE_SIZE` | per cloud, see above | Control-plane machine type |
| `STRESS_NODE_SIZE` | per cloud, see above | Worker machine type |
| `STRESS_NODE_VOLUME_TYPE` | per cloud, see above | Worker root volume type |
| `STRESS_NODE_VOLUME_SIZE` | `128` | Worker root volume size in GB |
| `STRESS_PHASES` | upstream periodic phase list | Stress tool phases, see `test/stress/phase.go` upstream |
| `STRESS_EXTRA_ARGS` | unset | Extra stress tool flags, passed verbatim |
| `STRESS_VALIDATE_WAIT` | `25m` | Cluster validation timeout |

The other `STRESS_*` knobs (`STRESS_FILL_PER_NODE`, `STRESS_PROBE_COUNT`, `STRESS_THROUGHPUT_COUNT`,
`STRESS_THROUGHPUT_MIN_SECONDS`, `STRESS_CREATE_CONCURRENCY`, `STRESS_CLAIMS_WARM`, `STRESS_TIMEOUT`,
`STRESS_NODE_CIDR_MASK_SIZE`, `STRESS_ETCD_QUOTA_BYTES`, `STRESS_KINDNET_NETWORK_POLICIES`) match
the upstream script.

### Output

The stress tool's raw data, the controller log and the rendered HTML report are written to
`${ARTIFACTS}/stress-test/` (or the temporary workspace when `ARTIFACTS` is unset). In prow,
the job page links to the report through `stress-test.link.txt`.
