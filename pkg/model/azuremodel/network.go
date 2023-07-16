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
	"fmt"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-05-01/network"
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
	var sshAccessIPv4, sshAccessIPv6 []string
	for _, cidr := range b.Cluster.Spec.SSHAccess {
		switch net.IPFamilyOfCIDRString(cidr) {
		case net.IPv4:
			sshAccessIPv4 = append(sshAccessIPv4, cidr)
		case net.IPv6:
			sshAccessIPv6 = append(sshAccessIPv6, cidr)
		default:
			return fmt.Errorf("unknown IP family for CIDR: %q", cidr)
		}
	}
	if len(sshAccessIPv4) > 0 {
		nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
			Name:                     fi.PtrTo("AllowSSH"),
			Priority:                 fi.PtrTo[int32](100),
			Access:                   network.SecurityRuleAccessAllow,
			Direction:                network.SecurityRuleDirectionInbound,
			Protocol:                 network.SecurityRuleProtocolTCP,
			SourceAddressPrefixes:    &sshAccessIPv4,
			SourcePortRange:          fi.PtrTo("*"),
			DestinationAddressPrefix: fi.PtrTo("*"),
			DestinationPortRange:     fi.PtrTo("22"),
		})
	}
	if len(sshAccessIPv6) > 0 {
		nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
			Name:                     fi.PtrTo("AllowSSH_v6"),
			Priority:                 fi.PtrTo[int32](101),
			Access:                   network.SecurityRuleAccessAllow,
			Direction:                network.SecurityRuleDirectionInbound,
			Protocol:                 network.SecurityRuleProtocolTCP,
			SourceAddressPrefixes:    &sshAccessIPv6,
			SourcePortRange:          fi.PtrTo("*"),
			DestinationAddressPrefix: fi.PtrTo("*"),
			DestinationPortRange:     fi.PtrTo("22"),
		})
	}
	var k8sAccessIPv4, k8sAccessIPv6 []string
	for _, cidr := range b.Cluster.Spec.API.Access {
		switch net.IPFamilyOfCIDRString(cidr) {
		case net.IPv4:
			k8sAccessIPv4 = append(k8sAccessIPv4, cidr)
		case net.IPv6:
			k8sAccessIPv6 = append(k8sAccessIPv6, cidr)
		default:
			return fmt.Errorf("unknown IP family for CIDR: %q", cidr)
		}
	}
	if len(k8sAccessIPv4) > 0 {
		nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
			Name:                     fi.PtrTo("AllowKubernetesAPI"),
			Priority:                 fi.PtrTo[int32](200),
			Access:                   network.SecurityRuleAccessAllow,
			Direction:                network.SecurityRuleDirectionInbound,
			Protocol:                 network.SecurityRuleProtocolTCP,
			SourceAddressPrefixes:    &k8sAccessIPv4,
			SourcePortRange:          fi.PtrTo("*"),
			DestinationAddressPrefix: fi.PtrTo("*"),
			DestinationPortRange:     fi.PtrTo(strconv.Itoa(wellknownports.KubeAPIServer)),
		})
	}
	if len(k8sAccessIPv6) > 0 {
		nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
			Name:                     fi.PtrTo("AllowKubernetesAPI_v6"),
			Priority:                 fi.PtrTo[int32](201),
			Access:                   network.SecurityRuleAccessAllow,
			Direction:                network.SecurityRuleDirectionInbound,
			Protocol:                 network.SecurityRuleProtocolTCP,
			SourceAddressPrefixes:    &k8sAccessIPv6,
			SourcePortRange:          fi.PtrTo("*"),
			DestinationAddressPrefix: fi.PtrTo("*"),
			DestinationPortRange:     fi.PtrTo(strconv.Itoa(wellknownports.KubeAPIServer)),
		})
	}
	if b.Cluster.UsesNoneDNS() {
		if b.Cluster.Spec.API.LoadBalancer != nil && b.Cluster.Spec.API.LoadBalancer.Type == kops.LoadBalancerTypeInternal {
			nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
				Name:                     fi.PtrTo("AllowKopsController"),
				Priority:                 fi.PtrTo[int32](210),
				Access:                   network.SecurityRuleAccessAllow,
				Direction:                network.SecurityRuleDirectionInbound,
				Protocol:                 network.SecurityRuleProtocolTCP,
				SourceAddressPrefix:      fi.PtrTo(b.Cluster.Spec.Networking.NetworkCIDR),
				SourcePortRange:          fi.PtrTo("*"),
				DestinationAddressPrefix: fi.PtrTo("*"),
				DestinationPortRange:     fi.PtrTo(strconv.Itoa(wellknownports.KopsControllerPort)),
			})
		} else {
			// TODO: Limit access to necessary source address prefixes instead of "0.0.0.0/0" and "::/0"
			nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
				Name:                     fi.PtrTo("AllowKopsController"),
				Priority:                 fi.PtrTo[int32](210),
				Access:                   network.SecurityRuleAccessAllow,
				Direction:                network.SecurityRuleDirectionInbound,
				Protocol:                 network.SecurityRuleProtocolTCP,
				SourceAddressPrefix:      fi.PtrTo("0.0.0.0/0"),
				SourcePortRange:          fi.PtrTo("*"),
				DestinationAddressPrefix: fi.PtrTo("*"),
				DestinationPortRange:     fi.PtrTo(strconv.Itoa(wellknownports.KopsControllerPort)),
			})
			nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
				Name:                     fi.PtrTo("AllowKopsController_v6"),
				Priority:                 fi.PtrTo[int32](211),
				Access:                   network.SecurityRuleAccessAllow,
				Direction:                network.SecurityRuleDirectionInbound,
				Protocol:                 network.SecurityRuleProtocolTCP,
				SourceAddressPrefix:      fi.PtrTo("::/0"),
				SourcePortRange:          fi.PtrTo("*"),
				DestinationAddressPrefix: fi.PtrTo("*"),
				DestinationPortRange:     fi.PtrTo(strconv.Itoa(wellknownports.KopsControllerPort)),
			})
		}
	}
	var nodePortAccessIPv4, nodePortAccessIPv6 []string
	for _, cidr := range b.Cluster.Spec.NodePortAccess {
		switch net.IPFamilyOfCIDRString(cidr) {
		case net.IPv4:
			nodePortAccessIPv4 = append(nodePortAccessIPv4, cidr)
		case net.IPv6:
			nodePortAccessIPv6 = append(nodePortAccessIPv6, cidr)
		default:
			return fmt.Errorf("unknown IP family for CIDR: %q", cidr)
		}
	}
	if len(nodePortAccessIPv4) > 0 {
		nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
			Name:                     fi.PtrTo("AllowNodePort"),
			Priority:                 fi.PtrTo[int32](300),
			Access:                   network.SecurityRuleAccessAllow,
			Direction:                network.SecurityRuleDirectionInbound,
			Protocol:                 network.SecurityRuleProtocolTCP,
			SourceAddressPrefixes:    &nodePortAccessIPv4,
			SourcePortRange:          fi.PtrTo("*"),
			DestinationAddressPrefix: fi.PtrTo("*"),
			DestinationPortRange:     fi.PtrTo("443"),
		})
	}
	if len(nodePortAccessIPv6) > 0 {
		nsgTask.SecurityRules = append(nsgTask.SecurityRules, &azuretasks.NetworkSecurityRule{
			Name:                     fi.PtrTo("AllowNodePort_v6"),
			Priority:                 fi.PtrTo[int32](301),
			Access:                   network.SecurityRuleAccessAllow,
			Direction:                network.SecurityRuleDirectionInbound,
			Protocol:                 network.SecurityRuleProtocolTCP,
			SourceAddressPrefixes:    &nodePortAccessIPv6,
			SourcePortRange:          fi.PtrTo("*"),
			DestinationAddressPrefix: fi.PtrTo("*"),
			DestinationPortRange:     fi.PtrTo("443"),
		})
	}
	c.AddTask(nsgTask)

	for _, subnetSpec := range b.Cluster.Spec.Networking.Subnets {
		subnetTask := &azuretasks.Subnet{
			Name:                 fi.PtrTo(subnetSpec.Name),
			Lifecycle:            b.Lifecycle,
			ResourceGroup:        b.LinkToResourceGroup(),
			VirtualNetwork:       b.LinkToVirtualNetwork(),
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
