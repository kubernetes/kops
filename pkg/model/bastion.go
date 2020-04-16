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
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
)

const (
	BastionELBSecurityGroupPrefix = "bastion"
	BastionELBDefaultIdleTimeout  = 5 * time.Minute
)

// BastionModelBuilder adds model objects to support bastions
//
// Bastion instances live in the utility subnets created in the private topology.
// All traffic goes through an ELB, and the ELB has port 22 open to SSHAccess.
// Bastion instances have access to all internal master and node instances.

type BastionModelBuilder struct {
	*KopsModelContext
	Lifecycle         *fi.Lifecycle
	SecurityLifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &BastionModelBuilder{}

func (b *BastionModelBuilder) Build(c *fi.ModelBuilderContext) error {
	var bastionInstanceGroups []*kops.InstanceGroup
	for _, ig := range b.InstanceGroups {
		if ig.Spec.Role == kops.InstanceGroupRoleBastion {
			bastionInstanceGroups = append(bastionInstanceGroups, ig)
		}
	}

	if len(bastionInstanceGroups) == 0 {
		return nil
	}

	bastionGroups, err := b.GetSecurityGroups(kops.InstanceGroupRoleBastion)
	if err != nil {
		return err
	}
	nodeGroups, err := b.GetSecurityGroups(kops.InstanceGroupRoleNode)
	if err != nil {
		return err
	}
	masterGroups, err := b.GetSecurityGroups(kops.InstanceGroupRoleMaster)
	if err != nil {
		return err
	}

	// Create security group for bastion instances
	for _, bastionGroup := range bastionGroups {
		bastionGroup.Task.Lifecycle = b.SecurityLifecycle
		c.AddTask(bastionGroup.Task)
	}

	for _, src := range bastionGroups {
		// Allow traffic from bastion instances to egress freely
		t := &awstasks.SecurityGroupRule{
			Name:          s("bastion-egress" + src.Suffix),
			Lifecycle:     b.SecurityLifecycle,
			SecurityGroup: src.Task,
			Egress:        fi.Bool(true),
			CIDR:          s("0.0.0.0/0"),
		}
		c.AddTask(t)
	}

	// Allow incoming SSH traffic to bastions, through the ELB
	// TODO: Could we get away without an ELB here?  Tricky to fix if dns-controller breaks though...
	for _, dest := range bastionGroups {
		t := &awstasks.SecurityGroupRule{
			Name:          s("ssh-elb-to-bastion" + dest.Suffix),
			Lifecycle:     b.SecurityLifecycle,
			SecurityGroup: dest.Task,
			SourceGroup:   b.LinkToELBSecurityGroup(BastionELBSecurityGroupPrefix),
			Protocol:      s("tcp"),
			FromPort:      i64(22),
			ToPort:        i64(22),
		}
		c.AddTask(t)
	}

	// Allow bastion nodes to SSH to masters
	for _, src := range bastionGroups {
		for _, dest := range masterGroups {
			t := &awstasks.SecurityGroupRule{
				Name:          s("bastion-to-master-ssh" + JoinSuffixes(src, dest)),
				Lifecycle:     b.SecurityLifecycle,
				SecurityGroup: dest.Task,
				SourceGroup:   src.Task,
				Protocol:      s("tcp"),
				FromPort:      i64(22),
				ToPort:        i64(22),
			}
			c.AddTask(t)
		}
	}

	// Allow bastion nodes to SSH to nodes
	for _, src := range bastionGroups {
		for _, dest := range nodeGroups {
			t := &awstasks.SecurityGroupRule{
				Name:          s("bastion-to-node-ssh" + JoinSuffixes(src, dest)),
				Lifecycle:     b.SecurityLifecycle,
				SecurityGroup: dest.Task,
				SourceGroup:   src.Task,
				Protocol:      s("tcp"),
				FromPort:      i64(22),
				ToPort:        i64(22),
			}
			c.AddTask(t)
		}
	}

	// Create security group for bastion ELB
	{
		t := &awstasks.SecurityGroup{
			Name:      s(b.ELBSecurityGroupName(BastionELBSecurityGroupPrefix)),
			Lifecycle: b.SecurityLifecycle,

			VPC:              b.LinkToVPC(),
			Description:      s("Security group for bastion ELB"),
			RemoveExtraRules: []string{"port=22"},
		}
		t.Tags = b.CloudTags(*t.Name, false)
		c.AddTask(t)
	}

	// Allow traffic from ELB to egress freely
	{
		t := &awstasks.SecurityGroupRule{
			Name:      s("bastion-elb-egress"),
			Lifecycle: b.SecurityLifecycle,

			SecurityGroup: b.LinkToELBSecurityGroup(BastionELBSecurityGroupPrefix),
			Egress:        fi.Bool(true),
			CIDR:          s("0.0.0.0/0"),
		}

		c.AddTask(t)
	}

	// Allow external access to ELB
	for _, sshAccess := range b.Cluster.Spec.SSHAccess {
		t := &awstasks.SecurityGroupRule{
			Name:      s("ssh-external-to-bastion-elb-" + sshAccess),
			Lifecycle: b.SecurityLifecycle,

			SecurityGroup: b.LinkToELBSecurityGroup(BastionELBSecurityGroupPrefix),
			Protocol:      s("tcp"),
			FromPort:      i64(22),
			ToPort:        i64(22),
			CIDR:          s(sshAccess),
		}
		c.AddTask(t)
	}

	var elbSubnets []*awstasks.Subnet
	{
		zones := sets.NewString()
		for _, ig := range bastionInstanceGroups {
			subnets, err := b.GatherSubnets(ig)
			if err != nil {
				return err
			}
			for _, s := range subnets {
				zones.Insert(s.Zone)
			}
		}

		for zoneName := range zones {
			utilitySubnet, err := b.LinkToUtilitySubnetInZone(zoneName)
			if err != nil {
				return err
			}
			elbSubnets = append(elbSubnets, utilitySubnet)
		}
	}

	// Create ELB itself
	var elb *awstasks.LoadBalancer
	{
		loadBalancerName := b.GetELBName32("bastion")

		idleTimeout := BastionELBDefaultIdleTimeout
		if b.Cluster.Spec.Topology != nil && b.Cluster.Spec.Topology.Bastion != nil && b.Cluster.Spec.Topology.Bastion.IdleTimeoutSeconds != nil {
			idleTimeout = time.Second * time.Duration(*b.Cluster.Spec.Topology.Bastion.IdleTimeoutSeconds)
		}

		tags := b.CloudTags(loadBalancerName, false)
		for k, v := range b.Cluster.Spec.CloudLabels {
			tags[k] = v
		}
		// Override the returned name to be the expected ELB name
		tags["Name"] = "bastion." + b.ClusterName()

		elb = &awstasks.LoadBalancer{
			Name:      s("bastion." + b.ClusterName()),
			Lifecycle: b.Lifecycle,

			LoadBalancerName: s(loadBalancerName),
			SecurityGroups: []*awstasks.SecurityGroup{
				b.LinkToELBSecurityGroup(BastionELBSecurityGroupPrefix),
			},
			Subnets: elbSubnets,
			Listeners: map[string]*awstasks.LoadBalancerListener{
				"22": {InstancePort: 22},
			},

			HealthCheck: &awstasks.LoadBalancerHealthCheck{
				Target:             s("TCP:22"),
				Timeout:            i64(5),
				Interval:           i64(10),
				HealthyThreshold:   i64(2),
				UnhealthyThreshold: i64(2),
			},

			ConnectionSettings: &awstasks.LoadBalancerConnectionSettings{
				IdleTimeout: i64(int64(idleTimeout.Seconds())),
			},

			Tags: tags,
		}
		// Add additional security groups to the ELB
		if b.Cluster.Spec.Topology != nil && b.Cluster.Spec.Topology.Bastion != nil && b.Cluster.Spec.Topology.Bastion.LoadBalancer != nil && b.Cluster.Spec.Topology.Bastion.LoadBalancer.AdditionalSecurityGroups != nil {
			for _, id := range b.Cluster.Spec.Topology.Bastion.LoadBalancer.AdditionalSecurityGroups {
				t := &awstasks.SecurityGroup{
					Name:      fi.String(id),
					Lifecycle: b.SecurityLifecycle,
					ID:        fi.String(id),
					Shared:    fi.Bool(true),
				}
				if err := c.EnsureTask(t); err != nil {
					return err
				}
				elb.SecurityGroups = append(elb.SecurityGroups, t)
			}
		}

		c.AddTask(elb)
	}

	// When Spotinst Elastigroups are used, there is no need to create
	// a separate task for the attachment of the load balancer since this
	// is already done as part of the Elastigroup's creation, if needed.
	if !featureflag.Spotinst.Enabled() {
		for _, ig := range bastionInstanceGroups {
			// We build the ASG when we iterate over the instance groups

			// Attach the ELB to the ASG
			t := &awstasks.LoadBalancerAttachment{
				Name:      s("bastion-elb-attachment"),
				Lifecycle: b.Lifecycle,

				LoadBalancer:     elb,
				AutoscalingGroup: b.LinkToAutoscalingGroup(ig),
			}
			c.AddTask(t)
		}
	}

	bastionPublicName := ""
	if b.Cluster.Spec.Topology != nil && b.Cluster.Spec.Topology.Bastion != nil {
		bastionPublicName = b.Cluster.Spec.Topology.Bastion.BastionPublicName
	}
	if bastionPublicName != "" {
		// Here we implement the bastion CNAME logic
		// By default bastions will create a CNAME that follows the `bastion-$clustername` formula
		t := &awstasks.DNSName{
			Name:      s(bastionPublicName),
			Lifecycle: b.Lifecycle,

			Zone:               b.LinkToDNSZone(),
			ResourceType:       s("A"),
			TargetLoadBalancer: elb,
		}
		c.AddTask(t)

	}
	return nil
}
