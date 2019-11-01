# Getting Started with kops on GCE

Make sure you have [installed kops](../install.md) and [installed kubectl](../install.md), and installed
the [gcloud tools](https://cloud.google.com/sdk/downloads).

You'll need a Google Cloud account, and make sure that gcloud is logged in to your account using `gcloud init`.

You should confirm that basic commands like `gcloud compute zones list` are working.

You'll also need to [configure default credentials](https://developers.google.com/accounts/docs/application-default-credentials), using `gcloud auth application-default login`.

<!-- TODO: Can we get rid of `gcloud auth application-default login` ? -->

# Creating a state store

kops needs a state store, to hold the configuration for your clusters.  The simplest configuration
for Google Cloud is to store it in a Google Cloud Storage bucket in the same account, so that's how we'll
start.

So, just create an empty bucket - you can use any (available) name - e.g. `gsutil mb gs://kubernetes-clusters/`

Further, rather than typing the `--state` argument every time, it's much easier to export the `KOPS_STATE_STORE`
environment variable:

```
export KOPS_STATE_STORE=gs://kubernetes-clusters/
```

You can also put this in your `~/.bashrc` or similar.

# Creating our first cluster

`kops create cluster` creates the Cluster object and InstanceGroup object you'll be working with in kops:


    PROJECT=`gcloud config get-value project`
    export KOPS_FEATURE_FLAGS=AlphaAllowGCE # to unlock the GCE features
    kops create cluster simple.k8s.local --zones us-central1-a --state ${KOPS_STATE_STORE}/ --project=${PROJECT}


You can now list the Cluster objects in your kops state store (the GCS bucket
we created).


    > kops get cluster --state ${KOPS_STATE_STORE}

    NAME                CLOUD        ZONES
    simple.k8s.local    gce          us-central1-a


<!-- TODO: Fix bug where zones not showing up -->

This shows that you have one Cluster object configured, named `simple.k8s.local`.  The cluster holds the cluster-wide configuration for
a kubernetes cluster - things like the kubernetes version, and the authorization policy in use.

The `kops` tool should feel a lot like `kubectl` - kops uses the same API machinery as kubernetes,
so it should behave similarly, although now you are managing kubernetes clusters, instead of managing
objects on a kubernetes cluster.

You can see the details of your Cluster object by doing:

    > kops get cluster --state ${KOPS_STATE_STORE}/ simple.k8s.local -oyaml

    apiVersion: kops.k8s.io/v1alpha2
    kind: Cluster
    metadata:
      creationTimestamp: 2017-10-03T05:07:27Z
      name: simple.k8s.local
    spec:
      api:
        loadBalancer:
          type: Public
      authorization:
        alwaysAllow: {}
      channel: stable
      cloudProvider: gce
      configBase: gs://kubernetes-clusters/simple.k8s.local
      etcdClusters:
      - etcdMembers:
        - instanceGroup: master-us-central1-a
          name: a
        name: main
      - etcdMembers:
        - instanceGroup: master-us-central1-a
          name: a
        name: events
      iam:
        legacy: false
      kubernetesApiAccess:
      - 0.0.0.0/0
      kubernetesVersion: 1.7.2
      masterPublicName: api.simple.k8s.local
      networking:
        kubenet: {}
      nonMasqueradeCIDR: 100.64.0.0/10
      project: my-gce-project
      sshAccess:
      - 0.0.0.0/0
      subnets:
      - name: us-central1
        region: us-central1
        type: Public
      topology:
        dns:
          type: Public
        masters: public
        nodes: public

Similarly, you can also see your InstanceGroups using:

    > kops get instancegroup --state ${KOPS_STATE_STORE}/ --name simple.k8s.local

    NAME                    ROLE    MACHINETYPE     MIN    MAX    SUBNETS
    master-us-central1-a    Master  n1-standard-1   1      1      us-central1
    nodes                   Node    n1-standard-2   2      2      us-central1


<!-- TODO: Fix subnets vs regions -->

InstanceGroups are the other main kops object - an InstanceGroup manages a set of cloud instances,
which then are registered in kubernetes as Nodes.  You have multiple InstanceGroups for different types
of instances / Nodes - in our simple example we have one for our master (which only has a single member),
and one for our nodes (and we have two nodes configured).

We'll see a lot more of Cluster objects and InstanceGroups as we use kops to reconfigure clusters.  But let's get
on with our first cluster.

# Creating a cluster

`kops create cluster` created the Cluster object & InstanceGroup object in our state store,
but didn't actually create any instances or other cloud objects in GCE.  To do that, we'll use
`kops update cluster`.

`kops update cluster` without `--yes` will show you a preview of all the changes will be made;
it is very useful to see what kops is about to do, before actually making the changes.

Run `kops update cluster simple.k8s.local` and peruse the changes.

We're now finally ready to create the object: `kops update cluster simple.k8s.local --yes`

(If you haven't created an SSH key, you'll have to `ssh-keygen -t rsa`)

<!-- TODO: We don't need this on GCE; remove SSH key requirement -->

Your cluster is created in the background - kops actually creates GCE Managed Instance Groups
that run the instances; this ensures that even if instances are terminated, they will automatically
be relaunched by GCE and your cluster will self-heal.

After a few minutes, you should be able to do `kubectl get nodes` and your first cluster should be ready!

# Enjoy

At this point you have a kubernetes cluster - the core commands to do so are as simple as `kops create cluster`
and `kops update cluster`.  There's a lot more power in kops, and even more power in kubernetes itself, so we've
put a few jumping off places here.  But when you're done, don't forget to [delete your cluster](#deleting-the-cluster).

* [Manipulate InstanceGroups](../tutorial/working-with-instancegroups.md) to add more nodes, change image

# Deleting the cluster

When you're done using the cluster, you should delete it to release the cloud resources.  `kops delete cluster` is
the command.  When run without `--yes` it shows a preview of the objects it will delete:


    > kops delete cluster simple.k8s.local
    TYPE                    NAME                                                    ID
    Address                 api-simple-k8s-local                                    api-simple-k8s-local
    Disk                    a-etcd-events-simple-k8s-local                          a-etcd-events-simple-k8s-local
    Disk                    a-etcd-main-simple-k8s-local                            a-etcd-main-simple-k8s-local
    ForwardingRule          api-simple-k8s-local                                    api-simple-k8s-local
    Instance                master-us-central1-a-9847                               us-central1-a/master-us-central1-a-9847
    Instance                nodes-0s0w                                              us-central1-a/nodes-0s0w
    Instance                nodes-dvlq                                              us-central1-a/nodes-dvlq
    InstanceGroupManager    a-master-us-central1-a-simple-k8s-local                 us-central1-a/a-master-us-central1-a-simple-k8s-local
    InstanceGroupManager    a-nodes-simple-k8s-local                                us-central1-a/a-nodes-simple-k8s-local
    InstanceTemplate        master-us-central1-a-simple-k8s-local-1507008700        master-us-central1-a-simple-k8s-local-1507008700
    InstanceTemplate        nodes-simple-k8s-local-1507008700                       nodes-simple-k8s-local-1507008700
    Route                   simple-k8s-local-715bb0c7-a7fc-11e7-93d7-42010a800002   simple-k8s-local-715bb0c7-a7fc-11e7-93d7-42010a800002
    Route                   simple-k8s-local-9a2a08e8-a7fc-11e7-93d7-42010a800002   simple-k8s-local-9a2a08e8-a7fc-11e7-93d7-42010a800002
    Route                   simple-k8s-local-9c17a4e6-a7fc-11e7-93d7-42010a800002   simple-k8s-local-9c17a4e6-a7fc-11e7-93d7-42010a800002
    TargetPool              api-simple-k8s-local                                    api-simple-k8s-local

    Must specify --yes to delete cluster


After you've double-checked you're deleting exactly what you want to delete, run `kops delete cluster simple.k8s.local --yes`.
