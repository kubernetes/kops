# Watching Kubernetes resources

### Before you begin - configure your 

- Install minikube and start a new cluster
- Create a new GO project
- In the GO project init the repo with `apiserver-boot init`
- Create a resource with `apiserver-boot create group version resource`
- Build the executables with `apiserver-boot build executables`
- Create the aggregation config with `apiserver-boot build config --local-minikube --name test --namespace default`
- Install the aggregated apiserver config into minikube with `kubectl create -f config/apiserver.yaml`

## Run the sharedinformers to watch the types

Setup the informers and related objects that are shared across all controllers
run by the controller manager.  The following steps are done once per-controller-manager
(go binary), not once per-controller (go package).

**Step 1:** If `pkg/controller/sharedinformer/informers.go` does not already exist, create it.

This is automatically created for you by `apiserver-boot` when creating a resource for
releases alpha.18+.

**Step 2:** Enable watching Kubernetes types and initializing the ClientSet

This will require that the controller-manager and apiserver are run as aggregated
with a core apiserver, as the controller-manager will be watch both core and custom
types.

Setting this to true will also allow your controller to read/write to the core apiserver using the
`ClientSet` passed into your controller's `Init` function through `si.KubernetesClientSet`.

```go
// SetupKubernetesTypes registers the config for watching Kubernetes types
func (si *SharedInformers) SetupKubernetesTypes() bool {
	return true
}
```

**Step 3:** Start the informers for the resources you want to watch

For each Kind you that want to get notified about when it is created/updated/deleted,
`Run` a corresponding Informer in `StartAdditionalInformers`.

**Note:** If you want to watch Deployments, you do not need to start informers for all
group/versions.  You only need to watch 1 group/version and other Deployment group/versions
will be converted to the group/version you are watching.

```go
// StartAdditionalInformers starts watching Deployments
func (si *SharedInformers) StartAdditionalInformers(shutdown <-chan struct{}) {
	go si.KubernetesFactory.Apps().V1beta1().Deployments().Informer().Run(shutdown)
}
```

## Register your controller with the informer

**Step 1:** Map the Kubernetes resource instance to the key of an instance of your resource

When receiving a notification for the Kubernetes resource (e.g. Deployment) you want
to run the `Reconcile` loop for your resource instance that owns the Kubernetes resource.
If you write an `OwnerReference` for each of the Kubernetes resources created by your resource,
you may look at that field to find the key for your resource.

- Cast the argument to the type
  - **Note:** double check the group/version are consistent between
    - the informer you started in pkg/controller/sharedinformers/informers.go
    - the type you cast the argument to
    - the informer with which you register your reconcile loop (Step 2)

```go
func (c *FooControllerImpl) DeploymentToFoo(i interface{}) (string, error) {
	d, _ := i.(*v1beta1.Deployment)
	log.Printf("Deployment update: %v", d.Name)
	if len(d.OwnerReferences) == 1 && d.OwnerReferences[0].Kind == "Foo" {
		return d.Namespace + "/" + d.OwnerReferences[0].Name, nil
	} else {
		// Not owned
		return "", nil
	}
}
```

**Step 2:** Register your resource Reconcile with the informer

In your controller init function tie it all together:
- The sharedinformer you started that watches for events
- The conversion from a Deployment to the key of your resource
- The reconcile function that takes the key of one of your resources

```go
func (c *FooControllerImpl) Init(
	config *rest.Config,
	si *sharedinformers.SharedInformers,
	reconcileKey func(key string) error) {
    ...
	si.Watch(
	    "FooDeployment",
	    si.KubernetesFactory.Extensions().V1beta1().Deployments().Informer(),
	    c.DeploymentToFoo, reconcileKey)
}
```

### Create/delete/update the Kubernetes from your controller reconcile loop

Add a PodSpec to your resource Spec.

Use the si.KubernetesClientSet from within your controllers `Reconcile` function
to update Kubernetes objects.

**Note**: Consider using a `Lister` for reading and indexing cached objects to reduce load
on the apiserver.

### Notes

Value validation will not be executed for the fields in your Spec when your resource
is created.  We hope to fix this in the future by making the Validation functions for
Kubernetes API objects available.

### Run aggregated with minikube

Run your extension server locally but aggregated with minikube:

`apiserver-boot run local-minikube`

To run without rebuilding the binaries or generating code.

`apiserver-boot run local-minikube --build=false`