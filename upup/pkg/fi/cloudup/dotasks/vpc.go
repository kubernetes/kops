/*
Copyright 2021 The Kubernetes Authors.

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

package dotasks

import (
	"context"

	"github.com/digitalocean/godo"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/do"
)

// +kops:fitask
type VPC struct {
	Name      *string
	ID        *string
	Lifecycle fi.Lifecycle
	IPRange   *string
	Region    *string
}

var _ fi.CompareWithID = &VPC{}

func (v *VPC) CompareWithID() *string {
	return v.ID
}

func (v *VPC) Find(c *fi.Context) (*VPC, error) {
	cloud := c.Cloud.(do.DOCloud)
	vpcService := cloud.VPCsService()

	opt := &godo.ListOptions{}
	vpcs, _, err := vpcService.List(context.TODO(), opt)
	if err != nil {
		return nil, err
	}

	for _, vpc := range vpcs {
		if vpc.Name == fi.StringValue(v.Name) {
			return &VPC{
				Name:      fi.String(vpc.Name),
				ID:        fi.String(vpc.ID),
				Lifecycle: v.Lifecycle,
				IPRange:   fi.String(vpc.IPRange),
				Region:    fi.String(vpc.RegionSlug),
			}, nil
		}
	}

	// VPC = nil if not found
	return nil, nil
}

func (v *VPC) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(v, c)
}

func (_ *VPC) CheckChanges(a, e, changes *VPC) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Region != nil {
			return fi.CannotChangeField("Region")
		}
	} else {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		if e.Region == nil {
			return fi.RequiredField("Region")
		}
	}
	return nil
}

func (_ *VPC) RenderDO(t *do.DOAPITarget, a, e, changes *VPC) error {
	if a != nil {
		return nil
	}

	vpcService := t.Cloud.VPCsService()
	_, _, err := vpcService.Create(context.TODO(), &godo.VPCCreateRequest{
		Name:       fi.StringValue(e.Name),
		RegionSlug: fi.StringValue(e.Region),
		IPRange:    fi.StringValue(e.IPRange),
	})

	return err
}
