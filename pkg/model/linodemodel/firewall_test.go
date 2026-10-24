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

package linodemodel

import (
	"regexp"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linodetasks"
)

func TestFirewallModelBuilderBuild(t *testing.T) {
	for _, test := range []struct {
		name   string
		subnet kops.SubnetType
		access []string
		rules  map[string]int
	}{
		{"public", kops.SubnetTypePublic, []string{"192.0.2.0/24", "2001:db8::/64"}, map[string]int{"control-plane": 2, "nodes": 3, "bastion": 1}},
		{"empty-access", kops.SubnetTypePublic, nil, map[string]int{"control-plane": 0, "nodes": 0, "bastion": 0}},
		{"private", kops.SubnetTypePrivate, nil, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			cluster := &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{Name: "example.k8s.local"},
				Spec: kops.ClusterSpec{
					SSHAccess: test.access, NodePortAccess: test.access,
					API: kops.APISpec{Access: test.access},
					Networking: kops.NetworkingSpec{Subnets: []kops.ClusterSubnetSpec{
						{Name: "subnet-a", Type: test.subnet},
					}},
				},
			}
			var instanceGroups []*kops.InstanceGroup
			for _, role := range []kops.InstanceGroupRole{kops.InstanceGroupRoleControlPlane, kops.InstanceGroupRoleNode, kops.InstanceGroupRoleBastion} {
				instanceGroups = append(instanceGroups, &kops.InstanceGroup{Spec: kops.InstanceGroupSpec{Role: role, Subnets: []string{"subnet-a"}}})
			}
			builder := &FirewallModelBuilder{
				LinodeModelContext: &LinodeModelContext{KopsModelContext: contextModel(cluster, instanceGroups)},
				Lifecycle:          fi.LifecycleSync,
			}
			context := &fi.CloudupModelBuilderContext{Tasks: map[string]fi.CloudupTask{}}
			if err := builder.Build(context); err != nil {
				t.Fatal(err)
			}
			if len(context.Tasks) != len(test.rules) {
				t.Fatalf("got %d firewalls, want %d", len(context.Tasks), len(test.rules))
			}
			seenRules := make(map[*linodetasks.FirewallRule]bool)
			for _, task := range context.Tasks {
				firewall := task.(*linodetasks.Firewall)
				role := strings.TrimPrefix(fi.ValueOf(firewall.Name), "example-k8s-local-")
				if count, ok := test.rules[role]; !ok || len(firewall.Rules) != count {
					t.Fatalf("unexpected rules for %q: %v", role, firewall.Rules)
				}
				for _, rule := range firewall.Rules {
					if seenRules[rule] {
						t.Fatal("firewalls must not share mutable rules")
					}
					seenRules[rule] = true
				}
			}
		})
	}
}

func TestLinodeFirewallName(t *testing.T) {
	validLabel := regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,30}[a-z0-9]$`)
	names := make(map[string]bool)
	for _, cluster := range []string{"example.k8s.local", "abcdefghijklmnopqrstuvwx.example", "example--cluster.k8s.local", "example-cluster.k8s.local"} {
		for _, role := range []string{linodeFirewallRoleControlPlane, linodeFirewallRoleNodes, linodeFirewallRoleBastion} {
			name := linodeFirewallName(cluster, role)
			if !validLabel.MatchString(name) || strings.Contains(name, "--") {
				t.Errorf("invalid firewall label %q for %q/%q", name, cluster, role)
			}
			if names[name] {
				t.Errorf("duplicate firewall label %q", name)
			}
			names[name] = true
		}
	}
}
