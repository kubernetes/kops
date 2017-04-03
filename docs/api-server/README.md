# Kops HTTP API Server

# Developing with the server in Kubernetes

Set your docker registry. For me this value is `krisnova` because of my [dockerhub registry](https://hub.docker.com/r/krisnova/kops-server).
It's important to note that the `kops` **Makefile** assumes a repository called `kops-server`.

```bash
export DOCKER_REGISTRY=krisnova
```

Build the kops API server container, and push the image up to your registry. The `helm` chart assumes `latest` as in `krisnova/kops-server:latest`.
The **Makefile** will handle this automatically.

```bash
make kops-server-push
```

This should compile the binary, and conveniently drop it off in your docker registry as a container image.
Once this done you should deploy the image to your cluster.

# Deploy the kops API server to a cluster

From the kops directory run the following `helm` command. More information on `helm` can be found [here](https://github.com/kubernetes/helm)

```bash
helm install charts/kops --namespace kops
```

For now (until we get something better) I usually just `helm ls` then `helm delete $name` and then run the `helm install` again. It's tedious but it works. Please help make this better!
