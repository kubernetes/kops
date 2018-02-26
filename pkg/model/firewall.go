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
	if err := b.buildNodeRules(c); err != nil {
		return err
	}
	if err := b.buildMasterRules(c); err != nil {
		return err
	}
	return nil
}

func (b *FirewallModelBuilder) buildNodeRules(c *fi.ModelBuilderContext) error {
	{
		t := &awstasks.SecurityGroup{
			Name:             s(b.SecurityGroupName(kops.InstanceGroupRoleNode)),
			Lifecycle:        b.Lifecycle,
			VPC:              b.LinkToVPC(),
			Description:      s("Security group for nodes"),
			RemoveExtraRules: []string{"port=22"},
		}
		c.AddTask(t)
	}

	// Allow full egress
	{
		t := &awstasks.SecurityGroupRule{
			Name:          s("node-egress"),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			Egress:        fi.Bool(true),
			CIDR:          s("0.0.0.0/0"),
		}
		c.AddTask(t)
	}

	// Nodes can talk to nodes
	{
		t := &awstasks.SecurityGroupRule{
			Name:          s("all-node-to-node"),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
		}
		c.AddTask(t)
	}

	// Pods running in Nodes could need to reach pods in master/s
	if b.Cluster.Spec.Networking != nil && b.Cluster.Spec.Networking.AmazonVPC != nil {
		// Nodes can talk to masters
		{
			t := &awstasks.SecurityGroupRule{
				Name:          s("all-nodes-to-master"),
				Lifecycle:     b.Lifecycle,
				SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
				SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			}
			c.AddTask(t)
		}
	}

	// We _should_ block per port... but:
	// * It causes e2e tests to break
	// * Users expect to be able to reach pods
	// * If users are running an overlay, we punch a hole in it anyway
	//b.applyNodeToMasterAllowSpecificPorts(c)
	b.applyNodeToMasterBlockSpecificPorts(c)
	// re-using provided security group for the node

	return nil
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

		if b.Cluster.Spec.Networking.Kuberouter != nil {
			protocols = append(protocols, ProtocolIPIP)
		}
	}

	for _, udpPort := range udpPorts {
		c.AddTask(&awstasks.SecurityGroupRule{
			Name:          s(fmt.Sprintf("node-to-master-udp-%d", udpPort)),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
			SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			FromPort:      i64(udpPort),
			ToPort:        i64(udpPort),
			Protocol:      s("udp"),
		})
	}
	for _, tcpPort := range tcpPorts {
		c.AddTask(&awstasks.SecurityGroupRule{
			Name:          s(fmt.Sprintf("node-to-master-tcp-%d", tcpPort)),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
			SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			FromPort:      i64(tcpPort),
			ToPort:        i64(tcpPort),
			Protocol:      s("tcp"),
		})
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

		c.AddTask(&awstasks.SecurityGroupRule{
			Name:          s("node-to-master-protocol-" + name),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
			SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			Protocol:      s(awsName),
		})
	}
}

func (b *FirewallModelBuilder) applyNodeToMasterBlockSpecificPorts(c *fi.ModelBuilderContext) {
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

	if b.Cluster.Spec.Networking.Kuberouter != nil {
		protocols = append(protocols, ProtocolIPIP)
	}

	for _, r := range udpRanges {
		c.AddTask(&awstasks.SecurityGroupRule{
			Name:          s(fmt.Sprintf("node-to-master-udp-%d-%d", r.From, r.To)),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
			SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			FromPort:      i64(int64(r.From)),
			ToPort:        i64(int64(r.To)),
			Protocol:      s("udp"),
		})
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

	for _, r := range tcpRanges {
		c.AddTask(&awstasks.SecurityGroupRule{
			Name:          s(fmt.Sprintf("node-to-master-tcp-%d-%d", r.From, r.To)),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
			SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			FromPort:      i64(int64(r.From)),
			ToPort:        i64(int64(r.To)),
			Protocol:      s("tcp"),
		})
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

		c.AddTask(&awstasks.SecurityGroupRule{
			Name:          s("node-to-master-protocol-" + name),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
			SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			Protocol:      s(awsName),
		})
	}
}

func (b *FirewallModelBuilder) buildMasterRules(c *fi.ModelBuilderContext) error {
	{
		t := &awstasks.SecurityGroup{
			Name:        s(b.SecurityGroupName(kops.InstanceGroupRoleMaster)),
			Lifecycle:   b.Lifecycle,
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
		c.AddTask(t)
	}

	// Allow full egress
	{
		t := &awstasks.SecurityGroupRule{
			Name:          s("master-egress"),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
			Egress:        fi.Bool(true),
			CIDR:          s("0.0.0.0/0"),
		}
		c.AddTask(t)
	}

	// Masters can talk to masters
	{
		t := &awstasks.SecurityGroupRule{
			Name:          s("all-master-to-master"),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
			SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
		}
		c.AddTask(t)
	}

	// Masters can talk to nodes
	{
		t := &awstasks.SecurityGroupRule{
			Name:          s("all-master-to-node"),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
		}
		c.AddTask(t)
	}

	return nil
}
