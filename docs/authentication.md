# Authentication

Kops has support for configuring authentication systems.  This support is
currently highly experimental, and should not be used with kubernetes versions
before 1.8.5 because of a serious bug with apimachinery (#55022)[https://github.com/kubernetes/kubernetes/issues/55022].

## kopeio authentication

If you want to experiment with kopeio authentication, you can use
`--authentication kopeio`.  However please be aware that kopeio authentication
has not yet been formally released, and thus there is not a lot of upstream
documentation.

Alternatively, you can add this block to your cluster:

```
authentication:
  kopeio: {}
```

For example:

```
apiVersion: kops/v1alpha2
kind: Cluster
metadata:
  name: cluster.example.com
spec:
  authentication:
    kopeio: {}
  authorization:
    rbac: {}
```

