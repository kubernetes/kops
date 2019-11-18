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
	"strings"
	"time"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog"
)

// RenderAWS is responsible for performing creating / updating the launch template
func (t *LaunchTemplate) RenderAWS(c *awsup.AWSAPITarget, a, ep, changes *LaunchTemplate) error {
	name := t.LaunchTemplateName()

	// @step: resolve the image id to an AMI for us
	image, err := c.Cloud.ResolveImage(fi.StringValue(t.ImageID))
	if err != nil {
		return err
	}

	// @step: lets build the launch template input
	input := &ec2.CreateLaunchTemplateInput{
		LaunchTemplateData: &ec2.RequestLaunchTemplateData{
			DisableApiTermination: fi.Bool(false),
			EbsOptimized:          t.RootVolumeOptimization,
			ImageId:               image.ImageId,
			InstanceType:          t.InstanceType,
		},
		LaunchTemplateName: aws.String(name),
	}
	lc := input.LaunchTemplateData

	// @step: add the actual block device mappings
	rootDevices, err := t.buildRootDevice(c.Cloud)
	if err != nil {
		return err
	}
	ephemeralDevices, err := buildEphemeralDevices(c.Cloud, fi.StringValue(t.InstanceType))
	if err != nil {
		return err
	}
	additionalDevices, err := buildAdditionalDevices(t.BlockDeviceMappings)
	if err != nil {
		return err
	}
	for _, x := range []map[string]*BlockDeviceMapping{rootDevices, ephemeralDevices, additionalDevices} {
		for name, device := range x {
			input.LaunchTemplateData.BlockDeviceMappings = append(input.LaunchTemplateData.BlockDeviceMappings, device.ToLaunchTemplateBootDeviceRequest(name))
		}
	}

	// @step: add the ssh key
	if t.SSHKey != nil {
		lc.KeyName = t.SSHKey.Name
	}
	var securityGroups []*string
	// @step: add the security groups
	for _, sg := range t.SecurityGroups {
		securityGroups = append(securityGroups, sg.ID)
	}
	// @step: add any tenacy details
	if t.Tenancy != nil {
		lc.Placement = &ec2.LaunchTemplatePlacementRequest{Tenancy: t.Tenancy}
	}
	// @step: set the instance monitoring
	lc.Monitoring = &ec2.LaunchTemplatesMonitoringRequest{Enabled: fi.Bool(false)}
	if t.InstanceMonitoring != nil {
		lc.Monitoring = &ec2.LaunchTemplatesMonitoringRequest{Enabled: t.InstanceMonitoring}
	}
	// @step: add the iam instance profile
	if t.IAMInstanceProfile != nil {
		lc.IamInstanceProfile = &ec2.LaunchTemplateIamInstanceProfileSpecificationRequest{
			Name: t.IAMInstanceProfile.Name,
		}
	}
	// @step: are the node publicly facing
	if fi.BoolValue(t.AssociatePublicIP) {
		lc.NetworkInterfaces = append(lc.NetworkInterfaces,
			&ec2.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest{
				AssociatePublicIpAddress: t.AssociatePublicIP,
				DeleteOnTermination:      aws.Bool(true),
				DeviceIndex:              fi.Int64(0),
				Groups:                   securityGroups,
			})
	} else {
		lc.SecurityGroupIds = securityGroups
	}
	// @step: add the userdata
	if t.UserData != nil {
		d, err := t.UserData.AsBytes()
		if err != nil {
			return fmt.Errorf("error rendering LaunchTemplate UserData: %v", err)
		}
		lc.UserData = aws.String(base64.StdEncoding.EncodeToString(d))
	}

	// @step: attempt to create the launch template
	err = func() error {
		for attempt := 0; attempt < 10; attempt++ {
			if _, err = c.Cloud.EC2().CreateLaunchTemplate(input); err == nil {
				return nil
			}

			if awsup.AWSErrorCode(err) == "ValidationError" {
				message := awsup.AWSErrorMessage(err)
				if strings.Contains(message, "not authorized") || strings.Contains(message, "Invalid IamInstance") {
					if attempt > 10 {
						return fmt.Errorf("IAM instance profile not yet created/propagated (original error: %v)", message)
					}
					klog.V(4).Infof("got an error indicating that the IAM instance profile %q is not ready: %q", fi.StringValue(ep.IAMInstanceProfile.Name), message)

					time.Sleep(5 * time.Second)
					continue
				}
				klog.V(4).Infof("ErrorCode=%q, Message=%q", awsup.AWSErrorCode(err), awsup.AWSErrorMessage(err))
			}
		}

		return err
	}()
	if err != nil {
		return fmt.Errorf("failed to create aws launch template: %s", err)
	}

	ep.ID = fi.String(name)

	return nil
}

// Find is responsible for finding the launch template for us
func (t *LaunchTemplate) Find(c *fi.Context) (*LaunchTemplate, error) {
	cloud, ok := c.Cloud.(awsup.AWSCloud)
	if !ok {
		return nil, fmt.Errorf("invalid cloud provider: %v, expected: %s", c.Cloud, "awsup.AWSCloud")
	}

	// @step: get the latest launch template version
	lt, err := t.findLatestLaunchTemplate(c)
	if err != nil {
		return nil, err
	}
	if lt == nil {
		return nil, nil
	}

	klog.V(3).Infof("found existing LaunchTemplate: %s", fi.StringValue(lt.LaunchTemplateName))

	actual := &LaunchTemplate{
		AssociatePublicIP:  fi.Bool(false),
		ID:                 lt.LaunchTemplateName,
		ImageID:            lt.LaunchTemplateData.ImageId,
		InstanceMonitoring: fi.Bool(false),
		InstanceType:       lt.LaunchTemplateData.InstanceType,
		Lifecycle:          t.Lifecycle,
		Name:               t.Name,
		RootVolumeOptimization: lt.LaunchTemplateData.EbsOptimized,
	}

	// @step: check if any of the interfaces are public facing
	for _, x := range lt.LaunchTemplateData.NetworkInterfaces {
		if aws.BoolValue(x.AssociatePublicIpAddress) {
			actual.AssociatePublicIP = fi.Bool(true)
			// @note: not sure i like this https://github.com/hashicorp/terraform/issues/2998
			for _, id := range x.Groups {
				actual.SecurityGroups = append(actual.SecurityGroups, &SecurityGroup{ID: id})
			}
		}
	}
	// @step: add at the security groups
	for _, id := range lt.LaunchTemplateData.SecurityGroupIds {
		actual.SecurityGroups = append(actual.SecurityGroups, &SecurityGroup{ID: id})
	}
	sort.Sort(OrderSecurityGroupsById(actual.SecurityGroups))

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

	// @step: get the image is order to find out the root device name as using the index
	// is not vaiable, under conditions they move
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
		actual.UserData = fi.WrapResource(fi.NewStringResource(string(ud)))
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
	var list []*ec2.LaunchTemplate

	cloud := c.Cloud.(awsup.AWSCloud)

	var next *string
	for {
		resp, err := cloud.EC2().DescribeLaunchTemplates(&ec2.DescribeLaunchTemplatesInput{
			NextToken: next,
		})
		if err != nil {
			return nil, err
		}
		list = append(list, resp.LaunchTemplates...)

		if resp.NextToken == nil {
			return list, nil
		}
		next = resp.NextToken
	}
}

// findAllLaunchTemplateVersions returns all the launch templates versions for us
func (t *LaunchTemplate) findAllLaunchTemplatesVersions(c *fi.Context) ([]*ec2.LaunchTemplateVersion, error) {
	var list []*ec2.LaunchTemplateVersion

	cloud, ok := c.Cloud.(awsup.AWSCloud)
	if !ok {
		return []*ec2.LaunchTemplateVersion{}, fmt.Errorf("invalid cloud provider: %v, expected: awsup.AWSCloud", c.Cloud)
	}

	templates, err := t.findAllLaunchTemplates(c)
	if err != nil {
		return nil, err
	}

	var next *string
	for _, x := range templates {
		err := func() error {
			for {
				resp, err := cloud.EC2().DescribeLaunchTemplateVersions(&ec2.DescribeLaunchTemplateVersionsInput{
					LaunchTemplateName: x.LaunchTemplateName,
					NextToken:          next,
				})
				if err != nil {
					return err
				}
				list = append(list, resp.LaunchTemplateVersions...)
				if resp.NextToken == nil {
					return nil
				}

				next = resp.NextToken
			}
		}()
		if err != nil {
			return nil, err
		}
	}

	return list, nil
}

// findLaunchTemplates returns a list of launch templates
func (t *LaunchTemplate) findLaunchTemplates(c *fi.Context) ([]*ec2.LaunchTemplateVersion, error) {
	// @step: get a list of the launch templates
	list, err := t.findAllLaunchTemplatesVersions(c)
	if err != nil {
		return nil, err
	}
	prefix := fmt.Sprintf("%s-", fi.StringValue(t.Name))

	// @step: filter out the templates we are interested in
	var filtered []*ec2.LaunchTemplateVersion
	for _, x := range list {
		if strings.HasPrefix(aws.StringValue(x.LaunchTemplateName), prefix) {
			filtered = append(filtered, x)
		}
	}

	// @step: we can sort the configurations in chronological order
	sort.Slice(filtered, func(i, j int) bool {
		ti := filtered[i].CreateTime
		tj := filtered[j].CreateTime
		if tj == nil {
			return true
		}
		if ti == nil {
			return false
		}
		return ti.UnixNano() < tj.UnixNano()
	})

	return filtered, nil
}

// findLatestLaunchTemplate returns the latest template
func (t *LaunchTemplate) findLatestLaunchTemplate(c *fi.Context) (*ec2.LaunchTemplateVersion, error) {
	// @step: get a list of configuration
	configurations, err := t.findLaunchTemplates(c)
	if err != nil {
		return nil, err
	}
	if len(configurations) == 0 {
		return nil, nil
	}

	return configurations[len(configurations)-1], nil
}

// deleteLaunchTemplate tracks a LaunchConfiguration that we're going to delete
// It implements fi.Deletion
type deleteLaunchTemplate struct {
	lc *ec2.LaunchTemplateVersion
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
