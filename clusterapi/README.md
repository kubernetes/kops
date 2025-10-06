This is experimental integration with the cluster-api.  It is very much not production ready (and currently barely works).

We plug in our own bootstrap provider with the goal of enabling cluster-api nodes to join a kOps cluster.

# Create a cluster on GCP

*Note*: the name & zone matter, we need to match the values we'll create later in the CAPI resources.

```
kops create cluster clusterapi.k8s.local --zones us-east4-a
kops update cluster clusterapi.k8s.local --yes --admin
kops validate cluster --wait=10m
```

#cd cluster-api-provider-gcp
#REGISTRY=${USER} make docker-build docker-push
#REGISTRY=${USER} make install-management-cluster # Doesn't yet exist in capg



# TODO: Install cert-manager

# Install CAPI and CAPG
```
REPO_ROOT=$(git rev-parse --show-toplevel)
kustomize build ${REPO_ROOT}/clusterapi/manifests/cluster-api | kubectl apply --server-side -f -
kustomize build ${REPO_ROOT}/clusterapi/manifests/cluster-api-provider-gcp | kubectl apply --server-side -f -
```

# Install our CRDs
```
kustomize build config | kubectl apply --server-side -f -
```

# Remove any stuff left over from previous runs
```
kubectl delete machinedeployment --all
kubectl delete gcpmachinetemplate --all
```

```
# Very carefully create a MachineDeployment matching our configuration
cat examples/manifest.yaml | IMAGE_ID=projects/ubuntu-os-cloud/global/images/family/ubuntu-2204-lts GCP_NODE_MACHINE_TYPE=e2-medium KUBERNETES_VERSION=v1.28.6 WORKER_MACHINE_COUNT=1  GCP_ZONE=us-east4-a GCP_REGION=us-east4 GCP_NETWORK_NAME=clusterapi-k8s-local GCP_SUBNET=us-east4-clusterapi-k8s-local GCP_PROJECT=$(gcloud config get project) CLUSTER_NAME=clusterapi-k8s-local envsubst | kubectl apply --server-side -n kube-system -f -
```

# IMAGE_ID=projects/debian-cloud/global/images/family/debian-12 doesn't work with user-data (????)

# Run our controller, which populates the secret with the bootstrap script
```
go run .
```
