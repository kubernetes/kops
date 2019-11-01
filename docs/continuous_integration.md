# Continuous Integration

Using Kops' declarative manifests it is possible to create and manage clusters entirely in a CI environment.
Rather than using `kops create cluster` and `kops edit cluster`, the cluster and instance group manifests can be stored in version control.
This allows cluster changes to be made through reviewable commits rather than on a local workstation.
This is ideal for larger teams in order to avoid possible collisions from simultaneous changes being made.
It also provides an audit trail, consistent environment, and centralized view for any Kops commands being ran.

Running Kops in a CI environment can also be useful for upgrading Kops.
Simply download a newer version in the CI environment and run a new pipeline.
This will allow any changes to be reviewed prior to being applied.
This strategy can be extended to sequentially upgrade Kops on multiple clusters, allowing changes to be tested on non-production environments first.

This page provides examples for managing Kops clusters in CI environments.
The [Manifest documentation](./manifests_and_customizing_via_api.md) describes how to create the YAML manifest files locally and includes high level examples of commands described below.

If you have a solution for a different CI platform or deployment strategy, feel free to open a Pull Request!

## GitLab CI

[GitLab CI](https://about.gitlab.com/product/continuous-integration/) is built into GitLab and allows commits to trigger CI pipelines.

### Requirements

* The GitLab runners that run the jobs need the appropriate permissions to invoke the Kops commands.
  For clusters in AWS this means providing AWS IAM credentials either with IAM User Keys set as secret variables in the project, or having the runner run on an EC2 instance with an instance profile attached.


### Example Workflow

1. A cluster administrator makes a change to a cluster manifest, commits and pushes to a feature branch on GitLab and opens a Merge Request
2. A reviewer reviews the change to confirm it is as intended, and approves or merges the MR
3. A "master" pipeline is triggered from this merge commit which runs a `kops update cluster`.
4. The administrator reviews the output of the `dryrun` job to confirm the desired changes and initiates the `update` job which runs `kops update cluster --yes`.
5. Once the cluster is updated, `kops rolling-update cluster` is ran which can be used to confirm any nodes that need replacement. The administrator then starts the `roll` job which runs `kops rolling-update cluster --yes` and replaces any nodes as necessary.

```yaml
# .gitlab-ci.yml
stages:
  - dryrun
  - update
  - roll

variables:
  KOPS_CLUSTER_NAME: ...
  KOPS_STATE_STORE: ...

dryrun:
  stage: dryrun
  only:
    - master@namespace/project_name
  script:
    - kops replace --force -f cluster.yml
    - kops update cluster

update:
  stage: update
  only:
    - master@namespace/project_name
  when: manual
  script:
    - kops update cluster --yes
    - kops rolling-update cluster

roll:
  stage: roll
  only:
    - master@namespace/project_name
  when: manual
  script:
    - kops rolling-update cluster --yes
```

### Considerations

* The `only:` field in each job will need to be updated to reflect the real project's [namespace](https://docs.gitlab.com/ce/user/group/#namespaces) and name.
  The two variables will also need to be set to real values.
* The jobs that make actual changes to the clusters are manually invoked (`when: manual`) though this could easily be removed to make them automatic.
* This pipeline setup will create and update existing clusters in place. It does not perform a "blue/green" deployment of multiple clusters.
* The pipeline can be extended to support multiple clusters by making separate jobs per cluster for each stage.
  Ensure the `KOPS_CLUSTER_NAME` variable is set correctly for each set of jobs.

  In this case, it is possible to use `kops toolbox template` to manage one YAML template and per-cluster values files with which to render the template.
  See the [Cluster Template](./operations/cluster_template.md) documentation for more information.
  `kops toolbox template` would then be ran before `kops replace`.

### Limitations

* This pipeline does not have a true "dryrun" job that can be ran on non-master branches, for example before a merge request is merged.
  This is because the required `kops replace` before the `kops update cluster` will update the live assets in the state store which could impact newly launched nodes that download these assets.
  [PR #6465](https://github.com/kubernetes/kops/pull/6465) could add support for copying the state store to a local filesystem prior to `kops replace`, allowing the dryrun pipeline to be completely isolated from the live state store.