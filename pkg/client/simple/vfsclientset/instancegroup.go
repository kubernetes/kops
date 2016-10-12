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

	key := "instancegroup"

	r := &InstanceGroupVFS{}
	r.init(key, c.basePath.Join(clusterName, key), v1alpha1.SchemeGroupVersion)
	return r
}

var _ simple.InstanceGroupInterface = &InstanceGroupVFS{}

func (c *InstanceGroupVFS) Get(name string) (*api.InstanceGroup, error) {
	v := &api.InstanceGroup{}
	found, err := c.get(name, v)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return v, nil
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
