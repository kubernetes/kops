# Release Process

The kops project is released on an as-needed basis. The process is as follows:

1. An issue is proposing a new release with a changelog since the last release
1. All [OWNERS](OWNERS) must LGTM this release
1. An OWNER runs `git tag -s $VERSION` and inserts the changelog and pushes the tag with `git push $VERSION`
1. The release issue is closed
1. An announcement email is sent to `kubernetes-dev@googlegroups.com` with the subject `[ANNOUNCE] kops $VERSION is released`

## Branches

We maintain a `release-1.4` branch for kops 1.4.X, `release-1.5` for kops 1.5.X
etc.

We create new branches from master as a new kops version is released (or in
preparation for the release).

Generally we don't encourage users to run older kops versions, or older
branches, because newer versions of kops should remain compatible with older
versions of Kubernetes.

Releases should be done from the `release-1.X` branch.  The tags should be made
on the release branches.

## Update versions

See [1.5.0-alpha4 commit](https://github.com/kubernetes/kops/commit/a60d7982e04c273139674edebcb03c9608ba26a0) for example

* Edit makefile
* If updating dns-controller: bump version in Makefile, code, manifests, and tests


## Check builds OK

```
make ci
```


## Push new dns-controller image if needed

```
make dns-controller-push DNS_CONTROLLER_TAG=1.5.1 DOCKER_REGISTRY=kope
```

## Upload new version

```
# export AWS_PROFILE=??? # If needed
make upload S3_BUCKET=s3://kubeupv2
```

## Tag new version

Make sure you are on the release branch `git checkout release-1.X`

```
export TAG=1.5.0-alpha4
git tag ${TAG}
git push --tags
```

## Update release branch

For the time being, we are also maintaining a release branch.  We push released
versions to that.

`git push origin release`

## Upload to github

Manually create a release on github & upload, but soon we'll publish shipbot which automates this...

```
bazel run //cmd/shipbot -- -tag ${TAG}
```


## Compile release notes

e.g. `git log 1.5.0-alpha2..1.5.0-alpha3 > /tmp/relnotes`

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
