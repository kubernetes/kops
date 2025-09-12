# Getting Started with kOps on Azure

**WARNING**: Azure support on kOps is currently in **alpha**, which means that it is in the early stages of development and subject to change, please use with caution.

## Features

* Create, update and delete clusters
* Create, edit and delete instance groups
* ...

## Requirements

* Latest kOps version installed
* kubectl installed
* Azure CLI installed
* Azure account with **Contributor** permissions for the cluster subscription and an existing storage account
* SSH key, both `id_ed25519` and `id_rsa` keys are supported

## Environment Variables

### Enable Azure

Since Azure support is currently in **alpha**, it is feature gated and you will need to set:

```bash
export KOPS_FEATURE_FLAGS="Azure"
```
### Azure-specific

```bash
export AZURE_SUBSCRIPTION_ID=<subscription-id>
export AZURE_STORAGE_ACCOUNT=<storage-account-name>
```

### kOps-specific

```bash
export KOPS_STATE_STORE=azureblob://<container-name>
```

## Creating a Single Master Cluster

```bash
# Create a cluster in zone northeurope-1
kops create cluster --cloud azure --name my.k8s --zones northeurope-1 --azure-admin-user ubuntu --yes
# Validate the cluster
kops validate cluster --name my.k8s --wait=10m
# Export the kubeconfig file with the cluster admin user (make sure you keep this user safe!)
kops export kubeconfig --name my.k8s --admin
```

## Updating a Cluster

```bash
# Edit the cluster configuration
kops edit cluster --name my.k8s 
# Edit the nodes instance group configuration
kops edit ig --name my.k8s nodes 
# Preview the changes to be applied to the cluster
kops update cluster --name my.k8s 
# Apply the changes to the cluster
kops update cluster --name my.k8s --yes 
# Preview the node that need to be updated
kops rolling-update cluster --name my.k8s 
# Replace the nodes that need to be updated
kops rolling-update cluster --name my.k8s --yes 
```

## Deleting a Cluster

```bash
# Preview the resources to be deleted
kops delete cluster --name my.k8s 
# Delete all the cluster resources
kops delete cluster --name my.k8s --yes 
```

## TODO

kOps for Azure currently does not support the following features:

* Azure AD Workload Identity
* Azure Disk volumes
* Azure Load Balancer
* Autoscaling (using Cluster Autoscaler or Karpenter)
* Terraform support
* Multi-master clusters
* ...

## Next steps

Now that you have a working kOps cluster, read through the recommendations for [production setups guide](production.md) to learn more about how to configure kOps for production workloads.
