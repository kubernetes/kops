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

package vfsclientset

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/v1alpha1"
	"k8s.io/kops/pkg/apis/kops/validation"
	kopsinternalversion "k8s.io/kops/pkg/client/clientset_generated/clientset/typed/kops/internalversion"
)

type FederationVFS struct {
	commonVFS
}

func newFederationVFS(c *VFSClientset) *FederationVFS {
	kind := "Federation"

	r := &FederationVFS{}
	r.init(kind, c.basePath.Join("_federation"), StoreVersion)
	defaultReadVersion := v1alpha1.SchemeGroupVersion.WithKind(kind)
	r.defaultReadVersion = &defaultReadVersion
	r.validate = func(o runtime.Object) error {
		return validation.ValidateFederation(o.(*api.Federation))
	}
	return r
}

var _ kopsinternalversion.FederationInterface = &FederationVFS{}

func (c *FederationVFS) Get(name string, options metav1.GetOptions) (*api.Federation, error) {
	if options.ResourceVersion != "" {
		return nil, fmt.Errorf("ResourceVersion not supported in FederationVFS::Get")
	}
	o, err := c.get(name)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, nil
	}
	return o.(*api.Federation), nil
}

func (c *FederationVFS) List(options metav1.ListOptions) (*api.FederationList, error) {
	list := &api.FederationList{}
	items, err := c.list(list.Items, options)
	if err != nil {
		return nil, err
	}
	list.Items = items.([]api.Federation)
	return list, nil
}

func (c *FederationVFS) Create(g *api.Federation) (*api.Federation, error) {
	err := c.create(g)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (c *FederationVFS) Update(g *api.Federation) (*api.Federation, error) {
	err := c.update(g)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (c *FederationVFS) Delete(name string, options *metav1.DeleteOptions) error {
	return c.delete(name, options)
}

func (r *FederationVFS) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return fmt.Errorf("FederationVFS DeleteCollection not implemented for vfs store")
}

func (r *FederationVFS) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return nil, fmt.Errorf("FederationVFS Watch not implemented for vfs store")
}

func (r *FederationVFS) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *api.Federation, err error) {
	return nil, fmt.Errorf("FederationVFS Patch not implemented for vfs store")
}
