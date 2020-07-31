** This file documents the new release process, as used from kops 1.19
onwards.  For the process used for versions up to kops 1.18, please
see [the old release process](development/release.md)**

# Release Process

The kops project is released on an as-needed basis. The process is as follows:

1. An issue is proposing a new release with a changelog since the last release
1. All [OWNERS](https://github.com/kubernetes/kops/blob/master/OWNERS) must LGTM this release
1. An OWNER runs `git tag -s $VERSION` and inserts the changelog and pushes the tag with `git push $VERSION`
1. The release issue is closed
1. An announcement email is sent to `kubernetes-dev@googlegroups.com` with the subject `[ANNOUNCE] kops $VERSION is released`

## Branches

We maintain a `release-1.17` branch for kops 1.17.X, `release-1.18` for kops 1.18.X
etc.

`master` is where development happens.  We create new branches from master as a
new kops version is released, or in preparation for a new release.  As we are
preparing for a new kubernetes release, we will try to advance the master branch
to focus on the new functionality, and start cherry-picking back more selectively
to the release branches only as needed.

Generally we don't encourage users to run older kops versions, or older
branches, because newer versions of kops should remain compatible with older
versions of Kubernetes.

Releases should be done from the `release-1.X` branch.  The tags should be made
on the release branches.

We do currently maintain a `release` branch which should point to the same tag as
the current `release-1.X` tag.

## New Kubernetes versions and release branches

Typically Kops alpha releases are created off the master branch and beta and stable releases are created off of release branches.
In order to create a new release branch off of master prior to a beta release, perform the following steps:

1. Create a new periodic E2E prow job for the "next" kubernetes minor version.
   * All Kops prow jobs are defined [here](https://github.com/kubernetes/test-infra/tree/master/config/jobs/kubernetes/kops)
2. Create a new presubmit E2E prow job for the new release branch.
3. Create a new milestone in the GitHub repo.
4. Update [prow's milestone_applier config](https://github.com/kubernetes/test-infra/blob/dc99617c881805981b85189da232d29747f87004/config/prow/plugins.yaml#L309-L313) to update master to use the new milestone and add an entry for the new branch that targets master's old milestone.
5. Create the new release branch in git and push it to the GitHub repo.

## Update versions

See [1.19.0-alpha.1 PR](https://github.com/kubernetes/kops/pull/9494) for example

* Use the hack/set-version script to update versions:  `hack/set-version 1.20.0 1.20.1`

The syntax is `hack/set-version <new-release-version> <new-ci-version>`

`new-release-version` is the version you are releasing.

`new-ci-version` is the version you are releasing "plus one"; this is used to avoid CI jobs being out of semver order.

Examples:

| new-release-version  | new-ci-version
| ---------------------| ---------------
| 1.20.1               | 1.20.2
| 1.21.0-alpha.1       | 1.21.0-alpha.2
| 1.21.0-beta.1        | 1.21.0-beta.2


* Update the golden tests: `hack/update-expected.sh`

* Commit the changes (without pushing yet): `git add -p && git commit -m "Release 1.X.Y"`

## Check builds OK

```
rm -rf .build/ .bazelbuild/
make ci
```

## Send Pull Request to propose a release

```
git push $USER
hub pull-request
```

Wait for the PR to merge

## Tag the branch

(TODO: Can we automate this?  Maybe we should have a tags.yaml file)

```
git checkout master
git fetch origin
git reset --hard origin/master
```

Make sure you are on the correct commit, and not a newer one!

```
VERSION=$(tools/get_version.sh | grep VERSION | awk '{print $2}')
echo ${VERSION}
```

```
git tag v${VERSION}
git show v${VERSION}
```

Double check it is the correct commit!

```
git push git@github.com:kubernetes/kops v${VERSION}
git fetch origin # sync back up
```


## Wait for CI job to complete

The staging CI job should now see the tag, and build it (from the
trusted prow cluster, using Google Cloud Build).

The job is here: https://testgrid.k8s.io/sig-cluster-lifecycle-kops#kops-postsubmit-push-to-staging

It (currently) takes about 10 minutes to run.

In the meantime, you can compile the release notes...

## Compile release notes

e.g.

```
git checkout -b relnotes_${VERSION}

FROM=1.18.0
TO=1.18.1
DOC=1.18
git log v${FROM}..v${TO} --oneline | grep Merge.pull | grep -v Revert..Merge.pull | cut -f 5 -d ' ' | tac  > /tmp/prs
echo -e "\n## ${FROM} to ${TO}\n"  >> docs/releases/${DOC}-NOTES.md
relnotes  -config .shipbot.yaml  < /tmp/prs  >> docs/releases/${DOC}-NOTES.md
```

Review then send a PR with the release notes:

```
git add -p && git commit -m "Release notes for ${VERSION}"
git push ${USER}
hub pull-request
```

## Update release branch

For the time being, we are also maintaining a release branch.  We push released
versions to that.

`git push git@github.com:kubernetes/kops release-1.17:release`

## Propose promotion of artifacts

Create container promotion PR:

```
cd ${GOPATH}/src/k8s.io/k8s.io

git co -b kops_images_${VERSION}

cd k8s.gcr.io/images/k8s-staging-kops
echo "" >> images.yaml
echo "# ${VERSION}" >> images.yaml
k8s-container-image-promoter --snapshot gcr.io/k8s-staging-kops --snapshot-tag ${VERSION} >> images.yaml
```

You can dry-run the promotion with 

```
cd ${GOPATH}/src/k8s.io/k8s.io
k8s-container-image-promoter --thin-manifest-dir k8s.gcr.io
```

Currently we send the image and non-image artifact promotion PRs separately.

```
git add -p
git commit -m "Promote kops $VERSION images"
git push ${USER}
hub pull-request
```


Create binary promotion PR:

```
cd ${GOPATH}/src/k8s.io/k8s.io

git co master
git co -b kops_artifacts_${VERSION}

mkdir -p ./k8s-staging-kops/kops/releases/${VERSION}/
gsutil rsync -r  gs://k8s-staging-kops/kops/releases/${VERSION}/ ./k8s-staging-kops/kops/releases/${VERSION}/

promobot-generate-manifest --src k8s-staging-kops/kops/releases/ >> artifacts/manifests/k8s-staging-kops/${VERSION}.yaml
```

Verify, then send a PR:

```
git add artifacts/manifests/k8s-staging-kops/${VERSION}.yaml
git commit -m "Promote kops $VERSION binary artifacts"
git push ${USER}
hub pull-request
```


## Promote to dockerhub / S3 / github (legacy)

We are in the process of moving to k8s.gcr.io for all images and to
artifacts.k8s.io for all non-image artifacts.

In the meantime (and for compatibility), we must also promote to the old locations:

Images to dockerhub:

```
crane cp gcr.io/k8s-staging-kops/kube-apiserver-healthcheck:${VERSION} kope/kube-apiserver-healthcheck:${VERSION}
crane cp gcr.io/k8s-staging-kops/dns-controller:${VERSION} kope/dns-controller:${VERSION}
crane cp gcr.io/k8s-staging-kops/kops-controller:${VERSION} kope/kops-controller:${VERSION}
```


Binaries to S3 bucket:

```
aws s3 sync --acl public-read k8s-staging-kops/kops/releases/${VERSION}/ s3://kubeupv2/kops/${VERSION}/
```

Binaries to github:

```
cd ${GOPATH}/src/k8s.io/kops/
shipbot -tag v${VERSION} -config .shipbot.yaml -src ${GOPATH}/src/k8s.io/k8s.io/k8s-staging-kops/kops/releases/${VERSION}/
```


Until the binary promoter is automatic, we also need to promote the binary artifacts manually:

```
mkdir -p /tmp/thin/artifacts/filestores/k8s-staging-kops/
mkdir -p /tmp/thin/artifacts/manifests/k8s-staging-kops/

cd ${GOPATH}/src/k8s.io/k8s.io
cp artifacts/manifests/k8s-staging-kops/${VERSION}.yaml /tmp/thin/artifacts/manifests/k8s-staging-kops/

cat > /tmp/thin/artifacts/filestores/k8s-staging-kops/filepromoter-manifest.yaml << EOF
filestores:
- base: gs://k8s-staging-kops/kops/releases/
  src: true
- base: gs://k8s-artifacts-prod/binaries/kops/
  service-account: k8s-infra-gcr-promoter@k8s-artifacts-prod.iam.gserviceaccount.com
EOF

promobot-files --filestores /tmp/thin/artifacts/filestores/k8s-staging-kops/filepromoter-manifest.yaml --files /tmp/thin/artifacts/manifests/k8s-staging-kops/ --dry-run=true
```

After validation of the dry-run output:
```
promobot-files --filestores /tmp/thin/artifacts/filestores/k8s-staging-kops/filepromoter-manifest.yaml --files /tmp/thin/artifacts/manifests/k8s-staging-kops/ --dry-run=false --use-service-account
```

## Smoketesting the release

```
wget https://artifacts.k8s.io/binaries/kops/${VERSION}/linux/amd64/kops

mv kops ko
chmod +x ko

ko version
```

Also run through a kops create cluster flow, ideally verifying that
everything is pulling from the new locations.

## On github

* Download release
* Validate it
* Add notes
* Publish it

## Release kops to homebrew

* Following the [documentation](homebrew.md) we must release a compatible homebrew formulae with the release.
* This should be done at the same time as the release, and we will iterate on how to improve timing of this.

## Update the alpha channel and/or stable channel

Once we are satisfied the release is sound:

* Bump the kops recommended version in the alpha channel

Once we are satisfied the release is stable:

* Bump the kops recommended version in the stable channel

## Update conformance results with CNCF

Use the following instructions: https://github.com/cncf/k8s-conformance/blob/master/instructions.md
