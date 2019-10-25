# Using A Manifest to Manage kops Clusters

This document also applies to using the `kops` API to customize a Kubernetes cluster with or without using YAML or JSON.

## Table of Contents

   * [Using A Manifest to Manage kops Clusters](#using-a-manifest-to-manage-kops-clusters)
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

> You can see all the options that are currently supported in Kops [here](https://github.com/kubernetes/kops/blob/master/pkg/apis/kops/componentconfig.go) or [more prettily here](https://godoc.org/k8s.io/kops/pkg/apis/kops#ClusterSpec)

The following is a list of the benefits of using a file to manage instances.

- Capability to access API values that are not accessible via the command line such as setting the max price for spot instances.
- Create, replace, update, and delete clusters without entering an interactive editor. This feature is helpful when automating cluster creation.
- Ability to check-in files to source control that represents an installation.
- Run commands such as `kops delete -f mycluster.yaml`.

## Exporting a Cluster

At this time you must run `kops create cluster` and then export the YAML from the state store. We plan in the future to have the capability to generate kops YAML via the command line. The following is an example of creating a cluster and exporting the YAML.

```shell
export NAME=k8s.example.com
export KOPS_STATE_STORE=s3://example-state-store
 kops create cluster $NAME \
    --zones "us-east-2a,us-east-2b,us-east-2c" \
    --master-zones "us-east-2a,us-east-2b,us-east-2c" \
    --networking weave \
    --topology private \
    --bastion \
    --node-count 3 \
    --node-size m4.xlarge \
    --kubernetes-version v1.6.6 \
    --master-size m4.large \
    --vpc vpc-6335dd1a \
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
  creationTimestamp: 2017-05-04T23:21:47Z
  name: k8s.example.com
spec:
  api:
    loadBalancer:
      type: Public
  authorization:
    alwaysAllow: {}
  channel: stable
  cloudProvider: aws
  configBase: s3://example-state-store/k8s.example.com
  etcdClusters:
  - etcdMembers:
    - instanceGroup: master-us-east-2d
      name: a
    - instanceGroup: master-us-east-2b
      name: b
    - instanceGroup: master-us-east-2c
      name: c
    name: main
  - etcdMembers:
    - instanceGroup: master-us-east-2d
      name: a
    - instanceGroup: master-us-east-2b
      name: b
    - instanceGroup: master-us-east-2c
      name: c
    name: events
  kubernetesApiAccess:
  - 0.0.0.0/0
  kubernetesVersion: 1.6.6
  masterPublicName: api.k8s.example.com
  networkCIDR: 172.20.0.0/16
  networkID: vpc-6335dd1a
  networking:
    weave: {}
  nonMasqueradeCIDR: 100.64.0.0/10
  sshAccess:
  - 0.0.0.0/0
  subnets:
  - cidr: 172.20.32.0/19
    name: us-east-2d
    type: Private
    zone: us-east-2d
  - cidr: 172.20.64.0/19
    name: us-east-2b
    type: Private
    zone: us-east-2b
  - cidr: 172.20.96.0/19
    name: us-east-2c
    type: Private
    zone: us-east-2c
  - cidr: 172.20.0.0/22
    name: utility-us-east-2d
    type: Utility
    zone: us-east-2d
  - cidr: 172.20.4.0/22
    name: utility-us-east-2b
    type: Utility
    zone: us-east-2b
  - cidr: 172.20.8.0/22
    name: utility-us-east-2c
    type: Utility
    zone: us-east-2c
  topology:
    bastion:
      bastionPublicName: bastion.k8s.example.com
    dns:
      type: Public
    masters: private
    nodes: private

---

apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  creationTimestamp: 2017-05-04T23:21:48Z
  labels:
    kops.k8s.io/cluster: k8s.example.com
  name: bastions
spec:
  image: kope.io/k8s-1.6-debian-jessie-amd64-hvm-ebs-2017-05-02
  machineType: t2.micro
  maxSize: 1
  minSize: 1
  role: Bastion
  subnets:
  - utility-us-east-2d
  - utility-us-east-2b
  - utility-us-east-2c


---

apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  creationTimestamp: 2017-05-04T23:21:47Z
  labels:
    kops.k8s.io/cluster: k8s.example.com
  name: master-us-east-2d
spec:
  image: kope.io/k8s-1.6-debian-jessie-amd64-hvm-ebs-2017-05-02
  machineType: m4.large
  maxSize: 1
  minSize: 1
  role: Master
  subnets:
  - us-east-2d


---

apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  creationTimestamp: 2017-05-04T23:21:47Z
  labels:
    kops.k8s.io/cluster: k8s.example.com
  name: master-us-east-2b
spec:
  image: kope.io/k8s-1.6-debian-jessie-amd64-hvm-ebs-2017-05-02
  machineType: m4.large
  maxSize: 1
  minSize: 1
  role: Master
  subnets:
  - us-east-2b


---

apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  creationTimestamp: 2017-05-04T23:21:48Z
  labels:
    kops.k8s.io/cluster: k8s.example.com
  name: master-us-east-2c
spec:
  image: kope.io/k8s-1.6-debian-jessie-amd64-hvm-ebs-2017-05-02
  machineType: m4.large
  maxSize: 1
  minSize: 1
  role: Master
  subnets:
  - us-east-2c


---

apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  creationTimestamp: 2017-05-04T23:21:48Z
  labels:
    kops.k8s.io/cluster: k8s.example.com
  name: nodes
spec:
  image: kope.io/k8s-1.6-debian-jessie-amd64-hvm-ebs-2017-05-02
  machineType: m4.xlarge
  maxSize: 3
  minSize: 3
  role: Node
  subnets:
  - us-east-2d
  - us-east-2b
  - us-east-2c
```

## YAML Examples

With the above YAML file, a user can add configurations that are not available via the command line. For instance, you can add a `maxPrice` value to a new instance group and use spot instances. Also add node and cloud labels for the new instance group.

```yaml
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  creationTimestamp: 2017-05-04T23:21:48Z
  labels:
    kops.k8s.io/cluster: k8s.example.com
  name: my-crazy-big-nodes
spec:
 nodeLabels:
    spot: "true"
  cloudLabels:
    team: example
    project: ion
  image: kope.io/k8s-1.6-debian-jessie-amd64-hvm-ebs-2017-05-02
  machineType: m4.10xlarge
  maxSize: 42
  minSize: 42
  maxPrice: "0.35"
  role: Node
  subnets:
  - us-east-2c
```

This configuration will create an autoscale group that will include 42 m4.10xlarge nodes running as spot instances with custom labels.

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
  creationTimestamp: 2017-05-04T23:21:47Z
  name: k8s.example.com
spec:
  api:
```

Full documentation is accessible via [godoc](https://godoc.org/k8s.io/kops/pkg/apis/kops#ClusterSpec).

The `ClusterSpec` allows a user to set configurations for such values as Docker log driver, Kubernetes API server log level, VPC for reusing a VPC (`NetworkID`), and the Kubernetes version.

More information about some of the elements in the `ClusterSpec` is available in the following:

-  Cluster Spec [document](cluster_spec.md) which outlines some of the values in the Cluster Specification.
- [Etcd Encryption](operations/etcd_backup_restore_encryption.md)
- [GPU](gpu.md) setup
- [IAM Roles](iam_roles.md) - adding additional IAM roles.
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
  creationTimestamp: 2017-05-04T23:21:48Z
  name: foo
spec:
```

Full documentation is accessible via [godocs](https://godoc.org/k8s.io/kops/pkg/apis/kops#InstanceGroupSpec).

Instance Groups map to Auto Scaling Groups in AWS, and Instance Groups in GCE. They are an API level description of a group of compute instances used as Masters or Nodes.

More documentation is available in the [Instance Group](instance_groups.md) document.

## Closing Thoughts

Using YAML or JSON-based configuration for building and managing kops clusters is powerful, but use this strategy with caution.

- If you do not need to define or customize a value, let kops set that value. Setting too many values prevents kops from doing its job in setting up the cluster and you may end up with strange bugs.
- If you end up with strange bugs, try letting kops do more.
- Be cautious, take care, and test outside of production!

If you need to run a custom version of Kubernetes Controller Manager, set `kubeControllerManager.image` and update your cluster. This is the beauty of using a manifest for your cluster!
