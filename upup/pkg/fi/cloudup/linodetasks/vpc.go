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
type VPC struct {
	Name      *string
	ID        *int
	Lifecycle fi.Lifecycle

	Description *string
	Region      *string
}

var _ fi.CloudupTask = &VPC{}
var _ fi.CompareWithID = &VPC{}

func (v *VPC) CompareWithID() *string {
	if v.ID == nil {
		return nil
	}
	id := strconv.Itoa(fi.ValueOf(v.ID))
	return new(id)
}

func (v *VPC) Find(c *fi.CloudupContext) (*VPC, error) {
	cloud := c.T.Cloud.(linode.LinodeCloud)

	vpcs, err := cloud.Client().ListVPCs(c.Context(), nil)
	if err != nil {
		return nil, fmt.Errorf("error listing Linode (Akamai) VPCs: %w", err)
	}

	var found *linodego.VPC
	name := fi.ValueOf(v.Name)
	for i := range vpcs {
		candidate := &vpcs[i]
		if candidate.Label != name {
			continue
		}
		if v.Region != nil && candidate.Region != fi.ValueOf(v.Region) {
			continue
		}
		if found != nil {
			return nil, fmt.Errorf("found multiple Linode (Akamai) VPCs named %q", name)
		}
		found = candidate
	}

	if found == nil {
		return nil, nil
	}

	actual := &VPC{
		Name:        new(found.Label),
		ID:          new(found.ID),
		Lifecycle:   v.Lifecycle,
		Description: new(found.Description),
		Region:      new(found.Region),
	}
	v.ID = actual.ID

	return actual, nil
}

func (v *VPC) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(v, c)
}

func (_ *VPC) CheckChanges(actual, expected, changes *VPC) error {
	if actual != nil {
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Region != nil {
			return fi.CannotChangeField("Region")
		}
	} else {
		if expected.Name == nil {
			return fi.RequiredField("Name")
		}
		if expected.Region == nil {
			return fi.RequiredField("Region")
		}
	}

	return nil
}

func (_ *VPC) RenderLinode(t *linode.APITarget, actual, expected, changes *VPC) error {
	if actual == nil {
		vpc, err := t.Cloud.Client().CreateVPC(context.Background(), linodego.VPCCreateOptions{
			Label:       fi.ValueOf(expected.Name),
			Description: fi.ValueOf(expected.Description),
			Region:      fi.ValueOf(expected.Region),
		})
		if err != nil {
			return fmt.Errorf("error creating Linode (Akamai) VPC %q: %w", fi.ValueOf(expected.Name), err)
		}
		expected.ID = new(vpc.ID)
		return nil
	}

	if changes == nil || (changes.Name == nil && changes.Description == nil) {
		expected.ID = actual.ID
		return nil
	}

	vpc, err := t.Cloud.Client().UpdateVPC(context.Background(), fi.ValueOf(actual.ID), linodego.VPCUpdateOptions{
		Label:       fi.ValueOf(expected.Name),
		Description: fi.ValueOf(expected.Description),
	})
	if err != nil {
		return fmt.Errorf("error updating Linode (Akamai) VPC %q: %w", fi.ValueOf(expected.Name), err)
	}
	expected.ID = new(vpc.ID)

	return nil
}
