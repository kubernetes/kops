# Testing new configurations

When adding support for a new cluster configuration — a new API field, a new
topology, a new networking option, or a new piece of cloud infrastructure —
we have three main kinds of test coverage. This page describes when to use
each. It applies equally to humans and to AI coding assistants.

## Prefer an `update cluster` golden test

The scenarios under [`tests/integration/update_cluster/`](https://github.com/kubernetes/kops/tree/master/tests/integration/update_cluster)
(registered in `cmd/kops/integration_test.go`) run the full
create → populate → model → render pipeline against mock clouds, and compare
the rendered Terraform (and generated data files) against golden output
checked into the tree.

This is normally the best first line of coverage for a new configuration:

* It exercises the real code path end-to-end, rather than a single function.
* The effect of a change is visible in review as a Terraform diff, which is
  much easier to evaluate than assertions in test code.
* Once the scenario exists, future PRs touching the same area automatically
  show their effect on the golden output.

In particular, **prefer adding or extending a golden scenario over writing a
bespoke unit test for a model builder**. A hand-written test that asserts a
task field is set is usually redundant with a golden scenario that renders
that task; the golden test gives the same coverage plus reviewable output.
Bespoke unit tests are still the right tool for isolated, pure functions
(parsers, validation, CIDR math, and similar).

To add a scenario:

1. Copy the closest existing scenario directory (for example
   `minimal_gce_private`) and edit `in-v1alpha2.yaml`.
2. Register a test function in `cmd/kops/integration_test.go`.
3. Seed the scenario's `data/` directory with at least one file (for example
   copy `data/aws_s3_object_kops-version.txt_content` from the scenario you
   started from), so the harness knows data files are expected.
4. Generate the golden output, then run the test again to confirm it passes:

   ```bash
   HACK_UPDATE_EXPECTED_IN_PLACE=1 go test ./cmd/kops/ -run TestMyScenario
   go test ./cmd/kops/ -run TestMyScenario
   ```

5. Review the generated `kubernetes.tf` and `data/` files carefully — they are
   the assertion. In particular check `data/aws_s3_object_cluster-completed.spec_content`,
   which records the fully-populated cluster spec; defaulting and assignment
   bugs show up there.

## Add a `create cluster` golden test when the CLI path needs coverage

The `update cluster` scenarios start from a cluster YAML file, so they do not
cover `kops create cluster` flag handling, defaulting, or subnet/topology
setup. When the new configuration involves CLI behavior — or when the create
path has historically had bugs for it — add a scenario under
[`tests/integration/create_cluster/`](https://github.com/kubernetes/kops/tree/master/tests/integration/create_cluster)
(registered in `cmd/kops/create_cluster_integration_test.go`). Each scenario
is an `options.yaml` (the `kops create cluster` options) plus an
`expected-v1alpha2.yaml` golden cluster spec, regenerated with the same
`HACK_UPDATE_EXPECTED_IN_PLACE=1` mechanism.

As an example, GCE IPv6 needed one of these: the golden `update cluster` test
started from YAML and passed, while `kops create cluster --ipv6` on GCE failed
outright.

## Add an e2e job in test-infra for real-cloud verification

Golden tests verify what we *would* create, not that it actually works. For
configurations whose correctness depends on real cloud behavior (instances
booting, CNI functioning, cloud APIs accepting what we render), add a periodic
e2e job in
[test-infra](https://github.com/kubernetes/test-infra/blob/master/config/jobs/kubernetes/kops/build_jobs.py):
edit `build_jobs.py`, re-run it to regenerate the job YAML files, and add the
job to the appropriate testgrid dashboards.

For a feature that is still a work in progress, it is fine — and useful — to
add the job as a canary that is expected to be red at first; it tracks
progress and catches regressions as the feature lands. Say so in the job's
comment and PR description.

## Verifying by hand against a real cloud account

When investigating whether a configuration behaves as intended, you can often
avoid creating any resources: `kops create cluster --dry-run -oyaml` and
`kops create -f` into a scratch state store followed by inspecting the stored
config only make read-only cloud API calls, and quickly show what the spec
population code actually did.
