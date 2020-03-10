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
	"k8s.io/apimachinery/pkg/util/intstr"

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
			allErrs = append(allErrs, validateNetworkingCalico(spec.Networking.Calico, spec.EtcdClusters[0], fieldPath.Child("networking", "calico"))...)
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

	if spec.RollingUpdate != nil {
		allErrs = append(allErrs, validateRollingUpdate(spec.RollingUpdate, fieldPath.Child("rollingUpdate"), false)...)
	}

	return allErrs
}

func validateCIDR(cidr string, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		detail := "Could not be parsed as a CIDR"
		if !strings.Contains(cidr, "/") {
			ip := net.ParseIP(cidr)
			if ip != nil {
				detail += fmt.Sprintf(" (did you mean \"%s/32\")", cidr)
			}
		}
		allErrs = append(allErrs, field.Invalid(fieldPath, cidr, detail))
	} else if !ip.Equal(ipNet.IP) {
		maskSize, _ := ipNet.Mask.Size()
		detail := fmt.Sprintf("Network contains bits outside prefix (did you mean \"%s/%d\")", ipNet.IP, maskSize)
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
				allErrs = append(allErrs, field.Duplicate(fieldPath.Index(i).Child("name"), name))
			}
			names.Insert(name)
		}
	}

	// cannot mix subnets with specified ID and without specified id
	if len(subnets) > 0 {
		hasID := subnets[0].ProviderID != ""
		for i := range subnets {
			if (subnets[i].ProviderID != "") != hasID {
				allErrs = append(allErrs, field.Forbidden(fieldPath.Index(i).Child("id"), "cannot mix subnets with specified ID and unspecified ID"))
			}
		}
	}

	return allErrs
}

func validateSubnet(subnet *kops.ClusterSubnetSpec, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	// name is required
	if subnet.Name == "" {
		allErrs = append(allErrs, field.Required(fieldPath.Child("name"), ""))
	}

	// CIDR
	if subnet.CIDR != "" {
		allErrs = append(allErrs, validateCIDR(subnet.CIDR, fieldPath.Child("cidr"))...)
	}

	return allErrs
}

// validateFileAssetSpec is responsible for checking a FileAssetSpec is ok
func validateFileAssetSpec(v *kops.FileAssetSpec, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if v.Name == "" {
		allErrs = append(allErrs, field.Required(fieldPath.Child("name"), ""))
	}
	if v.Content == "" {
		allErrs = append(allErrs, field.Required(fieldPath.Child("content"), ""))
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
		allErrs = append(allErrs, validateExecContainerAction(v.ExecContainer, fieldPath.Child("execContainer"))...)
	}

	return allErrs
}

func validateExecContainerAction(v *kops.ExecContainerAction, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if v.Image == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("image"), "image must be specified"))
	}

	return allErrs
}

func validateKubeAPIServer(v *kops.KubeAPIServerConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	proxyClientCertIsNil := v.ProxyClientCertFile == nil
	proxyClientKeyIsNil := v.ProxyClientKeyFile == nil

	if (proxyClientCertIsNil && !proxyClientKeyIsNil) || (!proxyClientCertIsNil && proxyClientKeyIsNil) {
		allErrs = append(allErrs, field.Forbidden(fldPath, "proxyClientCertFile and proxyClientKeyFile must both be specified (or neither)"))
	}

	if v.ServiceNodePortRange != "" {
		pr := &utilnet.PortRange{}
		err := pr.Set(v.ServiceNodePortRange)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("serviceNodePortRange"), v.ServiceNodePortRange, err.Error()))
		}
	}

	if v.AuthorizationMode != nil && strings.Contains(*v.AuthorizationMode, "Webhook") {
		if v.AuthorizationWebhookConfigFile == nil {
			allErrs = append(allErrs, field.Required(fldPath.Child("authorizationWebhookConfigFile"), "Authorization mode Webhook requires authorizationWebhookConfigFile to be specified"))
		}
	}

	return allErrs
}

func validateNetworking(c *kops.ClusterSpec, v *kops.NetworkingSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	optionTaken := false

	if v.Classic != nil {
		allErrs = append(allErrs, field.Invalid(fldPath, "classic", "classic networking is not supported"))
	}

	if v.Kubenet != nil {
		optionTaken = true
	}

	if v.External != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("external"), "only one networking option permitted"))
		}
		optionTaken = true
	}

	if v.CNI != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("cni"), "only one networking option permitted"))
		}
		optionTaken = true
	}

	if v.Kopeio != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("kopeio"), "only one networking option permitted"))
		}
		optionTaken = true
	}

	if v.Weave != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("weave"), "only one networking option permitted"))
		}
		optionTaken = true
	}

	if v.Flannel != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("flannel"), "only one networking option permitted"))
		}
		optionTaken = true

		allErrs = append(allErrs, validateNetworkingFlannel(v.Flannel, fldPath.Child("flannel"))...)
	}

	if v.Calico != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("calico"), "only one networking option permitted"))
		}
		optionTaken = true
	}

	if v.Canal != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("canal"), "only one networking option permitted"))
		}
		optionTaken = true

		allErrs = append(allErrs, validateNetworkingCanal(v.Canal, fldPath.Child("canal"))...)
	}

	if v.Kuberouter != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("kuberouter"), "only one networking option permitted"))
		}
		optionTaken = true
	}

	if v.Romana != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("romana"), "only one networking option permitted"))
		}
		optionTaken = true
	}

	if v.AmazonVPC != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("amazonvpc"), "only one networking option permitted"))
		}
		optionTaken = true

		if c.CloudProvider != "aws" {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("amazonvpc"), "amazon-vpc-routed-eni networking is supported only in AWS"))
		}
	}

	if v.Cilium != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("cilium"), "only one networking option permitted"))
		}
		optionTaken = true

		allErrs = append(allErrs, validateNetworkingCilium(c, v.Cilium, fldPath.Child("cilium"))...)
	}

	if v.LyftVPC != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("lyftvpc"), "only one networking option permitted"))
		}
		optionTaken = true

		if c.CloudProvider != "aws" {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("lyftvpc"), "amazon-vpc-routed-eni networking is supported only in AWS"))
		}
	}

	if v.GCE != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("gce"), "only one networking option permitted"))
		}
		optionTaken = true

		allErrs = append(allErrs, validateNetworkingGCE(c, v.GCE, fldPath.Child("gce"))...)
	}

	return allErrs
}

func validateNetworkingFlannel(v *kops.FlannelNetworkingSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if v.Backend == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("backend"), "Flannel backend must be specified"))
	} else {
		allErrs = append(allErrs, IsValidValue(fldPath.Child("backend"), &v.Backend, []string{"udp", "vxlan"})...)
	}

	return allErrs
}

func validateNetworkingCanal(v *kops.CanalNetworkingSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if v.DefaultEndpointToHostAction != "" {
		valid := []string{"ACCEPT", "DROP", "RETURN"}
		allErrs = append(allErrs, IsValidValue(fldPath.Child("defaultEndpointToHostAction"), &v.DefaultEndpointToHostAction, valid)...)
	}

	if v.ChainInsertMode != "" {
		valid := []string{"insert", "append"}
		allErrs = append(allErrs, IsValidValue(fldPath.Child("chainInsertMode"), &v.ChainInsertMode, valid)...)
	}

	if v.LogSeveritySys != "" {
		valid := []string{"INFO", "DEBUG", "WARNING", "ERROR", "CRITICAL", "NONE"}
		allErrs = append(allErrs, IsValidValue(fldPath.Child("logSeveritySys"), &v.LogSeveritySys, valid)...)
	}

	if v.IptablesBackend != "" {
		valid := []string{"Auto", "Legacy", "NFT"}
		allErrs = append(allErrs, IsValidValue(fldPath.Child("iptablesBackend"), &v.IptablesBackend, valid)...)
	}

	return allErrs
}

func validateNetworkingCilium(c *kops.ClusterSpec, v *kops.CiliumNetworkingSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if v.EnableNodePort && c.KubeProxy != nil && (c.KubeProxy.Enabled == nil || *c.KubeProxy.Enabled) {
		allErrs = append(allErrs, field.Forbidden(fldPath.Root().Child("spec", "kubeProxy", "enabled"), "When Cilium NodePort is enabled, kubeProxy must be disabled"))
	}

	if v.EnablePolicy != "" {
		allErrs = append(allErrs, IsValidValue(fldPath.Child("enablePolicy"), &v.EnablePolicy, []string{"default", "always", "never"})...)
	}

	if v.Tunnel != "" {
		allErrs = append(allErrs, IsValidValue(fldPath.Child("tunnel"), &v.Tunnel, []string{"vxlan", "geneve", "disabled"})...)
	}

	if v.MonitorAggregation != "" {
		allErrs = append(allErrs, IsValidValue(fldPath.Child("monitorAggregation"), &v.MonitorAggregation, []string{"low", "medium", "maximum"})...)
	}

	if v.ContainerRuntimeLabels != "" {
		allErrs = append(allErrs, IsValidValue(fldPath.Child("containerRuntimeLabels"), &v.ContainerRuntimeLabels, []string{"none", "containerd", "crio", "docker", "auto"})...)
	}

	if v.Ipam != "" {
		// "azure" not supported by kops
		allErrs = append(allErrs, IsValidValue(fldPath.Child("ipam"), &v.Ipam, []string{"crd", "eni"})...)

		if v.Ipam == kops.CiliumIpamEni {
			if c.CloudProvider != string(kops.CloudProviderAWS) {
				allErrs = append(allErrs, field.Forbidden(fldPath.Child("ipam"), "Cilum ENI IPAM is supported only in AWS"))
			}
			if !v.DisableMasquerade {
				allErrs = append(allErrs, field.Forbidden(fldPath.Child("disableMasquerade"), "Masquerade must be disabled when ENI IPAM is used"))
			}
		}
	}

	return allErrs
}

func validateNetworkingGCE(c *kops.ClusterSpec, v *kops.GCENetworkingSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if c.CloudProvider != "gce" {
		allErrs = append(allErrs, field.Forbidden(fldPath, "gce networking is supported only when on GCP"))
	}

	return allErrs
}

func validateAdditionalPolicy(role string, policy string, fldPath *field.Path) field.ErrorList {
	errs := field.ErrorList{}

	var valid []string
	for _, r := range kops.AllInstanceGroupRoles {
		valid = append(valid, strings.ToLower(string(r)))
	}
	errs = append(errs, IsValidValue(fldPath, &role, valid)...)

	statements, err := iam.ParseStatements(policy)
	if err != nil {
		errs = append(errs, field.Invalid(fldPath.Key(role), policy, "policy was not valid JSON: "+err.Error()))
	}

	// Trivial validation of policy, mostly to make sure it isn't some other random object
	for i, statement := range statements {
		fldEffect := fldPath.Key(role).Index(i).Child("Effect")
		if statement.Effect == "" {
			errs = append(errs, field.Required(fldEffect, "Effect must be specified for IAM policy"))
		} else {
			value := string(statement.Effect)
			errs = append(errs, IsValidValue(fldEffect, &value, []string{"Allow", "Deny"})...)
		}
	}

	return errs
}

func validateEtcdClusterSpec(spec *kops.EtcdClusterSpec, fieldPath *field.Path) field.ErrorList {
	errs := field.ErrorList{}

	if spec.Provider != "" {
		value := string(spec.Provider)
		errs = append(errs, IsValidValue(fieldPath.Child("provider"), &value, kops.SupportedEtcdProviderTypes)...)
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
		allErrs = append(allErrs, field.InternalError(fldPath.Child("majorVersion"), fmt.Errorf("failed to parse Etcd version to check compatibility: %s", err)))
	}

	if sem.Major != 3 {
		if e.Version == "" {
			allErrs = append(allErrs,
				field.Forbidden(fldPath.Child("majorVersion"),
					fmt.Sprintf("Unable to use v3 when ETCD version for %s cluster is default(%s)",
						e.Name, components.DefaultEtcd2Version)))
		} else {
			allErrs = append(allErrs,
				field.Forbidden(fldPath.Child("majorVersion"),
					fmt.Sprintf("Unable to use v3 when ETCD version for %s cluster is %s", e.Name, e.Version)))
		}
	}
	return allErrs
}

func validateNetworkingCalico(v *kops.CalicoNetworkingSpec, e *kops.EtcdClusterSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if v.TyphaReplicas < 0 {
		allErrs = append(allErrs,
			field.Invalid(fldPath.Child("typhaReplicas"), v.TyphaReplicas,
				fmt.Sprintf("Unable to set number of Typha replicas to less than 0, you've specified %d", v.TyphaReplicas)))
	}

	if v.MajorVersion != "" {
		valid := []string{"v3"}
		allErrs = append(allErrs, IsValidValue(fldPath.Child("majorVersion"), &v.MajorVersion, valid)...)
		if v.MajorVersion == "v3" {
			allErrs = append(allErrs, ValidateEtcdVersionForCalicoV3(e, v.MajorVersion, fldPath)...)
		}
	}

	if v.IptablesBackend != "" {
		valid := []string{"Auto", "Legacy", "NFT"}
		allErrs = append(allErrs, IsValidValue(fldPath.Child("iptablesBackend"), &v.IptablesBackend, valid)...)
	}

	return allErrs
}

func validateContainerRuntime(runtime *string, fldPath *field.Path) field.ErrorList {
	valid := []string{"containerd", "docker"}

	allErrs := field.ErrorList{}
	allErrs = append(allErrs, IsValidValue(fldPath, runtime, valid)...)

	return allErrs
}

func validateRollingUpdate(rollingUpdate *kops.RollingUpdate, fldpath *field.Path, onMasterInstanceGroup bool) field.ErrorList {
	allErrs := field.ErrorList{}
	if rollingUpdate.MaxUnavailable != nil {
		unavailable, err := intstr.GetValueFromIntOrPercent(rollingUpdate.MaxUnavailable, 1, false)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fldpath.Child("maxUnavailable"), rollingUpdate.MaxUnavailable,
				fmt.Sprintf("Unable to parse: %v", err)))
		}
		if unavailable < 0 {
			allErrs = append(allErrs, field.Invalid(fldpath.Child("maxUnavailable"), rollingUpdate.MaxUnavailable, "Cannot be negative"))
		}
	}
	if rollingUpdate.MaxSurge != nil {
		surge, err := intstr.GetValueFromIntOrPercent(rollingUpdate.MaxSurge, 1000, true)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fldpath.Child("maxSurge"), rollingUpdate.MaxSurge,
				fmt.Sprintf("Unable to parse: %v", err)))
		}
		if onMasterInstanceGroup && surge != 0 {
			allErrs = append(allErrs, field.Forbidden(fldpath.Child("maxSurge"), "Cannot surge instance groups with role \"Master\""))
		} else if surge < 0 {
			allErrs = append(allErrs, field.Invalid(fldpath.Child("maxSurge"), rollingUpdate.MaxSurge, "Cannot be negative"))
		}
	}

	return allErrs
}
