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

	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/util/subnet"
	"k8s.io/kops/upup/pkg/fi"
)

// legacy contains validation functions that don't match the apimachinery style

// ValidateCluster is responsible for checking the validity of the Cluster spec
func ValidateCluster(c *kops.Cluster, strict bool) field.ErrorList {
	fieldSpec := field.NewPath("spec")
	allErrs := field.ErrorList{}

	// KubernetesVersion
	// This is one case we return the error because a large part of the rest of the validation logic depends on a valid kubernetes version.

	if c.Spec.KubernetesVersion == "" {
		allErrs = append(allErrs, field.Required(fieldSpec.Child("kubernetesVersion"), ""))
		return allErrs
	} else if _, err := util.ParseKubernetesVersion(c.Spec.KubernetesVersion); err != nil {
		allErrs = append(allErrs, field.Invalid(fieldSpec.Child("kubernetesVersion"), c.Spec.KubernetesVersion, "unable to determine kubernetes version"))
		return allErrs
	}

	if c.ObjectMeta.Name == "" {
		allErrs = append(allErrs, field.Required(field.NewPath("objectMeta", "name"), "Cluster Name is required (e.g. --name=mycluster.myzone.com)"))
	} else {
		// Must be a dns name
		errs := validation.IsDNS1123Subdomain(c.ObjectMeta.Name)
		if len(errs) != 0 {
			allErrs = append(allErrs, field.Invalid(field.NewPath("objectMeta", "name"), c.ObjectMeta.Name, fmt.Sprintf("Cluster Name must be a valid DNS name (e.g. --name=mycluster.myzone.com) errors: %s", strings.Join(errs, ", "))))
		} else if !strings.Contains(c.ObjectMeta.Name, ".") {
			// Tolerate if this is a cluster we are importing for upgrade
			if c.ObjectMeta.Annotations[kops.AnnotationNameManagement] != kops.AnnotationValueManagementImported {
				allErrs = append(allErrs, field.Invalid(field.NewPath("objectMeta", "name"), c.ObjectMeta.Name, "Cluster Name must be a fully-qualified DNS name (e.g. --name=mycluster.myzone.com)"))
			}
		}
	}

	if c.Spec.Assets != nil && c.Spec.Assets.ContainerProxy != nil && c.Spec.Assets.ContainerRegistry != nil {
		allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("Assets", "ContainerProxy"), "ContainerProxy cannot be used in conjunction with ContainerRegistry as represent mutually exclusive concepts. Please consult the documentation for details."))
	}

	requiresSubnets := true
	requiresNetworkCIDR := true
	requiresSubnetCIDR := true
	switch kops.CloudProviderID(c.Spec.CloudProvider) {
	case "":
		allErrs = append(allErrs, field.Required(fieldSpec.Child("cloudProvider"), ""))
		requiresSubnets = false
		requiresSubnetCIDR = false
		requiresNetworkCIDR = false

	case kops.CloudProviderBareMetal:
		requiresSubnets = false
		requiresSubnetCIDR = false
		requiresNetworkCIDR = false
		if c.Spec.NetworkCIDR != "" {
			allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("networkCIDR"), "networkCIDR should not be set on bare metal"))
		}

	case kops.CloudProviderGCE:
		requiresNetworkCIDR = false
		if c.Spec.NetworkCIDR != "" {
			allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("networkCIDR"), "networkCIDR should not be set on GCE"))
		}
		requiresSubnetCIDR = false

	case kops.CloudProviderDO:
		requiresSubnets = false
		requiresSubnetCIDR = false
		requiresNetworkCIDR = false
		if c.Spec.NetworkCIDR != "" {
			allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("networkCIDR"), "networkCIDR should not be set on DigitalOcean"))
		}
	case kops.CloudProviderALI:
		requiresSubnets = false
		requiresSubnetCIDR = false
		requiresNetworkCIDR = false
	case kops.CloudProviderAWS:
	case kops.CloudProviderVSphere:
	case kops.CloudProviderOpenstack:
		requiresNetworkCIDR = false
		requiresSubnetCIDR = false

	default:
		allErrs = append(allErrs, field.NotSupported(fieldSpec.Child("cloudProvider"), c.Spec.CloudProvider, []string{
			string(kops.CloudProviderBareMetal),
			string(kops.CloudProviderGCE),
			string(kops.CloudProviderDO),
			string(kops.CloudProviderALI),
			string(kops.CloudProviderAWS),
			string(kops.CloudProviderVSphere),
			string(kops.CloudProviderOpenstack),
		}))
	}

	if requiresSubnets && len(c.Spec.Subnets) == 0 {
		// TODO: Auto choose zones from region?
		allErrs = append(allErrs, field.Required(fieldSpec.Child("subnets"), "must configure at least one subnet (use --zones)"))
	}

	if strict && c.Spec.Kubelet == nil {
		allErrs = append(allErrs, field.Required(fieldSpec.Child("kubelet"), "kubelet not configured"))
	}
	if strict && c.Spec.MasterKubelet == nil {
		allErrs = append(allErrs, field.Required(fieldSpec.Child("masterKubelet"), "masterKubelet not configured"))
	}
	if strict && c.Spec.KubeControllerManager == nil {
		allErrs = append(allErrs, field.Required(fieldSpec.Child("kubeControllerManager"), "kubeControllerManager not configured"))
	}
	if strict && c.Spec.KubeDNS == nil {
		allErrs = append(allErrs, field.Required(fieldSpec.Child("kubeDNS"), "kubeDNS not configured"))
	}
	if strict && c.Spec.KubeScheduler == nil {
		allErrs = append(allErrs, field.Required(fieldSpec.Child("kubeScheduler"), "kubeScheduler not configured"))
	}
	if strict && c.Spec.KubeAPIServer == nil {
		allErrs = append(allErrs, field.Required(fieldSpec.Child("kubeAPIServer"), "kubeAPIServer not configured"))
	}
	if strict && c.Spec.KubeProxy == nil {
		allErrs = append(allErrs, field.Required(fieldSpec.Child("kubeProxy"), "kubeProxy not configured"))
	}
	if strict && c.Spec.Docker == nil {
		allErrs = append(allErrs, field.Required(fieldSpec.Child("docker"), "docker not configured"))
	}

	// Check NetworkCIDR
	var networkCIDR *net.IPNet
	var err error
	{
		if c.Spec.NetworkCIDR == "" {
			if requiresNetworkCIDR {
				allErrs = append(allErrs, field.Required(fieldSpec.Child("networkCIDR"), "Cluster did not have networkCIDR set"))
			}
		} else {
			_, networkCIDR, err = net.ParseCIDR(c.Spec.NetworkCIDR)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(fieldSpec.Child("networkCIDR"), c.Spec.NetworkCIDR, fmt.Sprintf("Cluster had an invalid networkCIDR")))
			}
		}
	}

	// Check AdditionalNetworkCIDRs
	var additionalNetworkCIDRs []*net.IPNet
	{
		if len(c.Spec.AdditionalNetworkCIDRs) > 0 {
			for _, AdditionalNetworkCIDR := range c.Spec.AdditionalNetworkCIDRs {
				_, IPNetAdditionalNetworkCIDR, err := net.ParseCIDR(AdditionalNetworkCIDR)
				if err != nil {
					allErrs = append(allErrs, field.Invalid(fieldSpec.Child("additionalNetworkCIDRs"), AdditionalNetworkCIDR, fmt.Sprintf("Cluster had an invalid additionalNetworkCIDRs")))
				}
				additionalNetworkCIDRs = append(additionalNetworkCIDRs, IPNetAdditionalNetworkCIDR)
			}
		}
	}

	// nonMasqueradeCIDR is essentially deprecated, and we're moving to cluster-cidr instead (which is better named pod-cidr)
	nonMasqueradeCIDRRequired := true
	serviceClusterMustBeSubnetOfNonMasqueradeCIDR := true
	if c.Spec.Networking != nil && c.Spec.Networking.GCE != nil {
		nonMasqueradeCIDRRequired = false
		serviceClusterMustBeSubnetOfNonMasqueradeCIDR = false
	}

	// Check NonMasqueradeCIDR
	var nonMasqueradeCIDR *net.IPNet
	{
		nonMasqueradeCIDRString := c.Spec.NonMasqueradeCIDR
		if nonMasqueradeCIDRString == "" {
			if nonMasqueradeCIDRRequired {
				allErrs = append(allErrs, field.Required(fieldSpec.Child("nonMasqueradeCIDR"), "Cluster did not have nonMasqueradeCIDR set"))
			}
		} else {
			_, nonMasqueradeCIDR, err = net.ParseCIDR(nonMasqueradeCIDRString)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(fieldSpec.Child("nonMasqueradeCIDR"), nonMasqueradeCIDRString, "Cluster had an invalid nonMasqueradeCIDR"))
			}

			if networkCIDR != nil && subnet.Overlap(nonMasqueradeCIDR, networkCIDR) && c.Spec.Networking != nil && c.Spec.Networking.AmazonVPC == nil && c.Spec.Networking.LyftVPC == nil && (c.Spec.Networking.Cilium == nil || c.Spec.Networking.Cilium.Ipam != kops.CiliumIpamEni) {
				allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("nonMasqueradeCIDR"), fmt.Sprintf("nonMasqueradeCIDR %q cannot overlap with networkCIDR %q", nonMasqueradeCIDRString, c.Spec.NetworkCIDR)))
			}

			if c.Spec.Kubelet != nil && c.Spec.Kubelet.NonMasqueradeCIDR != nonMasqueradeCIDRString {
				// TODO Remove the Spec.Kubelet.NonMasqueradeCIDR field?
				if strict || c.Spec.Kubelet.NonMasqueradeCIDR != "" {
					allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("kubelet", "nonMasqueradeCIDR"), "kubelet nonMasqueradeCIDR did not match cluster nonMasqueradeCIDR"))
				}
			}
			if c.Spec.MasterKubelet != nil && c.Spec.MasterKubelet.NonMasqueradeCIDR != nonMasqueradeCIDRString {
				// TODO remove the Spec.MasterKubelet.NonMasqueradeCIDR field?
				if strict || c.Spec.MasterKubelet.NonMasqueradeCIDR != "" {
					allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("masterKubelet", "nonMasqueradeCIDR"), "masterKubelet nonMasqueradeCIDR did not match cluster nonMasqueradeCIDR"))
				}
			}
		}
	}

	// Check ServiceClusterIPRange
	var serviceClusterIPRange *net.IPNet
	{
		serviceClusterIPRangeString := c.Spec.ServiceClusterIPRange
		if serviceClusterIPRangeString == "" {
			if strict {
				allErrs = append(allErrs, field.Required(fieldSpec.Child("serviceClusterIPRange"), "Cluster did not have serviceClusterIPRange set"))
			}
		} else {
			_, serviceClusterIPRange, err = net.ParseCIDR(serviceClusterIPRangeString)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(fieldSpec.Child("serviceClusterIPRange"), serviceClusterIPRangeString, "Cluster had an invalid serviceClusterIPRange"))
			} else {
				if nonMasqueradeCIDR != nil && serviceClusterMustBeSubnetOfNonMasqueradeCIDR && !subnet.BelongsTo(nonMasqueradeCIDR, serviceClusterIPRange) {
					allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("serviceClusterIPRange"), fmt.Sprintf("serviceClusterIPRange %q must be a subnet of nonMasqueradeCIDR %q", serviceClusterIPRangeString, c.Spec.NonMasqueradeCIDR)))
				}

				if c.Spec.KubeAPIServer != nil && c.Spec.KubeAPIServer.ServiceClusterIPRange != serviceClusterIPRangeString {
					if strict || c.Spec.KubeAPIServer.ServiceClusterIPRange != "" {
						allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("kubeAPIServer", "serviceClusterIPRange"), "kubeAPIServer serviceClusterIPRange did not match cluster serviceClusterIPRange"))
					}
				}
			}
		}
	}

	// Check ClusterCIDR
	if c.Spec.KubeControllerManager != nil {
		var clusterCIDR *net.IPNet
		clusterCIDRString := c.Spec.KubeControllerManager.ClusterCIDR
		if clusterCIDRString != "" {
			_, clusterCIDR, err = net.ParseCIDR(clusterCIDRString)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(fieldSpec.Child("kubeControllerManager", "clusterCIDR"), clusterCIDRString, "cluster had an invalid kubeControllerManager.clusterCIDR"))
			} else if nonMasqueradeCIDR != nil && !subnet.BelongsTo(nonMasqueradeCIDR, clusterCIDR) {
				allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("kubeControllerManager", "clusterCIDR"), fmt.Sprintf("kubeControllerManager.clusterCIDR %q must be a subnet of nonMasqueradeCIDR %q", clusterCIDRString, c.Spec.NonMasqueradeCIDR)))
			}
		}
	}

	// @check the custom kubedns options are valid
	if c.Spec.KubeDNS != nil {
		if c.Spec.KubeDNS.ServerIP != "" {
			address := c.Spec.KubeDNS.ServerIP
			ip := net.ParseIP(address)
			if ip == nil {
				allErrs = append(allErrs, field.Invalid(fieldSpec.Child("kubeDNS", "serverIP"), address, "Cluster had an invalid kubeDNS.serverIP"))
			} else {
				if serviceClusterIPRange != nil && !serviceClusterIPRange.Contains(ip) {
					allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("kubeDNS", "serverIP"), fmt.Sprintf("ServiceClusterIPRange %q must contain the DNS Server IP %q", c.Spec.ServiceClusterIPRange, address)))
				}
				if !featureflag.ExperimentalClusterDNS.Enabled() {
					if isExperimentalClusterDNS(c.Spec.Kubelet, c.Spec.KubeDNS) {
						allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("kubelet", "clusterDNS"), "Kubelet ClusterDNS did not match cluster kubeDNS.serverIP or nodeLocalDNS.localIP"))
					}
					if isExperimentalClusterDNS(c.Spec.MasterKubelet, c.Spec.KubeDNS) {
						allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("masterKubelet", "clusterDNS"), "MasterKubelet ClusterDNS did not match cluster kubeDNS.serverIP or nodeLocalDNS.localIP"))
					}
				}
			}

			// @ check that NodeLocalDNS addon is configured correctly
			if c.Spec.KubeDNS.NodeLocalDNS != nil && fi.BoolValue(c.Spec.KubeDNS.NodeLocalDNS.Enabled) {
				if c.Spec.KubeDNS.Provider != "CoreDNS" {
					allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("kubeDNS", "provider"), "KubeDNS provider must be set to CoreDNS if NodeLocalDNS addon is enabled"))
				}

				allErrs = append(allErrs, validateNodeLocalDNS(&c.Spec, fieldSpec.Child("spec"))...)
			}
		}

		// @check the nameservers are valid
		for i, x := range c.Spec.KubeDNS.UpstreamNameservers {
			if ip := net.ParseIP(x); ip == nil {
				allErrs = append(allErrs, field.Invalid(fieldSpec.Child("kubeDNS", "upstreamNameservers").Index(i), x, "Invalid nameserver given, should be a valid ip address"))
			}
		}

		// @check the stubdomain if any
		for domain, nameservers := range c.Spec.KubeDNS.StubDomains {
			if len(nameservers) <= 0 {
				allErrs = append(allErrs, field.Invalid(fieldSpec.Child("kubeDNS", "stubDomains").Key(domain), domain, "No nameservers specified for the stub domain"))
			}
			for i, x := range nameservers {
				if ip := net.ParseIP(x); ip == nil {
					allErrs = append(allErrs, field.Invalid(fieldSpec.Child("kubeDNS", "stubDomains").Key(domain).Index(i), x, "Invalid nameserver given, should be a valid ip address"))
				}
			}
		}
	}

	// Check CloudProvider
	{

		var k8sCloudProvider string
		switch kops.CloudProviderID(c.Spec.CloudProvider) {
		case kops.CloudProviderAWS:
			k8sCloudProvider = "aws"
		case kops.CloudProviderGCE:
			k8sCloudProvider = "gce"
		case kops.CloudProviderDO:
			k8sCloudProvider = "external"
		case kops.CloudProviderVSphere:
			k8sCloudProvider = "vsphere"
		case kops.CloudProviderBareMetal:
			k8sCloudProvider = ""
		case kops.CloudProviderOpenstack:
			k8sCloudProvider = "openstack"
		case kops.CloudProviderALI:
			k8sCloudProvider = "alicloud"
		default:
			// We already added an error above
			k8sCloudProvider = "ignore"
		}

		if k8sCloudProvider != "ignore" {
			if c.Spec.Kubelet != nil && (strict || c.Spec.Kubelet.CloudProvider != "") {
				if c.Spec.Kubelet.CloudProvider != "external" && k8sCloudProvider != c.Spec.Kubelet.CloudProvider {
					allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("kubelet", "cloudProvider"), "Did not match cluster cloudProvider"))
				}
			}
			if c.Spec.MasterKubelet != nil && (strict || c.Spec.MasterKubelet.CloudProvider != "") {
				if c.Spec.MasterKubelet.CloudProvider != "external" && k8sCloudProvider != c.Spec.MasterKubelet.CloudProvider {
					allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("masterKubelet", "cloudProvider"), "Did not match cluster cloudProvider"))
				}
			}
			if c.Spec.KubeAPIServer != nil && (strict || c.Spec.KubeAPIServer.CloudProvider != "") {
				if c.Spec.KubeAPIServer.CloudProvider != "external" && k8sCloudProvider != c.Spec.KubeAPIServer.CloudProvider {
					allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("kubeAPIServer", "cloudProvider"), "Did not match cluster cloudProvider"))
				}
			}
			if c.Spec.KubeControllerManager != nil && (strict || c.Spec.KubeControllerManager.CloudProvider != "") {
				if c.Spec.KubeControllerManager.CloudProvider != "external" && k8sCloudProvider != c.Spec.KubeControllerManager.CloudProvider {
					allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("kubeControllerManager", "cloudProvider"), "Did not match cluster cloudProvider"))
				}
			}
		}
	}

	// Check that the subnet CIDRs are all consistent
	{
		for i, s := range c.Spec.Subnets {
			fieldSubnet := fieldSpec.Child("subnets").Index(i)
			if s.CIDR == "" {
				if requiresSubnetCIDR && strict {
					allErrs = append(allErrs, field.Required(fieldSubnet.Child("cidr"), "subnet did not have a cidr set"))
				}
			} else {
				_, subnetCIDR, err := net.ParseCIDR(s.CIDR)
				if err != nil {
					allErrs = append(allErrs, field.Invalid(fieldSubnet.Child("cidr"), s.CIDR, "subnet had an invalid cidr"))
				} else if networkCIDR != nil && !validateSubnetCIDR(networkCIDR, additionalNetworkCIDRs, subnetCIDR) {
					allErrs = append(allErrs, field.Forbidden(fieldSubnet.Child("cidr"), fmt.Sprintf("subnet %q had a cidr %q that was not a subnet of the networkCIDR %q", s.Name, s.CIDR, c.Spec.NetworkCIDR)))
				}
			}
		}
	}

	// NodeAuthorization
	if c.Spec.NodeAuthorization != nil {
		// @check the feature gate is enabled for this
		if !featureflag.EnableNodeAuthorization.Enabled() {
			allErrs = append(allErrs, field.Forbidden(field.NewPath("spec", "nodeAuthorization"), "node authorization is experimental feature; set `export KOPS_FEATURE_FLAGS=EnableNodeAuthorization`"))
		} else {
			if c.Spec.NodeAuthorization.NodeAuthorizer == nil {
				allErrs = append(allErrs, field.Forbidden(field.NewPath("spec", "nodeAuthorization"), "no node authorization policy has been set"))
			} else {
				path := field.NewPath("spec", "nodeAuthorization").Child("nodeAuthorizer")
				if c.Spec.NodeAuthorization.NodeAuthorizer.Port < 0 || c.Spec.NodeAuthorization.NodeAuthorizer.Port >= 65535 {
					allErrs = append(allErrs, field.Invalid(path.Child("port"), c.Spec.NodeAuthorization.NodeAuthorizer.Port, "invalid port"))
				}
				if c.Spec.NodeAuthorization.NodeAuthorizer.Timeout != nil && c.Spec.NodeAuthorization.NodeAuthorizer.Timeout.Duration <= 0 {
					allErrs = append(allErrs, field.Invalid(path.Child("timeout"), c.Spec.NodeAuthorization.NodeAuthorizer.Timeout, "must be greater than zero"))
				}
				if c.Spec.NodeAuthorization.NodeAuthorizer.TokenTTL != nil && c.Spec.NodeAuthorization.NodeAuthorizer.TokenTTL.Duration < 0 {
					allErrs = append(allErrs, field.Invalid(path.Child("tokenTTL"), c.Spec.NodeAuthorization.NodeAuthorizer.TokenTTL, "must be greater than or equal to zero"))
				}

				// @question: we could probably just default these settings in the model when the node-authorizer is enabled??
				if c.Spec.KubeAPIServer == nil {
					allErrs = append(allErrs, field.Required(field.NewPath("spec", "kubeAPIServer"), "bootstrap token authentication is not enabled in the kube-apiserver"))
				} else if c.Spec.KubeAPIServer.EnableBootstrapAuthToken == nil {
					allErrs = append(allErrs, field.Required(field.NewPath("spec", "kubeAPIServer", "enableBootstrapAuthToken"), "kube-apiserver has not been configured to use bootstrap tokens"))
				} else if !fi.BoolValue(c.Spec.KubeAPIServer.EnableBootstrapAuthToken) {
					allErrs = append(allErrs, field.Forbidden(field.NewPath("spec", "kubeAPIServer", "enableBootstrapAuthToken"), "bootstrap tokens in the kube-apiserver has been disabled"))
				}
			}
		}
	}

	// UpdatePolicy
	if c.Spec.UpdatePolicy != nil {
		switch *c.Spec.UpdatePolicy {
		case kops.UpdatePolicyExternal:
		// Valid
		default:
			allErrs = append(allErrs, field.NotSupported(fieldSpec.Child("updatePolicy"), *c.Spec.UpdatePolicy, []string{kops.UpdatePolicyExternal}))
		}
	}

	// KubeProxy
	if c.Spec.KubeProxy != nil {
		kubeProxyPath := fieldSpec.Child("kubeProxy")
		master := c.Spec.KubeProxy.Master

		for i, x := range c.Spec.KubeProxy.IPVSExcludeCIDRS {
			if _, _, err := net.ParseCIDR(x); err != nil {
				allErrs = append(allErrs, field.Invalid(kubeProxyPath.Child("ipvsExcludeCidrs").Index(i), x, "Invalid network CIDR"))
			}
		}

		if master != "" && !isValidAPIServersURL(master) {
			allErrs = append(allErrs, field.Invalid(kubeProxyPath.Child("master"), master, "Not a valid APIServer URL"))
		}
	}

	// KubeAPIServer
	if c.Spec.KubeAPIServer != nil {
		if c.IsKubernetesGTE("1.10") {
			if len(c.Spec.KubeAPIServer.AdmissionControl) > 0 {
				if len(c.Spec.KubeAPIServer.DisableAdmissionPlugins) > 0 {
					allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("kubeAPIServer", "disableAdmissionPlugins"),
						"disableAdmissionPlugins is mutually exclusive, you cannot use both admissionControl and disableAdmissionPlugins together"))
				}
			}
		}
	}

	// Kubelet
	allErrs = append(allErrs, validateKubelet(c.Spec.Kubelet, c, fieldSpec.Child("kubelet"))...)
	allErrs = append(allErrs, validateKubelet(c.Spec.MasterKubelet, c, fieldSpec.Child("masterKubelet"))...)

	// Topology support
	if c.Spec.Topology != nil {
		if c.Spec.Topology.Masters != "" && c.Spec.Topology.Nodes != "" {
			allErrs = append(allErrs, IsValidValue(fieldSpec.Child("topology", "masters"), &c.Spec.Topology.Masters, kops.SupportedTopologies)...)
			allErrs = append(allErrs, IsValidValue(fieldSpec.Child("topology", "nodes"), &c.Spec.Topology.Nodes, kops.SupportedTopologies)...)
		} else {
			allErrs = append(allErrs, field.Required(fieldSpec.Child("masters"), "topology requires non-nil values for masters and nodes"))
		}
		if c.Spec.Topology.Bastion != nil {
			bastion := c.Spec.Topology.Bastion
			if c.Spec.Topology.Masters == kops.TopologyPublic || c.Spec.Topology.Nodes == kops.TopologyPublic {
				allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("topology", "bastion"), "bastion requires masters and nodes to have private topology"))
			}
			if bastion.IdleTimeoutSeconds != nil && *bastion.IdleTimeoutSeconds <= 0 {
				allErrs = append(allErrs, field.Invalid(fieldSpec.Child("topology", "bastion", "idleTimeoutSeconds"), *bastion.IdleTimeoutSeconds, "bastion idleTimeoutSeconds should be greater than zero"))
			}
			if bastion.IdleTimeoutSeconds != nil && *bastion.IdleTimeoutSeconds > 3600 {
				allErrs = append(allErrs, field.Invalid(fieldSpec.Child("topology", "bastion", "idleTimeoutSeconds"), *bastion.IdleTimeoutSeconds, "bastion idleTimeoutSeconds cannot be greater than one hour"))
			}

		}
	}
	// Egress specification support
	{
		for i, s := range c.Spec.Subnets {
			if s.Egress == "" {
				continue
			}
			fieldSubnet := fieldSpec.Child("subnets").Index(i)
			if !strings.HasPrefix(s.Egress, "nat-") && !strings.HasPrefix(s.Egress, "i-") && s.Egress != kops.EgressExternal {
				allErrs = append(allErrs, field.Invalid(fieldSubnet.Child("egress"), s.Egress, "egress must be of type NAT Gateway or NAT EC2 Instance or 'External'"))
			}
			if s.Egress != kops.EgressExternal && s.Type != "Private" {
				allErrs = append(allErrs, field.Forbidden(fieldSubnet.Child("egress"), "egress can only be specified for private subnets"))
			}
		}
	}

	allErrs = append(allErrs, newValidateCluster(c)...)

	return allErrs
}

// validateSubnetCIDR is responsible for validating subnets are part of the CIDRs assigned to the cluster.
func validateSubnetCIDR(networkCIDR *net.IPNet, additionalNetworkCIDRs []*net.IPNet, subnetCIDR *net.IPNet) bool {
	if subnet.BelongsTo(networkCIDR, subnetCIDR) {
		return true
	}

	for _, additionalNetworkCIDR := range additionalNetworkCIDRs {
		if subnet.BelongsTo(additionalNetworkCIDR, subnetCIDR) {
			return true
		}
	}

	return false
}

// DeepValidate is responsible for validating the instancegroups within the cluster spec
func DeepValidate(c *kops.Cluster, groups []*kops.InstanceGroup, strict bool) error {
	if errs := ValidateCluster(c, strict); len(errs) != 0 {
		return errs.ToAggregate()
	}

	if len(groups) == 0 {
		return fmt.Errorf("must configure at least one InstanceGroup")
	}

	masterGroupCount := 0
	nodeGroupCount := 0
	for _, g := range groups {
		if g.IsMaster() {
			masterGroupCount++
		} else {
			nodeGroupCount++
		}
	}

	if masterGroupCount == 0 {
		return fmt.Errorf("must configure at least one Master InstanceGroup")
	}

	if nodeGroupCount == 0 {
		return fmt.Errorf("must configure at least one Node InstanceGroup")
	}

	for _, g := range groups {
		errs := CrossValidateInstanceGroup(g, c, strict)

		// Additional cloud-specific validation rules,
		// such as making sure that identifiers match the expected formats for the given cloud
		switch kops.CloudProviderID(c.Spec.CloudProvider) {
		case kops.CloudProviderAWS:
			errs = append(errs, awsValidateInstanceGroup(g)...)
		default:
			if len(g.Spec.Volumes) > 0 {
				errs = append(errs, field.Forbidden(field.NewPath("spec", "volumes"), "instancegroup volumes are only available with aws at present"))
			}
		}

		if len(errs) != 0 {
			return errs.ToAggregate()
		}
	}

	return nil
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

		if c.IsKubernetesGTE("1.10") {
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

	}
	return allErrs
}

func isExperimentalClusterDNS(k *kops.KubeletConfigSpec, dns *kops.KubeDNSConfig) bool {

	return k != nil && k.ClusterDNS != dns.ServerIP && dns.NodeLocalDNS != nil && k.ClusterDNS != dns.NodeLocalDNS.LocalIP

}
