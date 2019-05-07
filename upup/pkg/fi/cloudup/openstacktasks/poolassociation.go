/*
Copyright 2017 The Kubernetes Authors.

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

package openstacktasks

import (
	"fmt"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"

	v2pools "github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/pools"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

//go:generate fitask -type=PoolAssociation
type PoolAssociation struct {
	ID            *string
	Name          *string
	Lifecycle     *fi.Lifecycle
	Pool          *LBPool
	ServerGroup   *ServerGroup
	InterfaceName *string
	ProtocolPort  *int
}

// GetDependencies returns the dependencies of the Instance task
func (e *PoolAssociation) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task
	for _, task := range tasks {
		if _, ok := task.(*LB); ok {
			deps = append(deps, task)
		}
		if _, ok := task.(*LBPool); ok {
			deps = append(deps, task)
		}
		if _, ok := task.(*Instance); ok {
			deps = append(deps, task)
		}
	}
	return deps
}

var _ fi.CompareWithID = &PoolAssociation{}

func (s *PoolAssociation) CompareWithID() *string {
	return s.ID
}

func (p *PoolAssociation) Find(context *fi.Context) (*PoolAssociation, error) {
	cloud := context.Cloud.(openstack.OpenstackCloud)

	opt := v2pools.ListOpts{
		Name: fi.StringValue(p.Pool.Name),
		ID:   fi.StringValue(p.Pool.ID),
	}

	rs, err := cloud.ListPools(opt)
	if err != nil {
		return nil, err
	}
	if rs == nil {
		return nil, nil
	} else if len(rs) != 1 {
		return nil, fmt.Errorf("found multiple pools with name: %s", fi.StringValue(p.Pool.Name))
	}

	a := rs[0]
	// check is member already created
	found := false
	for _, member := range a.Members {
		poolMember, err := cloud.GetPool(a.ID, member.ID)
		if err != nil {
			return nil, err
		}
		if fi.StringValue(p.Name) == poolMember.Name {
			found = true
			break
		}
	}
	// if not found it is created by returning nil, nil
	// this is needed for instance in initial installation
	if !found {
		return nil, nil
	}
	pool, err := NewLBPoolTaskFromCloud(cloud, p.Lifecycle, &a, nil)
	if err != nil {
		return nil, fmt.Errorf("NewLBListenerTaskFromCloud: failed to fetch pool %s: %v", fi.StringValue(pool.Name), err)
	}

	actual := &PoolAssociation{
		ID:            p.ID,
		Name:          p.Name,
		Pool:          pool,
		ServerGroup:   p.ServerGroup,
		InterfaceName: p.InterfaceName,
		ProtocolPort:  p.ProtocolPort,
		Lifecycle:     p.Lifecycle,
	}
	p.ID = actual.ID
	return actual, nil
}

func (s *PoolAssociation) Run(context *fi.Context) error {
	return fi.DefaultDeltaRunMethod(s, context)
}

func (_ *PoolAssociation) CheckChanges(a, e, changes *PoolAssociation) error {
	if a == nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	} else {
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
	}
	return nil
}

func (_ *PoolAssociation) RenderOpenstack(t *openstack.OpenstackAPITarget, a, e, changes *PoolAssociation) error {
	if a == nil {

		for _, serverID := range e.ServerGroup.Members {
			server, err := servers.Get(t.Cloud.ComputeClient(), serverID).Extract()
			if err != nil {
				return fmt.Errorf("Failed to find server with id `%s`: %v", serverID, err)
			}

			memberAddress, err := openstack.GetServerFixedIP(server, fi.StringValue(e.InterfaceName))

			if err != nil {
				return fmt.Errorf("Failed to get fixed ip for associated pool: %v", err)
			}

			member, err := t.Cloud.AssociateToPool(server, fi.StringValue(e.Pool.ID), v2pools.CreateMemberOpts{
				Name:         fi.StringValue(e.Name),
				ProtocolPort: fi.IntValue(e.ProtocolPort),
				SubnetID:     fi.StringValue(e.Pool.Loadbalancer.VipSubnet),
				Address:      memberAddress,
			})
			if err != nil {
				return fmt.Errorf("Failed to create member: %v", err)
			}
			e.ID = fi.String(member.ID)
		}
		return nil
	} else {
		//TODO: Update Member, this is covered as `a` will always be nil
		klog.V(2).Infof("Openstack task PoolAssociation::RenderOpenstack Update not implemented!")
	}

	klog.V(2).Infof("Openstack task PoolAssociation::RenderOpenstack did nothing")
	return nil
}
