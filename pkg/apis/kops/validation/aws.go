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
		if c.Spec.Subnets[i].Egress != nil {
			allErrs = append(allErrs, awsValidateEgress(f.Child("Egress"), subnet.Egress, subnet.Type)...)
		}
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

func awsValidateEgress(fieldPath *field.Path, routes []kops.EgressSpec, t kops.SubnetType) field.ErrorList {
	allErrs := field.ErrorList{}

	ngwCount := 0

	// Each route must be valid
	for i, r := range routes {

		f := fieldPath.Index(i).Child("NatGateway")

		if t != kops.SubnetTypePrivate && r.NatGateway != "" {
			// NAT gateway routes are only for private subnet
			allErrs = append(allErrs, field.Invalid(f, r, "non-private subnet can't have NAT gateway"))
		} else if t == kops.SubnetTypePrivate && r.NatGateway != "" && ngwCount == 1 {
			// private subnet can't have multiple NAT gateway
			allErrs = append(allErrs, field.Invalid(f, routes, "private subnet can't have multiple NAT gateway"))
		} else if t == kops.SubnetTypePrivate && r.NatGateway != "" && ngwCount == 0 {
			ngwCount = 1
		}

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

	if strings.TrimSpace(route.CIDR) == "" && route.NatGateway == "" {
		// CIDR is required for NAT instance and VPC peering routes
		allErrs = append(allErrs, field.Required(fieldPath, "You must set destination CIDR block for the route"))
	} else if route.CIDR != "" && route.NatGateway != "" {
		// CIDR is not allowed for NAT gateway. Deafult: 0.0.0.0/0
		allErrs = append(allErrs, field.Invalid(fieldPath.Child("CIDR"), route, "You can't set destination CIDR block for the NAT gateway. Deafult: 0.0.0.0/0"))
	}

	if route.CIDR != "" {
		allErrs = append(allErrs, validateCIDR(route.CIDR, fieldPath.Child("CIDR"))...)
	}

	// instance or vpcPeeringConnection or natGateway is required
	if strings.TrimSpace(route.Instance) == "" && strings.TrimSpace(route.VpcPeeringConnection) == "" && strings.TrimSpace(route.NatGateway) == "" {
		allErrs = append(allErrs, field.Required(fieldPath, "You must set either instance or NAT gateway or vpcPeeringConnection for a egress route"))
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

	if strings.TrimSpace(route.NatGateway) != "" {
		if !strings.HasPrefix(route.NatGateway, "nat-") {
			allErrs = append(allErrs, field.Invalid(fieldPath.Child("NatGateway"), route, "natGateway does not match the expected AWS format"))
		}
	}

	return allErrs
}
