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
	"sort"
	"strings"

	secgroup "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/groups"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

// +kops:fitask
type Port struct {
	ID                       *string
	Name                     *string
	InstanceGroupName        *string
	Network                  *Network
	Subnets                  []*Subnet
	SecurityGroups           []*SecurityGroup
	AdditionalSecurityGroups []string
	Lifecycle                fi.Lifecycle
	Tags                     []string
	ForAPIServer             bool
	AllowedAddressPairs      []ports.AddressPair
}

// GetDependencies returns the dependencies of the Port task
func (e *Port) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	var deps []fi.CloudupTask
	for _, task := range tasks {
		if _, ok := task.(*Subnet); ok {
			deps = append(deps, task)
		}
		if _, ok := task.(*SecurityGroup); ok {
			deps = append(deps, task)
		}
		if _, ok := task.(*Network); ok {
			deps = append(deps, task)
		}
	}
	return deps
}

var _ fi.CompareWithID = &Port{}

func (s *Port) CompareWithID() *string {
	return s.ID
}

func (s *Port) FindAddresses(context *fi.CloudupContext) ([]string, error) {
	cloud := context.T.Cloud.(openstack.OpenstackCloud)
	if s.ID == nil {
		return nil, nil
	}
	port, err := cloud.GetPort(fi.ValueOf(s.ID))
	if err != nil {
		return nil, err
	}
	addrs := []string{}
	for _, addr := range port.FixedIPs {
		addrs = append(addrs, addr.IPAddress)
	}
	return addrs, nil
}

func (s *Port) IsForAPIServer() bool {
	return s.ForAPIServer
}

// getActualAllowedAddressPairs returns the actual allowed address pairs which kOps currently manages.
func getActualAllowedAddressPairs(port *ports.Port, find *Port) []ports.AddressPair {
	if find == nil {
		return port.AllowedAddressPairs
	}

	var allowedAddressPairs []ports.AddressPair
	for _, portAddressPair := range port.AllowedAddressPairs {
		// TODO: what if user set the macaddress in the config to the same one as the port?
		if portAddressPair.MACAddress == port.MACAddress {
			portAddressPair.MACAddress = ""
		}

		allowedAddressPairs = append(allowedAddressPairs, portAddressPair)
	}

	sort.Slice(allowedAddressPairs, func(i, j int) bool {
		return allowedAddressPairs[i].IPAddress < allowedAddressPairs[j].IPAddress
	})

	return allowedAddressPairs
}

func newPortTaskFromCloud(cloud openstack.OpenstackCloud, lifecycle fi.Lifecycle, port *ports.Port, find *Port) (*Port, error) {
	additionalSecurityGroupIDs := map[string]struct{}{}
	if find != nil {
		for _, sg := range find.AdditionalSecurityGroups {
			opt := secgroup.ListOpts{
				Name: sg,
			}
			gs, err := cloud.ListSecurityGroups(opt)
			if err != nil {
				continue
			}
			if len(gs) == 0 {
				continue
			}
			additionalSecurityGroupIDs[gs[0].ID] = struct{}{}
		}
	}
	sgs := []*SecurityGroup{}
	for _, sgid := range port.SecurityGroups {
		if _, ok := additionalSecurityGroupIDs[sgid]; ok {
			continue
		}
		sgs = append(sgs, &SecurityGroup{
			ID:        fi.PtrTo(sgid),
			Lifecycle: lifecycle,
		})
	}

	// sort for consistent comparison
	sort.Sort(SecurityGroupsByID(sgs))

	subnets := make([]*Subnet, len(port.FixedIPs))
	for i, subn := range port.FixedIPs {
		subnets[i] = &Subnet{
			ID:        fi.PtrTo(subn.SubnetID),
			Lifecycle: lifecycle,
		}
	}

	var tags []string

	if find != nil {
		for _, t := range find.Tags {
			if fi.ArrayContains(port.Tags, t) {
				tags = append(tags, t)
			}
		}
	} else {
		tags = port.Tags
	}

	var cloudInstanceGroupName *string
	for _, t := range port.Tags {
		prefix := fmt.Sprintf("%s=", openstack.TagKopsInstanceGroup)
		if !strings.HasPrefix(t, prefix) {
			continue
		}
		cloudInstanceGroupName = fi.PtrTo("")
		scanString := fmt.Sprintf("%s%%s", prefix)
		if _, err := fmt.Sscanf(t, scanString, cloudInstanceGroupName); err != nil {
			klog.V(2).Infof("Error extracting instance group for Port with name: %q", port.Name)
		}
	}

	actual := &Port{
		ID:                  fi.PtrTo(port.ID),
		InstanceGroupName:   cloudInstanceGroupName,
		Name:                fi.PtrTo(port.Name),
		Network:             &Network{ID: fi.PtrTo(port.NetworkID)},
		SecurityGroups:      sgs,
		Subnets:             subnets,
		Lifecycle:           lifecycle,
		Tags:                tags,
		AllowedAddressPairs: getActualAllowedAddressPairs(port, find),
	}
	if find != nil {
		find.ID = actual.ID
		actual.InstanceGroupName = find.InstanceGroupName
		actual.AdditionalSecurityGroups = find.AdditionalSecurityGroups
		actual.ForAPIServer = find.ForAPIServer
	}
	return actual, nil
}

func (s *Port) Find(context *fi.CloudupContext) (*Port, error) {
	cloud := context.T.Cloud.(openstack.OpenstackCloud)
	opt := ports.ListOpts{
		Name: fi.ValueOf(s.Name),
	}
	rs, err := cloud.ListPorts(opt)
	if err != nil {
		return nil, err
	}
	if rs == nil {
		return nil, nil
	} else if len(rs) != 1 {
		return nil, fmt.Errorf("found multiple ports with name: %s", fi.ValueOf(s.Name))
	}

	// sort for consistent comparison
	sort.Sort(SecurityGroupsByID(s.SecurityGroups))

	return newPortTaskFromCloud(cloud, s.Lifecycle, &rs[0], s)
}

func (s *Port) Run(context *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(s, context)
}

func (_ *Port) CheckChanges(a, e, changes *Port) error {
	if a == nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		if e.Network == nil {
			return fi.RequiredField("Network")
		}
	} else {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.Network != nil {
			return fi.CannotChangeField("Network")
		}
	}
	return nil
}

func (*Port) RenderOpenstack(t *openstack.OpenstackAPITarget, a, e, changes *Port) error {
	if a == nil {
		klog.V(2).Infof("Creating Port with name: %q", fi.ValueOf(e.Name))

		opt, err := portCreateOptsFromPortTask(t, a, e, changes)
		if err != nil {
			return fmt.Errorf("Error creating port cloud opts: %v", err)
		}

		v, err := t.Cloud.CreatePort(opt)
		if err != nil {
			return fmt.Errorf("Error creating port: %v", err)
		}

		if e.Tags != nil {
			for _, tag := range e.Tags {
				err = t.Cloud.AppendTag(openstack.ResourceTypePort, v.ID, tag)
				if err != nil {
					return fmt.Errorf("Error appending tag to port: %v", err)
				}
			}
		}
		e.ID = fi.PtrTo(v.ID)
		klog.V(2).Infof("Creating a new Openstack port, id=%s", v.ID)
		return nil
	}
	if changes != nil {
		if changes.Tags != nil {
			klog.V(2).Infof("Updating tags for Port with name: %q", fi.ValueOf(e.Name))
			for _, tag := range e.Tags {
				err := t.Cloud.AppendTag(openstack.ResourceTypePort, fi.ValueOf(a.ID), tag)
				if err != nil {
					return fmt.Errorf("Error appending tag to port: %v", err)
				}
			}
		}
		if changes.AllowedAddressPairs != nil {
			klog.V(2).Infof("Updating allowed address pairs for Port with name: %q", fi.ValueOf(e.Name))
			_, err := t.Cloud.UpdatePort(fi.ValueOf(a.ID), ports.UpdateOpts{
				AllowedAddressPairs: &e.AllowedAddressPairs,
			})
			if err != nil {
				return fmt.Errorf("error updating port: %v", err)
			}
		}
	}
	e.ID = a.ID
	klog.V(2).Infof("Using an existing Openstack port, id=%s", fi.ValueOf(e.ID))
	return nil
}

func portCreateOptsFromPortTask(t *openstack.OpenstackAPITarget, a, e, changes *Port) (ports.CreateOptsBuilder, error) {
	sgs := make([]string, len(e.SecurityGroups)+len(e.AdditionalSecurityGroups))
	for i, sg := range e.SecurityGroups {
		sgs[i] = fi.ValueOf(sg.ID)
	}
	for i, sg := range e.AdditionalSecurityGroups {
		opt := secgroup.ListOpts{
			Name: sg,
		}
		gs, err := t.Cloud.ListSecurityGroups(opt)
		if err != nil {
			continue
		}
		if len(gs) == 0 {
			return nil, fmt.Errorf("Additional SecurityGroup not found for name %s", sg)
		}
		sgs[i+len(e.SecurityGroups)] = gs[0].ID
	}
	fixedIPs := make([]ports.IP, len(e.Subnets))
	for i, subn := range e.Subnets {
		fixedIPs[i] = ports.IP{
			SubnetID: fi.ValueOf(subn.ID),
		}
	}

	return ports.CreateOpts{
		Name:                fi.ValueOf(e.Name),
		NetworkID:           fi.ValueOf(e.Network.ID),
		SecurityGroups:      &sgs,
		FixedIPs:            fixedIPs,
		AllowedAddressPairs: e.AllowedAddressPairs,
	}, nil
}
