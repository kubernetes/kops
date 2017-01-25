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
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
)

// APILoadBalancerBuilder builds a LoadBalancer for accessing the API
type APILoadBalancerBuilder struct {
	*KopsModelContext
}

var _ fi.ModelBuilder = &APILoadBalancerBuilder{}

func (b *APILoadBalancerBuilder) Build(c *fi.ModelBuilderContext) error {
	// Configuration where an ELB fronts the API
	if !b.UseLoadBalancerForAPI() {
		return nil
	}

	lbSpec := b.Cluster.Spec.API.LoadBalancer
	if lbSpec == nil {
		// Skipping API ELB creation; not requested in Spec
		return nil
	}

	switch lbSpec.Type {
	case kops.LoadBalancerTypeInternal, kops.LoadBalancerTypePublic:
	// OK

	default:
		return fmt.Errorf("unhandled LoadBalancer type %q", lbSpec.Type)
	}

	var elb *awstasks.LoadBalancer
	{
		elbID, err := b.GetELBName32("api")
		if err != nil {
			return err
		}

		var elbSubnets []*awstasks.Subnet
		for i := range b.Cluster.Spec.Subnets {
			subnet := &b.Cluster.Spec.Subnets[i]

			switch subnet.Type {
			case kops.SubnetTypePublic, kops.SubnetTypeUtility:
				if lbSpec.Type != kops.LoadBalancerTypePublic {
					continue
				}

			case kops.SubnetTypePrivate:
				if lbSpec.Type != kops.LoadBalancerTypeInternal {
					continue
				}

			default:
				return fmt.Errorf("subnet %q had unknown type %q", subnet.Name, subnet.Type)
			}

			elbSubnets = append(elbSubnets, b.LinkToSubnet(subnet))
		}

		elb = &awstasks.LoadBalancer{
			Name: s("api." + b.ClusterName()),
			ID:   s(elbID),
			SecurityGroups: []*awstasks.SecurityGroup{
				b.LinkToELBSecurityGroup("api"),
			},
			Subnets: elbSubnets,
			Listeners: map[string]*awstasks.LoadBalancerListener{
				"443": {InstancePort: 443},
			},

			// Configure fast-recovery health-checks
			HealthCheck: &awstasks.LoadBalancerHealthCheck{
				Target:             s("TCP:443"),
				Timeout:            i64(5),
				Interval:           i64(10),
				HealthyThreshold:   i64(2),
				UnhealthyThreshold: i64(2),
			},
		}

		switch lbSpec.Type {
		case kops.LoadBalancerTypeInternal:
			elb.Scheme = s("internal")
		case kops.LoadBalancerTypePublic:
			elb.Scheme = nil
		default:
			return fmt.Errorf("unknown elb Type: %q", lbSpec.Type)
		}

		c.AddTask(elb)
	}

	// Create security group for API ELB
	{
		t := &awstasks.SecurityGroup{
			Name:             s(b.ELBSecurityGroupName("api")),
			VPC:              b.LinkToVPC(),
			Description:      s("Security group for api ELB"),
			RemoveExtraRules: []string{"port=443"},
		}
		c.AddTask(t)
	}

	// Allow traffic from ELB to egress freely
	{
		t := &awstasks.SecurityGroupRule{
			Name:          s("api-elb-egress"),
			SecurityGroup: b.LinkToELBSecurityGroup("api"),
			Egress:        fi.Bool(true),
			CIDR:          s("0.0.0.0/0"),
		}
		c.AddTask(t)
	}

	// Allow traffic into the ELB from KubernetesAPIAccess CIDRs
	{
		for _, cidr := range b.Cluster.Spec.KubernetesAPIAccess {
			t := &awstasks.SecurityGroupRule{
				Name:          s("https-api-elb-" + cidr),
				SecurityGroup: b.LinkToELBSecurityGroup("api"),
				CIDR:          s(cidr),
				FromPort:      i64(443),
				ToPort:        i64(443),
				Protocol:      s("tcp"),
			}
			c.AddTask(t)
		}
	}

	// Allow HTTPS to the master instances from the ELB
	{
		t := &awstasks.SecurityGroupRule{
			Name:          s("https-elb-to-master"),
			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
			SourceGroup:   b.LinkToELBSecurityGroup("api"),
			FromPort:      i64(443),
			ToPort:        i64(443),
			Protocol:      s("tcp"),
		}
		c.AddTask(t)
	}

	for _, ig := range b.MasterInstanceGroups() {
		t := &awstasks.LoadBalancerAttachment{
			Name: s("api-" + ig.ObjectMeta.Name),

			LoadBalancer:     b.LinkToELB("api"),
			AutoscalingGroup: b.LinkToAutoscalingGroup(ig),
		}

		c.AddTask(t)
	}

	return nil

}
