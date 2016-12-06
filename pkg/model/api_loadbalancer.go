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
	// Configuration where an ELB fronts the master (apiservers in particular)

	if !b.UseLoadBalancerForAPI() {
		return nil
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
			case kops.SubnetTypePublic:
				if !b.Cluster.IsTopologyPublic() {
					continue
				}
			case kops.SubnetTypeUtility:
				if !b.Cluster.IsTopologyPrivate() {
					continue
				}

			case kops.SubnetTypePrivate:
				continue

			default:
				return fmt.Errorf("subnet %q had unknown type %q", subnet.SubnetName, subnet.Type)
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

		c.AddTask(elb)
	}

	// Create security group for API ELB
	{
		t := &awstasks.SecurityGroup{
			Name:             s(b.ELBSecurityGroupName("api")),
			VPC:              b.LinkToVPC(),
			Description:      s("Security group for ELB in front of API"),
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

	// Allow HTTPS to the master instances from the ELB
	{
		for _, cidr := range b.Cluster.Spec.APIAccess {
			t := &awstasks.SecurityGroupRule{
				Name:          s("https-api-elb-" + cidr),
				SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
				SourceGroup:   b.LinkToELBSecurityGroup("api"),
				CIDR:          s(cidr),
				FromPort:      i64(443),
				ToPort:        i64(443),
				Protocol:      s("tcp"),
			}
			c.AddTask(t)
		}
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
