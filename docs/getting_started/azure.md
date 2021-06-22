# Getting Started with kOps on Azure

Azure support on kOps is currently in alpha. The original issue
ticket is [#3957](https://github.com/kubernetes/kops/issues/3957).

Please see [#10412](https://github.com/kubernetes/kops/issues/10412)
for the remaining items and limitations. For example, Azure DNS is not
currently supported, and clusters need to be created with [Gossip
DNS](https://kops.sigs.k8s.io/gossip/).

# Create Creation Steps

## Step 1. Install Azure CLI

First, install Azure CLI.

```bash
$ curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash
```

Then type the following command to login to Azure. This will redirect
you to the browser login.

```bash
$ az login

...

You have logged in. Now let us find all the subscriptions to which you have access...
[
  {
	"cloudName": "AzureCloud",
	"homeTenantId": "76253...",
	"id": "7e232...",
	"isDefault": true,
	"managedByTenants": [],
	"name": "Your name...",
	"state": "Enabled",
	"tenantId": "76253...",
	"user": {
	  "name": "...",
	  "type": "user"
	}
  },
 ...
]
```

One Azure account has one or more than one “subscription”, which
serves as a single billing unit for Azure resources. Set the env var
`AZURE_SUBSCRIPTION_ID` to the ID of the subscription you want to
use.

```bash
$ export AZURE_SUBSCRIPTION_ID=7e232...
```

## Step 2. Create a Container in Azure Blob

Next, create an Azure Blob storage container for the kOps cluster store.

First, you need to create a resource group, which provides an isolated
namespace for resources.

```bash
$ az group create --name kops-test --location eastus
{
  "id": "/subscriptions/7e232.../resourceGroups/kops-test",
  "location": "eastus",
  "managedBy": null,
  "name": "kops-test",
  "properties": {
	"provisioningState": "Succeeded"
  },
  "tags": null,
  "type": "Microsoft.Resources/resourceGroups"
}
```

Then create a storage account for the resource group. The storage
account provides an isolated namespace for all storage resources. The
name must be unique across all Azure accounts.

```bash
$ az storage account create --name kopstest --resource-group kops-test
```

Set the env var `AZURE_STORAGE_ACCOUNT` to the storage account name for later use.

```bash
$ export AZURE_STORAGE_ACCOUNT=kopstest
```

Get an access key of the account and set it in env var `AZURE_STORAGE_KEY` for later use.

```bash
$ az storage account keys list --account-name kopstest
[
  {
	"keyName": "key1",
	"permissions": "Full",
	"value": "RHWWn..."
  },
  {
	"keyName": "key2",
	"permissions": "Full",
	"value": "..."
  }

]

$ export AZURE_STORAGE_KEY="RHWWn...“
```


Then create a blob container.

```bash
$ az storage container create --name cluster-configs
{
  "created": true
}
```

You can confirm that the container has been successfully created from
Storage Exporter or via `az storage container list`.

```bash
$ az storage container list --output table
Name             Lease Status    Last Modified
---------------  --------------  -------------------------
cluster-configs  unlocked        2020-10-06T21:12:36+00:00
```

Set the env var `KOPS_STATE_STORE` to the container name URL using kOps' `azureblob://` protocol.
The URL may include a path within the container.
kOps stores all of its cluster configuration within this path.

```bash
export KOPS_STATE_STORE=azureblob://cluster-configs
```

## Step 3. Set up Credentials for kOps

Use the following commands to generate kOps credentials.

First, create a service principal in Active Directory.

```bash
$ az ad sp create-for-rbac --name kops-test --role owner --sdk-auth

{
  "clientId": "8c6fddb5...",
  "clientSecret": "dUFzX1...",
  "subscriptionId": "7e232...",
  "tenantId": "76253...",
  ...
}
```

Set corresponding env vars:

- Set `AZURE_TENANT_ID` to the `tenantId` of the output
- Set `AZURE_CLIENT_ID` to the `clienteId` of the output
- Set `AZURE_CLIENT_SECRET` to the `clientSecret` of the output.

```bash
$ export AZURE_TENANT_ID="76253..."
$ export AZURE_CLIENT_ID="8c6fddb5..."
$ export AZURE_CLIENT_SECRET="dUFzX1..."
```

## Step 4. Run kOps Commands

Use the following command to create cluster configs in the blob container.
The command line flags prefixed with `--azure-` are for
Azure-specific configurations.

```bash
$ export KOPS_FEATURE_FLAGS=Azure

$ kops create cluster \
  --cloud azure \
  --name my-azure.k8s.local \
  --zones eastus-1 \
  --network-cidr 172.16.0.0/16 \
  --networking calico \
  --azure-subscription-id "${AZURE_SUBSCRIPTION_ID}" \
  --azure-tenant-id "${AZURE_TENANT_ID}" \
  --azure-resource-group-name kops-test \
  --azure-route-table-name kops-test \
  --azure-admin-user ubuntu
```

Confirm that config files are created in Blob storage.

```bash
$ az storage blob list --container-name cluster-configs --output table
```

Use the following command to preview the Azure resources
kOps will create for the k8s cluster.

```bash
$ kops update cluster  \
  --name my-azure.k8s.local
```

Now add the `--yes` flag to have kOps provision the resources
and create the cluster. This will also add a kubeconfig context
for the cluster.

```bash
$ kops update cluster  \
  --name my-azure.k8s.local \
  --yes
```

It may take a few minutes for the cluster's API server to become
reachable. Please run basic kubectl commands like `kubectl get
namespaces` to verify the API server is reachable.

Currently kOps creates the following resources in Azure:

- Virtual Machine Scale Sets (equivalent to AWS Auto Scaling Groups)
- Managed Disks (equivalent to AWS Elastic Volume Storage)
- Virtual network
- Subnet
- Route Table
- Role Assignment

By default, kOps create two VM Scale Sets - one for the k8s master and the
other for worker nodes. Managed Disks are used as etcd volumes ("main"
database and "event" database) and attached to the K8s master
VMs. Role assignments are needed to grant API access and Blob storage
access to the VMs.
