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

	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

//go:generate fitask -type=Network
type Network struct {
	ID        *string
	Name      *string
	Lifecycle *fi.Lifecycle
	Tag       *string
}

var _ fi.CompareWithID = &Network{}

func (n *Network) CompareWithID() *string {
	return n.ID
}

func NewNetworkTaskFromCloud(cloud openstack.OpenstackCloud, lifecycle *fi.Lifecycle, network *networks.Network, networkName *string) (*Network, error) {
	tag := ""
	if networkName != nil && fi.ArrayContains(network.Tags, fi.StringValue(networkName)) {
		tag = fi.StringValue(networkName)
	}

	task := &Network{
		ID:        fi.String(network.ID),
		Name:      fi.String(network.Name),
		Lifecycle: lifecycle,
		Tag:       fi.String(tag),
	}
	return task, nil
}

func (n *Network) Find(context *fi.Context) (*Network, error) {
	if n.Name == nil && n.ID == nil {
		return nil, nil
	}

	cloud := context.Cloud.(openstack.OpenstackCloud)
	opt := networks.ListOpts{
		ID:   fi.StringValue(n.ID),
		Name: fi.StringValue(n.Name),
	}
	ns, err := cloud.ListNetworks(opt)
	if err != nil {
		return nil, err
	}
	if ns == nil {
		return nil, nil
	} else if len(ns) != 1 {
		return nil, fmt.Errorf("found multiple networks with name: %s", fi.StringValue(n.Name))
	}
	v := ns[0]
	actual, err := NewNetworkTaskFromCloud(cloud, n.Lifecycle, &v, n.Tag)
	if err != nil {
		return nil, fmt.Errorf("Failed to create new Network object: %v", err)
	}
	n.ID = actual.ID
	return actual, nil
}

func (c *Network) Run(context *fi.Context) error {
	return fi.DefaultDeltaRunMethod(c, context)
}

func (_ *Network) CheckChanges(a, e, changes *Network) error {
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

func (_ *Network) ShouldCreate(a, e, changes *Network) (bool, error) {
	if a == nil || changes.Tag != nil {
		return true, nil
	}
	return false, nil
}

func (_ *Network) RenderOpenstack(t *openstack.OpenstackAPITarget, a, e, changes *Network) error {
	if a == nil {
		klog.V(2).Infof("Creating Network with name:%q", fi.StringValue(e.Name))

		opt := networks.CreateOpts{
			Name:         fi.StringValue(e.Name),
			AdminStateUp: fi.Bool(true),
		}

		v, err := t.Cloud.CreateNetwork(opt)
		if err != nil {
			return fmt.Errorf("Error creating network: %v", err)
		}

		err = t.Cloud.AppendTag(openstack.ResourceTypeNetwork, v.ID, fi.StringValue(e.Tag))
		if err != nil {
			return fmt.Errorf("Error appending tag to network: %v", err)
		}

		e.ID = fi.String(v.ID)
		klog.V(2).Infof("Creating a new Openstack network, id=%s", v.ID)
		return nil
	} else {
		err := t.Cloud.AppendTag(openstack.ResourceTypeNetwork, fi.StringValue(a.ID), fi.StringValue(changes.Tag))
		if err != nil {
			return fmt.Errorf("Error appending tag to network: %v", err)
		}
	}
	e.ID = a.ID
	klog.V(2).Infof("Using an existing Openstack network, id=%s", fi.StringValue(e.ID))
	return nil
}
