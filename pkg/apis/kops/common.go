package kops

import (
	"k8s.io/kubernetes/pkg/runtime"
)

// ApiType adds a Validate() method to runtime.Object
// TODO: use the real Validation infrastructure here
type ApiType interface {
	runtime.Object

	Validate() error
}
