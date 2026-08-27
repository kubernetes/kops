# Supporting a new Kubernetes version

kOps does not carry an allowlist of Kubernetes versions. Any version parses, and the only bounds
are `OldestSupportedKubernetesVersion` and `OldestRecommendedKubernetesVersion` in
`upup/pkg/fi/cloudup/apply_cluster.go`, plus a guard that rejects a Kubernetes minor newer than the
kOps minor.

So "adding support" is not a single change. It is a handful of independent pull requests, each
landing when the upstream artifact it depends on actually exists — historically spread over several
months. Dropping the oldest minor, by contrast, is one coherent pull request.

Find the most recent example of whichever operation you are doing and mirror it:

```bash
# the last removal
git log --oneline --all -S 'OldestSupportedKubernetesVersion' -- upup/pkg/fi/cloudup/apply_cluster.go
# the last channels addition
git log --oneline --all --grep='channels: Add Kubernetes' -i
```

## Adding a new minor

The pieces below are ordered by when they can land, not by importance. Each is its own pull
request.

### 1. Integration test fixtures

These land first, often before the Kubernetes version is even released — fixtures routinely pin an
`-alpha` or `-rc` build.

Copy the previous minor's directories and hand-write only these:

- `tests/integration/update_cluster/minimal-1.NN/in-v1alpha2.yaml` — copy the previous minor's and
  change only `kubernetesVersion`.
- `tests/integration/update_cluster/minimal-1.NN/id_rsa.pub` — copy verbatim; the test will not run
  without it.
- `tests/integration/create_cluster/minimal-1.NN/options.yaml` — a few lines, naming the version.
- `cmd/kops/integration_test.go` — add a `TestMinimal_v1_NN`.
- `cmd/kops/create_cluster_integration_test.go` — add the matching case to `TestCreateClusterMinimal`.

**These last two are separate registrations for the same minor.** Editing one and not the other
leaves either a dangling fixture directory or a test pointing at a directory that does not exist.

Everything else under those directories — `kubernetes.tf`, all of `data/`, `expected-v1alpha2.yaml`
— is generated. Create it with `./hack/update-expected.sh`, which writes missing files as well as
updating existing ones.

The integration harness deliberately bypasses the too-new guard, so a fixture for a Kubernetes minor
newer than the current kOps minor still passes locally while real users would be told to upgrade
kOps.

### 2. Asset hashes

Once real releases exist, add a `generate_k8s_hashes 1.NN <max-patch>` line to
`hack/generate-asset-hashes.sh` and run it. That writes `pkg/assets/assetdata/k8s-1.NN.yaml`, which
is generated — never hand-edit it.

This file has to be extended again for every subsequent patch release, which is a routine
standalone PR of its own. See [Updating third-party dependencies](../dependency-updates.md).

### 3. Kubernetes library dependencies

Bump the `k8s.io/*` staging libraries to the matching `v0.NN.x` across every `go.mod` in the repo,
then:

```bash
make gomod && make apimachinery && make crds
```

`k8s/crds/` changes as a result, because the vendored apimachinery types move. A Go toolchain bump
often rides along in the same PR, but nothing requires it — kOps tracks Go on its own cadence.

kOps no longer vendors `kubernetes/kubernetes` itself, only the staging modules, so this step is far
less painful than it once was.

### 4. Per-minor component ladders

Several components are selected by Kubernetes minor. Each gets its rung when the corresponding
upstream component ships:

| Where | What |
|---|---|
| `pkg/model/components/clusterautoscaler.go` | a `switch` on the minor — the new minor becomes `default:` and the previous default gets an explicit `case` |
| `pkg/nodemodel/wellknownassets/cni.go` | `defaultCNIAsset{Amd64,Arm64}K8s_NN` constants **and a new case at the top of both arch ladders** |
| `hack/generate-asset-hashes.sh` | a matching `generate_cni_hashes` line, then regenerate |
| `pkg/model/components/etcd.go` | the `LatestEtcd3XVersion` constants and the mapping ladder, when upstream changes its recommended etcd minor |
| `pkg/model/components/etcdmanager/options.go` | feature-gate guards keyed on the minor |
| `pkg/model/components/gcpcloudcontrollermanager.go` | a `switch` on the minor that currently has only a `default:` — it reads like dead code but is the intended hook |

The cloud controller managers for AWS and Azure are **no longer** keyed on the minor; they carry a
single pinned image, so bump the string rather than adding a case.

OpenStack is different again: its CCM and CSI image tags are *computed* from the cluster's
Kubernetes minor in `upup/pkg/fi/cloudup/template_functions.go`. A new minor therefore produces new
image references with no code change and no check that the tag exists upstream.

### 5. Channels

Always a separate pull request, landing after the Kubernetes release is GA. It touches exactly two
files, `channels/alpha` and `channels/stable`, prepending one entry to `kubernetesVersions` and one
to `kopsVersions` in each.

`alpha` and `stable` intentionally differ: `stable`'s `kopsVersions` entry for a kOps release that
is not out yet points back at the previous stable kOps.

### 6. Release notes

Add a line to `docs/releases/1.NN-NOTES.md` under `## Kubernetes`.

## Dropping the oldest minor

This is one pull request, and a large one — recent examples touched 90 to 276 files.

1. **Raise the floor.** `OldestSupportedKubernetesVersion` and `OldestRecommendedKubernetesVersion`
   in `upup/pkg/fi/cloudup/apply_cluster.go` move together, preserving the two-minor gap between
   them.

2. **Delete the dead rungs.** Grep for the dropped version across `pkg/model/components/`,
   `pkg/nodemodel/`, `nodeup/pkg/model/`, and `upup/pkg/fi/cloudup/`:

   ```bash
   git grep -n 'IsKubernetesLT("1.NN\|IsGTE("1.NN\|IsLT("1.NN\|case NN:'
   ```

   Expect hits in the CNI ladders, the cluster-autoscaler switch, kubelet and apiserver option
   builders, feature-gate guards in nodeup, addon guards in the bootstrap channel builder, the
   default-image selection in `new_cluster.go`, and version guards inside addon templates. When
   removing a rung leaves a `switch` with only a `default:`, collapse it.

3. **Delete the fixtures**: the `tests/integration/{create,update}_cluster/minimal-1.NN/` trees, the
   corresponding `pkg/assets/assetdata/k8s-1.NN.yaml`, and both test registrations.

4. **Re-pin every test input still on the dropped version.** This is the bulk of the diff, and it is
   *inputs*, not golden output:

   ```bash
   git grep -l '1\.NN\.0' -- tests/ nodeup/ pkg/ upup/
   ```

   Two traps here. `pkg/assets/assetdata/data_test.go` pins a real sha256 alongside the version, so
   the hash has to be looked up rather than just the version string edited. And the `cluster.yaml`
   files under `nodeup/pkg/model/tests/` are hand-written inputs that nothing validates against the
   floor — they drift silently and are only ever caught during a removal.

5. **Docs**: create the next `docs/releases/1.NN-NOTES.md` skeleton with the deprecation notice, and
   add its `mkdocs.yml` nav entry.

Channels are **not** touched by a removal. `channels/alpha` and `channels/stable` still carry
entries for very old versions by design; do not prune them.

## Commands

```bash
./hack/update-expected.sh          # regenerates every golden file; run after any change above
make test
make ci
```

`./hack/update-expected.sh` takes an optional package and `-run` filter, which is much faster while
iterating:

```bash
./hack/update-expected.sh ./cmd/kops/... TestMinimal_v1_NN
```

## Things that are easy to miss

- **`hack/generate-asset-hashes.sh` is never pruned.** It still lists Kubernetes minors whose
  `assetdata` YAML has been deleted, so running it resurrects those files. Delete the matching line
  when you delete the YAML, or discard the strays afterwards.
- **The CNI ladder is two parallel switches**, one per architecture. Editing one and not the other
  produces no compile error and no test failure on the architecture you did not test.
- **`tests/integration/update_cluster/complex/` has an `in-legacy-v1alpha2.yaml`** alongside its
  `in-v1alpha2.yaml`. It is the only fixture that does, and it is easy to bump one and miss the
  other.
- **Prow jobs live in `kubernetes/test-infra`**, not here. Adding or dropping a minor needs a
  matching change to `config/jobs/kubernetes/kops/`.

## Related

- [Updating third-party dependencies](../dependency-updates.md) — the pinning surfaces referenced
  above, and what regenerates each.
- [Updating the default base AMI](update_ami_versions.md) — the channel images list. A new
  Kubernetes minor only touches it when you are introducing a new distro boundary.
