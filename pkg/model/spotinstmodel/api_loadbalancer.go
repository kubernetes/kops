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

package spotinstmodel

import (
	"fmt"
	"time"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/model/awsmodel"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/fitasks"
)

// APILoadBalancerBuilder builds a LoadBalancer for accessing the API
type APILoadBalancerBuilder struct {
	*SpotinstModelContext

	Lifecycle         *fi.Lifecycle
	SecurityLifecycle *fi.Lifecycle
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

	// Compute the subnets - only one per zone, and then break ties based on ChooseBestSubnetForELB
	var elbSubnets []*awstasks.Subnet
	{
		subnetsByZone := make(map[string][]*kops.ClusterSubnetSpec)
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

			subnetsByZone[subnet.Zone] = append(subnetsByZone[subnet.Zone], subnet)
		}

		builder := &awsmodel.APILoadBalancerBuilder{
			AWSModelContext: &awsmodel.AWSModelContext{
				KopsModelContext: b.KopsModelContext,
			},
		}
		for zone, subnets := range subnetsByZone {
			subnet := builder.ChooseBestSubnetForELB(zone, subnets)
			elbSubnets = append(elbSubnets, b.LinkToSubnet(subnet))
		}
	}

	var elb *awstasks.LoadBalancer
	{
		loadBalancerName := b.GetELBName32("api")

		idleTimeout := awsmodel.LoadBalancerDefaultIdleTimeout
		if lbSpec.IdleTimeoutSeconds != nil {
			idleTimeout = time.Second * time.Duration(*lbSpec.IdleTimeoutSeconds)
		}

		elb = &awstasks.LoadBalancer{
			Name:      fi.String("api." + b.ClusterName()),
			Lifecycle: b.Lifecycle,

			LoadBalancerName: fi.String(loadBalancerName),
			SecurityGroups: []*awstasks.SecurityGroup{
				b.LinkToELBSecurityGroup("api"),
			},
			Subnets: elbSubnets,
			Listeners: map[string]*awstasks.LoadBalancerListener{
				"443": {InstancePort: 443},
			},

			// Configure fast-recovery health-checks
			HealthCheck: &awstasks.LoadBalancerHealthCheck{
				Target:             fi.String("SSL:443"),
				Timeout:            fi.Int64(5),
				Interval:           fi.Int64(10),
				HealthyThreshold:   fi.Int64(2),
				UnhealthyThreshold: fi.Int64(2),
			},

			ConnectionSettings: &awstasks.LoadBalancerConnectionSettings{
				IdleTimeout: fi.Int64(int64(idleTimeout.Seconds())),
			},
		}

		switch lbSpec.Type {
		case kops.LoadBalancerTypeInternal:
			elb.Scheme = fi.String("internal")
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
			Name:      fi.String(b.ELBSecurityGroupName("api")),
			Lifecycle: b.SecurityLifecycle,

			VPC:              b.LinkToVPC(),
			Description:      fi.String("Security group for api ELB"),
			RemoveExtraRules: []string{"port=443"},
		}
		t.Tags = b.CloudTags(*t.Name, false)
		c.AddTask(t)
	}

	// Allow traffic from ELB to egress freely
	{
		t := &awstasks.SecurityGroupRule{
			Name:      fi.String("api-elb-egress"),
			Lifecycle: b.SecurityLifecycle,

			SecurityGroup: b.LinkToELBSecurityGroup("api"),
			Egress:        fi.Bool(true),
			CIDR:          fi.String("0.0.0.0/0"),
		}
		c.AddTask(t)
	}

	// Allow traffic into the ELB from KubernetesAPIAccess CIDRs
	{
		for _, cidr := range b.Cluster.Spec.KubernetesAPIAccess {
			t := &awstasks.SecurityGroupRule{
				Name:      fi.String("https-api-elb-" + cidr),
				Lifecycle: b.SecurityLifecycle,

				SecurityGroup: b.LinkToELBSecurityGroup("api"),
				CIDR:          fi.String(cidr),
				FromPort:      fi.Int64(443),
				ToPort:        fi.Int64(443),
				Protocol:      fi.String("tcp"),
			}
			c.AddTask(t)
		}
	}

	// Add precreated additional security groups to the ELB
	{
		for _, id := range b.Cluster.Spec.API.LoadBalancer.AdditionalSecurityGroups {
			t := &awstasks.SecurityGroup{
				Name:   fi.String(id),
				ID:     fi.String(id),
				Shared: fi.Bool(true),

				Lifecycle: b.SecurityLifecycle,
			}
			if err := c.EnsureTask(t); err != nil {
				return err
			}
			elb.SecurityGroups = append(elb.SecurityGroups, t)
		}
	}

	// Allow HTTPS to the master instances from the ELB
	{
		t := &awstasks.SecurityGroupRule{
			Name:      fi.String("https-elb-to-master"),
			Lifecycle: b.SecurityLifecycle,

			SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
			SourceGroup:   b.LinkToELBSecurityGroup("api"),
			FromPort:      fi.Int64(443),
			ToPort:        fi.Int64(443),
			Protocol:      fi.String("tcp"),
		}
		c.AddTask(t)
	}

	if dns.IsGossipHostname(b.Cluster.Name) || b.UsePrivateDNS() {
		// Ensure the ELB hostname is included in the TLS certificate,
		// if we're not going to use an alias for it
		// TODO: I don't love this technique for finding the task by name & modifying it
		masterKeypairTask, found := c.Tasks["Keypair/master"]
		if !found {
			return fmt.Errorf("keypair/master task not found")
		}
		masterKeypair := masterKeypairTask.(*fitasks.Keypair)
		masterKeypair.AlternateNameTasks = append(masterKeypair.AlternateNameTasks, elb)
	}

	return nil
}
