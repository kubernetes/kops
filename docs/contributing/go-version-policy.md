# Go version policy

This page describes when kOps updates the Go version used to build its binaries, and how to
perform an update. It captures the policy discussed in
[issue #17458](https://github.com/kubernetes/kops/issues/17458).

## Policy

### Baseline: follow kubernetes/kubernetes

Each kOps branch depends on a specific Kubernetes minor version. The Go minor version used by a
kOps branch should be at least as new as the one used by the corresponding kubernetes/kubernetes
branch. Kubernetes pins its Go version in the
[`.go-version`](https://github.com/kubernetes/kubernetes/blob/master/.go-version) file at the root
of its repository.

kOps does not need to be an early adopter of a new Go minor version. Waiting until
kubernetes/kubernetes has moved to it is the safe default.

### Master branch

The master branch moves to a new Go minor version when the kubernetes/kubernetes master branch
adopts it. Moving earlier is possible when there is a strong reason, for example an important
feature or fix in the new Go release, but this is the exception.

### Release branches

Supported release branches receive Go updates as well, not only master:

* New Go patch releases, which typically contain security and bug fixes, are picked up on master
  and then backported to all supported release branches.
* New Go minor versions are also rolled out to supported release branches, following the upstream
  Kubernetes practice of keeping release branches on a
  [supported Go release](https://go.dev/doc/devel/release#policy). In practice, all supported kOps
  branches converge on the same Go version as master. This keeps maintenance simple, and
  dependency updates on release branches tend to require newer Go versions anyway, since updated
  dependencies pull in newer Kubernetes libraries and the Go requirements that come with them.

A kOps release branch may therefore end up on a newer Go minor version than the
kubernetes/kubernetes branch it depends on. That is acceptable. The rule that matters is the
baseline above: never older than the corresponding Kubernetes branch.

### Who performs the update

There is no dedicated owner. Maintainers usually notice new Go releases and open the update PR,
but any contributor is welcome to do so. Updates land on master first and are then cherry-picked
to the supported release branches.

## Where the Go version is pinned

There is no `.go-version` file or `toolchain` directive in kOps. The Go version is pinned in two
kinds of places:

1. The `go` directive in `go.mod` of every module in the repository:
    * `go.mod`
    * `hack/go.mod`
    * `tests/e2e/go.mod`
    * `tests/e2e/scenarios/ai-conformance/go.mod`
    * `tests/e2e/scenarios/ai-conformance/tools/check-aws-availability/go.mod`
    * `tools/metal/dhcp/go.mod`
    * `tools/metal/storage/go.mod`
    * `tools/otel/traceserver/go.mod`
2. The `golang` builder image tags in `cloudbuild.yaml`, which is used to build and push the
   official release artifacts.

GitHub Actions workflows do not pin a version of their own. They use `actions/setup-go` with
`go-version-file` pointing at the root `go.mod`, so they follow the `go` directive automatically.

## How to update the Go version

1. Update the `go` directive in all `go.mod` files listed above. The directive uses the full
   version, for example `go 1.26.5`.
2. Update the `golang` image tags in `cloudbuild.yaml` to the matching version.
3. Run `make gomod` to tidy all modules and refresh the `vendor/` tree.
4. Run `make test` and address any issues. New Go minor versions occasionally introduce new
   `go vet` or lint findings; fixes for those belong in the same PR.
5. Open the PR against master. Once merged, cherry-pick it to the supported release branches.

For examples, see past update PRs such as [#18546](https://github.com/kubernetes/kops/pull/18546)
(patch bump on master) and [#18397](https://github.com/kubernetes/kops/pull/18397) (minor bump on
a release branch).
