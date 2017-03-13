/*
Copyright 2017 The Kubernetes Authors.

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

package validation

import (
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kops/pkg/apis/kops"
	"strings"
)

func awsValidateCluster(c *kops.Cluster) field.ErrorList {
	return nil
}

func awsValidateInstanceGroup(ig *kops.InstanceGroup) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, awsValidateAdditionalSecurityGroups(field.NewPath("spec", "additionalSecurityGroups"), ig.Spec.AdditionalSecurityGroups)...)

	return allErrs
}

func awsValidateAdditionalSecurityGroups(fieldPath *field.Path, groups []string) field.ErrorList {
	allErrs := field.ErrorList{}

	for i, s := range groups {
		if strings.TrimSpace(s) == "" {
			allErrs = append(allErrs, field.Invalid(fieldPath.Index(i), s, "security group cannot be empty, if specified"))
			continue
		}
		if !strings.HasPrefix(s, "sg-") {
			allErrs = append(allErrs, field.Invalid(fieldPath.Index(i), s, "security group does not match the expected AWS format"))
		}
	}

	return allErrs
}
