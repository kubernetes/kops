/*
Copyright 2026 The Kubernetes Authors.

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

package linodetasks

import (
	"context"
	"fmt"
	"strconv"

	"github.com/linode/linodego/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

// +kops:fitask
type Subnet struct {
	Name      *string
	ID        *int
	Lifecycle fi.Lifecycle

	IPv4 *string
	VPC  *VPC
}

var _ fi.CloudupTask = &Subnet{}
var _ fi.CompareWithID = &Subnet{}

func (v *Subnet) CompareWithID() *string {
	if v.ID == nil {
		return nil
	}
	id := strconv.Itoa(fi.ValueOf(v.ID))
	return new(id)
}

func (v *Subnet) Find(c *fi.CloudupContext) (*Subnet, error) {
	cloud := c.T.Cloud.(linode.LinodeCloud)

	if v.VPC == nil {
		return nil, fmt.Errorf("Subnet.VPC is required")
	}
	if v.VPC.ID == nil {
		if v.VPC.Name != nil {
			if _, ok := c.Target.(*fi.CloudupDryRunTarget); ok {
				return nil, nil
			}
			return nil, fi.NewTryAgainLaterError("waiting for VPC ID")
		}
		return nil, fmt.Errorf("Subnet.VPC.ID is required")
	}

	subnets, err := cloud.Client().ListVPCSubnets(c.Context(), fi.ValueOf(v.VPC.ID), nil)
	if err != nil {
		return nil, fmt.Errorf("error listing Linode (Akamai) VPC Subnets: %w", err)
	}

	var foundByName *linodego.VPCSubnet
	var foundByIPv4 *linodego.VPCSubnet
	name := fi.ValueOf(v.Name)
	ipv4 := fi.ValueOf(v.IPv4)
	for i := range subnets {
		candidate := &subnets[i]
		if candidate.Label == name {
			if foundByName != nil {
				return nil, fmt.Errorf("found multiple Linode (Akamai) VPC Subnets named %q", name)
			}
			foundByName = candidate
		}
		if ipv4 != "" && candidate.IPv4 == ipv4 {
			if foundByIPv4 == nil {
				foundByIPv4 = candidate
			}
		}
	}

	found := foundByName
	if found == nil {
		found = foundByIPv4
		if found == nil {
			return nil, nil
		}
	}

	actual := &Subnet{
		Name:      new(found.Label),
		ID:        new(found.ID),
		Lifecycle: v.Lifecycle,
		IPv4:      new(found.IPv4),
		VPC:       v.VPC,
	}
	v.ID = actual.ID

	return actual, nil
}

func (v *Subnet) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(v, c)
}

func (_ *Subnet) CheckChanges(actual, expected, changes *Subnet) error {
	if actual != nil {
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.IPv4 != nil {
			return fi.CannotChangeField("IPv4")
		}
		if changes.VPC != nil {
			return fi.CannotChangeField("VPC")
		}
	} else {
		if expected.Name == nil {
			return fi.RequiredField("Name")
		}
		if expected.IPv4 == nil {
			return fi.RequiredField("IPv4")
		}
		if expected.VPC == nil {
			return fi.RequiredField("VPC")
		}
	}

	return nil
}

func (_ *Subnet) RenderLinode(t *linode.APITarget, actual, expected, changes *Subnet) error {
	if actual == nil {
		subnet, err := t.Cloud.Client().CreateVPCSubnet(context.Background(), linodego.VPCSubnetCreateOptions{
			Label: fi.ValueOf(expected.Name),
			IPv4:  fi.ValueOf(expected.IPv4),
		},
			fi.ValueOf(expected.VPC.ID),
		)
		if err != nil {
			return fmt.Errorf("error creating Linode (Akamai) Subnet %q: %w", fi.ValueOf(expected.Name), err)
		}
		expected.ID = new(subnet.ID)
		return nil
	}

	if changes == nil || (changes.Name == nil && changes.IPv4 == nil) {
		expected.ID = actual.ID
		return nil
	}

	subnet, err := t.Cloud.Client().UpdateVPCSubnet(context.Background(), fi.ValueOf(actual.VPC.ID), fi.ValueOf(actual.ID), linodego.VPCSubnetUpdateOptions{
		Label: fi.ValueOf(expected.Name),
	})
	if err != nil {
		return fmt.Errorf("error updating Linode (Akamai) Subnet %q: %w", fi.ValueOf(expected.Name), err)
	}
	expected.ID = new(subnet.ID)

	return nil
}
