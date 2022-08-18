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

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type AutoscalingLifecycleHook struct {
	ID        *string
	Name      *string
	Lifecycle fi.Lifecycle

	// HookName is the name of the lifecycle hook.
	// It needs to be unique within the autoscaling group.
	// If not set, Name will be used.
	HookName *string

	AutoscalingGroup    *AutoscalingGroup
	DefaultResult       *string
	HeartbeatTimeout    *int64
	LifecycleTransition *string

	Enabled *bool
}

var _ fi.CompareWithID = &AutoscalingLifecycleHook{}

func (h *AutoscalingLifecycleHook) CompareWithID() *string {
	return h.Name
}

func (h *AutoscalingLifecycleHook) Find(c *fi.Context) (*AutoscalingLifecycleHook, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	request := &autoscaling.DescribeLifecycleHooksInput{
		AutoScalingGroupName: h.AutoscalingGroup.Name,
		LifecycleHookNames:   []*string{h.GetHookName()},
	}

	response, err := cloud.Autoscaling().DescribeLifecycleHooks(request)
	if err != nil {
		return nil, fmt.Errorf("error listing ASG Lifecycle Hooks: %v", err)
	}
	if response == nil || len(response.LifecycleHooks) == 0 {
		if !fi.BoolValue(h.Enabled) {
			return h, nil
		}

		return nil, nil
	}
	if len(response.LifecycleHooks) > 1 {
		return nil, fmt.Errorf("found multiple ASG Lifecycle Hooks with the same name")
	}

	hook := response.LifecycleHooks[0]
	actual := &AutoscalingLifecycleHook{
		ID:                  h.Name,
		Name:                h.Name,
		HookName:            h.HookName,
		Lifecycle:           h.Lifecycle,
		AutoscalingGroup:    h.AutoscalingGroup,
		DefaultResult:       hook.DefaultResult,
		HeartbeatTimeout:    hook.HeartbeatTimeout,
		LifecycleTransition: hook.LifecycleTransition,
		Enabled:             fi.Bool(true),
	}

	return actual, nil
}

func (h *AutoscalingLifecycleHook) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(h, c)
}

func (_ *AutoscalingLifecycleHook) CheckChanges(a, e, changes *AutoscalingLifecycleHook) error {
	if a == nil {
		if e.Name == nil {
			return field.Required(field.NewPath("Name"), "")
		}
		if e.AutoscalingGroup == nil {
			return field.Required(field.NewPath("AutoScalingGroupName"), "")
		}
	}

	return nil
}

func (*AutoscalingLifecycleHook) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *AutoscalingLifecycleHook) error {
	if changes != nil {
		if fi.BoolValue(e.Enabled) {
			request := &autoscaling.PutLifecycleHookInput{
				AutoScalingGroupName: e.AutoscalingGroup.Name,
				DefaultResult:        e.DefaultResult,
				HeartbeatTimeout:     e.HeartbeatTimeout,
				LifecycleHookName:    e.GetHookName(),
				LifecycleTransition:  e.LifecycleTransition,
			}
			_, err := t.Cloud.Autoscaling().PutLifecycleHook(request)
			if err != nil {
				return err
			}
		} else {
			request := &autoscaling.DeleteLifecycleHookInput{
				AutoScalingGroupName: e.AutoscalingGroup.Name,
				LifecycleHookName:    e.GetHookName(),
			}
			_, err := t.Cloud.Autoscaling().DeleteLifecycleHook(request)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type terraformASGLifecycleHook struct {
	Name                 *string                  `cty:"name"`
	AutoScalingGroupName *terraformWriter.Literal `cty:"autoscaling_group_name"`
	DefaultResult        *string                  `cty:"default_result"`
	HeartbeatTimeout     *int64                   `cty:"heartbeat_timeout"`
	LifecycleTransition  *string                  `cty:"lifecycle_transition"`
}

func (_ *AutoscalingLifecycleHook) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *AutoscalingLifecycleHook) error {
	if !fi.BoolValue(e.Enabled) {
		return nil
	}
	tf := &terraformASGLifecycleHook{
		Name:                 e.GetHookName(),
		AutoScalingGroupName: e.AutoscalingGroup.TerraformLink(),
		DefaultResult:        e.DefaultResult,
		HeartbeatTimeout:     e.HeartbeatTimeout,
		LifecycleTransition:  e.LifecycleTransition,
	}

	return t.RenderResource("aws_autoscaling_lifecycle_hook", *e.Name, tf)
}

type cloudformationASGLifecycleHook struct {
	LifecycleHookName    *string
	AutoScalingGroupName *cloudformation.Literal
	DefaultResult        *string
	HeartbeatTimeout     *int64
	LifecycleTransition  *string
}

func (_ *AutoscalingLifecycleHook) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *AutoscalingLifecycleHook) error {
	if !fi.BoolValue(e.Enabled) {
		return nil
	}
	tf := &cloudformationASGLifecycleHook{
		LifecycleHookName:    e.GetHookName(),
		AutoScalingGroupName: e.AutoscalingGroup.CloudformationLink(),
		DefaultResult:        e.DefaultResult,
		HeartbeatTimeout:     e.HeartbeatTimeout,
		LifecycleTransition:  e.LifecycleTransition,
	}

	return t.RenderResource("AWS::AutoScaling::LifecycleHook", *e.Name, tf)
}

func (h *AutoscalingLifecycleHook) GetHookName() *string {
	if h.HookName != nil {
		return h.HookName
	}
	return h.Name
}
