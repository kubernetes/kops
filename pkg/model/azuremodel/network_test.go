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
	"slices"
	"strconv"
	"testing"

	network "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azuretasks"
)

func TestNetworkModelBuilder_Build(t *testing.T) {
	b := NetworkModelBuilder{
		AzureModelContext: newTestAzureModelContext(),
	}
	c := &fi.CloudupModelBuilderContext{
		Tasks: make(map[string]fi.CloudupTask),
	}
	err := b.Build(c)
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}
}

// TestNetworkModelBuilder_KindnetSecurityRules verifies the pod-CIDR NSG rules that
// Kindnet needs. Kindnet preserves pod source IPs, which aren't NIC-assigned and so
// don't match the ASG-based node rules. Without an explicit allow to the kube-apiserver,
// pods such as coredns can never reach the API and never become ready.
func TestNetworkModelBuilder_KindnetSecurityRules(t *testing.T) {
	// Without Kindnet, pod traffic is SNAT'd to the node IP and matches the node ASG
	// rules, so no pod-CIDR rules should be emitted.
	nsg := buildNetworkSecurityGroup(t, newTestAzureModelContext())
	if rule := findSecurityRule(nsg, "AllowPodCIDRToKubernetesAPI"); rule != nil {
		t.Errorf("did not expect AllowPodCIDRToKubernetesAPI rule without Kindnet")
	}

	ctx := newTestAzureModelContext()
	ctx.Cluster.Spec.Networking.Kindnet = &kops.KindnetNetworkingSpec{}
	ctx.Cluster.Spec.Networking.PodCIDR = "100.96.0.0/11"
	nsg = buildNetworkSecurityGroup(t, ctx)

	rule := findSecurityRule(nsg, "AllowPodCIDRToKubernetesAPI")
	if rule == nil {
		t.Fatal("expected AllowPodCIDRToKubernetesAPI security rule for Kindnet")
	}
	if got, want := fi.ValueOf(rule.Priority), int32(1007); got != want {
		t.Errorf("priority: got %d, want %d", got, want)
	}
	if got, want := rule.Protocol, network.SecurityRuleProtocolTCP; got != want {
		t.Errorf("protocol: got %q, want %q", got, want)
	}
	if got, want := fi.ValueOf(rule.SourceAddressPrefix), ctx.Cluster.Spec.Networking.PodCIDR; got != want {
		t.Errorf("source address prefix: got %q, want %q", got, want)
	}
	if got, want := fi.ValueOf(rule.DestinationPortRange), strconv.Itoa(wellknownports.KubeAPIServer); got != want {
		t.Errorf("destination port range: got %q, want %q", got, want)
	}
	var dstASGs []string
	for _, name := range rule.DestinationApplicationSecurityGroupNames {
		dstASGs = append(dstASGs, fi.ValueOf(name))
	}
	if want := []string{ctx.NameForApplicationSecurityGroupControlPlane()}; !slices.Equal(dstASGs, want) {
		t.Errorf("destination ASGs: got %v, want %v", dstASGs, want)
	}
}

func buildNetworkSecurityGroup(t *testing.T, ctx *AzureModelContext) *azuretasks.NetworkSecurityGroup {
	t.Helper()
	b := NetworkModelBuilder{AzureModelContext: ctx}
	c := &fi.CloudupModelBuilderContext{
		Tasks: make(map[string]fi.CloudupTask),
	}
	if err := b.Build(c); err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	for _, task := range c.Tasks {
		if nsg, ok := task.(*azuretasks.NetworkSecurityGroup); ok {
			return nsg
		}
	}
	t.Fatal("no NetworkSecurityGroup task was built")
	return nil
}

func findSecurityRule(nsg *azuretasks.NetworkSecurityGroup, name string) *azuretasks.NetworkSecurityRule {
	for _, rule := range nsg.SecurityRules {
		if fi.ValueOf(rule.Name) == name {
			return rule
		}
	}
	return nil
}
