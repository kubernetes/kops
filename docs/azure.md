# Kubernetes on Azure with kops

## Configuring your workstation to use azure

#### Storage account

First set up the kops state store using Azure storage

```
export AZURE_STORAGE_ACCOUNT=kopsdevel
export AZURE_STORAGE_ACCESS_KEY=123ABC456DEF
export KOPS_STATE_STORE="https://kopsdevel.blob.core.windows.net"
```

