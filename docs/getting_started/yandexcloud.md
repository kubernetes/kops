# Getting Started with kOps on Yandex Cloud

**WARNING**: Yandex Cloud support on kOps is currently in early **alpha**, meaning it is subject to change, so please use with caution.

## Requirements
* Yandex Cloud [account](https://cloud.yandex.com/en/)
* Yandex Cloud [cli](https://cloud.yandex.com/en/docs/cli/operations/install-cli)
* Yandex Cloud [service IAM key in JSON format](https://cloud.yandex.com/en/docs/cli/cli-ref/managed-services/iam/key/create)
* Yandex Cloud [IAM access-key](https://cloud.yandex.com/en/docs/iam/operations/sa/create-access-key) to access Yandex Object Storage via S3(https://cloud.yandex.com/en/docs/storage/tools/s3cmd).
* SSH public and private keys

## Environment Variables

It is important to set the following environment variables:
```bash
export KOPS_FEATURE_FLAGS=Yandex
export YANDEX_CREDENTIAL_FILE=<path to iam json file>
export S3_ENDPOINT=https://storage.yandexcloud.net
export S3_ACCESS_KEY_ID=<access-key>
export S3_SECRET_ACCESS_KEY=<secret-key>
export KOPS_STATE_STORE=s3://<bucket-name>
```

Create a bucket and make sure that it uses a KMS secret key to encrypt data.

## Creating a Simple Cluster with three master nodes and one worker node

In the following examples, `example.k8s.local` is a [gossip-based DNS ](../gossip.md) cluster name.

 * Save that configuration to a file (like example.yaml).
 * Update spec.project with your Yandex Folder Id
 * Set spec.configBase to your Yandex Object Storage S3 bucket
 * Update spec.assets if you use own assets or remove it.
 * Update spec.cloudConfig.gceServiceAccount with your Service Account Id
```yaml
apiVersion: kops.k8s.io/v1alpha2
kind: Cluster
metadata:
  creationTimestamp: "2017-01-01T00:00:00Z"
  name: example.k8s.local
spec:
  assets:
    fileRepository:
      https://storage.yandexcloud.net/<bucket-with-assets>/
  api:
    loadBalancer:
      type: Public
  authorization:
    rbac: {}
  channel: stable
  cloudConfig:
    gceServiceAccount: <service-account-id>
  cloudProvider: yandex
  configBase: s3://<bucket-name>
  etcdClusters:
  - cpuRequest: 200m
    etcdMembers:
    - instanceGroup: master-ru-central1-a
      name: a
      volumeSize: 1
    - instanceGroup: master-ru-central1-b
      name: b
      volumeSize: 1
    - instanceGroup: master-ru-central1-c
      name: c
      volumeSize: 1
    memoryRequest: 100Mi
    name: main
  - cpuRequest: 100m
    etcdMembers:
    - instanceGroup: master-ru-central1-a
      name: a
      volumeSize: 1
    - instanceGroup: master-ru-central1-b
      name: b
      volumeSize: 1
    - instanceGroup: master-ru-central1-c
      name: c
      volumeSize: 1
    memoryRequest: 100Mi
    name: events
  iam:
    allowContainerRegistry: true
    legacy: false
  kubelet:
    anonymousAuth: false
  kubernetesApiAccess:
  - 0.0.0.0/0
  - ::/0
  kubernetesVersion: v1.21.0
  masterPublicName: api.example.k8s.local
  networkCIDR: 172.20.0.0/16
  networking:
    calico: {}
  nonMasqueradeCIDR: 100.64.0.0/10
  project: <folder-id>
  sshAccess:
  - 0.0.0.0/0
  - ::/0
  subnets:
  - cidr: 172.20.32.0/19
    name: ru-central1-a
    type: Public
    zone: ru-central1-a
  - cidr: 172.20.64.0/19
    name: ru-central1-b
    type: Public
    zone: ru-central1-b
  - cidr: 172.20.96.0/19
    name: ru-central1-c
    type: Public
    zone: ru-central1-c
  topology:
    dns:
      type: Public
    masters: public
    nodes: public

---

apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  creationTimestamp: "2017-01-01T00:00:00Z"
  labels:
    kops.k8s.io/cluster: example.k8s.local
  name: master-ru-central1-a
spec:
  image: ubuntu-2004-lts
  machineType: standard-v1
  manager: CloudGroup
  maxSize: 1
  minSize: 1
  nodeLabels:
    kops.k8s.io/instancegroup: master-ru-central1-a
  role: Master
  subnets:
  - ru-central1-a
  zones:
  - ru-central1-a

---

apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  creationTimestamp: "2017-01-01T00:00:00Z"
  labels:
    kops.k8s.io/cluster: example.k8s.local
  name: master-ru-central1-b
spec:
  image: ubuntu-2004-lts
  machineType: standard-v1
  manager: CloudGroup
  maxSize: 1
  minSize: 1
  nodeLabels:
    kops.k8s.io/instancegroup: master-ru-central1-b
  role: Master
  subnets:
  - ru-central1-b
  zones:
  - ru-central1-b

---

apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  creationTimestamp: "2017-01-01T00:00:00Z"
  labels:
    kops.k8s.io/cluster: example.k8s.local
  name: master-ru-central1-c
spec:
  image: ubuntu-2004-lts
  machineType: standard-v1
  manager: CloudGroup
  maxSize: 1
  minSize: 1
  nodeLabels:
    kops.k8s.io/instancegroup: master-ru-central1-c
  role: Master
  subnets:
  - ru-central1-c
  zones:
  - ru-central1-c

---

apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  creationTimestamp: "2017-01-01T00:00:00Z"
  labels:
    kops.k8s.io/cluster: example.k8s.local
  name: nodes-ru-central1-a
spec:
  image: ubuntu-2004-lts
  machineType: standard-v1
  manager: CloudGroup
  maxSize: 1
  minSize: 1
  nodeLabels:
    kops.k8s.io/instancegroup: nodes-ru-central1-a
  role: Node
  subnets:
  - ru-central1-a
  zones:
  - ru-central1-a
```

Save above config to example.yaml

```bash
# create configuration for a ubuntu 20.04 + calico cluster
kops create -f example.yaml

# adds ssh key to access VMs later
kops create sshpublickey -i <your-ssh-key>.pub

# create cluster
kops update cluster --name example.k8s.local --yes

# export kubecfg
kops export kubeconfig -v9 --admin

# update your /etc/hosts with nlb api to access your k8s cluster
<your nlb api> api.example.k8s.local
```

Right now Yandex Cloud implementation uses etcd-manager external provider which require some manual actions.
```bash
#!/bin/bash

DATA=`yc compute instance list --folder-id <folder-id> --format json`

IPS=$(echo $DATA |jq -r '.[] | select(.name | startswith("master")).network_interfaces[].primary_v4_address.one_to_one_nat.address')
LOCAL_IPS=$(echo $DATA |jq -r '.[] | select(.name | startswith("master")).network_interfaces[].primary_v4_address.address')

for IP in $IPS;
do 
    ssh -i <your-ssh-key> -o StrictHostKeyChecking=no ubuntu@$IP "sudo mkdir -p /etc/kubernetes/etcd-manager/seeds"
    for LOCAL_IP in $LOCAL_IPS;
    do
        ssh -i <your-ssh-key> -o StrictHostKeyChecking=no ubuntu@$IP "sudo touch /etc/kubernetes/etcd-manager/seeds/$LOCAL_IP"
    done
done

for IP in $IPS;
do
    LIP=$(ssh -i <your-ssh-key> -o StrictHostKeyChecking=no ubuntu@$IP "hostname -I | tr \".\" \"-\" | tr -d ' '")
    ssh -i <your-ssh-key> -o StrictHostKeyChecking=no ubuntu@$IP "sudo mkdir -p \"/mnt/disks/ip-$LIP/mnt\""
done

```

Test ETCD main cluster.
```
# Run from any VM in your cluster
ETCDCTL_API=3 etcdctl --cacert /srv/kubernetes/kube-apiserver/etcd-ca.crt --key /srv/kubernetes/kube-apiserver/etcd-client.key --cert /srv/kubernetes/kube-apiserver/etcd-client.crt --endpoints https://127.0.0.1:4001 member list
```

Check your k8s cluster
```bash
kubectl get po -A
```
