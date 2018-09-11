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

	"github.com/golang/glog"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/listeners"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/loadbalancers"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/pools"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

//go:generate fitask -type=LB
type LB struct {
	ID        *string
	Name      *string
	Subnet    *Subnet
	Listeners []listeners.Listener
	Lifecycle *fi.Lifecycle
}

var _ fi.CompareWithID = &LB{}

func (s *LB) CompareWithID() *string {
	return s.ID
}

func (s *LB) Find(context *fi.Context) (*LB, error) {
	if s.ID == nil {
		return nil, nil
	}

	cloud := context.Cloud.(openstack.OpenstackCloud)
	lb, err := loadbalancers.Get(cloud.LoadBalancerClient(), fi.StringValue(s.ID)).Extract()
	if err != nil {
		return nil, err
	}

	sub, err := subnets.Get(cloud.NetworkingClient(), fi.StringValue(s.Subnet.ID)).Extract()
	if err != nil {
		return nil, err
	}

	a := &LB{
		ID:        fi.String(lb.ID),
		Name:      fi.String(lb.Name),
		Listeners: lb.Listeners,
		Lifecycle: s.Lifecycle,
		Subnet: &Subnet{
			ID:   fi.String(sub.ID),
			Name: fi.String(sub.Name),
			CIDR: fi.String(sub.CIDR),
			Network: &Network{
				ID:        fi.String(sub.NetworkID),
				Lifecycle: s.Lifecycle,
			},
			Lifecycle: s.Lifecycle,
		},
	}
	return a, nil
}

func (s *LB) Run(context *fi.Context) error {
	return fi.DefaultDeltaRunMethod(s, context)
}

func (_ *LB) CheckChanges(a, e, changes *LB) error {
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

func (_ *LB) RenderOpenstack(t *openstack.OpenstackAPITarget, a, e, changes *LB) error {
	if a == nil {
		glog.V(2).Infof("Creating LB with Name: %q", fi.StringValue(e.Name))

		var vipsubid string
		if subid := fi.StringValue(e.Subnet.ID); subid != "" {
			vipsubid = subid
		} else {
			subid, err := subnets.IDFromName(t.Cloud.NetworkingClient(), fi.StringValue(e.Subnet.Name))
			if err != nil {
				return err
			}
			vipsubid = subid
		}

		lbopts := loadbalancers.CreateOpts{
			Name:        fi.StringValue(e.Name),
			VipSubnetID: vipsubid,
		}
		lb, err := t.Cloud.CreateLB(lbopts)
		if err != nil {
			return fmt.Errorf("error creating LB: %v", err)
		}
		e.ID = fi.String(lb.ID)

		poolopts := pools.CreateOpts{
			Name:           lb.Name + "-https",
			LBMethod:       pools.LBMethodRoundRobin,
			Protocol:       pools.ProtocolTCP,
			LoadbalancerID: lb.ID,
		}
		pool, err := pools.Create(t.Cloud.LoadBalancerClient(), poolopts).Extract()
		if err != nil {
			return fmt.Errorf("error creating LB pool: %v", err)
		}

		listeneropts := listeners.CreateOpts{
			Name:           lb.Name + "-https",
			DefaultPoolID:  pool.ID,
			LoadbalancerID: lb.ID,
			Protocol:       listeners.ProtocolTCP,
			ProtocolPort:   443,
		}
		_, err = listeners.Create(t.Cloud.LoadBalancerClient(), listeneropts).Extract()
		if err != nil {
			return fmt.Errorf("error creating LB listener: %v", err)
		}

		return nil
	}

	glog.V(2).Infof("Openstack task LB::RenderOpenstack did nothing")
	return nil
}
