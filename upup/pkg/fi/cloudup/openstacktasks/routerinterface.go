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

	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/routers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

// +kops:fitask
type RouterInterface struct {
	ID        *string
	Name      *string
	Router    *Router
	Subnet    *Subnet
	Lifecycle fi.Lifecycle
}

// GetDependencies returns the dependencies of the RouterInterface task
func (e *RouterInterface) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	var deps []fi.CloudupTask
	for _, task := range tasks {
		if _, ok := task.(*Router); ok {
			deps = append(deps, task)
		}
		if _, ok := task.(*Subnet); ok {
			deps = append(deps, task)
		}
	}
	return deps
}

var _ fi.CompareWithID = &RouterInterface{}

func (i *RouterInterface) CompareWithID() *string {
	return i.ID
}

func (i *RouterInterface) Find(context *fi.CloudupContext) (*RouterInterface, error) {
	cloud := context.T.Cloud.(openstack.OpenstackCloud)
	opt := ports.ListOpts{
		NetworkID: fi.ValueOf(i.Subnet.Network.ID),
		DeviceID:  fi.ValueOf(i.Router.ID),
		ID:        fi.ValueOf(i.ID),
	}
	ps, err := cloud.ListPorts(opt)
	if err != nil {
		return nil, err
	}
	if ps == nil {
		return nil, nil
	}

	var actual *RouterInterface

	subnetID := fi.ValueOf(i.Subnet.ID)
	for _, p := range ps {
		for _, ip := range p.FixedIPs {
			if ip.SubnetID == subnetID {
				if actual != nil {
					return nil, fmt.Errorf("found multiple interfaces which subnet:%s attach to", subnetID)
				}
				actual = &RouterInterface{
					ID:        fi.PtrTo(p.ID),
					Name:      i.Name,
					Router:    i.Router,
					Subnet:    i.Subnet,
					Lifecycle: i.Lifecycle,
				}
				i.ID = actual.ID
			}
		}
	}
	return actual, nil
}

func (i *RouterInterface) Run(context *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(i, context)
}

func (*RouterInterface) CheckChanges(a, e, changes *RouterInterface) error {
	if a == nil {
		if e.Router == nil {
			return fi.RequiredField("Router")
		}
		if e.Subnet == nil {
			return fi.RequiredField("Subnet")
		}
	} else {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.Router != nil {
			return fi.CannotChangeField("Router")
		}
		if changes.Subnet != nil {
			return fi.CannotChangeField("Subnet")
		}
	}
	return nil
}

func (_ *RouterInterface) RenderOpenstack(t *openstack.OpenstackAPITarget, a, e, changes *RouterInterface) error {
	if a == nil {
		routerID := fi.ValueOf(e.Router.ID)
		subnetID := fi.ValueOf(e.Subnet.ID)
		klog.V(2).Infof("Creating RouterInterface for router:%s and subnet:%s", routerID, subnetID)

		opt := routers.AddInterfaceOpts{SubnetID: subnetID}
		v, err := t.Cloud.CreateRouterInterface(routerID, opt)
		if err != nil {
			return fmt.Errorf("Error creating router interface: %v", err)
		}

		e.ID = fi.PtrTo(v.PortID)
		klog.V(2).Infof("Creating a new Openstack router interface, id=%s", v.PortID)
		return nil
	}
	e.ID = a.ID
	klog.V(2).Infof("Using an existing Openstack router interface, id=%s", fi.ValueOf(e.ID))
	return nil
}
