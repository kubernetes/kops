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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/v1alpha1"
	"k8s.io/kops/pkg/client/simple"
)

type InstanceGroupVFS struct {
	commonVFS

	clusterName string
}

func newInstanceGroupVFS(c *VFSClientset, clusterName string) *InstanceGroupVFS {
	if clusterName == "" {
		glog.Fatalf("clusterName is required")
	}

	kind := "InstanceGroup"

	r := &InstanceGroupVFS{
		clusterName: clusterName,
	}
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

	ig := o.(*api.InstanceGroup)
	c.addLabels(ig)

	return ig, nil
}

func (c *InstanceGroupVFS) addLabels(ig *api.InstanceGroup) {
	if ig.ObjectMeta.Labels == nil {
		ig.ObjectMeta.Labels = make(map[string]string)
	}
	ig.ObjectMeta.Labels[api.LabelClusterName] = c.clusterName
}

func (c *InstanceGroupVFS) List(options metav1.ListOptions) (*api.InstanceGroupList, error) {
	list := &api.InstanceGroupList{}
	items, err := c.list(list.Items, options)
	if err != nil {
		return nil, err
	}
	list.Items = items.([]api.InstanceGroup)
	for i := range list.Items {
		c.addLabels(&list.Items[i])
	}
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

func (c *InstanceGroupVFS) Delete(name string, options *metav1.DeleteOptions) error {
	return c.delete(name, options)
}
