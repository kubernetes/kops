## Aug 11 2016

Reworked SSH keys and support for running CI builds

* SSH keys are now stored as secrets.  `--ssh-public-key` will be created when you do `kops create cluster`.
  You no longer need to specify a `--ssh-public-key` when you do an update, but if you do it will be imported.
* An SSH public key must exist for AWS, if you do not have one you can import one with:
  `kops create secret --name $CLUSTER_NAME sshpublickey admin -i ~/.ssh/id_rsa.pub`
* For AWS, only a single SSH key can be used; you can delete extra keys with `kops delete secret`
* To support changing SSH keys reliably, the name of the imported AWS SSH keypair will change to include
  the OpenSSH key fingerprint.  Existing clusters will continue to work, but you will likely be prompted to
  do a rolling update when you would otherwise not have to.  I suggest waiting till you next upgrade kubernetes.

* Builds that are not published as Docker images can be run.  `kops` will pass a list of images in the NodeUp
  configuration, and NodeUp will download and `docker load` these images.  For examples, see the
  [testing tips](../development/testing.md)

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

## Stable Channel:

* Update to kubernetes 1.4.3 (highly recommended update)
* Image update includes kernel 4.4.26 (address CVE-2016-5195)

## 1.4.1

* Fix dns-controller when multiple HostedZones with the same name
* Initial support for CentOS / RHEL7
* Initial k8s-style API & examples

## 1.4.0

* Initial stable release

