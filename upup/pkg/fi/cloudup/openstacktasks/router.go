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
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

// +kops:fitask
type Router struct {
	ID                    *string
	Name                  *string
	Lifecycle             fi.Lifecycle
	AvailabilityZoneHints []*string
}

var _ fi.CompareWithID = &Router{}

func (n *Router) CompareWithID() *string {
	return n.ID
}

// NewRouterTaskFromCloud initializes and returns a new Router
func NewRouterTaskFromCloud(cloud openstack.OpenstackCloud, lifecycle fi.Lifecycle, router *routers.Router, find *Router) (*Router, error) {
	actual := &Router{
		ID:                    fi.PtrTo(router.ID),
		Name:                  fi.PtrTo(router.Name),
		Lifecycle:             lifecycle,
		AvailabilityZoneHints: fi.StringSlice(router.AvailabilityZoneHints),
	}
	if find != nil {
		find.ID = actual.ID
	}
	return actual, nil
}

func (n *Router) Find(context *fi.CloudupContext) (*Router, error) {
	cloud := context.T.Cloud.(openstack.OpenstackCloud)
	opt := routers.ListOpts{
		Name: fi.ValueOf(n.Name),
		ID:   fi.ValueOf(n.ID),
	}
	rs, err := cloud.ListRouters(opt)
	if err != nil {
		return nil, err
	}
	if rs == nil {
		return nil, nil
	} else if len(rs) != 1 {
		return nil, fmt.Errorf("found multiple routers with name: %s", fi.ValueOf(n.Name))
	}
	return NewRouterTaskFromCloud(cloud, n.Lifecycle, &rs[0], n)
}

func (c *Router) Run(context *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(c, context)
}

func (_ *Router) CheckChanges(a, e, changes *Router) error {
	if a == nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	} else {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.AvailabilityZoneHints != nil {
			return fi.CannotChangeField("AvailabilityZoneHints")
		}
	}
	return nil
}

func (_ *Router) RenderOpenstack(t *openstack.OpenstackAPITarget, a, e, changes *Router) error {
	if a == nil {
		klog.V(2).Infof("Creating Router with name:%q", fi.ValueOf(e.Name))

		opt := routers.CreateOpts{
			Name:                  fi.ValueOf(e.Name),
			AdminStateUp:          fi.PtrTo(true),
			AvailabilityZoneHints: fi.StringSliceValue(e.AvailabilityZoneHints),
		}
		floatingNet, err := t.Cloud.GetExternalNetwork()
		if err != nil {
			return fmt.Errorf("Error creating router.  Could not list external networks for gateway: %v", err)
		}

		opt.GatewayInfo = &routers.GatewayInfo{
			NetworkID: floatingNet.ID,
		}

		routerFloatingSubnet, err := t.Cloud.GetExternalSubnet()
		if err != nil {
			return fmt.Errorf("Failed to find floatingip subnet: %v", err)
		}
		if routerFloatingSubnet != nil {
			opt.GatewayInfo.ExternalFixedIPs = []routers.ExternalFixedIP{
				{
					SubnetID: routerFloatingSubnet.ID,
				},
			}
		}

		v, err := t.Cloud.CreateRouter(opt)
		if err != nil {
			return fmt.Errorf("Error creating router: %v", err)
		}
		e.ID = fi.PtrTo(v.ID)
		klog.V(2).Infof("Creating a new Openstack router, id=%s", v.ID)
		return nil
	}
	e.ID = a.ID
	klog.V(2).Infof("Using an existing Openstack router, id=%s", fi.ValueOf(e.ID))
	return nil
}
