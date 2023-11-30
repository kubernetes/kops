# Bare Metal Support

***Bare metal support is experimental, and may be removed at any time***

## Introduction

kOps has some experimental bare-metal support, specifically for nodes.  The idea
we are exploring is that you can run your control-plane in a cloud, but you can
join physical machines as nodes to that control-plane, even though those nodes
are not located in the cloud.

This approach has some limitations and complexities - for example the
cloud-controller-manager for the control plane won't be able to attach volumes
to the nodes, because they aren't cloud VMs.  The advantage is that we can first
implement node bare-metal support, before tackling the complexities of the
control plane. 

## Walkthrough

Create a "normal" kOps cluster, but make sure the Metal feature-flag is set;
here we are using GCE:

```
export KOPS_FEATURE_FLAGS=Metal
kops create cluster foo.k8s.local --cloud gce --zones us-east4-a
kops update cluster --yes --admin foo.k8s.local

kops validate cluster --wait=10m
```

Create a kops-system namespace, to hold the host information that is generated
as part of joining the machine.  Although these are sensitive, they aren't
secrets, because they only hold public keys:

```
kubectl create ns kops-system
kubectl apply --server-side -f k8s/crds/kops.k8s.io_hosts.yaml
```

Create a ClusterRoleBinding and ClusterRole to allow kops-controller
to read the Host objects:

```
kubectl apply --server-side -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kops-controller:pki-verifier
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kops-controller:pki-verifier
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: User
  name: system:serviceaccount:kube-system:kops-controller
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kops-controller:pki-verifier
rules:
- apiGroups:
  - "kops.k8s.io"
  resources:
  - hosts
  verbs:
  - get
  - list
  - watch
EOF
```


### Create a VM

When first trying this out, we recommend creating a local VM instead of a true
bare-metal machine.

```
mkdir vm1
cd vm1
wget -O debian11.qcow2 https://cloud.debian.org/images/cloud/bullseye/20231013-1532/debian-11-nocloud-amd64-20231013-1532.qcow2

qemu-img create -o backing_file=debian11.qcow2,backing_fmt=qcow2 -f qcow2 vm1-root.qcow2 10G

qemu-system-x86_64 \
  -smp 2 \
  -enable-kvm \
  -netdev user,id=net0,net=192.168.76.0/24,dhcpstart=192.168.76.9,hostfwd=tcp::2222-:22 \
  -device rtl8139,netdev=net0 \
  -m 4G \
  -drive file=vm1-root.qcow2,if=virtio,format=qcow2 \
  -nographic -serial mon:stdio
```

Now login as root (with no password, and set up SSH and the machine name):

```
ssh-keygen -A
systemctl restart sshd
echo "vm1" > /etc/hostname
hostname vm1
```

Currently the `kops toolbox enroll` command only supports SSH agents for
the private key; so get your public key from `ssh-add -L`, and then you must
currently manually add it to the `authorized_keys` file on the VM.

```
mkdir ~/.ssh/
vim ~/.ssh/authorized_keys
```

After you've done this, open a new terminal and SSH should now work
from the host: `ssh  -p 2222 root@127.0.0.1 uptime`


### Joining the VM to the cluster

```
go run ./cmd/kops toolbox enroll --cluster foo.k8s.local --instance-group nodes-us-east4-a --ssh-user root --host 127.0.0.1 --ssh-port 2222
```

Within a minute or so, the node should appear in `kubectl get nodes`. 
If it doesn't work, first check the kops-configuration log:
`ssh root@127.0.0.1 -p 2222 journalctl -u kops-configuration`

And then if that looks OK (ends in "success"), check the kubelet log:
`ssh root@127.0.0.1 -p 2222 journalctl -u kubelet`.

### The state of the node

You should observe that the node is running, and pods are scheduled to the node.

```
kubectl get pods -A --field-selector spec.nodeName=vm1
```

Cilium will likely be running on the node.

The GCE PD CSI driver is scheduled, but is likely crash-looping
because it can't reach the GCE metadata service.  You can see this from the
logs on the VM in `/var/log/container`
(e.g. `ssh root@127.0.0.1 -p 2222 cat /var/log/containers/*gce-pd-driver*.log`)

If you try to use `kubectl logs`, you will see an error like the below, which
indicates another problem - that the control plane cannot reach the kubelet:
`Error from server: Get "https://192.168.76.9:10250/containerLogs/gce-pd-csi-driver/csi-gce-pd-node-l2rm8/csi-driver-registrar": dial tcp 192.168.76.9:10250: i/o timeout`

### Cleanup

Quit the qemu VM with Ctrl-a x.

Delete the node and the secret
```
kubectl delete node vm1
kubectl delete host -n kops-system vm1
```

If you're done with the cluster also:
```
kops delete cluster foo.k8s.local --yes
```
