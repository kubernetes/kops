/*
Copyright 2018 The Kubernetes Authors.

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

	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

//go:generate fitask -type=Port
type Port struct {
	ID             *string
	Name           *string
	Network        *Network
	SecurityGroups []*SecurityGroup
	Lifecycle      *fi.Lifecycle
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
	sgs := make([]*SecurityGroup, len(port.SecurityGroups))
	for i, sgid := range port.SecurityGroups {
		sgs[i] = &SecurityGroup{
			ID:        fi.String(sgid),
			Lifecycle: lifecycle,
		}
	}

	actual := &Port{
		ID:             fi.String(port.ID),
		Name:           fi.String(port.Name),
		Network:        &Network{ID: fi.String(port.NetworkID)},
		SecurityGroups: sgs,
		Lifecycle:      lifecycle,
	}
	if find != nil {
		find.ID = actual.ID
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
	if rs == nil {
		return nil, nil
	} else if len(rs) != 1 {
		return nil, fmt.Errorf("found multiple ports with name: %s", fi.StringValue(s.Name))
	}

	return NewPortTaskFromCloud(cloud, s.Lifecycle, &rs[0], s)
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
		if e.Network != nil {
			return fi.CannotChangeField("Network")
		}
	}
	return nil
}

func (_ *Port) RenderOpenstack(t *openstack.OpenstackAPITarget, a, e, changes *Port) error {
	if a == nil {
		klog.V(2).Infof("Creating Port with name: %q", fi.StringValue(e.Name))

		sgs := make([]string, len(e.SecurityGroups))
		for i, sg := range e.SecurityGroups {
			sgs[i] = fi.StringValue(sg.ID)
		}

		opt := ports.CreateOpts{
			Name:           fi.StringValue(e.Name),
			NetworkID:      fi.StringValue(e.Network.ID),
			SecurityGroups: &sgs,
		}

		v, err := t.Cloud.CreatePort(opt)
		if err != nil {
			return fmt.Errorf("Error creating port: %v", err)
		}

		e.ID = fi.String(v.ID)
		klog.V(2).Infof("Creating a new Openstack port, id=%s", v.ID)
		return nil
	}
	e.ID = a.ID
	klog.V(2).Infof("Using an existing Openstack port, id=%s", fi.StringValue(e.ID))
	return nil
}
