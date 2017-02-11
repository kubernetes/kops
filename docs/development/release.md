# Release Process

The kops project is released on an as-needed basis. The process is as follows:

1. An issue is proposing a new release with a changelog since the last release
1. All [OWNERS](OWNERS) must LGTM this release
1. An OWNER runs `git tag -s $VERSION` and inserts the changelog and pushes the tag with `git push $VERSION`
1. The release issue is closed
1. An announcement email is sent to `kubernetes-dev@googlegroups.com` with the subject `[ANNOUNCE] kops $VERSION is released`


## Update versions

See [1.5.0-alpha4 commit](https://github.com/kubernetes/kops/commit/a60d7982e04c273139674edebcb03c9608ba26a0) for example

* Edit makefile
* If updating dns-controller: bump version in Makefile, code, manifests, and tests


## Check builds OK

```
make ci
```


## Push new protokube image if needed

```
make dns-controller-push DNS_CONTROLLER_TAG=1.5.1 DOCKER_REGISTRY=kope
```

## Upload new version

```
# export AWS_PROFILE=??? # If needed
make upload S3_BUCKET=s3://kubeupv2
```

## Tag new version

```
export TAG=1.5.0-alpha4
git tag ${TAG}
git push --tags
```

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
