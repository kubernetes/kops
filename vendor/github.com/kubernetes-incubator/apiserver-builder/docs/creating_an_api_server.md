# Creating an API Server

**Note:** This document explains how to manually create the files generated
by `apiserver-boot`.  It is recommended to automatically create these files
instead.  See [getting started](https://github.com/kubernetes-incubator/apiserver-builder/blob/master/docs/getting_started.md)
for more details.

## Create the apiserver command

Create a file called `main.go` at the root of your project.  This
file bootstraps the apiserver by invoking the apiserver start function
with the generated API code.

*Location:* `GOPATH/src/YOUR/GO/PACKAGE/main.go`

```go
package main

import (
	"github.com/kubernetes-incubator/apiserver-builder/pkg/cmd/server"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // Enable cloud provider auth

	"YOUR/GO/PACKAGE/pkg/apis" // Change this
	"YOUR/GO/PACKAGE/pkg/openapi" // Change this
)

const storagePath = "/registry/YOUR.DOMAIN" // Change this

func main() {
	server.StartApiServer(storagePath, apis.GetAllApiBuilders(), openapi.GetOpenAPIDefinitions)
}

```

## Create the API root package

Create your API root under `pkg/apis`

*Location:*  `GOPATH/src/YOUR/GO/PACKAGE/pkg/apis/doc.go`

- Change `YOUR.DOMAIN` to the domain you want your API groups to appear under.

```go
// +domain=YOUR.DOMAIN

package apis
```

## Create an API group

Create your API group under `pkg/apis/GROUP`

*Location:* `GOPATH/src/YOUR/GO/PACKAGE/pkg/apis/GROUP/doc.go`

- Change GROUP to be the group name.
- Change YOUR.DOMAIN to be your domain.

```go
// +k8s:deepcopy-gen=package,register
// +groupName=GROUP.YOUR.DOMAIN

// Package api is the internal version of the API.
package GROUP
```

## Create an API version

Create your API group under `pkg/apis/GROUP/VERSION`

*Location:* `GOPATH/src/YOUR/GO/PACKAGE/pkg/apis/GROUP/VERSION/doc.go`

- Change GROUP to be the group name.
- Change VERSION to the be the api version name.
- Change YOUR/GO/PACKAGE to be the go package of you project.

```go
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=package,register
// +k8s:conversion-gen=YOUR/GO/PACKAGE/pkg/apis/GROUP
// +k8s:defaulter-gen=TypeMeta

// +groupName=GROUP.VERSION
package VERSION // import "YOUR/GO/PACKAGE/pkg/apis/GROUP/VERSION"
```

## Create the API type definitions

## Generate the code

## Create an integration test

## Start the server locally