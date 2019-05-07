## Checklist for a new kubernetes version

### Update bazel rules_go / rules_docker / gazelle etc

### Run gazelle

```bash
make gazelle
```

### Check go version used in k/k

### Update base images

### Update dependencies (apimachinery etc)

This is by far the most painful bit of the process.  First you have to persuade dep to update dependencies, and then you'll have to make code changes.

For dep, you'll probably have to remove some imports to any packages that have
been removed, otherwise dep will ignore your Gopkg.toml.  The path forward here
is to use vgo (also known as go), which has a much better model.

You'll then have to fix any changed code.  This is gradually getting better, but will be better if we:

* Stop using apimachinery / codegen and switch to CRDs / cluster-api
* Stop vendoring functionality from kubernetes/kubernetes - this is also gradually getting better.


### Update docker version installed by default
### Check CNI version

Sources:
*  [kube-up](https://github.com/kubernetes/kubernetes/blob/master/cluster/gce/gci/configure.sh#L27)

### Check admission plugins

Sources:
* https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#is-there-a-recommended-set-of-admission-controllers-to-use
* https://github.com/kubernetes/kubernetes/blob/master/cluster/gce/config-default.sh)

### Check for new deprecated flags

Review the e2e test output, looking for the artifacts from kube-apiserver, kubelet, kube-scheduler etc

e.g. Flag --address has been deprecated, see --insecure-bind-address instead.
Flag --insecure-port has been deprecated, This flag will be removed in a future version.

### Check for major new features (that are in beta or GA, not alpha)

### Check for new aws-sdk-go library (if we want to go newer than k8s)

