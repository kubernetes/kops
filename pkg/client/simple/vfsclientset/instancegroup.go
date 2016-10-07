package vfsclientset

import (
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/upup/pkg/api"
	k8sapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"time"
	"fmt"
	"os"
	"github.com/golang/glog"
	"k8s.io/kops/util/pkg/vfs"
	"k8s.io/kops/upup/pkg/api/registry"
)

type InstanceGroupVFS struct {
	clusterBasePath vfs.Path
}

var _ simple.ClusterInterface = &ClusterVFS{}

func (c *InstanceGroupVFS) Get(name string) (*api.InstanceGroup, error) {
	group := &api.InstanceGroup{}
	err := registry.ReadConfig(c.clusterBasePath.Join("instancegroup", name), group)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading InstanceGroup %q: %v", name, err)
	}
	return group, nil
}

func (c *InstanceGroupVFS) List(options k8sapi.ListOptions) (*api.InstanceGroupList, error) {
	items, err := c.readAll()
	if err != nil {
		return nil, err
	}
	ret := &api.InstanceGroupList{}
	for _, i := range items {
		ret.Items = append(ret.Items, *i)
	}
	return ret, nil
}

func (c *InstanceGroupVFS) Create(g *api.InstanceGroup) (*api.InstanceGroup, error) {
	err := g.Validate(true)
	if err != nil {
		return nil, err
	}

	if g.CreationTimestamp.IsZero() {
		g.CreationTimestamp = unversioned.NewTime(time.Now().UTC())
	}

	err = registry.WriteConfig(c.clusterBasePath.Join("instancegroup",  g.Name), g, vfs.WriteOptionCreate)
	if err != nil {
		return nil, fmt.Errorf("error writing InstanceGroup: %v", err)
	}

	return g, nil
}

func (c *InstanceGroupVFS) Update(g *api.InstanceGroup) (*api.InstanceGroup, error) {
	err := g.Validate(true)
	if err != nil {
		return nil, err
	}

	err = registry.WriteConfig(c.clusterBasePath.Join("instancegroup",  g.Name), g, vfs.WriteOptionOnlyIfExists)
	if err != nil {
		return nil, fmt.Errorf("error writing InstanceGroup %q: %v", g.Name, err)
	}

	return g, nil
}

func (c *InstanceGroupVFS) Delete(name string, options *k8sapi.DeleteOptions) (error) {
	p := c.clusterBasePath.Join("instancegroup", name)
	err := p.Remove()
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("error deleting instancegroup configuration %q: %v", name, err)
	}
	return nil
}

func (c *InstanceGroupVFS) listNames() ([]string, error) {
	keys, err := listChildNames(c.clusterBasePath.Join("instancegroup"))
	if err != nil {
		return nil, fmt.Errorf("error listing instancegroups in state store: %v", err)
	}
	return keys, nil
}

func (r *InstanceGroupVFS) readAll() ([]*api.InstanceGroup, error) {
	names, err := r.listNames()
	if err != nil {
		return nil, err
	}

	var instancegroups []*api.InstanceGroup
	for _, name := range names {
		g, err := r.Get(name)
		if err != nil {
			return nil, err
		}

		if g == nil {
			glog.Warningf("InstanceGroup was listed, but then not found %q", name)
		}

		instancegroups = append(instancegroups, g)
	}

	return instancegroups, nil
}

