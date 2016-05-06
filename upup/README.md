## UpUp - CloudUp & NodeUp

CloudUp and NodeUp are two tools that are aiming to replace kube-up:
the easiest way to get a production Kubernetes up and running.

(Currently work in progress, but working.  Some of these statements are forward-looking.)

Some of the more interesting features:

* Written in go, so hopefully easier to maintain and extend, as complexity inevitably increases
* Uses a state-sync model, so we get things like a dry-run mode and idempotency automatically
* Based on a simple meta-model defined in a directory tree
* Can produce configurations in other formats (currently Terraform & Cloud-Init), so that we can have working
  configurations for other tools also.

## Bringing up a cluster

Set `YOUR_GCE_PROJECT`, then:

```
cd upup
make
${GOPATH}/bin/cloudup --v=0 --logtostderr -cloud=gce -zone=us-central1-f -project=$YOUR_GCE_PROJECT -name=kubernetes -kubernetes-version=1.2.2
```

If you have problems, please set `--v=8 --logtostderr` and open an issue, and ping justinsb on slack!

For now, we don't build a local kubectl file.  So just ssh to the master, and run kubectl from there:

```
gcloud compute ssh kubernetes-master
...
kubectl get nodes
kubectl get pods --all-namespaces
```

## Other interesting modes:

See changes that would be applied: `${GOPATH}/bin/cloudup --dryrun`

Build a terrform model: `${GOPATH}/bin/cloudup $NORMAL_ARGS --target=terraform > tf/k8s.tf.json`

# How it works

Everything is driven by a local configuration directory tree, called the "model".  The model represents
the desired state of the world.

Each file in the tree describes a Task.

On the nodeup side, Tasks can manage files, systemd services, packages etc.
On the cloudup side, Tasks manage cloud resources: instances, networks, disks etc.
