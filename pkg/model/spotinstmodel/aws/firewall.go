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

package aws

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	awstasks "k8s.io/kops/upup/pkg/fi/cloudup/spotinsttasks/aws"
	"strconv"
)

// FirewallModelBuilder configures firewall network objects
type FirewallModelBuilder struct {
	*ModelContext
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
	name := "nodes." + b.ClusterName()

	{
		t := &awstasks.SecurityGroup{
			Name:             fi.String(name),
			Lifecycle:        b.Lifecycle,
			VPC:              b.LinkToVPC(),
			Description:      fi.String("Security group for nodes"),
			RemoveExtraRules: []string{"port=22"},
		}
		c.AddTask(t)
	}

	// Allow full egress
	{
		t := &awstasks.SecurityGroupRule{
			Name:          fi.String("node-egress"),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			Egress:        fi.Bool(true),
			CIDR:          fi.String("0.0.0.0/0"),
		}
		c.AddTask(t)
	}

	// Nodes can talk to nodes
	{
		t := &awstasks.SecurityGroupRule{
			Name:          fi.String("all-node-to-node"),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
		}
		c.AddTask(t)
	}

	// We _should_ block per port... but:
	// * It causes e2e tests to break
	// * Users expect to be able to reach pods
	// * If users are running an overlay, we punch a hole in it anyway
	//b.applyNodeToMasterAllowSpecificPorts(c)
	b.applyNodeToMasterBlockSpecificPorts(c)

	return nil
}

func (b *FirewallModelBuilder) applyNodeToMasterAllowSpecificPorts(c *fi.ModelBuilderContext) {
	// TODO: We need to remove the ALL rule
	//W1229 12:32:22.300132    9003 executor.go:109] error running task "SecurityGroupRule/node-to-master-443" (9m58s remaining to succeed): error creating SecurityGroupIngress: InvalidPermission.Duplicate: the specified rule "peer: sg-f6b1a68b, ALL, ALLOW" already exists
	//status code: 400, request id: 6a69627f-9a26-4bd0-b294-a9a96f89bc46

	udpPorts := []int64{}
	tcpPorts := []int64{}
	protocols := []model.Protocol{}

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
			udpPorts = append(udpPorts, 8285)
		}

		if b.Cluster.Spec.Networking.Calico != nil {
			// Calico needs to access etcd
			// TODO: Remove, replace with etcd in calico manifest
			glog.Warningf("Opening etcd port on masters for access from the nodes, for calico.  This is unsafe in untrusted environments.")
			tcpPorts = append(tcpPorts, 4001)

			tcpPorts = append(tcpPorts, 179)
			protocols = append(protocols, model.ProtocolIPIP)
		}
	}

	for _, udpPort := range udpPorts {
		c.AddTask(&awstasks.SecurityGroupRule{
			Name:          fi.String(fmt.Sprintf("node-to-master-udp-%d", udpPort)),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
			SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			FromPort:      fi.Int64(udpPort),
			ToPort:        fi.Int64(udpPort),
			Protocol:      fi.String("udp"),
		})
	}
	for _, tcpPort := range tcpPorts {
		c.AddTask(&awstasks.SecurityGroupRule{
			Name:          fi.String(fmt.Sprintf("node-to-master-tcp-%d", tcpPort)),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
			SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			FromPort:      fi.Int64(tcpPort),
			ToPort:        fi.Int64(tcpPort),
			Protocol:      fi.String("tcp"),
		})
	}
	for _, protocol := range protocols {
		awsName := strconv.Itoa(int(protocol))
		name := awsName
		switch protocol {
		case model.ProtocolIPIP:
			name = "ipip"
		default:
			glog.Warningf("unknown protocol %q - naming by number", awsName)
		}

		c.AddTask(&awstasks.SecurityGroupRule{
			Name:          fi.String("node-to-master-protocol-" + name),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
			SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			Protocol:      fi.String(awsName),
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
	tcpRanges := []portRange{{From: 1, To: 4000}, {From: 4003, To: 65535}}
	udpRanges := []portRange{{From: 1, To: 65535}}
	protocols := []model.Protocol{}

	if b.Cluster.Spec.Networking.Calico != nil {
		// Calico needs to access etcd
		// TODO: Remove, replace with etcd in calico manifest
		glog.Warningf("Opening etcd port on masters for access from the nodes, for calico.  This is unsafe in untrusted environments.")
		tcpRanges = []portRange{{From: 1, To: 4001}, {From: 4003, To: 65535}}
		protocols = append(protocols, model.ProtocolIPIP)
	}

	for _, r := range udpRanges {
		c.AddTask(&awstasks.SecurityGroupRule{
			Name:          fi.String(fmt.Sprintf("node-to-master-udp-%d-%d", r.From, r.To)),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
			SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			FromPort:      fi.Int64(int64(r.From)),
			ToPort:        fi.Int64(int64(r.To)),
			Protocol:      fi.String("udp"),
		})
	}
	for _, r := range tcpRanges {
		c.AddTask(&awstasks.SecurityGroupRule{
			Name:          fi.String(fmt.Sprintf("node-to-master-tcp-%d-%d", r.From, r.To)),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
			SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			FromPort:      fi.Int64(int64(r.From)),
			ToPort:        fi.Int64(int64(r.To)),
			Protocol:      fi.String("tcp"),
		})
	}
	for _, protocol := range protocols {
		awsName := strconv.Itoa(int(protocol))
		name := awsName
		switch protocol {
		case model.ProtocolIPIP:
			name = "ipip"
		default:
			glog.Warningf("unknown protocol %q - naming by number", awsName)
		}

		c.AddTask(&awstasks.SecurityGroupRule{
			Name:          fi.String("node-to-master-protocol-" + name),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
			SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			Protocol:      fi.String(awsName),
		})
	}
}

func (b *FirewallModelBuilder) buildMasterRules(c *fi.ModelBuilderContext) error {
	name := "masters." + b.ClusterName()

	{
		t := &awstasks.SecurityGroup{
			Name:        fi.String(name),
			Lifecycle:   b.Lifecycle,
			VPC:         b.LinkToVPC(),
			Description: fi.String("Security group for masters"),
			RemoveExtraRules: []string{
				"port=22",   // SSH
				"port=443",  // k8s api
				"port=4001", // etcd main (etcd events is 4002)
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
			Name:          fi.String("master-egress"),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
			Egress:        fi.Bool(true),
			CIDR:          fi.String("0.0.0.0/0"),
		}
		c.AddTask(t)
	}

	// Masters can talk to masters
	{
		t := &awstasks.SecurityGroupRule{
			Name:          fi.String("all-master-to-master"),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
			SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
		}
		c.AddTask(t)
	}

	// Masters can talk to nodes
	{
		t := &awstasks.SecurityGroupRule{
			Name:          fi.String("all-master-to-node"),
			Lifecycle:     b.Lifecycle,
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			SourceGroup:   b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
		}
		c.AddTask(t)
	}

	return nil
}
