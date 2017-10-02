/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package miskatonic

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
)

//var _ rest.CreaterUpdater = NewStudentREST()
//var _ rest.Patcher = NewStudentREST()

var _ rest.CreaterUpdater = &StudentREST{}
var _ rest.Patcher = &StudentREST{}

// +k8s:deepcopy-gen=false
type StudentREST struct {
	Registry StudentRegistry
}

func (r *StudentREST) Create(ctx request.Context, obj runtime.Object, includeUninitialized bool) (runtime.Object, error) {
	s := obj.(*Student)
	s.Spec.ID = s.Spec.ID + 1
	return s, nil
}

// Get retrieves the object from the storage. It is required to support Patch.
func (r *StudentREST) Get(ctx request.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return &Student{}, nil
}

// Update alters the status subset of an object.
func (r *StudentREST) Update(ctx request.Context, name string, objInfo rest.UpdatedObjectInfo) (runtime.Object, bool, error) {
	return nil, false, nil
}

func (r *StudentREST) New() runtime.Object {
	return &Student{}
}
