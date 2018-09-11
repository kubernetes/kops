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
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/schedulerhints"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

//go:generate fitask -type=Instance
type Instance struct {
	ID     *string
	Name   *string
	Port   *Port
	Region *string
	Flavor *string
	Image  *string
	SSHKey *string
	Tags   []string
	Count  int
	Role   *string

	Lifecycle *fi.Lifecycle
}

var _ fi.CompareWithID = &Instance{}

func (e *Instance) CompareWithID() *string {
	return e.ID
}

func (e *Instance) Find(c *fi.Context) (*Instance, error) {
	if e == nil || e.ID == nil {
		return nil, nil
	}
	id := *(e.ID)
	v, err := servers.Get(c.Cloud.(openstack.OpenstackCloud).ComputeClient(), id).Extract()
	if err != nil {
		return nil, fmt.Errorf("error finding server with id %s: %v", id, err)
	}

	a := new(Instance)
	a.ID = fi.String(v.ID)
	a.Name = fi.String(v.Name)
	a.SSHKey = fi.String(v.KeyName)
	a.Lifecycle = e.Lifecycle

	return a, nil
}

func (e *Instance) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *Instance) CheckChanges(a, e, changes *Instance) error {
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

func (_ *Instance) ShouldCreate(a, e, changes *Instance) (bool, error) {
	return a == nil, nil
}

func (_ *Instance) RenderOpenstack(t *openstack.OpenstackAPITarget, a, e, changes *Instance) error {
	if a == nil {
		glog.V(2).Infof("Creating Instance with name: %q", fi.StringValue(e.Name))

		opt := servers.CreateOpts{
			Name:       fi.StringValue(e.Name),
			ImageName:  fi.StringValue(e.Image),
			FlavorName: fi.StringValue(e.Flavor),
			Networks: []servers.Network{
				{
					Port: fi.StringValue(e.Port.ID),
				},
			},
		}
		keyext := keypairs.CreateOptsExt{
			CreateOptsBuilder: opt,
			KeyName:           fi.StringValue(e.SSHKey),
		}
		sgext := schedulerhints.CreateOptsExt{
			CreateOptsBuilder: keyext,
			SchedulerHints: &schedulerhints.SchedulerHints{
				Group: fi.StringValue(e.Role),
			},
		}
		v, err := t.Cloud.CreateInstance(sgext)
		if err != nil {
			return fmt.Errorf("Error creating instance: %v", err)
		}

		e.ID = fi.String(v.ID)
		glog.V(2).Infof("Creating a new Openstack instance, id=%s", v.ID)
		return nil
	}

	glog.V(2).Infof("Openstack task Instance::RenderOpenstack did nothing")
	return nil
}
