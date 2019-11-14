# Authentication

Kops has support for configuring authentication systems.  This should not be used with kubernetes versions
before 1.8.5 because of a serious bug with apimachinery [#55022](https://github.com/kubernetes/kubernetes/issues/55022).

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
apiVersion: kops.k8s.io/v1alpha2
kind: Cluster
metadata:
  name: cluster.example.com
spec:
  authentication:
    kopeio: {}
  authorization:
    rbac: {}
```

## AWS IAM Authenticator


:exclamation:AWS IAM Authenticator requires Kops 1.10 or newer and Kubernetes 1.10 or newer


To turn on AWS IAM Authenticator, you'll need to add the stanza bellow
to your cluster configuration.

```
authentication:
  aws: {}
```

For example:

```
apiVersion: kops.k8s.io/v1alpha2
kind: Cluster
metadata:
  name: cluster.example.com
spec:
  authentication:
    aws: {}
  authorization:
    rbac: {}
```

The creation of a AWS IAM authenticator config as a ConfigMap is also required.
For more details on AWS IAM authenticator please visit [kubernetes-sigs/aws-iam-authenticator](https://github.com/kubernetes-sigs/aws-iam-authenticator)

Example config:

```
---
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: kube-system
  name: aws-iam-authenticator
  labels:
    k8s-app: aws-iam-authenticator
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

### Creating a new cluster with IAM Authenticator on.

* Create a cluster following the [AWS getting started guide](https://github.com/kubernetes/kops/blob/master/docs/getting_started/aws.md)
* When you reach the "Customize Cluster Configuration" section of the guide modify the cluster spec and add the Authentication and Authorization configs to the YAML config.
* Continue following the cluster creation guide to build the cluster.
    * :warning: When the cluster first comes up the aws-iam-authenticator PODs will be in a bad state.
as it is trying to find the aws-iam-authenticator ConfigMap and we have not yet created it.
* Once the cluster is up, you'll need to create an aws-iam-authenticator configMap on the cluster `kubectl apply -f aws-iam-authenticator_example-config.yaml`
* Once the configuration is created you need to delete the initially created aws-iam-authenticator PODs, this will force new ones to come and correctly find the ConfigMap.
```
kubectl get pods -n kube-system | grep aws-iam-authenticator | awk '{print $1}' | xargs kubectl delete pod -n kube-system
```

### Turning on IAM Authenticator on an existing cluster.

* Create an aws-iam-authenticator configMap on the cluster `kubectl apply -f aws-iam-authenticator_example-config.yaml`
* Edit the clusters configuration `kops edit cluster ${NAME}` and add the Authentication and Authorization configs to the YAML config.
* Update the clusters configuration `kops update cluster ${CLUSTER_NAME} --yes`
* Temporarily disable aws-iam-authenticator DaemonSet `kubectl patch daemonset -n kube-system aws-iam-authenticator -p '{"spec": {"template": {"spec": {"nodeSelector": {"disable-aws-iam-authenticator": "true"}}}}}'`
* Perform a rolling update of the masters `kops rolling-update cluster ${CLUSTER_NAME} --instance-group-roles=Master --force --yes`
* Re-enable aws-iam-authenticator DaemonSet `kubectl patch daemonset -n kube-system aws-iam-authenticator --type json -p='[{"op": "remove", "path": "/spec/template/spec/nodeSelector/disable-aws-iam-authenticator"}]'` 
