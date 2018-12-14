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

	"github.com/golang/glog"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/floatingips"
	l3floatingip "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

//go:generate fitask -type=FloatingIP
type FloatingIP struct {
	Name      *string
	ID        *string
	Server    *Instance
	LB        *LB
	Lifecycle *fi.Lifecycle
}

var _ fi.HasAddress = &FloatingIP{}

func (e *FloatingIP) FindIPAddress(context *fi.Context) (*string, error) {
	if e.ID == nil {
		if e.Server != nil && e.Server.ID == nil {
			return nil, nil
		}
		if e.LB != nil && e.LB.ID == nil {
			return nil, nil
		}
	}

	cloud := context.Cloud.(openstack.OpenstackCloud)

	fip, err := cloud.GetFloatingIP(fi.StringValue(e.ID))
	if err != nil {
		return nil, err
	}
	return &fip.IP, nil
}

// GetDependencies returns the dependencies of the Instance task
func (e *FloatingIP) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task
	for _, task := range tasks {
		if _, ok := task.(*Instance); ok {
			deps = append(deps, task)
		}
		if _, ok := task.(*LB); ok {
			deps = append(deps, task)
		}
		// We cant create a floating IP until the router with access to the external network
		//  Has created an interface to our subnet
		if _, ok := task.(*RouterInterface); ok {
			deps = append(deps, task)
		}
	}
	return deps
}

var _ fi.CompareWithID = &FloatingIP{}

func (e *FloatingIP) CompareWithID() *string {
	return e.ID
}

func (e *FloatingIP) Find(c *fi.Context) (*FloatingIP, error) {
	if e == nil {
		return nil, nil
	}
	cloud := c.Cloud.(openstack.OpenstackCloud)
	if e.LB != nil && e.LB.PortID != nil {
		// Layer 3
		fips, err := cloud.ListL3FloatingIPs(l3floatingip.ListOpts{
			PortID: fi.StringValue(e.LB.PortID),
		})
		if err != nil {
			return nil, fmt.Errorf("Failed to list layer 3 floating ip's for port ID %s: %v", fi.StringValue(e.LB.PortID), err)
		}
		if len(fips) == 0 {
			return nil, nil
		}
		if len(fips) > 1 {
			return nil, fmt.Errorf("Multiple floating ip's associated to port: %s", fi.StringValue(e.LB.PortID))
		}
		actual := &FloatingIP{
			Name:      e.Name,
			ID:        fi.String(fips[0].ID),
			LB:        e.LB,
			Lifecycle: e.Lifecycle,
		}
		e.ID = actual.ID
		return actual, nil
	} else if e.Server != nil {
		fips, err := cloud.ListFloatingIPs()
		if err != nil {
			return nil, fmt.Errorf("Failed to list floating ip's: %v", err)
		}
		if len(fips) == 0 {
			return nil, nil
		}
		var actual *FloatingIP
		for _, fip := range fips {
			if fip.InstanceID == fi.StringValue(e.Server.ID) {
				if actual != nil {
					return nil, fmt.Errorf("Multiple floating ip's associated to server: %s", fi.StringValue(e.Server.ID))
				}
				actual = &FloatingIP{
					Name:      e.Name,
					ID:        fi.String(fips[0].ID),
					Server:    e.Server,
					Lifecycle: e.Lifecycle,
				}
				break
			}
		}
		if actual != nil {
			e.ID = actual.ID
		}
		return actual, nil
	}
	return nil, nil
}

func (e *FloatingIP) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *FloatingIP) CheckChanges(a, e, changes *FloatingIP) error {
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

func (_ *FloatingIP) ShouldCreate(a, e, changes *FloatingIP) (bool, error) {
	return a == nil, nil
}

func (f *FloatingIP) RenderOpenstack(t *openstack.OpenstackAPITarget, a, e, changes *FloatingIP) error {

	if a == nil {
		cloud := t.Cloud.(openstack.OpenstackCloud)
		external, err := cloud.GetExternalNetwork()
		if err != nil {
			return fmt.Errorf("Failed to find external network: %v", err)
		}

		if e.LB != nil {
			//Layer 3
			fip, err := cloud.CreateL3FloatingIP(l3floatingip.CreateOpts{
				FloatingNetworkID: external.ID,
				PortID:            fi.StringValue(e.LB.PortID),
			})
			if err != nil {
				return fmt.Errorf("Failed to create floating IP: %v", err)
			}

			e.ID = fi.String(fip.ID)

		} else if e.Server != nil {

			if err := e.Server.WaitForStatusActive(t); err != nil {
				return fmt.Errorf("Failed to associate floating IP to instance %s", *e.Name)
			}
			fip, err := cloud.CreateFloatingIP(floatingips.CreateOpts{
				Pool: external.Name,
			})
			if err != nil {
				return fmt.Errorf("Failed to create floating IP: %v", err)
			}
			err = cloud.AssociateFloatingIPToInstance(fi.StringValue(e.Server.ID), floatingips.AssociateOpts{
				FloatingIP: fip.IP,
			})
			if err != nil {
				return fmt.Errorf("Failed to associated floating IP to instance %s: %v", *e.Name, err)
			}

			e.ID = fi.String(fip.ID)

		} else {
			return fmt.Errorf("Must specify either Port or Server!")
		}
		return nil
	}

	glog.V(2).Infof("Openstack task Instance::RenderOpenstack did nothing")
	return nil
}
