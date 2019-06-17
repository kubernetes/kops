# Kops HTTP API Server

# Building the kops API server

Set your docker registry

```bash
cd ${GOPATH}/src/k8s.io/kops/
export DOCKER_REGISTRY=$registry
```

Build the kops API server container, and push the image up to your registry.

```bash
make kops-server-push
```

# Deploying the kops API server to a cluster

From the kops directory run the following `helm` command. More information on `helm` can be found [here](https://github.com/kubernetes/helm)

```bash
helm install charts/kops --namespace kops
```