# Ambassador

The [Ambassador API Gateway](https://getambassador.io/) provides all the functionality of a traditional ingress
controller (i.e., path-based routing) while exposing many additional capabilities such as authentication, URL rewriting,
CORS, rate limiting, and automatic metrics collection.

## Ambassador Addon

[Ambassador Operator](https://github.com/datawire/ambassador-operator) is a Kubernetes Operator that controls the
complete lifecycle of Ambassador in your cluster. It also automates many of the repeatable tasks you have to perform for
Ambassador. Once installed, the Operator will automatically complete rapid installations and seamless upgrades to new
versions of Ambassador.

This addon deploys Ambassador Operator which installs Ambassador in a kops cluster.

##### Note:
The operator requires widely scoped permissions in order to install and manage Ambassador's lifecycle. Both, the
operator and Ambassador, are deployed in the `ambassador` namespace. You can review the permissions granted to the
operator [here](https://github.com/kubernetes/kops/blob/master/addons/ambassador/ambassador-operator.yaml).

### Usage

#### As a kops addon

To deploy the addon, run the following before creating a cluster -
```console
kops edit cluster <cluster-name>
```

Now add the addon specification in the cluster manifest in the section - `spec.addons`

```
addons:
- manifest: ambassador
```

##### Note:

If you've already created the cluster, you'll have to run -
```console
kops update cluster <cluster-name> --yes
```
followed by -
```console
kops rolling-update cluster --yes
```
to install the addon.

For more information on how to enable addon during cluster creation refer [Kops Addon guide](https://github.com/kubernetes/kops/blob/master/docs/operations/addons.md#installing-kubernetes-addons).

#### Deploying using `kubectl`

After cluster creation, you can deploy Ambassador using the following command -

```console
kubectl create -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/ambassador/ambassador-operator.yaml
```
