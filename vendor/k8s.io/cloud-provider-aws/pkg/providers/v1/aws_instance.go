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

package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cloud-provider-aws/pkg/providers/v1/iface"
)

type awsInstance struct {
	ec2 iface.EC2

	// id in AWS
	awsID string

	// node name in k8s
	nodeName types.NodeName

	// availability zone the instance resides in
	availabilityZone string

	// ID of VPC the instance resides in
	vpcID string

	// ID of subnet the instance resides in
	subnetID string

	// instance type
	instanceType string
}

// newAWSInstance creates a new awsInstance object
func newAWSInstance(ec2Service iface.EC2, instance *ec2types.Instance) *awsInstance {
	az := ""
	if instance.Placement != nil {
		az = aws.ToString(instance.Placement.AvailabilityZone)
	}
	self := &awsInstance{
		ec2:              ec2Service,
		awsID:            aws.ToString(instance.InstanceId),
		nodeName:         mapInstanceToNodeName(instance),
		availabilityZone: az,
		instanceType:     string(instance.InstanceType),
		vpcID:            aws.ToString(instance.VpcId),
		subnetID:         aws.ToString(instance.SubnetId),
	}

	return self
}

// Gets the full information about this instance from the EC2 API
func (i *awsInstance) describeInstance(ctx context.Context) (*ec2types.Instance, error) {
	return describeInstance(ctx, i.ec2, InstanceID(i.awsID))
}
