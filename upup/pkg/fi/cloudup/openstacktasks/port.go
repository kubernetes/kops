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

	secgroup "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/groups"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

//go:generate fitask -type=Port
type Port struct {
	ID                       *string
	Name                     *string
	Network                  *Network
	Subnets                  []*Subnet
	SecurityGroups           []*SecurityGroup
	AdditionalSecurityGroups []string
	Lifecycle                *fi.Lifecycle
	Tag                      *string
}

// GetDependencies returns the dependencies of the Port task
func (e *Port) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task
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

func NewPortTaskFromCloud(cloud openstack.OpenstackCloud, lifecycle *fi.Lifecycle, port *ports.Port, find *Port) (*Port, error) {
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
			ID:        fi.String(sgid),
			Lifecycle: lifecycle,
		})
	}
	subnets := make([]*Subnet, len(port.FixedIPs))
	for i, subn := range port.FixedIPs {
		subnets[i] = &Subnet{
			ID:        fi.String(subn.SubnetID),
			Lifecycle: lifecycle,
		}
	}

	tag := ""
	if find != nil && fi.ArrayContains(port.Tags, fi.StringValue(find.Tag)) {
		tag = fi.StringValue(find.Tag)
	}

	actual := &Port{
		ID:             fi.String(port.ID),
		Name:           fi.String(port.Name),
		Network:        &Network{ID: fi.String(port.NetworkID)},
		SecurityGroups: sgs,
		Subnets:        subnets,
		Lifecycle:      lifecycle,
		Tag:            fi.String(tag),
	}
	if find != nil {
		find.ID = actual.ID
		actual.AdditionalSecurityGroups = find.AdditionalSecurityGroups
	}
	return actual, nil
}

func (s *Port) Find(context *fi.Context) (*Port, error) {
	cloud := context.Cloud.(openstack.OpenstackCloud)
	opt := ports.ListOpts{
		Name: fi.StringValue(s.Name),
	}
	rs, err := cloud.ListPorts(opt)
	if err != nil {
		return nil, err
	}
	filteredList := []ports.Port{}
	for _, port := range rs {
		if fi.ArrayContains(port.Tags, openstack.TagNameDetach) {
			continue
		}
		filteredList = append(filteredList, port)
	}
	if len(filteredList) == 0 {
		return nil, nil
	} else if len(filteredList) != 1 {
		return nil, fmt.Errorf("found multiple ports with name: %s", fi.StringValue(s.Name))
	}
	return NewPortTaskFromCloud(cloud, s.Lifecycle, &filteredList[0], s)
}

func (s *Port) Run(context *fi.Context) error {
	return fi.DefaultDeltaRunMethod(s, context)
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
		klog.V(2).Infof("Creating Port with name: %q", fi.StringValue(e.Name))

		opt, err := portCreateOptsFromPortTask(t, a, e, changes)
		if err != nil {
			return fmt.Errorf("Error creating port cloud opts: %v", err)
		}

		v, err := t.Cloud.CreatePort(opt)
		if err != nil {
			return fmt.Errorf("Error creating port: %v", err)
		}

		if e.Tag != nil {
			err = t.Cloud.AppendTag(openstack.ResourceTypePort, v.ID, fi.StringValue(e.Tag))
			if err != nil {
				return fmt.Errorf("Error appending tag to port: %v", err)
			}
		}
		e.ID = fi.String(v.ID)
		klog.V(2).Infof("Creating a new Openstack port, id=%s", v.ID)
		return nil
	} else if changes != nil && changes.Tag != nil {
		err := t.Cloud.AppendTag(openstack.ResourceTypePort, fi.StringValue(a.ID), fi.StringValue(changes.Tag))
		if err != nil {
			return fmt.Errorf("Error appending tag to port: %v", err)
		}
	}
	e.ID = a.ID
	klog.V(2).Infof("Using an existing Openstack port, id=%s", fi.StringValue(e.ID))
	return nil
}

func portCreateOptsFromPortTask(t *openstack.OpenstackAPITarget, a, e, changes *Port) (ports.CreateOptsBuilder, error) {
	sgs := make([]string, len(e.SecurityGroups)+len(e.AdditionalSecurityGroups))
	for i, sg := range e.SecurityGroups {
		sgs[i] = fi.StringValue(sg.ID)
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
			SubnetID: fi.StringValue(subn.ID),
		}
	}

	return ports.CreateOpts{
		Name:           fi.StringValue(e.Name),
		NetworkID:      fi.StringValue(e.Network.ID),
		SecurityGroups: &sgs,
		FixedIPs:       fixedIPs,
	}, nil
}
