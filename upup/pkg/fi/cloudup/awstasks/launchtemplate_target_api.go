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
	"context"
	"encoding/base64"
	"fmt"
	"sort"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/klog/v2"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

// RenderAWS is responsible for performing creating / updating the launch template
func (t *LaunchTemplate) RenderAWS(c *awsup.AWSAPITarget, a, e, changes *LaunchTemplate) error {
	ctx := context.TODO()
	// @step: resolve the image id to an AMI for us
	image, err := c.Cloud.ResolveImage(fi.ValueOf(t.ImageID))
	if err != nil {
		return err
	}

	// @step: lets build the launch template data
	data := &ec2types.RequestLaunchTemplateData{
		DisableApiTermination: fi.PtrTo(false),
		EbsOptimized:          t.RootVolumeOptimization,
		ImageId:               image.ImageId,
		InstanceType:          fi.ValueOf(t.InstanceType),
		MetadataOptions: &ec2types.LaunchTemplateInstanceMetadataOptionsRequest{
			HttpPutResponseHopLimit: t.HTTPPutResponseHopLimit,
			HttpTokens:              fi.ValueOf(t.HTTPTokens),
			HttpProtocolIpv6:        fi.ValueOf(t.HTTPProtocolIPv6),
		},
		NetworkInterfaces: []ec2types.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest{
			{
				AssociatePublicIpAddress: t.AssociatePublicIP,
				DeleteOnTermination:      aws.Bool(true),
				DeviceIndex:              fi.PtrTo(int32(0)),
				Ipv6AddressCount:         t.IPv6AddressCount,
			},
		},
	}

	// @step: add the actual block device mappings
	rootDevices, err := t.buildRootDevice(c.Cloud)
	if err != nil {
		return fmt.Errorf("failed to build root device: %w", err)
	}
	ephemeralDevices, err := buildEphemeralDevices(c.Cloud, fi.ValueOf(t.InstanceType))
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
		data.NetworkInterfaces[0].Groups = append(data.NetworkInterfaces[0].Groups, fi.ValueOf(sg.ID))
	}
	// @step: add any tenancy details
	if t.Tenancy != nil {
		data.Placement = &ec2types.LaunchTemplatePlacementRequest{Tenancy: fi.ValueOf(t.Tenancy)}
	}
	// @step: set the instance monitoring
	data.Monitoring = &ec2types.LaunchTemplatesMonitoringRequest{Enabled: fi.PtrTo(false)}
	if t.InstanceMonitoring != nil {
		data.Monitoring = &ec2types.LaunchTemplatesMonitoringRequest{Enabled: t.InstanceMonitoring}
	}
	// @step: add the iam instance profile
	if t.IAMInstanceProfile != nil {
		data.IamInstanceProfile = &ec2types.LaunchTemplateIamInstanceProfileSpecificationRequest{
			Name: t.IAMInstanceProfile.Name,
		}
	}
	// @step: add the tags
	var tags []ec2types.Tag
	if len(t.Tags) > 0 {
		for k, v := range t.Tags {
			tags = append(tags, ec2types.Tag{
				Key:   aws.String(k),
				Value: aws.String(v),
			})
		}
		data.TagSpecifications = append(data.TagSpecifications, ec2types.LaunchTemplateTagSpecificationRequest{
			ResourceType: ec2types.ResourceTypeInstance,
			Tags:         tags,
		})
		data.TagSpecifications = append(data.TagSpecifications, ec2types.LaunchTemplateTagSpecificationRequest{
			ResourceType: ec2types.ResourceTypeVolume,
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
	if fi.ValueOf(t.SpotPrice) != "" {
		s := &ec2types.LaunchTemplateSpotMarketOptionsRequest{
			BlockDurationMinutes: t.SpotDurationInMinutes,
			MaxPrice:             t.SpotPrice,
		}
		if t.InstanceInterruptionBehavior != nil {
			s.InstanceInterruptionBehavior = fi.ValueOf(t.InstanceInterruptionBehavior)
		}
		data.InstanceMarketOptions = &ec2types.LaunchTemplateInstanceMarketOptionsRequest{
			MarketType:  ec2types.MarketTypeSpot,
			SpotOptions: s,
		}
	}
	if fi.ValueOf(t.CPUCredits) != "" {
		data.CreditSpecification = &ec2types.CreditSpecificationRequest{
			CpuCredits: t.CPUCredits,
		}
	}
	// @step: attempt to create the launch template
	if a == nil {
		input := &ec2.CreateLaunchTemplateInput{
			LaunchTemplateName: t.Name,
			LaunchTemplateData: data,
			TagSpecifications: []ec2types.TagSpecification{
				{
					ResourceType: ec2types.ResourceTypeLaunchTemplate,
					Tags:         tags,
				},
			},
		}
		output, err := c.Cloud.EC2().CreateLaunchTemplate(ctx, input)
		if err != nil || output.LaunchTemplate == nil {
			return fmt.Errorf("error creating LaunchTemplate %q: %v", fi.ValueOf(t.Name), err)
		}
		e.ID = output.LaunchTemplate.LaunchTemplateId
	} else {
		input := &ec2.CreateLaunchTemplateVersionInput{
			LaunchTemplateName: t.Name,
			LaunchTemplateData: data,
		}
		if version, err := c.Cloud.EC2().CreateLaunchTemplateVersion(ctx, input); err != nil {
			return fmt.Errorf("error creating LaunchTemplateVersion: %v", err)
		} else {
			newDefault := strconv.FormatInt(*version.LaunchTemplateVersion.VersionNumber, 10)
			input := &ec2.ModifyLaunchTemplateInput{
				DefaultVersion:   &newDefault,
				LaunchTemplateId: version.LaunchTemplateVersion.LaunchTemplateId,
			}
			if _, err := c.Cloud.EC2().ModifyLaunchTemplate(ctx, input); err != nil {
				return fmt.Errorf("error updating launch template version: %w", err)
			}
		}
		if changes.Tags != nil {
			err = c.UpdateTags(fi.ValueOf(a.ID), e.Tags)
			if err != nil {
				return fmt.Errorf("error updating LaunchTemplate tags: %v", err)
			}
		}
		e.ID = a.ID

	}

	return nil
}

// Find is responsible for finding the launch template for us
func (t *LaunchTemplate) Find(c *fi.CloudupContext) (*LaunchTemplate, error) {
	cloud, ok := c.T.Cloud.(awsup.AWSCloud)
	if !ok {
		return nil, fmt.Errorf("invalid cloud provider: %v, expected: %s", c.T.Cloud, "awsup.AWSCloud")
	}

	// @step: get the latest launch template version
	lt, err := t.findLatestLaunchTemplateVersion(c)
	if err != nil {
		return nil, err
	}
	if lt == nil {
		return nil, nil
	}

	klog.V(3).Infof("found existing LaunchTemplate: %s", fi.ValueOf(lt.LaunchTemplateName))

	actual := &LaunchTemplate{
		AssociatePublicIP:      fi.PtrTo(false),
		ID:                     lt.LaunchTemplateId,
		ImageID:                lt.LaunchTemplateData.ImageId,
		InstanceMonitoring:     fi.PtrTo(false),
		Lifecycle:              t.Lifecycle,
		Name:                   t.Name,
		RootVolumeOptimization: lt.LaunchTemplateData.EbsOptimized,
	}
	if len(lt.LaunchTemplateData.InstanceType) > 0 {
		actual.InstanceType = fi.PtrTo(lt.LaunchTemplateData.InstanceType)
	}

	// @step: check if any of the interfaces are public facing
	for _, x := range lt.LaunchTemplateData.NetworkInterfaces {
		if aws.ToBool(x.AssociatePublicIpAddress) {
			actual.AssociatePublicIP = fi.PtrTo(true)
		}
		for _, id := range x.Groups {
			actual.SecurityGroups = append(actual.SecurityGroups, &SecurityGroup{ID: fi.PtrTo(id)})
		}
		actual.IPv6AddressCount = x.Ipv6AddressCount
	}
	// In older Kops versions, security groups were added to LaunchTemplateData.SecurityGroupIds
	for _, id := range lt.LaunchTemplateData.SecurityGroupIds {
		actual.SecurityGroups = append(actual.SecurityGroups, &SecurityGroup{ID: fi.PtrTo("legacy-" + id)})
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
	if lt.LaunchTemplateData.Placement != nil && len(lt.LaunchTemplateData.Placement.Tenancy) > 0 {
		actual.Tenancy = fi.PtrTo(lt.LaunchTemplateData.Placement.Tenancy)
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
	if imo != nil && imo.SpotOptions != nil && aws.ToString(imo.SpotOptions.MaxPrice) != "" {
		actual.SpotPrice = imo.SpotOptions.MaxPrice
		actual.SpotDurationInMinutes = imo.SpotOptions.BlockDurationMinutes
		if len(imo.SpotOptions.InstanceInterruptionBehavior) > 0 {
			actual.InstanceInterruptionBehavior = fi.PtrTo(imo.SpotOptions.InstanceInterruptionBehavior)
		}
	} else {
		actual.SpotPrice = aws.String("")
	}

	// @step: get the image is order to find out the root device name as using the index
	// is not variable, under conditions they move
	image, err := cloud.ResolveImage(fi.ValueOf(t.ImageID))
	if err != nil {
		return nil, err
	}

	// @step: find the root volume
	for _, b := range lt.LaunchTemplateData.BlockDeviceMappings {
		if b.Ebs == nil {
			continue
		}
		if b.DeviceName != nil && fi.ValueOf(b.DeviceName) == fi.ValueOf(image.RootDeviceName) {
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
		ud, err := base64.StdEncoding.DecodeString(aws.ToString(lt.LaunchTemplateData.UserData))
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
		if len(options.HttpTokens) > 0 {
			actual.HTTPTokens = fi.PtrTo(options.HttpTokens)
		}
		if len(options.HttpProtocolIpv6) > 0 {
			actual.HTTPProtocolIPv6 = fi.PtrTo(options.HttpProtocolIpv6)
		}
	}

	// @step: to avoid spurious changes on ImageId
	if t.ImageID != nil && actual.ImageID != nil && *actual.ImageID != *t.ImageID {
		image, err := cloud.ResolveImage(*t.ImageID)
		if err != nil {
			klog.Warningf("unable to resolve image: %q: %v", *t.ImageID, err)
		} else if image == nil {
			klog.Warningf("unable to resolve image: %q: not found", *t.ImageID)
		} else if aws.ToString(image.ImageId) == *actual.ImageID {
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
func (t *LaunchTemplate) findAllLaunchTemplates(c *fi.CloudupContext) ([]ec2types.LaunchTemplate, error) {
	ctx := c.Context()

	cloud, ok := c.T.Cloud.(awsup.AWSCloud)
	if !ok {
		return nil, fmt.Errorf("invalid cloud provider: %v, expected: %s", c.T.Cloud, "awsup.AWSCloud")
	}

	input := &ec2.DescribeLaunchTemplatesInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []string{fi.ValueOf(t.Name)},
			},
		},
	}

	var list []ec2types.LaunchTemplate
	paginator := ec2.NewDescribeLaunchTemplatesPaginator(cloud.EC2(), input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing AutoScaling LaunchTemplates: %v", err)
		}
		list = append(list, page.LaunchTemplates...)
	}

	return list, nil
}

// findLatestLaunchTemplateVersion returns the latest template version
func (t *LaunchTemplate) findLatestLaunchTemplateVersion(c *fi.CloudupContext) (*ec2types.LaunchTemplateVersion, error) {
	ctx := c.Context()

	cloud, ok := c.T.Cloud.(awsup.AWSCloud)
	if !ok {
		return nil, fmt.Errorf("invalid cloud provider: %v, expected: awsup.AWSCloud", c.T.Cloud)
	}

	input := &ec2.DescribeLaunchTemplateVersionsInput{
		LaunchTemplateName: t.Name,
		Versions:           []string{("$Latest")},
	}

	output, err := cloud.EC2().DescribeLaunchTemplateVersions(ctx, input)
	if err != nil {
		if awsup.AWSErrorCode(err) == "InvalidLaunchTemplateName.NotFoundException" {
			klog.V(4).Infof("Got InvalidLaunchTemplateName.NotFoundException error describing latest launch template version: %q", aws.ToString(t.Name))
			return nil, nil
		} else {
			return nil, err
		}
	}

	if len(output.LaunchTemplateVersions) == 0 {
		return nil, nil
	}

	return &output.LaunchTemplateVersions[0], nil
}

// deleteLaunchTemplate tracks a LaunchConfiguration that we're going to delete
// It implements fi.CloudupDeletion
type deleteLaunchTemplate struct {
	lc *ec2types.LaunchTemplate
}

var _ fi.CloudupDeletion = &deleteLaunchTemplate{}

// TaskName returns the task name
func (d *deleteLaunchTemplate) TaskName() string {
	return "LaunchTemplate"
}

// Item returns the launch template name
func (d *deleteLaunchTemplate) Item() string {
	return fi.ValueOf(d.lc.LaunchTemplateName)
}

func (d *deleteLaunchTemplate) Delete(t fi.CloudupTarget) error {
	ctx := context.TODO()
	awsTarget, ok := t.(*awsup.AWSAPITarget)
	if !ok {
		return fmt.Errorf("unexpected target type for deletion: %T", t)
	}

	if _, err := awsTarget.Cloud.EC2().DeleteLaunchTemplate(ctx, &ec2.DeleteLaunchTemplateInput{
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

func (d *deleteLaunchTemplate) DeferDeletion() bool {
	return false // TODO: Should we defer deletion?
}
