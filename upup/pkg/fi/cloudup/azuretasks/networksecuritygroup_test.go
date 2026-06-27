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

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

func TestNetworkSecurityRuleSourcePublicIPAddress(t *testing.T) {
	pip := &PublicIPAddress{
		Name:      fi.PtrTo("nat-public-ip"),
		IPAddress: fi.PtrTo("203.0.113.10"),
	}
	rule := &NetworkSecurityRule{
		Name:                  fi.PtrTo("AllowNodesToKubernetesAPI"),
		SourcePublicIPAddress: pip,
	}

	if err := rule.resolvePublicIPAddressSourcePrefix(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.SourceAddressPrefix != nil {
		t.Fatalf("expected source address prefix to be unset, got %q", fi.ValueOf(rule.SourceAddressPrefix))
	}
	if len(rule.SourceAddressPrefixes) != 1 {
		t.Fatalf("expected one source address prefix, got %v", rule.SourceAddressPrefixes)
	}
	if got, want := fi.ValueOf(rule.SourceAddressPrefixes[0]), "203.0.113.10"; got != want {
		t.Fatalf("unexpected source address prefix: got %q, want %q", got, want)
	}
	if rule.SourcePublicIPAddress != nil {
		t.Fatalf("expected SourcePublicIPAddress to be cleared after resolving source address prefix")
	}

	rule.SourcePublicIPAddress = pip
	tfRule := rule.toTerraform()
	if tfRule.SourceAddressPrefix != nil {
		t.Fatalf("expected Terraform source address prefix to be unset, got %q", tfRule.SourceAddressPrefix.String)
	}
	if len(tfRule.SourceAddressPrefixes) != 1 {
		t.Fatalf("expected one Terraform source address prefix, got %v", tfRule.SourceAddressPrefixes)
	}
	if got, want := tfRule.SourceAddressPrefixes[0].String, "azurerm_public_ip.nat-public-ip.ip_address"; got != want {
		t.Fatalf("unexpected Terraform source address prefix: got %q, want %q", got, want)
	}
}

func TestNetworkSecurityGroupPublicIPAddressDependency(t *testing.T) {
	pip := &PublicIPAddress{
		Name: fi.PtrTo("nat-public-ip"),
	}
	nsg := &NetworkSecurityGroup{
		SecurityRules: []*NetworkSecurityRule{
			{
				Name:                  fi.PtrTo("AllowNodesToKubernetesAPI"),
				SourcePublicIPAddress: pip,
			},
		},
	}

	deps := fi.FindDependencies[fi.CloudupSubContext](map[string]fi.CloudupTask{}, nsg)
	if len(deps) != 1 || deps[0] != pip {
		t.Fatalf("expected public IP dependency, got %v", deps)
	}
}

func TestNetworkSecurityGroupNormalizeSourcePublicIPAddress(t *testing.T) {
	pip := &PublicIPAddress{
		Name:      fi.PtrTo("nat-public-ip"),
		IPAddress: fi.PtrTo("203.0.113.10"),
	}
	nsg := &NetworkSecurityGroup{
		Tags: map[string]*string{},
		SecurityRules: []*NetworkSecurityRule{
			{
				Name:                  fi.PtrTo("AllowNodesToKubernetesAPI"),
				SourcePublicIPAddress: pip,
			},
		},
	}
	cloud := NewMockAzureCloud("eastus")
	ctx := &fi.CloudupContext{
		T: fi.CloudupSubContext{
			Cloud: cloud,
		},
		Target: azure.NewAzureAPITarget(cloud),
	}

	if err := nsg.Normalize(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rule := nsg.SecurityRules[0]
	if rule.SourcePublicIPAddress != nil {
		t.Fatalf("expected SourcePublicIPAddress to be cleared after Azure API normalization")
	}
	if len(rule.SourceAddressPrefixes) != 1 {
		t.Fatalf("expected one source address prefix, got %v", rule.SourceAddressPrefixes)
	}
	if got, want := fi.ValueOf(rule.SourceAddressPrefixes[0]), "203.0.113.10"; got != want {
		t.Fatalf("unexpected source address prefix: got %q, want %q", got, want)
	}

	nsg = &NetworkSecurityGroup{
		Tags: map[string]*string{},
		SecurityRules: []*NetworkSecurityRule{
			{
				Name:                  fi.PtrTo("AllowNodesToKubernetesAPI"),
				SourcePublicIPAddress: pip,
			},
		},
	}
	ctx.Target = terraform.NewTerraformTarget(cloud, "", "", nil)
	if err := nsg.Normalize(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rule = nsg.SecurityRules[0]
	if rule.SourcePublicIPAddress != pip {
		t.Fatalf("expected SourcePublicIPAddress to remain set for Terraform")
	}
	if len(rule.SourceAddressPrefixes) != 0 {
		t.Fatalf("expected source address prefixes to remain unset for Terraform, got %v", rule.SourceAddressPrefixes)
	}
}

func TestNetworkSecurityRuleSourcePublicIPAddressRequiresIPAddress(t *testing.T) {
	pip := &PublicIPAddress{
		Name: fi.PtrTo("nat-public-ip"),
	}
	rule := &NetworkSecurityRule{
		Name:                  fi.PtrTo("AllowNodesToKubernetesAPI"),
		SourcePublicIPAddress: pip,
	}

	if err := rule.resolvePublicIPAddressSourcePrefix(); err == nil {
		t.Fatalf("expected error resolving source address prefix without assigned IP address")
	}
}
