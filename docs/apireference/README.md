# How to update the apireference docs

These instructions assume:

- [apiserver-builder](https://github.com/kubernetes-incubator/apiserver-builder/) version v0.1-alpha.7 is installed and on the path.

Or build the most recent from src using:

```sh
# Install the apiserver-builder commands
go get -u github.com/kubernetes-incubator/apiserver-builder/cmd/...

# Install the reference docs commands (apiserver-builder commands invoke these)
go get -u github.com/kubernetes-incubator/reference-docs/gen-apidocs/...

go get -u k8s.io/code-generator/...
# Install the code generation commands (apiserver-builder commands invoke these)
go install k8s.io/code-generator/cmd/openapi-gen
go install k8s.io/code-generator/cmd/deepcopy-gen
go install k8s.io/code-generator/cmd/informer-gen
```

## Update the `pkg/openapi/openapi_generated.go`

From the root kops directory run:

```sh
apiserver-boot build generated --generator openapi --copyright hack/boilerplate/boilerplate.go.txt
```

This will run the openapi-gen code generator and update the openapi definition.

**Note:** This will print out the generator command that is run, and it is possible to run this directly.

## Update `docs/apireference`

```sh
go install k8s.io/kops/cmd/kops-server
apiserver-boot build docs --disable-delegated-auth=false --output-dir docs/apireference --server kops-server
```

This will build the apiserver, get the openapi definitions, and write them to
`docs/apireference/openapi-spec/swagger.json`.  It will then generate the reference
documentation from the openapi and delete the open api.  To keep the swagger.json and
intermediate files, run with `--cleanup=false`

## Run a local kops apiserver

```sh
go install k8s.io/kops/cmd/kops-server
apiserver-boot run local --build=false --disable-delegated-auth=false --run=etcd --run=apiserver --apiserver=kops-server
```

This will build and run the apiserver and etcd locally, and create a kubeconfig for kubectl.

```sh
kubectl --kubeconfig kubeconfig api-versions
```

This will connect to the apiserver with kubectl and print the versions
