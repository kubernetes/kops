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

package model

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"

	"k8s.io/klog/v2"
)

type Protocol int

const (
	ProtocolIPIP Protocol = 4
)

// FirewallModelBuilder configures firewall network objects
type FirewallModelBuilder struct {
	*KopsModelContext
	Cloud     awsup.AWSCloud
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &FirewallModelBuilder{}

func (b *FirewallModelBuilder) Build(c *fi.ModelBuilderContext) error {
	tasks, err := b.getExistingRulesFromCloud()
	if err != nil {
		return err
	}

	finalRules := make(map[string]*awstasks.SecurityGroupRule)
	for _, task := range tasks {
		klog.V(4).Infof("found rule %q", fi.StringValue(task.Name))
		finalRules[fi.StringValue(task.Name)] = task
	}

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

	for _, task := range b.SecurityGroupRules {
		finalRules[fi.StringValue(task.Name)] = task
	}

	for _, rule := range finalRules {
		c.AddTask(rule)
	}

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
				Name:          s("node-egress" + src.Suffix),
				Lifecycle:     b.Lifecycle,
				SecurityGroup: src.Task,
				Egress:        fi.Bool(true),
				CIDR:          s("0.0.0.0/0"),
			}
			b.AddDirectionalGroupRule(c, t)
		}

		// Nodes can talk to nodes
		for _, dest := range nodeGroups {
			suffix := JoinSuffixes(src, dest)

			t := &awstasks.SecurityGroupRule{
				Name:          s("all-node-to-node" + suffix),
				Lifecycle:     b.Lifecycle,
				SecurityGroup: dest.Task,
				SourceGroup:   src.Task,
			}
			b.AddDirectionalGroupRule(c, t)
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

	if b.Cluster.Spec.Networking.Kuberouter != nil {
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
					Name:          s(fmt.Sprintf("node-to-master-udp-%d-%d%s", r.From, r.To, suffix)),
					Lifecycle:     b.Lifecycle,
					SecurityGroup: masterGroup.Task,
					SourceGroup:   nodeGroup.Task,
					FromPort:      i64(int64(r.From)),
					ToPort:        i64(int64(r.To)),
					Protocol:      s("udp"),
				}
				b.AddDirectionalGroupRule(c, t)
			}
			for _, r := range tcpRanges {
				t := &awstasks.SecurityGroupRule{
					Name:          s(fmt.Sprintf("node-to-master-tcp-%d-%d%s", r.From, r.To, suffix)),
					Lifecycle:     b.Lifecycle,
					SecurityGroup: masterGroup.Task,
					SourceGroup:   nodeGroup.Task,
					FromPort:      i64(int64(r.From)),
					ToPort:        i64(int64(r.To)),
					Protocol:      s("tcp"),
				}
				b.AddDirectionalGroupRule(c, t)
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
					Name:          s(fmt.Sprintf("node-to-master-protocol-%s%s", name, suffix)),
					Lifecycle:     b.Lifecycle,
					SecurityGroup: masterGroup.Task,
					SourceGroup:   nodeGroup.Task,
					Protocol:      s(awsName),
				}
				b.AddDirectionalGroupRule(c, t)
			}
		}
	}

	// For AmazonVPC networking, pods running in Nodes could need to reach pods in master/s
	if b.Cluster.Spec.Networking != nil && b.Cluster.Spec.Networking.AmazonVPC != nil {
		// Nodes can talk to masters
		for _, src := range nodeGroups {
			for _, dest := range masterGroups {
				suffix := JoinSuffixes(src, dest)

				t := &awstasks.SecurityGroupRule{
					Name:          s("all-nodes-to-master" + suffix),
					Lifecycle:     b.Lifecycle,
					SecurityGroup: dest.Task,
					SourceGroup:   src.Task,
				}
				b.AddDirectionalGroupRule(c, t)
			}
		}
	}

}

func (b *FirewallModelBuilder) buildMasterRules(c *fi.ModelBuilderContext, nodeGroups []SecurityGroupInfo) ([]SecurityGroupInfo, error) {
	masterGroups, err := b.GetSecurityGroups(kops.InstanceGroupRoleMaster)
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
				Name:          s("master-egress" + src.Suffix),
				Lifecycle:     b.Lifecycle,
				SecurityGroup: src.Task,
				Egress:        fi.Bool(true),
				CIDR:          s("0.0.0.0/0"),
			}
			b.AddDirectionalGroupRule(c, t)
		}

		// Masters can talk to masters
		for _, dest := range masterGroups {
			suffix := JoinSuffixes(src, dest)

			t := &awstasks.SecurityGroupRule{
				Name:          s("all-master-to-master" + suffix),
				Lifecycle:     b.Lifecycle,
				SecurityGroup: dest.Task,
				SourceGroup:   src.Task,
			}
			b.AddDirectionalGroupRule(c, t)
		}

		// Masters can talk to nodes
		for _, dest := range nodeGroups {
			suffix := JoinSuffixes(src, dest)

			t := &awstasks.SecurityGroupRule{
				Name:          s("all-master-to-node" + suffix),
				Lifecycle:     b.Lifecycle,
				SecurityGroup: dest.Task,
				SourceGroup:   src.Task,
			}
			b.AddDirectionalGroupRule(c, t)
		}
	}

	return masterGroups, nil
}

type SecurityGroupInfo struct {
	Name   string
	Suffix string
	Task   *awstasks.SecurityGroup
}

func (b *KopsModelContext) GetSecurityGroups(role kops.InstanceGroupRole) ([]SecurityGroupInfo, error) {
	var baseGroup *awstasks.SecurityGroup
	if role == kops.InstanceGroupRoleMaster {
		name := b.SecurityGroupName(role)
		baseGroup = &awstasks.SecurityGroup{
			Name:        s(name),
			VPC:         b.LinkToVPC(),
			Description: s("Security group for masters"),
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
			Name:             s(name),
			VPC:              b.LinkToVPC(),
			Description:      s("Security group for nodes"),
			RemoveExtraRules: []string{"port=22"},
		}
		baseGroup.Tags = b.CloudTags(name, false)
	} else if role == kops.InstanceGroupRoleBastion {
		name := b.SecurityGroupName(role)
		baseGroup = &awstasks.SecurityGroup{
			Name:             s(name),
			VPC:              b.LinkToVPC(),
			Description:      s("Security group for bastion"),
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

		name := fi.StringValue(ig.Spec.SecurityGroupOverride)

		// De-duplicate security groups
		if done[name] {
			continue
		}
		done[name] = true

		sgName := fmt.Sprintf("%v-%v", fi.StringValue(ig.Spec.SecurityGroupOverride), role)
		t := &awstasks.SecurityGroup{
			Name:        &sgName,
			ID:          ig.Spec.SecurityGroupOverride,
			VPC:         b.LinkToVPC(),
			Shared:      fi.Bool(true),
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
			Name: fi.StringValue(baseGroup.Name),
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

func (b *KopsModelContext) AddDirectionalGroupRule(c *fi.ModelBuilderContext, t *awstasks.SecurityGroupRule) {

	name := generateName(t)
	b.SecurityGroupRules[name] = t
	t.Name = fi.String(name)
	t.Delete = fi.Bool(false)

	klog.V(4).Infof("Adding rule %q", name)

}

func generateName(o *awstasks.SecurityGroupRule) string {

	var target, dst, src, direction, proto string
	if o.SourceGroup != nil {
		target = fi.StringValue(o.SourceGroup.Name)
	} else if o.CIDR != nil && fi.StringValue(o.CIDR) != "" {
		target = fi.StringValue(o.CIDR)
	} else {
		target = "0.0.0.0/0"
	}

	if o.Protocol == nil || fi.StringValue(o.Protocol) == "" {
		proto = "all"
	} else {
		proto = fi.StringValue(o.Protocol)
	}

	if o.Egress == nil || !fi.BoolValue(o.Egress) {
		direction = "ingress"
		src = target
		dst = fi.StringValue(o.SecurityGroup.Name)
	} else {
		direction = "egress"
		dst = target
		src = fi.StringValue(o.SecurityGroup.Name)
	}

	return fmt.Sprintf("from-%s-%s-%s-%dto%d-%s", src, direction,
		proto, fi.Int64Value(o.FromPort), fi.Int64Value(o.ToPort), dst)
}

func (b *FirewallModelBuilder) getExistingRulesFromCloud() ([]*awstasks.SecurityGroupRule, error) {
	tasks := []*awstasks.SecurityGroupRule{}
	sgs, err := b.getClusterSecurityGroups()
	if err != nil {
		return nil, err
	}
	for _, sg := range sgs {
		name := fi.StringValue(sg.GroupName)

		klog.V(4).Infof("found group %q with id %s", name, fi.StringValue(sg.GroupId))
		// We assume that if the security group name ends with the cluster name it is owned by kOps.
		// We must ignore security groups about which we don't know anything, as we cannot make tasks of them.
		if !strings.HasSuffix(name, "."+b.ClusterName()) {
			klog.V(4).Infof("skipping EC2 security group %q", name)
			continue
		}
		for _, rule := range sg.IpPermissions {
			ts := b.createRulesFromPermissions(sg, rule, false, sgs)
			tasks = append(tasks, ts...)
		}
		for _, rule := range sg.IpPermissionsEgress {
			ts := b.createRulesFromPermissions(sg, rule, true, sgs)
			tasks = append(tasks, ts...)
		}
	}
	return tasks, nil
}

func (b *FirewallModelBuilder) getClusterSecurityGroups() (map[string]*ec2.SecurityGroup, error) {
	sgs := make(map[string]*ec2.SecurityGroup)

	request := &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			awsup.NewEC2Filter("tag:KubernetesCluster", b.ClusterName()),
		},
	}

	response, err := b.Cloud.EC2().DescribeSecurityGroups(request)
	if err != nil {
		return nil, fmt.Errorf("error listing EC2 security groups: %w", err)
	}

	for _, sg := range response.SecurityGroups {
		sgs[fi.StringValue(sg.GroupId)] = sg
	}
	return sgs, nil

}

func (b *FirewallModelBuilder) createRulesFromPermissions(sg *ec2.SecurityGroup, rule *ec2.IpPermission, egress bool, sgs map[string]*ec2.SecurityGroup) (tasks []*awstasks.SecurityGroupRule) {
	for _, cidr := range rule.IpRanges {
		// CIDRs with description starting with kubernetes.io are owned by CCM
		if strings.HasPrefix(fi.StringValue(cidr.Description), "kubernetes.io") {
			continue
		}
		rule := b.createBaseRule(sg, rule, egress)
		rule.Egress = fi.Bool(egress)
		rule.CIDR = cidr.CidrIp

		rule.Name = fi.String(generateName(rule))
		tasks = append(tasks, rule)
	}
	for _, p := range rule.UserIdGroupPairs {
		rule := b.createBaseRule(sg, rule, egress)
		rule.Egress = fi.Bool(egress)
		pGroup := sgs[fi.StringValue(p.GroupId)]
		groupName := fi.StringValue(pGroup.GroupName)
		// We assume that if the security group name ends with the cluster name it is owned by kOps.
		// We must ignore security groups about which we don't know anything, as we cannot make tasks of them.
		if !strings.HasSuffix(groupName, b.ClusterName()) {
			klog.V(4).Infof("Skipping rule; target EC2 security group %q not owned by kOps", groupName)
			continue
		}
		source := &awstasks.SecurityGroup{Name: pGroup.GroupName}

		rule.SourceGroup = source

		rule.Name = fi.String(generateName(rule))
		tasks = append(tasks, rule)
	}
	return tasks
}

func (b *FirewallModelBuilder) createBaseRule(sg *ec2.SecurityGroup, rule *ec2.IpPermission, egress bool) *awstasks.SecurityGroupRule {

	task := &awstasks.SecurityGroupRule{
		SecurityGroup: &awstasks.SecurityGroup{Name: sg.GroupName},
		FromPort:      rule.FromPort,
		ToPort:        rule.ToPort,
		Protocol:      rule.IpProtocol,
		Egress:        fi.Bool(egress),
		Delete:        fi.Bool(true),
	}
	if fi.StringValue(task.Protocol) == "-1" {
		task.Protocol = nil
	}
	return task
}
