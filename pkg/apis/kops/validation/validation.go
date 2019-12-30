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

package validation

import (
	"fmt"
	"net"
	"strings"

	"github.com/blang/semver"

	"k8s.io/apimachinery/pkg/api/validation"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/pkg/model/iam"
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

	// Additional cloud-specific validation rules
	switch kops.CloudProviderID(cluster.Spec.CloudProvider) {
	case kops.CloudProviderAWS:
		allErrs = append(allErrs, awsValidateCluster(cluster)...)
	case kops.CloudProviderGCE:
		allErrs = append(allErrs, gceValidateCluster(cluster)...)
	}

	return allErrs
}

func validateClusterSpec(spec *kops.ClusterSpec, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, validateSubnets(spec.Subnets, fieldPath.Child("subnets"))...)

	// SSHAccess
	for i, cidr := range spec.SSHAccess {
		allErrs = append(allErrs, validateCIDR(cidr, fieldPath.Child("sshAccess").Index(i))...)
	}

	// KubernetesAPIAccess
	for i, cidr := range spec.KubernetesAPIAccess {
		allErrs = append(allErrs, validateCIDR(cidr, fieldPath.Child("kubernetesAPIAccess").Index(i))...)
	}

	// NodePortAccess
	for i, cidr := range spec.NodePortAccess {
		allErrs = append(allErrs, validateCIDR(cidr, fieldPath.Child("nodePortAccess").Index(i))...)
	}

	// AdditionalNetworkCIDRs
	for i, cidr := range spec.AdditionalNetworkCIDRs {
		allErrs = append(allErrs, validateCIDR(cidr, fieldPath.Child("additionalNetworkCIDRs").Index(i))...)
	}

	// Hooks
	for i := range spec.Hooks {
		allErrs = append(allErrs, validateHookSpec(&spec.Hooks[i], fieldPath.Child("hooks").Index(i))...)
	}

	if spec.FileAssets != nil {
		for i, x := range spec.FileAssets {
			allErrs = append(allErrs, validateFileAssetSpec(&x, fieldPath.Child("fileAssets").Index(i))...)
		}
	}

	if spec.KubeAPIServer != nil {
		allErrs = append(allErrs, validateKubeAPIServer(spec.KubeAPIServer, fieldPath.Child("kubeAPIServer"))...)
	}

	if spec.Networking != nil {
		allErrs = append(allErrs, validateNetworking(spec, spec.Networking, fieldPath.Child("networking"))...)
		if spec.Networking.Calico != nil {
			allErrs = append(allErrs, validateNetworkingCalico(spec.Networking.Calico, spec.EtcdClusters[0], fieldPath.Child("networking").Child("Calico"))...)
		}
	}

	// IAM additionalPolicies
	if spec.AdditionalPolicies != nil {
		for k, v := range *spec.AdditionalPolicies {
			allErrs = append(allErrs, validateAdditionalPolicy(k, v, fieldPath.Child("additionalPolicies"))...)
		}
	}

	// EtcdClusters
	{
		for i, etcdCluster := range spec.EtcdClusters {
			allErrs = append(allErrs, validateEtcdClusterSpec(etcdCluster, fieldPath.Child("etcdClusters").Index(i))...)
		}
	}

	// Container Runtime
	if spec.ContainerRuntime != "" {
		allErrs = append(allErrs, validateContainerRuntime(&spec.ContainerRuntime, fieldPath.Child("containerRuntime"))...)
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

// validateFileAssetSpec is responsible for checking a FileAssetSpec is ok
func validateFileAssetSpec(v *kops.FileAssetSpec, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if v.Name == "" {
		allErrs = append(allErrs, field.Required(fieldPath.Child("Name"), ""))
	}
	if v.Content == "" {
		allErrs = append(allErrs, field.Required(fieldPath.Child("Content"), ""))
	}

	return allErrs
}

func validateHookSpec(v *kops.HookSpec, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	// if this unit is disabled, short-circuit and do not validate
	if v.Disabled {
		return allErrs
	}

	if v.ExecContainer == nil && v.Manifest == "" {
		allErrs = append(allErrs, field.Required(fieldPath, "you must set either manifest or execContainer for a hook"))
	}

	if v.ExecContainer != nil && v.UseRawManifest {
		allErrs = append(allErrs, field.Forbidden(fieldPath, "execContainer may not be used with useRawManifest (use manifest instead)"))
	}

	if v.Manifest == "" && v.UseRawManifest {
		allErrs = append(allErrs, field.Required(fieldPath, "you must set manifest when useRawManifest is true"))
	}

	if v.Before != nil && v.UseRawManifest {
		allErrs = append(allErrs, field.Forbidden(fieldPath, "before may not be used with useRawManifest"))
	}

	if v.Requires != nil && v.UseRawManifest {
		allErrs = append(allErrs, field.Forbidden(fieldPath, "requires may not be used with useRawManifest"))
	}

	if v.ExecContainer != nil {
		allErrs = append(allErrs, validateExecContainerAction(v.ExecContainer, fieldPath.Child("ExecContainer"))...)
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

func validateKubeAPIServer(v *kops.KubeAPIServerConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	proxyClientCertIsNil := v.ProxyClientCertFile == nil
	proxyClientKeyIsNil := v.ProxyClientKeyFile == nil

	if (proxyClientCertIsNil && !proxyClientKeyIsNil) || (!proxyClientCertIsNil && proxyClientKeyIsNil) {
		flds := [2]*string{v.ProxyClientCertFile, v.ProxyClientKeyFile}
		allErrs = append(allErrs, field.Invalid(fldPath, flds, "ProxyClientCertFile and ProxyClientKeyFile must both be specified (or not all)"))
	}

	if v.ServiceNodePortRange != "" {
		pr := &utilnet.PortRange{}
		err := pr.Set(v.ServiceNodePortRange)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath, v.ServiceNodePortRange, err.Error()))
		}
	}

	if v.AuthorizationMode != nil && strings.Contains(*v.AuthorizationMode, "Webhook") {
		if v.AuthorizationWebhookConfigFile == nil {
			flds := [2]*string{v.AuthorizationMode, v.AuthorizationWebhookConfigFile}
			allErrs = append(allErrs, field.Invalid(fldPath, flds, "Authorization mode Webhook requires AuthorizationWebhookConfigFile to be specified"))
		}
	}

	return allErrs
}

func validateNetworking(c *kops.ClusterSpec, v *kops.NetworkingSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if v.Flannel != nil {
		allErrs = append(allErrs, validateNetworkingFlannel(v.Flannel, fldPath.Child("Flannel"))...)
	}

	if v.GCE != nil {
		allErrs = append(allErrs, validateNetworkingGCE(c, v.GCE, fldPath.Child("gce"))...)
	}

	return allErrs
}

func validateNetworkingFlannel(v *kops.FlannelNetworkingSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	switch v.Backend {
	case "":
		allErrs = append(allErrs, field.Required(fldPath.Child("Backend"), "Flannel backend must be specified"))
	case "udp", "vxlan":
		// OK
	default:
		allErrs = append(allErrs, field.NotSupported(fldPath.Child("Backend"), v.Backend, []string{"udp", "vxlan"}))
	}

	return allErrs
}

func validateNetworkingGCE(c *kops.ClusterSpec, v *kops.GCENetworkingSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if c.CloudProvider != "gce" {
		allErrs = append(allErrs, field.Invalid(fldPath, "gce", "gce networking is supported only when on GCP"))
	}

	return allErrs
}

func validateAdditionalPolicy(role string, policy string, fldPath *field.Path) field.ErrorList {
	errs := field.ErrorList{}

	valid := sets.NewString()
	for _, r := range kops.AllInstanceGroupRoles {
		k := strings.ToLower(string(r))
		valid.Insert(k)
	}
	if !valid.Has(role) {
		message := fmt.Sprintf("role is not known (valid values: %s)", strings.Join(valid.List(), ","))
		errs = append(errs, field.Invalid(fldPath, role, message))
	}

	statements, err := iam.ParseStatements(policy)
	if err != nil {
		errs = append(errs, field.Invalid(fldPath.Key(role), policy, "policy was not valid JSON: "+err.Error()))
	}

	// Trivial validation of policy, mostly to make sure it isn't some other random object
	for i, statement := range statements {
		fldEffect := fldPath.Key(role).Index(i).Child("Effect")
		switch statement.Effect {
		case "Allow", "Deny":
			//valid

		case "":
			errs = append(errs, field.Required(fldEffect, "Effect must be specified for IAM policy"))

		default:
			errs = append(errs, field.Invalid(fldEffect, statement.Effect, "Effect must be 'Allow' or 'Deny'"))
		}
	}

	return errs
}

func validateEtcdClusterSpec(spec *kops.EtcdClusterSpec, fieldPath *field.Path) field.ErrorList {
	errs := field.ErrorList{}

	switch spec.Provider {
	case kops.EtcdProviderTypeManager:
		// ok
	case kops.EtcdProviderTypeLegacy:
		// ok

	case "":
		// blank means that the user accepts the recommendation

	default:
		errs = append(errs, field.Invalid(fieldPath.Child("provider"), spec.Provider, "Provider must be Manager or Legacy"))
	}

	return errs
}

func ValidateEtcdVersionForCalicoV3(e *kops.EtcdClusterSpec, majorVersion string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	version := e.Version
	if e.Version == "" {
		version = components.DefaultEtcd2Version
	}
	sem, err := semver.Parse(strings.TrimPrefix(version, "v"))
	if err != nil {
		allErrs = append(allErrs, field.InternalError(fldPath.Child("MajorVersion"), fmt.Errorf("failed to parse Etcd version to check compatibility: %s", err)))
	}

	if sem.Major != 3 {
		if e.Version == "" {
			allErrs = append(allErrs,
				field.Invalid(fldPath.Child("MajorVersion"), majorVersion,
					fmt.Sprintf("Unable to use v3 when ETCD version for %s cluster is default(%s)",
						e.Name, components.DefaultEtcd2Version)))
		} else {
			allErrs = append(allErrs,
				field.Invalid(fldPath.Child("MajorVersion"), majorVersion,
					fmt.Sprintf("Unable to use v3 when ETCD version for %s cluster is %s", e.Name, e.Version)))
		}
	}
	return allErrs
}

func validateNetworkingCalico(v *kops.CalicoNetworkingSpec, e *kops.EtcdClusterSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if v.TyphaReplicas >= 0 {

	} else {
		allErrs = append(allErrs,
			field.Invalid(fldPath.Child("TyphaReplicas"), v.TyphaReplicas,
				fmt.Sprintf("Unable to set number of Typha replicas to less than 0, you've specified %d", v.TyphaReplicas)))
	}
	switch v.MajorVersion {
	case "":
		// OK:
	case "v3":
		allErrs = append(allErrs, ValidateEtcdVersionForCalicoV3(e, v.MajorVersion, fldPath)...)
	default:
		allErrs = append(allErrs, field.NotSupported(fldPath.Child("MajorVersion"), v.MajorVersion, []string{"v3"}))
	}

	return allErrs
}

func validateContainerRuntime(runtime *string, fldPath *field.Path) field.ErrorList {
	valid := []string{"containerd", "docker"}

	allErrs := field.ErrorList{}
	allErrs = append(allErrs, IsValidValue(fldPath, runtime, valid)...)

	return allErrs
}
