# Testing preview versions

The kOps test infrastructure creates builds of git branches and pull requests.
These builds can easily be used for testing. Note that these builds are cleaned up after some time, so it is not safe to use these for production clusters.

This is handy as if you do not want to compile e.g the master branch to test a fix.

## Testing release branches

After each successful merge to a release branch, the build is made available through a release marker.

| branch | marker |
|--------|--------|
| https://storage.googleapis.com/kops-ci/bin/latest-ci-updown-green.txt | master branch |
| https://storage.googleapis.com/k8s-staging-kops/kops/releases/markers/release-1.21/latest-ci.txt | kOps 1.21 release branch |
| https://storage.googleapis.com/k8s-staging-kops/kops/releases/markers/release-1.22/latest-ci.txt | kOps 1.22 release branch |

You can create a cluster using these markers using the following scripts:

```sh
marker="https://storage.googleapis.com/kops-ci/bin/latest-ci-updown-green.txt"
export KOPS_BASE_URL="$(curl -s https://storage.googleapis.com/kops-ci/bin/latest-ci-updown-green.txt)"
wget -q "$KOPS_BASE_URL/$(go env GOOS)/$(go env GOARCH)/kops"
chmod +x ./kops
./kops version
```

Keep in mind you need to set `KOPS_BASE_URL` every time you use `./kops`

## Testing a pull request

When a PR builds successfully, you can test the PR using the following script:

```sh
pr=13208
sha=$(curl -s -H "Accept: application/vnd.github.v3+json" https://api.github.com/repos/kubernetes/kops/pulls/${pr} | jq -r .head.sha )
export KOPS_BASE_URL=https://storage.googleapis.com/kops-ci/pulls/pull-kops-e2e-kubernetes-aws/pull-v1.24.0-alpha.2-68-g8a1070a1b9
wget -q "$KOPS_BASE_URL/$(go env GOOS)/$(go env GOARCH)/kops"
chmod +x ./kops
./kops version
```