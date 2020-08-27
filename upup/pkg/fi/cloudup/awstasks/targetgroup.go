/*
Copyright 2020 The Kubernetes Authors.

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

	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

// TargetGroup manages an targetgroup used for an ALB/NLB.
// +kops:fitask
type TargetGroup struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	// ARN is the Amazon Resource Name for the Target Group
	ARN *string

	// Shared is set if this is an external LB (one we don't create or own)
	Shared *bool
}

var _ fi.CompareWithID = &TargetGroup{}

func (e *TargetGroup) CompareWithID() *string {
	return e.ARN
}

func (e *TargetGroup) Find(c *fi.Context) (*TargetGroup, error) {
	if e.ARN == nil {
		return nil, fmt.Errorf("ARN must be set for TargetGroup")
	}

	actual := &TargetGroup{}
	actual.ARN = e.ARN

	// Prevent spurious changes
	actual.Name = e.Name
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *TargetGroup) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *TargetGroup) ShouldCreate(a, e, changes *TargetGroup) (bool, error) {
	if fi.BoolValue(e.Shared) {
		return false, nil
	}
	return true, nil
}

func (s *TargetGroup) CheckChanges(a, e, changes *TargetGroup) error {
	if a == nil {
		if e.ARN == nil {
			return fi.RequiredField("ARN")
		}
	}
	return nil
}

func (_ *TargetGroup) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *TargetGroup) error {
	shared := fi.BoolValue(e.Shared)
	if shared {
		return nil
	}

	return fmt.Errorf("non shared Target Groups is not yet supported")
}

type terraformTargetGroup struct {
	Name *string `json:"name" cty:"name"`
}

func (_ *TargetGroup) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *TargetGroup) error {
	shared := fi.BoolValue(e.Shared)
	if shared {
		return nil
	}

	return fmt.Errorf("non shared Target Groups is not yet supported")
}

func (e *TargetGroup) TerraformLink(params ...string) *terraform.Literal {
	shared := fi.BoolValue(e.Shared)
	if shared {
		if e.ARN == nil {
			klog.Fatalf("ARN must be set for shared Target Group: %s", e)
		}

		klog.V(4).Infof("reusing existing Target Group with ARN %q", *e.ARN)
		return terraform.LiteralFromStringValue(*e.ARN)
	}

	return nil
}

type cloudformationTargetGroup struct {
	Name *string `json:"Name,omitempty"`
}

func (_ *TargetGroup) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *TargetGroup) error {
	shared := fi.BoolValue(e.Shared)
	if shared {
		return nil
	}

	return fmt.Errorf("non shared Target Groups is not yet supported")
}

func (e *TargetGroup) CloudformationLink() *cloudformation.Literal {
	shared := fi.BoolValue(e.Shared)
	if shared {
		if e.ARN == nil {
			klog.Fatalf("ARN must be set for shared Target Group: %s", e)
		}

		klog.V(4).Infof("reusing existing Target Group with ARN %q", *e.ARN)
		return cloudformation.LiteralString(*e.ARN)
	}

	return nil
}
