# Using A Manifest to Manage kOps Clusters

This document also applies to using the `kops` API to customize a Kubernetes cluster with or without using YAML or JSON.

## Table of Contents

   * [Using A Manifest to Manage kOps Clusters](#using-a-manifest-to-manage-kops-clusters)
   * [Background](#background)
   * [Exporting a Cluster](#exporting-a-cluster)
   * [YAML Examples](#yaml-examples)
   * [Further References](#further-references)
   * [Cluster Spec](#cluster-spec)
   * [Instance Groups](#instance-groups)
   * [Closing Thoughts](#closing-thoughts)

## Background

> We like to think of it as `kubectl` for Clusters.

Because of the above statement `kops` includes an API which provides a feature for users to utilize YAML or JSON manifests for managing their `kops` created Kubernetes installations. In the same way that you can use a YAML manifest to deploy a Job, you can deploy and manage a `kops` Kubernetes instance with a manifest. All of these values are also usable via the interactive editor with `kops edit`.

> You can see all the options that are currently supported in kOps [here](https://github.com/kubernetes/kops/blob/master/pkg/apis/kops/componentconfig.go) or [more prettily here](https://pkg.go.dev/k8s.io/kops/pkg/apis/kops#ClusterSpec)

The following is a list of the benefits of using a file to manage instances.

- Capability to access API values that are not accessible via the command line such as setting the max price for spot instances.
- Create, replace, update, and delete clusters without entering an interactive editor. This feature is helpful when automating cluster creation.
- Ability to check-in files to source control that represents an installation.
- Run commands such as `kops delete -f mycluster.yaml`.

## Exporting a Cluster

At this time you must run `kops create cluster` and then export the YAML from the state store. We plan in the future to have the capability to generate kOps YAML via the command line. The following is an example of creating a cluster and exporting the YAML.

```shell
export NAME=k8s.example.com
export KOPS_STATE_STORE=s3://example-state-store
kops create cluster $NAME \
    --zones "us-east-2a,us-east-2b,us-east-2c" \
    --control-plane-zones "us-east-2a,us-east-2b,us-east-2c" \
    --networking calico \
    --topology private \
    --bastion \
    --node-count 3 \
    --node-size m5.xlarge \
    --kubernetes-version v1.36.4 \
    --control-plane-size m5.large \
    --network-id vpc-6335dd1a \
    --dry-run \
    -o yaml > $NAME.yaml
```

The above command exports a YAML document which contains the definition of the cluster, `kind: Cluster`, and the definitions of the instance groups, `kind: InstanceGroup`.

NOTE: If you run `kops get cluster $NAME -o yaml > $NAME.yaml`, you will only get a cluster spec. Use the command above (`kops get $NAME ...`)for both the cluster spec and all instance groups.

The following is the contents of the exported YAML file.

```yaml
apiVersion: kops.k8s.io/v1alpha2
kind: Cluster
metadata:
  name: k8s.example.com
spec:
  api:
    loadBalancer:
      class: Network
      type: Public
  authorization:
    rbac: {}
  channel: stable
  cloudProvider: aws
  configBase: s3://example-state-store/k8s.example.com
  etcdClusters:
  - cpuRequest: 200m
    etcdMembers:
    - encryptedVolume: true
      instanceGroup: control-plane-us-east-2a
      name: a
    - encryptedVolume: true
      instanceGroup: control-plane-us-east-2b
      name: b
    - encryptedVolume: true
      instanceGroup: control-plane-us-east-2c
      name: c
    manager:
      backupRetentionDays: 90
    memoryRequest: 100Mi
    name: main
  - cpuRequest: 100m
    etcdMembers:
    - encryptedVolume: true
      instanceGroup: control-plane-us-east-2a
      name: a
    - encryptedVolume: true
      instanceGroup: control-plane-us-east-2b
      name: b
    - encryptedVolume: true
      instanceGroup: control-plane-us-east-2c
      name: c
    manager:
      backupRetentionDays: 90
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
  kubernetesVersion: v1.36.4
  networkCIDR: 172.20.0.0/16
  networking:
    calico: {}
  nonMasqueradeCIDR: 100.64.0.0/10
  sshAccess:
  - 0.0.0.0/0
  - ::/0
  subnets:
  - cidr: 172.20.64.0/18
    name: us-east-2a
    type: Private
    zone: us-east-2a
  - cidr: 172.20.128.0/18
    name: us-east-2b
    type: Private
    zone: us-east-2b
  - cidr: 172.20.192.0/18
    name: us-east-2c
    type: Private
    zone: us-east-2c
  - cidr: 172.20.0.0/21
    name: utility-us-east-2a
    type: Utility
    zone: us-east-2a
  - cidr: 172.20.8.0/21
    name: utility-us-east-2b
    type: Utility
    zone: us-east-2b
  - cidr: 172.20.16.0/21
    name: utility-us-east-2c
    type: Utility
    zone: us-east-2c
  topology:
    dns:
      type: None

---

apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: k8s.example.com
  name: control-plane-us-east-2a
spec:
  image: 099720109477/ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-20260714
  machineType: m5.large
  maxSize: 1
  minSize: 1
  nodeLabels:
    kops.k8s.io/instancegroup: control-plane-us-east-2a
  role: Master
  subnets:
  - us-east-2a

---

apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: k8s.example.com
  name: control-plane-us-east-2b
spec:
  image: 099720109477/ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-20260714
  machineType: m5.large
  maxSize: 1
  minSize: 1
  nodeLabels:
    kops.k8s.io/instancegroup: control-plane-us-east-2b
  role: Master
  subnets:
  - us-east-2b

---

apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: k8s.example.com
  name: control-plane-us-east-2c
spec:
  image: 099720109477/ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-20260714
  machineType: m5.large
  maxSize: 1
  minSize: 1
  nodeLabels:
    kops.k8s.io/instancegroup: control-plane-us-east-2c
  role: Master
  subnets:
  - us-east-2c

---

apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: k8s.example.com
  name: nodes-us-east-2a
spec:
  image: 099720109477/ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-20260714
  machineType: m5.xlarge
  maxSize: 1
  minSize: 1
  nodeLabels:
    kops.k8s.io/instancegroup: nodes-us-east-2a
  role: Node
  subnets:
  - us-east-2a

---

apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: k8s.example.com
  name: nodes-us-east-2b
spec:
  image: 099720109477/ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-20260714
  machineType: m5.xlarge
  maxSize: 1
  minSize: 1
  nodeLabels:
    kops.k8s.io/instancegroup: nodes-us-east-2b
  role: Node
  subnets:
  - us-east-2b

---

apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: k8s.example.com
  name: nodes-us-east-2c
spec:
  image: 099720109477/ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-20260714
  machineType: m5.xlarge
  maxSize: 1
  minSize: 1
  nodeLabels:
    kops.k8s.io/instancegroup: nodes-us-east-2c
  role: Node
  subnets:
  - us-east-2c

---

apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: k8s.example.com
  name: bastions
spec:
  image: 099720109477/ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-20260714
  machineType: t3.micro
  maxSize: 1
  minSize: 1
  nodeLabels:
    kops.k8s.io/instancegroup: bastions
  role: Bastion
  subnets:
  - us-east-2a
  - us-east-2b
  - us-east-2c
```

## YAML Examples

With the above YAML file, a user can add configurations that are not available via the command line. For instance, you can add a `maxPrice` value to a new instance group and use spot instances. Also add node and cloud labels for the new instance group.

```yaml
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: k8s.example.com
  name: my-crazy-big-nodes
spec:
  nodeLabels:
    spot: "true"
  cloudLabels:
    team: example
    project: ion
  image: 099720109477/ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-20260714
  machineType: m5.12xlarge
  maxSize: 42
  minSize: 42
  maxPrice: "0.35"
  role: Node
  subnets:
  - us-east-2c
```

This configuration will create an autoscale group that will include 42 m5.12xlarge nodes running as spot instances with custom labels.

To create the cluster execute:

```shell
kops create -f $NAME.yaml
kops create secret --name $NAME sshpublickey admin -i ~/.ssh/id_rsa.pub
kops update cluster $NAME --yes
kops rolling-update cluster $NAME --yes
```

Please refer to the rolling-update [documentation](cli/kops_rolling-update_cluster.md).

Update the cluster spec YAML file, and to update the cluster run:

```shell
kops replace -f $NAME.yaml
kops update cluster $NAME --yes
kops rolling-update cluster $NAME --yes
```

Please refer to the rolling-update [documentation](cli/kops_rolling-update_cluster.md).

## Further References

`kops` implements a full API that defines the various elements in the YAML file exported above. Two top level components exist; `ClusterSpec` and `InstanceGroup`.

### Cluster Spec

```yaml
apiVersion: kops.k8s.io/v1alpha2
kind: Cluster
metadata:
  name: k8s.example.com
spec:
  api:
```

Full documentation is accessible via [godoc](https://pkg.go.dev/k8s.io/kops/pkg/apis/kops#ClusterSpec).

The `ClusterSpec` allows a user to set configurations for such values as Kubernetes API server log level, VPC for reusing a VPC (`NetworkID`), and the Kubernetes version.

More information about some of the elements in the `ClusterSpec` is available in the following:

-  Cluster Spec [document](cluster_spec.md) which outlines some of the values in the Cluster Specification.
- [Etcd Encryption](operations/etcd_backup_restore_encryption.md)
- [GPU](gpu.md) setup
- [Instance IAM Roles](iam_roles.md) - adding additional instance IAM roles.
- [Labels](labels.md)
- [Run In Existing VPC](run_in_existing_vpc.md)

To access the full configuration that a `kops` installation is running execute:

```bash
kops get cluster $NAME --full -o yaml
```

This command prints the entire YAML configuration. But _do not_ use the full document, you may experience strange and unique unwanted behaviors.

### Instance Groups

```yaml
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  name: foo
spec:
```

Full documentation is accessible via [godocs](https://pkg.go.dev/k8s.io/kops/pkg/apis/kops#InstanceGroupSpec).

Instance Groups map to Auto Scaling Groups in AWS, and Instance Groups in GCE. They are an API level description of a group of compute instances used as Masters or Nodes.

More documentation is available in the [Instance Group](instance_groups.md) document.

## Closing Thoughts

Using YAML or JSON-based configuration for building and managing kOps clusters is powerful, but use this strategy with caution.

- If you do not need to define or customize a value, let kOps set that value. Setting too many values prevents kOps from doing its job in setting up the cluster and you may end up with strange bugs.
- If you end up with strange bugs, try letting kOps do more.
- Be cautious, take care, and test outside of production!

If you need to run a custom version of Kubernetes Controller Manager, set `kubeControllerManager.image` and update your cluster. This is the beauty of using a manifest for your cluster!
