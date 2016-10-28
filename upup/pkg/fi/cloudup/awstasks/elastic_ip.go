/*
Copyright 2016 The Kubernetes Authors.

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
	//"fmt"
	//
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"fmt"
)

//go:generate fitask -type=ElasticIP

// Elastic IP
// Representation the EIP AWS task
type ElasticIP struct {
	Name                  *string
	ID                    *string
	PublicIP              *string



	// Allow support for assicated subnets
	// If you need another resource you must add it

	AssociatedSubnet   *Subnet

	//AssociatedSubnetTagId *string

	//AssociatedElbTag      *string
	//AssociatedElbTagId    *string

}

var _ fi.HasAddress = &ElasticIP{}

func (e *ElasticIP) FindAddress(context *fi.Context) (*string, error) {
	actual, err := e.find(context.Cloud.(awsup.AWSCloud))
	if err != nil {
		return nil, fmt.Errorf("error querying for ElasticIP: %v", err)
	}
	if actual == nil {
		return nil, nil
	}
	return actual.PublicIP, nil
}

//
// Find (public wrapper for find()
//
func (e *ElasticIP) Find(context *fi.Context) (*ElasticIP, error) {
	return e.find(context.Cloud.(awsup.AWSCloud))
}

// Will attempt to look up the elastic IP from AWS
func (e *ElasticIP) find(cloud awsup.AWSCloud) (*ElasticIP, error) {
	publicIP := e.PublicIP
	allocationID := e.ID

	// Find via tag on foreign resource
	if allocationID == nil && publicIP == nil && e.AssociatedSubnet.ID != nil {
		var filters []*ec2.Filter
		filters = append(filters, awsup.NewEC2Filter("key", "AssociatedElasticIp"))
		filters = append(filters, awsup.NewEC2Filter("resource-id", e.AssociatedSubnet.ID))

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
		glog.V(2).Infof("Found public IP via tag: %v", *publicIP)
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

		actual.AssociatedSubnet = e.AssociatedSubnet

		e.ID = actual.ID

		return actual, nil
	}
	return nil, nil
}

// The Run() function is called to execute this task.
// This is the main entry point of the task, and will actually
// connect our internal resource representation to an actual
// resource in AWS
func (e *ElasticIP) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

// Validation for the resource. EIPs are simple, so virtually no
// validation
func (s *ElasticIP) CheckChanges(a, e, changes *ElasticIP) error {
	// This is a new EIP
	if a == nil {
		// No logic for EIPs - they are just created
	}

	// This is an existing EIP
	// We should never be changing this
	if a != nil {
		if changes.PublicIP != nil {
			return fi.CannotChangeField("PublicIP")
		}
		if changes.AssociatedSubnet != nil {
			return fi.CannotChangeField("AssociatedSubnet")
		}
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
	}
	return nil
}

// Here is where we actually apply changes to AWS
func (_ *ElasticIP) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *ElasticIP) error {

	var publicIp *string
	var eipId *string

	// If this is a new ElasticIP
	if a == nil {
		glog.V(2).Infof("Creating ElasticIP for VPC")

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
	}else {
		publicIp = a.PublicIP
		eipId = a.ID
	}


	// Tag the associated subnet
	if e.AssociatedSubnet == nil {
		return  fmt.Errorf("Subnet not set")
	} else if e.AssociatedSubnet.ID == nil {
		return  fmt.Errorf("Subnet ID not set")
	}
	tags := make(map[string]string)
	tags["AssociatedElasticIp"] = *publicIp
	tags["AssociatedElasticIpAllocationId"] = *eipId // Leaving this in for reference, even though we don't use it
	err := t.AddAWSTags(*e.AssociatedSubnet.ID, tags)
	if err != nil {
		return fmt.Errorf("Unable to tag subnet %v", err)
	}
	return nil
}
