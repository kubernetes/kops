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
	//"github.com/aws/aws-sdk-go/service/ec2"
	//"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	//"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=NATGateway
type NATGateway struct {
	Name         *string
	ElasticIp    *string
	Subnet       *string
}

var _ fi.CompareWithID = &NATGateway{} // Validate the IDs

func (e *NATGateway) CompareWithID() *string {
	s := ""
	return &s
}

func (e *NATGateway) Find(c *fi.Context) (*NATGateway, error) {
	return &NATGateway{}, nil
}

func (s *NATGateway) CheckChanges(a, e, changes *NATGateway) error {
	return nil
}

func (e *NATGateway) Run(c *fi.Context) error {
	return nil
}

func (_ *NATGateway) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *NATGateway) error {
	return nil
}

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