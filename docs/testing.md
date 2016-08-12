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