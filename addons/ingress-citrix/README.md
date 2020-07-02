# Deploying Citrix Ingress Controller through KOPS

This guide explains how to deploy [Citrix Ingress Controller](https://github.com/citrix/citrix-k8s-ingress-controller) through KOPS addon.

## Quick Deploy using `kops`

You can enable the Citrix Ingress Controller addon when creating the Kubernetes cluster through KOPS.

Edit the cluster before creating it

```
kops edit cluster <cluster-name>
```

Now add the addon specification in the cluster manifest in the section - `spec.addons`

```
addons:
  - manifest: ingress-citrix

```
For more information on how to enable addon during cluster creation refer [Kops Addon guide](https://github.com/kubernetes/kops/blob/master/docs/operations/addons.md#installing-kubernetes-addons)

## Quick Deploy using `kubectl`

After cluster creation, you can deploy [Citrix Ingress Controller](https://github.com/citrix/citrix-k8s-ingress-controller) using the below command

```
kubectl create secret generic nslogin --from-literal=username='nsroot' --from-literal=password=<password>
kubectl create -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/ingress-citrix/v1.1.1.yaml
```
