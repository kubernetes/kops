# Release process

The kOps project is released on an as-needed basis. The process is as follows:

## Release branches

We maintain a `release-1.21` branch for kOps 1.21.X, `release-1.22` for kOps 1.22.X,
etc.

`master` is where development happens.  We create new branches from master as we
prepare for a new minor release.  As we are
preparing for a new Kubernetes release, we will try to advance the master branch
to focus on the new functionality and cherry-pick back
to the release branches only as needed.

Generally we don't encourage users to run older kOps versions, or older
branches, because newer versions of kOps should remain compatible with older
versions of Kubernetes.

Beta and stable releases (excepting the first beta of a new minor version) should be made from the `release-1.X` branch.
Alpha releases may be made on either `master` or a release branch.

## Creating new release branches

Typically, kOps alpha releases are created off the master branch and beta and stable releases are created off of release branches.
The exception is the first beta release for a new minor version: it is where the release branch for the minor version branches off of master.
In order to create the first beta release for a new minor version and a new release branch off of master, perform the following steps:

1. Update the periodic E2E Prow jobs for the "next" kOps/Kubernetes minor version.
   * Edit [build_jobs.py](https://github.com/kubernetes/test-infra/tree/master/config/jobs/kubernetes/kops/build_jobs.py)
   to add the new minor version to `k8s_versions` and `kops_versions`.
     Also update the list of minor versions in `generate_versions()`, `generate_pipeline()`, and `generate_presubmits_e2e()`.
   * Edit the [testgrid config.yaml](https://github.com/kubernetes/test-infra/blob/master/config/testgrids/kubernetes/kops/config.yaml)
   to add the new minor version to both lists in the file, prefixed with `kops-k8s-`.
   * Remove the oldest minor version from each of those lists.
   * Run the `build_jobs.py` script.
2. Create a new milestone in the GitHub repo.
3. Update [prow's milestone_applier config](https://github.com/kubernetes/test-infra/blob/dc99617c881805981b85189da232d29747f87004/config/prow/plugins.yaml#L309-L313)
   to update master to use the new milestone and add an entry for the new feature branch. 
   Create this as a separate PR as it will require separate review.
4. Create the .0-beta.1 release per the instructions in the following section. GitHub Actions will create the release branch when it tags the release.
5. On the master branch, create a PR to update to the next minor version:
   * Update `OldestSupportedKubernetesVersion` and `OldestRecommendedKubernetesVersion` in
   [apply_cluster.go](https://github.com/kubernetes/kops/tree/master/upup/pkg/fi/cloudup/apply_cluster.go)
   * Add a row for the new minor version to [upgrade_k8s.md](https://github.com/kubernetes/kops/tree/master/permalinks/upgrade_k8s.md)
   * Fix any tests broken by the now-unsupported versions.
   * Create release notes for the next minor version. The release notes should mention the
   Kubernetes support removal and deprecation.
6. On master, off of the branch point, create the first alpha release for the new minor release.

## Creating releases

### Send Pull Request to propose a release

See [1.22.0-beta.2 PR](https://github.com/kubernetes/kops/pull/12467) for an example.

Use the hack/set-version script to update versions, using the new version as the argument.
Then update the golden tests.

```
hack/set-version 1.22.0
hack/update-expected.sh
```

Commit the changes (without pushing yet):

```
VERSION=$(tools/get_version.sh | grep VERSION | awk '{print $2}')
git checkout -b release_${VERSION}
git add . && git commit -m "Release ${VERSION}"
```

This is the "release commit". Push and create a PR.

```
gh pr create -f
```

Wait for the PR to merge.

### Reviewing the release commit PR

To review someone else's release commit, verify that:

* A release at that point is desired. (For example, there are no unfixed blocking bugs.)
* There is nothing in the commit besides version number updates and golden outputs.

The "verify-versions" CI task will ensure that the versions have been updated in all the
expected places.

### Wait for CI job to complete

After the PR merges, GitHub Actions will tag the release.
The [staging CI job](https://testgrid.k8s.io/sig-cluster-lifecycle-kops#kops-postsubmit-push-to-staging) should build from the tag (from the trusted prow cluster, using Google Cloud Build).

It (currently) takes about 30 minutes to run.

In the meantime, you can compile the release notes...

### Compile release notes

This step is not necessary for an ".0-alpha.1" release as these are made off
of the branch point for the previous minor release.

The `relnotes` tool is from [kopeio/shipbot](https://github.com/kopeio/shipbot).

For example:

```
git checkout master
git pull upstream master
git checkout -b relnotes_${VERSION}

FROM=1.21.0-alpha.2 # Replace "1.21.0-alpha.2" with the previous version
DOC=$(expr ${VERSION} : '\([^.]*.[^.]*\)')
git log v${FROM}..v${VERSION} --oneline | grep Merge.pull | grep -v Revert..Merge.pull | cut -f 5 -d ' ' | tac  > /tmp/prs
echo -e "\n## ${FROM} to ${VERSION}\n"  >> docs/releases/${DOC}-NOTES.md
relnotes  -config .shipbot.yaml  < /tmp/prs  >> docs/releases/${DOC}-NOTES.md
```

Review then send a PR with the release notes:

```
git add -p docs/releases/${DOC}-NOTES.md
git commit -m "Release notes for ${VERSION}"
gh pr create -f
```

### Propose promotion of artifacts

The following tools are prerequisites:

* [`gsutil`](https://cloud.google.com/storage/docs/gsutil_install)
* [`kpromo`](https://github.com/kubernetes-sigs/promo-tools)

Create container promotion PR:

```
cd ${GOPATH}/src/k8s.io/k8s.io

git checkout main
git pull upstream main
git checkout -b kops_images_${VERSION}

cd k8s.gcr.io/images/k8s-staging-kops
echo "" >> images.yaml
echo "# ${VERSION}" >> images.yaml
kpromo cip run --snapshot gcr.io/k8s-staging-kops --snapshot-tag ${VERSION} >> images.yaml
```

Currently we send the image and non-image artifact promotion PRs separately.

```
cd ${GOPATH}/src/k8s.io/k8s.io
git add -p k8s.gcr.io/images/k8s-staging-kops/images.yaml
git commit -m "Promote kOps $VERSION images"
gh pr create -f
```

Create binary promotion PR:

```
cd ${GOPATH}/src/k8s.io/k8s.io

git checkout main
git pull upstream main
git checkout -b kops_artifacts_${VERSION}

rm -rf ./k8s-staging-kops/kops/releases
mkdir -p ./k8s-staging-kops/kops/releases/${VERSION}/
gsutil rsync -r  gs://k8s-staging-kops/kops/releases/${VERSION}/ ./k8s-staging-kops/kops/releases/${VERSION}/

kpromo manifest files --src k8s-staging-kops/kops/releases/ >> artifacts/manifests/k8s-staging-kops/${VERSION}.yaml
```

Verify, then send a PR:

```
git add artifacts/manifests/k8s-staging-kops/${VERSION}.yaml
git commit -m "Promote kOps $VERSION binary artifacts"
gh pr create -f
```

Upon approval and merge of the binary promotion PR, artifacts will be promoted
to artifacts.k8s.io via postsubmit. The process is described in detail
[here](https://git.k8s.io/k8s.io/artifacts/README.md).

### Promote to GitHub (all releases)

The `shipbot` tool is from [kopeio/shipbot](https://github.com/kopeio/shipbot).

Binaries to github (all releases):

```
cd ${GOPATH}/src/k8s.io/kops/
git checkout v$VERSION
shipbot -tag v${VERSION} -config .shipbot.yaml -src ${GOPATH}/src/k8s.io/k8s.io/k8s-staging-kops/kops/releases/${VERSION}/
```

### Promote to S3 (stable releases < 1.22)

```
aws s3 sync --acl public-read ${GOPATH}/src/k8s.io/k8s.io/k8s-staging-kops/kops/releases/${VERSION}/ s3://kubeupv2/kops/${VERSION}/
```

### Smoke test the release

This step is only necessary for stable releases (as binary artifacts are not otherwise promoted to artifacts.k8s.io).

```
wget https://artifacts.k8s.io/binaries/kops/${VERSION}/linux/amd64/kops

mv kops ko
chmod +x ko

./ko version
```

Also run through a `kops create cluster` flow, ideally verifying that
everything is pulling from the new locations.

### Publish to GitHub

* Download release
* Validate it
* Add notes
* Publish it

### Release to Homebrew

This step is only necessary for stable releases in the latest stable minor version.

* Following the [documentation](homebrew.md) we must release a compatible homebrew formulae with the release.
* This should be done at the same time as the release, and we will iterate on how to improve timing of this.

### Update conformance results with CNCF

This step is only necessary for a first stable minor release (a ".0").

Use the following instructions: https://github.com/cncf/k8s-conformance/blob/master/instructions.md

### Update latest minor release in documentation

This step is only necessary for a first stable minor release (a ".0").

Create a PR that updates the following documents:

* Rotate the new version into the version matrix in both
[releases.md](https://github.com/kubernetes/kops/tree/master/docs/welcome/releases.md)
and [README-ES.md](https://github.com/kubernetes/kops/tree/master/README-ES.md).
* Remove the "has not been released yet" header in the version's release notes.

### Add link to release notes

This step is only necessary for a first beta minor release (a ".0-beta.1").

Create a PR that updates the following document:

* Add a reference to the version's release notes in [mkdocs.yml](https://github.com/kubernetes/kops/tree/master/mkdocs.yml)

### Update the alpha channel and/or stable channel

Once we are satisfied the release is sound:

* Bump the kOps recommended version in the alpha channel

Once we are satisfied the release is stable:

* Bump the kOps recommended version in the stable channel
