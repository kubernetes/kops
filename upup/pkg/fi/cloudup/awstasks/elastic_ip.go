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

package awstasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=ElasticIP

// ElasticIP manages an AWS Address (ElasticIP)
type ElasticIP struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	ID       *string
	PublicIP *string

	// ElasticIPs don't support tags.  We instead find it via a related resource.

	// TagOnSubnet tags a subnet with the ElasticIP.  Deprecated: doesn't round-trip with terraform.
	TagOnSubnet *Subnet

	Tags map[string]string

	// AssociatedNatGatewayRouteTable follows the RouteTable -> NatGateway -> ElasticIP
	AssociatedNatGatewayRouteTable *RouteTable
}

var _ fi.CompareWithID = &ElasticIP{}

func (e *ElasticIP) CompareWithID() *string {
	return e.ID
}

var _ fi.HasAddress = &ElasticIP{}

func (e *ElasticIP) FindIPAddress(context *fi.Context) (*string, error) {
	actual, err := e.find(context.Cloud.(awsup.AWSCloud))
	if err != nil {
		return nil, fmt.Errorf("error querying for ElasticIP: %v", err)
	}
	if actual == nil {
		return nil, nil
	}
	return actual.PublicIP, nil
}

// Find returns the actual ElasticIP state, or nil if not found
func (e *ElasticIP) Find(context *fi.Context) (*ElasticIP, error) {
	return e.find(context.Cloud.(awsup.AWSCloud))
}

// find will attempt to look up the elastic IP from AWS
func (e *ElasticIP) find(cloud awsup.AWSCloud) (*ElasticIP, error) {
	publicIP := e.PublicIP
	allocationID := e.ID

	// Find via RouteTable -> NatGateway -> ElasticIP
	if allocationID == nil && publicIP == nil && e.AssociatedNatGatewayRouteTable != nil {
		ngw, err := findNatGatewayFromRouteTable(cloud, e.AssociatedNatGatewayRouteTable)
		if err != nil {
			return nil, fmt.Errorf("error finding AssociatedNatGatewayRouteTable: %v", err)
		}

		if ngw == nil {
			klog.V(2).Infof("AssociatedNatGatewayRouteTable not found")
		} else {
			if len(ngw.NatGatewayAddresses) == 0 {
				return nil, fmt.Errorf("NatGateway %q has no addresses", *ngw.NatGatewayId)
			}
			if len(ngw.NatGatewayAddresses) > 1 {
				return nil, fmt.Errorf("NatGateway %q has multiple addresses", *ngw.NatGatewayId)
			}
			allocationID = ngw.NatGatewayAddresses[0].AllocationId
			if allocationID == nil {
				return nil, fmt.Errorf("NatGateway %q has nil addresses", *ngw.NatGatewayId)
			} else {
				klog.V(2).Infof("Found ElasticIP AllocationID %q via NatGateway", *allocationID)
			}
		}
	}

	// Find via tag on subnet
	// TODO: Deprecated, because doesn't round-trip with terraform
	if allocationID == nil && publicIP == nil && e.TagOnSubnet != nil && e.TagOnSubnet.ID != nil {
		var filters []*ec2.Filter
		filters = append(filters, awsup.NewEC2Filter("key", "AssociatedElasticIp"))
		filters = append(filters, awsup.NewEC2Filter("resource-id", *e.TagOnSubnet.ID))

		request := &ec2.DescribeTagsInput{
			Filters: filters,
		}

		response, err := cloud.EC2().DescribeTags(request)
		if err != nil {
			return nil, fmt.Errorf("error listing tags: %v", err)
		}

		if response == nil || len(response.Tags) == 0 {
			return nil, nil
		}

		if len(response.Tags) != 1 {
			return nil, fmt.Errorf("found multiple tags for: %v", e)
		}
		t := response.Tags[0]
		publicIP = t.Value
		klog.V(2).Infof("Found public IP via tag: %v", *publicIP)
	}

	if publicIP != nil || allocationID != nil {
		request := &ec2.DescribeAddressesInput{}
		if allocationID != nil {
			request.AllocationIds = []*string{allocationID}
		} else if publicIP != nil {
			request.Filters = []*ec2.Filter{awsup.NewEC2Filter("public-ip", *publicIP)}
		}

		response, err := cloud.EC2().DescribeAddresses(request)
		if err != nil {
			return nil, fmt.Errorf("error listing ElasticIPs: %v", err)
		}

		if response == nil || len(response.Addresses) == 0 {
			return nil, nil
		}

		if len(response.Addresses) != 1 {
			return nil, fmt.Errorf("found multiple ElasticIPs for: %v", e)
		}
		a := response.Addresses[0]
		actual := &ElasticIP{
			ID:       a.AllocationId,
			PublicIP: a.PublicIp,
		}
		actual.TagOnSubnet = e.TagOnSubnet
		actual.AssociatedNatGatewayRouteTable = e.AssociatedNatGatewayRouteTable

		{
			tags, err := cloud.EC2().DescribeTags(&ec2.DescribeTagsInput{
				Filters: []*ec2.Filter{
					{
						Name:   aws.String("resource-id"),
						Values: aws.StringSlice([]string{*a.AllocationId}),
					},
				},
			})
			if err != nil {
				return nil, fmt.Errorf("error querying tags for ElasticIP: %v", err)
			}
			var ec2Tags []*ec2.Tag
			for _, t := range tags.Tags {
				ec2Tags = append(ec2Tags, &ec2.Tag{
					Key:   t.Key,
					Value: t.Value,
				})
			}
			actual.Tags = intersectTags(ec2Tags, e.Tags)
		}

		// ElasticIP don't have a Name (no tags), so we set the name to avoid spurious changes
		actual.Name = e.Name

		e.ID = actual.ID

		// Avoid spurious changes
		actual.Lifecycle = e.Lifecycle

		return actual, nil
	}
	return nil, nil
}

// Run is called to execute this task.
// This is the main entry point of the task, and will actually
// connect our internal resource representation to an actual
// resource in AWS
func (e *ElasticIP) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

// CheckChanges validates the resource. EIPs are simple, so virtually no
// validation
func (_ *ElasticIP) CheckChanges(a, e, changes *ElasticIP) error {
	// This is a new EIP
	if a == nil {
		// No logic for EIPs - they are just created
		return nil
	}

	// This is an existing EIP
	// We should never be changing this
	if a != nil {
		if changes.PublicIP != nil {
			return fi.CannotChangeField("PublicIP")
		}
		if changes.TagOnSubnet != nil {
			return fi.CannotChangeField("TagOnSubnet")
		}
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
	}
	return nil
}

// RenderAWS is where we actually apply changes to AWS
func (_ *ElasticIP) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *ElasticIP) error {
	var publicIp *string
	var eipId *string

	// If this is a new ElasticIP
	if a == nil {
		klog.V(2).Infof("Creating ElasticIP for VPC")

		request := &ec2.AllocateAddressInput{}
		request.Domain = aws.String(ec2.DomainTypeVpc)

		response, err := t.Cloud.EC2().AllocateAddress(request)
		if err != nil {
			return fmt.Errorf("error creating ElasticIP: %v", err)
		}

		e.ID = response.AllocationId
		e.PublicIP = response.PublicIp
		publicIp = e.PublicIP
		eipId = response.AllocationId
	} else {
		publicIp = a.PublicIP
		eipId = a.ID
	}

	if err := t.AddAWSTags(*e.ID, e.Tags); err != nil {
		return err
	}

	// Tag the associated subnet
	if e.TagOnSubnet != nil {
		if e.TagOnSubnet.ID == nil {
			return fmt.Errorf("Subnet ID not set")
		}
		tags := make(map[string]string)
		tags["AssociatedElasticIp"] = *publicIp
		tags["AssociatedElasticIpAllocationId"] = *eipId // Leaving this in for reference, even though we don't use it
		err := t.AddAWSTags(*e.TagOnSubnet.ID, tags)
		if err != nil {
			return fmt.Errorf("Unable to tag subnet %v", err)
		}
	} else {
		// TODO: Figure out what we can do.  We're sort of stuck between wanting to have one code-path with
		// terraform, and having a bigger "window of loss" here before we create the NATGateway
		klog.V(2).Infof("ElasticIP %q not tagged on subnet; risk of leaking", fi.StringValue(publicIp))
	}

	return nil
}

type terraformElasticIP struct {
	VPC  *bool             `json:"vpc"`
	Tags map[string]string `json:"tags,omitempty"`
}

func (_ *ElasticIP) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *ElasticIP) error {
	tf := &terraformElasticIP{
		VPC:  aws.Bool(true),
		Tags: e.Tags,
	}

	return t.RenderResource("aws_eip", *e.Name, tf)
}

func (e *ElasticIP) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("aws_eip", *e.Name, "id")
}

type cloudformationElasticIP struct {
	Domain *string             `json:"Domain"`
	Tags   []cloudformationTag `json:"Tags,omitempty"`
}

func (_ *ElasticIP) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *ElasticIP) error {
	tf := &cloudformationElasticIP{
		Domain: aws.String("vpc"),
		Tags:   buildCloudformationTags(e.Tags),
	}

	return t.RenderResource("AWS::EC2::EIP", *e.Name, tf)
}

// Removed because you normally want CloudformationAllocationID
//func (e *ElasticIP) CloudformationLink() *cloudformation.Literal {
//	return cloudformation.Ref("AWS::EC2::EIP", *e.Name)
//}

func (e *ElasticIP) CloudformationAllocationID() *cloudformation.Literal {
	return cloudformation.GetAtt("AWS::EC2::EIP", *e.Name, "AllocationId")
}
