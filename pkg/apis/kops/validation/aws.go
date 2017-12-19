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

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
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

	for i, subnet := range c.Spec.Subnets {
		f := field.NewPath("spec", "Subnets").Index(i)
		allErrs = append(allErrs, awsValidateRoutes(f.Child("Egress"), subnet.Egress)...)
	}

	return allErrs
}

func awsValidateInstanceGroup(ig *kops.InstanceGroup) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, awsValidateAdditionalSecurityGroups(field.NewPath("spec", "additionalSecurityGroups"), ig.Spec.AdditionalSecurityGroups)...)

	allErrs = append(allErrs, awsValidateMachineType(field.NewPath(ig.GetName(), "spec", "machineType"), ig.Spec.MachineType)...)

	return allErrs
}

func awsValidateAdditionalSecurityGroups(fieldPath *field.Path, groups []string) field.ErrorList {
	allErrs := field.ErrorList{}

	names := sets.NewString()
	for i, s := range groups {
		if names.Has(s) {
			allErrs = append(allErrs, field.Invalid(fieldPath.Index(i), s, "security groups with duplicate name found"))
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
		if _, err := awsup.GetMachineTypeInfo(machineType); err != nil {
			allErrs = append(allErrs, field.Invalid(fieldPath, machineType, "machine type specified is invalid"))
		}
	}

	return allErrs
}

func awsValidateRoutes(fieldPath *field.Path, routes []kops.EgressSpec) field.ErrorList {
	allErrs := field.ErrorList{}

	// Each route must be valid
	for i := range routes {
		allErrs = append(allErrs, awsValidateRoute(&routes[i], fieldPath.Index(i))...)
	}

	// cannot duplicate route CIDR
	{
		cidrs := sets.NewString()
		for i := range routes {
			cidr := routes[i].CIDR
			if cidrs.Has(cidr) {
				allErrs = append(allErrs, field.Invalid(fieldPath, routes, "routes with duplicate destination CIDR block found"))
			}
			cidrs.Insert(cidr)
		}
	}

	return allErrs
}

func awsValidateRoute(route *kops.EgressSpec, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	// CIDR is required
	if strings.TrimSpace(route.CIDR) == "" {
		allErrs = append(allErrs, field.Required(fieldPath.Child("CIDR"), "You must set destination CIDR block for a route"))
	}

	allErrs = append(allErrs, validateCIDR(route.CIDR, fieldPath.Child("CIDR"))...)

	// vpcPeeringConnection or instance is required
	if strings.TrimSpace(route.Instance) == "" && strings.TrimSpace(route.VpcPeeringConnection) == "" {
		allErrs = append(allErrs, field.Required(fieldPath, "You must set either vpcPeeringConnection or instance for a route"))
	}

	if strings.TrimSpace(route.Instance) != "" {
		if !strings.HasPrefix(route.Instance, "i-") {
			allErrs = append(allErrs, field.Invalid(fieldPath.Child("Instance"), route, "instance does not match the expected AWS format"))
		}
	}

	if strings.TrimSpace(route.VpcPeeringConnection) != "" {
		if !strings.HasPrefix(route.VpcPeeringConnection, "pcx-") {
			allErrs = append(allErrs, field.Invalid(fieldPath.Child("VpcPeeringConnection"), route, "vpcPeeringConnection does not match the expected AWS format"))
		}
	}

	return allErrs
}
