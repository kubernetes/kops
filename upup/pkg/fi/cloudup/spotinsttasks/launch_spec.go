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

package spotinsttasks

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"

	"github.com/spotinst/spotinst-sdk-go/service/ocean/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/stringutil"
	"k8s.io/klog"
	"k8s.io/kops/pkg/resources/spotinst"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=LaunchSpec
type LaunchSpec struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	ID                 *string
	UserData           *fi.ResourceHolder
	SecurityGroups     []*awstasks.SecurityGroup
	IAMInstanceProfile *awstasks.IAMInstanceProfile
	ImageID            *string
	Tags               map[string]string
	Labels             map[string]string

	Ocean *Ocean
}

var _ fi.CompareWithID = &LaunchSpec{}

func (o *LaunchSpec) CompareWithID() *string {
	return o.Name
}

var _ fi.HasDependencies = &LaunchSpec{}

func (o *LaunchSpec) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task

	if o.IAMInstanceProfile != nil {
		deps = append(deps, o.IAMInstanceProfile)
	}

	if o.SecurityGroups != nil {
		for _, sg := range o.SecurityGroups {
			deps = append(deps, sg)
		}
	}

	if o.Ocean != nil {
		deps = append(deps, o.Ocean)
	}

	return deps
}

func (o *LaunchSpec) find(svc spotinst.LaunchSpecService, oceanID string) (*aws.LaunchSpec, error) {
	klog.V(4).Infof("Attempting to find LaunchSpec: %q", fi.StringValue(o.Name))

	specs, err := svc.List(context.Background(), oceanID)
	if err != nil {
		return nil, fmt.Errorf("spotinst: failed to find launch spec %q: %v", fi.StringValue(o.Name), err)
	}
	if len(specs) == 0 {
		return nil, fmt.Errorf("spotinst: no launch specs associated with ocean %q", oceanID)
	}

	var out *aws.LaunchSpec
	for _, spec := range specs {
		if spec.Name() == fi.StringValue(o.Name) {
			out = spec.Obj().(*aws.LaunchSpec)
			break
		}
	}
	if out == nil {
		return nil, fmt.Errorf("spotinst: failed to find launch spec %q", fi.StringValue(o.Name))
	}

	klog.V(4).Infof("LaunchSpec/%s: %s", fi.StringValue(o.Name), stringutil.Stringify(out))
	return out, nil
}

var _ fi.HasCheckExisting = &LaunchSpec{}

func (o *LaunchSpec) Find(c *fi.Context) (*LaunchSpec, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	ocean, err := o.Ocean.find(cloud.Spotinst().Ocean(), *o.Ocean.Name)
	if err != nil {
		return nil, err
	}

	spec, err := o.find(cloud.Spotinst().LaunchSpec(), *ocean.ID)
	if err != nil {
		return nil, err
	}

	actual := &LaunchSpec{}
	actual.ID = spec.ID
	actual.Name = spec.Name
	actual.Ocean = &Ocean{
		ID:   ocean.ID,
		Name: o.Ocean.Name,
	}

	// Image.
	{
		actual.ImageID = spec.ImageID

		if o.ImageID != nil && actual.ImageID != nil &&
			fi.StringValue(actual.ImageID) != fi.StringValue(o.ImageID) {
			image, err := resolveImage(cloud, fi.StringValue(o.ImageID))
			if err != nil {
				return nil, err
			}
			if fi.StringValue(image.ImageId) == fi.StringValue(spec.ImageID) {
				actual.ImageID = o.ImageID
			}
		}
	}

	// User data.
	{
		if spec.UserData != nil {
			userData, err := base64.StdEncoding.DecodeString(fi.StringValue(spec.UserData))
			if err != nil {
				return nil, err
			}
			actual.UserData = fi.WrapResource(fi.NewStringResource(string(userData)))
		}
	}

	// IAM instance profile.
	if spec.IAMInstanceProfile != nil {
		actual.IAMInstanceProfile = &awstasks.IAMInstanceProfile{Name: spec.IAMInstanceProfile.Name}
	}

	// Security groups.
	{
		if spec.SecurityGroupIDs != nil {
			for _, sgID := range spec.SecurityGroupIDs {
				actual.SecurityGroups = append(actual.SecurityGroups,
					&awstasks.SecurityGroup{ID: fi.String(sgID)})
			}
		}
	}

	// Tags.
	{
		if len(spec.Tags) > 0 {
			actual.Tags = make(map[string]string)
			for _, tag := range spec.Tags {
				actual.Tags[fi.StringValue(tag.Key)] = fi.StringValue(tag.Value)
			}
		}
	}

	// Labels.
	if spec.Labels != nil {
		actual.Labels = make(map[string]string)

		for _, label := range spec.Labels {
			actual.Labels[fi.StringValue(label.Key)] = fi.StringValue(label.Value)
		}
	}

	// Avoid spurious changes.
	actual.Lifecycle = o.Lifecycle

	return actual, nil
}

func (o *LaunchSpec) CheckExisting(c *fi.Context) bool {
	spec, err := o.Find(c)
	return err == nil && spec != nil
}

func (o *LaunchSpec) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(o, c)
}

func (s *LaunchSpec) CheckChanges(a, e, changes *LaunchSpec) error {
	if e.Name == nil {
		return fi.RequiredField("Name")
	}
	return nil
}

func (o *LaunchSpec) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *LaunchSpec) error {
	return o.createOrUpdate(t.Cloud.(awsup.AWSCloud), a, e, changes)
}

func (o *LaunchSpec) createOrUpdate(cloud awsup.AWSCloud, a, e, changes *LaunchSpec) error {
	if a == nil {
		return o.create(cloud, a, e, changes)
	} else {
		return o.update(cloud, a, e, changes)
	}
}

func (_ *LaunchSpec) create(cloud awsup.AWSCloud, a, e, changes *LaunchSpec) error {
	ocean, err := e.Ocean.find(cloud.Spotinst().Ocean(), *e.Ocean.Name)
	if err != nil {
		return err
	}

	klog.V(2).Infof("Creating Launch Spec for Ocean %q", *ocean.ID)

	spec := new(aws.LaunchSpec)
	spec.SetName(e.Name)
	spec.SetOceanId(ocean.ID)

	// Image.
	{
		if e.ImageID != nil {
			image, err := resolveImage(cloud, fi.StringValue(e.ImageID))
			if err != nil {
				return err
			}
			spec.SetImageId(image.ImageId)
		}
	}

	// User data.
	{
		if e.UserData != nil {
			userData, err := e.UserData.AsString()
			if err != nil {
				return err
			}
			encoded := base64.StdEncoding.EncodeToString([]byte(userData))
			spec.SetUserData(fi.String(encoded))
		}
	}

	// IAM instance profile.
	{
		if e.IAMInstanceProfile != nil {
			iprof := new(aws.IAMInstanceProfile)
			iprof.SetName(e.IAMInstanceProfile.GetName())
			spec.SetIAMInstanceProfile(iprof)
		}
	}

	// Security groups.
	{
		if e.SecurityGroups != nil {
			securityGroupIDs := make([]string, len(e.SecurityGroups))
			for i, sg := range e.SecurityGroups {
				securityGroupIDs[i] = *sg.ID
			}
			spec.SetSecurityGroupIDs(securityGroupIDs)
		}
	}

	// Tags.
	{
		if e.Tags != nil {
			spec.SetTags(e.buildTags())
		}
	}

	// Labels.
	{
		if e.Labels != nil && len(e.Labels) > 0 {
			var labels []*aws.Label
			for k, v := range e.Labels {
				labels = append(labels, &aws.Label{
					Key:   fi.String(k),
					Value: fi.String(v),
				})
			}
			spec.SetLabels(labels)
		}
	}

	// Wrap the raw object as an LaunchSpec.
	sp, err := spotinst.NewLaunchSpec(cloud.ProviderID(), spec)
	if err != nil {
		return err
	}

	// Create a new LaunchSpec.
	id, err := cloud.Spotinst().LaunchSpec().Create(context.Background(), sp)
	if err != nil {
		return fmt.Errorf("spotinst: failed to create launch spec: %v", err)
	}

	e.ID = fi.String(id)
	return nil
}

func (_ *LaunchSpec) update(cloud awsup.AWSCloud, a, e, changes *LaunchSpec) error {
	klog.V(2).Infof("Updating Launch Spec for Ocean %q", *a.Ocean.ID)

	actual, err := e.find(cloud.Spotinst().LaunchSpec(), *a.Ocean.ID)
	if err != nil {
		klog.Errorf("Unable to resolve Launch Spec %q, error: %v", *e.Name, err)
		return err
	}

	var changed bool
	spec := new(aws.LaunchSpec)
	spec.SetId(a.ID)

	// Image.
	{
		if changes.ImageID != nil {
			image, err := resolveImage(cloud, fi.StringValue(e.ImageID))
			if err != nil {
				return err
			}

			if *actual.ImageID != *image.ImageId {
				spec.SetImageId(image.ImageId)
			}

			changes.ImageID = nil
			changed = true
		}
	}

	// User data.
	{
		if changes.UserData != nil {
			userData, err := e.UserData.AsString()
			if err != nil {
				return err
			}
			encoded := base64.StdEncoding.EncodeToString([]byte(userData))

			spec.SetUserData(fi.String(encoded))
			changes.UserData = nil
			changed = true
		}
	}

	// IAM instance profile.
	{
		if changes.IAMInstanceProfile != nil {
			iprof := new(aws.IAMInstanceProfile)
			iprof.SetName(e.IAMInstanceProfile.GetName())

			spec.SetIAMInstanceProfile(iprof)
			changes.IAMInstanceProfile = nil
			changed = true
		}
	}

	// Security groups.
	{
		if changes.SecurityGroups != nil {
			securityGroupIDs := make([]string, len(e.SecurityGroups))
			for i, sg := range e.SecurityGroups {
				securityGroupIDs[i] = *sg.ID
			}

			spec.SetSecurityGroupIDs(securityGroupIDs)
			changes.SecurityGroups = nil
			changed = true
		}
	}

	// Tags.
	{
		if changes.Tags != nil {
			spec.SetTags(e.buildTags())
			changes.Tags = nil
			changed = true
		}
	}

	// Labels.
	{
		if changes.Labels != nil {
			labels := make([]*aws.Label, 0, len(e.Labels))
			for k, v := range e.Labels {
				labels = append(labels, &aws.Label{
					Key:   fi.String(k),
					Value: fi.String(v),
				})
			}

			spec.SetLabels(labels)
			changes.Labels = nil
			changed = true
		} else if a.Labels != nil { // Forcibly remove all existing labels.
			spec.SetLabels(nil)
			changed = true
		}
	}

	empty := &LaunchSpec{}
	if !reflect.DeepEqual(empty, changes) {
		klog.Warningf("Not all changes applied to Launch Spec %q: %v", *spec.ID, changes)
	}

	if !changed {
		klog.V(2).Infof("No changes detected in Launch Spec %q", *spec.ID)
		return nil
	}

	klog.V(2).Infof("Updating Launch Spec %q (config: %s)", *spec.ID, stringutil.Stringify(spec))

	// Wrap the raw object as an LaunchSpec.
	sp, err := spotinst.NewLaunchSpec(cloud.ProviderID(), spec)
	if err != nil {
		return err
	}

	// Update an existing LaunchSpec.
	if err := cloud.Spotinst().LaunchSpec().Update(context.Background(), sp); err != nil {
		return fmt.Errorf("spotinst: failed to update launch spec: %v", err)
	}

	return nil
}

type terraformLaunchSpec struct {
	Name    *string            `json:"name,omitempty" cty:"name"`
	OceanID *terraform.Literal `json:"ocean_id,omitempty" cty:"ocean_id"`

	*terraformOceanLaunchSpec
}

func (_ *LaunchSpec) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *LaunchSpec) error {
	cloud := t.Cloud.(awsup.AWSCloud)

	tf := &terraformLaunchSpec{
		Name:                     e.Name,
		OceanID:                  e.Ocean.TerraformLink(),
		terraformOceanLaunchSpec: &terraformOceanLaunchSpec{},
	}

	// Image.
	if e.ImageID != nil {
		image, err := resolveImage(cloud, fi.StringValue(e.ImageID))
		if err != nil {
			return err
		}
		tf.ImageID = image.ImageId
	}

	var role string
	for key := range e.Ocean.Tags {
		if strings.HasPrefix(key, awstasks.CloudTagInstanceGroupRolePrefix) {
			suffix := strings.TrimPrefix(key, awstasks.CloudTagInstanceGroupRolePrefix)
			if role != "" && role != suffix {
				return fmt.Errorf("spotinst: found multiple role tags %q vs %q", role, suffix)
			}
			role = suffix
		}
	}

	// Security groups.
	if e.SecurityGroups != nil {
		for _, sg := range e.SecurityGroups {
			tf.SecurityGroups = append(tf.SecurityGroups, sg.TerraformLink())
			if role != "" {
				if err := t.AddOutputVariableArray(role+"_security_groups", sg.TerraformLink()); err != nil {
					return err
				}
			}
		}
	}

	// User data.
	if e.UserData != nil {
		var err error
		tf.UserData, err = t.AddFile("spotinst_ocean_aws_launch_spec", *e.Name, "user_data", e.UserData)
		if err != nil {
			return err
		}
	}

	// IAM instance profile.
	if e.IAMInstanceProfile != nil {
		tf.IAMInstanceProfile = e.IAMInstanceProfile.TerraformLink()
	}

	// Tags.
	{
		if e.Tags != nil {
			for _, tag := range e.buildTags() {
				tf.Tags = append(tf.Tags, &terraformKV{
					Key:   tag.Key,
					Value: tag.Value,
				})
			}
		}
	}

	// Labels.
	{
		if e.Labels != nil {
			tf.Labels = make([]*terraformKV, 0, len(e.Labels))
			for k, v := range e.Labels {
				tf.Labels = append(tf.Labels, &terraformKV{
					Key:   fi.String(k),
					Value: fi.String(v),
				})
			}
		}
	}

	return t.RenderResource("spotinst_ocean_aws_launch_spec", *e.Name, tf)
}

func (o *LaunchSpec) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("spotinst_ocean_aws_launch_spec", *o.Name, "id")
}

func (o *LaunchSpec) buildTags() []*aws.Tag {
	tags := make([]*aws.Tag, 0, len(o.Tags))

	for key, value := range o.Tags {
		tags = append(tags, &aws.Tag{
			Key:   fi.String(key),
			Value: fi.String(value),
		})
	}

	return tags
}
