/*
Copyright 2026 The Kubernetes Authors.

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

package azuretasks

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	network "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

func newTestNetworkSecurityGroup(natGatewayPip *PublicIPAddress) *NetworkSecurityGroup {
	return &NetworkSecurityGroup{
		Name:      to.Ptr("nsg"),
		Lifecycle: fi.LifecycleSync,
		ResourceGroup: &ResourceGroup{
			Name: to.Ptr("rg"),
		},
		SecurityRules: []*NetworkSecurityRule{
			{
				Name:                     to.Ptr("AllowSSH"),
				Priority:                 to.Ptr[int32](100),
				Access:                   network.SecurityRuleAccessAllow,
				Direction:                network.SecurityRuleDirectionInbound,
				Protocol:                 network.SecurityRuleProtocolTCP,
				SourceAddressPrefix:      to.Ptr("*"),
				SourcePortRange:          to.Ptr("*"),
				DestinationAddressPrefix: to.Ptr("*"),
				DestinationPortRange:     to.Ptr("22"),
			},
			{
				Name:                     to.Ptr("AllowNodesToKubernetesAPI"),
				Priority:                 to.Ptr[int32](2000),
				Access:                   network.SecurityRuleAccessAllow,
				Direction:                network.SecurityRuleDirectionInbound,
				Protocol:                 network.SecurityRuleProtocolTCP,
				SourcePublicIPAddress:    natGatewayPip,
				SourcePortRange:          to.Ptr("*"),
				DestinationAddressPrefix: to.Ptr("*"),
				DestinationPortRange:     to.Ptr("443"),
			},
		},
		Tags: map[string]*string{
			testTagKey: to.Ptr(testTagValue),
		},
	}
}

func TestNetworkSecurityGroupRenderAzure(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	apiTarget := azure.NewAzureAPITarget(cloud)
	nsg := &NetworkSecurityGroup{}
	expected := newTestNetworkSecurityGroup(&PublicIPAddress{
		Name:      to.Ptr("natgw"),
		IPAddress: to.Ptr("192.0.2.1"),
	})
	if err := nsg.RenderAzure(apiTarget, nil, expected, nil); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	actual := cloud.NetworkSecurityGroupsClient.NSGs[*expected.Name]
	if a, e := *actual.Name, *expected.Name; a != e {
		t.Errorf("unexpected Name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.Properties.SecurityRules[0].Properties.SourceAddressPrefix, "*"; a != e {
		t.Errorf("unexpected SourceAddressPrefix: expected %s, but got %s", e, a)
	}
	if a, e := *actual.Properties.SecurityRules[1].Properties.SourceAddressPrefix, "192.0.2.1"; a != e {
		t.Errorf("unexpected SourceAddressPrefix: expected %s, but got %s", e, a)
	}
}

func TestNetworkSecurityGroupRenderAzureUnallocatedPublicIPAddress(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	apiTarget := azure.NewAzureAPITarget(cloud)
	nsg := &NetworkSecurityGroup{}
	expected := newTestNetworkSecurityGroup(&PublicIPAddress{
		Name: to.Ptr("natgw"),
	})
	if err := nsg.RenderAzure(apiTarget, nil, expected, nil); err == nil {
		t.Fatalf("expected error rendering a rule whose public IP has no allocated address")
	}
}

func TestNetworkSecurityGroupFind(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	ctx := &fi.CloudupContext{
		T: fi.CloudupSubContext{
			Cloud: cloud,
		},
	}

	natGatewayPip := &PublicIPAddress{
		Name:      to.Ptr("natgw"),
		IPAddress: to.Ptr("192.0.2.1"),
	}
	nsg := newTestNetworkSecurityGroup(natGatewayPip)
	nsg.SecurityRules = append(nsg.SecurityRules, &NetworkSecurityRule{
		Name:                     to.Ptr("AllowNodesToKopsController"),
		Priority:                 to.Ptr[int32](2001),
		Access:                   network.SecurityRuleAccessAllow,
		Direction:                network.SecurityRuleDirectionInbound,
		Protocol:                 network.SecurityRuleProtocolTCP,
		SourcePublicIPAddress:    natGatewayPip,
		SourcePortRange:          to.Ptr("*"),
		DestinationAddressPrefix: to.Ptr("*"),
		DestinationPortRange:     to.Ptr("3988"),
	})
	// Find will return nothing if there is no network security group created.
	actual, err := nsg.Find(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if actual != nil {
		t.Errorf("unexpected networkSecurityGroup found: %+v", actual)
	}

	cloud.NetworkSecurityGroupsClient.NSGs[*nsg.Name] = &network.SecurityGroup{
		Name: nsg.Name,
		ID:   to.Ptr("id"),
		Properties: &network.SecurityGroupPropertiesFormat{
			SecurityRules: []*network.SecurityRule{
				{
					Name: to.Ptr("AllowNodesToKubernetesAPI"),
					Properties: &network.SecurityRulePropertiesFormat{
						Priority:                 to.Ptr[int32](2000),
						Access:                   to.Ptr(network.SecurityRuleAccessAllow),
						Direction:                to.Ptr(network.SecurityRuleDirectionInbound),
						Protocol:                 to.Ptr(network.SecurityRuleProtocolTCP),
						SourceAddressPrefix:      to.Ptr("192.0.2.1"),
						SourcePortRange:          to.Ptr("*"),
						DestinationAddressPrefix: to.Ptr("*"),
						DestinationPortRange:     to.Ptr("443"),
					},
				},
				{
					Name: to.Ptr("AllowNodesToKopsController"),
					Properties: &network.SecurityRulePropertiesFormat{
						Priority:                 to.Ptr[int32](2001),
						Access:                   to.Ptr(network.SecurityRuleAccessAllow),
						Direction:                to.Ptr(network.SecurityRuleDirectionInbound),
						Protocol:                 to.Ptr(network.SecurityRuleProtocolTCP),
						SourceAddressPrefix:      to.Ptr("*"),
						SourcePortRange:          to.Ptr("*"),
						DestinationAddressPrefix: to.Ptr("*"),
						DestinationPortRange:     to.Ptr("3988"),
					},
				},
			},
		},
	}
	// Find again.
	actual, err = nsg.Find(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if a, e := *actual.Name, *nsg.Name; a != e {
		t.Errorf("unexpected networkSecurityGroup name: expected %s, but got %s", e, a)
	}
	// A source matching the referenced public IP maps back to the task reference.
	if a := actual.SecurityRules[0]; a.SourcePublicIPAddress != natGatewayPip || a.SourceAddressPrefix != nil {
		t.Errorf("expected rule source mapped to the referenced public IP, but got %+v", a)
	}
	// A non-matching source stays literal, so pre-existing wildcard rules show up as a change.
	if a := actual.SecurityRules[1]; a.SourcePublicIPAddress != nil || fi.ValueOf(a.SourceAddressPrefix) != "*" {
		t.Errorf("expected rule source to stay literal, but got %+v", a)
	}
}
