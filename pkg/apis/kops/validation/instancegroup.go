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

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/slice"

	"github.com/aws/aws-sdk-go/aws/arn"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateInstanceGroup is responsible for validating the configuration of a instancegroup
func ValidateInstanceGroup(g *kops.InstanceGroup) field.ErrorList {
	allErrs := field.ErrorList{}

	if g.ObjectMeta.Name == "" {
		allErrs = append(allErrs, field.Required(field.NewPath("Name"), ""))
	}

	switch g.Spec.Role {
	case "":
		allErrs = append(allErrs, field.Required(field.NewPath("Role"), "Role must be set"))
	case kops.InstanceGroupRoleMaster:
		if len(g.Spec.Subnets) == 0 {
			allErrs = append(allErrs, field.Required(field.NewPath("Subnets"), "master InstanceGroup must specify at least one Subnet"))
		}
	case kops.InstanceGroupRoleNode:
	case kops.InstanceGroupRoleBastion:
	default:
		allErrs = append(allErrs, field.Invalid(field.NewPath("Role"), g.Spec.Role, "Unknown role"))
	}

	if g.Spec.Tenancy != "" {
		if g.Spec.Tenancy != "default" && g.Spec.Tenancy != "dedicated" && g.Spec.Tenancy != "host" {
			allErrs = append(allErrs, field.Invalid(field.NewPath("Tenancy"), g.Spec.Tenancy, "Unknown tenancy. Must be Default, Dedicated or Host."))
		}
	}

	if g.Spec.MaxSize != nil && g.Spec.MinSize != nil {
		if *g.Spec.MaxSize < *g.Spec.MinSize {
			allErrs = append(allErrs, field.Invalid(field.NewPath("MaxSize"), *g.Spec.MaxSize, "maxSize must be greater than or equal to minSize."))
		}
	}

	if fi.Int32Value(g.Spec.RootVolumeIops) < 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("RootVolumeIops"), g.Spec.RootVolumeIops, "RootVolumeIops must be greater than 0"))
	}

	// @check all the hooks are valid in this instancegroup
	for i := range g.Spec.Hooks {
		allErrs = append(allErrs, validateHookSpec(&g.Spec.Hooks[i], field.NewPath("hooks").Index(i))...)
	}

	// @check the fileAssets for this instancegroup are valid
	for i := range g.Spec.FileAssets {
		allErrs = append(allErrs, validateFileAssetSpec(&g.Spec.FileAssets[i], field.NewPath("fileAssets").Index(i))...)
	}

	if g.Spec.MixedInstancesPolicy != nil {
		allErrs = append(allErrs, validatedMixedInstancesPolicy(field.NewPath(g.Name), g.Spec.MixedInstancesPolicy, g)...)
	}

	for _, UserDataInfo := range g.Spec.AdditionalUserData {
		allErrs = append(allErrs, validateExtraUserData(&UserDataInfo)...)
	}

	// @step: iterate and check the volume specs
	for i, x := range g.Spec.Volumes {
		devices := make(map[string]bool)
		path := field.NewPath("volumes").Index(i)

		allErrs = append(allErrs, validateVolumeSpec(path, x)...)

		// @check the device name has not been used already
		if _, found := devices[x.Device]; found {
			allErrs = append(allErrs, field.Invalid(path.Child("device"), x.Device, "duplicate device name found in volumes"))
		}

		devices[x.Device] = true
	}

	// @step: iterate and check the volume mount specs
	for i, x := range g.Spec.VolumeMounts {
		used := make(map[string]bool)
		path := field.NewPath("volumeMounts").Index(i)

		allErrs = append(allErrs, validateVolumeMountSpec(path, x)...)
		if _, found := used[x.Device]; found {
			allErrs = append(allErrs, field.Invalid(path.Child("device"), x.Device, "duplicate device reference"))
		}
		if _, found := used[x.Path]; found {
			allErrs = append(allErrs, field.Invalid(path.Child("path"), x.Path, "duplicate mount path specified"))
		}
	}

	allErrs = append(allErrs, validateInstanceProfile(g.Spec.IAM, field.NewPath("iam"))...)

	if g.Spec.RollingUpdate != nil {
		allErrs = append(allErrs, validateRollingUpdate(g.Spec.RollingUpdate, field.NewPath("rollingUpdate"))...)
	}

	return allErrs
}

// validatedMixedInstancesPolicy is responsible for validating the user input of a mixed instance policy
func validatedMixedInstancesPolicy(path *field.Path, spec *kops.MixedInstancesPolicySpec, ig *kops.InstanceGroup) field.ErrorList {
	var errs field.ErrorList

	if len(spec.Instances) < 2 {
		errs = append(errs, field.Invalid(path.Child("instances"), spec.Instances, "must be 2 or more instance types"))
	}
	// @step: check the instances are validate
	for i, x := range spec.Instances {
		errs = append(errs, awsValidateMachineType(path.Child("instances").Index(i).Child("instanceType"), x)...)
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

	if spec.SpotAllocationStrategy != nil && !slice.Contains(kops.SpotAllocationStrategies, fi.StringValue(spec.SpotAllocationStrategy)) {
		errs = append(errs, field.Invalid(path.Child("spotAllocationStrategy"), spec.SpotAllocationStrategy, "unsupported spot allocation strategy"))
	}

	return errs
}

// validateVolumeSpec is responsible for checking a volume spec is ok
func validateVolumeSpec(path *field.Path, v *kops.VolumeSpec) field.ErrorList {
	allErrs := field.ErrorList{}

	if v.Device == "" {
		allErrs = append(allErrs, field.Required(path.Child("device"), "device name required"))
	}
	if v.Size <= 0 {
		allErrs = append(allErrs, field.Invalid(path.Child("size"), v.Size, "must be greater than zero"))
	}

	return allErrs
}

// validateVolumeMountSpec is responsible for checking the volume mount is ok
func validateVolumeMountSpec(path *field.Path, spec *kops.VolumeMountSpec) field.ErrorList {
	allErrs := field.ErrorList{}

	if spec.Device == "" {
		allErrs = append(allErrs, field.Required(path.Child("device"), "device name required"))
	}
	if spec.Filesystem == "" {
		allErrs = append(allErrs, field.Required(path.Child("filesystem"), "filesystem type required"))
	}
	if spec.Path == "" {
		allErrs = append(allErrs, field.Required(path.Child("path"), "mount path required"))
	}
	if !slice.Contains(kops.SupportedFilesystems, spec.Filesystem) {
		allErrs = append(allErrs, field.Invalid(path.Child("filesystem"), spec.Filesystem,
			fmt.Sprintf("unsupported filesystem, available types: %s", strings.Join(kops.SupportedFilesystems, ","))))
	}

	return allErrs
}

// CrossValidateInstanceGroup performs validation of the instance group, including that it is consistent with the Cluster
// It calls ValidateInstanceGroup, so all that validation is included.
func CrossValidateInstanceGroup(g *kops.InstanceGroup, cluster *kops.Cluster, strict bool) field.ErrorList {
	allErrs := ValidateInstanceGroup(g)

	// Check that instance groups are defined in subnets that are defined in the cluster
	{
		clusterSubnets := make(map[string]*kops.ClusterSubnetSpec)
		for i := range cluster.Spec.Subnets {
			s := &cluster.Spec.Subnets[i]
			clusterSubnets[s.Name] = s
		}

		for i, z := range g.Spec.Subnets {
			if clusterSubnets[z] == nil {
				// TODO field.NotFound(field.NewPath("spec.subnets").Index(i), z) ?
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec.subnets").Index(i), z,
					fmt.Sprintf("InstanceGroup %q is configured in %q, but this is not configured as a Subnet in the cluster", g.ObjectMeta.Name, z)))
			}
		}
	}

	return allErrs
}

func validateExtraUserData(userData *kops.UserData) field.ErrorList {
	allErrs := field.ErrorList{}
	fieldPath := field.NewPath("AdditionalUserData")

	if userData.Name == "" {
		allErrs = append(allErrs, field.Required(fieldPath.Child("Name"), "field must be set"))
	}

	if userData.Content == "" {
		allErrs = append(allErrs, field.Required(fieldPath.Child("Content"), "field must be set"))
	}

	switch userData.Type {
	case "text/x-include-once-url":
	case "text/x-include-url":
	case "text/cloud-config-archive":
	case "text/upstart-job":
	case "text/cloud-config":
	case "text/part-handler":
	case "text/x-shellscript":
	case "text/cloud-boothook":

	default:
		allErrs = append(allErrs, field.Invalid(fieldPath.Child("Type"), userData.Type, "Invalid user-data content type"))
	}

	return allErrs
}

// validateInstanceProfile checks the String values for the AuthProfile
func validateInstanceProfile(v *kops.IAMProfileSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if v != nil && v.Profile != nil {
		instanceProfileARN := *v.Profile
		parsedARN, err := arn.Parse(instanceProfileARN)
		if err != nil || !strings.HasPrefix(parsedARN.Resource, "instance-profile") {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("Profile"), instanceProfileARN,
				"Instance Group IAM Instance Profile must be a valid aws arn such as arn:aws:iam::123456789012:instance-profile/KopsExampleRole"))
		}
	}
	return allErrs
}
