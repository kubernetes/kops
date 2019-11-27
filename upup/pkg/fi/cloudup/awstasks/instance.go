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

	"encoding/base64"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

// MaxUserDataSize is the max size of the userdata
const MaxUserDataSize = 16384

// Instance defines the instance specification
type Instance struct {
	ID        *string
	Lifecycle *fi.Lifecycle

	UserData fi.Resource

	Subnet           *Subnet
	PrivateIPAddress *string

	Name *string
	Tags map[string]string

	Shared *bool

	ImageID            *string
	InstanceType       *string
	SSHKey             *SSHKey
	SecurityGroups     []*SecurityGroup
	AssociatePublicIP  *bool
	IAMInstanceProfile *IAMInstanceProfile
}

var _ fi.CompareWithID = &Instance{}

func (s *Instance) CompareWithID() *string {
	return s.ID
}

func (e *Instance) Find(c *fi.Context) (*Instance, error) {
	cloud := c.Cloud.(awsup.AWSCloud)
	var request *ec2.DescribeInstancesInput

	if fi.BoolValue(e.Shared) {
		var instanceIds []*string
		instanceIds = append(instanceIds, e.ID)
		request = &ec2.DescribeInstancesInput{
			InstanceIds: instanceIds,
		}
	} else {
		filters := cloud.BuildFilters(e.Name)
		filters = append(filters, awsup.NewEC2Filter("instance-state-name", "pending", "running", "stopping", "stopped"))
		request = &ec2.DescribeInstancesInput{
			Filters: filters,
		}
	}

	response, err := cloud.EC2().DescribeInstances(request)
	if err != nil {
		return nil, fmt.Errorf("error listing instances: %v", err)
	}

	instances := []*ec2.Instance{}
	if response != nil {
		for _, reservation := range response.Reservations {
			instances = append(instances, reservation.Instances...)
		}
	}

	if len(instances) == 0 {
		return nil, nil
	}

	if len(instances) != 1 {
		return nil, fmt.Errorf("found multiple Instances with name: %s", *e.Name)
	}

	klog.V(2).Info("found existing instance")
	i := instances[0]

	if i.InstanceId == nil {
		return nil, fmt.Errorf("found instance, but InstanceId was nil")
	}

	actual := &Instance{
		ID:               i.InstanceId,
		PrivateIPAddress: i.PrivateIpAddress,
		InstanceType:     i.InstanceType,
		ImageID:          i.ImageId,
		Name:             findNameTag(i.Tags),
	}

	// Fetch instance UserData
	{
		request := &ec2.DescribeInstanceAttributeInput{}
		request.InstanceId = i.InstanceId
		request.Attribute = aws.String("userData")
		response, err := cloud.EC2().DescribeInstanceAttribute(request)
		if err != nil {
			return nil, fmt.Errorf("error querying EC2 for user metadata for instance %q: %v", *i.InstanceId, err)
		}
		if response.UserData != nil {
			b, err := base64.StdEncoding.DecodeString(aws.StringValue(response.UserData.Value))
			if err != nil {
				return nil, fmt.Errorf("error decoding EC2 UserData: %v", err)
			}
			actual.UserData = fi.NewBytesResource(b)
		}
	}

	if i.SubnetId != nil {
		actual.Subnet = &Subnet{ID: i.SubnetId}
	}
	if i.KeyName != nil {
		actual.SSHKey = &SSHKey{Name: i.KeyName}
	}

	for _, sg := range i.SecurityGroups {
		actual.SecurityGroups = append(actual.SecurityGroups, &SecurityGroup{ID: sg.GroupId})
	}

	associatePublicIpAddress := false
	for _, ni := range i.NetworkInterfaces {
		if aws.StringValue(ni.Association.PublicIp) != "" {
			associatePublicIpAddress = true
		}
	}
	actual.AssociatePublicIP = &associatePublicIpAddress

	if i.IamInstanceProfile != nil {
		actual.IAMInstanceProfile = &IAMInstanceProfile{Name: nameFromIAMARN(i.IamInstanceProfile.Arn)}
	}

	actual.Tags = intersectTags(i.Tags, e.Tags)

	actual.Lifecycle = e.Lifecycle
	actual.Shared = e.Shared

	e.ID = actual.ID

	// Avoid spurious changes on ImageId
	if e.ImageID != nil && actual.ImageID != nil && *actual.ImageID != *e.ImageID {
		image, err := cloud.ResolveImage(*e.ImageID)
		if err != nil {
			klog.Warningf("unable to resolve image: %q: %v", *e.ImageID, err)
		} else if image == nil {
			klog.Warningf("unable to resolve image: %q: not found", *e.ImageID)
		} else if aws.StringValue(image.ImageId) == *actual.ImageID {
			klog.V(4).Infof("Returning matching ImageId as expected name: %q -> %q", *actual.ImageID, *e.ImageID)
			actual.ImageID = e.ImageID
		}
	}

	return actual, nil
}

func nameFromIAMARN(arn *string) *string {
	if arn == nil {
		return nil
	}
	tokens := strings.Split(*arn, ":")
	last := tokens[len(tokens)-1]

	if !strings.HasPrefix(last, "instance-profile/") {
		klog.Warningf("Unexpected ARN for instance profile: %q", *arn)
	}

	name := strings.TrimPrefix(last, "instance-profile/")
	return &name
}

func (e *Instance) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *Instance) CheckChanges(a, e, changes *Instance) error {
	if a != nil {
		if !fi.BoolValue(e.Shared) && e.Name == nil {
			return fi.RequiredField("Name")
		}
	}
	return nil
}

func (_ *Instance) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *Instance) error {
	if a == nil {

		if fi.BoolValue(e.Shared) {
			return fmt.Errorf("NAT EC2 Instance %q not found", fi.StringValue(e.ID))
		}

		if e.ImageID == nil {
			return fi.RequiredField("ImageID")
		}
		image, err := t.Cloud.ResolveImage(fi.StringValue(e.ImageID))
		if err != nil {
			return err
		}

		klog.V(2).Infof("Creating Instance with Name:%q", fi.StringValue(e.Name))
		request := &ec2.RunInstancesInput{
			ImageId:      image.ImageId,
			InstanceType: e.InstanceType,
			MinCount:     aws.Int64(1),
			MaxCount:     aws.Int64(1),
		}

		if e.SSHKey != nil {
			request.KeyName = e.SSHKey.Name
		}

		securityGroupIDs := []*string{}
		for _, sg := range e.SecurityGroups {
			securityGroupIDs = append(securityGroupIDs, sg.ID)
		}
		request.NetworkInterfaces = []*ec2.InstanceNetworkInterfaceSpecification{
			{
				DeviceIndex:              aws.Int64(0),
				AssociatePublicIpAddress: e.AssociatePublicIP,
				SubnetId:                 e.Subnet.ID,
				PrivateIpAddress:         e.PrivateIPAddress,
				Groups:                   securityGroupIDs,
			},
		}

		// Build up the actual block device mappings
		// TODO: Support RootVolumeType & RootVolumeSize (see launchconfiguration)
		blockDeviceMappings, err := buildEphemeralDevices(t.Cloud, fi.StringValue(e.InstanceType))
		if err != nil {
			return err
		}

		if len(blockDeviceMappings) != 0 {
			request.BlockDeviceMappings = []*ec2.BlockDeviceMapping{}
			for deviceName, bdm := range blockDeviceMappings {
				request.BlockDeviceMappings = append(request.BlockDeviceMappings, bdm.ToEC2(deviceName))
			}
		}

		if e.UserData != nil {
			d, err := fi.ResourceAsBytes(e.UserData)
			if err != nil {
				return fmt.Errorf("error rendering Instance UserData: %v", err)
			}
			if len(d) > MaxUserDataSize {
				// TODO: Re-enable gzip?
				// But it exposes some bugs in the AWS console, so if we can avoid it, we should
				//d, err = fi.GzipBytes(d)
				//if err != nil {
				//	return fmt.Errorf("error while gzipping UserData: %v", err)
				//}
				return fmt.Errorf("Instance UserData was too large (%d bytes)", len(d))
			}
			request.UserData = aws.String(base64.StdEncoding.EncodeToString(d))
		}

		if e.IAMInstanceProfile != nil {
			request.IamInstanceProfile = &ec2.IamInstanceProfileSpecification{
				Name: e.IAMInstanceProfile.Name,
			}
		}

		response, err := t.Cloud.EC2().RunInstances(request)
		if err != nil {
			return fmt.Errorf("error creating Instance: %v", err)
		}

		e.ID = response.Instances[0].InstanceId
	}

	return t.AddAWSTags(*e.ID, e.Tags)
}

func (e *Instance) TerraformLink() *terraform.Literal {
	if fi.BoolValue(e.Shared) {
		if e.ID == nil {
			klog.Fatalf("ID must be set, if NAT Instance is shared: %s", e)
		}

		return terraform.LiteralFromStringValue(*e.ID)
	}

	return terraform.LiteralSelfLink("aws_instance", *e.Name)
}
