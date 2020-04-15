# Etcd Administration Tasks

## etcd-manager

etcd-manager is a kubernetes-associated project that kops uses to manage
etcd.

etcd-manager uses many of the same ideas as the existing etcd implementation
built into kops, but it addresses some limitations also:

* separate from kops - can be used by other projects
* allows etcd2 -> etcd3 upgrade (along with minor upgrades)
* allows cluster resizing (e.g. going from 1 to 3 nodes)

When using kubernetes >= 1.12 etcd-manager will be used by default. See [../etcd3-migration.md] for upgrades from older clusters.

## Backups

Backups and restores of etcd on kops are covered in [etcd_backup_restore_encryption.md](etcd_backup_restore_encryption.md)

## Direct Data Access

It's not typically necessary to view or manipulate the data inside of etcd directly with etcdctl, because all operations usually go through kubectl commands. However, it can be informative during troubleshooting, or just to understand kubernetes better. Here are the steps to accomplish that on kops.

1\. Connect to an etcd-manager pod

```bash
CONTAINER=$(kubectl get pods -n kube-system | grep etcd-manager-main | head -n 1 | awk '{print $1}')
kubectl exec -it -n kube-system $CONTAINER bash
```

2\. Determine which version of etcd is running

```bash
DIRNAME=$(ps -ef | grep --color=never /opt/etcd | head -n 1 | awk '{print $8}' | xargs dirname)
echo $DIRNAME
```

3\. Run etcdctl

```bash
ETCDCTL_API=3 $DIRNAME/etcdctl --cacert=/rootfs/etc/kubernetes/pki/kube-apiserver/etcd-ca.crt --cert=/rootfs/etc/kubernetes/pki/kube-apiserver/etcd-client.crt --key=/rootfs/etc/kubernetes/pki/kube-apiserver/etcd-client.key --endpoints=https://127.0.0.1:4001 get --prefix / | tee output.txt
```

The contents of etcd are now in output.txt. 

You may run any other etcdctl commands by replacing the "get --prefix /" with a different command.

The contents of the etcd dump are often garbled. See the next section for a better way to view the results.

## Dump etcd contents in clear text

Openshift's etcdhelper is a good way of exporting the contents of etcd in a readable format. Here are the steps.

1\. SSH into a master node

You can view the IP addresses of the nodes

```
kubectl get nodes -o wide
```

and then

```
ssh admin@<IP-of-master-node>
```

2\. Install golang

in whatever manner you prefer. Here is one example.

```
cd /usr/local
sudo wget https://dl.google.com/go/go1.13.3.linux-amd64.tar.gz
sudo tar -xvf go1.13.3.linux-amd64.tar.gz
cat <<EOT >> $HOME/.profile
export GOROOT=/usr/local/go
export GOPATH=\$HOME/go
export PATH=\$GOPATH/bin:\$GOROOT/bin:\$PATH
EOT
source $HOME/.profile
which go
```

3\. Install etcdhelper

```
mkdir -p ~/go/src/github.com/
cd ~/go/src/github.com/
git clone https://github.com/openshift/origin openshift
cd openshift/tools/etcdhelper
go build .
sudo cp etcdhelper /usr/local/bin/etcdhelper
which etcdhelper
```

4\. Run etcdhelper

```
sudo etcdhelper -key /etc/kubernetes/pki/kube-apiserver/etcd-client.key -cert /etc/kubernetes/pki/kube-apiserver/etcd-client.crt  -cacert /etc/kubernetes/pki/kube-apiserver/etcd-ca.crt -endpoint https://127.0.0.1:4001 dump | tee output.txt
```

The output of the command is now available in output.txt

Other etcdhelper commands are possible, like "ls":

```
sudo etcdhelper -key /etc/kubernetes/pki/kube-apiserver/etcd-client.key -cert /etc/kubernetes/pki/kube-apiserver/etcd-client.crt  -cacert /etc/kubernetes/pki/kube-apiserver/etcd-ca.crt -endpoint https://127.0.0.1:4001 ls
```