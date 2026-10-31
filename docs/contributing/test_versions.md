# Testing preview versions

The kOps test infrastructure creates builds of git branches.
These builds can easily be used for testing. Note that these builds are cleaned up after some time, so it is not safe to use these for production clusters.

This is handy as if you do not want to compile e.g the master branch to test a fix.

## Testing release branches

After each successful merge to a release branch, the build is made available through a release marker.

| branch | marker |
|--------|--------|
| https://storage.googleapis.com/k8s-staging-kops/kops/releases/markers/master/latest-ci-updown-green.txt | master branch |
| https://storage.googleapis.com/k8s-staging-kops/kops/releases/markers/release-1.21/latest-ci.txt | kOps 1.21 release branch |
| https://storage.googleapis.com/k8s-staging-kops/kops/releases/markers/release-1.22/latest-ci.txt | kOps 1.22 release branch |

You can create a cluster using these markers using the following scripts:

```sh
marker="https://storage.googleapis.com/k8s-staging-kops/kops/releases/markers/master/latest-ci-updown-green.txt"
export KOPS_BASE_URL="$(curl -s https://storage.googleapis.com/k8s-staging-kops/kops/releases/markers/master/latest-ci-updown-green.txt)"
wget -q "$KOPS_BASE_URL/$(go env GOOS)/$(go env GOARCH)/kops"
chmod +x ./kops
./kops version
```

Keep in mind you need to set `KOPS_BASE_URL` every time you use `./kops`

## Testing a pull request

The presubmit jobs stage their builds in temporary buckets that are deleted when the job finishes,
so pull request builds are not kept for testing. Instead, build the pull request yourself with one
of the development scripts. They upload the build to a bucket in your own account, install `kops`
to `$GOPATH/bin` and set `KOPS_BASE_URL` and `KOPS_STATE_STORE`:

```sh
pr=13208
git fetch https://github.com/kubernetes/kops.git "pull/${pr}/head"
git checkout FETCH_HEAD
. hack/dev-build-aws.sh # or . hack/dev-build-gce.sh
kops version
```

See the comments in the scripts for the permissions that the cluster needs to read the build.
