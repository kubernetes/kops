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
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
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

func awsValidateInstanceGroup(ig *kops.InstanceGroup, cloud awsup.AWSCloud) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, awsValidateAdditionalSecurityGroups(field.NewPath("spec", "additionalSecurityGroups"), ig.Spec.AdditionalSecurityGroups)...)

	if cloud != nil {
		allErrs = append(allErrs, awsValidateInstanceType(field.NewPath(ig.GetName(), "spec", "machineType"), ig.Spec.MachineType, cloud)...)
	}

	allErrs = append(allErrs, awsValidateSpotDurationInMinute(field.NewPath(ig.GetName(), "spec", "spotDurationInMinutes"), ig)...)

	allErrs = append(allErrs, awsValidateInstanceInterruptionBehavior(field.NewPath(ig.GetName(), "spec", "instanceInterruptionBehavior"), ig)...)

	if ig.Spec.MixedInstancesPolicy != nil {
		allErrs = append(allErrs, awsValidateMixedInstancesPolicy(field.NewPath("spec", "mixedInstancesPolicy"), ig.Spec.MixedInstancesPolicy, ig, cloud)...)
	}

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

func awsValidateInstanceType(fieldPath *field.Path, instanceType string, cloud awsup.AWSCloud) field.ErrorList {
	allErrs := field.ErrorList{}
	if instanceType != "" {
		for _, typ := range strings.Split(instanceType, ",") {
			if _, err := cloud.DescribeInstanceType(typ); err != nil {
				allErrs = append(allErrs, field.Invalid(fieldPath, typ, "machine type specified is invalid"))
			}
		}
	}

	return allErrs
}

func awsValidateSpotDurationInMinute(fieldPath *field.Path, ig *kops.InstanceGroup) field.ErrorList {
	allErrs := field.ErrorList{}
	if ig.Spec.SpotDurationInMinutes != nil {
		validSpotDurations := []string{"60", "120", "180", "240", "300", "360"}
		spotDurationStr := strconv.FormatInt(*ig.Spec.SpotDurationInMinutes, 10)
		allErrs = append(allErrs, IsValidValue(fieldPath, &spotDurationStr, validSpotDurations)...)
	}
	return allErrs
}

func awsValidateInstanceInterruptionBehavior(fieldPath *field.Path, ig *kops.InstanceGroup) field.ErrorList {
	allErrs := field.ErrorList{}
	if ig.Spec.InstanceInterruptionBehavior != nil {
		validInterruptionBehaviors := []string{"terminate", "hibernate", "stop"}
		instanceInterruptionBehavior := *ig.Spec.InstanceInterruptionBehavior
		allErrs = append(allErrs, IsValidValue(fieldPath, &instanceInterruptionBehavior, validInterruptionBehaviors)...)
	}
	return allErrs
}

// awsValidateMixedInstancesPolicy is responsible for validating the user input of a mixed instance policy
func awsValidateMixedInstancesPolicy(path *field.Path, spec *kops.MixedInstancesPolicySpec, ig *kops.InstanceGroup, cloud awsup.AWSCloud) field.ErrorList {
	var errs field.ErrorList

	// @step: check the instances are validate
	if cloud != nil {
		for i, x := range spec.Instances {
			errs = append(errs, awsValidateInstanceType(path.Child("instances").Index(i).Child("instanceType"), x, cloud)...)
		}
	}
	if spec.OnDemandBase != nil {
		if fi.Int64Value(spec.OnDemandBase) < 0 {
			errs = append(errs, field.Invalid(path.Child("onDemandBase"), spec.OnDemandBase, "cannot be less than zero"))
		}
		if fi.Int64Value(spec.OnDemandBase) > int64(fi.Int32Value(ig.Spec.MaxSize)) {
			errs = append(errs, field.Invalid(path.Child("onDemandBase"), spec.OnDemandBase, "cannot be greater than max size"))
		}
	}

	if spec.OnDemandAboveBase != nil {
		if fi.Int64Value(spec.OnDemandAboveBase) < 0 {
			errs = append(errs, field.Invalid(path.Child("onDemandAboveBase"), spec.OnDemandAboveBase, "cannot be less than 0"))
		}
		if fi.Int64Value(spec.OnDemandAboveBase) > 100 {
			errs = append(errs, field.Invalid(path.Child("onDemandAboveBase"), spec.OnDemandAboveBase, "cannot be greater than 100"))
		}
	}

	errs = append(errs, IsValidValue(path.Child("spotAllocationStrategy"), spec.SpotAllocationStrategy, kops.SpotAllocationStrategies)...)

	return errs
}
