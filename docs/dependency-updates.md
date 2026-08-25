# Updating Third-Party Dependencies

This guide describes how to update the third-party dependencies that kOps pins, and what else must
change alongside each one. It is written for an automated agent: each section names the file that
holds the version, the files coupled to it, the regeneration command, and the verification command.

Deliberately, this document names **files, not line numbers, and describes versions rather than
quoting them** — pins move constantly. Grep for the symbol or the surrounding construct.

Two mechanisms cut across almost every surface below. Read these first.

## The two cross-cutting mechanisms

### 1. `pkg/assets/assetdata/` — the embedded hash store

Every binary kOps downloads onto a node has its SHA256 pre-stored in
`pkg/assets/assetdata/*.yaml`, embedded via `//go:embed` in `pkg/assets/assetdata/data.go` and
looked up by canonical URL.

**A version constant bumped without a matching hash entry compiles fine and fails at
`kops update cluster` time.** Never hand-edit these YAML files; regenerate them:

```bash
# 1. Edit the invocation block at the bottom of hack/generate-asset-hashes.sh.
#    For k8s, runc, and containerd the second argument is the MAXIMUM PATCH version:
#        generate_k8s_hashes <minor> <max-patch>
#        generate_containerd_hashes <minor> <max-patch>
#    For CNI the arguments are a base path plus an explicit version, because the upstream
#    patch numbering has gaps — add a new line rather than raising a counter:
#        generate_cni_hashes <minor> <full-version>
# 2. Regenerate:
./hack/generate-asset-hashes.sh
# 3. Verify:
make verify-hashes
```

Generated files contain harmless junk rows whose `sha256` field is the literal string `Hash:`.
Leave them.

### 2. `./hack/update-expected.sh` — golden-file regeneration

Any change to a default version that reaches the cluster spec or nodeup config ripples into golden
files under `tests/integration/` — routinely dozens of files, sometimes several hundred.

```bash
./hack/update-expected.sh                                          # everything (slow, safe default)
./hack/update-expected.sh ./cmd/kops/...                           # one package
./hack/update-expected.sh ./cmd/kops/... TestMinimalAWS            # one test
./hack/update-expected.sh ./upup/pkg/fi/cloudup/... TestBootstrapChannelBuilder
```

The script sets `HACK_UPDATE_EXPECTED_IN_PLACE=1` (read by `pkg/testutils/golden/compare.go`) and
unsets `KOPS_BASE_URL`, `DNSCONTROLLER_IMAGE`, `KOPS_FEATURE_FLAGS`, cloud credentials, and forces
`KUBECONFIG=/dev/null`. **Never run the underlying `go test` with that env var directly** — without
the unsets, dev environment values get baked into the goldens.

Two gotchas:

- **The script never deletes stale goldens.** A manifest rename leaves an orphan file, and
  `cmd/kops/integration_test.go` compares directory listings, so the test fails until you
  `git rm` it by hand. There is a comment in that test saying so.
- **A version bump also changes `cluster-completed.spec_content` and `kubernetes.tf`** (the S3
  object hashes), not just the rendered addon manifest. A regen diff that touches only the manifest
  file means the regen did not actually run.

By repo convention the regeneration lands as its own commit inside the same PR, titled literally
`./hack/update-expected.sh`, so reviewers can read the real diff separately from the churn.

## Surface index

| # | Dependency surface | Version lives in | Section |
|---|---|---|---|
| 1 | Go modules | `go.mod` and the nested modules | [Go modules](#go-modules) |
| 2 | Go toolchain | every `go.mod` + `cloudbuild.yaml` | [Go toolchain](#go-toolchain) |
| 3 | Build/lint tooling | `hack/go.mod`, `Makefile`, `hack/verify-*.sh` | [Build and lint tooling](#build-and-lint-tooling) |
| 4 | GitHub Actions | `.github/workflows/*.yml` (SHA-pinned) | [GitHub Actions](#github-actions) |
| 5 | Addon manifests | `upup/models/cloudup/resources/addons/**` | [Addon manifests](#addon-manifests) |
| 6 | Addon default versions | `pkg/model/components/*.go` | [Addon manifests](#addon-manifests) |
| 7 | containerd / runc | `pkg/model/components/containerd.go` | [Node components](#node-components) |
| 8 | CNI plugins | `pkg/nodemodel/wellknownassets/cni.go` | [Node components](#node-components) |
| 9 | crictl / nerdctl | `pkg/nodemodel/wellknownassets/{crictl,nerdctl}.go` | [Node components](#node-components) |
| 10 | kubelet/kubectl hashes | `pkg/assets/assetdata/k8s-*.yaml` | [Node components](#node-components) |
| 11 | Credential providers | `pkg/nodemodel/fileassets.go` | [Node components](#node-components) |
| 12 | etcd / etcd-manager | `pkg/model/components/etcd.go`, `etcdmanager/model.go` | [etcd](#etcd-and-etcd-manager) |
| 13 | Kubernetes version support | `upup/pkg/fi/cloudup/apply_cluster.go`, `channels/*` | [Kubernetes versions](#kubernetes-version-support) |
| 14 | OS images / AMIs | `channels/{alpha,stable}` | [OS images](#os-images-and-amis) |
| 15 | e2e scenario charts | `tests/e2e/scenarios/**` | [e2e scenarios](#e2e-scenario-dependencies) |
| 16 | Cluster API manifests | `clusterapi/manifests/**` | [Miscellaneous](#miscellaneous-surfaces) |
| 17 | Docs tooling | `images/mkdocs/*`, `netlify.toml` | [Miscellaneous](#miscellaneous-surfaces) |

**Only surface 4 is automated.** `.github/dependabot.yml` configures the `github-actions` ecosystem
and nothing else — there is no `gomod`, `docker`, or `pip` entry. Every other bump in this document
is a manual pull request.

## Go modules

The repo has several modules and a single vendor tree — **only the root module is vendored**. There
is no `go.work`. Enumerate them before starting:

```bash
find . -name go.mod -not -path './vendor/*'
```

Beyond the root module, expect a `hack/` tools module, a `tests/e2e/` module, one or more modules
under `tests/e2e/scenarios/`, and small modules under `tools/`. Points to know:

- The root module holds the `vendor/` tree and any `replace` directives.
- `hack/go.mod` uses a Go `tool` directive block to pin the lint/codegen tooling.
- `tests/e2e/go.mod` carries a `replace` pointing at the repo root, so it tracks local changes.
- Some of the small `tools/` modules have **no `replace` back to the root**, so they resolve the
  published `k8s.io/kops` from the module proxy and do not see local changes.

### Recipe

```bash
# 1. Edit the version in the owning go.mod (add a `replace` only for a genuine pin).
# 2. Regenerate everything:
make gomod
# 3. Verify:
make verify-gomod
```

`make gomod` runs `go mod tidy` + `go mod vendor` at the root, then `go mod tidy` in every other
module directory. Never hand-edit `go.sum` or `vendor/`.

`make gomod` does **not** unify a version across modules. If the dependency you bumped also appears
in another module's `go.mod`, edit each one.

### If the dependency is `k8s.io/*` or `sigs.k8s.io/controller-tools`

```bash
make apimachinery     # NOT apimachinery-codegen — the generators don't emit goimports-clean output
make crds             # regenerates k8s/crds/*.yaml
make verify-apimachinery verify-crds
```

The Kubernetes staging libraries (`k8s.io/api`, `apimachinery`, `cli-runtime`, `client-go`,
`component-base`, `kubectl`, `kubelet`, `mount-utils`) are pinned as a **coordinated set at the
same patch version** in `go.mod`. Bump them together. Indirect stragglers such as
`k8s.io/apiextensions-apiserver` and `k8s.io/kube-openapi` are not part of that set and routinely
lag.

**`CODEGEN_VERSION` in the `Makefile` pins `k8s.io/code-generator` and must be bumped in the same
change.** Because it is fetched over the network by `go run k8s.io/code-generator/cmd/<gen>@$(CODEGEN_VERSION)`,
`make gomod` cannot see it and nothing will fail if you forget — it simply drifts out of step with
the staging libraries. Keep it at the same version as `k8s.io/api`.

Note that `hack/verify-apimachinery.sh` runs `git status --porcelain --untracked-files=no`
repo-wide and unfiltered, so **any unrelated uncommitted edit makes it fail locally**.

### Conventions

- Commit `go.mod`, `go.sum`, and the `vendor/` diff **together in one commit**. Never split them.
- Keep the dependency PR separate from feature work.

## Go toolchain

There is no `.go-version` file and no `toolchain` directive. The version is a **patch-level `go`
directive repeated in every `go.mod` in the repo**, plus the CI builder image. The root `go.mod`
carries a comment reminding you to keep it in sync with `cloudbuild.yaml` and the other modules.

Files that must change together:

| File | What to change |
|---|---|
| every `go.mod` found by `find . -name go.mod -not -path './vendor/*'` | the `go` directive |
| root `go.mod` | the `godebug default=` directive — **only on a minor bump** |
| `cloudbuild.yaml` | the `golang:` builder image, which appears in **several** build steps |

Then `make gomod`.

**Nothing in `.github/workflows/` needs editing** — every `actions/setup-go` step uses
`go-version-file` pointing at `go.mod`.

`.golangci.yaml` has its own `go:` setting under `run:`. It is a lint language-level version and
does not have to match the toolchain, but it **does** have to be compatible with the pinned
golangci-lint. Check it when bumping either the toolchain or golangci-lint.

On a Go **minor** bump, also re-sync `third_party/forked/text/template`, which is a snapshot of the
stdlib package taken at a specific Go version (see its `README.md`).

**Do not combine a Go toolchain bump with a general dependency refresh.** The toolchain bump must
stay independently cherry-pickable to every open release branch — see
[Known failure modes](#known-failure-modes).

## Build and lint tooling

Most linters are declared in the `hack/` tools module and carry **no version string in the shell
script** — bump them with `cd hack && go get -u <module>` followed by `make gomod`.

| Tool | Pinned in | Invoked by |
|---|---|---|
| golangci-lint | `hack/go.mod` | `hack/verify-golangci-lint.sh` |
| controller-gen | `hack/go.mod` | `make crds` |
| goimports | `hack/go.mod` | `hack/{verify,update}-goimports.sh` |
| misspell | `hack/go.mod` | `hack/verify-spelling.sh` |
| staticcheck | `hack/go.mod` | not called by any script directly |

**A controller-tools bump rewrites `k8s/crds/*.yaml`** — run `make crds` and commit the result. A
golangci-lint bump may also require adjusting the `go:` setting in `.golangci.yaml`.

Versions pinned as literals instead, each needing a hand edit:

| File | Pin |
|---|---|
| `Makefile` | `CODEGEN_VERSION` (see [Go modules](#go-modules)) and the `ko` version in the `KO` variable |
| `hack/verify-shellcheck.sh` | `SHELLCHECK_VERSION` **and** `SHELLCHECK_IMAGE`, which carries both a tag and a digest — a comment in the file says to keep them in sync, and all three values change together |
| `hack/verify-terraform.sh` | `TF_TAG` |
| `hack/{verify,update}-gofumpt.sh` | a `go install mvdan.cc/gofumpt@<version>` line in each |
| `clusterapi/gen.go` | a `go run sigs.k8s.io/controller-tools/cmd/controller-gen@<version>` line |
| `images/mkdocs/requirements.txt` | the mkdocs Python packages — see [Miscellaneous](#miscellaneous-surfaces) |

Some tool invocations use `@latest` and are therefore not reproducible. Treat these as known debt:
do not "bump" them, and do not convert them to pins as a side effect of an unrelated change.

## GitHub Actions

Every third-party action is pinned to a **bare 40-character commit SHA with no version comment**.
You cannot read the version from the workflow file. To recover it:

```bash
grep -rhoE 'uses: [^ ]+' .github/workflows/ | sort -u   # current pins
git log -p --follow .github/workflows/main.yml          # the bump commits name the versions
```

`.github/dependabot.yml` registers the `github-actions` ecosystem for the repo root, on a weekly
schedule, applying the `ok-to-test` label. Dependabot commits use
`build(deps): bump <owner>/<action> from <X> to <Y>`. A given action's SHA is identical across every
workflow file, so all occurrences move in one commit.

**Verify dependabot is actually keeping up rather than assuming it.** It can be disabled at the repo
or org level without any change to `dependabot.yml`, in which case bumping these actions becomes the
agent's job. Check both halves — that PRs are still being opened, and that the pins actually match
upstream:

```bash
# Is dependabot still opening PRs?
gh pr list --repo kubernetes/kops --author "app/dependabot" --state all --limit 10 \
  --json number,title,createdAt

# Are the current pins the latest upstream releases?
for r in actions/checkout actions/setup-go actions/upload-artifact actions/dependency-review-action; do
  echo -n "$r "; gh release view --repo "$r" --json tagName -q .tagName
done
```

If a pin lags the latest upstream release and no dependabot PR is open for it, bump it manually:
resolve the release tag to its commit SHA (`gh api repos/<owner>/<repo>/commits/<tag> -q .sha`) and
replace every occurrence.

Note also that a long gap between dependabot PRs is normal when no new action releases have shipped
— compare against upstream before concluding anything is broken.

**Runner images are NOT covered by dependabot** and need a manual PR. Grep for `runs-on:` across
`.github/workflows/` to find them.

`.github/workflows/depsreview.yaml` runs `dependency-review-action`, but only on a path filter for
the root `go.mod` — the **nested** modules' go.mod files do not trigger it.

## Addon manifests

Addon manifests live in `upup/models/cloudup/resources/addons/<addon-key>/` and are embedded at
compile time via `//go:embed` in `upup/models/vfs.go`. **There is no codegen step for manifests.**

File naming is `k8s-<minMinor>.yaml[.template]`, where the number is the *minimum Kubernetes version
the manifest supports* — not the addon version. `.template` files are rendered against the cluster
spec; plain `.yaml` files are used verbatim.

The manifest filename is hardcoded per addon in
`upup/pkg/fi/cloudup/bootstrapchannelbuilder/bootstrapchannelbuilder.go` (Cilium has its own file in
that package), so **renaming a manifest requires a matching Go edit**.

`channels/alpha` and `channels/stable` are *not* addon bookkeeping — they hold OS images and
Kubernetes version recommendations only, and an addon bump never touches them.

### Three pinning patterns

| Pattern | Where the tag lives | Examples |
|---|---|---|
| **A. Go default fed into the template** | `pkg/model/components/<addon>.go` | cilium, aws-ebs-csi, node-termination-handler, karpenter, kindnet, cluster-autoscaler, node-problem-detector, nodelocaldns |
| **B. Template literal with an `or` fallback** | the `.yaml.template` itself | calico, aws-load-balancer-controller, coredns, metrics-server |
| **C. Fully hardcoded in the template** | the `.yaml.template` itself | cert-manager, snapshot-controller, external-dns, flannel, kube-router, amazon-vpc-cni, and the CSI **sidecar** images of aws-ebs-csi |

Templates using `{{ .Version }}` open with a `{{ with .Path.To.Spec }}` block — that is why the bare
field resolves.

### Finding the current version and the upstream source

**Generated-from-Helm addons** carry a machine-readable pointer in `kustomization.yaml`
(`helmCharts[].repo`, `.name`, `.version`) and most have a `regenerate.sh` beside them:

```bash
ls upup/models/cloudup/resources/addons/*/regenerate.sh
ls upup/models/cloudup/resources/addons/*/kustomization.yaml
cd upup/models/cloudup/resources/addons/<addon> && ./regenerate.sh   # needs kustomize with --enable-helm
```

Some addons are helm-templated without a `regenerate.sh`; the `helm template` command is in the
header comment of their `helm-values.yaml`. A couple of `regenerate.sh` scripts post-process their
output — read the script before assuming it is a plain `kustomize build`.

**Vendored addons** carry the upstream URL in a comment on the first few lines of the template,
usually with the release tag embedded in the URL. **Bump the tag in that comment when you bump the
addon** — it is the only record of the source revision.

> **These header comments are frequently stale**, sometimes by many minor versions. Read the
> `image:` line for the current version, not the header — then fix the header.

Some addons have **no upstream pointer at all**; infer the source from the image registry and tag.

### Recipe

1. Find the new upstream version (above).
2. Bump the version in its home:
   - Pattern A → the literal in `pkg/model/components/<addon>.go`.
   - Helm/kustomize → `kustomization.yaml` `helmCharts[].version` **and** `helm-values.yaml`
     `image.tag`. These are independent fields; both must move.
   - Pattern B/C → every `image:` line **and** every `app.kubernetes.io/version:` label.
3. Refresh the manifest body: `./regenerate.sh`, or the `helm template` command from
   `helm-values.yaml`, or re-download the upstream YAML and re-apply the kOps deltas by hand.
4. Update the upstream-source URL comment at the top of the template.
5. `./hack/update-expected.sh`, and `git rm` any orphaned golden file.
6. `make test` (or at minimum `go test ./cmd/kops/... ./upup/pkg/fi/cloudup/...`).

To find which integration cases an addon touches:

```bash
ls tests/integration/update_cluster/*/data/*addons-<addon-key>-*
```

High-coverage cases include `many-addons`, `many-addons-ccm*`, `minimal-aws`, `privatecilium*`,
`minimal-warmpool`, `aws-lb-controller`, and `karpenter`.

### Addon gotchas

- **Cilium minor bumps are gated by validation.** `pkg/apis/kops/validation/validation.go` hardcodes
  the single supported Cilium minor and rejects everything else. Bumping the minor requires editing
  that check and its test in `validation_test.go`, or every cluster is rejected. Grep the file for
  `Only version` to find it.
- **Two-place bumps:** karpenter (Go image literal *and* chart version in `kustomization.yaml`),
  aws-load-balancer-controller (chart version in two `kustomization.yaml` files, tag in two
  `helm-values.yaml` files), aws-ebs-csi (Go default for the driver, plus the sidecar image literals
  in the template that the chart does not template).
- **cluster-autoscaler is keyed on the Kubernetes minor**, not a single default —
  `pkg/model/components/clusterautoscaler.go` is a `switch` over `v.Minor`. Adding a Kubernetes
  minor means adding a `case`.
- **Calico has no Go default at all.** The version literal exists only in the template's `or`
  fallbacks, and there is no version-range validation.
- If the new version adds or removes a resource Kind, check `alwaysPruneGroupKinds` in
  `upup/pkg/fi/cloudup/bootstrapchannelbuilder/pruning.go`.
- If the upgrade needs node churn or certificates, check the `NeedsRollingUpdate` / `NeedsPKI` flags
  on the `AddonSpec`.
- If the bump adds a new tunable, it becomes an API change: edit `pkg/apis/kops/networking.go`,
  `v1alpha2/`, `v1alpha3/`, then `make apimachinery && make crds`.
- **Controllers that talk to a cloud API often need a matching IAM policy change.** Compare the
  upstream `iam_policy.json` against `pkg/model/iam/iam_builder.go` when bumping the AWS load
  balancer controller, a CSI driver, or Karpenter — see
  [Known failure modes](#known-failure-modes).

### kOps-owned addons — never bump these by hand

`dns-controller.addons.k8s.io` and `kops-controller.addons.k8s.io` use `{{ KopsVersionImageTag }}`,
driven by `kops-version.go`. They move only via `hack/set-version`, enforced by
`make verify-versions`.

Also kOps-authored, with no upstream to track: `kubelet-api.rbac.addons.k8s.io`,
`limit-range.addons.k8s.io`, `storage-{aws,gce,azure,openstack}.addons.k8s.io`,
`hcloud-config.addons.k8s.io`, `azure-cloud-config.addons.k8s.io`, and
`karpenter.sh/instancegroups.yaml.template`.

## Node components

### containerd and runc

Both defaults are set in the **same `if` block** in `pkg/model/components/containerd.go`, so they
bump together. `DefaultSandboxImage` (the `pause` image) is a constant in the same file.

Coupled files: `pkg/nodemodel/wellknownassets/{containerd,runc}.go` (URL templates and a
`semver.LT` floor), `pkg/apis/kops/validation/validation.go` (**the containerd floor is duplicated
there** — raising it means editing both places plus `validation_test.go`), and the hash files under
`pkg/assets/assetdata/`.

Read the comment above the containerd default before changing it. It records upstream regressions
that specific patch releases must be avoided for, and that reasoning is not recorded anywhere else.

### CNI plugins — pinned per Kubernetes minor

`pkg/nodemodel/wellknownassets/cni.go` defines `defaultCNIAsset{Amd64,Arm64}K8s_NN` constants,
selected by **two parallel descending `IsGTE` ladders** — one for amd64, one for arm64. It is easy
to update one and miss the other. There is a unit test alongside it that also encodes the mapping.

Policy: match the version upstream Kubernetes pins for that minor.

### crictl and nerdctl

Installed on demand only, gated by `spec.containerd.installCriCtl` / `installNerdCtl`.

- crictl (`pkg/nodemodel/wellknownassets/crictl.go`) — the version is **inline in the URL**, there is
  no version variable. Hashes live in `pkg/assets/assetdata/crictl.yaml`.
- nerdctl (`pkg/nodemodel/wellknownassets/nerdctl.go`) — URLs **plus inline
  `nerdctlAssetHash{Amd64,Arm64}` constants**. nerdctl is the one component whose hashes live in Go
  rather than in `assetdata/`.

### kubelet / kubectl

The download base URL is derived from the cluster's Kubernetes version in
`pkg/nodemodel/fileassets.go`, so there is no version to bump — but
**`pkg/assets/assetdata/k8s-*.yaml` must be extended for every new Kubernetes patch release**, or
kOps refuses that version. That is a routine, standalone PR: raise the max-patch argument in
`hack/generate-asset-hashes.sh`, run the script, done. No golden-file churn.

### Cloud credential providers

`pkg/nodemodel/fileassets.go` contains full download URLs with the version inline and no variable,
one for the AWS ECR credential provider and one for auth-provider-gcp. Each is paired with a hash
file (`pkg/assets/assetdata/ecr.yaml`, `pkg/assets/assetdata/gcp.yaml`).

The GCP binary must stay in step with the image tag in
`pkg/model/components/gcpcloudcontrollermanager.go` — same upstream release, two independent
literals.

### GPU and sandboxing

- NVIDIA: the driver package and device-plugin image are constants in
  `pkg/apis/kops/containerdconfig.go`.
- gVisor: `nodeup/pkg/model/gvisor.go` uses a rolling apt repository — **nothing to pin**.

Docker support has been removed; there is no docker version anywhere, despite `"docker"` still
appearing as a valid `containerRuntime` string in validation.

### Recipe

```
1. pkg/model/components/containerd.go  (or the relevant wellknownassets/<x>.go)
2. hack/generate-asset-hashes.sh       (raise the max-patch argument / add the version)
3. ./hack/generate-asset-hashes.sh     -> pkg/assets/assetdata/<component>-<minor>.yaml
4. ./hack/update-expected.sh           -> tests/integration/**            [SEPARATE COMMIT]
5. docs/releases/<current>-NOTES.md    (one line)
6. If raising a floor: validation.go AND wellknownassets/<x>.go (duplicated guard) + tests
```

## etcd and etcd-manager

| Pin | Location |
|---|---|
| Latest etcd per minor | `LatestEtcd<NN>Version` constants in `pkg/model/components/etcd.go` |
| Kubernetes → etcd minor mapping | the `switch` below those constants in the same file |
| Image table | `etcdLatestImages` in `pkg/model/components/etcdmanager/options.go`, derived from the constants |
| etcd-manager image tag | the pod spec in `pkg/model/components/etcdmanager/model.go` |

Older patch versions are **generated, not listed**: `etcdSupportedVersions()` in
`etcdmanager/options.go` synthesizes entries for every patch below the configured latest, so bumping
a `LatestEtcd<NN>Version` constant automatically widens the accepted set.

The etcd-manager tag also appears in checked-in expected output under
`pkg/model/components/etcdmanager/tests/` and `tests/integration/update_cluster/*/assets.yaml` —
`./hack/update-expected.sh` handles these. Its blast radius is among the largest of any bump.

## Kubernetes version support

| What | Where |
|---|---|
| Oldest supported / oldest recommended | `OldestSupportedKubernetesVersion` and `OldestRecommendedKubernetesVersion` in `upup/pkg/fi/cloudup/apply_cluster.go` |
| Default version for new clusters | `channels/{alpha,stable}`, `spec.kopsVersions[].kubernetesVersion` |
| Upgrade nag ranges | `channels/{alpha,stable}`, `spec.kubernetesVersions[]` |
| Fallback when no channel | `upup/pkg/fi/cloudup/defaults.go` |

`pkg/apis/kops/validation/` contains **no** min/max Kubernetes version — the bounds live only in
`apply_cluster.go`.

**Adding support for a new minor** is a two-line-per-file channel change: one
`spec.kubernetesVersions` entry and one `spec.kopsVersions` entry at the top of each list in both
`channels/alpha` and `channels/stable`.

**Dropping the oldest minor** is a large, mechanical sweep: the `apply_cluster.go` constants,
deleting that minor's `pkg/assets/assetdata/k8s-*.yaml`, removing the per-minor branches in
`pkg/model/components/` and both ladder arms in `pkg/nodemodel/wellknownassets/cni.go`, deleting the
`tests/integration/{create,update}_cluster/minimal-<minor>/` trees, adding a release-notes skeleton
under `docs/releases/`, and a `mkdocs.yml` nav line. Find the most recent such commit and mirror it:

```bash
git log --oneline --all -- upup/pkg/fi/cloudup/apply_cluster.go | head
```

## OS images and AMIs

Default images for AWS, GCE, and Azure live in `channels/{alpha,stable}` under `spec.images[]`,
resolved by `FindImage` in `pkg/apis/kops/channel.go`.

The cadence is strict and documented in `docs/contributing/update_ami_versions.md`: bump `alpha`,
let it bake 7-10 days, then promote to `stable`. Commit messages alternate between
`chore(channels): bump k8s and ubuntu ami versions in alpha channel` and
`chore(channels): promote alpha to stable`.

Two things that make this less trivial than it looks:

- Channel bumps stay single-file only when they touch image rows that no test fixture embeds. Some
  `tests/integration/create_cluster/*/expected-v1alpha2.yaml` fixtures embed the AMI for the current
  default distro, so promoting that row to `stable` also needs `./hack/update-expected.sh`. Note
  this touches **`create_cluster`** goldens, not `update_cluster`.
- A new image *family* additionally needs an entry in `HasUpstreamImagePrefix` in
  `pkg/apis/kops/channel.go`, which is the allowlist of prefixes `kops upgrade cluster` will
  rewrite.

Providers absent from the channels use hardcoded `default<Provider>Image*` constants in
`upup/pkg/fi/cloudup/populate_instancegroup_spec.go`.

Distro lifecycle changes (experimental → stable → deprecated) are documented in the support matrix
in `docs/operations/images.md` and the current release notes.

## e2e scenario dependencies

`tests/e2e/scenarios/**` pins upstream charts, manifests, and CLIs. Nothing here is automated.

Pins are spread across `run-test.sh` / `test.sh` scripts, image tags in `**/testdata/*.yaml`, and
occasionally in Go test files. Some scenarios pin a dozen upstream components in a single script.

Three things worth knowing before editing:

- **The same component is often pinned in several scenarios at different versions.** Grep the whole
  scenarios tree for the component name before claiming a bump is complete.
- **Some scenarios derive the version from the running cluster and must not be "bumped".** Several
  read the image tag off the deployed workload and then `git checkout` that tag upstream. There is
  no pin to edit.
- **Some references deliberately float**: the Kubernetes stable-release markers, the kOps CI
  markers, and a few manifests fetched from an upstream `main` branch. Leave them.

Image tags are sometimes duplicated between testdata YAML and Go test files. Grep, don't assume.

After editing any scenario shell script: `make verify-shellcheck`. After editing `tests/e2e` Go code
or `tests/e2e/go.mod`: `make test-e2e-install`.

## Miscellaneous surfaces

**Cluster API manifests** (`clusterapi/manifests/`): each provider has a `kustomization.yaml` with a
`?ref=<tag>` and a `patches/set_manager_image.yaml` with an image tag. These are two independent
literals per provider; bump both and check they agree.

**Docs tooling**: `images/mkdocs/Dockerfile`, `images/mkdocs/requirements.txt`, and `netlify.toml`
move together.

**Unmaintained Dockerfiles** under `hooks/` pin EOL base images and their scripts are excluded via
`hack/.shellcheck_failures`. Do not bump these speculatively — touching them is a separate cleanup
decision.

## Verification

Run the narrowest gate that covers what you changed, then the broad one.

| Changed | Command |
|---|---|
| any `go.mod` / `go.sum` / `vendor/` | `make gomod && make verify-gomod` |
| `k8s.io/*` or controller-tools | `make apimachinery && make crds && make verify-apimachinery verify-crds` |
| `hack/go.mod` tool versions | `make gomod`, `make verify-golangci-lint`, `make verify-misspelling`, `hack/verify-crds.sh` |
| `pkg/assets/assetdata/*.yaml` | `make verify-hashes` |
| addon manifests / component defaults | `./hack/update-expected.sh` then `make test` |
| shell scripts | `make verify-shellcheck` |
| `tests/e2e` Go code | `make test-e2e-install` |
| `hack/verify-terraform.sh` (`TF_TAG`) | `hack/verify-terraform.sh` (needs Docker) |
| `hack/verify-shellcheck.sh` version/image | `hack/verify-shellcheck.sh` (needs Docker) |
| broad gate | `make quick-ci` (what the GitHub Actions `verify` job runs) |
| broadest | `make ci` (adds `verify-gomod`, `verify-golangci-lint`, `verify-terraform`, `nodeup`, `examples`, `test`) |

**The two CI systems are complementary, not overlapping**, so neither one alone proves a dependency
change is clean:

| Runs only in GitHub Actions (`make quick-ci`) | Runs only in Prow presubmits |
|---|---|
| `verify-crds`, `verify-goimports`, `verify-versions`, `verify-misspelling`, `verify-shellcheck`, `verify-gendocs`, `verify-apimachinery`, `verify-codegen` | `verify-gomod`, `verify-golangci-lint`, `verify-hashes`, `verify-terraform`, `verify-gofmt` |

Each Prow check is its own job invoking a single make target (`pull-kops-verify-gomod` runs
`make verify-gomod`, and so on); `pull-kops-verify-generated` runs `make verify-generate`, which
resolves to just `verify-crds`. `make ci` is the closest single local approximation to the union.

One gap: `hack/verify-gofumpt.sh` is wired into no `make` target and is run by no job in either
system. If you care about it, run it by hand.

## Commit and pull request conventions

### Commits

- **One dependency concern per PR.** Dependency bumps are kept separate from feature work.
- **kOps merges with merge commits, not squash.** Your individual commit messages land verbatim on
  `master`, so they matter.
- **Split the bump from the regeneration.** The established shape is two commits: the hand-written
  change, then the generated churn in a commit named after the command that produced it. All three
  spellings appear in recent history: `./hack/update-expected.sh`, `hack/update-expected.sh`,
  `test: hack/update-expected`. Other regeneration commits are titled `make gomod` or
  `./hack/generate-asset-hashes.sh`.
- **Titles are free-form.** Despite `AGENTS.md` citing conventional commits, most dependency PR
  titles carry no prefix at all. Three styles are all normal:

  | Style | Shape |
  |---|---|
  | Bare imperative (most common) | `Update <Thing> to <version>`, `Upgrade <Thing> to <version>`, `Update dependencies` |
  | `<component>: <imperative>` | `aws: Update EBS CSI driver to <version>`, `etcd-manager: upgrade to <version>`, `deps: bump containerd to <version> to address vulnerabilities` |
  | Conventional commits | `chore(networking): bump <thing> to <version>`, `chore(channels): ...`, `chore: upgrade <thing> to <version>` |

  Match the surrounding history for the file you are editing rather than imposing a format.

### Pull request body

`.github/PULL_REQUEST_TEMPLATE/pull_request_template.md` has **no `release-note` block** and no
`/kind` section — only "What this PR does / why we need it", "Which issue(s) this PR fixes", and
"Special notes for your reviewer". A `release-note` block appears **only** on cherry-pick PRs,
injected by `cherry_pick_pull.sh`, and is left empty for dependency picks. Release notes are
hand-written later in `docs/releases/`.

Many maintainer bump PRs have a nearly empty body, but the ones that get reviewed fastest state
**the upstream release link and the concrete reason** — the upstream changelog entry, the API that
changed, and the user-visible symptom it fixes. A vague "bump for CVEs" PR with no named advisory
gets no traction; name the CVE or the upstream issue.

### Labels, review, and CI

**Do not hand-apply labels.** Every label on these PRs is bot-applied: `cncf-cla`, `size/*`,
`area/*` by path, and `lgtm`/`approved` by prow. Dependabot gets `ok-to-test` from
`.github/dependabot.yml`; a first-time human contributor needs a Kubernetes org member to type
`/ok-to-test` before presubmits run at all.

Approvers are listed in the root `OWNERS`. Roughly 45 sub-directory `OWNERS` files exist, so a
provider-scoped bump can be approved by that directory's approvers.

Presubmits that most often fail on a dependency change:

| Job | Why it fails | Fix |
|---|---|---|
| `pull-kops-test` | Golden fixtures under `tests/integration/` and `upup/pkg/fi/cloudup/tests/bootstrapchannelbuilder/` embed the image tag | `./hack/update-expected.sh` |
| `pull-kops-verify-gomod` | `go.mod` was edited without regenerating `vendor/` and tidying the nested modules | `make gomod` — a bare `go get` is never enough |
| `pull-kops-verify-generated` | codegen output is stale | `make apimachinery` / `make crds` |
| e2e jobs | usually flakes; maintainers `/override` rather than block | `/retest`, or `/test <job>` for path-triggered optional jobs |

Some e2e jobs are path-triggered and optional (the load balancer controller, Karpenter, kube-router,
and several Azure scenarios among them); a bump touching those components should request them
explicitly with `/test <job>`.

Pushing new commits after an `/lgtm` automatically removes the label and resets review.

### Cherry-picking

Dependency updates are cherry-picked heavily — roughly a quarter of them target a release branch.
Policy is `docs/contributing/proposing-a-cherry-pick.md`; the mechanism is `cherry_pick_pull.sh`
from kubernetes/kubernetes, producing `Automated cherry pick of #NNNNN: <original title>`.

| Usually cherry-picked | Usually not |
|---|---|
| Go toolchain bumps (to **every** open branch — security patches) | routine `chore(channels)` alpha bumps |
| Vulnerability fixes, sometimes authored straight on the branch | dependabot GitHub Actions bumps |
| etcd / etcd-manager / containerd bumps | generic `Update dependencies` |
| Addon bumps that fix a real bug | addon version bumps with no bug attached |

Golden files differ between `master` and release branches, so a cherry-pick frequently needs
`./hack/update-expected.sh` re-run against the branch rather than a clean `git cherry-pick`.

### Attribution

Maintainers note AI assistance in the **PR body**, not in a commit trailer — for example
`Assisted by Claude Opus`. Commit trailers in use are only `Co-authored-by:`.

## Known failure modes

Real examples, each of which cost a PR:

1. **A CNI bump that fails the e2e apply must be fixed, not retested.** PR #17266 (Cilium) failed a
   server-side-apply typed-patch on `.spec.template.spec.volumes`, was retested repeatedly, went
   `lifecycle/rotten`, and was closed after seven months in favour of a fresh PR.
2. **LBC, CSI, and Karpenter bumps usually need a matching IAM policy change.** PR #18003 burned ten
   e2e retries before a maintainer identified the real cause: the upstream `iam_policy.json` had
   changed and `pkg/model/iam/iam_builder.go` needed the same update. PR #18220 shows the complete
   shape — the version, `iam_builder.go`, the `pkg/model/iam/tests/iam_builder_master_*.json`
   goldens, and the integration goldens.
3. **Do not combine a Go toolchain bump with a dependency refresh.** PR #18057 was rejected for
   exactly this and replaced by a narrower PR, because the toolchain bump has to be cherry-pickable
   to every open release branch on its own.
4. **A mangled template gets the lgtm revoked.** PR #17833 failed `pull-kops-test` because the
   regenerated `.yaml.template` was malformed. Read the regenerated diff before pushing.
5. **Never hand-edit golden fixtures.** A commit exists purely to stop a developer's kubeconfig
   leaking into the goldens. Always go through `./hack/update-expected.sh`.
6. **Stalled bumps rot.** `k8s-triage-robot` marks a PR stale after 90 days and rotten 30 days
   later. Finish or close.
