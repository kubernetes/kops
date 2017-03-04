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

// FederationInterface has methods to work with Federation resources.
type FederationInterface interface {
	Create(*api.Federation) (*api.Federation, error)
	Update(*api.Federation) (*api.Federation, error)
	//UpdateStatus(*api.Federation) (*api.Federation, error)
	Delete(name string, options *metav1.DeleteOptions) error
	//DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error
	Get(name string) (*api.Federation, error)
	List(opts metav1.ListOptions) (*api.FederationList, error)
	//Watch(opts k8sapi.ListOptions) (watch.Interface, error)
	//Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *api.Federation, err error)
	//FederationExpansion
}
