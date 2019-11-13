/*
Copyright 2019 The Kubernetes Authors.

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

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

//go:generate fitask -type=Subnet
type Subnet struct {
	ID         *string
	Name       *string
	Network    *Network
	CIDR       *string
	DNSServers []*string
	Tag        *string
	Lifecycle  *fi.Lifecycle
}

// GetDependencies returns the dependencies of the Port task
func (e *Subnet) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task
	for _, task := range tasks {
		if _, ok := task.(*Network); ok {
			deps = append(deps, task)
		}
	}
	return deps
}

var _ fi.CompareWithID = &Subnet{}

func (s *Subnet) CompareWithID() *string {
	return s.ID
}

func NewSubnetTaskFromCloud(cloud openstack.OpenstackCloud, lifecycle *fi.Lifecycle, subnet *subnets.Subnet, find *Subnet) (*Subnet, error) {
	network, err := cloud.GetNetwork(subnet.NetworkID)
	if err != nil {
		return nil, fmt.Errorf("NewSubnetTaskFromCloud: Failed to get network with ID %s: %v", subnet.NetworkID, err)
	}
	networkTask, err := NewNetworkTaskFromCloud(cloud, lifecycle, network, find.Tag)
	if err != nil {
		return nil, fmt.Errorf("error creating network task from cloud: %v", err)
	}

	nameservers := make([]*string, len(subnet.DNSNameservers))
	for i, ns := range subnet.DNSNameservers {
		nameservers[i] = fi.String(ns)
	}

	tag := ""
	if find != nil && fi.ArrayContains(subnet.Tags, fi.StringValue(find.Tag)) {
		tag = fi.StringValue(find.Tag)
	}

	actual := &Subnet{
		ID:         fi.String(subnet.ID),
		Name:       fi.String(subnet.Name),
		Network:    networkTask,
		CIDR:       fi.String(subnet.CIDR),
		Lifecycle:  lifecycle,
		DNSServers: nameservers,
		Tag:        fi.String(tag),
	}
	if find != nil {
		find.ID = actual.ID
	}
	return actual, nil
}

func (s *Subnet) Find(context *fi.Context) (*Subnet, error) {
	cloud := context.Cloud.(openstack.OpenstackCloud)
	opt := subnets.ListOpts{
		ID:         fi.StringValue(s.ID),
		Name:       fi.StringValue(s.Name),
		NetworkID:  fi.StringValue(s.Network.ID),
		CIDR:       fi.StringValue(s.CIDR),
		EnableDHCP: fi.Bool(true),
		IPVersion:  4,
	}
	rs, err := cloud.ListSubnets(opt)
	if err != nil {
		return nil, err
	}
	if rs == nil {
		return nil, nil
	} else if len(rs) != 1 {
		return nil, fmt.Errorf("found multiple subnets with name: %s", fi.StringValue(s.Name))
	}
	return NewSubnetTaskFromCloud(cloud, s.Lifecycle, &rs[0], s)
}

func (s *Subnet) Run(context *fi.Context) error {
	return fi.DefaultDeltaRunMethod(s, context)
}

func (_ *Subnet) CheckChanges(a, e, changes *Subnet) error {
	if a == nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		if e.Network == nil {
			return fi.RequiredField("Network")
		}
		if e.CIDR == nil {
			return fi.RequiredField("CIDR")
		}
	} else {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.DNSServers != nil {
			return fi.CannotChangeField("DNSServers")
		}
		if changes.Network != nil {
			return fi.CannotChangeField("Network")
		}
		if changes.CIDR != nil {
			return fi.CannotChangeField("CIDR")
		}
	}
	return nil
}

func (_ *Subnet) RenderOpenstack(t *openstack.OpenstackAPITarget, a, e, changes *Subnet) error {
	if a == nil {
		klog.V(2).Infof("Creating Subnet with name:%q", fi.StringValue(e.Name))

		opt := subnets.CreateOpts{
			Name:       fi.StringValue(e.Name),
			NetworkID:  fi.StringValue(e.Network.ID),
			IPVersion:  gophercloud.IPv4,
			CIDR:       fi.StringValue(e.CIDR),
			EnableDHCP: fi.Bool(true),
		}

		if len(e.DNSServers) > 0 {
			dnsNameSrv := make([]string, len(e.DNSServers))
			for i, ns := range e.DNSServers {
				dnsNameSrv[i] = fi.StringValue(ns)
			}
			opt.DNSNameservers = dnsNameSrv
		}
		v, err := t.Cloud.CreateSubnet(opt)
		if err != nil {
			return fmt.Errorf("Error creating subnet: %v", err)
		}

		err = t.Cloud.AppendTag(openstack.ResourceTypeSubnet, v.ID, fi.StringValue(e.Tag))
		if err != nil {
			return fmt.Errorf("Error appending tag to subnet: %v", err)
		}

		e.ID = fi.String(v.ID)
		klog.V(2).Infof("Creating a new Openstack subnet, id=%s", v.ID)
		return nil
	} else {
		err := t.Cloud.AppendTag(openstack.ResourceTypeSubnet, fi.StringValue(a.ID), fi.StringValue(changes.Tag))
		if err != nil {
			return fmt.Errorf("Error appending tag to subnet: %v", err)
		}
	}
	e.ID = a.ID
	klog.V(2).Infof("Using an existing Openstack subnet, id=%s", fi.StringValue(e.ID))
	return nil
}
