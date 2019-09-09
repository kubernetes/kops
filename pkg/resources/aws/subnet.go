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

package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"k8s.io/kops/pkg/resources"
)

func DumpSubnet(op *resources.DumpOperation, r *resources.Resource) error {
	data := make(map[string]interface{})
	data["id"] = r.ID
	data["type"] = r.Type
	data["raw"] = r.Obj
	op.Dump.Resources = append(op.Dump.Resources, data)

	ec2Subnet := r.Obj.(*ec2.Subnet)
	s := &resources.Subnet{
		ID:   aws.StringValue(ec2Subnet.SubnetId),
		Zone: aws.StringValue(ec2Subnet.AvailabilityZone),
	}
	op.Dump.Subnets = append(op.Dump.Subnets, s)

	return nil
}
