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

package alitasks

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ess"
	"k8s.io/klog"

	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/aliup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=LaunchConfiguration

const dateFormat = "2006-01-02T15:04Z"

// defaultRetainLaunchConfigurationCount is the number of launch configurations (matching the name prefix) that we should
// keep, we delete older ones
var defaultRetainLaunchConfigurationCount = 3

// RetainLaunchConfigurationCount returns the number of launch configurations to keep
func RetainLaunchConfigurationCount() int {
	if featureflag.KeepLaunchConfigurations.Enabled() {
		return math.MaxInt32
	}
	return defaultRetainLaunchConfigurationCount
}

// LaunchConfiguration is the specification for a launch configuration
type LaunchConfiguration struct {
	Lifecycle *fi.Lifecycle
	ID        *string
	Name      *string

	ImageID            *string
	InstanceType       *string
	SystemDiskSize     *int
	SystemDiskCategory *string

	RAMRole       *RAMRole
	ScalingGroup  *ScalingGroup
	SSHKey        *SSHKey
	UserData      *fi.ResourceHolder
	SecurityGroup *SecurityGroup

	Tags map[string]string
}

var _ fi.CompareWithID = &LaunchConfiguration{}

func (l *LaunchConfiguration) CompareWithID() *string {
	return l.ID
}

func (l *LaunchConfiguration) Find(c *fi.Context) (*LaunchConfiguration, error) {
	if l.ScalingGroup == nil || l.ScalingGroup.ScalingGroupId == nil {
		klog.V(4).Infof("ScalingGroup / ScalingGroupId not found for %s, skipping Find", fi.StringValue(l.Name))
		return nil, nil
	}

	configurations, err := l.findLaunchConfigurations(c)
	if err != nil {
		return nil, fmt.Errorf("error finding ScalingConfigurations: %v", err)
	}

	// No ScalingConfigurations with specified Name.
	if len(configurations) == 0 {
		klog.V(2).Infof("can't found matching LaunchConfiguration: %q", fi.StringValue(l.Name))
		return nil, nil
	}

	lc := configurations[len(configurations)-1]

	klog.V(2).Infof("found matching LaunchConfiguration: %q", lc.ScalingConfigurationName)

	actual := &LaunchConfiguration{
		Name:               l.Name,
		ID:                 fi.String(lc.ScalingConfigurationId),
		ImageID:            fi.String(lc.ImageId),
		InstanceType:       fi.String(lc.InstanceType),
		SystemDiskSize:     fi.Int(lc.SystemDiskSize),
		SystemDiskCategory: fi.String(string(lc.SystemDiskCategory)),
	}

	if lc.KeyPairName != "" {
		actual.SSHKey = &SSHKey{
			Name: fi.String(lc.KeyPairName),
		}
	}

	if lc.RamRoleName != "" {
		actual.RAMRole = &RAMRole{
			Name: fi.String(lc.RamRoleName),
		}
	}

	if lc.UserData != "" {
		userData, err := base64.StdEncoding.DecodeString(lc.UserData)
		if err != nil {
			return nil, fmt.Errorf("error decoding UserData: %v", err)
		}
		actual.UserData = fi.WrapResource(fi.NewStringResource(string(userData)))
	}

	actual.ScalingGroup = &ScalingGroup{
		ScalingGroupId: fi.String(lc.ScalingGroupId),
	}
	actual.SecurityGroup = &SecurityGroup{
		SecurityGroupId: fi.String(lc.SecurityGroupId),
	}

	if len(lc.Tags.Tag) != 0 {
		actual.Tags = make(map[string]string)
		for _, tag := range lc.Tags.Tag {
			actual.Tags[tag.Key] = tag.Value
		}
	}

	// Ignore "system" fields
	actual.Lifecycle = l.Lifecycle
	return actual, nil
}

func (l *LaunchConfiguration) findLaunchConfigurations(c *fi.Context) ([]*ess.ScalingConfigurationItemType, error) {
	cloud := c.Cloud.(aliup.ALICloud)
	prefix := *l.Name + "-"

	var configurations []*ess.ScalingConfigurationItemType

	pageNumber := 1
	pageSize := 50
	for {
		describeSCArgs := &ess.DescribeScalingConfigurationsArgs{
			RegionId: common.Region(cloud.Region()),
			Pagination: common.Pagination{
				PageNumber: pageNumber,
				PageSize:   pageSize,
			},
		}

		if l.ScalingGroup != nil && l.ScalingGroup.ScalingGroupId != nil {
			describeSCArgs.ScalingGroupId = fi.StringValue(l.ScalingGroup.ScalingGroupId)
		}

		configs, _, err := cloud.EssClient().DescribeScalingConfigurations(describeSCArgs)
		if err != nil {
			return nil, fmt.Errorf("error finding ScalingConfigurations: %v", err)
		}

		for _, c := range configs {
			if strings.HasPrefix(c.ScalingConfigurationName, prefix) {

				// Verify the CreationTime is parseble here, so we can ignore errors when sorting
				_, err := time.Parse(dateFormat, c.CreationTime)
				if err != nil {
					return nil, fmt.Errorf("error parse CreationTime %s: %v", c.CreationTime, err)
				}

				cc := c // Copy the pointer during iteration
				configurations = append(configurations, &cc)
			}
		}

		if len(configs) < pageSize {
			break
		} else {
			pageNumber++
		}

		klog.V(4).Infof("Describing ScalingConfigurations page %v...", pageNumber)
	}

	sort.Slice(configurations, func(i, j int) bool {
		ti, _ := time.Parse(dateFormat, configurations[i].CreationTime)
		tj, _ := time.Parse(dateFormat, configurations[j].CreationTime)
		return ti.UnixNano() < tj.UnixNano()
	})

	return configurations, nil
}

func (l *LaunchConfiguration) Run(c *fi.Context) error {
	c.Cloud.(aliup.ALICloud).AddClusterTags(l.Tags)
	return fi.DefaultDeltaRunMethod(l, c)
}

func (_ *LaunchConfiguration) CheckChanges(a, e, changes *LaunchConfiguration) error {
	//Configuration can not be modified, we need to create a new one
	if e.Name == nil {
		return fi.RequiredField("Name")
	}

	if e.ImageID == nil {
		return fi.RequiredField("ImageId")
	}

	if e.InstanceType == nil {
		return fi.RequiredField("InstanceType")
	}

	return nil
}

func (_ *LaunchConfiguration) RenderALI(t *aliup.ALIAPITarget, a, e, changes *LaunchConfiguration) error {
	launchConfigurationName := *e.Name + "-" + fi.BuildTimestampString()
	klog.V(2).Infof("Creating LaunchConfiguration with name:%q", launchConfigurationName)

	createScalingConfiguration := &ess.CreateScalingConfigurationArgs{
		ScalingConfigurationName: launchConfigurationName,
		ScalingGroupId:           fi.StringValue(e.ScalingGroup.ScalingGroupId),
		ImageId:                  fi.StringValue(e.ImageID),
		InstanceType:             fi.StringValue(e.InstanceType),
		SecurityGroupId:          fi.StringValue(e.SecurityGroup.SecurityGroupId),
		SystemDisk_Size:          common.UnderlineString(strconv.Itoa(fi.IntValue(e.SystemDiskSize))),
		SystemDisk_Category:      common.UnderlineString(fi.StringValue(e.SystemDiskCategory)),
	}

	if e.RAMRole != nil && e.RAMRole.Name != nil {
		createScalingConfiguration.RamRoleName = fi.StringValue(e.RAMRole.Name)
	}

	if e.UserData != nil {
		userData, err := e.UserData.AsString()
		if err != nil {
			return fmt.Errorf("error rendering ScalingLaunchConfiguration UserData: %v", err)
		}
		createScalingConfiguration.UserData = userData
	}

	if e.SSHKey != nil && e.SSHKey.Name != nil {
		createScalingConfiguration.KeyPairName = fi.StringValue(e.SSHKey.Name)
	}

	if e.Tags != nil {
		tagItem, err := json.Marshal(e.Tags)
		if err != nil {
			return fmt.Errorf("error rendering LaunchConfiguration Tags: %v", err)
		}
		createScalingConfiguration.Tags = string(tagItem)
	}

	createScalingConfigurationResponse, err := t.Cloud.EssClient().CreateScalingConfiguration(createScalingConfiguration)
	if err != nil {
		return fmt.Errorf("error creating scalingConfiguration: %v", err)
	}
	e.ID = fi.String(createScalingConfigurationResponse.ScalingConfigurationId)

	// Disable ScalingGroup, used to bind scalingConfig, we should execute EnableScalingGroup in the task LaunchConfiguration
	// If the ScalingGroup is active, we can not execute EnableScalingGroup.
	if e.ScalingGroup.Active != nil && fi.BoolValue(e.ScalingGroup.Active) {
		klog.V(2).Infof("Disabling LoadBalancer with id:%q", fi.StringValue(e.ScalingGroup.ScalingGroupId))

		disableScalingGroupArgs := &ess.DisableScalingGroupArgs{
			ScalingGroupId: fi.StringValue(e.ScalingGroup.ScalingGroupId),
		}
		_, err := t.Cloud.EssClient().DisableScalingGroup(disableScalingGroupArgs)
		if err != nil {
			return fmt.Errorf("error disabling scalingGroup: %v", err)
		}
	}

	//Enable this configuration
	enableScalingGroupArgs := &ess.EnableScalingGroupArgs{
		ScalingGroupId:               fi.StringValue(e.ScalingGroup.ScalingGroupId),
		ActiveScalingConfigurationId: fi.StringValue(e.ID),
	}

	klog.V(2).Infof("Enabling new LaunchConfiguration of LoadBalancer with id:%q", fi.StringValue(e.ScalingGroup.ScalingGroupId))

	_, err = t.Cloud.EssClient().EnableScalingGroup(enableScalingGroupArgs)
	if err != nil {
		return fmt.Errorf("error enabling scalingGroup: %v", err)
	}

	return nil
}

type terraformLaunchConfiguration struct {
	ImageID            *string `json:"image_id,omitempty" cty:"image_id"`
	InstanceType       *string `json:"instance_type,omitempty" cty:"instance_type"`
	SystemDiskCategory *string `json:"system_disk_category,omitempty" cty:"system_disk_category"`
	UserData           *string `json:"user_data,omitempty" cty:"user_data"`

	RAMRole       *terraform.Literal `json:"role_name,omitempty" cty:"role_name"`
	ScalingGroup  *terraform.Literal `json:"scaling_group_id,omitempty" cty:"scaling_group_id"`
	SSHKey        *terraform.Literal `json:"key_name,omitempty" cty:"key_name"`
	SecurityGroup *terraform.Literal `json:"security_group_id,omitempty" cty:"security_group_id"`
}

func (_ *LaunchConfiguration) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *LaunchConfiguration) error {
	data, err := e.UserData.AsBytes()
	if err != nil {
		return fmt.Errorf("error rendering ScalingLaunchConfiguration UserData: %v", err)
	}

	userData := base64.StdEncoding.EncodeToString(data)

	tf := &terraformLaunchConfiguration{
		ImageID:            e.ImageID,
		InstanceType:       e.InstanceType,
		SystemDiskCategory: e.SystemDiskCategory,
		UserData:           &userData,

		RAMRole:       e.RAMRole.TerraformLink(),
		ScalingGroup:  e.ScalingGroup.TerraformLink(),
		SSHKey:        e.SSHKey.TerraformLink(),
		SecurityGroup: e.SecurityGroup.TerraformLink(),
	}

	return t.RenderResource("alicloud_ess_scaling_configuration", *e.Name, tf)
}

func (l *LaunchConfiguration) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("alicloud_ess_scaling_configuration", fi.StringValue(l.Name), "id")
}

// deleteLaunchConfiguration tracks a LaunchConfiguration that we're going to delete
// It implements fi.Deletion
type deleteLaunchConfiguration struct {
	lc *ess.ScalingConfigurationItemType
}

var _ fi.Deletion = &deleteLaunchConfiguration{}

func (d *deleteLaunchConfiguration) TaskName() string {
	return "LaunchConfiguration"
}

func (d *deleteLaunchConfiguration) Item() string {
	return d.lc.ScalingConfigurationName
}

func (d *deleteLaunchConfiguration) Delete(t fi.Target) error {
	klog.V(2).Infof("deleting launch configuration %v", d)

	aliTarget, ok := t.(*aliup.ALIAPITarget)
	if !ok {
		return fmt.Errorf("unexpected target type for deletion: %T", t)
	}

	request := &ess.DeleteScalingConfigurationArgs{
		ScalingConfigurationId: d.lc.ScalingConfigurationId,
	}

	id := request.ScalingConfigurationId
	klog.V(2).Infof("Calling ESS DeleteScalingConfiguration for %s", id)
	_, err := aliTarget.Cloud.EssClient().DeleteScalingConfiguration(request)
	if err != nil {
		return fmt.Errorf("error deleting ESS LaunchConfiguration %s: %v", id, err)
	}

	return nil
}

func (d *deleteLaunchConfiguration) String() string {
	return d.TaskName() + "-" + d.Item()
}

func (e *LaunchConfiguration) FindDeletions(c *fi.Context) ([]fi.Deletion, error) {
	var removals []fi.Deletion

	configurations, err := e.findLaunchConfigurations(c)
	if err != nil {
		return nil, err
	}

	if len(configurations) <= RetainLaunchConfigurationCount() {
		return nil, nil
	}

	configurations = configurations[:len(configurations)-RetainLaunchConfigurationCount()]

	for _, configuration := range configurations {
		removals = append(removals, &deleteLaunchConfiguration{lc: configuration})
	}

	klog.V(2).Infof("will delete launch configurations: %v", removals)

	return removals, nil
}
