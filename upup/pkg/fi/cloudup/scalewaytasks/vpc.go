/*
Copyright 2024 The Kubernetes Authors.

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

package scalewaytasks

import (
	"fmt"
	"strings"

	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type VPC struct {
	ID     *string
	Name   *string
	Region *string
	Tags   []string

	Lifecycle fi.Lifecycle
}

var _ fi.CloudupTask = &VPC{}
var _ fi.CompareWithID = &VPC{}

func (v *VPC) CompareWithID() *string {
	return v.ID
}

func (v *VPC) Find(context *fi.CloudupContext) (*VPC, error) {
	cloud := context.T.Cloud.(scaleway.ScwCloud)
	vpcs, err := cloud.VPCService().ListVPCs(&vpc.ListVPCsRequest{
		Region: scw.Region(cloud.Region()),
		Name:   v.Name,
		Tags:   []string{fmt.Sprintf("%s=%s", scaleway.TagClusterName, scaleway.ClusterNameFromTags(v.Tags))},
	}, scw.WithContext(context.Context()), scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("listing VPCs: %w", err)
	}

	if vpcs.TotalCount == 0 {
		return nil, nil
	}
	if vpcs.TotalCount > 1 {
		return nil, fmt.Errorf("expected exactly 1 VPC, got %d", vpcs.TotalCount)
	}
	vpcFound := vpcs.Vpcs[0]

	return &VPC{
		ID:        fi.PtrTo(vpcFound.ID),
		Name:      fi.PtrTo(vpcFound.Name),
		Region:    fi.PtrTo(vpcFound.Region.String()),
		Tags:      vpcFound.Tags,
		Lifecycle: v.Lifecycle,
	}, nil
}

func (v *VPC) Run(context *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(v, context)
}

func (_ *VPC) CheckChanges(actual, expected, changes *VPC) error {
	if actual != nil {
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
		if expected.Name == nil {
			return fi.RequiredField("Name")
		}
		if expected.Region == nil {
			return fi.RequiredField("Region")
		}
	}
	return nil
}

func (_ *VPC) RenderScw(t *scaleway.ScwAPITarget, actual, expected, _ *VPC) error {
	if actual != nil {
		//TODO(Mia-Cross): update tags
		return nil
	}

	cloud := t.Cloud.(scaleway.ScwCloud)
	region := scw.Region(fi.ValueOf(expected.Region))

	vpcCreated, err := cloud.VPCService().CreateVPC(&vpc.CreateVPCRequest{
		Region: region,
		Name:   fi.ValueOf(expected.Name),
		Tags:   expected.Tags,
	})
	if err != nil {
		return fmt.Errorf("creating VPC: %w", err)
	}

	expected.ID = &vpcCreated.ID

	return nil
}

type terraformVPC struct {
	Name *string  `cty:"name"`
	Tags []string `cty:"tags"`
}

func (_ *VPC) RenderTerraform(t *terraform.TerraformTarget, actual, expected, changes *VPC) error {
	tfName := strings.ReplaceAll(fi.ValueOf(expected.Name), ".", "-")
	tfVPC := terraformVPC{
		Name: expected.Name,
		Tags: expected.Tags,
	}
	return t.RenderResource("scaleway_vpc", tfName, tfVPC)
}

func (v *VPC) TerraformLink() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("scaleway_vpc", fi.ValueOf(v.Name), "id")
}
