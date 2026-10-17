# Testing preview versions

The kOps test infrastructure creates builds of git branches and pull requests.
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

### Version-only markers

Cloud Build also publishes markers containing just the build version under `https://storage.googleapis.com/k8s-staging-kops/kops/releases/`:

| Branch | Marker |
|--------|--------|
| `master` | `latest.txt` |
| `release-1.X` | `latest-1.X.txt` |

Release branches publish these markers after the publishing change is applied to them. These markers identify the latest published build; they do not indicate that E2E tests passed. The existing URL-valued markers, including the updown-green markers, remain available.

`kubetest2 kops --kops-version-marker` accepts both formats. For a version-only marker, it downloads artifacts from `<marker-directory>/<version>/`. The version string and artifact directories are unchanged. For example:

```sh
marker="https://storage.googleapis.com/k8s-staging-kops/kops/releases/latest.txt"
version=$(curl -fsSL "$marker")
export KOPS_BASE_URL="${marker%/*}/${version}"
```

The `--publish-version-marker` flag copies the source marker body unchanged. Do not copy a version-only marker to an existing green-marker path such as `markers/master/latest-ci-updown-green.txt`: downloads would resolve relative to that directory, and existing shell consumers expect an absolute artifact URL. Before migrating jobs that publish green markers, update the publisher to write absolute artifact URLs to the existing destinations, or place new green markers beside the version directories and migrate their consumers together.

Version-only markers allow a CDN to serve markers and artifacts under the same URL prefix. Configure and verify that route before switching jobs to CDN marker URLs. Keep existing markers until their consumers have migrated. CDN configuration and cache purging are managed separately from the build.

## Testing a pull request

When a PR builds successfully, you can test the PR using the following script:

```sh
pr=13208
sha=$(curl -s -H "Accept: application/vnd.github.v3+json" https://api.github.com/repos/kubernetes/kops/pulls/${pr} | jq -r .head.sha )
export KOPS_BASE_URL="https://storage.googleapis.com/k8s-staging-kops/pulls/pull-kops-aws-distro-debian12/pull-${sha}"
wget -q "$KOPS_BASE_URL/$(go env GOOS)/$(go env GOARCH)/kops"
chmod +x ./kops
./kops version
```
