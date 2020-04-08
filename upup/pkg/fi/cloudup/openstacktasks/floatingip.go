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
	"time"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/floatingips"
	l3floatingip "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/kops/util/pkg/vfs"
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

var readBackoff = wait.Backoff{
	Duration: time.Second,
	Factor:   1.5,
	Jitter:   0.1,
	Steps:    10,
}

// this function tries to find l3 floating, and retries x times to find that. In some cases the floatingip is not in place in first request
func findL3Floating(cloud openstack.OpenstackCloud, opts l3floatingip.ListOpts) ([]l3floatingip.FloatingIP, error) {
	var result []l3floatingip.FloatingIP
	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		fips, err := cloud.ListL3FloatingIPs(opts)
		if err != nil {
			return false, fmt.Errorf("Failed to list L3 floating ip: %v", err)
		}
		if len(fips) == 0 {
			return false, nil
		}
		result = fips
		return true, nil
	})
	if !done {
		if err == nil {
			err = wait.ErrWaitTimeout
		}
		return result, err
	}
	return result, nil
}

func (e *FloatingIP) findServerFloatingIP(context *fi.Context, cloud openstack.OpenstackCloud) (*string, error) {
	ips, err := cloud.ListServerIPs(fi.StringValue(e.Server.ID), openstack.FloatingType)
	if err != nil {
		return nil, err
	}
	// assumes that we do have only one interface and 0-1 floatingip in server, the setup that kops does
	if len(ips) > 0 {
		return ips[0], nil
	}
	return nil, nil
}

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
	// try to find ip address using LB port
	if e.ID == nil && e.LB != nil && e.LB.PortID != nil {
		fips, err := findL3Floating(cloud, l3floatingip.ListOpts{
			PortID: fi.StringValue(e.LB.PortID),
		})
		if err != nil {
			return nil, err
		}
		if len(fips) == 1 && fips[0].PortID == fi.StringValue(e.LB.PortID) {
			return &fips[0].FloatingIP, nil
		}
		return nil, fmt.Errorf("Could not find port floatingips port=%s", fi.StringValue(e.LB.PortID))
	} else if e.ID == nil && e.Server != nil {
		floating, err := e.findServerFloatingIP(context, cloud)
		if err != nil {
			return nil, fmt.Errorf("Could not find server (\"%s\") floating ip: %v", fi.StringValue(e.Server.ID), err)
		}
		return floating, nil
	}

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
		// We can't create a floating IP until the router with access to the external network
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
		for i, fip := range fips {
			if fip.InstanceID == fi.StringValue(e.Server.ID) {
				if actual != nil {
					return nil, fmt.Errorf("Multiple floating ip's associated to server: %s", fi.StringValue(e.Server.ID))
				}
				actual = &FloatingIP{
					Name:      e.Name,
					ID:        fi.String(fips[i].ID),
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

			opts := l3floatingip.CreateOpts{
				FloatingNetworkID: external.ID,
				PortID:            fi.StringValue(e.LB.PortID),
			}
			lbSubnet, err := cloud.GetLBFloatingSubnet()
			if err != nil {
				return fmt.Errorf("Failed to find floatingip subnet: %v", err)
			}
			if lbSubnet != nil {
				opts.SubnetID = lbSubnet.ID
			}
			fip, err := cloud.CreateL3FloatingIP(opts)
			if err != nil {
				return fmt.Errorf("Failed to create floating IP: %v", err)
			}

			e.ID = fi.String(fip.ID)

		} else if e.Server != nil {

			if err := e.Server.WaitForStatusActive(t); err != nil {
				return fmt.Errorf("Failed to associate floating IP to instance %s", *e.Name)
			}

			// recheck is there floatingip already in port
			// this can happen for instance when recreating bastion host
			fips, err := cloud.ListFloatingIPs()
			if err != nil {
				return fmt.Errorf("Failed to list floating ip's: %v", err)
			}
			for _, fip := range fips {
				if fip.InstanceID == fi.StringValue(e.Server.ID) {
					e.ID = fi.String(fip.ID)
					klog.V(2).Infof("Openstack::RenderOpenstack floatingip found after server is active")
					return nil
				}
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

	klog.V(2).Infof("Openstack task Instance::RenderOpenstack did nothing")
	return nil
}
