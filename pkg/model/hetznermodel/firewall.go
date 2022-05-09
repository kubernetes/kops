/*
Copyright 2022 The Kubernetes Authors.

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

package hetznermodel

import (
	"fmt"
	"net"
	"strings"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetznertasks"
)

// ExternalAccessModelBuilder configures Firewall objects
type ExternalAccessModelBuilder struct {
	*HetznerModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.ModelBuilder = &ExternalAccessModelBuilder{}

func (b *ExternalAccessModelBuilder) Build(c *fi.ModelBuilderContext) error {
	var sshAccess []net.IPNet
	for _, cidr := range b.Cluster.Spec.SSHAccess {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return err
		}
		sshAccess = append(sshAccess, *ipNet)
	}
	controlPlaneLabelSelector := []string{
		fmt.Sprintf("%s=%s", hetzner.TagKubernetesClusterName, b.ClusterName()),
		fmt.Sprintf("%s=%s", hetzner.TagKubernetesInstanceRole, string(kops.InstanceGroupRoleMaster)),
	}
	controlPlaneFirewall := &hetznertasks.Firewall{
		Name:      fi.String("control-plane." + b.ClusterName()),
		Lifecycle: b.Lifecycle,
		Selector:  strings.Join(controlPlaneLabelSelector, ","),
		Rules: []*hetznertasks.FirewallRule{
			{
				Direction: string(hcloud.FirewallRuleDirectionIn),
				SourceIPs: sshAccess,
				Protocol:  string(hcloud.FirewallRuleProtocolTCP),
				Port:      fi.String("22"),
			},
		},
		Labels: map[string]string{
			hetzner.TagKubernetesClusterName:  b.ClusterName(),
			hetzner.TagKubernetesFirewallRole: "control-plane",
		},
	}
	nodesLabelSelector := []string{
		fmt.Sprintf("%s=%s", hetzner.TagKubernetesClusterName, b.ClusterName()),
		fmt.Sprintf("%s=%s", hetzner.TagKubernetesInstanceRole, string(kops.InstanceGroupRoleNode)),
	}
	nodesFirewall := &hetznertasks.Firewall{
		Name:      fi.String("nodes." + b.ClusterName()),
		Lifecycle: b.Lifecycle,
		Selector:  strings.Join(nodesLabelSelector, ","),
		Rules: []*hetznertasks.FirewallRule{
			{
				Direction: string(hcloud.FirewallRuleDirectionIn),
				SourceIPs: sshAccess,
				Protocol:  string(hcloud.FirewallRuleProtocolTCP),
				Port:      fi.String("22"),
			},
		},
		Labels: map[string]string{
			hetzner.TagKubernetesClusterName:  b.ClusterName(),
			hetzner.TagKubernetesFirewallRole: "nodes",
		},
	}

	if !b.UseLoadBalancerForAPI() {
		var apiAccess []net.IPNet
		for _, cidr := range b.Cluster.Spec.KubernetesAPIAccess {
			_, ipNet, err := net.ParseCIDR(cidr)
			if err != nil {
				return err
			}
			apiAccess = append(apiAccess, *ipNet)
		}
		controlPlaneFirewall.Rules = append(controlPlaneFirewall.Rules, &hetznertasks.FirewallRule{
			Direction: string(hcloud.FirewallRuleDirectionIn),
			SourceIPs: apiAccess,
			Protocol:  string(hcloud.FirewallRuleProtocolTCP),
			Port:      fi.String("443"),
		})
	}

	if len(b.Cluster.Spec.NodePortAccess) > 0 {
		var nodePortAccess []net.IPNet
		for _, cidr := range b.Cluster.Spec.NodePortAccess {
			_, ipNet, err := net.ParseCIDR(cidr)
			if err != nil {
				return err
			}
			nodePortAccess = append(nodePortAccess, *ipNet)
		}
		nodePortRange, err := b.NodePortRange()
		if err != nil {
			return err
		}
		nodesFirewall.Rules = append(nodesFirewall.Rules, &hetznertasks.FirewallRule{
			Direction: string(hcloud.FirewallRuleDirectionIn),
			SourceIPs: nodePortAccess,
			Protocol:  string(hcloud.FirewallRuleProtocolTCP),
			Port:      fi.String(fmt.Sprintf("%d-%d", nodePortRange.Base, nodePortRange.Base+nodePortRange.Size-1)),
		})
		nodesFirewall.Rules = append(nodesFirewall.Rules, &hetznertasks.FirewallRule{
			Direction: string(hcloud.FirewallRuleDirectionIn),
			SourceIPs: nodePortAccess,
			Protocol:  string(hcloud.FirewallRuleProtocolTCP),
			Port:      fi.String(fmt.Sprintf("%d-%d", nodePortRange.Base, nodePortRange.Base+nodePortRange.Size-1)),
		})
	}

	c.AddTask(controlPlaneFirewall)
	c.AddTask(nodesFirewall)

	return nil
}
