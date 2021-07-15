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
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/blang/semver/v4"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/sets"
	utilvalidation "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/utils"
)

func newValidateCluster(cluster *kops.Cluster) field.ErrorList {
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

	allErrs = append(allErrs, validateClusterSpec(&cluster.Spec, cluster, field.NewPath("spec"))...)

	// Additional cloud-specific validation rules
	switch kops.CloudProviderID(cluster.Spec.CloudProvider) {
	case kops.CloudProviderAWS:
		allErrs = append(allErrs, awsValidateCluster(cluster)...)
	case kops.CloudProviderGCE:
		allErrs = append(allErrs, gceValidateCluster(cluster)...)
	case kops.CloudProviderOpenstack:
		allErrs = append(allErrs, openstackValidateCluster(cluster)...)
	}

	return allErrs
}

func validateClusterSpec(spec *kops.ClusterSpec, c *kops.Cluster, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, validateSubnets(spec, fieldPath.Child("subnets"))...)

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

	if spec.Topology != nil {
		allErrs = append(allErrs, validateTopology(spec.Topology, fieldPath.Child("topology"))...)
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
		allErrs = append(allErrs, validateKubeAPIServer(spec.KubeAPIServer, c, fieldPath.Child("kubeAPIServer"))...)
	}

	if spec.KubeProxy != nil {
		allErrs = append(allErrs, validateKubeProxy(spec.KubeProxy, fieldPath.Child("kubeProxy"))...)
	}

	if spec.Kubelet != nil {
		allErrs = append(allErrs, validateKubelet(spec.Kubelet, c, fieldPath.Child("kubelet"))...)
	}

	if spec.MasterKubelet != nil {
		allErrs = append(allErrs, validateKubelet(spec.MasterKubelet, c, fieldPath.Child("masterKubelet"))...)
	}

	if spec.Networking != nil {
		allErrs = append(allErrs, validateNetworking(c, spec.Networking, fieldPath.Child("networking"))...)
		if spec.Networking.Calico != nil {
			allErrs = append(allErrs, validateNetworkingCalico(spec.Networking.Calico, spec.EtcdClusters[0], fieldPath.Child("networking", "calico"))...)
		}
	}

	if spec.NodeAuthorization != nil {
		allErrs = append(allErrs, field.Forbidden(fieldPath.Child("nodeAuthorization"), "NodeAuthorization must be empty. The functionality has been reimplemented and is enabled on kubernetes >= 1.19.0."))
	}

	if spec.ClusterAutoscaler != nil {
		allErrs = append(allErrs, validateClusterAutoscaler(c, spec.ClusterAutoscaler, fieldPath.Child("clusterAutoscaler"))...)
	}

	if spec.ExternalDNS != nil {
		allErrs = append(allErrs, validateExternalDNS(c, spec.ExternalDNS, fieldPath.Child("externalDNS"))...)
	}

	if spec.NodeTerminationHandler != nil {
		allErrs = append(allErrs, validateNodeTerminationHandler(c, spec.NodeTerminationHandler, fieldPath.Child("nodeTerminationHandler"))...)
	}

	if spec.MetricsServer != nil {
		allErrs = append(allErrs, validateMetricsServer(c, spec.MetricsServer, fieldPath.Child("metricsServer"))...)

	}

	if spec.AWSLoadBalancerController != nil {
		allErrs = append(allErrs, validateAWSLoadBalancerController(c, spec.AWSLoadBalancerController, fieldPath.Child("awsLoadBalanceController"))...)
	}

	if spec.SnapshotController != nil {
		allErrs = append(allErrs, validateSnapshotController(c, spec.SnapshotController, fieldPath.Child("snapshotController"))...)

	}

	// IAM additional policies
	if spec.AdditionalPolicies != nil {
		for k, v := range *spec.AdditionalPolicies {
			allErrs = append(allErrs, validateAdditionalPolicy(k, v, fieldPath.Child("additionalPolicies"))...)
		}
	}
	// IAM external policies
	if spec.ExternalPolicies != nil {
		for k, v := range *spec.ExternalPolicies {
			allErrs = append(allErrs, validateExternalPolicies(k, v, fieldPath.Child("externalPolicies"))...)
		}
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
			allErrs = append(allErrs, validateEtcdTLS(spec.EtcdClusters, fieldEtcdClusters)...)
			allErrs = append(allErrs, validateEtcdStorage(spec.EtcdClusters, fieldEtcdClusters)...)
		}
	}

	if spec.ContainerRuntime != "" {
		allErrs = append(allErrs, validateContainerRuntime(&spec.ContainerRuntime, fieldPath.Child("containerRuntime"))...)
	}

	if spec.Containerd != nil {
		allErrs = append(allErrs, validateContainerdConfig(spec.Containerd, fieldPath.Child("containerd"))...)
	}

	if spec.Docker != nil {
		allErrs = append(allErrs, validateDockerConfig(spec.Docker, fieldPath.Child("docker"))...)
	}

	if spec.Assets != nil {
		if spec.Assets.ContainerProxy != nil && spec.Assets.ContainerRegistry != nil {
			allErrs = append(allErrs, field.Forbidden(fieldPath.Child("assets", "containerProxy"), "containerProxy cannot be used in conjunction with containerRegistry"))
		}
	}

	if spec.IAM == nil || spec.IAM.Legacy {
		allErrs = append(allErrs, field.Forbidden(fieldPath.Child("iam", "legacy"), "legacy IAM permissions are no longer supported"))
	}

	if spec.RollingUpdate != nil {
		allErrs = append(allErrs, validateRollingUpdate(spec.RollingUpdate, fieldPath.Child("rollingUpdate"), false)...)
	}

	if spec.API != nil && spec.API.LoadBalancer != nil && spec.CloudProvider == "aws" {
		value := string(spec.API.LoadBalancer.Class)
		allErrs = append(allErrs, IsValidValue(fieldPath.Child("class"), &value, kops.SupportedLoadBalancerClasses)...)
		if spec.API.LoadBalancer.SSLCertificate != "" && spec.API.LoadBalancer.Class != kops.LoadBalancerClassNetwork && c.IsKubernetesGTE("1.19") {
			allErrs = append(allErrs, field.Forbidden(fieldPath, "sslCertificate requires network loadbalancer for K8s 1.19+ see https://github.com/kubernetes/kops/blob/master/permalinks/acm_nlb.md"))
		}
		if spec.API.LoadBalancer.Class == kops.LoadBalancerClassNetwork && spec.API.LoadBalancer.UseForInternalApi && spec.API.LoadBalancer.Type == kops.LoadBalancerTypeInternal {
			allErrs = append(allErrs, field.Forbidden(fieldPath, "useForInternalApi cannot be used with internal NLB due lack of hairpinning support"))
		}
	}

	if spec.CloudConfig != nil {
		allErrs = append(allErrs, validateCloudConfiguration(spec.CloudConfig, fieldPath.Child("cloudConfig"))...)
	}

	if spec.WarmPool != nil {
		if kops.CloudProviderID(spec.CloudProvider) != kops.CloudProviderAWS {
			allErrs = append(allErrs, field.Forbidden(field.NewPath("spec", "warmPool"), "warm pool only supported on AWS"))
		} else {
			allErrs = append(allErrs, validateWarmPool(spec.WarmPool, fieldPath.Child("warmPool"))...)
		}
	}

	if spec.IAM != nil {
		if len(spec.IAM.ServiceAccountExternalPermissions) > 0 {
			if spec.ServiceAccountIssuerDiscovery == nil || !spec.ServiceAccountIssuerDiscovery.EnableAWSOIDCProvider {
				allErrs = append(allErrs, field.Forbidden(fieldPath.Child("iam", "serviceAccountExternalPermissions"), "serviceAccountExternalPermissions requires AWS OIDC Provider to be enabled"))
			}
			allErrs = append(allErrs, validateSAExternalPermissions(spec.IAM.ServiceAccountExternalPermissions, fieldPath.Child("iam", "serviceAccountExternalPermissions"))...)
		}
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

func validateCIDR(cidr string, fieldPath *field.Path) field.ErrorList {
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
	return allErrs
}

func validateIPv6CIDR(cidr string, fieldPath *field.Path) field.ErrorList {
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
		allErrs = append(allErrs, validateCIDR(cidr, fieldPath)...)

		if !utils.IsIPv6CIDR(cidr) {
			allErrs = append(allErrs, field.Invalid(fieldPath, cidr, "Network is not an IPv6 CIDR"))
		}
	}

	return allErrs
}

func validateTopology(topology *kops.TopologySpec, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if topology.Masters == "" {
		allErrs = append(allErrs, field.Required(fieldPath.Child("masters"), ""))
	} else {
		allErrs = append(allErrs, IsValidValue(fieldPath.Child("masters"), &topology.Masters, kops.SupportedTopologies)...)
	}

	if topology.Nodes == "" {
		allErrs = append(allErrs, field.Required(fieldPath.Child("nodes"), ""))
	} else {
		allErrs = append(allErrs, IsValidValue(fieldPath.Child("nodes"), &topology.Nodes, kops.SupportedTopologies)...)
	}

	if topology.Bastion != nil {
		bastion := topology.Bastion
		if topology.Masters == kops.TopologyPublic || topology.Nodes == kops.TopologyPublic {
			allErrs = append(allErrs, field.Forbidden(fieldPath.Child("bastion"), "bastion requires masters and nodes to have private topology"))
		}
		if bastion.IdleTimeoutSeconds != nil && *bastion.IdleTimeoutSeconds <= 0 {
			allErrs = append(allErrs, field.Invalid(fieldPath.Child("bastion", "idleTimeoutSeconds"), *bastion.IdleTimeoutSeconds, "bastion idleTimeoutSeconds should be greater than zero"))
		}
		if bastion.IdleTimeoutSeconds != nil && *bastion.IdleTimeoutSeconds > 3600 {
			allErrs = append(allErrs, field.Invalid(fieldPath.Child("bastion", "idleTimeoutSeconds"), *bastion.IdleTimeoutSeconds, "bastion idleTimeoutSeconds cannot be greater than one hour"))
		}
	}

	if topology.DNS != nil {
		value := string(topology.DNS.Type)
		allErrs = append(allErrs, IsValidValue(fieldPath.Child("dns", "type"), &value, kops.SupportedDnsTypes)...)
	}

	return allErrs
}

func validateSubnets(cluster *kops.ClusterSpec, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	subnets := cluster.Subnets

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

	if kops.CloudProviderID(cluster.CloudProvider) != kops.CloudProviderAWS {
		for i := range subnets {
			if subnets[i].IPv6CIDR != "" {
				allErrs = append(allErrs, field.Forbidden(fieldPath.Index(i).Child("ipv6CIDR"), "ipv6CIDR can only be specified for AWS"))
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
	// IPv6CIDR
	if subnet.IPv6CIDR != "" {
		allErrs = append(allErrs, validateIPv6CIDR(subnet.IPv6CIDR, fieldPath.Child("ipv6CIDR"))...)
	}

	if subnet.Egress != "" {
		egressType := strings.Split(subnet.Egress, "-")[0]
		if egressType != kops.EgressNatGateway && egressType != kops.EgressElasticIP && egressType != kops.EgressNatInstance && egressType != kops.EgressExternal && egressType != kops.EgressTransitGateway {
			allErrs = append(allErrs, field.Invalid(fieldPath.Child("egress"), subnet.Egress,
				"egress must be of type NAT Gateway, NAT Gateway with existing ElasticIP, NAT EC2 Instance, Transit Gateway, or External"))
		}
		if subnet.Egress != kops.EgressExternal && subnet.Type != "Private" {
			allErrs = append(allErrs, field.Forbidden(fieldPath.Child("egress"), "egress can only be specified for private subnets"))
		}
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

func validateKubeAPIServer(v *kops.KubeAPIServerConfig, c *kops.Cluster, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(v.AdmissionControl) > 0 {
		if len(v.DisableAdmissionPlugins) > 0 {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("disableAdmissionPlugins"),
				"disableAdmissionPlugins is mutually exclusive, you cannot use both admissionControl and disableAdmissionPlugins together"))
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
			if kops.CloudProviderID(c.Spec.CloudProvider) == kops.CloudProviderAWS && c.IsKubernetesGTE("1.19") {
				if !hasNode || !hasRBAC {
					allErrs = append(allErrs, field.Required(fldPath.Child("authorizationMode"), "As of kubernetes 1.19 on AWS, authorizationMode must include RBAC and Node"))
				}
			}
		}
	}

	if v.LogFormat != "" {
		allErrs = append(allErrs, IsValidValue(fldPath.Child("logFormat"), &v.LogFormat, []string{"text", "json"})...)
	}

	return allErrs
}

func validateKubeProxy(k *kops.KubeProxyConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	master := k.Master

	for i, x := range k.IPVSExcludeCIDRS {
		if _, _, err := net.ParseCIDR(x); err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("ipvsExcludeCidrs").Index(i), x, "Invalid network CIDR"))
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

		if k.BootstrapKubeconfig != "" {
			if c.Spec.KubeAPIServer == nil {
				allErrs = append(allErrs, field.Required(kubeletPath.Root().Child("spec").Child("kubeAPIServer"), "bootstrap token require the NodeRestriction admissions controller"))
			}
		}

		if k.TopologyManagerPolicy != "" {
			allErrs = append(allErrs, IsValidValue(kubeletPath.Child("topologyManagerPolicy"), &k.TopologyManagerPolicy, []string{"none", "best-effort", "restricted", "single-numa-node"})...)
			if !c.IsKubernetesGTE("1.18") {
				allErrs = append(allErrs, field.Forbidden(kubeletPath.Child("topologyManagerPolicy"), "topologyManagerPolicy requires at least Kubernetes 1.18"))
			}
		}

		if k.EnableCadvisorJsonEndpoints != nil {
			if c.IsKubernetesLT("1.18") && c.IsKubernetesGTE("1.21") {
				allErrs = append(allErrs, field.Forbidden(kubeletPath.Child("enableCadvisorJsonEndpoints"), "enableCadvisorJsonEndpoints requires Kubernetes 1.18-1.20"))
			}
		}

		if k.LogFormat != "" {
			allErrs = append(allErrs, IsValidValue(kubeletPath.Child("logFormat"), &k.LogFormat, []string{"text", "json"})...)
		}

	}
	return allErrs
}

func validateNetworking(cluster *kops.Cluster, v *kops.NetworkingSpec, fldPath *field.Path) field.ErrorList {
	c := &cluster.Spec
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

	if v.Kopeio != nil {
		if optionTaken {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("kopeio"), "only one networking option permitted"))
		}
		optionTaken = true
	}

	if v.CNI != nil && optionTaken {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("cni"), "only one networking option permitted"))
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
		if c.KubeProxy != nil && (c.KubeProxy.Enabled == nil || *c.KubeProxy.Enabled) {
			allErrs = append(allErrs, field.Forbidden(fldPath.Root().Child("spec", "kubeProxy", "enabled"), "kube-router requires kubeProxy to be disabled"))
		}
		optionTaken = true
	}

	if v.Romana != nil {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("romana"), "support for Romana has been removed"))
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

		allErrs = append(allErrs, validateNetworkingCilium(cluster, v.Cilium, fldPath.Child("cilium"))...)
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

		if !(version.Minor >= 8 && version.Minor <= 10) {
			allErrs = append(allErrs, field.Invalid(versionFld, v.Version, "Only versions 1.8 through 1.10 are supported"))
		}

		if version.Minor < 10 && c.IsIPv6Only() {
			allErrs = append(allErrs, field.Invalid(versionFld, v.Version, "kOps only supports IPv6 on version 1.10 or later"))
		}

		if v.Hubble != nil && fi.BoolValue(v.Hubble.Enabled) {
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

	if v.ContainerRuntimeLabels != "" {
		allErrs = append(allErrs, IsValidValue(fldPath.Child("containerRuntimeLabels"), &v.ContainerRuntimeLabels, []string{"none", "containerd", "crio", "docker", "auto"})...)
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

	if fi.BoolValue(v.EnableL7Proxy) && v.IPTablesRulesNoinstall {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("enableL7Proxy"), "Cilium L7 Proxy requires IPTablesRules to be installed"))
	}

	if v.Ipam != "" {
		// "azure" not supported by kops
		allErrs = append(allErrs, IsValidValue(fldPath.Child("ipam"), &v.Ipam, []string{"hostscope", "kubernetes", "crd", "eni"})...)

		if v.Ipam == kops.CiliumIpamEni {
			if c.CloudProvider != string(kops.CloudProviderAWS) {
				allErrs = append(allErrs, field.Forbidden(fldPath.Child("ipam"), "Cilum ENI IPAM is supported only in AWS"))
			}
			if v.DisableMasquerade != nil && !*v.DisableMasquerade {
				allErrs = append(allErrs, field.Forbidden(fldPath.Child("disableMasquerade"), "Masquerade must be disabled when ENI IPAM is used"))
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
				if cluster.Provider == kops.EtcdProviderTypeLegacy {
					allErrs = append(allErrs, field.Invalid(fldPath.Root().Child("etcdClusters"), kops.EtcdProviderTypeLegacy, "Legacy etcd provider is not supported for the cilium cluster"))
				}
				hasCiliumCluster = true
				break
			}
		}
		if !hasCiliumCluster {
			allErrs = append(allErrs, field.Required(fldPath.Root().Child("etcdClusters"), "Cilium with managed etcd requires a dedicated etcd cluster"))
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
	allErrs := field.ErrorList{}

	var valid []string
	for _, r := range kops.AllInstanceGroupRoles {
		valid = append(valid, strings.ToLower(string(r)))
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
			value := string(statement.Effect)
			allErrs = append(allErrs, IsValidValue(fldEffect, &value, []string{"Allow", "Deny"})...)
		}
	}

	return allErrs
}

func validateExternalPolicies(role string, policies []string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	var valid []string
	for _, r := range kops.AllInstanceGroupRoles {
		valid = append(valid, strings.ToLower(string(r)))
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

	if spec.Name == "" {
		allErrs = append(allErrs, field.Required(fieldPath.Child("name"), "etcdCluster did not have name"))
	}
	if spec.Provider != "" {
		value := string(spec.Provider)
		allErrs = append(allErrs, IsValidValue(fieldPath.Child("provider"), &value, kops.SupportedEtcdProviderTypes)...)
		if spec.Provider == kops.EtcdProviderTypeLegacy && c.IsKubernetesGTE("1.18") {
			allErrs = append(allErrs, field.Forbidden(fieldPath.Child("provider"), "support for Legacy mode removed as of Kubernetes 1.18"))
		}
	}
	if len(spec.Members) == 0 {
		allErrs = append(allErrs, field.Required(fieldPath.Child("etcdMembers"), "No members defined in etcd cluster"))
	} else if (len(spec.Members) % 2) == 0 {
		// Not technically a requirement, but doesn't really make sense to allow
		allErrs = append(allErrs, field.Invalid(fieldPath.Child("etcdMembers"), len(spec.Members), "Should be an odd number of master-zones for quorum. Use --zones and --master-zones to declare node zones and master zones separately"))
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

// validateEtcdTLS checks the TLS settings for etcd are valid
func validateEtcdTLS(specs []kops.EtcdClusterSpec, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	var usingTLS int
	for _, x := range specs {
		if x.EnableEtcdTLS {
			usingTLS++
		}
	}
	// check both clusters are using tls if one is enabled
	if usingTLS > 0 && usingTLS != len(specs) {
		allErrs = append(allErrs, field.Forbidden(fieldPath.Index(0).Child("enableEtcdTLS"), "both etcd clusters must have TLS enabled or none at all"))
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
// @TODO semvar package doesn't appear to ignore a 'v' in v1.1.1; could be a problem later down the line
func validateEtcdVersion(spec kops.EtcdClusterSpec, fieldPath *field.Path, minimalVersion *semver.Version) field.ErrorList {
	// @check if the storage is specified that it's valid

	if minimalVersion == nil {
		v := semver.MustParse("0.0.0")
		minimalVersion = &v
	}

	version := spec.Version
	if spec.Version == "" {
		version = components.DefaultEtcd3Version_1_17
	}

	sem, err := semver.Parse(strings.TrimPrefix(version, "v"))
	if err != nil {
		return field.ErrorList{field.Invalid(fieldPath.Child("version"), version, "the storage version is invalid")}
	}

	// we only support v3 for now
	if sem.Major == 3 {
		if sem.LT(*minimalVersion) {
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

	if fi.StringValue(spec.InstanceGroup) == "" {
		allErrs = append(allErrs, field.Required(fieldPath.Child("instanceGroup"), "etcdMember did not have instanceGroup"))
	}

	return allErrs
}

func validateNetworkingCalico(v *kops.CalicoNetworkingSpec, e kops.EtcdClusterSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if v.AWSSrcDstCheck != "" {
		valid := []string{"Enable", "Disable", "DoNothing"}
		allErrs = append(allErrs, IsValidValue(fldPath.Child("awsSrcDstCheck"), &v.AWSSrcDstCheck, valid)...)
	}

	if v.CrossSubnet != nil {
		if fi.BoolValue(v.CrossSubnet) && v.AWSSrcDstCheck != "Disable" {
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
		// Don't tolerate "None" for now, which would disable encapsulation in the default IPPool
		// object. Note that with no encapsulation, we'd need to select the "bird" networking
		// backend in order to allow use of BGP to distribute routes for pod traffic.
		valid := []string{"ipip", "vxlan"}
		allErrs = append(allErrs, IsValidValue(fldPath.Child("encapsulationMode"), &v.EncapsulationMode, valid)...)
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

func validateContainerRuntime(runtime *string, fldPath *field.Path) field.ErrorList {
	valid := []string{"containerd", "docker"}

	allErrs := field.ErrorList{}
	allErrs = append(allErrs, IsValidValue(fldPath, runtime, valid)...)

	return allErrs
}

func validateContainerdConfig(config *kops.ContainerdConfig, fldPath *field.Path) field.ErrorList {
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

	if config.Packages != nil {
		if config.Packages.UrlAmd64 != nil && config.Packages.HashAmd64 != nil {
			u := fi.StringValue(config.Packages.UrlAmd64)
			_, err := url.Parse(u)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("packageUrl"), config.Packages.UrlAmd64,
					fmt.Sprintf("cannot parse package URL: %v", err)))
			}
			h := fi.StringValue(config.Packages.HashAmd64)
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
			u := fi.StringValue(config.Packages.UrlArm64)
			_, err := url.Parse(u)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("packageUrlArm64"), config.Packages.UrlArm64,
					fmt.Sprintf("cannot parse package URL: %v", err)))
			}
			h := fi.StringValue(config.Packages.HashArm64)
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

	return allErrs
}

func validateDockerConfig(config *kops.DockerConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if config.Version != nil {
		sv, err := semver.ParseTolerant(*config.Version)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("version"), config.Version,
				fmt.Sprintf("unable to parse version string: %s", err.Error())))
		}
		if sv.LT(semver.MustParse("1.14.0")) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("version"), config.Version,
				"version is no longer available: https://www.docker.com/blog/changes-dockerproject-org-apt-yum-repositories"))
		} else if sv.LT(semver.MustParse("17.3.0")) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("version"), config.Version,
				"unsupported legacy version"))
		}
	}

	if config.Packages != nil {
		if config.Packages.UrlAmd64 != nil && config.Packages.HashAmd64 != nil {
			u := fi.StringValue(config.Packages.UrlAmd64)
			_, err := url.Parse(u)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("packageUrl"), config.Packages.UrlAmd64,
					fmt.Sprintf("unable parse package URL string: %v", err)))
			}
			h := fi.StringValue(config.Packages.HashAmd64)
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
			u := fi.StringValue(config.Packages.UrlArm64)
			_, err := url.Parse(u)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("packageUrlArm64"), config.Packages.UrlArm64,
					fmt.Sprintf("unable parse package URL string: %v", err)))
			}
			h := fi.StringValue(config.Packages.HashArm64)
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

	if config.Storage != nil {
		valid := []string{"aufs", "btrfs", "devicemapper", "overlay", "overlay2", "zfs"}
		values := strings.Split(*config.Storage, ",")
		for _, value := range values {
			allErrs = append(allErrs, IsValidValue(fldPath.Child("storage"), &value, valid)...)
		}
	}

	return allErrs
}

func validateRollingUpdate(rollingUpdate *kops.RollingUpdate, fldpath *field.Path, onMasterInstanceGroup bool) field.ErrorList {
	allErrs := field.ErrorList{}
	var err error
	unavailable := 1
	if rollingUpdate.MaxUnavailable != nil {
		unavailable, err = intstr.GetValueFromIntOrPercent(rollingUpdate.MaxUnavailable, 1, false)
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

	if (spec.KubeProxy != nil && spec.KubeProxy.ProxyMode == "ipvs") || (spec.Networking != nil && spec.Networking.Cilium != nil) {
		if spec.Kubelet != nil && spec.Kubelet.ClusterDNS != "" && spec.Kubelet.ClusterDNS != spec.KubeDNS.NodeLocalDNS.LocalIP {
			allErrs = append(allErrs, field.Forbidden(fldpath.Child("kubelet", "clusterDNS"), "Kubelet ClusterDNS must be set to the default IP address for LocalIP"))
		}

		if spec.MasterKubelet != nil && spec.MasterKubelet.ClusterDNS != "" && spec.MasterKubelet.ClusterDNS != spec.KubeDNS.NodeLocalDNS.LocalIP {
			allErrs = append(allErrs, field.Forbidden(fldpath.Child("kubelet", "clusterDNS"), "MasterKubelet ClusterDNS must be set to the default IP address for LocalIP"))
		}
	}

	return allErrs
}

func validateClusterAutoscaler(cluster *kops.Cluster, spec *kops.ClusterAutoscalerConfig, fldPath *field.Path) (allErrs field.ErrorList) {
	allErrs = append(allErrs, IsValidValue(fldPath.Child("expander"), spec.Expander, []string{"least-waste", "random", "most-pods"})...)

	if kops.CloudProviderID(cluster.Spec.CloudProvider) == kops.CloudProviderOpenstack {
		allErrs = append(allErrs, field.Forbidden(fldPath, "Cluster autoscaler is not supported on OpenStack"))
	}

	return allErrs
}

func validateExternalDNS(cluster *kops.Cluster, spec *kops.ExternalDNSConfig, fldPath *field.Path) (allErrs field.ErrorList) {
	allErrs = append(allErrs, IsValidValue(fldPath.Child("provider"), (*string)(&spec.Provider), []string{"", "dns-controller", "external-dns"})...)

	if spec.WatchNamespace != "" {
		if spec.WatchNamespace != "kube-system" {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("watchNamespace"), "externalDNS must watch either all namespaces or only kube-system"))
		}
	}

	if spec.Provider == kops.ExternalDNSProviderExternalDNS {
		if dns.IsGossipHostname(cluster.Spec.MasterInternalName) {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("provider"), "external-dns does not supprot gossip clusters"))
		}
	}

	return allErrs

}

func validateNodeTerminationHandler(cluster *kops.Cluster, spec *kops.NodeTerminationHandlerConfig, fldPath *field.Path) (allErrs field.ErrorList) {
	if kops.CloudProviderID(cluster.Spec.CloudProvider) != kops.CloudProviderAWS {
		allErrs = append(allErrs, field.Forbidden(fldPath, "Node Termination Handler supports only AWS"))
	}
	return allErrs
}

func validateMetricsServer(cluster *kops.Cluster, spec *kops.MetricsServerConfig, fldPath *field.Path) (allErrs field.ErrorList) {
	if spec != nil && fi.BoolValue(spec.Enabled) {
		if !fi.BoolValue(spec.Insecure) && !components.IsCertManagerEnabled(cluster) {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("insecure"), "Secure metrics server requires that cert manager is enabled"))
		}
	}

	return allErrs
}

func validateAWSLoadBalancerController(cluster *kops.Cluster, spec *kops.AWSLoadBalancerControllerConfig, fldPath *field.Path) (allErrs field.ErrorList) {
	if spec != nil && fi.BoolValue(spec.Enabled) {
		if !components.IsCertManagerEnabled(cluster) {
			allErrs = append(allErrs, field.Forbidden(fldPath, "AWS Load Balancer Controller requires that cert manager is enabled"))
		}
	}
	return allErrs
}

func validateCloudConfiguration(cloudConfig *kops.CloudConfiguration, fldPath *field.Path) (allErrs field.ErrorList) {
	if cloudConfig.ManageStorageClasses != nil && cloudConfig.Openstack != nil &&
		cloudConfig.Openstack.BlockStorage != nil && cloudConfig.Openstack.BlockStorage.CreateStorageClass != nil {
		if *cloudConfig.Openstack.BlockStorage.CreateStorageClass != *cloudConfig.ManageStorageClasses {
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
	if spec != nil && fi.BoolValue(spec.Enabled) {
		if !cluster.IsKubernetesGTE("1.20") {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("enabled"), "Snapshot controller requires kubernetes 1.20+"))
		}
		if !components.IsCertManagerEnabled(cluster) {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("enabled"), "Snapshot controller requires that cert manager is enabled"))
		}
		if !hasAWSEBSCSIDriver(cluster.Spec) {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("enabled"), "Snapshot controller requires external CSI Driver"))
		}
	}
	return allErrs
}
