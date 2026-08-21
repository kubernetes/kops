# E2E Failure Troubleshooting Guide

This guide describes how to investigate kops E2E CI job failures, identify root causes, and find correlations across groups of failing jobs.

## E2E Job Structure

### Job Naming Conventions

Periodic jobs follow these naming patterns:
- `e2e-kops-grid-{networking}-{distro}-k{k8s_version}-ko{kops_version}` — grid matrix jobs testing combinations of networking, distro, and version
- `e2e-kops-{cloud}-{feature/scenario}` — feature-specific tests (e.g., load balancer controller, nftables, upgrades)
- `ci-kubernetes-kops-{cloud}-{scenario}` — upstream Kubernetes CI jobs that use kops

Presubmit jobs use the pattern:
- `pull-kops-{description}` — triggered by pull requests

### Job Metadata Dimensions

Each job is annotated with `test.kops.k8s.io/*` keys in its Prow config that describe its test parameters:

| Dimension | Annotation | Example Values |
|---|---|---|
| Cloud provider | `test.kops.k8s.io/cloud` | aws, gce, digitalocean, azure |
| OS distribution | `test.kops.k8s.io/distro` | al2023, deb11, deb12, flatcar, rhel9, rocky9, u2404 |
| Networking plugin | `test.kops.k8s.io/networking` | cilium, calico, amazonvpc, kubenet, flannel, kuberouter, cilium-eni, kindnet |
| Kubernetes version | `test.kops.k8s.io/k8s_version` | ci, stable, 1.32, 1.33, 1.34 |
| kops version | `test.kops.k8s.io/kops_version` | latest, specific version markers |

These dimensions are critical for correlation analysis (see below).

### Job Definitions

Job configs are auto-generated YAML files located in the `test-infra` repository at `config/jobs/kubernetes/kops/`:
- `kops-periodics-*.yaml` — periodic job definitions (grid, conformance, distros, gce, misc, network-plugins, nftables, pipeline, upgrades, versions, ai-conformance)
- `kops-presubmits*.yaml` — presubmit job definitions (main, ai-conformance, branch, distros, e2e, network-plugins, scale)
- `kops-presets*.yaml` — shared credential and config presets

### Periodic vs Presubmit Jobs

| Aspect | Periodic Jobs | Presubmit Jobs |
|---|---|---|
| YAML root key | `periodics:` | `presubmits: kubernetes/kops:` |
| Trigger | `cron:` schedule | Pull request events |
| Name prefix | `e2e-kops-*` | `pull-kops-*` |
| GCS path | `logs/{job-name}/` | `pr-logs/directory/{job-name}/` |

## Artifact Locations

### GCS Bucket

All job artifacts are stored in the `kubernetes-ci-logs` GCS bucket.

**Finding the latest build:**
```
https://storage.googleapis.com/kubernetes-ci-logs/logs/{job-name}/latest-build.txt
```

**Build artifacts base URL:**
```
# Periodic jobs:
https://storage.googleapis.com/kubernetes-ci-logs/logs/{job-name}/{build-id}/

# Presubmit (PR) jobs:
https://storage.googleapis.com/kubernetes-ci-logs/pr-logs/directory/{job-name}/{build-id}/
```

### Prow UI

Each build is viewable at:
```
# Periodic:
https://prow.k8s.io/view/gs/kubernetes-ci-logs/logs/{job-name}/{build-id}

# Presubmit:
https://prow.k8s.io/view/gs/kubernetes-ci-logs/pr-logs/directory/{job-name}/{build-id}
```

### Testgrid Dashboards

- Periodic jobs: `https://testgrid.k8s.io/sig-cluster-lifecycle-kops`
- Presubmit jobs: `https://testgrid.k8s.io/kops-presubmits`

### Failures Dashboard

The failures dashboard at `https://storage.googleapis.com/k8s-metrics/failures-latest.html` tracks all consecutively failing Prow jobs. It loads data from `failures-latest.json` in the same bucket. Each entry contains the job name and consecutive failing days. The dashboard classifies severity as: Critical (365+ days), Severe (180-364), Warning (90-179), Moderate (30-89), Low (1-29).

## Artifact Structure

Each build directory contains the following hierarchy:

```
{build-id}/
├── build-log.txt              # Overall test runner stdout/stderr
├── finished.json              # Build result: {"result": "SUCCESS"|"FAILURE", ...}
├── started.json               # Build start metadata
└── artifacts/
    ├── junit_*.xml            # JUnit test results
    ├── toolbox-dump.yaml      # kops toolbox dump output
    ├── cluster.yaml           # kops get cluster -o yaml
    ├── instancegroups.yaml    # kops get instancegroups -o yaml
    ├── cluster-info/          # Kubernetes API resource dumps and pod logs
    │   ├── nodes.yaml         # All Node objects (status, conditions, capacity, addresses)
    │   ├── {resource}.yaml    # Other cluster-scoped resources (namespaces, clusterroles, etc.)
    │   └── {namespace}/       # Per-namespace resources and pod logs
    │       ├── events.yaml    # Events in this namespace (warnings, errors, scheduling decisions)
    │       ├── pods.yaml      # All Pod objects in this namespace
    │       ├── services.yaml  # All Services
    │       ├── deployments.apps.yaml  # Deployments, DaemonSets, etc.
    │       ├── configmaps.yaml
    │       └── {pod-name}/    # Container logs for each pod
    │           └── {container-name}.log        # Current container logs
    │           └── {container-name}.previous.log  # Previous (crashed) container logs
    └── {node-name}/           # Per-node directory (one per dumped VM)
        ├── journal.log        # Full systemd journal (journalctl --output=short-precise)
        ├── kern.log           # Kernel log (journalctl -k)
        ├── kubelet.log        # Kubelet service log
        ├── containerd.log     # Container runtime log
        ├── kube-proxy.log     # kube-proxy log
        ├── kops-configuration.log  # kops node configuration service (nodeup)
        ├── node-problem-detector.log
        ├── docker.log         # Docker service (if applicable)
        │
        │   # Control plane nodes only:
        ├── kube-apiserver.log
        ├── kube-controller-manager.log
        ├── kube-scheduler.log (from /var/log/)
        ├── etcd.log
        ├── etcd-events.log
        ├── etcd-cilium.log    # (if Cilium networking)
        ├── kops-controller.log
        │
        │   # Networking state:
        ├── iptables-nat.log   # iptables -t nat --list-rules
        ├── iptables-filter.log # iptables -t filter --list-rules
        ├── nftables-ruleset.log # nft list ruleset
        ├── ip-routes.log      # ip route show table all
        ├── ip-rules.log       # ip rule list
        ├── ip-link.log        # ip -s link
        ├── netstat.log        # ss -s (socket statistics)
        ├── bgp-protocols.log  # birdcl show protocols all (Calico only)
        ├── bird.cfg           # confd-generated BIRD config (Calico only)
        ├── bgp-sockets.log    # TCP sockets on port 179 (Calico only)
        ├── bgp-conntrack.log  # conntrack entries for port 179 (Calico only)
        │
        │   # System state:
        ├── etchosts           # /etc/hosts
        ├── sysctls            # sysctl -a
        ├── kubelet.conf       # /var/lib/kubelet/kubelet.conf
        ├── modules            # /proc/modules (loaded kernel modules)
        ├── cloud-init-output.log
        │
        │   # AWS-specific (if applicable):
        ├── aws-routed-eni_ipamd.log   # AWS VPC CNI IPAMD
        ├── aws-routed-eni_plugin.log  # AWS VPC CNI plugin
        │
        │   # Pod logs collected via kubectl:
        ├── external-dns.log   # external-dns pod logs
        └── dns-controller.log # dns-controller pod logs
```

### Node Dump Priority

Nodes are dumped in this priority order (from `pkg/dump/dumper.go`):
1. **Control plane nodes** — always dumped first
2. **Unregistered nodes** — cloud instances that didn't join the Kubernetes cluster (indicates bootstrap failure)
3. **Regular worker nodes** — up to the `--max-nodes` limit

## Failure Categories

Failures can be classified by examining JUnit XML results:

### `build` — Infrastructure failure
No JUnit test results found at all. The test infrastructure failed before any tests ran — the build, compilation, or cluster setup tooling broke before producing test output.

**How to identify:** No `junit_*.xml` files in the artifacts directory.

### `cluster_up` — Cluster creation failed
JUnit contains a failure for a test named "Up" or a test name containing "cluster". The kops cluster failed to create or validate.

**How to identify:** Parse JUnit XML, look for `<testcase>` elements with `<failure>` children where the test name is "Up" or contains "cluster".

### `e2e_test` — Individual test failures
The cluster came up successfully but specific e2e/conformance/feature tests failed. This is the most common failure mode for grid jobs.

**How to identify:** JUnit XML contains test failures for specific test cases that are not cluster creation.

## Diagnostic Process

### Step 1: Identify the failure type

1. Fetch `finished.json` to confirm the build failed
2. Fetch JUnit XML files (`artifacts/junit_*.xml`) and parse for `<failure>` elements
3. Classify the failure as `build`, `cluster_up`, or `e2e_test`

### Step 2: Read the build log

Start with `build-log.txt` — this is the overall test runner output and shows the high-level flow of what happened. Search for error messages, timeouts, and validation failures.

### Step 3: Check Kubernetes API resource dumps

The `cluster-info/` directory contains dumps of all Kubernetes API objects. These are essential for understanding cluster state at the time of failure:

- **`cluster-info/nodes.yaml`** — Check Node conditions (Ready, MemoryPressure, DiskPressure, PIDPressure, NetworkUnavailable). Nodes showing `NotReady` or `NetworkUnavailable` conditions point to kubelet or CNI issues. Also check node addresses and capacity.
- **`cluster-info/{namespace}/events.yaml`** — Check `cluster-info/kube-system/events.yaml` first for system-level issues, then other namespaces as needed. Search for Warning events that reveal scheduling failures, image pull errors, failed mounts, unhealthy pods, and node lifecycle issues. Look for patterns like `FailedScheduling`, `FailedCreatePodSandBox`, `Unhealthy`, `BackOff`, `NodeNotReady`.
- **`cluster-info/kube-system/pods.yaml`** — Check the status of system pods. Look for pods not in `Running` phase, containers in `CrashLoopBackOff` or `Waiting` state, and pods with restart counts > 0.

### Step 4: Check pod container logs

Pod container logs in `cluster-info/{namespace}/{pod-name}/{container-name}.log` are critical for diagnosing addon and system component failures:

- **CNI plugin pods** — For networking failures, check the logs of the CNI addon pods (e.g., cilium, calico, aws-node for VPC CNI, kube-router). These reveal why pod networking may not be functional. Look for errors during initialization, connectivity checks, or IPAM allocation.
- **CrashLooping pods** — Any pod with `CrashLoopBackOff` status (visible in `pods.yaml`) should have its container logs examined. Check both `{container}.log` (current) and `{container}.previous.log` (previous crash) for the root cause.
- **CoreDNS / kube-dns** — DNS failures cascade into many other errors. Check CoreDNS pod logs for upstream resolution failures or crash loops.
- **kube-proxy** — Networking issues at the Service level may appear in kube-proxy logs.
- **Cloud controller manager** — Node registration and load balancer issues may appear in cloud-controller-manager logs.

### Step 5: Check node-level logs

Use this priority order for investigating per-node logs in the `{node-name}/` directories:

1. **`journal.log`** — Full systemd journal with kubelet, networking, and service errors
2. **`kubelet.log`** — Kubelet-specific issues (pod scheduling, volume mounts, node registration)
3. **`kube-apiserver.log`**, **`kube-controller-manager.log`** — Control plane issues (control plane nodes only)
4. **Other service logs** — containerd.log, kube-proxy.log, kops-configuration.log

### Step 6: Check networking state

For networking-related failures, examine per-node networking dumps:
- `ip-routes.log` — Are expected routes present?
- `ip-link.log` — Are network interfaces up? Are CNI interfaces (e.g., cilium_*, cali*, eni*) present?
- `iptables-nat.log` / `iptables-filter.log` / `nftables-ruleset.log` — Are firewall rules correct?
- CNI-specific logs on nodes (e.g., `aws-routed-eni_ipamd.log` for AWS VPC CNI)

Cross-reference node networking state with CNI pod logs from `cluster-info/` to correlate node-level symptoms with CNI control plane errors.

### Step 7: Check cluster configuration

- `cluster.yaml` — Was the cluster configured correctly?
- `instancegroups.yaml` — Are instance groups as expected?
- `toolbox-dump.yaml` — Cloud resources dump (EC2 instances, security groups, load balancers, VPCs, subnets)

## Error Signal Patterns

When scanning log files, search for these error patterns to quickly identify root causes:

| Signal Type | Regex Pattern | What It Indicates |
|---|---|---|
| Network connectivity | `connection refused\|dial tcp.*timeout\|no route to host\|i/o timeout` | CNI not working, security groups misconfigured, DNS broken |
| Image pull | `ErrImagePull\|ImagePullBackOff` | Container images can't be pulled — registry issues, network issues, missing image |
| OOM | `OOMKilled\|cgroup memory` | Memory exhaustion — undersized nodes or memory leaks |
| Auth/cert | `x509:.*certificate\|certificate.*expired\|Unauthorized\|401 Unauthorized\|RBAC: access denied` | Certificate expiry, RBAC misconfiguration, webhook auth issues |
| Config error | `unknown flag\|invalid value` | Version skew, deprecated flags, bad configuration |
| Validation failure | `VALIDATION ERRORS\|Validation Failed\|is not ready` | Cluster didn't pass kops validation — nodes not joining, components unhealthy |
| CrashLoop | `CrashLoopBackOff\|Back-off restarting failed container` | Pods crashing repeatedly — config issues, missing dependencies |
| Node not ready | `node ".*" is not ready\|NotReady` | Nodes not joining cluster — kubelet issues, networking, cloud controller |

For each signal found, note the source file and surrounding context (2-3 lines before and after the match) to understand the root cause.

## Correlation Analysis

To identify systemic issues affecting multiple jobs, group failing jobs by their metadata dimensions and look for shared failure patterns.

### Process

1. **Collect metadata** for each failing job from its Prow config annotations (`test.kops.k8s.io/*`)
2. **Group jobs by each dimension** (cloud, distro, networking, k8s_version, kops_version)
3. **Within each group**, count how many jobs share the same failure category and error signal type
4. **Flag systemic patterns** when 3 or more jobs with the same dimension value share the same failure pattern, especially if they represent more than half the group

### Examples of Systemic Patterns

- **All jobs with a specific networking plugin failing with `network_connectivity` errors** → likely a bug or breaking change in that CNI plugin
- **All jobs with a specific distro failing with `config_error`** → likely a package version incompatibility on that OS
- **All jobs with a specific k8s version failing with `validation_failure`** → likely a regression in that Kubernetes release affecting kops
- **All jobs with a specific kops version failing** → likely a kops regression in that release branch

### Cross-referencing with Testgrid

Use Testgrid dashboards to see the historical failure trend:
- Check when failures started to identify the triggering change
- Compare with other jobs in the same dimension to confirm correlation
- Look at the git history around the failure start date for relevant changes

## Key Source Files

| File | Purpose |
|---|---|
| `tests/e2e/kubetest2-kops/deployer/dumplogs.go` | Orchestrates artifact collection via `kops toolbox dump` and `kops get` |
| `pkg/dump/dumper.go` | SSH-based per-node log dumper — defines which services, files, and commands are collected |
| `tests/e2e/kubetest2-kops/deployer/deployer.go` | Main kubetest2-kops deployer struct and configuration |
| `tests/e2e/kubetest2-kops/deployer/up.go` | Cluster creation logic |
| `tests/e2e/kubetest2-kops/deployer/down.go` | Cluster teardown (calls DumpClusterLogs before deletion) |
| `tests/e2e/pkg/tester/tester.go` | Test execution wrapper for Kubernetes e2e tests |
| `tests/e2e/scenarios/` | Individual test scenarios (each with run-test.sh and test.sh) |
