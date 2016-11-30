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

The [e2e](../e2e/README.md) directory has a docker image and some scripts which make it easy to run
the kubernetes e2e tests, using kops.

### Uploading a custom build

If you want to upload a custom build, here is one way to do so:

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

aws s3 sync  --acl public-read  kubernetes/server/bin/ s3://${S3_BUCKET_NAME}/kubernetes/dev/v1.6.0-dev/bin/linux/amd64/
```

(note the `v1.6.0-dev`: we insert a kubernetes version so that kops can
automatically detect which k8s version is in use, which it uses to control
flags that are not compatible between versions)

Then:

* `kops create cluster ... --kubernetes-version https://${S3_BUCKET_NAME}.s3.amazonaws.com/kubernetes/dev/v1.6.0-dev/`

* for an existing cluster: `kops edit cluster` and set `KubernetesVersion` to `https://${S3_BUCKET_NAME}.s3.amazonaws.com/kubernetes/dev/v1.6.0-dev/`
