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
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

func awsValidateCluster(c *kops.Cluster) field.ErrorList {
	allErrs := field.ErrorList{}

	if c.Spec.API != nil {
		if c.Spec.API.LoadBalancer != nil {
			allErrs = append(allErrs, awsValidateAdditionalSecurityGroups(field.NewPath("spec", "api", "loadBalancer", "additionalSecurityGroups"), c.Spec.API.LoadBalancer.AdditionalSecurityGroups)...)
		}
	}

	return allErrs
}

func awsValidateInstanceGroup(ig *kops.InstanceGroup) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, awsValidateAdditionalSecurityGroups(field.NewPath("spec", "additionalSecurityGroups"), ig.Spec.AdditionalSecurityGroups)...)

	allErrs = append(allErrs, awsValidateMachineType(field.NewPath(ig.GetName(), "spec", "machineType"), ig.Spec.MachineType)...)

	allErrs = append(allErrs, awsValidateAMIforNVMe(field.NewPath(ig.GetName(), "spec", "machineType"), ig)...)

	return allErrs
}

func awsValidateAdditionalSecurityGroups(fieldPath *field.Path, groups []string) field.ErrorList {
	allErrs := field.ErrorList{}

	names := sets.NewString()
	for i, s := range groups {
		if names.Has(s) {
			allErrs = append(allErrs, field.Duplicate(fieldPath.Index(i), s))
		}
		names.Insert(s)
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

func awsValidateMachineType(fieldPath *field.Path, machineType string) field.ErrorList {
	allErrs := field.ErrorList{}

	if machineType != "" {
		for _, typ := range strings.Split(machineType, ",") {
			if _, err := awsup.GetMachineTypeInfo(typ); err != nil {
				allErrs = append(allErrs, field.Invalid(fieldPath, typ, "machine type specified is invalid"))
			}
		}
	}

	return allErrs
}

// TODO: make image validation smarter? graduate from jessie to stretch? This is quick and dirty because we keep getting reports
func awsValidateAMIforNVMe(fieldPath *field.Path, ig *kops.InstanceGroup) field.ErrorList {
	// TODO: how can we put this list somewhere better?
	NVMe_INSTANCE_PREFIXES := []string{"P3", "C5", "M5", "H1", "I3"}

	allErrs := field.ErrorList{}

	for _, prefix := range NVMe_INSTANCE_PREFIXES {
		for _, machineType := range strings.Split(ig.Spec.MachineType, ",") {
			if strings.Contains(strings.ToUpper(machineType), strings.ToUpper(prefix)) {
				klog.V(2).Infof("machineType %s requires an image based on stretch to operate. Trying to check compatibility", machineType)
				if strings.Contains(ig.Spec.Image, "jessie") {
					errString := fmt.Sprintf("%s cannot use machineType %s with image based on Debian jessie.", ig.Name, machineType)
					allErrs = append(allErrs, field.Forbidden(fieldPath, errString))
					continue
				}
			}
		}
	}
	return allErrs
}
