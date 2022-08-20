# Troubleshooting kOps clusters

The first step to debugging a kOps cluster is to run `kops validate cluster --name <clustername> --wait 10m`. If the cluster has not validated by then, something is wrong.

# The Control Plane

If the above-mentioned command complains about an unavailable API server, it means the control plane isn't working properly.
In order to diagnose further, you need to log into one of the control plane nodes.

Run `kops get instances` (1.19+) or look in the AWS console to identify a node with the master role. Then ssh into the IP address listed.

The logs on the control plane resides in`/var/log`. Assume the logs are there unless otherwise noted.

## Nodeup

Nodeup is the process responsible for the initial provisioning of a node. It is a oneshot systemd service called `kops-configuration.service`. You can see the logs for this service running `journalctl -u kops-configuration.service`.

If it succeed, you should be able to see the following log entries:

```
nodeup[X]: success
systemd[1]: kops-configuration.service: Succeeded.
systemd[1]: Finished Run kops bootstrap (nodeup).
```

Note that if the node booted some time ago, the logs for this unit may be empty.

If the nodeup either exists with an error or keeps looping through a task that cannot continue, the cluster has most likely been misconfigured. Hopefully the error messages gives enough for further investigation.

Either way, we would appreciate a GitHub issue as we try to avoid clusters running into problems during the nodeup process.

## API Server

If nodeup succeeds, the core kube containers should have started. Look for the API server logs in `kube-apiserver.log`. 

Often the issue is obvious such as passing incorrect CLI flags.

### API Server hangs after etcd restore

After resizing an etcd cluster or restoring backup, the kubernetes API can contain too many endpoints.
You can confirm this by running `kubectl get endpoints -n default kubernetes`. This command should list exactly as many IPs as you have control plane nodes.

Check the [backup and restore documentation](etcd_backup_restore_encryption.md) for more details about this problem.

## etcd

The API server makes use of two etcd servers, main and events.

One of the more common reasons for the API server not working properly is that etcd is unavailable. If you see connection errors to port 4001 or 4002, it means that main and/or events respectively is unavailable.

The etcd clusters are managed by etcd-manager and most likely it is something wrong with the manager rather than etcd in itself. The logs for etcd is passed through etcd-manager, so you will be able to find the logs for both in `etcd.log` and `etcd-events.log`. Since both etcd-manager and etcd are quorum-based clusters there can be some misleading errors in these files that may suggest that etcd is broken, when in fact it is etcd-manager that is.

# DNS

Troubleshooting Kubernetes DNS is perhaps worth a whole book. The Kubernetes docs have [a fairly good writeup](https://kubernetes.io/docs/tasks/administer-cluster/dns-debugging-resolution/) on how to debug DNS.

It is worth mentioning that failing DNS is often a *symptom* of a broken pod network. So you may want to ensure that two pods can talk to each other using IP addresses before starting to troubleshoot DNS.

# CNI

## missing files in `/opt/cni/bin`

### empty directory

If the CNI bin directory is completely empty it may be a symptom of nodeup not working properly. See more on troubleshooting nodeup above. In most cases, nodeup will write the most common CNI plugins to that directory so it should rarely be completely empty.

### CNI plugin file missing

If the directory is there, but the CNI plugin and configuration is missing, it means that the process responsible for writing these files are not working properly. In most cases this is a `DaemonSet` running in `kube-system`.

At this point it is worth repeating that the control plane _will work_ without CNI. Most control plane nodes do not use the pod network but communicates using the host's network. If you cannot talk to the API server, e.g running `kubectl get nodes`, the problem is not CNI.

If the API is working, and the CNI is installed through a `DaemonSet`, check that the pods are running. If pods are expected, but absent, it may be an issue with installing the CNI addon. kOps will try to install addons regularly, so run `journalctl -f` on a control plane node to spot any errors.
