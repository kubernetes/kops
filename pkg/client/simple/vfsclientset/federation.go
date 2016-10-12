package vfsclientset

import (
	"k8s.io/kops/pkg/client/simple"
	api "k8s.io/kops/pkg/apis/kops"
	k8sapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kops/pkg/apis/kops/v1alpha1"
)

type FederationVFS struct {
	commonVFS
}

func newFederationVFS(c *VFSClientset) *FederationVFS {
	key := "_federation"

	r := &FederationVFS{}
	r.init(key, c.basePath.Join(key), v1alpha1.SchemeGroupVersion)
	return r
}
var _ simple.FederationInterface = &FederationVFS{}

func (c *FederationVFS) Get(name string) (*api.Federation, error) {
	v := &api.Federation{}
	found, err := c.get(name, v)
	if err != nil {
		return nil ,err
	}
	if !found {
		return nil, nil
	}
	return v, nil
}

func (c *FederationVFS) List(options k8sapi.ListOptions) (*api.FederationList, error) {
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

func (c *FederationVFS) Delete(name string, options *k8sapi.DeleteOptions) (error) {
	return c.delete(name, options)
}

