# Updating dependencies

kOps pins third-party versions in a lot of places: Go modules, addon manifests, the binaries nodeup
downloads onto nodes, GitHub Actions, e2e scenario charts, and the OS images in the channels. This
page covers how to go about a dependency update and how to get the pull request merged.

For the mechanics of an individual pin — which file holds it, what else must change with it, and how
to regenerate and verify — see the
[third-party dependency reference](../dependency-updates.md), which enumerates every surface.

## Almost none of this is automated

`.github/dependabot.yml` registers the `github-actions` ecosystem and nothing else. There is no
dependabot for Go modules, container images, or Python packages. Every other bump in the repo is a
manual pull request, which is why the surfaces below are worth knowing about.

Dependabot has also been disabled in the past without any change to `dependabot.yml`. If GitHub
Actions pins look stale, check whether it is still opening pull requests before assuming they are up
to date.

## The shape of a dependency update

Almost every update follows the same four steps:

1. **Find the new upstream version**, and read what changed. For a vendored addon manifest this
   means the upstream release; for a Go module, the changelog.
2. **Edit the pin**, along with anything coupled to it. Coupling is the part that catches people
   out: a version often appears in more than one place, and the second place rarely fails loudly.
3. **Regenerate.** Never hand-edit generated output.
4. **Verify**, then open the PR.

### Regeneration is not optional

Two generated artifacts cause most of the failed CI runs on dependency PRs.

**Golden test fixtures.** Files under `tests/integration/` embed image tags, versions and the hashes
of the objects kOps writes. Changing a default version changes dozens — sometimes hundreds — of
them. `./hack/update-expected.sh` regenerates them. It deliberately unsets your cloud credentials
and kubeconfig first, so always run it rather than editing fixtures by hand or invoking the
underlying `go test` yourself.

**Asset hashes.** Every binary kOps downloads onto a node has its SHA256 stored in
`pkg/assets/assetdata/`. These are *not* checked at compile time: a version bumped without a
matching hash entry builds fine, passes most tests, and then fails for users at
`kops update cluster`. Regenerate with `./hack/generate-asset-hashes.sh`.

### Vendoring Go dependencies

kOps commits its `vendor/` directory, and `go.mod` files exist in several nested modules as well as
the root. To add or update a Go dependency:

1. Edit the version in the owning `go.mod`, or add the import to a `.go` file.
2. Run `make gomod`. This tidies the root module, revendors, and tidies every nested module.
3. Review the diff — particularly for a dependency you have not used before.
4. Commit `go.mod`, `go.sum` and `vendor/` together.

A bare `go get` is never enough; `pull-kops-verify-gomod` will fail. Keep the dependency PR separate
from feature work so the vendor churn does not bury the change under review.

## Commits

- **One dependency concern per PR.** In particular, do not combine a Go toolchain bump with a
  general dependency refresh — the toolchain bump needs to be cherry-pickable to every open release
  branch on its own.
- **kOps merges with merge commits, not squash**, so your individual commit messages land verbatim
  on `master` and are worth writing carefully.
- **Split the bump from the regeneration.** The established shape is two commits: the hand-written
  change, then the generated churn in a commit named after the command that produced it —
  `./hack/update-expected.sh`, `make gomod`, `./hack/generate-asset-hashes.sh`. Reviewers read the
  first and skim the second.
- **Titles are free-form.** Despite `AGENTS.md` citing conventional commits, most dependency PR
  titles carry no prefix at all. Three styles are all normal:

  | Style | Shape |
  |---|---|
  | Bare imperative (most common) | `Update <Thing> to <version>`, `Upgrade <Thing> to <version>`, `Update dependencies` |
  | `<component>: <imperative>` | `aws: Update EBS CSI driver to <version>`, `etcd-manager: upgrade to <version>` |
  | Conventional commits | `chore(networking): bump <thing> to <version>`, `chore(channels): ...` |

  Match the surrounding history for the file you are editing rather than imposing a format.

## Describing the change

The bump PRs that get reviewed fastest say **which upstream release this is and why it matters**:
link the changelog or release notes, and name the behaviour that changed. If you are bumping for a
vulnerability, name the advisory — a PR that just says "bump for CVEs" gives a reviewer nothing to
check against.

Say so too if the bump is expected to be inert, for example a patch release you are taking purely to
stay current. That is useful information, not a reason to skip the description.

## When a dependency PR fails CI

Most red runs on a dependency change come from generated output that was not regenerated:

| Symptom | Cause | Fix |
|---|---|---|
| integration test failures showing an old version | golden fixtures embed the version you changed | `./hack/update-expected.sh` |
| go.mod verification fails | `go.mod` edited without revendoring and tidying the nested modules | `make gomod` |
| generated-code verification fails | codegen output is stale | `make apimachinery` / `make crds` |

e2e failures are more often flakes than real regressions, but do not assume it: a bump to a CNI, a
CSI driver or a controller is exactly the kind of change that breaks an e2e job for real. Read the
failure before retrying it.

Some e2e jobs are optional and only run when asked. A bump to the load balancer controller,
Karpenter, kube-router or the Azure scenarios should be tested against the job that covers it rather
than merged on the default set.

## Cherry-picking

Dependency updates are cherry-picked to release branches more often than most changes. See
[proposing a cherry-pick](proposing-a-cherry-pick.md) for how.

What tends to qualify: Go toolchain bumps, which carry security patches and go to every open branch;
vulnerability fixes; etcd, etcd-manager and containerd bumps; and addon bumps that fix a real bug.
What tends not to: routine channel bumps, GitHub Actions bumps, and version bumps with no bug
attached.

Golden files differ between `master` and the release branches, so a cherry-pick often needs
`./hack/update-expected.sh` re-run against the branch rather than a clean `git cherry-pick`.

## Things that have gone wrong before

Each of these cost a pull request:

1. **A CNI bump that fails the e2e apply must be fixed, not retested.** A Cilium bump failed a
   server-side-apply typed-patch on `.spec.template.spec.volumes`, was retested repeatedly, went
   `lifecycle/rotten`, and was closed after seven months in favour of a fresh PR.
2. **Load balancer controller, CSI and Karpenter bumps usually need a matching IAM policy change.**
   One LBC bump burned ten e2e retries before a maintainer spotted that the upstream
   `iam_policy.json` had changed and `pkg/model/iam/iam_builder.go` needed the same update. Check
   the upstream policy whenever you bump a controller that talks to a cloud API.
3. **Read the regenerated diff before pushing.** A Cilium bump failed its tests because the
   regenerated `.yaml.template` was malformed, and the breakage was in generated output nobody had
   looked at.
4. **Never hand-edit golden fixtures.** A commit exists purely to stop a developer's kubeconfig
   leaking into them.
