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
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

//go:generate fitask -type=Subnet
type Subnet struct {
	ID        *string
	Name      *string
	Network   *Network
	CIDR      *string
	Lifecycle *fi.Lifecycle
}

var _ fi.CompareWithID = &Subnet{}

func (s *Subnet) CompareWithID() *string {
	return s.ID
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
	v := rs[0]
	actual := &Subnet{
		ID:        fi.String(v.ID),
		Name:      fi.String(v.Name),
		Network:   &Network{ID: fi.String(v.NetworkID)},
		CIDR:      fi.String(v.CIDR),
		Lifecycle: s.Lifecycle,
	}
	return actual, nil
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
		if e.Network != nil {
			return fi.CannotChangeField("Network")
		}
		if e.CIDR != nil {
			return fi.CannotChangeField("CIDR")
		}
	}
	return nil
}

func (_ *Subnet) RenderOpenstack(t *openstack.OpenstackAPITarget, a, e, changes *Subnet) error {
	if a == nil {
		glog.V(2).Infof("Creating Subnet with name:%q", fi.StringValue(e.Name))

		opt := subnets.CreateOpts{
			Name:       fi.StringValue(e.Name),
			NetworkID:  fi.StringValue(e.Network.ID),
			IPVersion:  gophercloud.IPv4,
			CIDR:       fi.StringValue(e.CIDR),
			EnableDHCP: fi.Bool(true),
		}

		v, err := t.Cloud.CreateSubnet(opt)
		if err != nil {
			return fmt.Errorf("Error creating subnet: %v", err)
		}

		e.ID = fi.String(v.ID)
		glog.V(2).Infof("Creating a new Openstack subnet, id=%s", v.ID)
		return nil
	}
	e.ID = a.ID
	glog.V(2).Infof("Using an existing Openstack subnet, id=%s", fi.StringValue(e.ID))
	return nil
}
