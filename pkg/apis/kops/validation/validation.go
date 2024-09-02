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
	"errors"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/blang/semver/v4"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/sets"
	utilvalidation "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kops/pkg/util/subnet"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/utils"
)

func newValidateCluster(cluster *kops.Cluster, strict bool) field.ErrorList {
	allErrs := validation.ValidateObjectMeta(&cluster.ObjectMeta, false, validation.NameIsDNSSubdomain, field.NewPath("metadata"))

	clusterName := cluster.ObjectMeta.Name
	if clusterName == "" {
		allErrs = append(allErrs, field.Required(field.NewPath("objectMeta", "name"), "Cluster Name is required (e.g. --name=mycluster.myzone.com)"))
	} else {
		// Must be a dns name
		errs := utilvalidation.IsDNS1123Subdomain(clusterName)
		if len(errs) != 0 {
			allErrs = append(allErrs, field.Invalid(field.NewPath("objectMeta", "name"), clusterName, fmt.Sprintf("Cluster Name must be a valid DNS name (e.g. --name=mycluster.myzone.com) errors: %s", strings.Join(errs, ", "))))
		} else if !strings.Contains(clusterName, ".") {
			// Tolerate if this is a cluster we are importing for upgrade
			if cluster.ObjectMeta.Annotations[kops.AnnotationNameManagement] != kops.AnnotationValueManagementImported {
				allErrs = append(allErrs, field.Invalid(field.NewPath("objectMeta", "name"), clusterName, "Cluster Name must be a fully-qualified DNS name (e.g. --name=mycluster.myzone.com)"))
			}
		}
	}

	allErrs = append(allErrs, validateClusterSpec(&cluster.Spec, cluster, field.NewPath("spec"), strict)...)

	// Additional cloud-specific validation rules
	switch cluster.GetCloudProvider() {
	case kops.CloudProviderAWS:
		allErrs = append(allErrs, awsValidateCluster(cluster, strict)...)
	case kops.CloudProviderGCE:
		allErrs = append(allErrs, gceValidateCluster(cluster)...)
	}

	return allErrs
}

func validateClusterSpec(spec *kops.ClusterSpec, c *kops.Cluster, fieldPath *field.Path, strict bool) field.ErrorList {
	allErrs, providerConstraints := validateCloudProvider(c, &spec.CloudProvider, fieldPath.Child("cloudProvider"))

	// SSHAccess
	for i, cidr := range spec.SSHAccess {
		if strings.HasPrefix(cidr, "pl-") {
			if c.GetCloudProvider() != kops.CloudProviderAWS {
				allErrs = append(allErrs, field.Invalid(fieldPath.Child("sshAccess").Index(i), cidr, "Prefix List ID only supported for AWS"))
			}
		} else {
			allErrs = append(allErrs, validateCIDR(fieldPath.Child("sshAccess").Index(i), cidr)...)
		}
	}

	// KubernetesAPIAccess
	for i, cidr := range spec.API.Access {
		if strings.HasPrefix(cidr, "pl-") {
			if c.GetCloudProvider() != kops.CloudProviderAWS {
				allErrs = append(allErrs, field.Invalid(fieldPath.Child("kubernetesAPIAccess").Index(i), cidr, "Prefix List ID only supported for AWS"))
			}
		} else {
			allErrs = append(allErrs, validateCIDR(fieldPath.Child("kubernetesAPIAccess").Index(i), cidr)...)
		}
	}

	// NodePortAccess
	for i, cidr := range spec.NodePortAccess {
		if strings.HasPrefix(cidr, "pl-") {
			if c.GetCloudProvider() != kops.CloudProviderAWS {
				allErrs = append(allErrs, field.Invalid(fieldPath.Child("nodePortAccess").Index(i), cidr, "Prefix List ID only supported for AWS"))
			}
		} else {
			allErrs = append(allErrs, validateCIDR(fieldPath.Child("nodePortAccess").Index(i), cidr)...)
		}
	}

	// UpdatePolicy
	allErrs = append(allErrs, IsValidValue(fieldPath.Child("updatePolicy"), spec.UpdatePolicy, []string{kops.UpdatePolicyAutomatic, kops.UpdatePolicyExternal})...)

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
		allErrs = append(allErrs, validateKubeAPIServer(spec.KubeAPIServer, c, fieldPath.Child("kubeAPIServer"), strict)...)
	}

	if spec.KubeControllerManager != nil {
		allErrs = append(allErrs, validateKubeControllerManager(spec.KubeControllerManager, c, fieldPath.Child("kubeControllerManager"), strict)...)
	}

	if spec.KubeScheduler != nil {
		allErrs = append(allErrs, validateKubeScheduler(spec.KubeScheduler, c, fieldPath.Child("kubeScheduler"), strict)...)
	}

	if spec.KubeProxy != nil {
		allErrs = append(allErrs, validateKubeProxy(spec.KubeProxy, fieldPath.Child("kubeProxy"))...)
	}

	if spec.Kubelet != nil {
		allErrs = append(allErrs, validateKubelet(spec.Kubelet, c, fieldPath.Child("kubelet"))...)
	}

	if spec.ControlPlaneKubelet != nil {
		allErrs = append(allErrs, validateKubelet(spec.ControlPlaneKubelet, c, fieldPath.Child("controlPlaneKubelet"))...)
	}

	allErrs = append(allErrs, validateNetworking(c, &spec.Networking, fieldPath.Child("networking"), strict, providerConstraints)...)

	if spec.NodeAuthorization != nil {
		allErrs = append(allErrs, field.Forbidden(fieldPath.Child("nodeAuthorization"), "NodeAuthorization must be empty. The functionality has been reimplemented and is enabled on kubernetes >= 1.19.0."))
	}

	if spec.ClusterAutoscaler != nil {
		allErrs = append(allErrs, validateClusterAutoscaler(c, spec.ClusterAutoscaler, fieldPath.Child("clusterAutoscaler"))...)
	}

	if spec.ExternalDNS != nil {
		allErrs = append(allErrs, validateExternalDNS(c, spec.ExternalDNS, fieldPath.Child("externalDNS"))...)
	}

	if spec.MetricsServer != nil {
		allErrs = append(allErrs, validateMetricsServer(c, spec.MetricsServer, fieldPath.Child("metricsServer"))...)
	}

	if spec.SnapshotController != nil {
		allErrs = append(allErrs, validateSnapshotController(c, spec.SnapshotController, fieldPath.Child("snapshotController"))...)
	}

	// IAM additional policies
	for k, v := range spec.AdditionalPolicies {
		allErrs = append(allErrs, validateAdditionalPolicy(k, v, fieldPath.Child("additionalPolicies"))...)
	}
	// IAM external policies
	for k, v := range spec.ExternalPolicies {
		allErrs = append(allErrs, validateExternalPolicies(k, v, fieldPath.Child("externalPolicies"))...)
	}

	// EtcdClusters
	{
		fieldEtcdClusters := fieldPath.Child("etcdClusters")

		if len(spec.EtcdClusters) == 0 {
			allErrs = append(allErrs, field.Required(fieldEtcdClusters, ""))
		} else {
			for i, etcdCluster := range spec.EtcdClusters {
				allErrs = append(allErrs, validateEtcdClusterSpec(etcdCluster, c, fieldEtcdClusters.Index(i))...)
			}
			allErrs = append(allErrs, validateEtcdBackupStore(spec.EtcdClusters, fieldEtcdClusters)...)
			allErrs = append(allErrs, validateEtcdStorage(spec.EtcdClusters, fieldEtcdClusters)...)
		}
	}

	if spec.ContainerRuntime != "" {
		allErrs = append(allErrs, validateContainerRuntime(c, spec.ContainerRuntime, fieldPath.Child("containerRuntime"))...)
	}

	if spec.Containerd != nil {
		allErrs = append(allErrs, validateContainerdConfig(c, spec.Containerd, fieldPath.Child("containerd"), true)...)
	}

	if spec.Docker != nil {
		allErrs = append(allErrs, field.Forbidden(fieldPath.Child("docker"), "Docker CRI support was removed in Kubernetes 1.24: https://kubernetes.io/blog/2020/12/02/dockershim-faq"))
	}

	if spec.Assets != nil {
		if spec.Assets.ContainerProxy != nil && spec.Assets.ContainerRegistry != nil {
			allErrs = append(allErrs, field.Forbidden(fieldPath.Child("assets", "containerProxy"), "containerProxy cannot be used in conjunction with containerRegistry"))
		}
	}

	for i, sysctlParameter := range spec.SysctlParameters {
		if !strings.ContainsRune(sysctlParameter, '=') {
			allErrs = append(allErrs, field.Invalid(fieldPath.Child("sysctlParameters").Index(i), sysctlParameter, "must contain a \"=\" character"))
		}
	}

	if spec.RollingUpdate != nil {
		allErrs = append(allErrs, validateRollingUpdate(spec.RollingUpdate, fieldPath.Child("rollingUpdate"), false)...)
	}

	if spec.API.LoadBalancer != nil {
		lbSpec := spec.API.LoadBalancer
		lbPath := fieldPath.Child("api", "loadBalancer")
		if c.GetCloudProvider() != kops.CloudProviderAWS {
			if lbSpec.Class != "" {
				allErrs = append(allErrs, field.Forbidden(lbPath.Child("class"), "class is only supported on AWS"))
			}
			if lbSpec.IdleTimeoutSeconds != nil {
				allErrs = append(allErrs, field.Forbidden(lbPath.Child("idleTimeoutSeconds"), "idleTimeoutSeconds is only supported on AWS"))
			}
			if lbSpec.SecurityGroupOverride != nil {
				allErrs = append(allErrs, field.Forbidden(lbPath.Child("securityGroupOverride"), "securityGroupOverride is only supported on AWS"))
			}
			if lbSpec.AdditionalSecurityGroups != nil {
				allErrs = append(allErrs, field.Forbidden(lbPath.Child("additionalSecurityGroups"), "additionalSecurityGroups is only supported on AWS"))
			}
			if lbSpec.SSLCertificate != "" {
				allErrs = append(allErrs, field.Forbidden(lbPath.Child("sslCertificate"), "sslCertificate is only supported on AWS"))
			}
			if lbSpec.SSLPolicy != nil {
				allErrs = append(allErrs, field.Forbidden(lbPath.Child("sslPolicy"), "sslPolicy is only supported on AWS"))
			}
			if lbSpec.CrossZoneLoadBalancing != nil {
				allErrs = append(allErrs, field.Forbidden(lbPath.Child("crossZoneLoadBalancing"), "crossZoneLoadBalancing is only supported on AWS"))
			}
			if lbSpec.AccessLog != nil {
				allErrs = append(allErrs, field.Forbidden(lbPath.Child("accessLog"), "accessLog is only supported on AWS"))
			}
		}

		if lbSpec.Type == kops.LoadBalancerTypeInternal {
			var hasPrivate bool
			for _, subnet := range spec.Networking.Subnets {
				if subnet.Type == kops.SubnetTypePrivate {
					hasPrivate = true
					break
				}
			}
			if !hasPrivate {
				allErrs = append(allErrs, field.Forbidden(lbPath.Child("type"), "Internal LoadBalancers must have at least one subnet of type Private"))
			}
		}
	}

	if spec.CloudConfig != nil {
		allErrs = append(allErrs, validateCloudConfiguration(spec.CloudConfig, spec, fieldPath.Child("cloudConfig"))...)
	}

	if spec.IAM != nil {
		if spec.IAM.Legacy {
			allErrs = append(allErrs, field.Forbidden(fieldPath.Child("iam", "legacy"), "legacy IAM permissions are no longer supported"))
		}

		if len(spec.IAM.ServiceAccountExternalPermissions) > 0 {
			allErrs = append(allErrs, validateSAExternalPermissions(spec.IAM.ServiceAccountExternalPermissions, fieldPath.Child("iam", "serviceAccountExternalPermissions"))...)
		}
	}

	if spec.Karpenter != nil && spec.Karpenter.Enabled {
		fldPath := fieldPath.Child("karpenter", "enabled")
		if !fi.ValueOf(spec.IAM.UseServiceAccountExternalPermissions) {
			allErrs = append(allErrs, field.Forbidden(fldPath, "Karpenter requires that service accounts use external permissions"))
		}
	}

	if spec.CertManager != nil && fi.ValueOf(spec.CertManager.Enabled) {
		allErrs = append(allErrs, validateCertManager(c, spec.CertManager, fieldPath.Child("certManager"))...)
	}

	return allErrs
}

type cloudProviderConstraints struct {
	requiresSubnets               bool
	requiresNetworkCIDR           bool
	prohibitsNetworkCIDR          bool
	prohibitsMultipleNetworkCIDRs bool
	requiresNonMasqueradeCIDR     bool
	requiresSubnetCIDR            bool
	requiresSubnetRegion          bool
}

func validateCloudProvider(c *kops.Cluster, provider *kops.CloudProviderSpec, fieldSpec *field.Path) (allErrs field.ErrorList, constraints *cloudProviderConstraints) {
	constraints = &cloudProviderConstraints{
		requiresSubnets:               true,
		requiresNetworkCIDR:           true,
		prohibitsMultipleNetworkCIDRs: true,
		requiresNonMasqueradeCIDR:     true,
		requiresSubnetCIDR:            true,
		requiresSubnetRegion:          false,
	}

	optionTaken := false
	if c.Spec.CloudProvider.AWS != nil {
		optionTaken = true
		allErrs = append(allErrs, validateAWS(c, provider.AWS, fieldSpec.Child("aws"))...)
		constraints.prohibitsMultipleNetworkCIDRs = false
	}
	if c.Spec.CloudProvider.Azure != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("azure"), "only one cloudProvider option permitted"))
		}
		optionTaken = true
		constraints.requiresSubnetRegion = true
	}
	if c.Spec.CloudProvider.DO != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("do"), "only one cloudProvider option permitted"))
		}
		optionTaken = true
		constraints.requiresSubnets = false
		constraints.requiresSubnetCIDR = false
		constraints.requiresSubnetRegion = true
		constraints.requiresNetworkCIDR = false
	}
	if c.Spec.CloudProvider.GCE != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("gce"), "only one cloudProvider option permitted"))
		}
		optionTaken = true
		constraints.requiresNetworkCIDR = false
		constraints.requiresSubnetCIDR = false
		constraints.requiresSubnetRegion = true
		constraints.prohibitsNetworkCIDR = true
		constraints.requiresNonMasqueradeCIDR = false
	}
	if c.Spec.CloudProvider.Hetzner != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("hetzner"), "only one cloudProvider option permitted"))
		}
		optionTaken = true
		constraints.requiresNetworkCIDR = false
		constraints.requiresSubnets = false
		constraints.requiresSubnetCIDR = false
	}
	if c.Spec.CloudProvider.Openstack != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("openstack"), "only one cloudProvider option permitted"))
		}
		optionTaken = true
		constraints.requiresNetworkCIDR = false
		constraints.requiresSubnetCIDR = false
		// TODO Not required on cluster creation, but used in buildInstances?
		// constraints.requiresSubnetRegion = true
	}
	if c.Spec.CloudProvider.Scaleway != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("scaleway"), "only one cloudProvider option permitted"))
		}
		optionTaken = true
		constraints.requiresNetworkCIDR = false
		constraints.requiresSubnetCIDR = false
	}
	if c.GetCloudProvider() == kops.CloudProviderMetal {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("metal"), "only one cloudProvider option permitted"))
		}
		optionTaken = true
		constraints.requiresNetworkCIDR = false
		constraints.requiresSubnetCIDR = false
	}
	if !optionTaken {
		allErrs = append(allErrs, field.Required(fieldSpec, ""))
		constraints.requiresSubnets = false
		constraints.requiresSubnetCIDR = false
		constraints.requiresNetworkCIDR = false
	}

	return allErrs, constraints
}

func validateAWS(c *kops.Cluster, aws *kops.AWSSpec, path *field.Path) (allErrs field.ErrorList) {
	if aws.NodeTerminationHandler != nil {
		allErrs = append(allErrs, validateNodeTerminationHandler(c, aws.NodeTerminationHandler, path.Child("nodeTerminationHandler"))...)
	}

	if aws.LoadBalancerController != nil {
		allErrs = append(allErrs, validateAWSLoadBalancerController(c, aws.LoadBalancerController, path.Child("awsLoadBalanceController"))...)
	}

	if aws.WarmPool != nil {
		allErrs = append(allErrs, validateWarmPool(aws.WarmPool, path.Child("warmPool"))...)
	}

	if aws.PodIdentityWebhook != nil && aws.PodIdentityWebhook.Enabled {
		allErrs = append(allErrs, validatePodIdentityWebhook(c, aws.PodIdentityWebhook, path.Child("podIdentityWebhook"))...)
	}

	return allErrs
}

func validateSAExternalPermissions(externalPermissions []kops.ServiceAccountExternalPermission, path *field.Path) (allErrs field.ErrorList) {
	if len(externalPermissions) == 0 {
		return allErrs
	}

	sas := make(map[string]string)
	for _, sa := range externalPermissions {
		key := fmt.Sprintf("%s/%s", sa.Namespace, sa.Name)
		p := path.Key(key)
		if sa.Namespace == "" {
			allErrs = append(allErrs, field.Required(p.Child("namespace"), "namespace cannot be empty"))
		}
		if sa.Name == "" {
			allErrs = append(allErrs, field.Required(p.Child("name"), "name cannot be empty"))
		}
		_, duplicate := sas[key]
		if duplicate {
			allErrs = append(allErrs, field.Duplicate(p, key))
		}
		sas[key] = ""
		aws := sa.AWS
		ap := p.Child("aws")
		if aws == nil {
			allErrs = append(allErrs, field.Required(ap, "AWS permissions must be set"))
			continue
		}

		if len(aws.PolicyARNs) == 0 && aws.InlinePolicy == "" {
			allErrs = append(allErrs, field.Required(ap, "either inlinePolicy or policyARN must be set"))
		}
		if len(aws.PolicyARNs) > 0 && aws.InlinePolicy != "" {
			allErrs = append(allErrs, field.Forbidden(ap, "cannot set both inlinePolicy and policyARN"))
		}
	}
	return allErrs
}

// validateCIDR verifies that the cidr string can be parsed as a valid net.IPNet.
// Behaviour should be consistent with parseCIDR.
func validateCIDR(fieldPath *field.Path, cidr string) field.ErrorList {
	_, errs := parseCIDR(fieldPath, cidr)
	return errs
}

// parseCIDR is like net.ParseCIDR, but returns an error message that includes the field path.
// We also try to give some more hints on common errors.
// Hint: use validateCIDR if we don't need the parsed CIDR value.
func parseCIDR(fieldPath *field.Path, cidr string) (*net.IPNet, field.ErrorList) {
	allErrs := field.ErrorList{}

	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		detail := "Could not be parsed as a CIDR"
		if !strings.Contains(cidr, "/") {
			ip := net.ParseIP(cidr)
			if ip != nil {
				if ip.To4() != nil && !strings.Contains(cidr, ":") {
					detail += fmt.Sprintf(" (did you mean \"%s/32\")", cidr)
				} else {
					detail += fmt.Sprintf(" (did you mean \"%s/64\")", cidr)
				}
			}
		}
		allErrs = append(allErrs, field.Invalid(fieldPath, cidr, detail))
	} else if !ip.Equal(ipNet.IP) {
		maskSize, _ := ipNet.Mask.Size()
		detail := fmt.Sprintf("Network contains bits outside prefix (did you mean \"%s/%d\")", ipNet.IP, maskSize)
		allErrs = append(allErrs, field.Invalid(fieldPath, cidr, detail))
	}

	return ipNet, allErrs
}

// validateIPv6CIDR verifies that `cidr` specifies a valid IPv6 network range.
// We recognize the normal CIDR syntax - e.g. `2001:db8::/32`
// We also recognize values like /64#0, meaning "the first available /64 subnet", for dynamic allocations.
// See utils.ParseCIDRNotation for details.
func validateIPv6CIDR(fieldPath *field.Path, cidr string, serviceClusterIPRange *net.IPNet) field.ErrorList {
	allErrs := field.ErrorList{}

	if strings.HasPrefix(cidr, "/") {
		newSize, _, err := utils.ParseCIDRNotation(cidr)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fieldPath, cidr, fmt.Sprintf("IPv6 CIDR subnet is not parsable: %v", err)))
			return allErrs
		}
		if newSize < 0 || newSize > 128 {
			allErrs = append(allErrs, field.Invalid(fieldPath, cidr, "IPv6 CIDR subnet size must be a value between 0 and 128"))
		}
	} else {
		subnetCIDR, errs := parseCIDR(fieldPath, cidr)
		allErrs = append(allErrs, errs...)

		if !utils.IsIPv6CIDR(cidr) {
			allErrs = append(allErrs, field.Invalid(fieldPath, cidr, "Network is not an IPv6 CIDR"))
		}
		if subnet.Overlap(subnetCIDR, serviceClusterIPRange) {
			allErrs = append(allErrs, field.Forbidden(fieldPath, fmt.Sprintf("ipv6CIDR %q must not overlap serviceClusterIPRange %q", cidr, serviceClusterIPRange)))
		}
	}

	return allErrs
}

func validateTopology(c *kops.Cluster, topology *kops.TopologySpec, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if topology.DNS != "" {
		allErrs = append(allErrs, IsValidValue(fieldPath.Child("dns", "type"), &topology.DNS, kops.SupportedDnsTypes)...)
	}

	return allErrs
}

func validateSubnets(cluster *kops.Cluster, subnets []kops.ClusterSubnetSpec, fieldPath *field.Path, strict bool, providerConstraints *cloudProviderConstraints, networkCIDRs []*net.IPNet, podCIDR, serviceClusterIPRange *net.IPNet) field.ErrorList {
	allErrs := field.ErrorList{}

	if providerConstraints.requiresSubnets && len(subnets) == 0 {
		// TODO: Auto choose zones from region?
		allErrs = append(allErrs, field.Required(fieldPath, "must configure at least one subnet (use --zones)"))
	}

	// Each subnet must be valid
	for i := range subnets {
		allErrs = append(allErrs, validateSubnet(&subnets[i], &cluster.Spec, fieldPath.Index(i), strict, providerConstraints, networkCIDRs, podCIDR, serviceClusterIPRange)...)
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
		hasID := subnets[0].ID != ""
		for i := range subnets {
			if (subnets[i].ID != "") != hasID {
				allErrs = append(allErrs, field.Forbidden(fieldPath.Index(i).Child("id"), "cannot mix subnets with specified ID and unspecified ID"))
			}
		}
	}

	if providerConstraints.requiresSubnetRegion {
		region := ""
		for i, subnet := range subnets {
			if subnet.Region == "" {
				allErrs = append(allErrs, field.Required(fieldPath.Index(i).Child("region"), "region must be specified"))
			} else {
				if region == "" {
					region = subnet.Region
				} else if region != subnet.Region {
					allErrs = append(allErrs, field.Forbidden(fieldPath.Index(i).Child("region"), "clusters cannot span regions"))
				}
			}
		}
	}

	if cluster.GetCloudProvider() != kops.CloudProviderAWS {
		for i := range subnets {
			if subnets[i].IPv6CIDR != "" {
				allErrs = append(allErrs, field.Forbidden(fieldPath.Index(i).Child("ipv6CIDR"), "ipv6CIDR can only be specified for AWS"))
			}
		}
	}

	return allErrs
}

func validateSubnet(subnetSpec *kops.ClusterSubnetSpec, c *kops.ClusterSpec, fieldPath *field.Path, strict bool, providerConstraints *cloudProviderConstraints, networkCIDRs []*net.IPNet, podCIDR, serviceClusterIPRange *net.IPNet) field.ErrorList {
	allErrs := field.ErrorList{}

	// name is required
	if subnetSpec.Name == "" {
		allErrs = append(allErrs, field.Required(fieldPath.Child("name"), ""))
	}

	// CIDR
	if subnetSpec.CIDR == "" {
		if providerConstraints.requiresSubnetCIDR && strict {
			if !strings.Contains(c.Networking.NonMasqueradeCIDR, ":") || subnetSpec.IPv6CIDR == "" {
				allErrs = append(allErrs, field.Required(fieldPath.Child("cidr"), "subnet does not have a cidr set"))
			}
		}
	} else {
		subnetCIDR, errs := parseCIDR(fieldPath.Child("cidr"), subnetSpec.CIDR)
		allErrs = append(allErrs, errs...)
		if len(networkCIDRs) > 0 && subnetCIDR != nil {
			found := false
			for _, networkCIDR := range networkCIDRs {
				if subnet.BelongsTo(networkCIDR, subnetCIDR) {
					found = true
				}
			}
			if !found {
				extraMsg := ""
				if len(networkCIDRs) > 1 {
					extraMsg = " or an additionalNetworkCIDR"
				}
				allErrs = append(allErrs, field.Forbidden(fieldPath.Child("cidr"), fmt.Sprintf("subnet %q cidr %q is not a subnet of the networkCIDR %q%s", subnetSpec.Name, subnetSpec.CIDR, c.Networking.NetworkCIDR, extraMsg)))
			}
		}
		if subnet.Overlap(subnetCIDR, podCIDR) && c.Networking.AmazonVPC == nil {
			allErrs = append(allErrs, field.Forbidden(fieldPath.Child("cidr"), fmt.Sprintf("subnet %q cidr %q must not overlap podCIDR %q", subnetSpec.Name, subnetSpec.CIDR, podCIDR)))
		}
		if subnet.Overlap(subnetCIDR, serviceClusterIPRange) {
			allErrs = append(allErrs, field.Forbidden(fieldPath.Child("cidr"), fmt.Sprintf("subnet %q cidr %q must not overlap serviceClusterIPRange %q", subnetSpec.Name, subnetSpec.CIDR, serviceClusterIPRange)))
		}
	}

	// IPv6CIDR
	if subnetSpec.IPv6CIDR != "" {
		allErrs = append(allErrs, validateIPv6CIDR(fieldPath.Child("ipv6CIDR"), subnetSpec.IPv6CIDR, serviceClusterIPRange)...)
	}

	if subnetSpec.Egress != "" {
		egressType := strings.Split(subnetSpec.Egress, "-")[0]
		if egressType != kops.EgressNatGateway && egressType != kops.EgressElasticIP && egressType != kops.EgressNatInstance && egressType != kops.EgressExternal && egressType != kops.EgressTransitGateway {
			allErrs = append(allErrs, field.Invalid(fieldPath.Child("egress"), subnetSpec.Egress,
				"egress must be of type NAT Gateway, NAT Gateway with existing ElasticIP, NAT EC2 Instance, Transit Gateway, or External"))
		}
		if subnetSpec.Egress != kops.EgressExternal && subnetSpec.Type != "DualStack" && subnetSpec.Type != "Private" && (subnetSpec.IPv6CIDR == "" || subnetSpec.Type != "Public") {
			allErrs = append(allErrs, field.Forbidden(fieldPath.Child("egress"), "egress can only be specified for private or IPv6-capable public subnets"))
		}
	}

	allErrs = append(allErrs, IsValidValue(fieldPath.Child("type"), &subnetSpec.Type, []kops.SubnetType{
		kops.SubnetTypePublic,
		kops.SubnetTypePrivate,
		kops.SubnetTypeDualStack,
		kops.SubnetTypeUtility,
	})...)

	if subnetSpec.Type == kops.SubnetTypeDualStack && !c.IsIPv6Only() {
		allErrs = append(allErrs, field.Forbidden(fieldPath.Child("type"), "subnet type DualStack may only be used in IPv6 clusters"))
	}

	if c.CloudProvider.Openstack != nil {
		if c.CloudProvider.Openstack.Router == nil || c.CloudProvider.Openstack.Router.ExternalNetwork == nil {
			if subnetSpec.Type == kops.SubnetTypePublic {
				allErrs = append(allErrs, field.Forbidden(fieldPath.Child("type"), "subnet type Public requires an external network"))
			}
		}
	}

	if c.CloudProvider.AWS != nil && subnetSpec.AdditionalRoutes != nil {
		if len(subnetSpec.ID) > 0 {
			allErrs = append(allErrs, field.Forbidden(fieldPath.Child("additionalRoutes"), "additional routes cannot be added if the subnet is shared"))
		} else if subnetSpec.Type != kops.SubnetTypePrivate {
			allErrs = append(allErrs, field.Forbidden(fieldPath.Child("additionalRoutes"), "additional routes can only be added on private subnets"))
		}
		allErrs = append(allErrs, awsValidateAdditionalRoutes(fieldPath.Child("additionalRoutes"), subnetSpec.AdditionalRoutes, networkCIDRs)...)
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
	if v.Enabled != nil && !*v.Enabled {
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

func validateKubeAPIServer(v *kops.KubeAPIServerConfig, c *kops.Cluster, fldPath *field.Path, strict bool) field.ErrorList {
	allErrs := field.ErrorList{}

	if v.AuthenticationConfigFile != "" && c.Spec.Authentication != nil && c.Spec.Authentication.OIDC != nil {
		o := c.Spec.Authentication.OIDC
		if o.UsernameClaim != nil || o.UsernamePrefix != nil || o.GroupsClaims != nil || o.GroupsPrefix != nil || o.IssuerURL != nil || o.ClientID != nil || o.RequiredClaims != nil {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("authenticationConfigFile"), "authenticationConfigFile is mutually exclusive with OIDC options, remove all existing OIDC options to use authenticationConfigFile"))
		}
	}

	if fi.ValueOf(v.EnableBootstrapAuthToken) {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("enableBootstrapTokenAuth"), "bootstrap tokens are not supported"))
	}

	if len(v.AdmissionControl) > 0 {
		if len(v.DisableAdmissionPlugins) > 0 {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("admissionControl"),
				"admissionControl is mutually exclusive with disableAdmissionPlugins˚"))
		}

		if c.IsKubernetesGTE("1.26") {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("admissionControl"), "admissionControl has been replaced with enableAdmissionPlugins"))
		}
	}

	for _, plugin := range v.EnableAdmissionPlugins {
		if plugin == "PodSecurityPolicy" && c.IsKubernetesGTE("1.25") {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("enableAdmissionPlugins"),
				"PodSecurityPolicy has been removed from Kubernetes 1.25"))
		}
	}

	for _, plugin := range v.AdmissionControl {
		if plugin == "PodSecurityPolicy" && c.IsKubernetesGTE("1.25") {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("admissionControl"),
				"PodSecurityPolicy has been removed from Kubernetes 1.25"))
		}
	}

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

	if v.AuthorizationMode != nil {
		if strings.Contains(*v.AuthorizationMode, "Webhook") {
			if v.AuthorizationWebhookConfigFile == nil {
				allErrs = append(allErrs, field.Required(fldPath.Child("authorizationWebhookConfigFile"), "Authorization mode Webhook requires authorizationWebhookConfigFile to be specified"))
			}
		}

		if c.Spec.Authorization != nil && c.Spec.Authorization.RBAC != nil {

			var hasNode, hasRBAC bool
			for _, mode := range strings.Split(*v.AuthorizationMode, ",") {
				switch mode {
				case "Node":
					hasNode = true
				case "RBAC":
					hasRBAC = true
				default:
					allErrs = append(allErrs, IsValidValue(fldPath.Child("authorizationMode"), &mode, []string{"ABAC", "Webhook", "Node", "RBAC", "AlwaysAllow", "AlwaysDeny"})...)
				}
			}
			if c.GetCloudProvider() == kops.CloudProviderAWS {
				if !hasNode || !hasRBAC {
					allErrs = append(allErrs, field.Required(fldPath.Child("authorizationMode"), "As of kubernetes 1.19 on AWS, authorizationMode must include RBAC and Node"))
				}
			}
		}
	}

	if v.LogFormat != "" {
		allErrs = append(allErrs, IsValidValue(fldPath.Child("logFormat"), &v.LogFormat, []string{"text", "json"})...)
	}

	if v.InsecurePort != nil {
		field.Forbidden(fldPath.Child("insecurePort"), "insecurePort must not be set as of Kubernetes 1.24")
	}

	if v.AuditPolicyFile == "" && v.AuditWebhookConfigFile != "" {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("auditWebhookConfigFile"), v.AuditWebhookConfigFile, "an audit policy is required for the audit webhook config"))
	}
	if v.AuditPolicyFile != "" && v.AuditWebhookConfigFile != "" {
		auditPolicyDir := filepath.Dir(v.AuditPolicyFile)
		auditWebhookConfigDir := filepath.Dir(v.AuditWebhookConfigFile)
		if auditPolicyDir != auditWebhookConfigDir {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("auditWebhookConfigFile"), v.AuditWebhookConfigFile, "the audit webhook config must be placed in the same directory as the audit policy"))
		}
	}

	if v.ServiceClusterIPRange != c.Spec.Networking.ServiceClusterIPRange {
		if strict || v.ServiceClusterIPRange != "" {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("serviceClusterIPRange"), "kubeAPIServer serviceClusterIPRange did not match cluster serviceClusterIPRange"))
		}
	}

	return allErrs
}

func validateKubeControllerManager(v *kops.KubeControllerManagerConfig, c *kops.Cluster, fldPath *field.Path, strict bool) field.ErrorList {
	allErrs := field.ErrorList{}

	// We aren't aiming to do comprehensive validation, but we can add some best-effort validation where it helps guide users
	// Users reported encountered this in #15909
	if v.ExperimentalClusterSigningDuration != nil {
		if c.IsKubernetesGTE("1.25") {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("experimentalClusterSigningDuration"), "experimentalClusterSigningDuration has been replaced with clusterSigningDuration as of kubernetes 1.25"))
		}
	}

	return allErrs
}

func validateKubeScheduler(v *kops.KubeSchedulerConfig, c *kops.Cluster, fldPath *field.Path, strict bool) field.ErrorList {
	allErrs := field.ErrorList{}

	// We aren't aiming to do comprehensive validation, but we can add some best-effort validation where it helps guide users.
	// Users reported encountered this in #16388
	if v.UsePolicyConfigMap != nil {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("usePolicyConfigMap"), "usePolicyConfigMap is deprecated, use KubeSchedulerConfiguration"))
	}

	return allErrs
}

func validateKubeProxy(k *kops.KubeProxyConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	master := k.Master

	for i, x := range k.IPVSExcludeCIDRs {
		if _, _, err := net.ParseCIDR(x); err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("ipvsExcludeCIDRs").Index(i), x, "Invalid network CIDR"))
		}
	}

	if master != "" && !isValidAPIServersURL(master) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("master"), master, "Not a valid APIServer URL"))
	}

	return allErrs
}

func validateKubelet(k *kops.KubeletConfigSpec, c *kops.Cluster, kubeletPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if k != nil {

		{
			// Flag removed in 1.5
			if k.ConfigureCBR0 != nil {
				allErrs = append(allErrs, field.Forbidden(
					kubeletPath.Child("ConfigureCBR0"),
					"configure-cbr0 flag was removed in 1.5"))
			}
		}

		{
			// Flag removed in 1.6
			if k.APIServers != "" {
				allErrs = append(allErrs, field.Forbidden(
					kubeletPath.Child("apiServers"),
					"api-servers flag was removed in 1.6"))
			}
		}

		{
			// Flag removed in 1.10
			if k.RequireKubeconfig != nil {
				allErrs = append(allErrs, field.Forbidden(
					kubeletPath.Child("requireKubeconfig"),
					"require-kubeconfig flag was removed in 1.10.  (Please be sure you are not using a cluster config from `kops get cluster --full`)"))
			}
		}

		if k.ExperimentalAllowedUnsafeSysctls != nil {
			allErrs = append(allErrs, field.Forbidden(
				kubeletPath.Child("experimentalAllowedUnsafeSysctls"),
				"experimentalAllowedUnsafeSysctls was renamed in k8s 1.11; please use allowedUnsafeSysctls instead"))
		}

		if k.BootstrapKubeconfig != "" {
			allErrs = append(allErrs, field.Forbidden(kubeletPath.Child("bootstrapKubeconfig"), "bootstrap tokens are not supported"))
		}

		if k.TopologyManagerPolicy != "" {
			allErrs = append(allErrs, IsValidValue(kubeletPath.Child("topologyManagerPolicy"), &k.TopologyManagerPolicy, []string{"none", "best-effort", "restricted", "single-numa-node"})...)
		}

		if k.EnableCadvisorJsonEndpoints != nil {
			allErrs = append(allErrs, field.Forbidden(kubeletPath.Child("enableCadvisorJsonEndpoints"), "enableCadvisorJsonEndpoints requires Kubernetes 1.18-1.20"))
		}

		if k.LogFormat != "" {
			allErrs = append(allErrs, IsValidValue(kubeletPath.Child("logFormat"), &k.LogFormat, []string{"text", "json"})...)
		}

		if k.CPUCFSQuotaPeriod != nil {
			allErrs = append(allErrs, field.Forbidden(kubeletPath.Child("cpuCFSQuotaPeriod"), "cpuCFSQuotaPeriod has been removed on Kubernetes >=1.20"))
		}

		if k.NetworkPluginName != nil {
			allErrs = append(allErrs, field.Forbidden(kubeletPath.Child("networkPluginName"), "networkPluginName has been removed on Kubernetes >=1.24"))
		}
		if k.NetworkPluginMTU != nil {
			allErrs = append(allErrs, field.Forbidden(kubeletPath.Child("networkPluginMTU"), "networkPluginMTU has been removed on Kubernetes >=1.24"))
		}
		if k.NonMasqueradeCIDR != nil {
			allErrs = append(allErrs, field.Forbidden(kubeletPath.Child("nonMasqueradeCIDR"), "nonMasqueradeCIDR has been removed on Kubernetes >=1.24"))
		}

		if k.ShutdownGracePeriodCriticalPods != nil {
			if k.ShutdownGracePeriod == nil {
				allErrs = append(allErrs, field.Forbidden(kubeletPath.Child("shutdownGracePeriodCriticalPods"), "shutdownGracePeriodCriticalPods require shutdownGracePeriod"))
			}
			if k.ShutdownGracePeriod.Duration.Seconds() < k.ShutdownGracePeriodCriticalPods.Seconds() {
				allErrs = append(allErrs, field.Invalid(kubeletPath.Child("shutdownGracePeriodCriticalPods"), k.ShutdownGracePeriodCriticalPods.String(), "shutdownGracePeriodCriticalPods cannot be greater than shutdownGracePeriod"))
			}
		}

		if k.MemorySwapBehavior != "" {
			allErrs = append(allErrs, IsValidValue(kubeletPath.Child("memorySwapBehavior"), &k.MemorySwapBehavior, []string{"LimitedSwap", "UnlimitedSwap"})...)
		}
	}
	return allErrs
}

func validateNetworking(cluster *kops.Cluster, v *kops.NetworkingSpec, fldPath *field.Path, strict bool, providerConstraints *cloudProviderConstraints) field.ErrorList {
	c := &cluster.Spec
	allErrs := field.ErrorList{}

	var networkCIDRs []*net.IPNet

	if v.NetworkCIDR == "" {
		if providerConstraints.requiresNetworkCIDR {
			allErrs = append(allErrs, field.Required(fldPath.Child("networkCIDR"), "Cluster does not have networkCIDR set"))
		}
	} else if providerConstraints.prohibitsNetworkCIDR {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("networkCIDR"), fmt.Sprintf("%s doesn't support networkCIDR", cluster.GetCloudProvider())))
	} else {
		networkCIDR, errs := parseCIDR(fldPath.Child("networkCIDR"), v.NetworkCIDR)
		allErrs = append(allErrs, errs...)
		if networkCIDR != nil {
			networkCIDRs = append(networkCIDRs, networkCIDR)
		}

		if cluster.GetCloudProvider() == kops.CloudProviderDO {
			// verify if the NetworkCIDR is in a private range as per RFC1918
			if networkCIDR != nil && !networkCIDR.IP.IsPrivate() {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("networkCIDR"), v.NetworkCIDR, "networkCIDR must be within a private IP range"))
			}
			// verify if networkID is not specified. In case of DO, this is mutually exclusive.
			if v.NetworkID != "" {
				allErrs = append(allErrs, field.Forbidden(fldPath.Child("networkCIDR"), "DO doesn't support specifying both NetworkID and NetworkCIDR"))
			}
		}
	}

	if len(v.AdditionalNetworkCIDRs) > 0 && providerConstraints.prohibitsMultipleNetworkCIDRs {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("additionalNetworkCIDRs"), fmt.Sprintf("%s doesn't support additionalNetworkCIDRs", cluster.GetCloudProvider())))
	} else {
		for i, cidr := range v.AdditionalNetworkCIDRs {
			networkCIDR, errs := parseCIDR(fldPath.Child("additionalNetworkCIDRs").Index(i), cidr)
			allErrs = append(allErrs, errs...)
			if networkCIDR != nil {
				networkCIDRs = append(networkCIDRs, networkCIDR)
			}
		}
	}

	var nonMasqueradeCIDRs []*net.IPNet
	{
		if v.NonMasqueradeCIDR == "" {
			if providerConstraints.requiresNonMasqueradeCIDR {
				allErrs = append(allErrs, field.Required(fldPath.Child("nonMasqueradeCIDR"), "Cluster does not have nonMasqueradeCIDR set"))
			}
		} else {
			nonMasqueradeCIDR, errs := parseCIDR(fldPath.Child("nonMasqueradeCIDR"), v.NonMasqueradeCIDR)
			allErrs = append(allErrs, errs...)
			if nonMasqueradeCIDR != nil {
				nonMasqueradeCIDRs = append(nonMasqueradeCIDRs, nonMasqueradeCIDR)
			}

			if strings.Contains(v.NonMasqueradeCIDR, ":") && v.NonMasqueradeCIDR != "::/0" {
				allErrs = append(allErrs, field.Forbidden(fldPath.Child("nonMasqueradeCIDR"), "IPv6 clusters must have a nonMasqueradeCIDR of \"::/0\""))
			}

			if len(nonMasqueradeCIDRs) > 0 && len(networkCIDRs) > 0 && v.AmazonVPC == nil && (v.Cilium == nil || v.Cilium.IPAM != kops.CiliumIpamEni) {
				if subnet.Overlap(nonMasqueradeCIDRs[0], networkCIDRs[0]) {
					allErrs = append(allErrs, field.Forbidden(fldPath.Child("nonMasqueradeCIDR"), fmt.Sprintf("nonMasqueradeCIDR %q cannot overlap with networkCIDR %q", v.NonMasqueradeCIDR, v.NetworkCIDR)))
				}
				for i, cidr := range networkCIDRs[1:] {
					if subnet.Overlap(nonMasqueradeCIDRs[0], cidr) {
						allErrs = append(allErrs, field.Forbidden(fldPath.Child("nonMasqueradeCIDR"), fmt.Sprintf("nonMasqueradeCIDR %q cannot overlap with additionalNetworkCIDRs[%d] %q", v.NonMasqueradeCIDR, i, cidr)))
					}
				}
			}
		}
	}

	var podCIDR *net.IPNet
	{
		if v.PodCIDR == "" {
			if strict && !cluster.Spec.IsKopsControllerIPAM() {
				allErrs = append(allErrs, field.Required(fldPath.Child("podCIDR"), "Cluster did not have podCIDR set"))
			}
		} else {
			var errs field.ErrorList
			podCIDR, errs = parseCIDR(fldPath.Child("podCIDR"), v.PodCIDR)
			allErrs = append(allErrs, errs...)

			if podCIDR != nil {
				if len(nonMasqueradeCIDRs) > 0 && !subnet.BelongsTo(nonMasqueradeCIDRs[0], podCIDR) {
					allErrs = append(allErrs, field.Forbidden(fldPath.Child("podCIDR"), fmt.Sprintf("podCIDR %q must be a subnet of nonMasqueradeCIDR %q", podCIDR, nonMasqueradeCIDRs[0])))
				}
			}
		}
	}

	var serviceClusterIPRange *net.IPNet
	{
		if v.ServiceClusterIPRange == "" {
			if strict {
				allErrs = append(allErrs, field.Required(fldPath.Child("serviceClusterIPRange"), "Cluster did not have serviceClusterIPRange set"))
			}
		} else {
			var errs field.ErrorList
			serviceClusterIPRange, errs = parseCIDR(fldPath.Child("serviceClusterIPRange"), v.ServiceClusterIPRange)
			allErrs = append(allErrs, errs...)

			// Removed as part of #16340; we previously supported this and it seems to work fine.
			// We may add back if we find problems and have a path for migrating existing clusters.
			// if subnet.Overlap(podCIDR, serviceClusterIPRange) {
			// 	allErrs = append(allErrs, field.Forbidden(fldPath.Child("serviceClusterIPRange"), fmt.Sprintf("serviceClusterIPRange %q must not overlap podCIDR %q", serviceClusterIPRange, podCIDR)))
			// }
		}
	}

	allErrs = append(allErrs, validateSubnets(cluster, v.Subnets, fldPath.Child("subnets"), strict, providerConstraints, networkCIDRs, podCIDR, serviceClusterIPRange)...)

	if v.Topology != nil {
		allErrs = append(allErrs, validateTopology(cluster, v.Topology, fldPath.Child("topology"))...)
	}

	optionTaken := false

	if v.Classic != nil {
		allErrs = append(allErrs, field.Invalid(fldPath, "classic", "classic networking is not supported"))
	}

	if v.Kubenet != nil {
		optionTaken = true

		if cluster.Spec.IsIPv6Only() {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("kubenet"), "Kubenet does not support IPv6"))
		}
	}

	if v.External != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("external"), "only one networking option permitted"))
		}

		if cluster.IsKubernetesGTE("1.26") {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("external"), "external is not supported for Kubernetes >= 1.26"))
		}
		optionTaken = true
	}

	if v.Kopeio != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("kopeio"), "only one networking option permitted"))
		}
		optionTaken = true

		if cluster.Spec.IsIPv6Only() {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("kopeio"), "Kopeio does not support IPv6"))
		}
	}

	if v.CNI != nil && optionTaken {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("cni"), "only one networking option permitted"))
	}

	if v.Weave != nil {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("weave"), "Weave is no longer supported"))
	}

	if v.Flannel != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("flannel"), "only one networking option permitted"))
		}
		optionTaken = true

		if cluster.IsKubernetesGTE("1.28") {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("flannel"), "Flannel is not supported for Kubernetes >= 1.28"))
		} else {
			allErrs = append(allErrs, validateNetworkingFlannel(cluster, v.Flannel, fldPath.Child("flannel"))...)
		}
	}

	if v.Calico != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("calico"), "only one networking option permitted"))
		}
		optionTaken = true

		allErrs = append(allErrs, validateNetworkingCalico(&cluster.Spec, v.Calico, fldPath.Child("calico"))...)
	}

	if v.Canal != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("canal"), "only one networking option permitted"))
		}
		optionTaken = true

		if cluster.IsKubernetesGTE("1.28") {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("canal"), "Canal is not supported for Kubernetes >= 1.28"))
		} else {
			allErrs = append(allErrs, validateNetworkingCanal(cluster, v.Canal, fldPath.Child("canal"))...)
		}
	}

	if v.KubeRouter != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("kubeRouter"), "only one networking option permitted"))
		}
		if c.KubeProxy != nil && (c.KubeProxy.Enabled == nil || *c.KubeProxy.Enabled) {
			allErrs = append(allErrs, field.Forbidden(fldPath.Root().Child("spec", "kubeProxy", "enabled"), "kube-router requires kubeProxy to be disabled"))
		}
		optionTaken = true

		if cluster.Spec.IsIPv6Only() {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("kubeRouter"), "kube-router does not support IPv6"))
		}
	}

	if v.Romana != nil {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("romana"), "support for Romana has been removed"))
	}

	if v.AmazonVPC != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("amazonVPC"), "only one networking option permitted"))
		}
		optionTaken = true

		if cluster.GetCloudProvider() != kops.CloudProviderAWS {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("amazonVPC"), "amazon-vpc-routed-eni networking is supported only in AWS"))
		}

		if cluster.Spec.IsIPv6Only() {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("amazonVPC"), "amazon-vpc-routed-eni networking does not support IPv6"))
		}

	}

	if v.Cilium != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("cilium"), "only one networking option permitted"))
		}
		optionTaken = true

		allErrs = append(allErrs, validateNetworkingCilium(cluster, v.Cilium, fldPath.Child("cilium"))...)
	}

	if v.LyftVPC != nil {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("lyftvp"), "support for LyftVPC has been removed"))
	}

	if v.GCP != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("gcp"), "only one networking option permitted"))
		}

		allErrs = append(allErrs, validateNetworkingGCP(cluster, v.GCP, fldPath.Child("gcp"))...)
	}

	return allErrs
}

func validateNetworkingFlannel(c *kops.Cluster, v *kops.FlannelNetworkingSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if c.Spec.IsIPv6Only() {
		allErrs = append(allErrs, field.Forbidden(fldPath, "Flannel does not support IPv6"))
	}

	if v.Backend == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("backend"), "Flannel backend must be specified"))
	} else {
		allErrs = append(allErrs, IsValidValue(fldPath.Child("backend"), &v.Backend, []string{"udp", "vxlan"})...)
	}

	return allErrs
}

func validateNetworkingCanal(c *kops.Cluster, v *kops.CanalNetworkingSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if c.Spec.IsIPv6Only() {
		allErrs = append(allErrs, field.Forbidden(fldPath, "Canal does not support IPv6"))
	}

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

func validateNetworkingCilium(cluster *kops.Cluster, v *kops.CiliumNetworkingSpec, fldPath *field.Path) field.ErrorList {
	c := &cluster.Spec
	allErrs := field.ErrorList{}

	if v.Version != "" {
		versionFld := fldPath.Child("version")
		if !strings.HasPrefix(v.Version, "v") {
			return append(allErrs, field.Invalid(versionFld, v.Version, "Cilium version must be prefixed with 'v'"))
		}
		versionString := strings.TrimPrefix(v.Version, "v")
		version, err := semver.Parse(versionString)

		version.Pre = nil
		version.Build = nil
		if err != nil {
			allErrs = append(allErrs, field.Invalid(versionFld, v.Version, "Could not parse as semantic version"))
		}

		if version.Minor != 16 {
			allErrs = append(allErrs, field.Invalid(versionFld, v.Version, "Only version 1.16 is supported"))
		}

		if v.Hubble != nil && fi.ValueOf(v.Hubble.Enabled) {
			if !components.IsCertManagerEnabled(cluster) {
				allErrs = append(allErrs, field.Forbidden(fldPath.Child("hubble", "enabled"), "Hubble requires that cert manager is enabled"))
			}
		}
	}

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

	if v.IdentityAllocationMode != "" {
		allErrs = append(allErrs, IsValidValue(fldPath.Child("identityAllocationMode"), &v.IdentityAllocationMode, []string{"crd", "kvstore"})...)

		if v.IdentityAllocationMode == "kvstore" && !v.EtcdManaged {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("identityAllocationMode"), "Cilium requires managed etcd to allocate identities on kvstore mode"))
		}
	}

	if v.BPFLBAlgorithm != "" {
		allErrs = append(allErrs, IsValidValue(fldPath.Child("bpfLBAlgorithm"), &v.BPFLBAlgorithm, []string{"random", "maglev"})...)
	}

	if v.EnableEncryption && c.IsIPv6Only() {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("enableEncryption"), "encryption is not supported on IPv6 clusters"))
	}

	if v.EncryptionType != "" {
		if !v.EnableEncryption {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("encryptionType"), "encryptionType requires enableEncryption"))
		}

		allErrs = append(allErrs, IsValidValue(fldPath.Child("encryptionType"), &v.EncryptionType, []kops.CiliumEncryptionType{kops.CiliumEncryptionTypeIPSec, kops.CiliumEncryptionTypeWireguard})...)
	}

	if fi.ValueOf(v.EnableL7Proxy) && v.InstallIptablesRules != nil && !*v.InstallIptablesRules {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("enableL7Proxy"), "Cilium L7 Proxy requires installIptablesRules."))
	}

	if v.IPAM != "" {
		// "azure" not supported by kops
		allErrs = append(allErrs, IsValidValue(fldPath.Child("ipam"), &v.IPAM, []string{"hostscope", "kubernetes", "crd", "eni"})...)

		if v.IPAM == kops.CiliumIpamEni {
			if cluster.GetCloudProvider() != kops.CloudProviderAWS {
				allErrs = append(allErrs, field.Forbidden(fldPath.Child("ipam"), "Cilum ENI IPAM is supported only in AWS"))
			}
			if v.Masquerade != nil && !*v.Masquerade {
				allErrs = append(allErrs, field.Forbidden(fldPath.Child("masquerade"), "Masquerade must be enabled when ENI IPAM is used"))
			}
			if c.IsIPv6Only() {
				allErrs = append(allErrs, field.Forbidden(fldPath.Child("ipam"), "Cilium ENI IPAM does not support IPv6"))
			}
		}
	}

	if v.EtcdManaged {
		hasCiliumCluster := false
		for _, cluster := range c.EtcdClusters {
			if cluster.Name == "cilium" {
				hasCiliumCluster = true
				break
			}
		}
		if !hasCiliumCluster {
			allErrs = append(allErrs, field.Required(fldPath.Root().Child("etcdClusters"), "Cilium with managed etcd requires a dedicated etcd cluster"))
		}
	}

	if v.Ingress != nil && fi.ValueOf(v.Ingress.Enabled) {
		if v.Ingress.DefaultLoadBalancerMode != "" {
			allErrs = append(allErrs, IsValidValue(fldPath.Child("ingress", "defaultLoadBalancerMode"), &v.Ingress.DefaultLoadBalancerMode, []string{"shared", "dedicated"})...)
		}
	}

	return allErrs
}

func validateNetworkingGCP(cluster *kops.Cluster, v *kops.GCPNetworkingSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	c := cluster.Spec

	if cluster.GetCloudProvider() != kops.CloudProviderGCE {
		allErrs = append(allErrs, field.Forbidden(fldPath, "GCP networking is supported only when on GCP"))
	}

	if c.IsIPv6Only() {
		allErrs = append(allErrs, field.Forbidden(fldPath, "GCP networking does not support IPv6"))
	}

	return allErrs
}

func validateAdditionalPolicy(role string, policy string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	var valid []string
	for _, r := range kops.AllInstanceGroupRoles {
		valid = append(valid, r.ToLowerString())
	}
	allErrs = append(allErrs, IsValidValue(fldPath, &role, valid)...)

	statements, err := iam.ParseStatements(policy)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath.Key(role), policy, "policy was not valid JSON: "+err.Error()))
	}

	// Trivial validation of policy, mostly to make sure it isn't some other random object
	for i, statement := range statements {
		fldEffect := fldPath.Key(role).Index(i).Child("Effect")
		if statement.Effect == "" {
			allErrs = append(allErrs, field.Required(fldEffect, "Effect must be specified for IAM policy"))
		} else {
			allErrs = append(allErrs, IsValidValue(fldEffect, &statement.Effect, []iam.StatementEffect{iam.StatementEffectAllow, iam.StatementEffectDeny})...)
		}
	}

	return allErrs
}

func validateExternalPolicies(role string, policies []string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	var valid []string
	for _, r := range kops.AllInstanceGroupRoles {
		valid = append(valid, r.ToLowerString())
	}
	allErrs = append(allErrs, IsValidValue(fldPath, &role, valid)...)

	for _, policy := range policies {
		parsedARN, err := arn.Parse(policy)
		if err != nil || !strings.HasPrefix(parsedARN.Resource, "policy/") {
			allErrs = append(allErrs, field.Invalid(fldPath.Child(role), policy,
				"Policy must be a valid AWS ARN such as arn:aws:iam::123456789012:policy/KopsExamplePolicy"))
		}
	}

	return allErrs
}

func validateEtcdClusterSpec(spec kops.EtcdClusterSpec, c *kops.Cluster, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, IsValidValue(fieldPath.Child("name"), &spec.Name, []string{"cilium", "main", "events"})...)

	if spec.Provider != "" {
		allErrs = append(allErrs, IsValidValue(fieldPath.Child("provider"), &spec.Provider, []kops.EtcdProviderType{kops.EtcdProviderTypeManager})...)
	}
	if len(spec.Members) == 0 {
		allErrs = append(allErrs, field.Required(fieldPath.Child("etcdMembers"), "No members defined in etcd cluster"))
	} else if (len(spec.Members) % 2) == 0 {
		// Not technically a requirement, but doesn't really make sense to allow
		allErrs = append(allErrs, field.Invalid(fieldPath.Child("etcdMembers"), len(spec.Members), "Should be an odd number of control-plane-zones for quorum. Use --zones and --control-plane-zones to declare node zones and control-plane zones separately"))
	}
	allErrs = append(allErrs, validateEtcdVersion(spec, fieldPath, nil)...)
	for i, m := range spec.Members {
		allErrs = append(allErrs, validateEtcdMemberSpec(m, fieldPath.Child("etcdMembers").Index(i))...)
	}

	return allErrs
}

// validateEtcdBackupStore checks that the etcd clusters backupStore path is unique.
func validateEtcdBackupStore(specs []kops.EtcdClusterSpec, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	etcdBackupStore := make(map[string]bool)
	for _, x := range specs {
		if _, alreadyUsed := etcdBackupStore[x.Name]; alreadyUsed {
			allErrs = append(allErrs, field.Forbidden(fieldPath.Index(0).Child("backupStore"), "the backup store must be unique for each etcd cluster"))
		}
		etcdBackupStore[x.Name] = true
	}

	return allErrs
}

// validateEtcdStorage is responsible for checking versions are identical.
func validateEtcdStorage(specs []kops.EtcdClusterSpec, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	version := specs[0].Version
	for i, x := range specs {
		if x.Version != "" && x.Version != version {
			allErrs = append(allErrs, field.Forbidden(fieldPath.Index(i).Child("version"), fmt.Sprintf("cluster: %q, has a different storage version: %q, both must be the same", x.Name, x.Version)))
		}
	}

	return allErrs
}

// validateEtcdVersion is responsible for validating the storage version of etcd
func validateEtcdVersion(spec kops.EtcdClusterSpec, fieldPath *field.Path, minimalVersion *semver.Version) field.ErrorList {
	if spec.Version == "" {
		return nil
	}

	version := spec.Version

	sem, err := semver.Parse(strings.TrimPrefix(version, "v"))
	if err != nil {
		return field.ErrorList{field.Invalid(fieldPath.Child("version"), version, "the storage version is invalid")}
	}

	// we only support v3 for now
	if sem.Major == 3 {
		if minimalVersion != nil && sem.LT(*minimalVersion) {
			return field.ErrorList{field.Invalid(fieldPath.Child("version"), version, fmt.Sprintf("minimum version required is %s", minimalVersion.String()))}
		}
		return nil
	}

	return field.ErrorList{field.Invalid(fieldPath.Child("version"), version, "unsupported storage version, we only support major version 3")}
}

// validateEtcdMemberSpec is responsible for validate the cluster member
func validateEtcdMemberSpec(spec kops.EtcdMemberSpec, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if spec.Name == "" {
		allErrs = append(allErrs, field.Required(fieldPath.Child("name"), "etcdMember did not have name"))
	}

	if fi.ValueOf(spec.InstanceGroup) == "" {
		allErrs = append(allErrs, field.Required(fieldPath.Child("instanceGroup"), "etcdMember did not have instanceGroup"))
	}

	return allErrs
}

func validateNetworkingCalico(c *kops.ClusterSpec, v *kops.CalicoNetworkingSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if v.AWSSrcDstCheck != "" {
		if c.IsIPv6Only() && v.AWSSrcDstCheck != "DoNothing" {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("awsSrcDstCheck"), "awsSrcDstCheck may only be \"DoNothing\" for IPv6 clusters"))
		} else {
			valid := []string{"Enable", "Disable", "DoNothing"}
			allErrs = append(allErrs, IsValidValue(fldPath.Child("awsSrcDstCheck"), &v.AWSSrcDstCheck, valid)...)
		}
	}

	if v.CrossSubnet != nil {
		if fi.ValueOf(v.CrossSubnet) && v.AWSSrcDstCheck != "Disable" {
			field.Invalid(fldPath.Child("crossSubnet"), v.CrossSubnet, "crossSubnet is deprecated, use awsSrcDstCheck instead")
		}
	}

	if v.BPFExternalServiceMode != "" {
		valid := []string{"Tunnel", "DSR"}
		allErrs = append(allErrs, IsValidValue(fldPath.Child("bpfExternalServiceMode"), &v.BPFExternalServiceMode, valid)...)
	}

	if v.BPFLogLevel != "" {
		valid := []string{"Off", "Info", "Debug"}
		allErrs = append(allErrs, IsValidValue(fldPath.Child("bpfLogLevel"), &v.BPFLogLevel, valid)...)
	}

	if v.ChainInsertMode != "" {
		valid := []string{"insert", "append"}
		allErrs = append(allErrs, IsValidValue(fldPath.Child("chainInsertMode"), &v.ChainInsertMode, valid)...)
	}

	if v.EncapsulationMode != "" {
		valid := []string{"ipip", "vxlan", "none"}
		allErrs = append(allErrs, IsValidValue(fldPath.Child("encapsulationMode"), &v.EncapsulationMode, valid)...)

		if v.EncapsulationMode != "none" && c.IsIPv6Only() {
			// IPv6 doesn't support encapsulation and kops only uses the "none" networking backend.
			// The bird networking backend could also be added in the future if there's any valid use case.
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("encapsulationMode"), "IPv6 requires an encapsulationMode of \"none\""))
		} else if v.EncapsulationMode == "none" && !c.IsIPv6Only() {
			// Don't tolerate "None" for now, which would disable encapsulation in the default IPPool
			// object. Note that with no encapsulation, we'd need to select the "bird" networking
			// backend in order to allow use of BGP to distribute routes for pod traffic.
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("encapsulationMode"), "encapsulationMode \"none\" is only supported for IPv6 clusters"))
		}
	}

	if v.IPIPMode != "" {
		child := fldPath.Child("ipipMode")
		allErrs = append(allErrs, validateCalicoEncapsulationMode(v.IPIPMode, child)...)
		if v.IPIPMode != "Never" {
			if v.EncapsulationMode != "" && v.EncapsulationMode != "ipip" {
				allErrs = append(allErrs, field.Forbidden(child, `IP-in-IP encapsulation requires use of Calico's "ipip" encapsulation mode`))
			}
		}
	}

	if v.VXLANMode != "" {
		child := fldPath.Child("vxlanMode")
		allErrs = append(allErrs, validateCalicoEncapsulationMode(v.VXLANMode, child)...)
		if v.VXLANMode != "Never" {
			if v.EncapsulationMode != "" && v.EncapsulationMode != "vxlan" {
				allErrs = append(allErrs, field.Forbidden(child, `VXLAN encapsulation requires use of Calico's "vxlan" encapsulation mode`))
			}
		}
	}

	if v.IPv4AutoDetectionMethod != "" {
		allErrs = append(allErrs, validateCalicoAutoDetectionMethod(fldPath.Child("ipv4AutoDetectionMethod"), v.IPv4AutoDetectionMethod, ipv4.Version)...)
	}

	if v.IPv6AutoDetectionMethod != "" {
		allErrs = append(allErrs, validateCalicoAutoDetectionMethod(fldPath.Child("ipv6AutoDetectionMethod"), v.IPv6AutoDetectionMethod, ipv6.Version)...)
	}

	if v.IptablesBackend != "" {
		valid := []string{"Auto", "Legacy", "NFT"}
		allErrs = append(allErrs, IsValidValue(fldPath.Child("iptablesBackend"), &v.IptablesBackend, valid)...)
	}

	if v.TyphaReplicas < 0 {
		allErrs = append(allErrs,
			field.Invalid(fldPath.Child("typhaReplicas"), v.TyphaReplicas,
				fmt.Sprintf("Unable to set number of Typha replicas to less than 0, you've specified %d", v.TyphaReplicas)))
	}

	if v.WireguardEnabled && c.IsIPv6Only() {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("wireguardEnabled"), `WireGuard is not supported on IPv6 clusters`))
	}

	return allErrs
}

func validateCalicoAutoDetectionMethod(fldPath *field.Path, runtime string, version int) field.ErrorList {
	validationError := field.ErrorList{}

	// validation code is based on the checks in calico/node startup code
	// valid formats are "first-found", "can-reach=DEST", or
	// "(skip-)interface=<COMMA-SEPARATED LIST OF INTERFACES>"
	//
	// We won't do deep validation of the values in this check, since they can
	// be actual interface names or regexes
	method := strings.Split(runtime, "=")
	if len(method) == 0 {
		return field.ErrorList{field.Invalid(fldPath, runtime, "missing autodetection method")}
	}
	if len(method) > 2 {
		return field.ErrorList{field.Invalid(fldPath, runtime, "malformed autodetection method")}
	}

	// 'method' should contain something like "[interface eth0,en.*]" or "[first-found]"
	switch method[0] {
	case "first-found":
		return nil
	case "can-reach":
		destStr := method[1]
		if version == ipv4.Version {
			return utilvalidation.IsValidIPv4Address(fldPath, destStr)
		} else if version == ipv6.Version {
			return utilvalidation.IsValidIPv6Address(fldPath, destStr)
		}

		return field.ErrorList{field.InternalError(fldPath, errors.New("IP version is incorrect"))}
	case "interface":
		ifRegexes := regexp.MustCompile(`\s*,\s*`).Split(method[1], -1)
		if len(ifRegexes) == 0 || ifRegexes[0] == "" {
			validationError = append(validationError, field.Invalid(fldPath, runtime, "'interface=' must be followed by a comma separated list of interface regular expressions"))
		}
		for _, r := range ifRegexes {
			_, e := regexp.Compile(r)
			if e != nil {
				validationError = append(validationError, field.Invalid(fldPath, runtime, fmt.Sprintf("regexp %s does not compile: %s", r, e.Error())))
			}
		}
		return validationError
	case "skip-interface":
		ifRegexes := regexp.MustCompile(`\s*,\s*`).Split(method[1], -1)
		if len(ifRegexes) == 0 || ifRegexes[0] == "" {
			validationError = append(validationError, field.Invalid(fldPath, runtime, "'skip-interface=' must be followed by a comma separated list of interface regular expressions"))
		}
		for _, r := range ifRegexes {
			_, e := regexp.Compile(r)
			if e != nil {
				validationError = append(validationError, field.Invalid(fldPath, runtime, fmt.Sprintf("regexp %s does not compile: %s", r, e.Error())))
			}
		}
		return validationError
	default:
		return field.ErrorList{field.Invalid(fldPath, runtime, "unsupported autodetection method")}
	}
}

func validateCalicoEncapsulationMode(mode string, fldPath *field.Path) field.ErrorList {
	valid := []string{"Always", "CrossSubnet", "Never"}

	allErrs := field.ErrorList{}
	allErrs = append(allErrs, IsValidValue(fldPath, &mode, valid)...)

	return allErrs
}

func validateContainerRuntime(c *kops.Cluster, runtime string, fldPath *field.Path) field.ErrorList {
	valid := []string{"containerd", "docker"}

	allErrs := field.ErrorList{}
	allErrs = append(allErrs, IsValidValue(fldPath, &runtime, valid)...)

	if runtime == "docker" {
		allErrs = append(allErrs, field.Forbidden(fldPath, "Docker CRI support was removed in Kubernetes 1.24: https://kubernetes.io/blog/2020/12/02/dockershim-faq"))
	}

	return allErrs
}

func validateContainerdConfig(cluster *kops.Cluster, config *kops.ContainerdConfig, fldPath *field.Path, inClusterConfig bool) field.ErrorList {
	allErrs := field.ErrorList{}

	if config.Version != nil {
		sv, err := semver.ParseTolerant(*config.Version)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("version"), config.Version,
				fmt.Sprintf("unable to parse version string: %s", err.Error())))
		}
		if sv.LT(semver.MustParse("1.3.4")) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("version"), config.Version,
				"unsupported legacy version"))
		}
	}

	if config.NRI != nil {
		allErrs = append(allErrs, validateNriConfig(config, fldPath.Child("nri"))...)
	}

	if config.Packages != nil {
		if config.Packages.UrlAmd64 != nil && config.Packages.HashAmd64 != nil {
			u := fi.ValueOf(config.Packages.UrlAmd64)
			_, err := url.Parse(u)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("packageUrl"), config.Packages.UrlAmd64,
					fmt.Sprintf("cannot parse package URL: %v", err)))
			}
			h := fi.ValueOf(config.Packages.HashAmd64)
			if len(h) > 64 {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("packageHash"), config.Packages.HashAmd64,
					"Package hash must be 64 characters long"))
			}
		} else if config.Packages.UrlAmd64 != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("packageUrl"), config.Packages.HashAmd64,
				"Package hash must also be set"))
		} else if config.Packages.HashAmd64 != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("packageHash"), config.Packages.HashAmd64,
				"Package URL must also be set"))
		}

		if config.Packages.UrlArm64 != nil && config.Packages.HashArm64 != nil {
			u := fi.ValueOf(config.Packages.UrlArm64)
			_, err := url.Parse(u)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("packageUrlArm64"), config.Packages.UrlArm64,
					fmt.Sprintf("cannot parse package URL: %v", err)))
			}
			h := fi.ValueOf(config.Packages.HashArm64)
			if len(h) > 64 {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("packageHashArm64"), config.Packages.HashArm64,
					"Package hash must be 64 characters long"))
			}
		} else if config.Packages.UrlArm64 != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("packageUrlArm64"), config.Packages.HashArm64,
				"Package hash must also be set"))
		} else if config.Packages.HashArm64 != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("packageHashArm64"), config.Packages.HashArm64,
				"Package URL must also be set"))
		}
	}

	if config.NvidiaGPU != nil {
		allErrs = append(allErrs, validateNvidiaConfig(cluster, config.NvidiaGPU, fldPath.Child("nvidia"), inClusterConfig)...)
	}

	return allErrs
}

func validateNriConfig(containerd *kops.ContainerdConfig, fldPath *field.Path) (allErrs field.ErrorList) {
	if containerd.NRI.Enabled == nil || !fi.ValueOf(containerd.NRI.Enabled) {
		return allErrs
	}
	v, err := semver.Parse(*containerd.Version)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("version"), containerd.Version,
			fmt.Sprintf("unable to parse version string: %s", err.Error())))
	}
	expectedRange, err := semver.ParseRange(">=1.7.0")
	if err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("version"), containerd.Version,
			fmt.Sprintf("unable to parse version range: %s", err.Error())))
	}
	if !expectedRange(v) {
		allErrs = append(allErrs, field.Forbidden(fldPath, "NRI is available starting from version 1.7.0 and above"))
	}
	return allErrs
}

func validateNvidiaConfig(cluster *kops.Cluster, nvidia *kops.NvidiaGPUConfig, fldPath *field.Path, inClusterConfig bool) (allErrs field.ErrorList) {
	if !fi.ValueOf(nvidia.Enabled) {
		return allErrs
	}
	if cluster.GetCloudProvider() != kops.CloudProviderAWS && cluster.GetCloudProvider() != kops.CloudProviderOpenstack {
		allErrs = append(allErrs, field.Forbidden(fldPath, "Nvidia is only supported on AWS and OpenStack"))
	}
	if cluster.GetCloudProvider() == kops.CloudProviderOpenstack && inClusterConfig {
		allErrs = append(allErrs, field.Forbidden(fldPath, "OpenStack supports nvidia configuration only in instance group"))
	}
	return allErrs
}

func validateRollingUpdate(rollingUpdate *kops.RollingUpdate, fldpath *field.Path, onControlPlaneInstanceGroup bool) field.ErrorList {
	allErrs := field.ErrorList{}
	var err error
	unavailable := 1
	if rollingUpdate.MaxUnavailable != nil {
		unavailable, err = intstr.GetScaledValueFromIntOrPercent(rollingUpdate.MaxUnavailable, 1, false)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fldpath.Child("maxUnavailable"), rollingUpdate.MaxUnavailable,
				fmt.Sprintf("Unable to parse: %v", err)))
		}
		if unavailable < 0 {
			allErrs = append(allErrs, field.Invalid(fldpath.Child("maxUnavailable"), rollingUpdate.MaxUnavailable, "Cannot be negative"))
		}
	}
	if rollingUpdate.MaxSurge != nil {
		surge, err := intstr.GetScaledValueFromIntOrPercent(rollingUpdate.MaxSurge, 1000, true)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fldpath.Child("maxSurge"), rollingUpdate.MaxSurge,
				fmt.Sprintf("Unable to parse: %v", err)))
		}
		if onControlPlaneInstanceGroup && surge != 0 {
			allErrs = append(allErrs, field.Forbidden(fldpath.Child("maxSurge"), "Cannot surge instance groups with role \"ControlPlane\""))
		} else if surge < 0 {
			allErrs = append(allErrs, field.Invalid(fldpath.Child("maxSurge"), rollingUpdate.MaxSurge, "Cannot be negative"))
		}
		if unavailable == 0 && surge == 0 {
			allErrs = append(allErrs, field.Forbidden(fldpath.Child("maxSurge"), "Cannot be zero if maxUnavailable is zero"))
		}
	}
	return allErrs
}

func validateNodeLocalDNS(spec *kops.ClusterSpec, fldpath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if spec.KubeDNS.NodeLocalDNS.LocalIP != "" {
		address := spec.KubeDNS.NodeLocalDNS.LocalIP
		ip := net.ParseIP(address)
		if ip == nil {
			allErrs = append(allErrs, field.Invalid(fldpath.Child("kubeDNS", "nodeLocalDNS", "localIP"), address, "Cluster had an invalid kubeDNS.nodeLocalDNS.localIP"))
		}
	}

	if (spec.KubeProxy != nil && spec.KubeProxy.ProxyMode == "ipvs") || spec.Networking.Cilium != nil {
		if spec.Kubelet != nil && spec.Kubelet.ClusterDNS != "" && spec.Kubelet.ClusterDNS != spec.KubeDNS.NodeLocalDNS.LocalIP {
			allErrs = append(allErrs, field.Forbidden(fldpath.Child("kubelet", "clusterDNS"), "Kubelet ClusterDNS must be set to the default IP address for LocalIP"))
		}

		if spec.ControlPlaneKubelet != nil && spec.ControlPlaneKubelet.ClusterDNS != "" && spec.ControlPlaneKubelet.ClusterDNS != spec.KubeDNS.NodeLocalDNS.LocalIP {
			allErrs = append(allErrs, field.Forbidden(fldpath.Child("kubelet", "clusterDNS"), "ControlPlaneKubelet ClusterDNS must be set to the default IP address for LocalIP"))
		}
	}

	return allErrs
}

func validateClusterAutoscaler(cluster *kops.Cluster, spec *kops.ClusterAutoscalerConfig, fldPath *field.Path) (allErrs field.ErrorList) {
	if spec.Expander != "" {
		allErrs = append(allErrs, IsValidValue(fldPath.Child("expander"), &spec.Expander, []string{"least-waste", "random", "most-pods", "price", "priority"})...)
	}

	if spec.Expander == "price" && cluster.Spec.CloudProvider.GCE == nil {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("expander"), "Cluster autoscaler price expander is only supported on GCE"))
	}

	if cluster.GetCloudProvider() == kops.CloudProviderOpenstack {
		allErrs = append(allErrs, field.Forbidden(fldPath, "Cluster autoscaler is not supported on OpenStack"))
	}

	return allErrs
}

func validateExternalDNS(cluster *kops.Cluster, spec *kops.ExternalDNSConfig, fldPath *field.Path) (allErrs field.ErrorList) {
	allErrs = append(allErrs, IsValidValue(fldPath.Child("provider"), &spec.Provider, []kops.ExternalDNSProvider{"", kops.ExternalDNSProviderDNSController, kops.ExternalDNSProviderExternalDNS, kops.ExternalDNSProviderNone})...)

	if spec.WatchNamespace != "" {
		if spec.WatchNamespace != "kube-system" {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("watchNamespace"), "externalDNS must watch either all namespaces or only kube-system"))
		}
	}

	if spec.Provider == kops.ExternalDNSProviderExternalDNS {
		if cluster.UsesLegacyGossip() || cluster.UsesNoneDNS() {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("provider"), "external-dns requires public or private DNS topology"))
		}
	}

	return allErrs
}

func validateMetricsServer(cluster *kops.Cluster, spec *kops.MetricsServerConfig, fldPath *field.Path) (allErrs field.ErrorList) {
	if spec != nil && fi.ValueOf(spec.Enabled) {
		if !fi.ValueOf(spec.Insecure) && !components.IsCertManagerEnabled(cluster) {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("insecure"), "Secure metrics server requires that cert manager is enabled"))
		}
	}

	return allErrs
}

func validateNodeTerminationHandler(cluster *kops.Cluster, spec *kops.NodeTerminationHandlerSpec, fldPath *field.Path) (allErrs field.ErrorList) {
	if spec.IsQueueMode() {
		if spec.EnableSpotInterruptionDraining != nil && !*spec.EnableSpotInterruptionDraining {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("enableSpotInterruptionDraining"), "spot interruption draining cannot be disabled in Queue Processor mode"))
		}
		if spec.EnableScheduledEventDraining != nil && !*spec.EnableScheduledEventDraining {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("enableScheduledEventDraining"), "scheduled event draining cannot be disabled in Queue Processor mode"))
		}
		if !fi.ValueOf(spec.EnableRebalanceDraining) && fi.ValueOf(spec.EnableRebalanceMonitoring) {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("enableRebalanceMonitoring"), "rebalance events can only drain in Queue Processor mode"))
		}
	}
	return allErrs
}

func validateAWSLoadBalancerController(cluster *kops.Cluster, spec *kops.LoadBalancerControllerSpec, fldPath *field.Path) (allErrs field.ErrorList) {
	if spec != nil && fi.ValueOf(spec.Enabled) {
		if !components.IsCertManagerEnabled(cluster) {
			allErrs = append(allErrs, field.Forbidden(fldPath, "AWS Load Balancer Controller requires that cert manager is enabled"))
		}
	}
	return allErrs
}

func validateCloudConfiguration(cloudConfig *kops.CloudConfiguration, spec *kops.ClusterSpec, fldPath *field.Path) (allErrs field.ErrorList) {
	if cloudConfig.ManageStorageClasses != nil && spec.CloudProvider.Openstack != nil &&
		spec.CloudProvider.Openstack.BlockStorage != nil && spec.CloudProvider.Openstack.BlockStorage.CreateStorageClass != nil {
		if *spec.CloudProvider.Openstack.BlockStorage.CreateStorageClass != *cloudConfig.ManageStorageClasses {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("manageStorageClasses"),
				"Management of storage classes and OpenStack block storage classes are both specified but disagree"))
		}
	}
	return allErrs
}

func validateWarmPool(warmPool *kops.WarmPoolSpec, fldPath *field.Path) (allErrs field.ErrorList) {
	if warmPool.MaxSize != nil {
		if *warmPool.MaxSize < 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("maxSize"), *warmPool.MaxSize, "warm pool maxSize cannot be negative"))
		} else if warmPool.MinSize > *warmPool.MaxSize {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("maxSize"), *warmPool.MaxSize, "warm pool maxSize cannot be set to lower than minSize"))
		}
	}
	if warmPool.MinSize < 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("minSize"), warmPool.MinSize, "warm pool minSize cannot be negative"))
	}
	return allErrs
}

func validateSnapshotController(cluster *kops.Cluster, spec *kops.SnapshotControllerConfig, fldPath *field.Path) (allErrs field.ErrorList) {
	if spec != nil && fi.ValueOf(spec.Enabled) {
		if !components.IsCertManagerEnabled(cluster) {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("enabled"), "Snapshot controller requires that cert manager is enabled"))
		}
	}
	return allErrs
}

func validatePodIdentityWebhook(cluster *kops.Cluster, spec *kops.PodIdentityWebhookSpec, fldPath *field.Path) (allErrs field.ErrorList) {
	if spec != nil && spec.Enabled {
		if !components.IsCertManagerEnabled(cluster) {
			allErrs = append(allErrs, field.Forbidden(fldPath, "EKS Pod Identity Webhook requires that cert manager is enabled"))
		}
	}

	return allErrs
}

func validateCertManager(cluster *kops.Cluster, spec *kops.CertManagerConfig, fldPath *field.Path) (allErrs field.ErrorList) {
	if len(spec.HostedZoneIDs) > 0 {
		if !fi.ValueOf(cluster.Spec.IAM.UseServiceAccountExternalPermissions) {
			allErrs = append(allErrs, field.Forbidden(fldPath, "Cert Manager requires that service accounts use external permissions in order to do dns-01 validation"))
		}
	}
	return allErrs
}
