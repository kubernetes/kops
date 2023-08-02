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

	"github.com/gophercloud/gophercloud"
	v2pools "github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/pools"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/kops/util/pkg/vfs"
)

// +kops:fitask
type PoolAssociation struct {
	ID            *string
	Name          *string
	ServerPrefix  *string
	Lifecycle     fi.Lifecycle
	Pool          *LBPool
	InterfaceName *string
	ProtocolPort  *int
	Weight        *int
}

// GetDependencies returns the dependencies of the Instance task
func (e *PoolAssociation) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	var deps []fi.CloudupTask
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

func (p *PoolAssociation) Find(context *fi.CloudupContext) (*PoolAssociation, error) {
	cloud := context.T.Cloud.(openstack.OpenstackCloud)

	opt := v2pools.ListOpts{
		Name: fi.ValueOf(p.Pool.Name),
		ID:   fi.ValueOf(p.Pool.ID),
	}

	rs, err := cloud.ListPools(opt)
	if err != nil {
		return nil, err
	}
	if rs == nil {
		return nil, nil
	} else if len(rs) != 1 {
		return nil, fmt.Errorf("found multiple pools with name: %s", fi.ValueOf(p.Pool.Name))
	}

	a := rs[0]
	// check is member already created
	var found *v2pools.Member
	for _, member := range a.Members {
		poolMember, err := cloud.GetPoolMember(a.ID, member.ID)
		if err != nil {
			return nil, err
		}
		if fi.ValueOf(p.Name) == poolMember.Name {
			found = poolMember
			break
		}
	}
	// if not found it is created by returning nil, nil
	// this is needed for instance in initial installation
	if found == nil {
		return nil, nil
	}
	pool, err := NewLBPoolTaskFromCloud(cloud, p.Lifecycle, &a, nil)
	if err != nil {
		return nil, fmt.Errorf("NewLBListenerTaskFromCloud: failed to fetch pool %s: %v", fi.ValueOf(pool.Name), err)
	}

	actual := &PoolAssociation{
		ID:            fi.PtrTo(found.ID),
		Name:          fi.PtrTo(found.Name),
		Pool:          pool,
		ServerPrefix:  p.ServerPrefix,
		InterfaceName: p.InterfaceName,
		ProtocolPort:  p.ProtocolPort,
		Lifecycle:     p.Lifecycle,
		Weight:        fi.PtrTo(found.Weight),
	}
	p.ID = actual.ID
	return actual, nil
}

func (s *PoolAssociation) Run(context *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(s, context)
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

func GetServerFixedIP(client *gophercloud.ServiceClient, server *servers.Server, interfaceName string) (memberAddress string, err error) {
	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		memberAddress, err = openstack.GetServerFixedIP(server, interfaceName)
		if err != nil {
			// sometimes provisioning interfaces is slow, that is why we need retry the interface from the server
			return false, fmt.Errorf("Failed to get fixed ip for associated pool: %v", err)
		}
		return true, nil
	})
	if done {
		return memberAddress, nil
	}
	return memberAddress, err
}

func (_ *PoolAssociation) RenderOpenstack(t *openstack.OpenstackAPITarget, a, e, changes *PoolAssociation) error {
	if a == nil {
		serverPage, err := servers.List(t.Cloud.ComputeClient(), servers.ListOpts{
			Name: fmt.Sprintf("^%s", fi.ValueOf(e.ServerPrefix)),
		}).AllPages()
		if err != nil {
			return fmt.Errorf("error listing servers: %v", err)
		}
		serverList, err := servers.ExtractServers(serverPage)
		if err != nil {
			return fmt.Errorf("error extracting server page: %v", err)
		}

		for _, server := range serverList {
			memberAddress, err := GetServerFixedIP(t.Cloud.ComputeClient(), &server, fi.ValueOf(e.InterfaceName))
			if err != nil {
				return err
			}

			member, err := t.Cloud.AssociateToPool(&server, fi.ValueOf(e.Pool.ID), v2pools.CreateMemberOpts{
				Name:         fi.ValueOf(e.Name),
				ProtocolPort: fi.ValueOf(e.ProtocolPort),
				SubnetID:     fi.ValueOf(e.Pool.Loadbalancer.VipSubnet),
				Address:      memberAddress,
			})
			if err != nil {
				return fmt.Errorf("Failed to create member: %v", err)
			}
			e.ID = fi.PtrTo(member.ID)
		}
	} else {
		_, err := t.Cloud.UpdateMemberInPool(fi.ValueOf(a.Pool.ID), fi.ValueOf(a.ID), v2pools.UpdateMemberOpts{
			Weight: e.Weight,
		})
		if err != nil {
			return fmt.Errorf("Failed to update member: %v", err)
		}
	}
	return nil
}
