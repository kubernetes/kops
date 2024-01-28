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
		Name:          fi.PtrTo(b.NameForVirtualNetwork()),
		Lifecycle:     b.Lifecycle,
		ResourceGroup: b.LinkToResourceGroup(),
		CIDR:          fi.PtrTo(b.Cluster.Spec.Networking.NetworkCIDR),
		Tags:          map[string]*string{},
		Shared:        fi.PtrTo(b.Cluster.SharedVPC()),
	}
	c.AddTask(networkTask)

	nsgTask := &azuretasks.NetworkSecurityGroup{
		Name:          fi.PtrTo(b.NameForVirtualNetwork()),
		Lifecycle:     b.Lifecycle,
		ResourceGroup: b.LinkToResourceGroup(),
		Tags:          map[string]*string{},
	}
	sshAccessIPv4 := ipv4CIDRs(b.Cluster.Spec.SSHAccess)
	if len(sshAccessIPv4) > 0 {
		nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
			Name:                  fi.PtrTo("AllowSSH"),
			Priority:              fi.PtrTo[int32](100),
			Access:                network.SecurityRuleAccessAllow,
			Direction:             network.SecurityRuleDirectionInbound,
			Protocol:              network.SecurityRuleProtocolTCP,
			SourceAddressPrefixes: sshAccessIPv4,
			SourcePortRange:       fi.PtrTo("*"),
			DestinationApplicationSecurityGroupNames: []*string{
				fi.PtrTo(b.NameForApplicationSecurityGroupControlPlane()),
				fi.PtrTo(b.NameForApplicationSecurityGroupNodes()),
			},
			DestinationPortRange: fi.PtrTo("22"),
		})
	}
	sshAccessIPv6 := ipv6CIDRs(b.Cluster.Spec.SSHAccess)
	if len(sshAccessIPv6) > 0 {
		nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
			Name:                  fi.PtrTo("AllowSSH_v6"),
			Priority:              fi.PtrTo[int32](101),
			Access:                network.SecurityRuleAccessAllow,
			Direction:             network.SecurityRuleDirectionInbound,
			Protocol:              network.SecurityRuleProtocolTCP,
			SourceAddressPrefixes: sshAccessIPv6,
			SourcePortRange:       fi.PtrTo("*"),
			DestinationApplicationSecurityGroupNames: []*string{
				fi.PtrTo(b.NameForApplicationSecurityGroupControlPlane()),
				fi.PtrTo(b.NameForApplicationSecurityGroupNodes()),
			},
			DestinationPortRange: fi.PtrTo("22"),
		})
	}
	k8sAccessIPv4 := ipv4CIDRs(b.Cluster.Spec.API.Access)
	if len(k8sAccessIPv4) > 0 {
		nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
			Name:                                     fi.PtrTo("AllowKubernetesAPI"),
			Priority:                                 fi.PtrTo[int32](200),
			Access:                                   network.SecurityRuleAccessAllow,
			Direction:                                network.SecurityRuleDirectionInbound,
			Protocol:                                 network.SecurityRuleProtocolTCP,
			SourceAddressPrefixes:                    k8sAccessIPv4,
			SourcePortRange:                          fi.PtrTo("*"),
			DestinationApplicationSecurityGroupNames: []*string{fi.PtrTo(b.NameForApplicationSecurityGroupControlPlane())},
			DestinationPortRange:                     fi.PtrTo(strconv.Itoa(wellknownports.KubeAPIServer)),
		})
	}
	k8sAccessIPv6 := ipv6CIDRs(b.Cluster.Spec.API.Access)
	if len(k8sAccessIPv6) > 0 {
		nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
			Name:                                     fi.PtrTo("AllowKubernetesAPI_v6"),
			Priority:                                 fi.PtrTo[int32](201),
			Access:                                   network.SecurityRuleAccessAllow,
			Direction:                                network.SecurityRuleDirectionInbound,
			Protocol:                                 network.SecurityRuleProtocolTCP,
			SourceAddressPrefixes:                    k8sAccessIPv6,
			SourcePortRange:                          fi.PtrTo("*"),
			DestinationApplicationSecurityGroupNames: []*string{fi.PtrTo(b.NameForApplicationSecurityGroupControlPlane())},
			DestinationPortRange:                     fi.PtrTo(strconv.Itoa(wellknownports.KubeAPIServer)),
		})
	}
	nodePortAccessIPv4 := ipv4CIDRs(b.Cluster.Spec.NodePortAccess)
	if len(nodePortAccessIPv4) > 0 {
		nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
			Name:                                     fi.PtrTo("AllowNodePortTCP"),
			Priority:                                 fi.PtrTo[int32](300),
			Access:                                   network.SecurityRuleAccessAllow,
			Direction:                                network.SecurityRuleDirectionInbound,
			Protocol:                                 network.SecurityRuleProtocolAsterisk,
			SourceAddressPrefixes:                    nodePortAccessIPv4,
			SourcePortRange:                          fi.PtrTo("*"),
			DestinationApplicationSecurityGroupNames: []*string{fi.PtrTo(b.NameForApplicationSecurityGroupNodes())},
			DestinationPortRange:                     fi.PtrTo("30000-32767"),
		})
	}
	nodePortAccessIPv6 := ipv6CIDRs(b.Cluster.Spec.NodePortAccess)
	if len(nodePortAccessIPv6) > 0 {
		nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
			Name:                                     fi.PtrTo("AllowNodePortTCP_v6"),
			Priority:                                 fi.PtrTo[int32](301),
			Access:                                   network.SecurityRuleAccessAllow,
			Direction:                                network.SecurityRuleDirectionInbound,
			Protocol:                                 network.SecurityRuleProtocolAsterisk,
			SourceAddressPrefixes:                    nodePortAccessIPv6,
			SourcePortRange:                          fi.PtrTo("*"),
			DestinationApplicationSecurityGroupNames: []*string{fi.PtrTo(b.NameForApplicationSecurityGroupNodes())},
			DestinationPortRange:                     fi.PtrTo("30000-32767"),
		})
	}
	nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
		Name:                                     fi.PtrTo("AllowControlPlaneToControlPlane"),
		Priority:                                 fi.PtrTo[int32](1000),
		Access:                                   network.SecurityRuleAccessAllow,
		Direction:                                network.SecurityRuleDirectionInbound,
		Protocol:                                 network.SecurityRuleProtocolAsterisk,
		SourceApplicationSecurityGroupNames:      []*string{fi.PtrTo(b.NameForApplicationSecurityGroupControlPlane())},
		SourcePortRange:                          fi.PtrTo("*"),
		DestinationApplicationSecurityGroupNames: []*string{fi.PtrTo(b.NameForApplicationSecurityGroupControlPlane())},
		DestinationPortRange:                     fi.PtrTo("*"),
	})
	nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
		Name:                                     fi.PtrTo("AllowControlPlaneToNodes"),
		Priority:                                 fi.PtrTo[int32](1001),
		Access:                                   network.SecurityRuleAccessAllow,
		Direction:                                network.SecurityRuleDirectionInbound,
		Protocol:                                 network.SecurityRuleProtocolAsterisk,
		SourceApplicationSecurityGroupNames:      []*string{fi.PtrTo(b.NameForApplicationSecurityGroupControlPlane())},
		SourcePortRange:                          fi.PtrTo("*"),
		DestinationApplicationSecurityGroupNames: []*string{fi.PtrTo(b.NameForApplicationSecurityGroupNodes())},
		DestinationPortRange:                     fi.PtrTo("*"),
	})
	nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
		Name:                                     fi.PtrTo("AllowNodesToNodes"),
		Priority:                                 fi.PtrTo[int32](1002),
		Access:                                   network.SecurityRuleAccessAllow,
		Direction:                                network.SecurityRuleDirectionInbound,
		Protocol:                                 network.SecurityRuleProtocolAsterisk,
		SourceApplicationSecurityGroupNames:      []*string{fi.PtrTo(b.NameForApplicationSecurityGroupNodes())},
		SourcePortRange:                          fi.PtrTo("*"),
		DestinationApplicationSecurityGroupNames: []*string{fi.PtrTo(b.NameForApplicationSecurityGroupNodes())},
		DestinationPortRange:                     fi.PtrTo("*"),
	})
	nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
		Name:                                     fi.PtrTo("DenyNodesToEtcdManager"),
		Priority:                                 fi.PtrTo[int32](1003),
		Access:                                   network.SecurityRuleAccessDeny,
		Direction:                                network.SecurityRuleDirectionInbound,
		Protocol:                                 network.SecurityRuleProtocolTCP,
		SourceApplicationSecurityGroupNames:      []*string{fi.PtrTo(b.NameForApplicationSecurityGroupNodes())},
		SourcePortRange:                          fi.PtrTo("*"),
		DestinationApplicationSecurityGroupNames: []*string{fi.PtrTo(b.NameForApplicationSecurityGroupControlPlane())},
		DestinationPortRange:                     fi.PtrTo("2380-2381"),
	})
	nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
		Name:                                     fi.PtrTo("DenyNodesToEtcd"),
		Priority:                                 fi.PtrTo[int32](1004),
		Access:                                   network.SecurityRuleAccessDeny,
		Direction:                                network.SecurityRuleDirectionInbound,
		Protocol:                                 network.SecurityRuleProtocolTCP,
		SourceApplicationSecurityGroupNames:      []*string{fi.PtrTo(b.NameForApplicationSecurityGroupNodes())},
		SourcePortRange:                          fi.PtrTo("*"),
		DestinationApplicationSecurityGroupNames: []*string{fi.PtrTo(b.NameForApplicationSecurityGroupControlPlane())},
		DestinationPortRange:                     fi.PtrTo("4000-4001"),
	})
	nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
		Name:                                     fi.PtrTo("AllowNodesToControlPlane"),
		Priority:                                 fi.PtrTo[int32](1005),
		Access:                                   network.SecurityRuleAccessAllow,
		Direction:                                network.SecurityRuleDirectionInbound,
		Protocol:                                 network.SecurityRuleProtocolAsterisk,
		SourceApplicationSecurityGroupNames:      []*string{fi.PtrTo(b.NameForApplicationSecurityGroupNodes())},
		SourcePortRange:                          fi.PtrTo("*"),
		DestinationApplicationSecurityGroupNames: []*string{fi.PtrTo(b.NameForApplicationSecurityGroupControlPlane())},
		DestinationPortRange:                     fi.PtrTo("*"),
	})
	if b.Cluster.UsesNoneDNS() && b.Cluster.Spec.API.LoadBalancer != nil && b.Cluster.Spec.API.LoadBalancer.Type == kops.LoadBalancerTypePublic {
		// TODO: Limit access to necessary source address prefixes instead of "0.0.0.0/0" and "::/0"
		nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
			Name:                                     fi.PtrTo("AllowNodesToKubernetesAPI"),
			Priority:                                 fi.PtrTo[int32](2000),
			Access:                                   network.SecurityRuleAccessAllow,
			Direction:                                network.SecurityRuleDirectionInbound,
			Protocol:                                 network.SecurityRuleProtocolTCP,
			SourceAddressPrefix:                      fi.PtrTo("*"),
			SourcePortRange:                          fi.PtrTo("*"),
			DestinationApplicationSecurityGroupNames: []*string{fi.PtrTo(b.NameForApplicationSecurityGroupControlPlane())},
			DestinationPortRange:                     fi.PtrTo(strconv.Itoa(wellknownports.KubeAPIServer)),
		})
		nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
			Name:                                     fi.PtrTo("AllowNodesToKopsController"),
			Priority:                                 fi.PtrTo[int32](2001),
			Access:                                   network.SecurityRuleAccessAllow,
			Direction:                                network.SecurityRuleDirectionInbound,
			Protocol:                                 network.SecurityRuleProtocolTCP,
			SourceAddressPrefix:                      fi.PtrTo("*"),
			SourcePortRange:                          fi.PtrTo("*"),
			DestinationApplicationSecurityGroupNames: []*string{fi.PtrTo(b.NameForApplicationSecurityGroupControlPlane())},
			DestinationPortRange:                     fi.PtrTo(strconv.Itoa(wellknownports.KopsControllerPort)),
		})
	}
	nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
		Name:                     fi.PtrTo("AllowAzureLoadBalancer"),
		Priority:                 fi.PtrTo[int32](4000),
		Access:                   network.SecurityRuleAccessAllow,
		Direction:                network.SecurityRuleDirectionInbound,
		Protocol:                 network.SecurityRuleProtocolAsterisk,
		SourceAddressPrefix:      fi.PtrTo("AzureLoadBalancer"),
		SourcePortRange:          fi.PtrTo("*"),
		DestinationAddressPrefix: fi.PtrTo("VirtualNetwork"),
		DestinationPortRange:     fi.PtrTo("*"),
	})
	nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
		Name:                                     fi.PtrTo("DenyAllToControlPlane"),
		Priority:                                 fi.PtrTo[int32](4001),
		Access:                                   network.SecurityRuleAccessDeny,
		Direction:                                network.SecurityRuleDirectionInbound,
		Protocol:                                 network.SecurityRuleProtocolAsterisk,
		SourceAddressPrefix:                      fi.PtrTo("*"),
		SourcePortRange:                          fi.PtrTo("*"),
		DestinationApplicationSecurityGroupNames: []*string{fi.PtrTo(b.NameForApplicationSecurityGroupControlPlane())},
		DestinationPortRange:                     fi.PtrTo("*"),
	})
	nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
		Name:                                     fi.PtrTo("DenyAllToNodes"),
		Priority:                                 fi.PtrTo[int32](4002),
		Access:                                   network.SecurityRuleAccessDeny,
		Direction:                                network.SecurityRuleDirectionInbound,
		Protocol:                                 network.SecurityRuleProtocolAsterisk,
		SourceAddressPrefix:                      fi.PtrTo("*"),
		SourcePortRange:                          fi.PtrTo("*"),
		DestinationApplicationSecurityGroupNames: []*string{fi.PtrTo(b.NameForApplicationSecurityGroupNodes())},
		DestinationPortRange:                     fi.PtrTo("*"),
	})
	c.AddTask(nsgTask)

	ngwPipTask := &azuretasks.PublicIPAddress{
		Name:          fi.PtrTo(b.NameForVirtualNetwork()),
		Lifecycle:     b.Lifecycle,
		ResourceGroup: b.LinkToResourceGroup(),
		Tags:          map[string]*string{},
	}
	c.AddTask(ngwPipTask)
	ngwTask := &azuretasks.NatGateway{
		Name:              fi.PtrTo(b.NameForVirtualNetwork()),
		Lifecycle:         b.Lifecycle,
		PublicIPAddresses: []*azuretasks.PublicIPAddress{ngwPipTask},
		ResourceGroup:     b.LinkToResourceGroup(),
		Tags:              map[string]*string{},
	}
	c.AddTask(ngwTask)

	for _, subnetSpec := range b.Cluster.Spec.Networking.Subnets {
		subnetTask := &azuretasks.Subnet{
			Name:                 fi.PtrTo(subnetSpec.Name),
			Lifecycle:            b.Lifecycle,
			ResourceGroup:        b.LinkToResourceGroup(),
			VirtualNetwork:       b.LinkToVirtualNetwork(),
			NatGateway:           ngwTask,
			NetworkSecurityGroup: nsgTask,
			CIDR:                 fi.PtrTo(subnetSpec.CIDR),
			Shared:               fi.PtrTo(b.Cluster.SharedVPC()),
		}
		c.AddTask(subnetTask)
	}

	rtTask := &azuretasks.RouteTable{
		Name:          fi.PtrTo(b.NameForRouteTable()),
		Lifecycle:     b.Lifecycle,
		ResourceGroup: b.LinkToResourceGroup(),
		Tags:          map[string]*string{},
		Shared:        fi.PtrTo(b.Cluster.IsSharedAzureRouteTable()),
	}
	c.AddTask(rtTask)

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
