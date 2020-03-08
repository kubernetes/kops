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

package awsmodel

import (
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/model"
)

// AWSModelContext provides the context for the aws model
type AWSModelContext struct {
	*model.KopsModelContext
}

// UseMixedInstancePolicies indicates if we are using mixed instance policies
func UseMixedInstancePolicies(ig *kops.InstanceGroup) bool {
	return ig.Spec.MixedInstancesPolicy != nil
}

// UseLaunchTemplate checks if we need to use a launch template rather than configuration
func UseLaunchTemplate(ig *kops.InstanceGroup) bool {
	if featureflag.EnableLaunchTemplates.Enabled() {
		return true
	}

	if ig.Spec.UseLaunchTemplate != nil && *ig.Spec.UseLaunchTemplate {
		return true
	}

	return UseMixedInstancePolicies(ig)
}
