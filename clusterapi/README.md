This is experimental integration with the cluster-api.  It is very much not production ready (and currently barely works).

We plug in our own bootstrap provider with the goal of enabling cluster-api nodes to join a kOps cluster.

# Create a cluster on GCP

*Note*: the name & zone matter, we need to match the values we'll create later in the CAPI resources.

```
go run ./cmd/kops create cluster clusterapi.k8s.local --zones us-east4-a
go run ./cmd/kops update cluster clusterapi.k8s.local --yes --admin
go run ./cmd/kops validate cluster --wait=10m
```

# Install cert-manager

```
kubectl apply --server-side -f https://github.com/cert-manager/cert-manager/releases/download/v1.18.2/cert-manager.yaml

kubectl wait --for=condition=Available --timeout=5m -n cert-manager deployment/cert-manager
kubectl wait --for=condition=Available --timeout=5m -n cert-manager deployment/cert-manager-cainjector
kubectl wait --for=condition=Available --timeout=5m -n cert-manager deployment/cert-manager-webhook
```

# Install CAPI and CAPG
```
REPO_ROOT=$(git rev-parse --show-toplevel)
kustomize build ${REPO_ROOT}/clusterapi/manifests/cluster-api | kubectl apply --server-side -f -
kustomize build ${REPO_ROOT}/clusterapi/manifests/cluster-api-provider-gcp | kubectl apply --server-side -f -
```

# Install our CRDs
```
kustomize build ${REPO_ROOT}/k8s | kubectl apply --server-side -f -
kustomize build ${REPO_ROOT}/clusterapi/config | kubectl apply --server-side -f -
```

## Create our Cluster object
```
go run ./cmd/kops get cluster clusterapi.k8s.local -oyaml | kubectl apply --server-side -n kube-system -f -
```

## Create our instancegroup object

```
go run ./cmd/kops get ig nodes-us-east4-a --name clusterapi.k8s.local -oyaml | kubectl apply --server-side -n kube-system -f -
```

# Remove any stuff left over from previous runs
```
kubectl delete machinedeployment --all
kubectl delete gcpmachinetemplate --all
```

```
# Very carefully create a MachineDeployment matching our configuration
cat ${REPO_ROOT}/clusterapi/examples/machinedeployment-direct.yaml | IMAGE_ID=projects/ubuntu-os-cloud/global/images/family/ubuntu-2404-lts-amd64 GCP_NODE_MACHINE_TYPE=e2-medium KUBERNETES_VERSION=v1.34.0 WORKER_MACHINE_COUNT=1 GCP_ZONE=us-east4-a GCP_SUBNET=us-east4-clusterapi-k8s-local CLUSTER_NAME=clusterapi.k8s.local CLUSTER_NAME_ESCAPED=clusterapi-k8s-local envsubst | kubectl apply --server-side -n kube-system -f -
```

# IMAGE_ID=projects/debian-cloud/global/images/family/debian-12 doesn't work with user-data (????)

# Run our controller, which populates the secret with the bootstrap script
```
go run .
```
