# How it works

## 1) The CLI

Everything in `kops` is currently driven by a command line interface. We use [cobra](https://github.com/spf13/cobra) to define all of our command line UX.

All of the CLI code for kops can be found in `/cmd/kops` [link](https://github.com/kubernetes/kops/tree/master/cmd/kops)

For instance, if you are interested in finding the entry point to `kops create cluster` you would look in `/cmd/kops/create_cluster.go`. There you would find a function called `RunCreateCluster()`. That is the entry point of the command.

## 2) Storage

We have an abstracted way of interacting with our storage, called the `clientset`. This is an abstract way of interacting with remote storage for `kops`. This is commonly referred to as the `kops` **STATE STORE**.

We can access it through the `util.Factory` struct as in :

```Go
func RunMyCommand(f *util.Factory, out io.Writer, c *MyCommandOptions) error {
        clientset, _ := f.Clientset()
        cluster, _ := clientset.Clusters().Get(clusterName)
        fmt.Println(cluster)
}
```

The available Clientset functions are :

```
Create()
Update()
Get()
List()
```
## 3) The API

#### a) Cluster Spec

The `kops` API is a definition of struct members in Go code found [here](https://github.com/kubernetes/kops/tree/master/pkg/apis/kops). The `kops` API does **NOT** match the command line interface (by design). We use the native Kubernetes API machinery to manage versioning of the `kops` API.

The base level struct of the API is `api.Cluster{}` which is defined [here](https://github.com/kubernetes/kops/blob/master/pkg/apis/kops/cluster.go#L40). The top level struct contains meta information about the object such as the kind and version and the main data for the cluster itself can be found in `cluster.Spec`

It is important to note that the API members are a representation of a Kubernetes cluster. These values are stored in the `kops` **STATE STORE** mentioned above for later use. By design kops does not store information about the state of the cloud in the state store, if it can infer it from looking at the actual state of the cloud.

More information on the API can be found [here](https://github.com/kubernetes/kops/blob/master/docs/cluster_spec.md).

#### b) Instance Groups

In order for `kops` to create any servers, we will need to define instance groups. These are a slice of pointers to `kops.InstanceGroup` structs.

```go
var instanceGroups []*kops.InstanceGroup
```

Each instance group represents a group of instances in a cloud. Each instance group (or IG) defines values about the group of instances such as their size, volume information, etc. The definition can also be found in the `/pkg/apis/kops/instancegroup.go` file [here](https://github.com/kubernetes/kops/blob/master/pkg/apis/kops/instancegroup.go#L59).


## 4) Cloudup

#### a) The `ApplyClusterCmd`

After a user has built out a valid `api.Cluster{}` and valid `[]*kops.InstanceGroup` they can then begin interacting with the core logic in `kops`.

A user can build a `cloudup.ApplyClusterCmd` defined [here](https://github.com/kubernetes/kops/blob/master/upup/pkg/fi/cloudup/apply_cluster.go#L54) as follows: 


```go
applyCmd := &cloudup.ApplyClusterCmd{
    Cluster:         cluster,
    Models:          []string{"config", "proto", "cloudup"}, // ${GOPATH}/src/k8s.io/kops/upup/pkg/fi/cloudup/apply_cluster.go:52
    Clientset:       clientset,
    TargetName:      "target",                               // ${GOPATH}/src/k8s.io/kops/upup/pkg/fi/cloudup/target.go:19
    OutDir:          c.OutDir,
    DryRun:          isDryrun,
    MaxTaskDuration: 10 * time.Minute,                       // ${GOPATH}/src/k8s.io/kops/upup/pkg/fi/cloudup/apply_cluster.go
    InstanceGroups:  instanceGroups,
}
```

Now that the `ApplyClusterCmd` has been populated, we can attempt to run our apply. 

```go
err = applyCmd.Run()
```

This is where we enter the **core** of `kops` logic. The starting point can be found [here](https://github.com/kubernetes/kops/blob/master/upup/pkg/fi/cloudup/apply_cluster.go#L91). Based on the directives defined in the `ApplyClusterCmd` above, the apply operation will behave differently based on the input provided.
 
#### b) Validation
 
 From within the `ApplyClusterCmd.Run()` function we will first attempt to sanitize our input by validating the operation. There are many examples at the top of the function where we validate the input.
  
 
#### c) The Cloud
 
 The `cluster.Spec.CloudProvider` should have been populated earlier, and can be used to switch on to build our cloud as in [here](https://github.com/kubernetes/kops/blob/master/upup/pkg/fi/cloudup/utils.go#L37). If you are interested in creating a new cloud implementation the interface is defined [here](https://github.com/kubernetes/kops/blob/master/upup/pkg/fi/cloud.go#L26), with the AWS example [here](https://github.com/kubernetes/kops/blob/master/upup/pkg/fi/cloudup/awsup/aws_cloud.go#L65).
 
 **Note** As it stands the `FindVPCInfo()` function is a defined member of the interface. This is AWS only, and will eventually be pulled out of the interface. For now please implement the function as a no-op.
 
#### d) Models

A model is an evolution from the static YAML models in `kops v1.4`. There is a lot of improvements planned for these in the next major kops release. The models are indexed by a string. With the 3 primary models being 

```
config
proto
cloudup
```

Models are what map an ambiguous Cluster Spec (defined earlier) to **tasks**. Each **task** is a representation of an API request against a cloud. If you plan on implementing a new cloud, one option would be to define a new model, and build custom model code for your new model.

The `cloudup` model is what is used to map a cluster spec with `cluster.Spec.CloudProvider` = `aws`. 

**Note** this name is probably a misnomer, and is a reflection of the evolution of the `kops` core.

The existing `cloudup` model code can be found [here](https://github.com/kubernetes/kops/tree/master/pkg/model). 

**Note** that there is room here to redefine the directory structure based on models. EG: Moving these models into a new package, and renaming the model key.

Once a model builder has been defined as in [here](https://github.com/kubernetes/kops/blob/master/upup/pkg/fi/cloudup/apply_cluster.go#L373) the code will automatically be called.

From within the builder, we notice there is concrete logic for each builder. The logic will dictate which tasks need to be called in order to apply a resource to a cloud. The tasks are added by calling the `AddTask()` function as in [here](https://github.com/kubernetes/kops/blob/master/pkg/model/network.go#L69).
 
Once the models have been parsed, all the tasks should have been set.

#### e) Tasks

A task is typically a representation of a single API call. The task interface is defined [here](https://github.com/kubernetes/kops/blob/master/upup/pkg/fi/task.go#L26).

**Note** for more advanced clouds like AWS, there is also `Find()` and `Render()` functions in the core logic of executing the tasks defined [here](https://github.com/kubernetes/kops/blob/master/upup/pkg/fi/executor.go).

## 5) Nodeup

Nodeup is a standalone binary that handles bootstrapping the Kubernetes cluster. There is a shell script [here](https://github.com/kubernetes/kops/blob/master/pkg/model/resources/nodeup.go) that will bootstrap nodeup. The AWS implementation uses `cloud-init` to run the script on an instance. All new clouds will need to figure out best practices for bootstrapping `nodeup` on their platform.

