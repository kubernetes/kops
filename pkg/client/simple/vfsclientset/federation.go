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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/v1alpha1"
	"k8s.io/kops/pkg/client/simple"
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
	return r
}

var _ simple.FederationInterface = &FederationVFS{}

func (c *FederationVFS) Get(name string) (*api.Federation, error) {
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
