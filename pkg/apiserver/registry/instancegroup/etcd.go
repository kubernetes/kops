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

package instancegroup

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"

	"k8s.io/kops/pkg/apis/kops"
)

// rest implements a RESTStorage for kops InstanceGroups against etcd
type REST struct {
	*genericregistry.Store
}

// NewREST returns a RESTStorage object that will work against kops InstanceGroups.
func NewREST(optsGetter generic.RESTOptionsGetter) *REST {
	store := &genericregistry.Store{
		Copier: kops.Scheme,
		NewFunc: func() runtime.Object {
			return &kops.InstanceGroup{}
		},
		NewListFunc: func() runtime.Object {
			return &kops.InstanceGroupList{}
		},
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return obj.(*kops.InstanceGroup).Name, nil
		},
		PredicateFunc:     MatchInstanceGroup,
		QualifiedResource: kops.Resource("instancegroups"),

		CreateStrategy: Strategy,
		UpdateStrategy: Strategy,
		DeleteStrategy: Strategy,
	}
	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		panic(err) // TODO: Propagate error up
	}
	return &REST{store}
}
