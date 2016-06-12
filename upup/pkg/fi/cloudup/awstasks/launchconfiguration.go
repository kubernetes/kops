package awstasks

import (
	"fmt"

	"encoding/base64"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/terraform"
	"strings"
)

//go:generate fitask -type=LaunchConfiguration
type LaunchConfiguration struct {
	Name *string

	UserData *fi.ResourceHolder

	ImageID             *string
	InstanceType        *string
	SSHKey              *SSHKey
	SecurityGroups      []*SecurityGroup
	AssociatePublicIP   *bool
	BlockDeviceMappings map[string]*BlockDeviceMapping
	IAMInstanceProfile  *IAMInstanceProfile

	ID *string
}

var _ fi.CompareWithID = &LaunchConfiguration{}

func (e *LaunchConfiguration) CompareWithID() *string {
	return e.ID
}

func (e *LaunchConfiguration) Find(c *fi.Context) (*LaunchConfiguration, error) {
	cloud := c.Cloud.(*awsup.AWSCloud)

	request := &autoscaling.DescribeLaunchConfigurationsInput{}

	prefix := *e.Name + "-"

	configurations := map[string]*autoscaling.LaunchConfiguration{}
	err := cloud.Autoscaling.DescribeLaunchConfigurationsPages(request, func(page *autoscaling.DescribeLaunchConfigurationsOutput, lastPage bool) bool {
		for _, l := range page.LaunchConfigurations {
			name := aws.StringValue(l.LaunchConfigurationName)
			if strings.HasPrefix(name, prefix) {
				suffix := name[len(prefix):]
				configurations[suffix] = l
			}
		}
		return true
	})

	if len(configurations) == 0 {
		return nil, nil
	}

	var newest *autoscaling.LaunchConfiguration
	var newestTime int64
	for _, lc := range configurations {
		t := lc.CreatedTime.UnixNano()
		if t > newestTime {
			newestTime = t
			newest = lc
		}
	}

	lc := newest

	glog.V(2).Infof("found existing AutoscalingLaunchConfiguration: %q", *lc.LaunchConfigurationName)

	actual := &LaunchConfiguration{
		Name:               e.Name,
		ID:                 lc.LaunchConfigurationName,
		ImageID:            lc.ImageId,
		InstanceType:       lc.InstanceType,
		SSHKey:             &SSHKey{Name: lc.KeyName},
		AssociatePublicIP:  lc.AssociatePublicIpAddress,
		IAMInstanceProfile: &IAMInstanceProfile{Name: lc.IamInstanceProfile},
	}

	securityGroups := []*SecurityGroup{}
	for _, sgID := range lc.SecurityGroups {
		securityGroups = append(securityGroups, &SecurityGroup{ID: sgID})
	}
	actual.SecurityGroups = securityGroups

	actual.BlockDeviceMappings = make(map[string]*BlockDeviceMapping)
	for _, b := range lc.BlockDeviceMappings {
		deviceName, bdm := BlockDeviceMappingFromAutoscaling(b)
		actual.BlockDeviceMappings[deviceName] = bdm
	}
	userData, err := base64.StdEncoding.DecodeString(*lc.UserData)
	if err != nil {
		return nil, fmt.Errorf("error decoding UserData: %v", err)
	}
	actual.UserData = fi.WrapResource(fi.NewStringResource(string(userData)))

	// Avoid spurious changes on ImageId
	if e.ImageID != nil && actual.ImageID != nil && *actual.ImageID != *e.ImageID {
		image, err := cloud.ResolveImage(*e.ImageID)
		if err != nil {
			glog.Warningf("unable to resolve image: %q: %v", *e.ImageID, err)
		} else if image == nil {
			glog.Warningf("unable to resolve image: %q: not found", *e.ImageID)
		} else if aws.StringValue(image.ImageId) == *actual.ImageID {
			glog.V(4).Infof("Returning matching ImageId as expected name: %q -> %q", *actual.ImageID, *e.ImageID)
			actual.ImageID = e.ImageID
		}
	}

	if e.ID == nil {
		e.ID = actual.ID
	}

	return actual, nil
}

func addEphemeralDevices(instanceTypeName *string, blockDeviceMappings map[string]*BlockDeviceMapping) (map[string]*BlockDeviceMapping, error) {
	// TODO: Any reason not to always attach the ephemeral devices?
	if instanceTypeName == nil {
		return nil, fi.RequiredField("InstanceType")
	}
	instanceType, err := awsup.GetMachineTypeInfo(*instanceTypeName)
	if err != nil {
		return nil, err
	}
	if blockDeviceMappings == nil {
		blockDeviceMappings = make(map[string]*BlockDeviceMapping)
	}
	for _, ed := range instanceType.EphemeralDevices() {
		if _, found := blockDeviceMappings[ed.DeviceName]; found {
			glog.Warningf("not attach ephemeral device - found duplicate device mapping: %q", ed.DeviceName)
			continue
		}
		blockDeviceMappings[ed.DeviceName] = &BlockDeviceMapping{VirtualName: fi.String(ed.VirtualName)}
	}
	return blockDeviceMappings, nil
}

func (e *LaunchConfiguration) Run(c *fi.Context) error {
	blockDeviceMappings, err := addEphemeralDevices(e.InstanceType, e.BlockDeviceMappings)
	if err != nil {
		return err
	}
	e.BlockDeviceMappings = blockDeviceMappings

	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *LaunchConfiguration) CheckChanges(a, e, changes *LaunchConfiguration) error {
	if a != nil {
		if e.InstanceType == nil {
			return fi.RequiredField("InstanceType")
		}
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	}
	return nil
}

func (_ *LaunchConfiguration) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *LaunchConfiguration) error {
	launchConfigurationName := *e.Name + "-" + fi.BuildTimestampString()
	glog.V(2).Infof("Creating AutoscalingLaunchConfiguration with Name:%q", launchConfigurationName)

	if e.ImageID == nil {
		return fi.RequiredField("ImageID")
	}
	image, err := t.Cloud.ResolveImage(*e.ImageID)
	if err != nil {
		return err
	}

	request := &autoscaling.CreateLaunchConfigurationInput{}
	request.LaunchConfigurationName = &launchConfigurationName
	request.ImageId = image.ImageId
	request.InstanceType = e.InstanceType
	if e.SSHKey != nil {
		request.KeyName = e.SSHKey.Name
	}
	securityGroupIDs := []*string{}
	for _, sg := range e.SecurityGroups {
		securityGroupIDs = append(securityGroupIDs, sg.ID)
	}
	request.SecurityGroups = securityGroupIDs
	request.AssociatePublicIpAddress = e.AssociatePublicIP
	if e.BlockDeviceMappings != nil {
		request.BlockDeviceMappings = []*autoscaling.BlockDeviceMapping{}
		for device, bdm := range e.BlockDeviceMappings {
			request.BlockDeviceMappings = append(request.BlockDeviceMappings, bdm.ToAutoscaling(device))
		}
	}

	if e.UserData != nil {
		d, err := e.UserData.AsBytes()
		if err != nil {
			return fmt.Errorf("error rendering AutoScalingLaunchConfiguration UserData: %v", err)
		}
		request.UserData = aws.String(base64.StdEncoding.EncodeToString(d))
	}
	if e.IAMInstanceProfile != nil {
		request.IamInstanceProfile = e.IAMInstanceProfile.Name
	}

	_, err = t.Cloud.Autoscaling.CreateLaunchConfiguration(request)
	if err != nil {
		return fmt.Errorf("error creating AutoscalingLaunchConfiguration: %v", err)
	}

	e.ID = fi.String(launchConfigurationName)

	return nil // No tags on a launch configuration
}

type terraformLaunchConfiguration struct {
	NamePrefix               *string                 `json:"name_prefix,omitempty"`
	ImageID                  *string                 `json:"image_id,omitempty"`
	InstanceType             *string                 `json:"instance_type,omitempty"`
	KeyName                  *terraform.Literal      `json:"key_name,omitempty"`
	IAMInstanceProfile       *terraform.Literal      `json:"iam_instance_profile,omitempty"`
	SecurityGroups           []*terraform.Literal    `json:"security_groups,omitempty"`
	AssociatePublicIpAddress *bool                   `json:"associate_public_ip_address,omitempty"`
	UserData                 *terraform.Literal      `json:"user_data,omitempty"`
	EphemeralBlockDevice     []*terraformBlockDevice `json:"ephemeral_block_device,omitempty"`
	Lifecycle                *terraformLifecycle     `json:"lifecycle,omitempty"`
}

type terraformBlockDevice struct {
	DeviceName  *string `json:"device_name"`
	VirtualName *string `json:"virtual_name"`
}

type terraformLifecycle struct {
	CreateBeforeDestroy *bool `json:"create_before_destroy,omitempty"`
}

func (_ *LaunchConfiguration) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *LaunchConfiguration) error {
	cloud := t.Cloud.(*awsup.AWSCloud)

	if e.ImageID == nil {
		return fi.RequiredField("ImageID")
	}
	image, err := cloud.ResolveImage(*e.ImageID)
	if err != nil {
		return err
	}

	tf := &terraformLaunchConfiguration{
		NamePrefix:   fi.String(*e.Name + "-"),
		ImageID:      image.ImageId,
		InstanceType: e.InstanceType,
	}

	if e.SSHKey != nil {
		tf.KeyName = e.SSHKey.TerraformLink()
	}

	for _, sg := range e.SecurityGroups {
		tf.SecurityGroups = append(tf.SecurityGroups, sg.TerraformLink())
	}
	tf.AssociatePublicIpAddress = e.AssociatePublicIP

	if e.BlockDeviceMappings != nil {
		tf.EphemeralBlockDevice = []*terraformBlockDevice{}
		for deviceName, bdm := range e.BlockDeviceMappings {
			tf.EphemeralBlockDevice = append(tf.EphemeralBlockDevice, &terraformBlockDevice{
				VirtualName: bdm.VirtualName,
				DeviceName:  fi.String(deviceName),
			})
		}
	}

	if e.UserData != nil {
		tf.UserData, err = t.AddFile("aws_launch_configuration", *e.Name, "user_data", e.UserData)
		if err != nil {
			return err
		}
	}
	if e.IAMInstanceProfile != nil {
		tf.IAMInstanceProfile = e.IAMInstanceProfile.TerraformLink()
	}

	// So that we can update configurations
	tf.Lifecycle = &terraformLifecycle{CreateBeforeDestroy: fi.Bool(true)}

	return t.RenderResource("aws_launch_configuration", *e.Name, tf)
}

func (e *LaunchConfiguration) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("aws_launch_configuration", *e.Name, "id")
}
