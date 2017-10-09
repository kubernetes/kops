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

package awsmodel

import (
	"fmt"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
)

// APILoadBalancerSecurityGroupBuilder builds a LoadBalancer for accessing the API
type APILoadBalancerSecurityGroupBuilder struct {
	*AWSModelContext
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &APILoadBalancerSecurityGroupBuilder{}

func (b *APILoadBalancerSecurityGroupBuilder) Build(c *fi.ModelBuilderContext) error {
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

	// Create security group for API ELB
	{
		t := &awstasks.SecurityGroup{
			Name:      s(b.ELBSecurityGroupName("api")),
			Lifecycle: b.Lifecycle,

			VPC:              b.LinkToVPC(),
			Description:      s("Security group for api ELB"),
			RemoveExtraRules: []string{"port=443"},
		}
		c.AddTask(t)
	}

	// Allow traffic from ELB to egress freely
	{
		t := &awstasks.SecurityGroupRule{
			Name:      s("api-elb-egress"),
			Lifecycle: b.Lifecycle,

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
				Name:      s("https-api-elb-" + cidr),
				Lifecycle: b.Lifecycle,

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
			Name:      s("https-elb-to-master"),
			Lifecycle: b.Lifecycle,

			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
			SourceGroup:   b.LinkToELBSecurityGroup("api"),
			FromPort:      i64(443),
			ToPort:        i64(443),
			Protocol:      s("tcp"),
		}
		c.AddTask(t)
	}

	return nil

}
