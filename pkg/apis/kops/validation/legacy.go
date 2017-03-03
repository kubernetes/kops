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
	"github.com/blang/semver"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"net"
	"strings"
)

// legacy contains validation functions that don't match the apimachinery style

func ValidateCluster(c *kops.Cluster, strict bool) error {
	specField := field.NewPath("Spec")

	var err error

	specPath := field.NewPath("Cluster").Child("Spec")

	// kubernetesRelease is the version with only major & minor fields
	var kubernetesRelease semver.Version

	// KubernetesVersion
	if c.Spec.KubernetesVersion == "" {
		return field.Required(specField.Child("KubernetesVersion"), "")
	} else {
		sv, err := util.ParseKubernetesVersion(c.Spec.KubernetesVersion)
		if err != nil {
			return field.Invalid(specField.Child("KubernetesVersion"), c.Spec.KubernetesVersion, "unable to determine kubernetes version")
		}

		kubernetesRelease = semver.Version{Major: sv.Major, Minor: sv.Minor}
	}

	if c.ObjectMeta.Name == "" {
		return field.Required(field.NewPath("Name"), "Cluster Name is required (e.g. --name=mycluster.myzone.com)")
	}

	{
		// Must be a dns name
		errs := validation.IsDNS1123Subdomain(c.ObjectMeta.Name)
		if len(errs) != 0 {
			return fmt.Errorf("Cluster Name must be a valid DNS name (e.g. --name=mycluster.myzone.com) errors: %s", strings.Join(errs, ", "))
		}

		if !strings.Contains(c.ObjectMeta.Name, ".") {
			// Tolerate if this is a cluster we are importing for upgrade
			if c.ObjectMeta.Annotations[kops.AnnotationNameManagement] != kops.AnnotationValueManagementImported {
				return fmt.Errorf("Cluster Name must be a fully-qualified DNS name (e.g. --name=mycluster.myzone.com)")
			}
		}
	}

	if len(c.Spec.Subnets) == 0 {
		// TODO: Auto choose zones from region?
		return fmt.Errorf("must configure at least one Subnet (use --zones)")
	}

	if strict && c.Spec.Kubelet == nil {
		return fmt.Errorf("Kubelet not configured")
	}
	if strict && c.Spec.MasterKubelet == nil {
		return fmt.Errorf("MasterKubelet not configured")
	}
	if strict && c.Spec.KubeControllerManager == nil {
		return fmt.Errorf("KubeControllerManager not configured")
	}
	if strict && c.Spec.KubeDNS == nil {
		return fmt.Errorf("KubeDNS not configured")
	}
	if strict && c.Spec.Kubelet == nil {
		return fmt.Errorf("Kubelet not configured")
	}
	if strict && c.Spec.KubeAPIServer == nil {
		return fmt.Errorf("KubeAPIServer not configured")
	}
	if strict && c.Spec.KubeProxy == nil {
		return fmt.Errorf("KubeProxy not configured")
	}
	if strict && c.Spec.Docker == nil {
		return fmt.Errorf("Docker not configured")
	}

	// Check NetworkCIDR
	var networkCIDR *net.IPNet
	{
		networkCIDRString := c.Spec.NetworkCIDR
		if networkCIDRString == "" {
			return field.Required(specField.Child("NetworkCIDR"), "Cluster did not have NetworkCIDR set")
		}
		_, networkCIDR, err = net.ParseCIDR(networkCIDRString)
		if err != nil {
			return field.Invalid(specField.Child("NetworkCIDR"), networkCIDRString, fmt.Sprintf("Cluster had an invalid NetworkCIDR"))
		}
	}

	// Check NonMasqueradeCIDR
	var nonMasqueradeCIDR *net.IPNet
	{
		nonMasqueradeCIDRString := c.Spec.NonMasqueradeCIDR
		if nonMasqueradeCIDRString == "" {
			return fmt.Errorf("Cluster did not have NonMasqueradeCIDR set")
		}
		_, nonMasqueradeCIDR, err = net.ParseCIDR(nonMasqueradeCIDRString)
		if err != nil {
			return fmt.Errorf("Cluster had an invalid NonMasqueradeCIDR: %q", nonMasqueradeCIDRString)
		}

		if subnetsOverlap(nonMasqueradeCIDR, networkCIDR) {
			return fmt.Errorf("NonMasqueradeCIDR %q cannot overlap with NetworkCIDR %q", nonMasqueradeCIDRString, c.Spec.NetworkCIDR)
		}

		if c.Spec.Kubelet != nil && c.Spec.Kubelet.NonMasqueradeCIDR != nonMasqueradeCIDRString {
			if strict || c.Spec.Kubelet.NonMasqueradeCIDR != "" {
				return fmt.Errorf("Kubelet NonMasqueradeCIDR did not match cluster NonMasqueradeCIDR")
			}
		}
		if c.Spec.MasterKubelet != nil && c.Spec.MasterKubelet.NonMasqueradeCIDR != nonMasqueradeCIDRString {
			if strict || c.Spec.MasterKubelet.NonMasqueradeCIDR != "" {
				return fmt.Errorf("MasterKubelet NonMasqueradeCIDR did not match cluster NonMasqueradeCIDR")
			}
		}
	}

	// Check ServiceClusterIPRange
	var serviceClusterIPRange *net.IPNet
	{
		serviceClusterIPRangeString := c.Spec.ServiceClusterIPRange
		if serviceClusterIPRangeString == "" {
			if strict {
				return fmt.Errorf("Cluster did not have ServiceClusterIPRange set")
			}
		} else {
			_, serviceClusterIPRange, err = net.ParseCIDR(serviceClusterIPRangeString)
			if err != nil {
				return fmt.Errorf("Cluster had an invalid ServiceClusterIPRange: %q", serviceClusterIPRangeString)
			}

			if !isSubnet(nonMasqueradeCIDR, serviceClusterIPRange) {
				return fmt.Errorf("ServiceClusterIPRange %q must be a subnet of NonMasqueradeCIDR %q", serviceClusterIPRangeString, c.Spec.NonMasqueradeCIDR)
			}

			if c.Spec.KubeAPIServer != nil && c.Spec.KubeAPIServer.ServiceClusterIPRange != serviceClusterIPRangeString {
				if strict || c.Spec.KubeAPIServer.ServiceClusterIPRange != "" {
					return fmt.Errorf("KubeAPIServer ServiceClusterIPRange did not match cluster ServiceClusterIPRange")
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
				return fmt.Errorf("Cluster had an invalid KubeControllerManager.ClusterCIDR: %q", clusterCIDRString)
			}

			if !isSubnet(nonMasqueradeCIDR, clusterCIDR) {
				return fmt.Errorf("KubeControllerManager.ClusterCIDR %q must be a subnet of NonMasqueradeCIDR %q", clusterCIDRString, c.Spec.NonMasqueradeCIDR)
			}
		}
	}

	// Check KubeDNS.ServerIP
	if c.Spec.KubeDNS != nil {
		serverIPString := c.Spec.KubeDNS.ServerIP
		if serverIPString == "" {
			return fmt.Errorf("Cluster did not have KubeDNS.ServerIP set")
		}

		dnsServiceIP := net.ParseIP(serverIPString)
		if dnsServiceIP == nil {
			return fmt.Errorf("Cluster had an invalid KubeDNS.ServerIP: %q", serverIPString)
		}

		if !serviceClusterIPRange.Contains(dnsServiceIP) {
			return fmt.Errorf("ServiceClusterIPRange %q must contain the DNS Server IP %q", c.Spec.ServiceClusterIPRange, serverIPString)
		}

		if c.Spec.Kubelet != nil && c.Spec.Kubelet.ClusterDNS != c.Spec.KubeDNS.ServerIP {
			return fmt.Errorf("Kubelet ClusterDNS did not match cluster KubeDNS.ServerIP")
		}
		if c.Spec.MasterKubelet != nil && c.Spec.MasterKubelet.ClusterDNS != c.Spec.KubeDNS.ServerIP {
			return fmt.Errorf("MasterKubelet ClusterDNS did not match cluster KubeDNS.ServerIP")
		}
	}

	// Check CloudProvider
	{
		cloudProvider := c.Spec.CloudProvider

		if cloudProvider == "" {
			return field.Required(specPath.Child("CloudProvider"), "")
		}
		if c.Spec.Kubelet != nil && (strict || c.Spec.Kubelet.CloudProvider != "") {
			if cloudProvider != c.Spec.Kubelet.CloudProvider {
				return field.Invalid(specPath.Child("Kubelet", "CloudProvider"), c.Spec.Kubelet.CloudProvider, "Did not match cluster CloudProvider")
			}
		}
		if c.Spec.MasterKubelet != nil && (strict || c.Spec.MasterKubelet.CloudProvider != "") {
			if cloudProvider != c.Spec.MasterKubelet.CloudProvider {
				return field.Invalid(specPath.Child("MasterKubelet", "CloudProvider"), c.Spec.MasterKubelet.CloudProvider, "Did not match cluster CloudProvider")
			}
		}
		if c.Spec.KubeAPIServer != nil && (strict || c.Spec.KubeAPIServer.CloudProvider != "") {
			if cloudProvider != c.Spec.KubeAPIServer.CloudProvider {
				return field.Invalid(specPath.Child("KubeAPIServer", "CloudProvider"), c.Spec.KubeAPIServer.CloudProvider, "Did not match cluster CloudProvider")
			}
		}
		if c.Spec.KubeControllerManager != nil && (strict || c.Spec.KubeControllerManager.CloudProvider != "") {
			if cloudProvider != c.Spec.KubeControllerManager.CloudProvider {
				return field.Invalid(specPath.Child("KubeControllerManager", "CloudProvider"), c.Spec.KubeControllerManager.CloudProvider, "Did not match cluster CloudProvider")
			}
		}
	}

	// Check that the subnet CIDRs are all consistent
	{
		for _, s := range c.Spec.Subnets {
			if s.CIDR == "" {
				if strict {
					return fmt.Errorf("Subnet %q did not have a CIDR set", s.Name)
				}
			} else {
				_, subnetCIDR, err := net.ParseCIDR(s.CIDR)
				if err != nil {
					return fmt.Errorf("Subnet %q had an invalid CIDR: %q", s.Name, s.CIDR)
				}

				if !isSubnet(networkCIDR, subnetCIDR) {
					return fmt.Errorf("Subnet %q had a CIDR %q that was not a subnet of the NetworkCIDR %q", s.Name, s.CIDR, c.Spec.NetworkCIDR)
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
			return fmt.Errorf("unrecognized value for UpdatePolicy: %v", *c.Spec.UpdatePolicy)
		}
	}

	// AdminAccess
	if strict && len(c.Spec.SSHAccess) == 0 {
		// TODO: We may want to allow this
		return fmt.Errorf("SSHAccess not configured")
	}

	// AdminAccess
	if strict && len(c.Spec.KubernetesAPIAccess) == 0 {
		// TODO: We may want to allow this (maybe)
		return fmt.Errorf("KubernetesAPIAccess not configured")
	}

	// KubeProxy
	if c.Spec.KubeProxy != nil {
		kubeProxyPath := specPath.Child("KubeProxy")

		master := c.Spec.KubeProxy.Master
		if strict && master == "" {
			return field.Required(kubeProxyPath.Child("Master"), "")
		}
		if master != "" && !isValidAPIServersURL(master) {
			return field.Invalid(kubeProxyPath.Child("Master"), master, "Not a valid APIServer URL")
		}
	}

	// Kubelet
	if c.Spec.Kubelet != nil {
		kubeletPath := specPath.Child("Kubelet")

		if kubernetesRelease.GTE(semver.MustParse("1.6.0")) {
			// Flag removed in 1.6
			if c.Spec.Kubelet.APIServers != "" {
				return field.Invalid(
					kubeletPath.Child("APIServers"),
					c.Spec.Kubelet.APIServers,
					"api-servers flag was removed in 1.6")
			}
		} else {
			if strict && c.Spec.Kubelet.APIServers == "" {
				return field.Required(kubeletPath.Child("APIServers"), "")
			}
		}

		if c.Spec.Kubelet.APIServers != "" && !isValidAPIServersURL(c.Spec.Kubelet.APIServers) {
			return field.Invalid(kubeletPath.Child("APIServers"), c.Spec.Kubelet.APIServers, "Not a valid APIServer URL")
		}
	}

	// MasterKubelet
	if c.Spec.MasterKubelet != nil {
		masterKubeletPath := specPath.Child("MasterKubelet")

		if kubernetesRelease.GTE(semver.MustParse("1.6.0")) {
			// Flag removed in 1.6
			if c.Spec.MasterKubelet.APIServers != "" {
				return field.Invalid(
					masterKubeletPath.Child("APIServers"),
					c.Spec.MasterKubelet.APIServers,
					"api-servers flag was removed in 1.6")
			}
		} else {
			if strict && c.Spec.MasterKubelet.APIServers == "" {
				return field.Required(masterKubeletPath.Child("APIServers"), "")
			}
		}

		if c.Spec.MasterKubelet.APIServers != "" && !isValidAPIServersURL(c.Spec.MasterKubelet.APIServers) {
			return field.Invalid(masterKubeletPath.Child("APIServers"), c.Spec.MasterKubelet.APIServers, "Not a valid APIServer URL")
		}
	}

	// Topology support
	if c.Spec.Topology != nil {
		if c.Spec.Topology.Masters != "" && c.Spec.Topology.Nodes != "" {
			if c.Spec.Topology.Masters != kops.TopologyPublic && c.Spec.Topology.Masters != kops.TopologyPrivate {
				return fmt.Errorf("Invalid Masters value for Topology")
			} else if c.Spec.Topology.Nodes != kops.TopologyPublic && c.Spec.Topology.Nodes != kops.TopologyPrivate {
				return fmt.Errorf("Invalid Nodes value for Topology")
			}

		} else {
			return fmt.Errorf("Topology requires non-nil values for Masters and Nodes")
		}
		if c.Spec.Topology.Bastion != nil {
			bastion := c.Spec.Topology.Bastion
			if c.Spec.Topology.Masters == kops.TopologyPublic || c.Spec.Topology.Nodes == kops.TopologyPublic {
				return fmt.Errorf("Bastion supports only Private Masters and Nodes")
			}
			if bastion.IdleTimeoutSeconds != nil && *bastion.IdleTimeoutSeconds <= 0 {
				return fmt.Errorf("Bastion IdleTimeoutSeconds should be greater than zero")
			}
			if bastion.IdleTimeoutSeconds != nil && *bastion.IdleTimeoutSeconds > 3600 {
				return fmt.Errorf("Bastion IdleTimeoutSeconds cannot be greater than one hour")
			}

		}
	}
	// Egress specification support
	{
		for _, s := range c.Spec.Subnets {
			if s.Egress != "" && !strings.HasPrefix(s.Egress, "nat-") {
				return fmt.Errorf("egress must be of type NAT Gateway")
			}
			if s.Egress != "" && !(s.Type == "Private") {
				return fmt.Errorf("egress can only be specified for Private subnets")
			}
		}
	}

	// Etcd
	{
		if len(c.Spec.EtcdClusters) == 0 {
			return field.Required(specField.Child("EtcdClusters"), "")
		}
		for _, etcd := range c.Spec.EtcdClusters {
			if etcd.Name == "" {
				return fmt.Errorf("EtcdCluster did not have name")
			}
			if len(etcd.Members) == 0 {
				return fmt.Errorf("No members defined in etcd cluster %q", etcd.Name)
			}
			if (len(etcd.Members) % 2) == 0 {
				// Not technically a requirement, but doesn't really make sense to allow
				return fmt.Errorf("There should be an odd number of master-zones, for etcd's quorum.  Hint: Use --zones and --master-zones to declare node zones and master zones separately.")
			}
			for _, m := range etcd.Members {
				if m.Name == "" {
					return fmt.Errorf("EtcdMember did not have Name in cluster %q", etcd.Name)
				}
				if fi.StringValue(m.InstanceGroup) == "" {
					return fmt.Errorf("EtcdMember did not have InstanceGroup in cluster %q", etcd.Name)
				}
			}
		}
	}

	if kubernetesRelease.GTE(semver.MustParse("1.4.0")) {
		if c.Spec.Networking != nil && c.Spec.Networking.Classic != nil {
			return field.Invalid(specField.Child("Networking"), "classic", "classic networking is not supported with kubernetes versions 1.4 and later")
		}
	}

	errs := newValidateCluster(c)
	if len(errs) != 0 {
		return errs[0]
	}

	return nil
}

func DeepValidate(c *kops.Cluster, groups []*kops.InstanceGroup, strict bool) error {
	err := ValidateCluster(c, strict)
	if err != nil {
		return err
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
		err := g.CrossValidate(c, strict)
		if err != nil {
			return err
		}
	}

	return nil
}
