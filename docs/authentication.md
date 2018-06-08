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

## Heptio Authenticator for AWS

If you want to turn on Heptio Authenticator for AWS, you can add this block 
to your cluster:

```
authentication:
  heptio: {}
```

For example:

```
apiVersion: kops/v1alpha2
kind: Cluster
metadata:
  name: cluster.example.com
spec:
  authentication:
    heptio: {}
  authorization:
    rbac: {}
```

Once the cluster is up you will need to create the heptio authenticator
config as a config map. (This can also be done when boostrapping a cluster using addons)
For more details on heptio authenticator please visit (heptio/authenticator)[https://github.com/heptio/authenticator]
Example config:

```
---
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: kube-system
  name: heptio-authenticator-aws
  labels:
    k8s-app: heptio-authenticator-aws
data:
  config.yaml: |
    # a unique-per-cluster identifier to prevent replay attacks
    # (good choices are a random token or a domain name that will be unique to your cluster)
    clusterID: my-dev-cluster.example.com
    server:
      # each mapRoles entry maps an IAM role to a username and set of groups
      # Each username and group can optionally contain template parameters:
      #  1) "{{AccountID}}" is the 12 digit AWS ID.
      #  2) "{{SessionName}}" is the role session name.
      mapRoles:
      # statically map arn:aws:iam::000000000000:role/KubernetesAdmin to a cluster admin
      - roleARN: arn:aws:iam::000000000000:role/KubernetesAdmin
        username: kubernetes-admin
        groups:
        - system:masters
      # map EC2 instances in my "KubernetesNode" role to users like
      # "aws:000000000000:instance:i-0123456789abcdef0". Only use this if you
      # trust that the role can only be assumed by EC2 instances. If an IAM user
      # can assume this role directly (with sts:AssumeRole) they can control
      # SessionName.
      - roleARN: arn:aws:iam::000000000000:role/KubernetesNode
        username: aws:{{AccountID}}:instance:{{SessionName}}
        groups:
        - system:bootstrappers
        - aws:instances
      # map federated users in my "KubernetesAdmin" role to users like
      # "admin:alice-example.com". The SessionName is an arbitrary role name
      # like an e-mail address passed by the identity provider. Note that if this
      # role is assumed directly by an IAM User (not via federation), the user
      # can control the SessionName.
      - roleARN: arn:aws:iam::000000000000:role/KubernetesAdmin
        username: admin:{{SessionName}}
        groups:
        - system:masters
      # each mapUsers entry maps an IAM role to a static username and set of groups
      mapUsers:
      # map user IAM user Alice in 000000000000 to user "alice" in "system:masters"
      - userARN: arn:aws:iam::000000000000:user/Alice
        username: alice
        groups:
        - system:masters
```