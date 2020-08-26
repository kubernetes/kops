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

package openstackmodel

import (
	"fmt"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstacktasks"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/wellknownports"

	sg "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/groups"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/rules"
)

const (
	IPProtocolTCP   = string(rules.ProtocolTCP)
	IPProtocolUDP   = string(rules.ProtocolUDP)
	IPV4            = string(rules.EtherType4)
	ProtocolIPEncap = "4" // IP in IPv4/IPv6
)

// FirewallModelBuilder configures firewall network objects
type FirewallModelBuilder struct {
	*OpenstackModelContext
	Lifecycle *fi.Lifecycle
	Rules     map[string]*openstacktasks.SecurityGroupRule
}

var _ fi.ModelBuilder = &FirewallModelBuilder{}

func (b *FirewallModelBuilder) usesOctavia() bool {
	if b.Cluster.Spec.CloudConfig != nil &&
		b.Cluster.Spec.CloudConfig.Openstack != nil &&
		b.Cluster.Spec.CloudConfig.Openstack.Loadbalancer != nil {
		return fi.BoolValue(b.Cluster.Spec.CloudConfig.Openstack.Loadbalancer.UseOctavia)
	}
	return false
}

// addDirectionalGroupRule - create a rule on the source group to the dest group provided a securityGroupRuleTask
//  Example
//  Create an Ingress rule on source allowing traffic from dest with the options in the SecurityGroupRule
//  Create an Egress rule on source allowing traffic to dest with the options in the SecurityGroupRule
func (b *FirewallModelBuilder) addDirectionalGroupRule(c *fi.ModelBuilderContext, source, dest *openstacktasks.SecurityGroup, sgr *openstacktasks.SecurityGroupRule) {
	t := &openstacktasks.SecurityGroupRule{
		Direction:      sgr.Direction,
		EtherType:      sgr.EtherType,
		Lifecycle:      sgr.Lifecycle,
		PortRangeMin:   sgr.PortRangeMin,
		PortRangeMax:   sgr.PortRangeMax,
		Protocol:       sgr.Protocol,
		RemoteGroup:    dest,
		RemoteIPPrefix: sgr.RemoteIPPrefix,
		SecGroup:       source,
		Delete:         fi.Bool(false),
	}

	klog.V(8).Infof("Adding rule %v", fi.StringValue(t.GetName()))
	b.Rules[fi.StringValue(t.GetName())] = t

}

// addSSHRules - sets the ssh rules based on the presence of a bastion
func (b *FirewallModelBuilder) addSSHRules(c *fi.ModelBuilderContext, sgMap map[string]*openstacktasks.SecurityGroup) error {

	masterName := b.SecurityGroupName(kops.InstanceGroupRoleMaster)
	nodeName := b.SecurityGroupName(kops.InstanceGroupRoleNode)
	bastionName := b.SecurityGroupName(kops.InstanceGroupRoleBastion)
	masterSG := sgMap[masterName]
	nodeSG := sgMap[nodeName]
	bastionSG := sgMap[bastionName]

	sshIngress := &openstacktasks.SecurityGroupRule{
		Lifecycle:    b.Lifecycle,
		Direction:    s(string(rules.DirIngress)),
		Protocol:     s(string(rules.ProtocolTCP)),
		EtherType:    s(string(rules.EtherType4)),
		PortRangeMin: i(22),
		PortRangeMax: i(22),
	}

	if b.UsesSSHBastion() {
		for _, sshAccess := range b.Cluster.Spec.SSHAccess {
			sshRule := &openstacktasks.SecurityGroupRule{
				Lifecycle:      b.Lifecycle,
				Direction:      s(string(rules.DirIngress)),
				Protocol:       s(string(rules.ProtocolTCP)),
				EtherType:      s(string(rules.EtherType4)),
				PortRangeMin:   i(22),
				PortRangeMax:   i(22),
				RemoteIPPrefix: s(sshAccess),
			}
			b.addDirectionalGroupRule(c, bastionSG, nil, sshRule)
		}
		//Allow ingress ssh from the bastion on the masters and nodes
		b.addDirectionalGroupRule(c, masterSG, bastionSG, sshIngress)
		b.addDirectionalGroupRule(c, nodeSG, bastionSG, sshIngress)
	} else {
		for _, sshAccess := range b.Cluster.Spec.SSHAccess {
			sshRule := &openstacktasks.SecurityGroupRule{
				Lifecycle:      b.Lifecycle,
				Direction:      s(string(rules.DirIngress)),
				Protocol:       s(string(rules.ProtocolTCP)),
				EtherType:      s(string(rules.EtherType4)),
				PortRangeMin:   i(22),
				PortRangeMax:   i(22),
				RemoteIPPrefix: s(sshAccess),
			}
			b.addDirectionalGroupRule(c, masterSG, nil, sshRule)
			b.addDirectionalGroupRule(c, nodeSG, nil, sshRule)
		}
	}
	return nil
}

// addETCDRules - Add ETCD access rules based on which CNI might need to access __ETCD_ENDPOINTS__
func (b *FirewallModelBuilder) addETCDRules(c *fi.ModelBuilderContext, sgMap map[string]*openstacktasks.SecurityGroup) error {

	masterName := b.SecurityGroupName(kops.InstanceGroupRoleMaster)
	nodeName := b.SecurityGroupName(kops.InstanceGroupRoleNode)
	masterSG := sgMap[masterName]
	nodeSG := sgMap[nodeName]

	// ETCD Peer Discovery
	etcdRule := &openstacktasks.SecurityGroupRule{
		Lifecycle:    b.Lifecycle,
		Direction:    s(string(rules.DirIngress)),
		Protocol:     s(string(rules.ProtocolTCP)),
		EtherType:    s(IPV4),
		PortRangeMin: i(4001),
		PortRangeMax: i(4002),
	}
	etcdPeerRule := &openstacktasks.SecurityGroupRule{
		Lifecycle:    b.Lifecycle,
		Direction:    s(string(rules.DirIngress)),
		Protocol:     s(string(rules.ProtocolTCP)),
		EtherType:    s(IPV4),
		PortRangeMin: i(2380),
		PortRangeMax: i(2381),
	}
	b.addDirectionalGroupRule(c, masterSG, masterSG, etcdRule)
	b.addDirectionalGroupRule(c, masterSG, masterSG, etcdPeerRule)

	for _, portRange := range wellknownports.ETCDPortRanges() {
		etcdMgmrRule := &openstacktasks.SecurityGroupRule{
			Lifecycle:    b.Lifecycle,
			Direction:    s(string(rules.DirIngress)),
			Protocol:     s(string(rules.ProtocolTCP)),
			EtherType:    s(string(rules.EtherType4)),
			PortRangeMin: i(portRange.Min),
			PortRangeMax: i(portRange.Max),
		}
		b.addDirectionalGroupRule(c, masterSG, masterSG, etcdMgmrRule)
	}

	if b.Cluster.Spec.Networking.Calico != nil {

		etcdCNIRule := &openstacktasks.SecurityGroupRule{
			Lifecycle:    b.Lifecycle,
			Direction:    s(string(rules.DirIngress)),
			Protocol:     s(string(rules.ProtocolTCP)),
			EtherType:    s(IPV4),
			PortRangeMin: i(4001),
			PortRangeMax: i(4001),
		}
		// Master access from other masters covered above
		// Allow nodes to reach ETCD endpoints
		b.addDirectionalGroupRule(c, masterSG, nodeSG, etcdCNIRule)
	}
	return nil
}

// addNodePortRules - Add node port rules to nodes give the NodePortRange
func (b *FirewallModelBuilder) addNodePortRules(c *fi.ModelBuilderContext, sgMap map[string]*openstacktasks.SecurityGroup) error {

	nodeName := b.SecurityGroupName(kops.InstanceGroupRoleNode)
	nodeSG := sgMap[nodeName]

	for _, nodePortAccess := range b.Cluster.Spec.NodePortAccess {

		nodePortRange, err := b.NodePortRange()
		if err != nil {
			return err
		}

		for _, protocol := range []string{IPProtocolTCP, IPProtocolUDP} {
			nodePortRule := &openstacktasks.SecurityGroupRule{
				Lifecycle:      b.Lifecycle,
				Direction:      s(string(rules.DirIngress)),
				Protocol:       s(protocol),
				EtherType:      s(IPV4),
				PortRangeMin:   i(nodePortRange.Base),
				PortRangeMax:   i(nodePortRange.Base + nodePortRange.Size - 1),
				RemoteIPPrefix: s(nodePortAccess),
			}
			b.addDirectionalGroupRule(c, nodeSG, nil, nodePortRule)
		}
	}
	return nil
}

// addHTTPSRules - Add rules to 443 access given the presence of a loadbalancer or not
func (b *FirewallModelBuilder) addHTTPSRules(c *fi.ModelBuilderContext, sgMap map[string]*openstacktasks.SecurityGroup, useVIPACL bool) error {

	masterName := b.SecurityGroupName(kops.InstanceGroupRoleMaster)
	nodeName := b.SecurityGroupName(kops.InstanceGroupRoleNode)
	lbSGName := b.Cluster.Spec.MasterPublicName
	lbSG := sgMap[lbSGName]
	masterSG := sgMap[masterName]
	nodeSG := sgMap[nodeName]

	httpsIngress := &openstacktasks.SecurityGroupRule{
		Lifecycle:    b.Lifecycle,
		Direction:    s(string(rules.DirIngress)),
		Protocol:     s(IPProtocolTCP),
		EtherType:    s(IPV4),
		PortRangeMin: i(443),
		PortRangeMax: i(443),
	}

	//Allow all local communication for kubernetes.svc and to the api.internal lb/gossip for kubelet's
	b.addDirectionalGroupRule(c, masterSG, nodeSG, httpsIngress)
	b.addDirectionalGroupRule(c, masterSG, masterSG, httpsIngress)

	if b.UseLoadBalancerForAPI() {
		if !useVIPACL {
			//Allow API Access to the lb sg
			for _, apiAccess := range b.Cluster.Spec.KubernetesAPIAccess {
				b.addDirectionalGroupRule(c, lbSG, nil, &openstacktasks.SecurityGroupRule{
					Lifecycle:      b.Lifecycle,
					Direction:      s(string(rules.DirIngress)),
					Protocol:       s(IPProtocolTCP),
					EtherType:      s(IPV4),
					PortRangeMin:   i(443),
					PortRangeMax:   i(443),
					RemoteIPPrefix: s(apiAccess),
				})
			}
			//Allow masters ingress from the sg
			b.addDirectionalGroupRule(c, masterSG, lbSG, httpsIngress)
		}

		//FIXME: Octavia port traffic appears to be denied though its port is in lbSG
		if b.usesOctavia() {
			b.addDirectionalGroupRule(c, masterSG, nil, &openstacktasks.SecurityGroupRule{
				Lifecycle:      b.Lifecycle,
				Direction:      s(string(rules.DirIngress)),
				Protocol:       s(IPProtocolTCP),
				EtherType:      s(IPV4),
				PortRangeMin:   i(443),
				PortRangeMax:   i(443),
				RemoteIPPrefix: s(b.Cluster.Spec.NetworkCIDR),
			})
		}

	} else {
		// Allow the masters to receive connections from KubernetesAPIAccess
		for _, apiAccess := range b.Cluster.Spec.KubernetesAPIAccess {

			b.addDirectionalGroupRule(c, masterSG, nil, &openstacktasks.SecurityGroupRule{
				Lifecycle:      b.Lifecycle,
				Direction:      s(string(rules.DirIngress)),
				Protocol:       s(IPProtocolTCP),
				EtherType:      s(IPV4),
				PortRangeMin:   i(443),
				PortRangeMax:   i(443),
				RemoteIPPrefix: s(apiAccess),
			})
		}
	}

	return nil
}

// addKubeletRules - Add rules to 10250 port
func (b *FirewallModelBuilder) addKubeletRules(c *fi.ModelBuilderContext, sgMap map[string]*openstacktasks.SecurityGroup) error {

	//TODO: This is the default port for kubelet and may be overridden
	masterName := b.SecurityGroupName(kops.InstanceGroupRoleMaster)
	nodeName := b.SecurityGroupName(kops.InstanceGroupRoleNode)
	masterSG := sgMap[masterName]
	nodeSG := sgMap[nodeName]

	kubeletRule := &openstacktasks.SecurityGroupRule{
		Lifecycle:    b.Lifecycle,
		Direction:    s(string(rules.DirIngress)),
		Protocol:     s(IPProtocolTCP),
		EtherType:    s(IPV4),
		PortRangeMin: i(10250),
		PortRangeMax: i(10250),
	}

	// allow node-node, node-master and master-master and master-node
	for _, sgName := range []*openstacktasks.SecurityGroup{masterSG, nodeSG} {
		b.addDirectionalGroupRule(c, masterSG, sgName, kubeletRule)
		b.addDirectionalGroupRule(c, nodeSG, sgName, kubeletRule)
	}
	return nil
}

// addNodeExporterRules - Allow 9100 TCP port from nodesg
func (b *FirewallModelBuilder) addNodeExporterRules(c *fi.ModelBuilderContext, sgMap map[string]*openstacktasks.SecurityGroup) error {
	masterName := b.SecurityGroupName(kops.InstanceGroupRoleMaster)
	nodeName := b.SecurityGroupName(kops.InstanceGroupRoleNode)
	masterSG := sgMap[masterName]
	nodeSG := sgMap[nodeName]
	nodeExporterIngress := &openstacktasks.SecurityGroupRule{
		Lifecycle:    b.Lifecycle,
		Direction:    s(string(rules.DirIngress)),
		Protocol:     s(IPProtocolTCP),
		EtherType:    s(IPV4),
		PortRangeMin: i(9100),
		PortRangeMax: i(9100),
	}
	// allow 9100 port from nodeSG
	b.addDirectionalGroupRule(c, masterSG, nodeSG, nodeExporterIngress)
	b.addDirectionalGroupRule(c, nodeSG, nodeSG, nodeExporterIngress)
	return nil
}

// addDNSRules - Add DNS rules for internal DNS queries
func (b *FirewallModelBuilder) addDNSRules(c *fi.ModelBuilderContext, sgMap map[string]*openstacktasks.SecurityGroup) error {

	masterName := b.SecurityGroupName(kops.InstanceGroupRoleMaster)
	nodeName := b.SecurityGroupName(kops.InstanceGroupRoleNode)
	masterSG := sgMap[masterName]
	nodeSG := sgMap[nodeName]
	for _, protocol := range []string{IPProtocolTCP, IPProtocolUDP} {
		dnsRule := &openstacktasks.SecurityGroupRule{
			Lifecycle:    b.Lifecycle,
			Direction:    s(string(rules.DirIngress)),
			Protocol:     s(protocol),
			EtherType:    s(IPV4),
			PortRangeMin: i(53),
			PortRangeMax: i(53),
		}
		b.addDirectionalGroupRule(c, masterSG, nodeSG, dnsRule)
		b.addDirectionalGroupRule(c, nodeSG, masterSG, dnsRule)
		b.addDirectionalGroupRule(c, masterSG, masterSG, dnsRule)
	}
	return nil
}

// addCNIRules - Add ports required for different CNI implementations
func (b *FirewallModelBuilder) addCNIRules(c *fi.ModelBuilderContext, sgMap map[string]*openstacktasks.SecurityGroup) error {

	udpPorts := []int{}
	tcpPorts := []int{}
	protocols := []string{}

	// allow cadvisor
	tcpPorts = append(tcpPorts, 4194)

	if b.Cluster.Spec.Networking != nil {
		if b.Cluster.Spec.Networking.Kopeio != nil {
			// VXLAN over UDP
			// https://tools.ietf.org/html/rfc7348
			udpPorts = append(udpPorts, 4789)
		}

		if b.Cluster.Spec.Networking.Cilium != nil {
			udpPorts = append(udpPorts, 8472)
			tcpPorts = append(tcpPorts, 4240)
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
				klog.Warningf("unknown flannel networking backend %q", b.Cluster.Spec.Networking.Flannel.Backend)
			}
		}

		if b.Cluster.Spec.Networking.Calico != nil {
			tcpPorts = append(tcpPorts, 179)
			protocols = append(protocols, ProtocolIPEncap)
		}

		if b.Cluster.Spec.Networking.Kuberouter != nil {
			protocols = append(protocols, ProtocolIPEncap)
		}
	}

	masterName := b.SecurityGroupName(kops.InstanceGroupRoleMaster)
	nodeName := b.SecurityGroupName(kops.InstanceGroupRoleNode)
	masterSG := sgMap[masterName]
	nodeSG := sgMap[nodeName]

	for _, udpPort := range udpPorts {
		udpRule := &openstacktasks.SecurityGroupRule{
			Lifecycle:    b.Lifecycle,
			Direction:    s(string(rules.DirIngress)),
			Protocol:     s(string(rules.ProtocolUDP)),
			EtherType:    s(string(rules.EtherType4)),
			PortRangeMin: i(udpPort),
			PortRangeMax: i(udpPort),
		}
		b.addDirectionalGroupRule(c, masterSG, masterSG, udpRule)
		b.addDirectionalGroupRule(c, nodeSG, masterSG, udpRule)
		b.addDirectionalGroupRule(c, masterSG, nodeSG, udpRule)
		b.addDirectionalGroupRule(c, nodeSG, nodeSG, udpRule)
	}
	for _, tcpPort := range tcpPorts {
		tcpRule := &openstacktasks.SecurityGroupRule{
			Lifecycle:    b.Lifecycle,
			Direction:    s(string(rules.DirIngress)),
			Protocol:     s(string(rules.ProtocolTCP)),
			EtherType:    s(string(rules.EtherType4)),
			PortRangeMin: i(tcpPort),
			PortRangeMax: i(tcpPort),
		}
		b.addDirectionalGroupRule(c, masterSG, masterSG, tcpRule)
		b.addDirectionalGroupRule(c, nodeSG, masterSG, tcpRule)
		b.addDirectionalGroupRule(c, masterSG, nodeSG, tcpRule)
		b.addDirectionalGroupRule(c, nodeSG, nodeSG, tcpRule)
	}

	for _, protocol := range protocols {
		protocolRule := &openstacktasks.SecurityGroupRule{
			Lifecycle: b.Lifecycle,
			Direction: s(string(rules.DirIngress)),
			Protocol:  s(protocol),
			EtherType: s(string(rules.EtherType4)),
		}
		b.addDirectionalGroupRule(c, masterSG, nil, protocolRule)
		b.addDirectionalGroupRule(c, nodeSG, nil, protocolRule)
	}

	return nil
}

// addProtokubeRules - Add rules for protokube if gossip DNS is enabled
func (b *FirewallModelBuilder) addProtokubeRules(c *fi.ModelBuilderContext, sgMap map[string]*openstacktasks.SecurityGroup) error {

	if dns.IsGossipHostname(b.ClusterName()) {
		masterName := b.SecurityGroupName(kops.InstanceGroupRoleMaster)
		nodeName := b.SecurityGroupName(kops.InstanceGroupRoleNode)
		masterSG := sgMap[masterName]
		nodeSG := sgMap[nodeName]
		for _, portRange := range wellknownports.DNSGossipPortRanges() {
			protokubeRule := &openstacktasks.SecurityGroupRule{
				Lifecycle:    b.Lifecycle,
				Direction:    s(string(rules.DirIngress)),
				Protocol:     s(string(rules.ProtocolTCP)),
				EtherType:    s(string(rules.EtherType4)),
				PortRangeMin: i(portRange.Min),
				PortRangeMax: i(portRange.Max),
			}
			b.addDirectionalGroupRule(c, masterSG, nodeSG, protokubeRule)
			b.addDirectionalGroupRule(c, nodeSG, masterSG, protokubeRule)
			b.addDirectionalGroupRule(c, masterSG, masterSG, protokubeRule)
			b.addDirectionalGroupRule(c, nodeSG, nodeSG, protokubeRule)
		}
	}
	return nil
}

func (b *FirewallModelBuilder) getExistingRules(sgMap map[string]*openstacktasks.SecurityGroup) error {
	osCloud, err := b.createCloud()
	if err != nil {
		return err
	}
	sgIdMap := make(map[string]*openstacktasks.SecurityGroup)
	for sgName, sgt := range sgMap {

		sgs, err := osCloud.ListSecurityGroups(sg.ListOpts{
			Name: sgName,
		})
		if err != nil {
			return err
		}
		if len(sgs) != 1 {
			continue
		}
		sg := sgs[0]
		sgt.Name = fi.String(sg.Name)
		sgIdMap[sg.ID] = sgt
	}

	for sgid := range sgIdMap {

		sgRules, err := osCloud.ListSecurityGroupRules(rules.ListOpts{
			SecGroupID: sgid,
		})
		if err != nil {
			return err
		}
		for _, rule := range sgRules {

			t := &openstacktasks.SecurityGroupRule{
				ID:             fi.String(rule.ID),
				Direction:      fi.String(rule.Direction),
				EtherType:      fi.String(rule.EtherType),
				PortRangeMax:   fi.Int(rule.PortRangeMax),
				PortRangeMin:   fi.Int(rule.PortRangeMin),
				Protocol:       fi.String(rule.Protocol),
				RemoteIPPrefix: fi.String(rule.RemoteIPPrefix),
				RemoteGroup:    sgIdMap[rule.RemoteGroupID],
				Lifecycle:      b.Lifecycle,
				SecGroup:       sgIdMap[rule.SecGroupID],
				Delete:         fi.Bool(true),
			}
			klog.V(8).Infof("Adding existing rule %v", t)
			b.Rules[fi.StringValue(t.GetName())] = t
		}
	}
	return nil

}

func (b *FirewallModelBuilder) addDefaultEgress(c *fi.ModelBuilderContext, sgMap map[string]*openstacktasks.SecurityGroup) {
	for _, sg := range sgMap {
		t := &openstacktasks.SecurityGroupRule{
			Lifecycle:    b.Lifecycle,
			Direction:    s(string(rules.DirEgress)),
			EtherType:    s(string(rules.EtherType4)),
			PortRangeMin: i(0),
			PortRangeMax: i(0),
		}
		b.addDirectionalGroupRule(c, sg, nil, t)

		t = &openstacktasks.SecurityGroupRule{
			Lifecycle:    b.Lifecycle,
			Direction:    s(string(rules.DirEgress)),
			EtherType:    s(string(rules.EtherType6)),
			PortRangeMin: i(0),
			PortRangeMax: i(0),
		}
		b.addDirectionalGroupRule(c, sg, nil, t)
	}
}

// Build - schedule security groups and security group rule tasks for Openstack
func (b *FirewallModelBuilder) Build(c *fi.ModelBuilderContext) error {
	roles := []kops.InstanceGroupRole{kops.InstanceGroupRoleMaster, kops.InstanceGroupRoleNode}
	if b.UsesSSHBastion() {
		roles = append(roles, kops.InstanceGroupRoleBastion)
	}

	sgMap := make(map[string]*openstacktasks.SecurityGroup)

	useVIPACL := false
	if b.UseLoadBalancerForAPI() && b.UseVIPACL() {
		useVIPACL = true
	}
	sg := &openstacktasks.SecurityGroup{
		Name:             s(b.Cluster.Spec.MasterPublicName),
		Lifecycle:        b.Lifecycle,
		RemoveExtraRules: []string{"port=443"},
	}
	if useVIPACL {
		sg.RemoveGroup = true
	}
	c.AddTask(sg)
	sgMap[b.Cluster.Spec.MasterPublicName] = sg
	for _, role := range roles {

		// Create Security Group for Role
		groupName := b.SecurityGroupName(role)
		sg := &openstacktasks.SecurityGroup{
			Name:        s(groupName),
			Lifecycle:   b.Lifecycle,
			RemoveGroup: false,
		}
		if role == kops.InstanceGroupRoleBastion {
			sg.RemoveExtraRules = []string{"port=22"}
		} else if role == kops.InstanceGroupRoleNode {
			sg.RemoveExtraRules = []string{"port=22", "port=10250"}
		} else if role == kops.InstanceGroupRoleMaster {
			sg.RemoveExtraRules = []string{"port=22", "port=443", "port=10250"}
		}
		c.AddTask(sg)
		sgMap[groupName] = sg
	}

	b.Rules = make(map[string]*openstacktasks.SecurityGroupRule)

	err := b.getExistingRules(sgMap)
	// if this fails, it just means that existing rules won't be cleaned up
	if err != nil {
		klog.Warningf("Failed to list existing security groups: %v", err)
	}

	b.addDefaultEgress(c, sgMap)

	//Add API Server Rules
	b.addHTTPSRules(c, sgMap, useVIPACL)

	//Add SSH
	b.addSSHRules(c, sgMap)
	//Allow overlay DNS
	b.addDNSRules(c, sgMap)
	//Add Kubelet Rules
	b.addKubeletRules(c, sgMap)
	//Add Node exporter Rules
	b.addNodeExporterRules(c, sgMap)
	// Protokube Rules
	b.addProtokubeRules(c, sgMap)
	//Allow necessary local traffic
	b.addCNIRules(c, sgMap)
	//ETCD Leader Election
	b.addETCDRules(c, sgMap)
	// Add NodePort Rules:
	err = b.addNodePortRules(c, sgMap)
	if err != nil {
		return fmt.Errorf("failed to add node port rules: %v", err)
	}

	for _, r := range b.Rules {
		c.AddTask(r)
	}

	return nil

}
