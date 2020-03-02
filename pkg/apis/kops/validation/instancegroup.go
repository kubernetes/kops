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
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

// ValidateInstanceGroup is responsible for validating the configuration of a instancegroup
func ValidateInstanceGroup(g *kops.InstanceGroup) field.ErrorList {
	allErrs := field.ErrorList{}

	if g.ObjectMeta.Name == "" {
		allErrs = append(allErrs, field.Required(field.NewPath("objectMeta", "name"), ""))
	}

	switch g.Spec.Role {
	case "":
		allErrs = append(allErrs, field.Required(field.NewPath("spec", "role"), "Role must be set"))
	case kops.InstanceGroupRoleMaster:
		if len(g.Spec.Subnets) == 0 {
			allErrs = append(allErrs, field.Required(field.NewPath("spec", "subnets"), "master InstanceGroup must specify at least one Subnet"))
		}
	case kops.InstanceGroupRoleNode:
	case kops.InstanceGroupRoleBastion:
	default:
		var supported []string
		for _, role := range kops.AllInstanceGroupRoles {
			supported = append(supported, string(role))
		}
		allErrs = append(allErrs, field.NotSupported(field.NewPath("spec", "role"), g.Spec.Role, supported))
	}

	if g.Spec.Tenancy != "" {
		allErrs = append(allErrs, IsValidValue(field.NewPath("spec", "tenancy"), &g.Spec.Tenancy, []string{"default", "dedicated", "host"})...)
	}

	if g.Spec.MaxSize != nil && g.Spec.MinSize != nil {
		if *g.Spec.MaxSize < *g.Spec.MinSize {
			allErrs = append(allErrs, field.Forbidden(field.NewPath("spec", "maxSize"), "maxSize must be greater than or equal to minSize."))
		}
	}

	if fi.Int32Value(g.Spec.RootVolumeIops) < 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "rootVolumeIops"), g.Spec.RootVolumeIops, "RootVolumeIops must be greater than 0"))
	}

	// @check all the hooks are valid in this instancegroup
	for i := range g.Spec.Hooks {
		allErrs = append(allErrs, validateHookSpec(&g.Spec.Hooks[i], field.NewPath("spec", "hooks").Index(i))...)
	}

	// @check the fileAssets for this instancegroup are valid
	for i := range g.Spec.FileAssets {
		allErrs = append(allErrs, validateFileAssetSpec(&g.Spec.FileAssets[i], field.NewPath("spec", "fileAssets").Index(i))...)
	}

	if g.Spec.MixedInstancesPolicy != nil {
		allErrs = append(allErrs, validatedMixedInstancesPolicy(field.NewPath("spec", "mixedInstancesPolicy"), g.Spec.MixedInstancesPolicy, g)...)
	}

	for _, UserDataInfo := range g.Spec.AdditionalUserData {
		allErrs = append(allErrs, validateExtraUserData(&UserDataInfo)...)
	}

	// @step: iterate and check the volume specs
	for i, x := range g.Spec.Volumes {
		devices := make(map[string]bool)
		path := field.NewPath("spec", "volumes").Index(i)

		allErrs = append(allErrs, validateVolumeSpec(path, x)...)

		// @check the device name has not been used already
		if _, found := devices[x.Device]; found {
			allErrs = append(allErrs, field.Duplicate(path.Child("device"), x.Device))
		}

		devices[x.Device] = true
	}

	// @step: iterate and check the volume mount specs
	for i, x := range g.Spec.VolumeMounts {
		used := make(map[string]bool)
		path := field.NewPath("spec", "volumeMounts").Index(i)

		allErrs = append(allErrs, validateVolumeMountSpec(path, x)...)
		if _, found := used[x.Device]; found {
			allErrs = append(allErrs, field.Duplicate(path.Child("device"), x.Device))
		}
		if _, found := used[x.Path]; found {
			allErrs = append(allErrs, field.Duplicate(path.Child("path"), x.Path))
		}
	}

	allErrs = append(allErrs, validateInstanceProfile(g.Spec.IAM, field.NewPath("spec", "iam"))...)

	if g.Spec.RollingUpdate != nil {
		allErrs = append(allErrs, validateRollingUpdate(g.Spec.RollingUpdate, field.NewPath("spec", "rollingUpdate"))...)
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

	errs = append(errs, IsValidValue(path.Child("spotAllocationStrategy"), spec.SpotAllocationStrategy, kops.SpotAllocationStrategies)...)

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
	allErrs = append(allErrs, IsValidValue(path.Child("filesystem"), &spec.Filesystem, kops.SupportedFilesystems)...)

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
				allErrs = append(allErrs, field.NotFound(field.NewPath("spec", "subnets").Index(i), z))
			}
		}
	}

	return allErrs
}

var validUserDataTypes = []string{
	"text/x-include-once-url",
	"text/x-include-url",
	"text/cloud-config-archive",
	"text/upstart-job",
	"text/cloud-config",
	"text/part-handler",
	"text/x-shellscript",
	"text/cloud-boothook",
}

func validateExtraUserData(userData *kops.UserData) field.ErrorList {
	allErrs := field.ErrorList{}
	fieldPath := field.NewPath("additionalUserData")

	if userData.Name == "" {
		allErrs = append(allErrs, field.Required(fieldPath.Child("name"), "field must be set"))
	}

	if userData.Content == "" {
		allErrs = append(allErrs, field.Required(fieldPath.Child("content"), "field must be set"))
	}

	allErrs = append(allErrs, IsValidValue(fieldPath.Child("type"), &userData.Type, validUserDataTypes)...)

	return allErrs
}

// validateInstanceProfile checks the String values for the AuthProfile
func validateInstanceProfile(v *kops.IAMProfileSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if v != nil && v.Profile != nil {
		instanceProfileARN := *v.Profile
		parsedARN, err := arn.Parse(instanceProfileARN)
		if err != nil || !strings.HasPrefix(parsedARN.Resource, "instance-profile") {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("profile"), instanceProfileARN,
				"Instance Group IAM Instance Profile must be a valid aws arn such as arn:aws:iam::123456789012:instance-profile/KopsExampleRole"))
		}
	}
	return allErrs
}
