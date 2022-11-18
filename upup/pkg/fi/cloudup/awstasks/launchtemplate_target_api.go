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
	"encoding/base64"
	"fmt"
	"sort"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog/v2"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

// RenderAWS is responsible for performing creating / updating the launch template
func (t *LaunchTemplate) RenderAWS(c *awsup.AWSAPITarget, a, e, changes *LaunchTemplate) error {
	// @step: resolve the image id to an AMI for us
	image, err := c.Cloud.ResolveImage(fi.StringValue(t.ImageID))
	if err != nil {
		return err
	}

	// @step: lets build the launch template data
	data := &ec2.RequestLaunchTemplateData{
		DisableApiTermination: fi.PtrTo(false),
		EbsOptimized:          t.RootVolumeOptimization,
		ImageId:               image.ImageId,
		InstanceType:          t.InstanceType,
		MetadataOptions: &ec2.LaunchTemplateInstanceMetadataOptionsRequest{
			HttpPutResponseHopLimit: t.HTTPPutResponseHopLimit,
			HttpTokens:              t.HTTPTokens,
			HttpProtocolIpv6:        t.HTTPProtocolIPv6,
		},
		NetworkInterfaces: []*ec2.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest{
			{
				AssociatePublicIpAddress: t.AssociatePublicIP,
				DeleteOnTermination:      aws.Bool(true),
				DeviceIndex:              fi.PtrTo(int64(0)),
				Ipv6AddressCount:         t.IPv6AddressCount,
			},
		},
	}

	// @step: add the actual block device mappings
	rootDevices, err := t.buildRootDevice(c.Cloud)
	if err != nil {
		return fmt.Errorf("failed to build root device: %w", err)
	}
	ephemeralDevices, err := buildEphemeralDevices(c.Cloud, fi.StringValue(t.InstanceType))
	if err != nil {
		return fmt.Errorf("failed to build ephemeral devices: %w", err)
	}
	additionalDevices, err := buildAdditionalDevices(t.BlockDeviceMappings)
	if err != nil {
		return err
	}
	for _, x := range []map[string]*BlockDeviceMapping{rootDevices, ephemeralDevices, additionalDevices} {
		for name, device := range x {
			data.BlockDeviceMappings = append(data.BlockDeviceMappings, device.ToLaunchTemplateBootDeviceRequest(name))
		}
	}

	// @step: add the ssh key
	if t.SSHKey != nil {
		data.KeyName = t.SSHKey.Name
	}
	// @step: add the security groups
	for _, sg := range t.SecurityGroups {
		data.NetworkInterfaces[0].Groups = append(data.NetworkInterfaces[0].Groups, sg.ID)
	}
	// @step: add any tenancy details
	if t.Tenancy != nil {
		data.Placement = &ec2.LaunchTemplatePlacementRequest{Tenancy: t.Tenancy}
	}
	// @step: set the instance monitoring
	data.Monitoring = &ec2.LaunchTemplatesMonitoringRequest{Enabled: fi.PtrTo(false)}
	if t.InstanceMonitoring != nil {
		data.Monitoring = &ec2.LaunchTemplatesMonitoringRequest{Enabled: t.InstanceMonitoring}
	}
	// @step: add the iam instance profile
	if t.IAMInstanceProfile != nil {
		data.IamInstanceProfile = &ec2.LaunchTemplateIamInstanceProfileSpecificationRequest{
			Name: t.IAMInstanceProfile.Name,
		}
	}
	// @step: add the tags
	var tags []*ec2.Tag
	if len(t.Tags) > 0 {
		for k, v := range t.Tags {
			tags = append(tags, &ec2.Tag{
				Key:   aws.String(k),
				Value: aws.String(v),
			})
		}
		data.TagSpecifications = append(data.TagSpecifications, &ec2.LaunchTemplateTagSpecificationRequest{
			ResourceType: aws.String(ec2.ResourceTypeInstance),
			Tags:         tags,
		})
		data.TagSpecifications = append(data.TagSpecifications, &ec2.LaunchTemplateTagSpecificationRequest{
			ResourceType: aws.String(ec2.ResourceTypeVolume),
			Tags:         tags,
		})
	}
	// @step: add the userdata
	if t.UserData != nil {
		d, err := fi.ResourceAsBytes(t.UserData)
		if err != nil {
			return fmt.Errorf("error rendering LaunchTemplate UserData: %v", err)
		}
		data.UserData = aws.String(base64.StdEncoding.EncodeToString(d))
	}
	// @step: add market options
	if fi.StringValue(t.SpotPrice) != "" {
		s := &ec2.LaunchTemplateSpotMarketOptionsRequest{
			BlockDurationMinutes:         t.SpotDurationInMinutes,
			InstanceInterruptionBehavior: t.InstanceInterruptionBehavior,
			MaxPrice:                     t.SpotPrice,
		}
		data.InstanceMarketOptions = &ec2.LaunchTemplateInstanceMarketOptionsRequest{
			MarketType:  fi.PtrTo("spot"),
			SpotOptions: s,
		}
	}
	if fi.StringValue(t.CPUCredits) != "" {
		data.CreditSpecification = &ec2.CreditSpecificationRequest{
			CpuCredits: t.CPUCredits,
		}
	}
	// @step: attempt to create the launch template
	if a == nil {
		input := &ec2.CreateLaunchTemplateInput{
			LaunchTemplateName: t.Name,
			LaunchTemplateData: data,
			TagSpecifications: []*ec2.TagSpecification{
				{
					ResourceType: aws.String(ec2.ResourceTypeLaunchTemplate),
					Tags:         tags,
				},
			},
		}
		output, err := c.Cloud.EC2().CreateLaunchTemplate(input)
		if err != nil || output.LaunchTemplate == nil {
			return fmt.Errorf("error creating LaunchTemplate %q: %v", fi.StringValue(t.Name), err)
		}
		e.ID = output.LaunchTemplate.LaunchTemplateId
	} else {
		input := &ec2.CreateLaunchTemplateVersionInput{
			LaunchTemplateName: t.Name,
			LaunchTemplateData: data,
		}
		if version, err := c.Cloud.EC2().CreateLaunchTemplateVersion(input); err != nil {
			return fmt.Errorf("error creating LaunchTemplateVersion: %v", err)
		} else {
			newDefault := strconv.FormatInt(*version.LaunchTemplateVersion.VersionNumber, 10)
			input := &ec2.ModifyLaunchTemplateInput{
				DefaultVersion:   &newDefault,
				LaunchTemplateId: version.LaunchTemplateVersion.LaunchTemplateId,
			}
			if _, err := c.Cloud.EC2().ModifyLaunchTemplate(input); err != nil {
				return fmt.Errorf("error updating launch template version: %w", err)
			}
		}
		if changes.Tags != nil {
			err = c.UpdateTags(fi.StringValue(a.ID), e.Tags)
			if err != nil {
				return fmt.Errorf("error updating LaunchTemplate tags: %v", err)
			}
		}
		e.ID = a.ID

	}

	return nil
}

// Find is responsible for finding the launch template for us
func (t *LaunchTemplate) Find(c *fi.Context) (*LaunchTemplate, error) {
	cloud, ok := c.Cloud.(awsup.AWSCloud)
	if !ok {
		return nil, fmt.Errorf("invalid cloud provider: %v, expected: %s", c.Cloud, "awsup.AWSCloud")
	}

	// @step: get the latest launch template version
	lt, err := t.findLatestLaunchTemplateVersion(c)
	if err != nil {
		return nil, err
	}
	if lt == nil {
		return nil, nil
	}

	klog.V(3).Infof("found existing LaunchTemplate: %s", fi.StringValue(lt.LaunchTemplateName))

	actual := &LaunchTemplate{
		AssociatePublicIP:      fi.PtrTo(false),
		ID:                     lt.LaunchTemplateId,
		ImageID:                lt.LaunchTemplateData.ImageId,
		InstanceMonitoring:     fi.PtrTo(false),
		InstanceType:           lt.LaunchTemplateData.InstanceType,
		Lifecycle:              t.Lifecycle,
		Name:                   t.Name,
		RootVolumeOptimization: lt.LaunchTemplateData.EbsOptimized,
	}

	// @step: check if any of the interfaces are public facing
	for _, x := range lt.LaunchTemplateData.NetworkInterfaces {
		if aws.BoolValue(x.AssociatePublicIpAddress) {
			actual.AssociatePublicIP = fi.PtrTo(true)
		}
		for _, id := range x.Groups {
			actual.SecurityGroups = append(actual.SecurityGroups, &SecurityGroup{ID: id})
		}
		actual.IPv6AddressCount = x.Ipv6AddressCount
	}
	// In older Kops versions, security groups were added to LaunchTemplateData.SecurityGroupIds
	for _, id := range lt.LaunchTemplateData.SecurityGroupIds {
		actual.SecurityGroups = append(actual.SecurityGroups, &SecurityGroup{ID: fi.PtrTo("legacy-" + *id)})
	}
	sort.Sort(OrderSecurityGroupsById(actual.SecurityGroups))

	if lt.LaunchTemplateData.CreditSpecification != nil && lt.LaunchTemplateData.CreditSpecification.CpuCredits != nil {
		actual.CPUCredits = lt.LaunchTemplateData.CreditSpecification.CpuCredits
	} else {
		actual.CPUCredits = aws.String("")
	}
	// @step: check if monitoring it enabled
	if lt.LaunchTemplateData.Monitoring != nil {
		actual.InstanceMonitoring = lt.LaunchTemplateData.Monitoring.Enabled
	}
	// @step: add the tenancy
	if lt.LaunchTemplateData.Placement != nil {
		actual.Tenancy = lt.LaunchTemplateData.Placement.Tenancy
	}
	// @step: add the ssh if there is one
	if lt.LaunchTemplateData.KeyName != nil {
		actual.SSHKey = &SSHKey{Name: lt.LaunchTemplateData.KeyName}
	}
	// @step: add a instance if there is one
	if lt.LaunchTemplateData.IamInstanceProfile != nil {
		actual.IAMInstanceProfile = &IAMInstanceProfile{Name: lt.LaunchTemplateData.IamInstanceProfile.Name}
	}
	// @step: add InstanceMarketOptions if there are any
	imo := lt.LaunchTemplateData.InstanceMarketOptions
	if imo != nil && imo.SpotOptions != nil && aws.StringValue(imo.SpotOptions.MaxPrice) != "" {
		actual.SpotPrice = imo.SpotOptions.MaxPrice
		actual.SpotDurationInMinutes = imo.SpotOptions.BlockDurationMinutes
		actual.InstanceInterruptionBehavior = imo.SpotOptions.InstanceInterruptionBehavior
	} else {
		actual.SpotPrice = aws.String("")
	}

	// @step: get the image is order to find out the root device name as using the index
	// is not variable, under conditions they move
	image, err := cloud.ResolveImage(fi.StringValue(t.ImageID))
	if err != nil {
		return nil, err
	}

	// @step: find the root volume
	for _, b := range lt.LaunchTemplateData.BlockDeviceMappings {
		if b.Ebs == nil {
			continue
		}
		if b.DeviceName != nil && fi.StringValue(b.DeviceName) == fi.StringValue(image.RootDeviceName) {
			actual.RootVolumeSize = b.Ebs.VolumeSize
			actual.RootVolumeType = b.Ebs.VolumeType
			actual.RootVolumeIops = b.Ebs.Iops
			actual.RootVolumeThroughput = b.Ebs.Throughput
			actual.RootVolumeEncryption = b.Ebs.Encrypted
			if b.Ebs.KmsKeyId != nil {
				actual.RootVolumeKmsKey = b.Ebs.KmsKeyId
			} else {
				actual.RootVolumeKmsKey = fi.PtrTo("")
			}
		} else {
			_, d := BlockDeviceMappingFromLaunchTemplateBootDeviceRequest(b)
			actual.BlockDeviceMappings = append(actual.BlockDeviceMappings, d)
		}
	}

	if lt.LaunchTemplateData.UserData != nil {
		ud, err := base64.StdEncoding.DecodeString(aws.StringValue(lt.LaunchTemplateData.UserData))
		if err != nil {
			return nil, fmt.Errorf("error decoding userdata: %s", err)
		}
		actual.UserData = fi.NewStringResource(string(ud))
	}

	// @step: add tags
	if len(lt.LaunchTemplateData.TagSpecifications) > 0 {
		ts := lt.LaunchTemplateData.TagSpecifications[0]
		if ts.Tags != nil {
			tags := mapEC2TagsToMap(ts.Tags)
			actual.Tags = tags
		}
	}

	// @step: add instance metadata options
	if options := lt.LaunchTemplateData.MetadataOptions; options != nil {
		actual.HTTPPutResponseHopLimit = options.HttpPutResponseHopLimit
		actual.HTTPTokens = options.HttpTokens
		actual.HTTPProtocolIPv6 = options.HttpProtocolIpv6
	}

	// @step: to avoid spurious changes on ImageId
	if t.ImageID != nil && actual.ImageID != nil && *actual.ImageID != *t.ImageID {
		image, err := cloud.ResolveImage(*t.ImageID)
		if err != nil {
			klog.Warningf("unable to resolve image: %q: %v", *t.ImageID, err)
		} else if image == nil {
			klog.Warningf("unable to resolve image: %q: not found", *t.ImageID)
		} else if aws.StringValue(image.ImageId) == *actual.ImageID {
			klog.V(4).Infof("Returning matching ImageId as expected name: %q -> %q", *actual.ImageID, *t.ImageID)
			actual.ImageID = t.ImageID
		}
	}

	if t.ID == nil {
		t.ID = actual.ID
	}

	return actual, nil
}

// findAllLaunchTemplates returns all the launch templates for us
func (t *LaunchTemplate) findAllLaunchTemplates(c *fi.Context) ([]*ec2.LaunchTemplate, error) {
	cloud, ok := c.Cloud.(awsup.AWSCloud)
	if !ok {
		return nil, fmt.Errorf("invalid cloud provider: %v, expected: %s", c.Cloud, "awsup.AWSCloud")
	}

	input := &ec2.DescribeLaunchTemplatesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []*string{t.Name},
			},
		},
	}

	var list []*ec2.LaunchTemplate
	err := cloud.EC2().DescribeLaunchTemplatesPages(input, func(p *ec2.DescribeLaunchTemplatesOutput, lastPage bool) (shouldContinue bool) {
		list = append(list, p.LaunchTemplates...)
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("error listing AutoScaling LaunchTemplates: %v", err)
	}

	return list, nil
}

// findLatestLaunchTemplateVersion returns the latest template version
func (t *LaunchTemplate) findLatestLaunchTemplateVersion(c *fi.Context) (*ec2.LaunchTemplateVersion, error) {
	cloud, ok := c.Cloud.(awsup.AWSCloud)
	if !ok {
		return nil, fmt.Errorf("invalid cloud provider: %v, expected: awsup.AWSCloud", c.Cloud)
	}

	input := &ec2.DescribeLaunchTemplateVersionsInput{
		LaunchTemplateName: t.Name,
		Versions:           []*string{aws.String("$Latest")},
	}

	output, err := cloud.EC2().DescribeLaunchTemplateVersions(input)
	if err != nil {
		if awsup.AWSErrorCode(err) == "InvalidLaunchTemplateName.NotFoundException" {
			klog.V(4).Infof("Got InvalidLaunchTemplateName.NotFoundException error describing latest launch template version: %q", aws.StringValue(t.Name))
			return nil, nil
		} else {
			return nil, err
		}
	}

	if len(output.LaunchTemplateVersions) == 0 {
		return nil, nil
	}

	return output.LaunchTemplateVersions[0], nil
}

// deleteLaunchTemplate tracks a LaunchConfiguration that we're going to delete
// It implements fi.Deletion
type deleteLaunchTemplate struct {
	lc *ec2.LaunchTemplate
}

var _ fi.Deletion = &deleteLaunchTemplate{}

// TaskName returns the task name
func (d *deleteLaunchTemplate) TaskName() string {
	return "LaunchTemplate"
}

// Item returns the launch template name
func (d *deleteLaunchTemplate) Item() string {
	return fi.StringValue(d.lc.LaunchTemplateName)
}

func (d *deleteLaunchTemplate) Delete(t fi.Target) error {
	awsTarget, ok := t.(*awsup.AWSAPITarget)
	if !ok {
		return fmt.Errorf("unexpected target type for deletion: %T", t)
	}

	if _, err := awsTarget.Cloud.EC2().DeleteLaunchTemplate(&ec2.DeleteLaunchTemplateInput{
		LaunchTemplateName: d.lc.LaunchTemplateName,
	}); err != nil {
		return fmt.Errorf("error deleting LaunchTemplate %s: error: %s", d.Item(), err)
	}

	return nil
}

// String returns a string representation of the task
func (d *deleteLaunchTemplate) String() string {
	return d.TaskName() + "-" + d.Item()
}
