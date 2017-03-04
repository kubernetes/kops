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

package simple

import (
	api "k8s.io/kops/pkg/apis/kops"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InstanceGroupInterface has methods to work with InstanceGroup resources.
type InstanceGroupInterface interface {
	Create(*api.InstanceGroup) (*api.InstanceGroup, error)
	Update(*api.InstanceGroup) (*api.InstanceGroup, error)
	//UpdateStatus(*api.InstanceGroup) (*api.InstanceGroup, error)
	Delete(name string, options *metav1.DeleteOptions) error
	//DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error
	Get(name string) (*api.InstanceGroup, error)
	List(opts metav1.ListOptions) (*api.InstanceGroupList, error)
	//Watch(opts k8sapi.ListOptions) (watch.Interface, error)
	//Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *api.InstanceGroup, err error)
	//InstanceGroupExpansion
}
