/*
Copyright 2020 The Kubernetes Authors.

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

package azuremodel

import (
	"strconv"

	network "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azuretasks"
	"k8s.io/utils/net"
)

// NetworkModelBuilder configures a Virtual Network and subnets.
type NetworkModelBuilder struct {
	*AzureModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.CloudupModelBuilder = &NetworkModelBuilder{}

// Build builds tasks for creating a virtual network and subnets.
func (b *NetworkModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	networkTask := &azuretasks.VirtualNetwork{
		Name:          new(b.NameForVirtualNetwork()),
		Lifecycle:     b.Lifecycle,
		ResourceGroup: b.LinkToResourceGroup(),
		CIDR:          new(b.Cluster.Spec.Networking.NetworkCIDR),
		Tags:          map[string]*string{},
		Shared:        new(b.Cluster.SharedVPC()),
	}
	c.AddTask(networkTask)

	ngwPipTask := &azuretasks.PublicIPAddress{
		Name:             new(b.NameForVirtualNetwork()),
		Lifecycle:        b.Lifecycle,
		ResourceGroup:    b.LinkToResourceGroup(),
		IPVersion:        network.IPVersionIPv4,
		AllocationMethod: network.IPAllocationMethodStatic,
		SKU:              network.PublicIPAddressSKUNameStandard,
		Tags:             map[string]*string{},
	}
	c.AddTask(ngwPipTask)

	nsgTask := &azuretasks.NetworkSecurityGroup{
		Name:          new(b.Cluster.AzureNetworkSecurityGroupName()),
		Lifecycle:     b.Lifecycle,
		ResourceGroup: b.LinkToResourceGroup(),
		ApplicationSecurityGroups: []*azuretasks.ApplicationSecurityGroup{
			b.LinkToApplicationSecurityGroupControlPlane(),
			b.LinkToApplicationSecurityGroupNodes(),
		},
		Tags: map[string]*string{},
	}
	sshAccessIPv4 := ipv4CIDRs(b.Cluster.Spec.SSHAccess)
	if len(sshAccessIPv4) > 0 {
		nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
			Name:                  new("AllowSSH"),
			Priority:              new(int32(100)),
			Access:                network.SecurityRuleAccessAllow,
			Direction:             network.SecurityRuleDirectionInbound,
			Protocol:              network.SecurityRuleProtocolTCP,
			SourceAddressPrefixes: sshAccessIPv4,
			SourcePortRange:       new("*"),
			DestinationApplicationSecurityGroupNames: []*string{
				new(b.NameForApplicationSecurityGroupControlPlane()),
				new(b.NameForApplicationSecurityGroupNodes()),
			},
			DestinationPortRange: new("22"),
		})
	}
	sshAccessIPv6 := ipv6CIDRs(b.Cluster.Spec.SSHAccess)
	if len(sshAccessIPv6) > 0 {
		nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
			Name:                  new("AllowSSH_v6"),
			Priority:              new(int32(101)),
			Access:                network.SecurityRuleAccessAllow,
			Direction:             network.SecurityRuleDirectionInbound,
			Protocol:              network.SecurityRuleProtocolTCP,
			SourceAddressPrefixes: sshAccessIPv6,
			SourcePortRange:       new("*"),
			DestinationApplicationSecurityGroupNames: []*string{
				new(b.NameForApplicationSecurityGroupControlPlane()),
				new(b.NameForApplicationSecurityGroupNodes()),
			},
			DestinationPortRange: new("22"),
		})
	}
	k8sAccessIPv4 := ipv4CIDRs(b.Cluster.Spec.API.Access)
	if len(k8sAccessIPv4) > 0 {
		nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
			Name:                                     new("AllowKubernetesAPI"),
			Priority:                                 new(int32(200)),
			Access:                                   network.SecurityRuleAccessAllow,
			Direction:                                network.SecurityRuleDirectionInbound,
			Protocol:                                 network.SecurityRuleProtocolTCP,
			SourceAddressPrefixes:                    k8sAccessIPv4,
			SourcePortRange:                          new("*"),
			DestinationApplicationSecurityGroupNames: []*string{new(b.NameForApplicationSecurityGroupControlPlane())},
			DestinationPortRange:                     new(strconv.Itoa(wellknownports.KubeAPIServer)),
		})
	}
	k8sAccessIPv6 := ipv6CIDRs(b.Cluster.Spec.API.Access)
	if len(k8sAccessIPv6) > 0 {
		nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
			Name:                                     new("AllowKubernetesAPI_v6"),
			Priority:                                 new(int32(201)),
			Access:                                   network.SecurityRuleAccessAllow,
			Direction:                                network.SecurityRuleDirectionInbound,
			Protocol:                                 network.SecurityRuleProtocolTCP,
			SourceAddressPrefixes:                    k8sAccessIPv6,
			SourcePortRange:                          new("*"),
			DestinationApplicationSecurityGroupNames: []*string{new(b.NameForApplicationSecurityGroupControlPlane())},
			DestinationPortRange:                     new(strconv.Itoa(wellknownports.KubeAPIServer)),
		})
	}
	nodePortAccessIPv4 := ipv4CIDRs(b.Cluster.Spec.NodePortAccess)
	if len(nodePortAccessIPv4) > 0 {
		nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
			Name:                                     new("AllowNodePortTCP"),
			Priority:                                 new(int32(300)),
			Access:                                   network.SecurityRuleAccessAllow,
			Direction:                                network.SecurityRuleDirectionInbound,
			Protocol:                                 network.SecurityRuleProtocolAsterisk,
			SourceAddressPrefixes:                    nodePortAccessIPv4,
			SourcePortRange:                          new("*"),
			DestinationApplicationSecurityGroupNames: []*string{new(b.NameForApplicationSecurityGroupNodes())},
			DestinationPortRange:                     new("30000-32767"),
		})
	}
	nodePortAccessIPv6 := ipv6CIDRs(b.Cluster.Spec.NodePortAccess)
	if len(nodePortAccessIPv6) > 0 {
		nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
			Name:                                     new("AllowNodePortTCP_v6"),
			Priority:                                 new(int32(301)),
			Access:                                   network.SecurityRuleAccessAllow,
			Direction:                                network.SecurityRuleDirectionInbound,
			Protocol:                                 network.SecurityRuleProtocolAsterisk,
			SourceAddressPrefixes:                    nodePortAccessIPv6,
			SourcePortRange:                          new("*"),
			DestinationApplicationSecurityGroupNames: []*string{new(b.NameForApplicationSecurityGroupNodes())},
			DestinationPortRange:                     new("30000-32767"),
		})
	}
	nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
		Name:                                     new("AllowControlPlaneToControlPlane"),
		Priority:                                 new(int32(1000)),
		Access:                                   network.SecurityRuleAccessAllow,
		Direction:                                network.SecurityRuleDirectionInbound,
		Protocol:                                 network.SecurityRuleProtocolAsterisk,
		SourceApplicationSecurityGroupNames:      []*string{new(b.NameForApplicationSecurityGroupControlPlane())},
		SourcePortRange:                          new("*"),
		DestinationApplicationSecurityGroupNames: []*string{new(b.NameForApplicationSecurityGroupControlPlane())},
		DestinationPortRange:                     new("*"),
	})
	nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
		Name:                                     new("AllowControlPlaneToNodes"),
		Priority:                                 new(int32(1001)),
		Access:                                   network.SecurityRuleAccessAllow,
		Direction:                                network.SecurityRuleDirectionInbound,
		Protocol:                                 network.SecurityRuleProtocolAsterisk,
		SourceApplicationSecurityGroupNames:      []*string{new(b.NameForApplicationSecurityGroupControlPlane())},
		SourcePortRange:                          new("*"),
		DestinationApplicationSecurityGroupNames: []*string{new(b.NameForApplicationSecurityGroupNodes())},
		DestinationPortRange:                     new("*"),
	})
	nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
		Name:                                     new("AllowNodesToNodes"),
		Priority:                                 new(int32(1002)),
		Access:                                   network.SecurityRuleAccessAllow,
		Direction:                                network.SecurityRuleDirectionInbound,
		Protocol:                                 network.SecurityRuleProtocolAsterisk,
		SourceApplicationSecurityGroupNames:      []*string{new(b.NameForApplicationSecurityGroupNodes())},
		SourcePortRange:                          new("*"),
		DestinationApplicationSecurityGroupNames: []*string{new(b.NameForApplicationSecurityGroupNodes())},
		DestinationPortRange:                     new("*"),
	})
	// Kindnet preserves pod-CIDR source IPs to host services; Azure NSG ASG
	// matching needs an explicit allow since pod IPs aren't NIC-assigned.
	// Pods reach the nodes on all ports and the kube-apiserver on 443 only,
	// so DenyAllToControlPlane and the etcd denies stay effective.
	if b.Cluster.Spec.Networking.Kindnet != nil && b.Cluster.Spec.Networking.PodCIDR != "" {
		nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
			Name:                                     new("AllowPodCIDRToNodes"),
			Priority:                                 new(int32(1006)),
			Access:                                   network.SecurityRuleAccessAllow,
			Direction:                                network.SecurityRuleDirectionInbound,
			Protocol:                                 network.SecurityRuleProtocolAsterisk,
			SourceAddressPrefix:                      new(b.Cluster.Spec.Networking.PodCIDR),
			SourcePortRange:                          new("*"),
			DestinationApplicationSecurityGroupNames: []*string{new(b.NameForApplicationSecurityGroupNodes())},
			DestinationPortRange:                     new("*"),
		})
		nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
			Name:                                     new("AllowPodCIDRToKubernetesAPI"),
			Priority:                                 new(int32(1007)),
			Access:                                   network.SecurityRuleAccessAllow,
			Direction:                                network.SecurityRuleDirectionInbound,
			Protocol:                                 network.SecurityRuleProtocolTCP,
			SourceAddressPrefix:                      new(b.Cluster.Spec.Networking.PodCIDR),
			SourcePortRange:                          new("*"),
			DestinationApplicationSecurityGroupNames: []*string{new(b.NameForApplicationSecurityGroupControlPlane())},
			DestinationPortRange:                     new(strconv.Itoa(wellknownports.KubeAPIServer)),
		})
	}
	etcdPeerMax := wellknownports.EtcdEventsPeerPort
	for _, c := range b.Cluster.Spec.EtcdClusters {
		if c.Name == "leases" {
			etcdPeerMax = wellknownports.EtcdLeasesPeerPort
			break
		}
	}
	nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
		Name:                                     new("DenyNodesToEtcdManager"),
		Priority:                                 new(int32(1003)),
		Access:                                   network.SecurityRuleAccessDeny,
		Direction:                                network.SecurityRuleDirectionInbound,
		Protocol:                                 network.SecurityRuleProtocolTCP,
		SourceApplicationSecurityGroupNames:      []*string{new(b.NameForApplicationSecurityGroupNodes())},
		SourcePortRange:                          new("*"),
		DestinationApplicationSecurityGroupNames: []*string{new(b.NameForApplicationSecurityGroupControlPlane())},
		DestinationPortRange:                     new(strconv.Itoa(wellknownports.EtcdMainPeerPort) + "-" + strconv.Itoa(etcdPeerMax)),
	})
	nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
		Name:                                     new("DenyNodesToEtcd"),
		Priority:                                 new(int32(1004)),
		Access:                                   network.SecurityRuleAccessDeny,
		Direction:                                network.SecurityRuleDirectionInbound,
		Protocol:                                 network.SecurityRuleProtocolTCP,
		SourceApplicationSecurityGroupNames:      []*string{new(b.NameForApplicationSecurityGroupNodes())},
		SourcePortRange:                          new("*"),
		DestinationApplicationSecurityGroupNames: []*string{new(b.NameForApplicationSecurityGroupControlPlane())},
		DestinationPortRange:                     new(strconv.Itoa(wellknownports.ProtokubeGossipMemberlist) + "-" + strconv.Itoa(wellknownports.EtcdMainClientPort)),
	})
	nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
		Name:                                     new("AllowNodesToControlPlane"),
		Priority:                                 new(int32(1005)),
		Access:                                   network.SecurityRuleAccessAllow,
		Direction:                                network.SecurityRuleDirectionInbound,
		Protocol:                                 network.SecurityRuleProtocolAsterisk,
		SourceApplicationSecurityGroupNames:      []*string{new(b.NameForApplicationSecurityGroupNodes())},
		SourcePortRange:                          new("*"),
		DestinationApplicationSecurityGroupNames: []*string{new(b.NameForApplicationSecurityGroupControlPlane())},
		DestinationPortRange:                     new("*"),
	})
	if b.Cluster.UsesLoadBalancerForKopsController() && b.Cluster.Spec.API.LoadBalancer != nil && b.Cluster.Spec.API.LoadBalancer.Type == kops.LoadBalancerTypePublic {
		// Node traffic to the public load balancer frontend egresses through the NAT gateway, so it
		// arrives with the NAT gateway public IP as its source and cannot be matched by the nodes ASG.
		nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
			Name:                                     new("AllowNodesToKubernetesAPI"),
			Priority:                                 new(int32(2000)),
			Access:                                   network.SecurityRuleAccessAllow,
			Direction:                                network.SecurityRuleDirectionInbound,
			Protocol:                                 network.SecurityRuleProtocolTCP,
			SourcePublicIPAddress:                    ngwPipTask,
			SourcePortRange:                          new("*"),
			DestinationApplicationSecurityGroupNames: []*string{new(b.NameForApplicationSecurityGroupControlPlane())},
			DestinationPortRange:                     new(strconv.Itoa(wellknownports.KubeAPIServer)),
		})
		nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
			Name:                                     new("AllowNodesToKopsController"),
			Priority:                                 new(int32(2001)),
			Access:                                   network.SecurityRuleAccessAllow,
			Direction:                                network.SecurityRuleDirectionInbound,
			Protocol:                                 network.SecurityRuleProtocolTCP,
			SourcePublicIPAddress:                    ngwPipTask,
			SourcePortRange:                          new("*"),
			DestinationApplicationSecurityGroupNames: []*string{new(b.NameForApplicationSecurityGroupControlPlane())},
			DestinationPortRange:                     new(strconv.Itoa(wellknownports.KopsControllerPort)),
		})
	}
	nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
		Name:                     new("AllowAzureLoadBalancer"),
		Priority:                 new(int32(4000)),
		Access:                   network.SecurityRuleAccessAllow,
		Direction:                network.SecurityRuleDirectionInbound,
		Protocol:                 network.SecurityRuleProtocolAsterisk,
		SourceAddressPrefix:      new("AzureLoadBalancer"),
		SourcePortRange:          new("*"),
		DestinationAddressPrefix: new("VirtualNetwork"),
		DestinationPortRange:     new("*"),
	})
	nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
		Name:                                     new("DenyAllToControlPlane"),
		Priority:                                 new(int32(4001)),
		Access:                                   network.SecurityRuleAccessDeny,
		Direction:                                network.SecurityRuleDirectionInbound,
		Protocol:                                 network.SecurityRuleProtocolAsterisk,
		SourceAddressPrefix:                      new("*"),
		SourcePortRange:                          new("*"),
		DestinationApplicationSecurityGroupNames: []*string{new(b.NameForApplicationSecurityGroupControlPlane())},
		DestinationPortRange:                     new("*"),
	})
	nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
		Name:                                     new("DenyAllToNodes"),
		Priority:                                 new(int32(4002)),
		Access:                                   network.SecurityRuleAccessDeny,
		Direction:                                network.SecurityRuleDirectionInbound,
		Protocol:                                 network.SecurityRuleProtocolAsterisk,
		SourceAddressPrefix:                      new("*"),
		SourcePortRange:                          new("*"),
		DestinationApplicationSecurityGroupNames: []*string{new(b.NameForApplicationSecurityGroupNodes())},
		DestinationPortRange:                     new("*"),
	})
	c.AddTask(nsgTask)

	ngwTask := &azuretasks.NatGateway{
		Name:              new(b.NameForVirtualNetwork()),
		Lifecycle:         b.Lifecycle,
		PublicIPAddresses: []*azuretasks.PublicIPAddress{ngwPipTask},
		ResourceGroup:     b.LinkToResourceGroup(),
		SKU:               network.NatGatewaySKUNameStandard,
		Tags:              map[string]*string{},
	}
	c.AddTask(ngwTask)

	rtTask := &azuretasks.RouteTable{
		Name:          new(b.NameForRouteTable()),
		Lifecycle:     b.Lifecycle,
		ResourceGroup: b.LinkToResourceGroup(),
		Tags:          map[string]*string{},
		Shared:        new(b.Cluster.IsSharedAzureRouteTable()),
	}
	c.AddTask(rtTask)

	for _, subnetSpec := range b.Cluster.Spec.Networking.Subnets {
		subnetTask := &azuretasks.Subnet{
			Name:                 new(subnetSpec.Name),
			Lifecycle:            b.Lifecycle,
			ResourceGroup:        b.LinkToResourceGroup(),
			VirtualNetwork:       b.LinkToVirtualNetwork(),
			NatGateway:           ngwTask,
			NetworkSecurityGroup: nsgTask,
			RouteTable:           rtTask,
			CIDR:                 new(subnetSpec.CIDR),
			Shared:               new(b.Cluster.SharedVPC()),
		}
		c.AddTask(subnetTask)
	}

	return nil
}

func ipv4CIDRs(mixedCIDRs []string) []*string {
	var cidrs []*string
	for i := range mixedCIDRs {
		cidr := mixedCIDRs[i]
		if net.IPFamilyOfCIDRString(cidr) == net.IPv4 {
			cidrs = append(cidrs, &cidr)
		}
	}
	return cidrs
}

func ipv6CIDRs(mixedCIDRs []string) []*string {
	var cidrs []*string
	for i := range mixedCIDRs {
		cidr := mixedCIDRs[i]
		if net.IPFamilyOfCIDRString(cidr) == net.IPv6 {
			cidrs = append(cidrs, &cidr)
		}
	}
	return cidrs
}
