/*
Copyright 2016 The Kubernetes Authors.

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

package kops

import (
	"k8s.io/apimachinery/pkg/apimachinery/announced"
	"k8s.io/apimachinery/pkg/apimachinery/registered"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"os"
)

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

// GroupFactoryRegistry is the APIGroupFactoryRegistry (overlaps a bit with Registry, see comments in package for details)
var GroupFactoryRegistry = make(announced.APIGroupFactoryRegistry)

// Registry is an instance of an API registry.  This is an interim step to start removing the idea of a global
// API registry.
var Registry = registered.NewOrDie(os.Getenv("KOPS_API_VERSIONS"))

// Scheme is the default instance of runtime.Scheme to which types in the Kops API are already registered.
var Scheme = runtime.NewScheme()

// GroupName is the group name use in this package
const GroupName = "kops"

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: runtime.APIVersionInternal}

// Kind takes an unqualified kind and returns a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Cluster{},
		&ClusterList{},
		&InstanceGroup{},
		&InstanceGroupList{},
		&Federation{},
		&FederationList{},
	)
	return nil
}

func (obj *Cluster) GetObjectKind() schema.ObjectKind {
	return &obj.TypeMeta
}
func (obj *InstanceGroup) GetObjectKind() schema.ObjectKind {
	return &obj.TypeMeta
}
func (obj *Federation) GetObjectKind() schema.ObjectKind {
	return &obj.TypeMeta
}
