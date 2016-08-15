package api

import (
	"fmt"
	"k8s.io/kubernetes/pkg/util/validation"
	"net"
	"net/url"
	"strings"
)

func (c *Cluster) Validate(strict bool) error {
	var err error

	if c.Name == "" {
		return fmt.Errorf("Cluster Name is required (e.g. --name=mycluster.myzone.com)")
	}

	{
		// Must be a dns name
		errs := validation.IsDNS1123Subdomain(c.Name)
		if len(errs) != 0 {
			return fmt.Errorf("Cluster Name must be a valid DNS name (e.g. --name=mycluster.myzone.com) errors: %s", strings.Join(errs, ", "))
		}

		if !strings.Contains(c.Name, ".") {
			return fmt.Errorf("Cluster Name must be a fully-qualified DNS name (e.g. --name=mycluster.myzone.com)")
		}
	}

	if len(c.Spec.Zones) == 0 {
		// TODO: Auto choose zones from region?
		return fmt.Errorf("must configure at least one Zone (use --zones)")
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
		if c.Spec.NetworkCIDR == "" {
			return fmt.Errorf("Cluster did not have NetworkCIDR set")
		}
		_, networkCIDR, err = net.ParseCIDR(c.Spec.NetworkCIDR)
		if err != nil {
			return fmt.Errorf("Cluster had an invalid NetworkCIDR: %q", c.Spec.NetworkCIDR)
		}
	}

	// Check NonMasqueradeCIDR
	var nonMasqueradeCIDR *net.IPNet
	{
		if c.Spec.NonMasqueradeCIDR == "" {
			return fmt.Errorf("Cluster did not have NonMasqueradeCIDR set")
		}
		_, nonMasqueradeCIDR, err = net.ParseCIDR(c.Spec.NonMasqueradeCIDR)
		if err != nil {
			return fmt.Errorf("Cluster had an invalid NonMasqueradeCIDR: %q", c.Spec.NonMasqueradeCIDR)
		}

		if subnetsOverlap(nonMasqueradeCIDR, networkCIDR) {
			return fmt.Errorf("NonMasqueradeCIDR %q cannot overlap with NetworkCIDR %q", c.Spec.NonMasqueradeCIDR, c.Spec.NetworkCIDR)
		}

		if c.Spec.Kubelet != nil && c.Spec.Kubelet.NonMasqueradeCIDR != c.Spec.NonMasqueradeCIDR {
			return fmt.Errorf("Kubelet NonMasqueradeCIDR did not match cluster NonMasqueradeCIDR")
		}
		if c.Spec.MasterKubelet != nil && c.Spec.MasterKubelet.NonMasqueradeCIDR != c.Spec.NonMasqueradeCIDR {
			return fmt.Errorf("MasterKubelet NonMasqueradeCIDR did not match cluster NonMasqueradeCIDR")
		}
	}

	// Check ServiceClusterIPRange
	var serviceClusterIPRange *net.IPNet
	{
		if c.Spec.ServiceClusterIPRange == "" {
			if strict {
				return fmt.Errorf("Cluster did not have ServiceClusterIPRange set")
			}
		} else {
			_, serviceClusterIPRange, err = net.ParseCIDR(c.Spec.ServiceClusterIPRange)
			if err != nil {
				return fmt.Errorf("Cluster had an invalid ServiceClusterIPRange: %q", c.Spec.ServiceClusterIPRange)
			}

			if !isSubnet(nonMasqueradeCIDR, serviceClusterIPRange) {
				return fmt.Errorf("ServiceClusterIPRange %q must be a subnet of NonMasqueradeCIDR %q", c.Spec.ServiceClusterIPRange, c.Spec.NonMasqueradeCIDR)
			}

			if c.Spec.KubeAPIServer != nil && c.Spec.KubeAPIServer.ServiceClusterIPRange != c.Spec.ServiceClusterIPRange {
				return fmt.Errorf("KubeAPIServer ServiceClusterIPRange did not match cluster ServiceClusterIPRange")
			}
		}
	}

	// Check ClusterCIDR
	if c.Spec.KubeControllerManager != nil {
		var clusterCIDR *net.IPNet
		if c.Spec.KubeControllerManager.ClusterCIDR != "" {
			_, clusterCIDR, err = net.ParseCIDR(c.Spec.KubeControllerManager.ClusterCIDR)
			if err != nil {
				return fmt.Errorf("Cluster had an invalid KubeControllerManager.ClusterCIDR: %q", c.Spec.KubeControllerManager.ClusterCIDR)
			}

			if !isSubnet(nonMasqueradeCIDR, clusterCIDR) {
				return fmt.Errorf("KubeControllerManager.ClusterCIDR %q must be a subnet of NonMasqueradeCIDR %q", c.Spec.KubeControllerManager.ClusterCIDR, c.Spec.NonMasqueradeCIDR)
			}
		}
	}

	// Check KubeDNS.ServerIP
	if c.Spec.KubeDNS != nil {
		if c.Spec.KubeDNS.ServerIP == "" {
			return fmt.Errorf("Cluster did not have KubeDNS.ServerIP set")
		}

		dnsServiceIP := net.ParseIP(c.Spec.KubeDNS.ServerIP)
		if dnsServiceIP == nil {
			return fmt.Errorf("Cluster had an invalid KubeDNS.ServerIP: %q", c.Spec.KubeDNS.ServerIP)
		}

		if !serviceClusterIPRange.Contains(dnsServiceIP) {
			return fmt.Errorf("ServiceClusterIPRange %q must contain the DNS Server IP %q", c.Spec.ServiceClusterIPRange, c.Spec.KubeDNS.ServerIP)
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
		if c.Spec.CloudProvider == "" {
			return fmt.Errorf("CloudProvider is not set")
		}
		if c.Spec.Kubelet != nil && c.Spec.Kubelet.CloudProvider != "" && c.Spec.Kubelet.CloudProvider != c.Spec.CloudProvider {
			return fmt.Errorf("Kubelet CloudProvider did not match cluster CloudProvider")
		}
		if c.Spec.MasterKubelet != nil && c.Spec.MasterKubelet.CloudProvider != "" && c.Spec.MasterKubelet.CloudProvider != c.Spec.CloudProvider {
			return fmt.Errorf("MasterKubelet CloudProvider did not match cluster CloudProvider")
		}
		if c.Spec.KubeAPIServer != nil && c.Spec.KubeAPIServer.CloudProvider != "" && c.Spec.KubeAPIServer.CloudProvider != c.Spec.CloudProvider {
			return fmt.Errorf("KubeAPIServer CloudProvider did not match cluster CloudProvider")
		}
		if c.Spec.KubeControllerManager != nil && c.Spec.KubeControllerManager.CloudProvider != "" && c.Spec.KubeControllerManager.CloudProvider != c.Spec.CloudProvider {
			return fmt.Errorf("KubeControllerManager CloudProvider did not match cluster CloudProvider")
		}
	}

	// Check that the zone CIDRs are all consistent
	{

		for _, z := range c.Spec.Zones {
			if z.CIDR == "" {
				if strict {
					return fmt.Errorf("Zone %q did not have a CIDR set", z.Name)
				}
			} else {
				_, zoneCIDR, err := net.ParseCIDR(z.CIDR)
				if err != nil {
					return fmt.Errorf("Zone %q had an invalid CIDR: %q", z.Name, z.CIDR)
				}

				if !isSubnet(networkCIDR, zoneCIDR) {
					return fmt.Errorf("Zone %q had a CIDR %q that was not a subnet of the NetworkCIDR %q", z.Name, z.CIDR, c.Spec.NetworkCIDR)
				}
			}
		}
	}

	// AdminAccess
	if strict && len(c.Spec.AdminAccess) == 0 {
		return fmt.Errorf("AdminAccess not configured")
	}

	for _, adminAccess := range c.Spec.AdminAccess {
		_, _, err := net.ParseCIDR(adminAccess)
		if err != nil {
			return fmt.Errorf("AdminAccess rule %q could not be parsed (invalid CIDR)", adminAccess)
		}
	}

	// KubeProxy
	if c.Spec.KubeProxy != nil {
		if strict && c.Spec.KubeProxy.Master == "" {
			return fmt.Errorf("KubeProxy.Master not set")
		}
		if c.Spec.KubeProxy.Master != "" && !isValidAPIServersURL(c.Spec.KubeProxy.Master) {
			return fmt.Errorf("KubeProxy.Master not valid")
		}
	}

	// Kubelet
	if c.Spec.Kubelet != nil {
		if strict && c.Spec.Kubelet.APIServers == "" {
			return fmt.Errorf("Kubelet.APIServers not set")
		}
		if c.Spec.Kubelet.APIServers != "" && !isValidAPIServersURL(c.Spec.Kubelet.APIServers) {
			return fmt.Errorf("Kubelet.APIServers not valid")
		}
	}

	// MasterKubelet
	if c.Spec.MasterKubelet != nil {
		if strict && c.Spec.MasterKubelet.APIServers == "" {
			return fmt.Errorf("MasterKubelet.APIServers not set")
		}
		if c.Spec.MasterKubelet.APIServers != "" && !isValidAPIServersURL(c.Spec.MasterKubelet.APIServers) {
			return fmt.Errorf("MasterKubelet.APIServers not valid")
		}
	}

	// Etcd
	{
		if strict && len(c.Spec.EtcdClusters) == 0 {
			return fmt.Errorf("EtcdClusters not configured")
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
				if m.Zone == "" {
					return fmt.Errorf("EtcdMember did not have Zone in cluster %q", etcd.Name)
				}
			}
		}
	}

	return nil
}

func isValidAPIServersURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	if u.Host == "" || u.Scheme == "" {
		return false
	}
	return true
}

func DeepValidate(c *Cluster, groups []*InstanceGroup, strict bool) error {
	err := c.Validate(strict)
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

// isSubnet checks if child is a subnet of parent
func isSubnet(parent *net.IPNet, child *net.IPNet) bool {
	parentOnes, parentBits := parent.Mask.Size()
	childOnes, childBits := child.Mask.Size()
	if childBits != parentBits {
		return false
	}
	if parentOnes > childOnes {
		return false
	}
	childMasked := child.IP.Mask(parent.Mask)
	parentMasked := parent.IP.Mask(parent.Mask)
	return childMasked.Equal(parentMasked)
}

// subnetsOverlap checks if two subnets overlap
func subnetsOverlap(l *net.IPNet, r *net.IPNet) bool {
	return l.Contains(r.IP) || r.Contains(l.IP)
}
