## Jul 21 2016

More rational model/UX - `kops create cluster` just creates spec, `kops update cluster` does real creation:

* `kops create cluster` now creates the spec, but will not normally create the actual cloud resources.  You can
  specify `--yes` to force immediate creation if you want to.  create will now fail on an existing cluster.
* `kops update cluster` will now apply changes from the spec to the cloud - it will create or update your cluster.
  It also defaults to dryrun mode, so you should pass `--yes` (normally after checking the preview).
* Most commands accept positional arguments for the cluster name (you can specify `kops update cluster <name>`,
  instead of `kops update cluster --name <name>`)
* Dry-run should be the default for anything that makes changes to cloud resources.  Pass `--yes` to confirm.
  (cleaning up an inconsistency between `--dryrun` and `--yes` by removing `--dryrun` and making it the default)