# agent-sandbox benchmark

Runs the [kubernetes-sigs/agent-sandbox](https://github.com/kubernetes-sigs/agent-sandbox)
benchmark and stress test (`test/stress` in that repository) against a kOps cluster on GCE.

Upstream runs this test in its `benchmarks-kops-gcp-*` jobs with a pinned kOps release.
This scenario runs it with the kOps build under test, so kOps changes that affect pod
launch throughput or control-plane behaviour show up in presubmits.

The cluster tuning is the same as upstream's `test/benchmarks/scenarios/benchmarks-kops-gcp/run`,
and the `STRESS_*` variables keep the same names and defaults, so results are comparable with
the upstream jobs. The controller is deployed from the published `registry.k8s.io/agent-sandbox/`
images, and the agent-sandbox checkout is pinned to the same release tag, so no images are built.

### Running locally

Requires `gcloud` (authenticated), `kubectl`, `go` and `python3`.

Set the GCP project, the kOps state store and a cluster name:

```
export GCP_PROJECT=...
export KOPS_STATE_STORE=gs://...
export CLUSTER_NAME=agent-sandbox.k8s.local
```

By default the script builds kOps from the current tree and stages it in GCS. To test the
latest kOps CI build instead, set `JOB_TYPE=periodic`.

The full upstream periodic profile uses 20 `n2-standard-8` workers. For a smaller run, use the
upstream cilium presubmit profile:

```
export STRESS_NODE_COUNT=3
export STRESS_CONTROL_PLANE_SIZE=c3-standard-8
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

### Variables

| Variable | Default | Purpose |
|---|---|---|
| `CNI` | `cilium` | kOps `--networking` value; upstream CI runs `cilium` and `kindnet` |
| `CGROUP_DRIVER` | `systemd` | `cgroupfs` switches containerd and the kubelet off the systemd cgroup driver |
| `ZONES` | `us-central1-a` | GCE zone |
| `K8S_VERSION` | `stable` | Kubernetes version passed to kubetest2 |
| `AGENT_SANDBOX_VERSION` | newest tag published on `registry.k8s.io` | agent-sandbox tag to check out and deploy |
| `AGENT_SANDBOX_ROOT` | fresh clone in the workspace | Use an existing agent-sandbox checkout as is |
| `IMAGE_PREFIX` / `IMAGE_TAG` | `registry.k8s.io/agent-sandbox/` / `AGENT_SANDBOX_VERSION` | Controller images to deploy |
| `CONTROLLER_ARGS` | unset | Extra controller flags, e.g. `--enable-pprof-debug` |
| `STRESS_NODE_COUNT` | `20` | Worker node count |
| `STRESS_CONTROL_PLANE_SIZE` | `c3-standard-22` | Control-plane machine type |
| `STRESS_NODE_VOLUME_TYPE` | `pd-ssd` | Worker root volume type |
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
