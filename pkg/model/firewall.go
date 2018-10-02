/*
Copyright 2016 The Kubernetes Authors.

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

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"

	"github.com/golang/glog"
)

type Protocol int

const (
	ProtocolIPIP Protocol = 4
)

// FirewallModelBuilder configures firewall network objects
type FirewallModelBuilder struct {
	*KopsModelContext
	Lifecycle *fi.Lifecycle
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

func (b *FirewallModelBuilder) buildNodeRules(c *fi.ModelBuilderContext) (map[string]*awstasks.SecurityGroup, error) {

	nodeGroups, err := b.createSecurityGroups(kops.InstanceGroupRoleNode, b.Lifecycle, c)
	if err != nil {
		return nil, err
	}

	for nodeGroupName, secGroup := range nodeGroups {
		suffix := GetGroupSuffix(nodeGroupName, nodeGroups)
		// Allow full egress
		{
			t := &awstasks.SecurityGroupRule{
				Name:          s(fmt.Sprintf("node-egress%s", suffix)),
				Lifecycle:     b.Lifecycle,
				SecurityGroup: secGroup,
				Egress:        fi.Bool(true),
				CIDR:          s("0.0.0.0/0"),
			}
			c.AddTask(t)
		}

		// Nodes can talk to nodes
		{
			t := &awstasks.SecurityGroupRule{
				Name:          s(fmt.Sprintf("all-node-to-node%s", suffix)),
				Lifecycle:     b.Lifecycle,
				SecurityGroup: secGroup,
				SourceGroup:   secGroup,
			}
			c.AddTask(t)
		}

		// Pods running in Nodes could need to reach pods in master/s
		if b.Cluster.Spec.Networking != nil && b.Cluster.Spec.Networking.AmazonVPC != nil {
			// Nodes can talk to masters
			{
				t := &awstasks.SecurityGroupRule{
					Name:          s(fmt.Sprintf("all-nodes-to-master%s", suffix)),
					Lifecycle:     b.Lifecycle,
					SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
					SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
				}
				c.AddTask(t)
			}
		}
	}

	return nodeGroups, nil
}

func (b *FirewallModelBuilder) applyNodeToMasterAllowSpecificPorts(c *fi.ModelBuilderContext) {
	// TODO: We need to remove the ALL rule
	//W1229 12:32:22.300132    9003 executor.go:109] error running task "SecurityGroupRule/node-to-master-443" (9m58s remaining to succeed): error creating SecurityGroupIngress: InvalidPermission.Duplicate: the specified rule "peer: sg-f6b1a68b, ALL, ALLOW" already exists
	//status code: 400, request id: 6a69627f-9a26-4bd0-b294-a9a96f89bc46

	udpPorts := []int64{}
	tcpPorts := []int64{}
	protocols := []Protocol{}

	// allow access to API
	tcpPorts = append(tcpPorts, 443)

	// allow cadvisor
	tcpPorts = append(tcpPorts, 4194)

	// kubelet read-only used by heapster
	tcpPorts = append(tcpPorts, 10255)

	if b.Cluster.Spec.Networking != nil {
		if b.Cluster.Spec.Networking.Kopeio != nil {
			// VXLAN over UDP
			udpPorts = append(udpPorts, 4789)
		}

		if b.Cluster.Spec.Networking.Weave != nil {
			udpPorts = append(udpPorts, 6783)
			tcpPorts = append(tcpPorts, 6783)
			udpPorts = append(udpPorts, 6784)
		}

		if b.Cluster.Spec.Networking.Flannel != nil {
			switch b.Cluster.Spec.Networking.Flannel.Backend {
			case "", "udp":
				udpPorts = append(udpPorts, 8285)
			case "vxlan":
				udpPorts = append(udpPorts, 8472)
			default:
				glog.Warningf("unknown flannel networking backend %q", b.Cluster.Spec.Networking.Flannel.Backend)
			}
		}

		if b.Cluster.Spec.Networking.Calico != nil {
			// Calico needs to access etcd
			// TODO: Remove, replace with etcd in calico manifest
			// https://coreos.com/etcd/docs/latest/v2/configuration.html
			glog.Warningf("Opening etcd port on masters for access from the nodes, for calico.  This is unsafe in untrusted environments.")
			tcpPorts = append(tcpPorts, 4001)
			tcpPorts = append(tcpPorts, 179)
			protocols = append(protocols, ProtocolIPIP)
		}

		if b.Cluster.Spec.Networking.Romana != nil {
			// Romana needs to access etcd
			glog.Warningf("Opening etcd port on masters for access from the nodes, for romana.  This is unsafe in untrusted environments.")
			tcpPorts = append(tcpPorts, 4001)
			tcpPorts = append(tcpPorts, 9600)
		}

		if b.Cluster.Spec.Networking.Cilium != nil {
			// Cilium needs to access etcd
			glog.Warningf("Opening etcd port on masters for access from the nodes, for Cilium.  This is unsafe in untrusted environments.")
			tcpPorts = append(tcpPorts, 4001)
		}

		if b.Cluster.Spec.Networking.Kuberouter != nil {
			protocols = append(protocols, ProtocolIPIP)
		}
	}

	for _, udpPort := range udpPorts {
		t := &awstasks.SecurityGroupRule{
			Name:          s(fmt.Sprintf("node-to-master-udp-%d", udpPort)),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
			SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			FromPort:      i64(udpPort),
			ToPort:        i64(udpPort),
			Protocol:      s("udp"),
		}
		c.AddTask(t)
	}
	for _, tcpPort := range tcpPorts {
		t := &awstasks.SecurityGroupRule{
			Name:          s(fmt.Sprintf("node-to-master-tcp-%d", tcpPort)),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
			SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			FromPort:      i64(tcpPort),
			ToPort:        i64(tcpPort),
			Protocol:      s("tcp"),
		}
		c.AddTask(t)
	}
	for _, protocol := range protocols {
		awsName := strconv.Itoa(int(protocol))
		name := awsName
		switch protocol {
		case ProtocolIPIP:
			name = "ipip"
		default:
			glog.Warningf("unknown protocol %q - naming by number", awsName)
		}

		t := &awstasks.SecurityGroupRule{
			Name:          s("node-to-master-protocol-" + name),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
			SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			Protocol:      s(awsName),
		}
		c.AddTask(t)
	}
}

func (b *FirewallModelBuilder) applyNodeToMasterBlockSpecificPorts(c *fi.ModelBuilderContext, nodeGroups map[string]*awstasks.SecurityGroup, masterGroups map[string]*awstasks.SecurityGroup) {
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

	if b.Cluster.Spec.Networking.Calico != nil {
		// Calico needs to access etcd
		// TODO: Remove, replace with etcd in calico manifest
		glog.Warningf("Opening etcd port on masters for access from the nodes, for calico.  This is unsafe in untrusted environments.")
		tcpBlocked[4001] = false
		protocols = append(protocols, ProtocolIPIP)
	}

	if b.Cluster.Spec.Networking.Romana != nil {
		// Romana needs to access etcd
		glog.Warningf("Opening etcd port on masters for access from the nodes, for romana.  This is unsafe in untrusted environments.")
		tcpBlocked[4001] = false
		protocols = append(protocols, ProtocolIPIP)
	}

	if b.Cluster.Spec.Networking.Cilium != nil {
		// Cilium needs to access etcd
		glog.Warningf("Opening etcd port on masters for access from the nodes, for Cilium.  This is unsafe in untrusted environments.")
		tcpBlocked[4001] = false
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

	for masterSecGroupName, masterGroup := range masterGroups {
		var masterSuffix string
		var nodeSuffix string

		if len(masterGroups) != 1 {
			masterSuffix = "-" + masterSecGroupName
		} else {
			masterSuffix = ""
		}

		for nodeSecGroupName, nodeGroup := range nodeGroups {

			if len(masterGroups) == 1 && len(nodeGroups) == 1 {
				nodeSuffix = ""
			} else {
				nodeSuffix = fmt.Sprintf("%s-%s", masterSuffix, nodeSecGroupName)
			}

			for _, r := range udpRanges {
				t := &awstasks.SecurityGroupRule{
					Name:          s(fmt.Sprintf("node-to-master-udp-%d-%d%s", r.From, r.To, nodeSuffix)),
					Lifecycle:     b.Lifecycle,
					SecurityGroup: masterGroup,
					SourceGroup:   nodeGroup,
					FromPort:      i64(int64(r.From)),
					ToPort:        i64(int64(r.To)),
					Protocol:      s("udp"),
				}
				c.AddTask(t)
			}
			for _, r := range tcpRanges {
				t := &awstasks.SecurityGroupRule{
					Name:          s(fmt.Sprintf("node-to-master-tcp-%d-%d%s", r.From, r.To, nodeSuffix)),
					Lifecycle:     b.Lifecycle,
					SecurityGroup: masterGroup,
					SourceGroup:   nodeGroup,
					FromPort:      i64(int64(r.From)),
					ToPort:        i64(int64(r.To)),
					Protocol:      s("tcp"),
				}
				c.AddTask(t)
			}
			for _, protocol := range protocols {
				awsName := strconv.Itoa(int(protocol))
				name := awsName
				switch protocol {
				case ProtocolIPIP:
					name = "ipip"
				default:
					glog.Warningf("unknown protocol %q - naming by number", awsName)
				}

				t := &awstasks.SecurityGroupRule{
					Name:          s(fmt.Sprintf("node-to-master-protocol-%s%s", name, nodeSuffix)),
					Lifecycle:     b.Lifecycle,
					SecurityGroup: masterGroup,
					SourceGroup:   nodeGroup,
					Protocol:      s(awsName),
				}
				c.AddTask(t)
			}
		}
	}
}

func (b *FirewallModelBuilder) buildMasterRules(c *fi.ModelBuilderContext, nodeGroups map[string]*awstasks.SecurityGroup) (map[string]*awstasks.SecurityGroup, error) {
	masterGroups, err := b.createSecurityGroups(kops.InstanceGroupRoleMaster, b.Lifecycle, c)
	if err != nil {
		return nil, err
	}

	for masterSecGroupName, masterGroup := range masterGroups {
		suffix := GetGroupSuffix(masterSecGroupName, masterGroups)
		// Allow full egress
		{
			t := &awstasks.SecurityGroupRule{
				Name:          s(fmt.Sprintf("master-egress%s", suffix)),
				Lifecycle:     b.Lifecycle,
				SecurityGroup: masterGroup,
				Egress:        fi.Bool(true),
				CIDR:          s("0.0.0.0/0"),
			}
			c.AddTask(t)
		}

		// Masters can talk to masters
		{
			t := &awstasks.SecurityGroupRule{
				Name:          s(fmt.Sprintf("all-master-to-master%s", suffix)),
				Lifecycle:     b.Lifecycle,
				SecurityGroup: masterGroup,
				SourceGroup:   masterGroup,
			}
			c.AddTask(t)
		}
		for nodeSecGroupName, nodeGroup := range nodeGroups {

			if len(masterGroups) == 1 && len(nodeGroups) == 1 {
				nodeSecGroupName = ""
			} else {
				nodeSecGroupName = fmt.Sprintf("%s-%s", masterSecGroupName, nodeSecGroupName)
			}

			// Masters can talk to nodes
			{
				t := &awstasks.SecurityGroupRule{
					Name:          s(fmt.Sprintf("all-master-to-node%s", nodeSecGroupName)),
					Lifecycle:     b.Lifecycle,
					SecurityGroup: nodeGroup,
					SourceGroup:   masterGroup,
				}
				c.AddTask(t)
			}
		}
	}

	return masterGroups, nil
}

func (b *KopsModelContext) GetSecurityGroups(role kops.InstanceGroupRole) (map[string]*awstasks.SecurityGroup, error) {
	return b.createSecurityGroups(role, nil, nil)
}

func (b *KopsModelContext) createSecurityGroups(role kops.InstanceGroupRole, lifecycle *fi.Lifecycle, c *fi.ModelBuilderContext) (map[string]*awstasks.SecurityGroup, error) {

	var baseGroup *awstasks.SecurityGroup
	if role == kops.InstanceGroupRoleMaster {
		name := "masters." + b.ClusterName()
		baseGroup = &awstasks.SecurityGroup{
			Name:        s(name),
			Lifecycle:   lifecycle,
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

				// TODO: UDP vs TCP
				// TODO: Protocol 4 for calico
			},
		}
		baseGroup.Tags = b.CloudTags(*baseGroup.Name, false)
	} else if role == kops.InstanceGroupRoleNode {
		name := "nodes." + b.ClusterName()
		baseGroup = &awstasks.SecurityGroup{
			Name:             s(name),
			Lifecycle:        lifecycle,
			VPC:              b.LinkToVPC(),
			Description:      s("Security group for nodes"),
			RemoveExtraRules: []string{"port=22"},
		}
		baseGroup.Tags = b.CloudTags(*baseGroup.Name, false)
	} else if role == kops.InstanceGroupRoleBastion {
		return nil, fmt.Errorf("bastion are not supported yet with instancegroup securitygroup")
		/*
			// TODO use this instead of the hardcoded names??
			// b.SecurityGroupName(kops.InstanceGroupRoleBastion))
			// TODO implement
			name := "bastion." + b.ClusterName()
			baseGroup = &awstasks.SecurityGroup{
				Name:             s(name),
				Lifecycle:        lifecycle,
				VPC:              b.LinkToVPC(),
				Description:      s("Security group for bastion"),
				RemoveExtraRules: []string{"port=22"},
			}
			baseGroup.Tags = b.CloudTags(*baseGroup.Name, false)
		*/
	} else {
		return nil, fmt.Errorf("not a supported security group type")
	}

	groups := make(map[string]*awstasks.SecurityGroup)
	for _, ig := range b.InstanceGroups {
		if ig.Spec.SecurityGroupOverride != nil && ig.Spec.Role == role {
			name := fi.StringValue(ig.Spec.SecurityGroupOverride)
			t := &awstasks.SecurityGroup{
				Name:        ig.Spec.SecurityGroupOverride,
				ID:          ig.Spec.SecurityGroupOverride,
				Lifecycle:   lifecycle,
				VPC:         b.LinkToVPC(),
				Shared:      fi.Bool(true),
				Description: baseGroup.Description,
			}
			groups[name] = t
		}
	}

	for name, t := range groups {
		if c != nil {
			glog.V(8).Infof("adding security group: %q", name)
			c.AddTask(t)
		}
	}

	if len(groups) == 0 {
		groups[fi.StringValue(baseGroup.Name)] = baseGroup
		if c != nil {
			glog.V(8).Infof("adding security group: %q", fi.StringValue(baseGroup.Name))
			c.AddTask(baseGroup)
		}
	}

	return groups, nil
}

// GetGroupSuffix returns the name of the security groups suffix.
func GetGroupSuffix(name string, groups map[string]*awstasks.SecurityGroup) string {
	if len(groups) != 1 {
		glog.V(8).Infof("adding group suffix: %q", name)
		return "-" + name
	}

	return ""
}
