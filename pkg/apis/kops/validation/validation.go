/*
Copyright 2016 The Kubernetes Authors.

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
	"net"
	"strings"

	"k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kops/pkg/apis/kops"
)

var validDockerConfigStorageValues = []string{"aufs", "btrfs", "devicemapper", "overlay", "overlay2", "zfs"}

func ValidateDockerConfig(config *kops.DockerConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, IsValidValue(fldPath.Child("storage"), config.Storage, validDockerConfigStorageValues)...)
	return allErrs
}

func newValidateCluster(cluster *kops.Cluster) field.ErrorList {
	allErrs := validation.ValidateObjectMeta(&cluster.ObjectMeta, false, validation.NameIsDNSSubdomain, field.NewPath("metadata"))
	allErrs = append(allErrs, validateClusterSpec(&cluster.Spec, field.NewPath("spec"))...)
	return allErrs
}

func validateClusterSpec(spec *kops.ClusterSpec, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, validateSubnets(spec.Subnets, field.NewPath("spec"))...)

	// SSHAccess
	for i, cidr := range spec.SSHAccess {
		allErrs = append(allErrs, validateCIDR(cidr, fieldPath.Child("sshAccess").Index(i))...)
	}

	// AdminAccess
	for i, cidr := range spec.KubernetesAPIAccess {
		allErrs = append(allErrs, validateCIDR(cidr, fieldPath.Child("kubernetesAPIAccess").Index(i))...)
	}

	for i := range spec.Hooks {
		allErrs = append(allErrs, validateHook(&spec.Hooks[i], fieldPath.Child("hooks").Index(i))...)
	}

	return allErrs
}

func validateCIDR(cidr string, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		detail := "Could not be parsed as a CIDR"
		if !strings.Contains(cidr, "/") {
			ip := net.ParseIP(cidr)
			if ip != nil {
				detail += fmt.Sprintf(" (did you mean \"%s/32\")", cidr)
			}
		}
		allErrs = append(allErrs, field.Invalid(fieldPath, cidr, detail))
	}
	return allErrs
}

func validateSubnets(subnets []kops.ClusterSubnetSpec, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	// cannot be empty
	if len(subnets) == 0 {
		allErrs = append(allErrs, field.Required(fieldPath, ""))
	}

	// Each subnet must be valid
	for i := range subnets {
		allErrs = append(allErrs, validateSubnet(&subnets[i], fieldPath.Index(i))...)
	}

	// cannot duplicate subnet name
	{
		names := sets.NewString()
		for i := range subnets {
			name := subnets[i].Name
			if names.Has(name) {
				allErrs = append(allErrs, field.Invalid(fieldPath, subnets, fmt.Sprintf("subnets with duplicate name %q found", name)))
			}
			names.Insert(name)
		}
	}

	// cannot mix subnets with specified ID and without specified id
	{
		hasID := 0
		for i := range subnets {
			if subnets[i].ProviderID != "" {
				hasID++
			}
		}
		if hasID != 0 && hasID != len(subnets) {
			allErrs = append(allErrs, field.Invalid(fieldPath, subnets, "cannot mix subnets with specified ID and unspecified ID"))
		}
	}

	return allErrs
}

func validateSubnet(subnet *kops.ClusterSubnetSpec, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	// name is required
	if subnet.Name == "" {
		allErrs = append(allErrs, field.Required(fieldPath.Child("Name"), ""))
	}

	return allErrs
}

func validateHook(v *kops.HookSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if v.ExecContainer == nil {
		allErrs = append(allErrs, field.Required(fldPath, "An action is required"))
	}

	if v.ExecContainer != nil {
		allErrs = append(allErrs, validateExecContainerAction(v.ExecContainer, fldPath.Child("ExecContainer"))...)
	}
	return allErrs
}

func validateExecContainerAction(v *kops.ExecContainerAction, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if v.Image == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("Image"), "Image must be specified"))
	}

	return allErrs
}
