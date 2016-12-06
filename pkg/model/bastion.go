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
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"github.com/kopeio/gladish/pkg/sets"
	"time"
)

const BastionELBSecurityGroupPrefix = "bastion"
const BastionELBIdleTimeout = 2 * time.Minute

//
//type BastionSpec struct {
//	// Controls if a private topology should deploy a bastion host or not
//	// The bastion host is designed to be a simple, and secure bridge between
//	// the public subnet and the private subnet
//	Enable      bool   `json:"enable,omitempty"`
//	MachineType string `json:"machineType,omitempty"`
//	PublicName  string `json:"name,omitempty"`
//	// Bastion's Loadbalancer idle timeout
//	IdleTimeout int `json:"idleTimeout,omitempty"`
//}


// BastionModelBuilder add model objects to support bastions
type BastionModelBuilder struct {
	*KopsModelContext
}

var _ fi.ModelBuilder = &BastionModelBuilder{}

func (b *BastionModelBuilder) Build(c *fi.ModelBuilderContext) error {
	//{{ if and WithBastion IsTopologyPrivate }}
	//securityGroupRule/bastion-to-master:
	//securityGroup: securityGroup/masters.{{ ClusterName }}
	//sourceGroup: securityGroup/bastion.{{ ClusterName }}
	//{{ end }}
	var bastionGroups []*kops.InstanceGroup
	for _, ig := range b.InstanceGroups {
		if ig.Spec.Role == kops.InstanceGroupRoleBastion {
			bastionGroups = append(bastionGroups, ig)
		}
	}

	if len(bastionGroups) == 0 {
		return nil
	}

	//	-{{ if and WithBastion IsTopologyPrivate }}
	//-securityGroupRule/bastion-to-master:
	//-  securityGroup: securityGroup/masters.{{ ClusterName }}
	//-  sourceGroup: securityGroup/bastion.{{ ClusterName }}
	//-{{ end }}

	// TODO: Replace with objects and LinkTo functions
	//bastionSecurityGroupName := b.SecurityGroupName(kops.InstanceGroupRoleBastion)
	//bastionELBSecurityGroupName := b.ELBSecurityGroupName("bastion")

	// Create security group for bastion instances
	{
		t := &awstasks.SecurityGroup{
			Name: s(b.SecurityGroupName(kops.InstanceGroupRoleBastion)),
			VPC: b.LinkToVPC(),
			Description: s("Security group for bastion"),
			RemoveExtraRules: []string{"22" },
		}
		c.AddTask(t)
	}



	//// Allow traffic from bastion instances to egress freely
	//{
	//	r := &awstasks.SecurityGroupRule{
	//		Name: "bastion-egress",
	//		SecurityGroup: &awstasks.SecurityGroup{Name: bastionSGName},
	//		Egress: true,
	//		CIDR: "0.0.0.0/0",
	//	}
	//
	//	c.Tasks[r.Name] = r
	//}

	//-# TODO Kris - I don't think we need to open these
	//-#securityGroupRule/all-node-to-bastion:
	//-#  securityGroup: securityGroup/bastion.{{ ClusterName }}
	//-#  sourceGroup: securityGroup/nodes.{{ ClusterName }}
	//-#securityGroupRule/all-master-to-bastion:
	//-#  securityGroup: securityGroup/bastion.{{ ClusterName }}
	//-#  sourceGroup: securityGroup/masters.{{ ClusterName }}


	// Allow incoming SSH traffic to bastions, through the ELB
	// TODO: Could we get away without an ELB here?  Tricky if dns-controller is broken though...
	{
		rule := &awstasks.SecurityGroupRule{
			Name: s("ssh-external-to-bastion"),
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleBastion),
			SourceGroup: b.LinkToELBSecurityGroup(BastionELBSecurityGroupPrefix),
			Protocol: s("tcp"),
			FromPort: i64(22),
			ToPort: i64(22),
		}
		c.AddTask(rule)
	}


	// Allow bastion nodes to reach masters
	{
		rule := &awstasks.SecurityGroupRule{
			Name: s("bastion-to-master"),
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
			SourceGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleBastion),
		}
		c.AddTask(rule)
	}


	// Allow bastion nodes to reach nodes
	// If we are creating a bastion, we need to poke a hole in the
	// firewall for it to talk to our masters
	{
		rule := &awstasks.SecurityGroupRule{
			Name: s("bastion-to-nodes"),
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
			SourceGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleBastion),
		}
		c.AddTask(rule)
	}



	// Create security group for bastion ELB
	{
		t := &awstasks.SecurityGroup{
			Name: s(b.ELBSecurityGroupName(BastionELBSecurityGroupPrefix)),
			VPC: b.LinkToVPC(),
			Description: s("Security group for bastion ELB"),
			RemoveExtraRules: []string{"22"},
		}
		c.AddTask(t)
	}

	// Allow ELB egress
	//{
	//	-securityGroupRule/bastion-elb-egress:
	//	-  securityGroup: securityGroup/bastion-elb.{{ ClusterName }}
	//	-  egress: true
	//	-  cidr: 0.0.0.0/0
	//}


	// Allow external access to ELB
	for _, sshAccess := range b.Cluster.Spec.SSHAccess {
		t := &awstasks.SecurityGroupRule{
			Name: s("ssh-external-to-bastion-elb-" + sshAccess),
			SecurityGroup: b.LinkToELBSecurityGroup(BastionELBSecurityGroupPrefix),
			Protocol: s("tcp"),
			FromPort: i64(22),
			ToPort: i64(22),
			CIDR: s(sshAccess),
		}
		c.AddTask(t)
	}

	var elbSubnets []*awstasks.Subnet
	{
		//	{{ range $zone :=.Zones }}
		//- subnet/utility-{{ $zone.Name }}.{{ ClusterName }}
		//{{ end }}

		zones := sets.NewString()
		for _, ig := range bastionGroups {
			subnets, err := b.GatherSubnets(ig)
			if err != nil {
				return err
			}
			for _, s := range subnets {
				zones.Insert(s.Zone)
			}
		}

		for zoneName := range zones {
			utilitySubnet, err := b.LinkToPublicSubnetInZone(zoneName)
			if err != nil {
				return err
			}
			elbSubnets = append(elbSubnets, utilitySubnet)
		}
	}

	// Create ELB itself
	var elb *awstasks.LoadBalancer
	{
		elbID, err := b.GetELBName32("bastion")
		if err != nil {
			return err
		}

		elb = &awstasks.LoadBalancer{
			Name: s("bastion." + b.ClusterName()),
			ID: s(elbID),
			SecurityGroups: []*awstasks.SecurityGroup{
				b.LinkToELBSecurityGroup(BastionELBSecurityGroupPrefix),
			},
			Subnets: elbSubnets,
			Listeners: map[string]*awstasks.LoadBalancerListener{
				"22": {InstancePort: 22},
			},
		}

		c.AddTask(elb)
	}

	for _, ig := range bastionGroups {
		asg, err := b.buildASG(ig)
		if err != nil {
			return err
		}

		// Attach the ELB to the ASG
		{
			t := &awstasks.LoadBalancerAttachment{
				Name: s("bastion-elb-attachment"),
				LoadBalancer: elb,
				AutoscalingGroup: asg,
			}
			c.AddTask(t)
		}
	}

	// Build ELB attributes
	{
		//idleTimeout := GetBastionIdleTimeout()
		idleTimeout := BastionELBIdleTimeout

		//-loadBalancerConnectionSettings/bastion.{{ ClusterName }}:
		//-  loadBalancer: loadBalancer/bastion.{{ ClusterName }}
		//-  idleTimeout: {{ GetBastionIdleTimeout }}
		elbSettings := &awstasks.LoadBalancerConnectionSettings{
			//Name: elb.Name,
			//LoadBalancer: elb,
			IdleTimeout: i64(int64(idleTimeout.Seconds())),
		}

		//		-# ---------------------------------------------------------------------
		//		-# Loadbalancer attributes are configurable now
		//	-# By default ELB has an idle timeout of 60 seconds to close connection
		//-# Modified the idle timeout for bastion elb
		//-# --------------------------------------------------------------------
		//-loadBalancerAttributes/bastion.{{ ClusterName }}:
		//-  loadBalancer: loadBalancer/bastion.{{ ClusterName }}
		//-  connectionSettings: loadBalancerConnectionSettings/bastion.{{ ClusterName }}
		t := &awstasks.LoadBalancerAttributes{
			Name: elb.Name,
			LoadBalancer: elb,
			ConnectionSettings: elbSettings,
		}
		c.AddTask(t)
	}

	// TODO: Re-enable bastion DNS
	bastionDNS := "" //getBastionDNS()
	if bastionDNS != "" {
		//-{{ if IsBastionDNS }}
		//-# ------------------------------------------------------------------------
		//-# By default Bastion is not reachable from outside because of security concerns.
		//-# But if the user specifies bastion name using edit cluster, we configure
		//-# the bastion DNS entry for it to be reachable from outside.
		//-# BastionPublicName --> Bastion LoadBalancer
		//-# ------------------------------------------------------------------------
		//-dnsName/{{ GetBastionDNS }}:
		//-  Zone: dnsZone/{{ .DNSZone }}
		//-  ResourceType: "A"
		//-  TargetLoadBalancer: loadBalancer/bastion.{{ ClusterName }}
		//-{{ end }}
		//-{{ end }}

		t := &awstasks.DNSName{
			Name: s(bastionDNS),
			Zone: b.LinkToDNSZone(),
			ResourceType: s("A"),
			TargetLoadBalancer: elb,
		}
		c.AddTask(t)
	}
	return nil
}

func (b*BastionModelBuilder) buildASG(ig *kops.InstanceGroup) (*awstasks.AutoscalingGroup, error) {
	//-# ---------------------------------------------------------------
	//-# ASG - The Bastion itself
	//-#
	//-# Define the bastion host.
	//-# Machine type configurable.
	//-#
	//-# The bastion host will live in one of the utility subnets
	//-# created in the private topology. The bastion host will have
	//-# port 22 TCP open to 0.0.0.0/0. And will have internal SSH
	//-# access to all private subnets.
	//-#
	//-# ---------------------------------------------------------------

	name := ig.ObjectMeta.Name + "." + b.ClusterName()

	sshKey, err := b.LinkToSSHKey()
	if err != nil {
		return nil, err
	}

	lc := &awstasks.LaunchConfiguration{
		Name:  s(name),
		SSHKey: sshKey,
		SecurityGroups: []*awstasks.SecurityGroup{
			b.LinkToSecurityGroup(ig.Spec.Role),
		},
		IAMInstanceProfile: b.LinkToIAMInstanceProfile(ig),
		ImageID: s(ig.Spec.Image),
		InstanceType: s(ig.Spec.MachineType),
		AssociatePublicIP: fi.Bool(false),
		RootVolumeSize: i64(20),
		RootVolumeType:s("gp2"),
	}

	var subnets []*awstasks.Subnet
	{
		subnetSpecs, err := b.GatherSubnets(ig)
		if err != nil {
			return nil, err
		}
		for _, subnetSpec := range subnetSpecs {
			subnet := b.LinkToSubnet(subnetSpec)
			subnets = append(subnets, subnet)
		}
	}

	asg := &awstasks.AutoscalingGroup{
		Name: s(name),
		MinSize: i64(1),
		MaxSize: i64(1),
		LaunchConfiguration: lc,
		Tags: map[string]string{
			"Name": name,
			// TODO: Aren't these added automatically?
			"KubernetesCluster": b.ClusterName(),
		},
		Subnets: subnets,
	}

	return asg, nil
}

//func (tf *TemplateFunctions) IsBastionDNS() bool {
//	if tf.cluster.Spec.Topology.Bastion.PublicName == "" {
//		return false
//	} else {
//		return true
//	}
//}
//
//func (tf *TemplateFunctions) GetBastionDNS() string {
//	return tf.cluster.GetBastionPublicName()
//}

//// This function is replacing existing yaml
//func (tf *TemplateFunctions) GetBastionZone() (string, error) {
//	var name string
//	if len(tf.cluster.Spec.Zones) < 1 {
//		return "", fmt.Errorf("Unable to detect zone name for bastion")
//	} else {
//		// If we have a list, always use the first one
//		name = tf.cluster.Spec.Zones[0].Name
//	}
//	return name, nil
//}

//func (tf *TemplateFunctions) GetBastionMachineType() (string, error) {
//	defaultMachineType := tf.cluster.GetBastionMachineType()
//	if defaultMachineType == "" {
//		return "", fmt.Errorf("DefaultMachineType for bastion can not be empty")
//	}
//	return defaultMachineType, nil
//}

//func (tf *TemplateFunctions) GetBastionIdleTimeout() (int, error) {
//	timeout := tf.cluster.GetBastionIdleTimeout()
//	if timeout <= 0 {
//		return 0, fmt.Errorf("IdleTimeout for Bastion can not be negative")
//	}
//	return timeout, nil
//}


//func (tf *TemplateFunctions) GetBastionImageId() (string, error) {
//	if len(tf.instanceGroups) == 0 {
//		return "", fmt.Errorf("Unable to find AMI in instance group")
//	} else if len(tf.instanceGroups) > 0 {
//		ami := tf.instanceGroups[0].Spec.Image
//		for i := 1; i < len(tf.instanceGroups); i++ {
//			// If we can't be sure all AMIs are the same, we don't know which one to use for the bastion host
//			if tf.instanceGroups[i].Spec.Image != ami {
//				return "", fmt.Errorf("Unable to use multiple image id's with a private bastion")
//			}
//		}
//		return ami, nil
//	}
//	return "", nil
//}


//func (c *Cluster) GetBastionMachineType() string {
//	return c.Spec.Topology.Bastion.MachineType
//}
//func (c *Cluster) GetBastionPublicName() string {
//	return c.Spec.Topology.Bastion.PublicName
//}
//func (c *Cluster) GetBastionIdleTimeout() int {
//	return c.Spec.Topology.Bastion.IdleTimeout
//}


