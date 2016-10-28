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
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	//"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"fmt"
)

//go:generate fitask -type=NATGateway
type NatGateway struct {
	Name         *string
	ElasticIp    *ElasticIP
	Subnet       *Subnet
	ID *string
}

var _ fi.CompareWithID = &NatGateway{} // Validate the IDs

func (e *NatGateway) CompareWithID() *string {
	s := ""
	return &s
}

func (e *NatGateway) Find(c *fi.Context) (*NatGateway, error) {
	return nil, nil
}

func (s *NatGateway) CheckChanges(a, e, changes *NatGateway) error {
	return nil
}

func (e *NatGateway) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *NatGateway) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *NatGateway) error {

	// New NGW
	var id *string
	if a == nil {
		glog.V(2).Infof("Creating Nat Gateway")

		request := &ec2.CreateNatGatewayInput{}
		request.AllocationId = e.ElasticIp.ID
		request.SubnetId = e.Subnet.ID
		response, err := t.Cloud.EC2().CreateNatGateway(request)
		if err != nil {
			return fmt.Errorf("Error creating Nat Gateway: %v", err)
		}
		e.ID = response.NatGateway.NatGatewayId
		id = e.ID
	}else {
		id = a.ID
	}

	// Tag the associated subnet
	if e.Subnet == nil {
		return  fmt.Errorf("Subnet not set")
	} else if e.Subnet.ID == nil {
		return  fmt.Errorf("Subnet ID not set")
	}
	tags := make(map[string]string)
	tags["AssociatedNatgateway"] = *id
	err := t.AddAWSTags(*e.Subnet.ID, tags)
	if err != nil {
		return fmt.Errorf("Unable to tag subnet %v", err)
	}
	return nil
}

// TODO Kris - We need to support NGW for Terraform

//type terraformNATGateway struct {
//	AllocationId *string           `json:"AllocationID,omitempty"`
//	SubnetID     *bool             `json:"SubnetID,omitempty"`
//}
//
//func (_ *NATGateway) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *NATGateway) error {
//	//	cloud := t.Cloud.(awsup.AWSCloud)
//
//	tf := &terraformNatGateway{
//		AllocationId:  e.AllocationID,
//		//SubnetID:      e.SubnetID,
//	}
//
//	return t.RenderResource("aws_natgateway", *e.AllocationID, tf)
//}
//
//func (e *NATGateway) TerraformLink() *terraform.Literal {
//	return terraform.LiteralProperty("aws_natgateway", *e.AllocationID, "id")
//}