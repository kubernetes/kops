# Release Process

The kops project is released on an as-needed basis. The process is as follows:

1. An issue is proposing a new release with a changelog since the last release
1. All [OWNERS](https://github.com/kubernetes/kops/blob/master/OWNERS) must LGTM this release
1. An OWNER runs `git tag -s $VERSION` and inserts the changelog and pushes the tag with `git push $VERSION`
1. The release issue is closed
1. An announcement email is sent to `kubernetes-dev@googlegroups.com` with the subject `[ANNOUNCE] kops $VERSION is released`

## Branches

We maintain a `release-1.4` branch for kops 1.4.X, `release-1.5` for kops 1.5.X
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


## Update versions

See [1.5.0-alpha4 commit](https://github.com/kubernetes/kops/commit/a60d7982e04c273139674edebcb03c9608ba26a0) for example

* Edit makefile
* If updating dns-controller: bump version in Makefile, code, manifests, and tests

`git commit -m "Release 1.X.Y`

## Check builds OK

```
rm -rf .build/ .bazelbuild/
make ci
```


## Push new kops-controller / dns-controller image if needed

```
make dns-controller-push DOCKER_IMAGE_PREFIX=kope/  DOCKER_REGISTRY=index.docker.io
make kops-controller-push DOCKER_IMAGE_PREFIX=kope/  DOCKER_REGISTRY=index.docker.io
```

## Upload new version

```
# export AWS_PROFILE=??? # If needed
make upload UPLOAD_DEST=s3://kubeupv2
```

## Tag new version

Make sure you are on the release branch `git checkout release-1.X`

```
make release-tag
git push git@github.com:kubernetes/kops
git push --tags git@github.com:kubernetes/kops
```

## Update release branch

For the time being, we are also maintaining a release branch.  We push released
versions to that.

`git push origin release-1.8:release`

## Pull request to master branch (for release commit)

## Upload to github

Use [shipbot](https://github.com/kopeio/shipbot) to upload the release:

```
make release-github
```


## Compile release notes

e.g.

```
git log 1.14.0-beta.1..1.14.0 --oneline | grep Merge.pull | cut -f 5 -d ' ' | tac  > /tmp/prs
relnotes  -config .shipbot.yaml  < /tmp/prs  >> docs/releases/1.14-NOTES.md
```

## On github

* Download release
* Validate it
* Add notes
* Publish it

## Update the alpha channel and/or stable channel

Once we are satisfied the release is sound:

* Bump the kops recommended version in the alpha channel

Once we are satisfied the release is stable:

* Bump the kops recommended version in the stable channel

## Update conformance results with CNCF

Use the following instructions: https://github.com/cncf/k8s-conformance/blob/master/instructions.md

