# Kops HTTP API Server

### Building

To build the API container run the following

Note: This is a MAJOR clockwork, and a lot of assumptions and hardcoding exist here for now.

```bash
make uas-build
```

### Deploy to a cluster

From the kops directory run the following `helm` command

```bash
helm install charts/kops --namespace kops
```