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
	"github.com/golang/glog"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/v1alpha1"
	"k8s.io/kops/pkg/client/simple"
	k8sapi "k8s.io/kubernetes/pkg/api"
)

type InstanceGroupVFS struct {
	commonVFS
}

func newInstanceGroupVFS(c *VFSClientset, clusterName string) *InstanceGroupVFS {
	if clusterName == "" {
		glog.Fatalf("clusterName is required")
	}

	kind := "InstanceGroup"

	r := &InstanceGroupVFS{}
	r.init(kind, c.basePath.Join(clusterName, "instancegroup"), StoreVersion)
	defaultReadVersion := v1alpha1.SchemeGroupVersion.WithKind(kind)
	r.defaultReadVersion = &defaultReadVersion
	return r
}

var _ simple.InstanceGroupInterface = &InstanceGroupVFS{}

func (c *InstanceGroupVFS) Get(name string) (*api.InstanceGroup, error) {
	o, err := c.get(name)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, nil
	}
	return o.(*api.InstanceGroup), nil
}

func (c *InstanceGroupVFS) List(options k8sapi.ListOptions) (*api.InstanceGroupList, error) {
	list := &api.InstanceGroupList{}
	items, err := c.list(list.Items, options)
	if err != nil {
		return nil, err
	}
	list.Items = items.([]api.InstanceGroup)
	return list, nil
}

func (c *InstanceGroupVFS) Create(g *api.InstanceGroup) (*api.InstanceGroup, error) {
	err := c.create(g)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (c *InstanceGroupVFS) Update(g *api.InstanceGroup) (*api.InstanceGroup, error) {
	err := c.update(g)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (c *InstanceGroupVFS) Delete(name string, options *k8sapi.DeleteOptions) error {
	return c.delete(name, options)
}
