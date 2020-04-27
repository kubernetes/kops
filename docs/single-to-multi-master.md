# Migrating from single to multi-master

Switching from a single-master to a multi-maser Kubernetes cluster is an entirely graceful procedure when using etcd-manager.
If you are still using legacy etcd, you need to migrate to etcd-manager first.

## Create instance groups

### Create new subnets

Start out by deciding which availability zones you want to deploy the masters to. You can only have one master per availability zone.

Then you need to add new subnets for your availability zones. Which subnets you need to add depend on which topology you have chosen. Simplest is to copy the sections you already have. Make sure that you add additional subnets per type. E.g if you have a `private` and a `utility` subnet, you need to copy both.

```bash
kops get cluster -o yaml > mycluster.yaml
```

Change the subnet section to look something like this:
```yaml
  - cidr: 172.20.32.0/19
    name: eu-west-1a
    type: Private
    zone: eu-west-1a
  - cidr: 172.20.64.0/19
    name: eu-west-1b
    type: Private
    zone: eu-west-1b
  - cidr: 172.20.96.0/19
    name: eu-west-1c
    type: Private
    zone: eu-west-1c
  - cidr: 172.20.0.0/22
    name: utility-eu-west-1a
    type: Utility
    zone: eu-west-1a
  - cidr: 172.20.4.0/22
    name: utility-eu-west-1b
    type: Utility
    zone: eu-west-1b
  - cidr: 172.20.8.0/22
    name: utility-eu-west-1c
    type: Utility
    zone: eu-west-1c
```

### Create new master instance groups

The next step is creating two new instance groups for the new masters. 

```bash
kops create instancegroup master-<subnet name> --subnet <subnet name> --role Master
```

Example:

```bash
kops create ig master-us-west-1d --subnet us-west-1d --role Master
```

This command will bring up an editor with the default values. Ensure that:

 * `maxSize` and `minSize` is 1
 * only one zone is listed
 * you have the correct image and machine type

### Reference the new masters in your cluster configuration

Bring up `mycluster.yaml` again to add etcd members to each of new masters.

```bash
$EDITOR mycluster.yaml
```

 * In `.spec.etcdClusters` add 2 new members in each cluster, one for each new
 availability zone.

```yaml
    - instanceGroup: master-<availability-zone2>
      name: <availability-zone2-name>
    - instanceGroup: master-<availability-zone3>
      name: <availability-zone3-name>
```

Example:

```yaml
etcdClusters:
  - etcdMembers:
    - instanceGroup: master-eu-west-1a
      name: a
    - instanceGroup: master-eu-west-1b
      name: b
    - instanceGroup: master-eu-west-1c
      name: c
    name: main
  - etcdMembers:
    - instanceGroup: master-eu-west-1a
      name: a
    - instanceGroup: master-eu-west-1b
      name: b
    - instanceGroup: master-eu-west-1c
      name: c
    name: events
```

### Update Cluster to launch new masters

Update the cluster spec and apply the config by running the following:

```bash
kops replace -f mycluster.yaml
kops update cluster example.com
kops update cluster example.com --yes
```

This will launch the two new masters. You will also need to roll the old master so that it can join the new etcd cluster.

After about 5 minutes all three masters should have found each other. Run the following to ensure everything is running as expected.

```bash
kops validate cluster
```

While rotating the original master is not strictly necessary, kops will say it needs updating because of the configuration change.

```
kops rolling-update cluster --yes
```