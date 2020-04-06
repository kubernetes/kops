## Testing tips

If you are running kops as part of an e2e test, the following tips may be useful.

### CI Kubernetes Build

Set the KubernetesVersion to a `http://` or `https://` base url, such as `https://storage.googleapis.com/kubernetes-release-dev/ci/v1.4.0-alpha.2.677+ea69570f61af8e/`

We expect the base url to have `bin/linux/amd64` directory containing:

* kubelet
* kubelet.sha1
* kubectl
* kubectl.sha1
* kube-apiserver.docker_tag
* kube-apiserver.tar
* kube-apiserver.tar.sha1
* kube-controller-manager.docker_tag
* kube-controller-manager.tar
* kube-controller-manager.tar.sha1
* kube-proxy.docker_tag
* kube-proxy.tar
* kube-proxy.tar.sha1
* kube-scheduler.docker_tag
* kube-scheduler.tar
* kube-scheduler.tar.sha1


Do this with `kops edit cluster <clustername>`.  The spec should look like

```
...
spec:
  kubernetesVersion: "https://storage.googleapis.com/kubernetes-release-dev/ci/v1.4.0-alpha.2.677+ea69570f61af8e/"
  cloudProvider: aws
  etcdClusters:
  - etcdMembers:
    - name: us-east-1c
      zone: us-east-1c
    name: main
...
```


### Running the kubernetes e2e test suite

The simple way:

```
# cd wherever you tend to put git repos
git clone https://github.com/kubernetes/test-infra.git
export KOPS_E2E_STATE_STORE=s3://your-kops-state-store # Change to your state store path
export KOPS_E2E_CLUSTER_NAME=e2e.cluster.name          # Change to an FQDN for your e2e cluster name
test-infra/jobs/ci-kubernetes-e2e-kops-aws.sh |& tee /tmp/testlog
```

This:

* Brings up a cluster using the latest `kops` build from `master` (see below for how to use your current build)
* Runs the default series of tests (which the Kubernetes team is [also
running here](https://k8s-testgrid.appspot.com/google-aws#kops-aws)) (see below for how to override the test list)
* Tears down the cluster
* Pipes all output to `/tmp/testlog`

(**Note**: By default this script assumes that your AWS credentials are in
`~/.aws/credentials`, and the SSH keypair you want to use is
`~/.ssh/kube_aws_rsa`. You can override `JENKINS_AWS_CREDENTIALS_FILE`,
`JENKINS_AWS_SSH_PRIVATE_KEY_FILE` and `JENKINS_AWS_SSH_PUBLIC_KEY_FILE` if you
want to change this.)

This isn't yet terribly useful, though - it just shows how to replicate the
existing job, but not with your custom code. To test a custom `kops` build, you
can do the following:

To use S3:
```
# cd to your kops repo
export S3_BUCKET_NAME=kops-dev-${USER}
make kops-install dev-upload UPLOAD_DEST=s3://${S3_BUCKET_NAME}

KOPS_VERSION=`bazel run //cmd/kops version -- --short`
export KOPS_BASE_URL=https://${S3_BUCKET_NAME}.s3.amazonaws.com/kops/${KOPS_VERSION}/
```

To use GCS:
```
export GCS_BUCKET_NAME=kops-dev-${USER}
make kops-install dev-upload UPLOAD_DEST=gs://${GCS_BUCKET_NAME}

KOPS_VERSION=`bazel run //cmd/kops version -- --short`
export KOPS_BASE_URL=https://${GCS_BUCKET_NAME}.storage.googleapis.com/kops/${KOPS_VERSION}/
```

Whether using GCS or S3, you probably want to upload dns-controller &
kops-contoller images if you have changed them:

For dns-controller:

```bash
KOPS_VERSION=`bazel run //cmd/kops version -- --short`
export DOCKER_IMAGE_PREFIX=${USER}/
export DOCKER_REGISTRY=
make dns-controller-push
export DNSCONTROLLER_IMAGE=${DOCKER_IMAGE_PREFIX}dns-controller:${KOPS_VERSION}
```

For kops-controller:

```bash
KOPS_VERSION=`bazel run //cmd/kops version -- --short`
export DOCKER_IMAGE_PREFIX=${USER}/
export DOCKER_REGISTRY=
make kops-controller-push
export KOPSCONTROLLER_IMAGE=${DOCKER_IMAGE_PREFIX}kops-controller:${KOPS_VERSION}
```

You can create a cluster using `kops create cluster <clustername> --zones us-east-1b`

Then follow the test directions above.

To override the test list for the job, you need to familiar with the
`ginkgo.focus` and `ginkgo.skip`
flags. Using these flags, you can do:

```
export GINKGO_TEST_ARGS="--ginkgo.focus=\[Feature:Performance\]"
```

and follow the instructions above. [Here are some other examples from the `e2e.go` documentation.](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-testing/e2e-tests.md).

If you want to test against an existing cluster, you can do:

```
export E2E_UP=false; export E2E_DOWN=false
```

and follow the instructions above. This is particularly useful for testing the
myriad of `kops` configuration/topology options without having to modify the
testing infrastructure. *Note:* This is also the only way currently to test a
custom Kubernetes build
(see
[kubernetes/test-infra#1454](https://github.com/kubernetes/test-infra/issues/1454)).


### Uploading a custom build

If you want to upload a custom Kubernetes build, here is a simple way (note:
this assumes you've run `make quick-release` in the Kubernetes repo first):


```
# cd wherever you tend to put git repos
git clone https://github.com/kubernetes/release.git

# cd back to your kubernetes repo
/path/to/release/push-build.sh # Fix /path/to/release with wherever you cloned the release repo
```

That will upload the release to a GCS bucket and make it public. You can then
use the outputted URL in `kops` with `--kubernetes-version`.

If you need it private in S3, here's a manual way:

```
make quick-release
cd ./_output/release-tars/
# ??? rm -rf kubernetes/
tar zxf kubernetes-server-linux-amd64.tar.gz

rm kubernetes/server/bin/federation*
rm kubernetes/server/bin/hyperkube
rm kubernetes/server/bin/kubeadm
rm kubernetes/server/bin/kube-apiserver
rm kubernetes/server/bin/kube-controller-manager
rm kubernetes/server/bin/kube-discovery
rm kubernetes/server/bin/kube-dns
rm kubernetes/server/bin/kubemark
rm kubernetes/server/bin/kube-proxy
rm kubernetes/server/bin/kube-scheduler
rm kubernetes/kubernetes-src.tar.gz


find kubernetes/server/bin -type f -name "*.tar" | xargs -I {} /bin/bash -c "sha1sum {} | cut -f1 -d ' ' > {}.sha1"
find kubernetes/server/bin -type f -name "kube???" | xargs -I {} /bin/bash -c "sha1sum {} | cut -f1 -d ' ' > {}.sha1"

aws s3 sync  --acl public-read kubernetes/server/bin/ s3://${S3_BUCKET_NAME}/kubernetes/dev/v1.6.0-dev/bin/linux/amd64/
```

### Example e2e command

```
go run hack/e2e.go -v -up -down -kops `which kops` -kops-cluster test.test-aws.k8s.io -kops-state s3://k8s-kops-state-store/ -kops-nodes=
4 -deployment kops --kops-kubernetes-version https://storage.googleapis.com/kubernetes-release-dev/ci/$(curl  -SsL https://storage.googleapis.com/kubernetes-release-dev/ci/latest-green.txt)
```

(note the `v1.6.0-dev`: we insert a kubernetes version so that kops can
automatically detect which k8s version is in use, which it uses to control
flags that are not compatible between versions)

Then:

* `kops create cluster ... --kubernetes-version https://${S3_BUCKET_NAME}.s3.amazonaws.com/kubernetes/dev/v1.6.0-dev/`

* for an existing cluster: `kops edit cluster` and set `KubernetesVersion` to `https://${S3_BUCKET_NAME}.s3.amazonaws.com/kubernetes/dev/v1.6.0-dev/`
