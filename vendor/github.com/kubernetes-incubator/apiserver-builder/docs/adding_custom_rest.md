# Adding resources

**Important:** Read [this doc](https://github.com/kubernetes-incubator/apiserver-builder/blob/master/docs/adding_resources.md)
first to understand how resources are added

## Create a resource with custom rest

You can implement your own REST implementation instead of using the
standard storage by providing the `rest=KindREST` parameter
and providing a `newKindREST() rest.Storage {}` function to return the
storage.

For more information on custom REST implementations, see the
[subresources doc](https://github.com/kubernetes-incubator/apiserver-builder/blob/master/docs/adding_subresources.md)

```go
// +genclient=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +resource:path=foos,rest=FooREST
// +k8s:openapi-gen=true
// Foo defines some thing
type Foo struct {
    // Your resource definition here
}

// Initialize custom REST storage
func NewFooREST() rest.Storage {
    // Initialize fields of custom REST implementation
}

// Your rest.Storage implementation below
// ...
```

**Warning:** NewFooREST() should not contain any non-trivial logic, besides
simply initializing the fields of the struct, that represents the custom REST.
See [this issue](https://github.com/kubernetes-incubator/apiserver-builder/issues/92) for details.
