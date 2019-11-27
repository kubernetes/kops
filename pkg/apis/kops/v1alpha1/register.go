/*
Copyright 2019 The Kubernetes Authors.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	SchemeBuilder      runtime.SchemeBuilder
	localSchemeBuilder = &SchemeBuilder
	AddToScheme        = localSchemeBuilder.AddToScheme
)

func init() {
	// We only register manually written functions here. The registration of the
	// generated functions takes place in the generated files. The separation
	// makes the code compile even when the generated files are missing.
	localSchemeBuilder.Register(addKnownTypes, addDefaultingFuncs, addConversionFuncs)
}

// GroupName is the group name use in this package
const GroupName = "kops.k8s.io"

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}

//// Kind takes an unqualified kind and returns a Group qualified GroupKind
//func Kind(kind string) schema.GroupKind {
//	return SchemeGroupVersion.WithKind(kind).GroupKind()
//}
//
//// Resource takes an unqualified resource and returns a Group qualified GroupResource
//func Resource(resource string) schema.GroupResource {
//	return SchemeGroupVersion.WithResource(resource).GroupResource()
//}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Cluster{},
		&ClusterList{},
		&InstanceGroup{},
		&InstanceGroupList{},
		&SSHCredential{},
		&SSHCredentialList{},
	)

	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)

	return nil
}

func (obj *Cluster) GetObjectKind() schema.ObjectKind {
	return &obj.TypeMeta
}
func (obj *InstanceGroup) GetObjectKind() schema.ObjectKind {
	return &obj.TypeMeta
}
func (obj *SSHCredential) GetObjectKind() schema.ObjectKind {
	return &obj.TypeMeta
}

func addConversionFuncs(scheme *runtime.Scheme) error {
	// Add non-generated conversion functions
	err := scheme.AddConversionFuncs(
		Convert_v1alpha1_BastionSpec_To_kops_BastionSpec,
		Convert_kops_BastionSpec_To_v1alpha1_BastionSpec,

		Convert_v1alpha1_ClusterSpec_To_kops_ClusterSpec,
		Convert_kops_ClusterSpec_To_v1alpha1_ClusterSpec,

		Convert_v1alpha1_EtcdMemberSpec_To_kops_EtcdMemberSpec,
		Convert_kops_EtcdMemberSpec_To_v1alpha1_EtcdMemberSpec,

		Convert_v1alpha1_InstanceGroupSpec_To_kops_InstanceGroupSpec,
		Convert_kops_InstanceGroupSpec_To_v1alpha1_InstanceGroupSpec,

		Convert_v1alpha1_TopologySpec_To_kops_TopologySpec,
		Convert_kops_TopologySpec_To_v1alpha1_TopologySpec,
	)
	if err != nil {
		return err
	}

	return nil
}
