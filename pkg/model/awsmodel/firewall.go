/*
Copyright 2019 The Kubernetes Authors.

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

package awsmodel

import (
	"fmt"
	"strconv"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"

	"k8s.io/klog/v2"
)

type Protocol int

const (
	ProtocolIPIP Protocol = 4
)

// FirewallModelBuilder configures firewall network objects
type FirewallModelBuilder struct {
	*AWSModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.ModelBuilder = &FirewallModelBuilder{}

func (b *FirewallModelBuilder) Build(c *fi.ModelBuilderContext) error {
	nodeGroups, err := b.buildNodeRules(c)
	if err != nil {
		return err
	}

	masterGroups, err := b.buildMasterRules(c, nodeGroups)
	if err != nil {
		return err
	}

	// We _should_ block per port... but:
	// * It causes e2e tests to break
	// * Users expect to be able to reach pods
	// * If users are running an overlay, we punch a hole in it anyway
	// b.applyNodeToMasterAllowSpecificPorts(c)
	b.applyNodeToMasterBlockSpecificPorts(c, nodeGroups, masterGroups)

	return nil
}

func (b *FirewallModelBuilder) buildNodeRules(c *fi.ModelBuilderContext) ([]SecurityGroupInfo, error) {
	nodeGroups, err := b.GetSecurityGroups(kops.InstanceGroupRoleNode)
	if err != nil {
		return nil, err
	}

	for _, group := range nodeGroups {
		group.Task.Lifecycle = b.Lifecycle
		c.AddTask(group.Task)
	}

	for _, src := range nodeGroups {
		// Allow full egress
		{
			t := &awstasks.SecurityGroupRule{
				Name:          fi.PtrTo("ipv4-node-egress" + src.Suffix),
				Lifecycle:     b.Lifecycle,
				SecurityGroup: src.Task,
				Egress:        fi.PtrTo(true),
				CIDR:          fi.PtrTo("0.0.0.0/0"),
			}
			AddDirectionalGroupRule(c, t)
		}
		{
			t := &awstasks.SecurityGroupRule{
				Name:          fi.PtrTo("ipv6-node-egress" + src.Suffix),
				Lifecycle:     b.Lifecycle,
				SecurityGroup: src.Task,
				Egress:        fi.PtrTo(true),
				IPv6CIDR:      fi.PtrTo("::/0"),
			}
			AddDirectionalGroupRule(c, t)
		}

		// Nodes can talk to nodes
		for _, dest := range nodeGroups {
			suffix := JoinSuffixes(src, dest)

			t := &awstasks.SecurityGroupRule{
				Name:          fi.PtrTo("all-node-to-node" + suffix),
				Lifecycle:     b.Lifecycle,
				SecurityGroup: dest.Task,
				SourceGroup:   src.Task,
			}
			AddDirectionalGroupRule(c, t)
		}

	}

	return nodeGroups, nil
}

func (b *FirewallModelBuilder) applyNodeToMasterBlockSpecificPorts(c *fi.ModelBuilderContext, nodeGroups []SecurityGroupInfo, masterGroups []SecurityGroupInfo) {
	type portRange struct {
		From int
		To   int
	}

	// TODO: Make less hacky
	// TODO: Fix management - we need a wildcard matcher now
	tcpBlocked := make(map[int]bool)

	// Don't allow nodes to access etcd client port
	tcpBlocked[4001] = true
	tcpBlocked[4002] = true

	// Don't allow nodes to access etcd peer port
	tcpBlocked[2380] = true
	tcpBlocked[2381] = true

	udpRanges := []portRange{{From: 1, To: 65535}}
	protocols := []Protocol{}

	if b.Cluster.Spec.Networking.Cilium != nil && b.Cluster.Spec.Networking.Cilium.EtcdManaged {
		// Block the etcd peer port
		tcpBlocked[2382] = true
	}

	if b.Cluster.Spec.Networking.Calico != nil {
		protocols = append(protocols, ProtocolIPIP)
	}

	if b.Cluster.Spec.Networking.KubeRouter != nil {
		protocols = append(protocols, ProtocolIPIP)
	}

	tcpRanges := []portRange{
		{From: 1, To: 0},
	}
	for port := 1; port < 65536; port++ {
		previous := &tcpRanges[len(tcpRanges)-1]
		if !tcpBlocked[port] {
			if (previous.To + 1) == port {
				previous.To = port
			} else {
				tcpRanges = append(tcpRanges, portRange{From: port, To: port})
			}
		}
	}

	for _, masterGroup := range masterGroups {
		for _, nodeGroup := range nodeGroups {
			suffix := JoinSuffixes(nodeGroup, masterGroup)

			for _, r := range udpRanges {
				t := &awstasks.SecurityGroupRule{
					Name:          fi.PtrTo(fmt.Sprintf("node-to-master-udp-%d-%d%s", r.From, r.To, suffix)),
					Lifecycle:     b.Lifecycle,
					SecurityGroup: masterGroup.Task,
					SourceGroup:   nodeGroup.Task,
					FromPort:      fi.PtrTo(int64(r.From)),
					ToPort:        fi.PtrTo(int64(r.To)),
					Protocol:      fi.PtrTo("udp"),
				}
				AddDirectionalGroupRule(c, t)
			}
			for _, r := range tcpRanges {
				t := &awstasks.SecurityGroupRule{
					Name:          fi.PtrTo(fmt.Sprintf("node-to-master-tcp-%d-%d%s", r.From, r.To, suffix)),
					Lifecycle:     b.Lifecycle,
					SecurityGroup: masterGroup.Task,
					SourceGroup:   nodeGroup.Task,
					FromPort:      fi.PtrTo(int64(r.From)),
					ToPort:        fi.PtrTo(int64(r.To)),
					Protocol:      fi.PtrTo("tcp"),
				}
				AddDirectionalGroupRule(c, t)
			}
			for _, protocol := range protocols {
				awsName := strconv.Itoa(int(protocol))
				name := awsName
				switch protocol {
				case ProtocolIPIP:
					name = "ipip"
				default:
					klog.Warningf("unknown protocol %q - naming by number", awsName)
				}

				t := &awstasks.SecurityGroupRule{
					Name:          fi.PtrTo(fmt.Sprintf("node-to-master-protocol-%s%s", name, suffix)),
					Lifecycle:     b.Lifecycle,
					SecurityGroup: masterGroup.Task,
					SourceGroup:   nodeGroup.Task,
					Protocol:      fi.PtrTo(awsName),
				}
				AddDirectionalGroupRule(c, t)
			}
		}
	}

	// For AmazonVPC networking, pods running in Nodes could need to reach pods in master/s
	if b.Cluster.Spec.Networking.AmazonVPC != nil {
		// Nodes can talk to masters
		for _, src := range nodeGroups {
			for _, dest := range masterGroups {
				suffix := JoinSuffixes(src, dest)

				t := &awstasks.SecurityGroupRule{
					Name:          fi.PtrTo("all-nodes-to-master" + suffix),
					Lifecycle:     b.Lifecycle,
					SecurityGroup: dest.Task,
					SourceGroup:   src.Task,
				}
				AddDirectionalGroupRule(c, t)
			}
		}
	}
}

func (b *FirewallModelBuilder) buildMasterRules(c *fi.ModelBuilderContext, nodeGroups []SecurityGroupInfo) ([]SecurityGroupInfo, error) {
	masterGroups, err := b.GetSecurityGroups(kops.InstanceGroupRoleControlPlane)
	if err != nil {
		return nil, err
	}

	for _, group := range masterGroups {
		group.Task.Lifecycle = b.Lifecycle
		c.AddTask(group.Task)
	}

	for _, src := range masterGroups {
		// Allow full egress
		{
			t := &awstasks.SecurityGroupRule{
				Name:          fi.PtrTo("ipv4-master-egress" + src.Suffix),
				Lifecycle:     b.Lifecycle,
				SecurityGroup: src.Task,
				Egress:        fi.PtrTo(true),
				CIDR:          fi.PtrTo("0.0.0.0/0"),
			}
			AddDirectionalGroupRule(c, t)
		}
		{
			t := &awstasks.SecurityGroupRule{
				Name:          fi.PtrTo("ipv6-master-egress" + src.Suffix),
				Lifecycle:     b.Lifecycle,
				SecurityGroup: src.Task,
				Egress:        fi.PtrTo(true),
				IPv6CIDR:      fi.PtrTo("::/0"),
			}
			AddDirectionalGroupRule(c, t)
		}

		// Masters can talk to masters
		for _, dest := range masterGroups {
			suffix := JoinSuffixes(src, dest)

			t := &awstasks.SecurityGroupRule{
				Name:          fi.PtrTo("all-master-to-master" + suffix),
				Lifecycle:     b.Lifecycle,
				SecurityGroup: dest.Task,
				SourceGroup:   src.Task,
			}
			AddDirectionalGroupRule(c, t)
		}

		// Masters can talk to nodes
		for _, dest := range nodeGroups {
			suffix := JoinSuffixes(src, dest)

			t := &awstasks.SecurityGroupRule{
				Name:          fi.PtrTo("all-master-to-node" + suffix),
				Lifecycle:     b.Lifecycle,
				SecurityGroup: dest.Task,
				SourceGroup:   src.Task,
			}
			AddDirectionalGroupRule(c, t)
		}
	}

	return masterGroups, nil
}

type SecurityGroupInfo struct {
	Name   string
	Suffix string
	Task   *awstasks.SecurityGroup
}

func (b *AWSModelContext) GetSecurityGroups(role kops.InstanceGroupRole) ([]SecurityGroupInfo, error) {
	var baseGroup *awstasks.SecurityGroup
	if role == kops.InstanceGroupRoleControlPlane {
		name := b.SecurityGroupName(role)
		baseGroup = &awstasks.SecurityGroup{
			Name:        fi.PtrTo(name),
			VPC:         b.LinkToVPC(),
			Description: fi.PtrTo("Security group for masters"),
			RemoveExtraRules: []string{
				"port=22",   // SSH
				"port=443",  // k8s api
				"port=2380", // etcd main peer
				"port=2381", // etcd events peer
				"port=4001", // etcd main
				"port=4002", // etcd events
				"port=4789", // VXLAN
				"port=179",  // Calico
				"port=8443", // k8s api secondary listener

				// TODO: UDP vs TCP
				// TODO: Protocol 4 for calico
			},
		}
		baseGroup.Tags = b.CloudTags(name, false)
	} else if role == kops.InstanceGroupRoleNode {
		name := b.SecurityGroupName(role)
		baseGroup = &awstasks.SecurityGroup{
			Name:             fi.PtrTo(name),
			VPC:              b.LinkToVPC(),
			Description:      fi.PtrTo("Security group for nodes"),
			RemoveExtraRules: []string{"port=22"},
		}
		baseGroup.Tags = b.CloudTags(name, false)
	} else if role == kops.InstanceGroupRoleBastion {
		name := b.SecurityGroupName(role)
		baseGroup = &awstasks.SecurityGroup{
			Name:             fi.PtrTo(name),
			VPC:              b.LinkToVPC(),
			Description:      fi.PtrTo("Security group for bastion"),
			RemoveExtraRules: []string{"port=22"},
		}
		baseGroup.Tags = b.CloudTags(name, false)
	} else {
		return nil, fmt.Errorf("not a supported security group type")
	}

	var groups []SecurityGroupInfo

	done := make(map[string]bool)

	// Build groups that specify a SecurityGroupOverride
	allOverrides := true
	for _, ig := range b.InstanceGroups {
		if ig.Spec.Role != role {
			continue
		}

		if ig.Spec.SecurityGroupOverride == nil {
			allOverrides = false
			continue
		}

		name := fi.ValueOf(ig.Spec.SecurityGroupOverride)

		// De-duplicate security groups
		if done[name] {
			continue
		}
		done[name] = true

		sgName := fmt.Sprintf("%v-%v", fi.ValueOf(ig.Spec.SecurityGroupOverride), role)
		t := &awstasks.SecurityGroup{
			Name:        &sgName,
			ID:          ig.Spec.SecurityGroupOverride,
			VPC:         b.LinkToVPC(),
			Shared:      fi.PtrTo(true),
			Description: baseGroup.Description,
		}
		// Because the SecurityGroup is shared, we don't set RemoveExtraRules
		// This does mean we don't check them.  We might want to revisit this in future.

		suffix := "-" + name

		groups = append(groups, SecurityGroupInfo{
			Name:   name,
			Suffix: suffix,
			Task:   t,
		})
	}

	// Add the default SecurityGroup, if any InstanceGroups are using the default
	if !allOverrides {
		groups = append(groups, SecurityGroupInfo{
			Name: fi.ValueOf(baseGroup.Name),
			Task: baseGroup,
		})
	}

	return groups, nil
}

// JoinSuffixes constructs a suffix for traffic from the src to the dest group
// We have to avoid ambiguity in the case where one has a suffix and the other does not,
// where normally l.Suffix + r.Suffix would equal r.Suffix + l.Suffix
func JoinSuffixes(src SecurityGroupInfo, dest SecurityGroupInfo) string {
	if src.Suffix == "" && dest.Suffix == "" {
		return ""
	}

	s := src.Suffix
	if s == "" {
		s = "-default"
	}

	d := dest.Suffix
	if d == "" {
		d = "-default"
	}

	return s + d
}

func AddDirectionalGroupRule(c *fi.ModelBuilderContext, t *awstasks.SecurityGroupRule) {
	name := generateName(t)
	t.Name = fi.PtrTo(name)
	tags := make(map[string]string)
	for key, value := range t.SecurityGroup.Tags {
		tags[key] = value
	}
	tags["Name"] = *t.Name
	t.Tags = tags

	klog.V(8).Infof("Adding rule %v", name)
	c.AddTask(t)
}

func generateName(o *awstasks.SecurityGroupRule) string {
	var target, dst, src, direction, proto string
	if o.SourceGroup != nil {
		target = fi.ValueOf(o.SourceGroup.Name)
	} else if o.CIDR != nil {
		target = fi.ValueOf(o.CIDR)
	} else if o.IPv6CIDR != nil {
		target = fi.ValueOf(o.IPv6CIDR)
	} else {
		target = "0.0.0.0/0"
	}

	if o.Protocol == nil || fi.ValueOf(o.Protocol) == "" {
		proto = "all"
	} else {
		proto = fi.ValueOf(o.Protocol)
	}

	if o.Egress == nil || !fi.ValueOf(o.Egress) {
		direction = "ingress"
		src = target
		dst = fi.ValueOf(o.SecurityGroup.Name)
	} else {
		direction = "egress"
		dst = target
		src = fi.ValueOf(o.SecurityGroup.Name)
	}

	return fmt.Sprintf("from-%s-%s-%s-%dto%d-%s", src, direction,
		proto, fi.ValueOf(o.FromPort), fi.ValueOf(o.ToPort), dst)
}
