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

package gcemodel

import (
	"fmt"
	"strconv"

	"golang.org/x/exp/slices"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/pkg/wellknownservices"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
)

// APILoadBalancerBuilder builds a LoadBalancer for accessing the API
type APILoadBalancerBuilder struct {
	*GCEModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.CloudupModelBuilder = &APILoadBalancerBuilder{}

// createPublicLB validates the existence of a target pool with the given name,
// and creates an IP address and forwarding rule pointing to that target pool.
func (b *APILoadBalancerBuilder) createPublicLB(c *fi.CloudupModelBuilderContext) error {
	region := ""
	if b.UseRegionalLoadBalancer() {
		for _, subnet := range b.Cluster.Spec.Networking.Subnets {
			if subnet.Region != "" {
				region = subnet.Region
			}
		}
	}

	hc := &gcetasks.HealthCheck{
		Name:        s(b.NameForHealthCheck("api")),
		Port:        i64(wellknownports.KubeAPIServerHealthCheck),
		RequestPath: s("/healthz"),
		Lifecycle:   b.Lifecycle,
		Region:      region,
	}
	c.AddTask(hc)
	var igms []*gcetasks.InstanceGroupManager
	for _, ig := range b.InstanceGroups {
		if ig.Spec.Role != kops.InstanceGroupRoleControlPlane {
			continue
		}
		if len(ig.Spec.Zones) > 1 {
			return fmt.Errorf("instance group %q has %d zones, which is not yet supported for GCP", ig.GetName(), len(ig.Spec.Zones))
		}
		if len(ig.Spec.Zones) == 0 {
			return fmt.Errorf("instance group %q must specify exactly one zone", ig.GetName())
		}
		zone := ig.Spec.Zones[0]
		igms = append(igms, &gcetasks.InstanceGroupManager{Name: s(gce.NameForInstanceGroupManager(b.Cluster.ObjectMeta.Name, ig.ObjectMeta.Name, zone)), Zone: s(zone)})
	}

	bs := &gcetasks.BackendService{
		Name:                  s(b.NameForBackendService("api")),
		Protocol:              s("TCP"),
		HealthChecks:          []*gcetasks.HealthCheck{hc},
		Lifecycle:             b.Lifecycle,
		LoadBalancingScheme:   s("EXTERNAL"),
		Region:                region,
		InstanceGroupManagers: igms,
	}
	c.AddTask(bs)

	ipAddress := &gcetasks.Address{
		Name: s(b.NameForIPAddress("api")),

		Lifecycle:         b.Lifecycle,
		Region:            region,
		WellKnownServices: []wellknownservices.WellKnownService{wellknownservices.KubeAPIServer},
	}
	c.AddTask(ipAddress)

	clusterLabel := gce.LabelForCluster(b.ClusterName())

	c.AddTask(&gcetasks.ForwardingRule{
		Name:                s(b.NameForForwardingRule("api")),
		Lifecycle:           b.Lifecycle,
		PortRange:           s(strconv.Itoa(wellknownports.KubeAPIServer) + "-" + strconv.Itoa(wellknownports.KubeAPIServer)),
		BackendService:      bs,
		IPAddress:           ipAddress,
		Region:              region,
		IPProtocol:          "TCP",
		LoadBalancingScheme: s("EXTERNAL"),
		Labels: map[string]string{
			clusterLabel.Key: clusterLabel.Value,
			"name":           "api",
		},
	})
	if b.Cluster.UsesNoneDNS() {
		ipAddress.WellKnownServices = append(ipAddress.WellKnownServices, wellknownservices.KopsController)

		c.AddTask(&gcetasks.ForwardingRule{
			Name:                s(b.NameForForwardingRule("kops-controller")),
			Lifecycle:           b.Lifecycle,
			PortRange:           s(strconv.Itoa(wellknownports.KopsControllerPort) + "-" + strconv.Itoa(wellknownports.KopsControllerPort)),
			BackendService:      bs,
			IPAddress:           ipAddress,
			IPProtocol:          "TCP",
			Region:              region,
			LoadBalancingScheme: s("EXTERNAL"),
			Labels: map[string]string{
				clusterLabel.Key: clusterLabel.Value,
				"name":           "kops-controller",
			},
		})
	}

	return b.addFirewallRules(c)
}

func (b *APILoadBalancerBuilder) addFirewallRules(c *fi.CloudupModelBuilderContext) error {
	// Allow traffic into the API from KubernetesAPIAccess CIDRs
	{
		network, err := b.LinkToNetwork()
		if err != nil {
			return err
		}
		b.AddFirewallRulesTasks(c, "https-api", &gcetasks.FirewallRule{
			Lifecycle:    b.Lifecycle,
			Network:      network,
			SourceRanges: b.Cluster.Spec.API.Access,
			TargetTags:   []string{b.GCETagForRole(kops.InstanceGroupRoleControlPlane)},
			Allowed:      []string{"tcp:" + strconv.Itoa(wellknownports.KubeAPIServer)},
		})

		if b.NetworkingIsIPAlias() {
			c.AddTask(&gcetasks.FirewallRule{
				Name:         s(b.NameForFirewallRule("pod-cidrs-to-https-api")),
				Lifecycle:    b.Lifecycle,
				Network:      network,
				Family:       gcetasks.AddressFamilyIPv4, // ip alias is always ipv4
				SourceRanges: []string{b.Cluster.Spec.Networking.PodCIDR},
				TargetTags:   []string{b.GCETagForRole(kops.InstanceGroupRoleControlPlane)},
				Allowed:      []string{"tcp:" + strconv.Itoa(wellknownports.KubeAPIServer)},
			})
		}

		if b.Cluster.UsesNoneDNS() {
			b.AddFirewallRulesTasks(c, "kops-controller", &gcetasks.FirewallRule{
				Lifecycle:    b.Lifecycle,
				Network:      network,
				SourceRanges: b.Cluster.Spec.API.Access,
				TargetTags:   []string{b.GCETagForRole(kops.InstanceGroupRoleControlPlane)},
				Allowed:      []string{"tcp:" + strconv.Itoa(wellknownports.KopsControllerPort)},
			})
		}
	}
	return nil

}

// createInternalLB creates an internal load balancer for the cluster.  In
// GCP this entails creating a health check, backend service, and one forwarding rule
// per specified subnet pointing to that backend service.
func (b *APILoadBalancerBuilder) createInternalLB(c *fi.CloudupModelBuilderContext) error {
	clusterLabel := gce.LabelForCluster(b.ClusterName())

	// internal Loadbalancers are always regional
	region := ""
	for _, subnet := range b.Cluster.Spec.Networking.Subnets {
		if subnet.Region != "" {
			region = subnet.Region
		}
	}

	hc := &gcetasks.HealthCheck{
		Name:        s(b.NameForHealthCheck("api")),
		Port:        i64(wellknownports.KubeAPIServer),
		RequestPath: s("/healthz"),
		Region:      region,
		Lifecycle:   b.Lifecycle,
	}
	c.AddTask(hc)
	var igms []*gcetasks.InstanceGroupManager
	for _, ig := range b.InstanceGroups {
		if ig.Spec.Role != kops.InstanceGroupRoleControlPlane {
			continue
		}
		if len(ig.Spec.Zones) > 1 {
			return fmt.Errorf("instance group %q has %d zones, which is not yet supported for GCP", ig.GetName(), len(ig.Spec.Zones))
		}
		if len(ig.Spec.Zones) == 0 {
			return fmt.Errorf("instance group %q must specify exactly one zone", ig.GetName())
		}
		zone := ig.Spec.Zones[0]
		igms = append(igms, &gcetasks.InstanceGroupManager{Name: s(gce.NameForInstanceGroupManager(b.Cluster.ObjectMeta.Name, ig.ObjectMeta.Name, zone)), Zone: s(zone)})
	}
	bs := &gcetasks.BackendService{
		Name:                  s(b.NameForBackendService("api")),
		Protocol:              s("TCP"),
		HealthChecks:          []*gcetasks.HealthCheck{hc},
		Lifecycle:             b.Lifecycle,
		LoadBalancingScheme:   s("INTERNAL"),
		Region:                region,
		InstanceGroupManagers: igms,
	}
	c.AddTask(bs)

	network, err := b.LinkToNetwork()
	if err != nil {
		return err
	}

	for _, sn := range b.Cluster.Spec.Networking.Subnets {
		var subnet *gcetasks.Subnet
		for _, ig := range b.InstanceGroups {
			if ig.HasAPIServer() && slices.Contains(ig.Spec.Subnets, sn.Name) {
				subnet = b.LinkToSubnet(&sn)
				break
			}
		}
		if subnet == nil {
			continue
		}

		ipAddress := &gcetasks.Address{
			Name:          s(b.NameForIPAddress("api-" + sn.Name)),
			IPAddressType: s("INTERNAL"),
			Purpose:       s("SHARED_LOADBALANCER_VIP"),
			Subnetwork:    subnet,

			WellKnownServices: []wellknownservices.WellKnownService{wellknownservices.KubeAPIServer},
			Region:            region,
			Lifecycle:         b.Lifecycle,
		}
		c.AddTask(ipAddress)

		c.AddTask(&gcetasks.ForwardingRule{
			Name:                s(b.NameForForwardingRule("api-" + sn.Name)),
			Lifecycle:           b.Lifecycle,
			BackendService:      bs,
			Ports:               []string{strconv.Itoa(wellknownports.KubeAPIServer)},
			IPAddress:           ipAddress,
			IPProtocol:          "TCP",
			LoadBalancingScheme: s("INTERNAL"),
			Network:             network,
			Region:              region,
			Subnetwork:          subnet,
			Labels: map[string]string{
				clusterLabel.Key: clusterLabel.Value,
				"name":           "api-" + sn.Name,
			},
		})
		if b.Cluster.UsesNoneDNS() {
			ipAddress.WellKnownServices = append(ipAddress.WellKnownServices, wellknownservices.KopsController)

			c.AddTask(&gcetasks.ForwardingRule{
				Name:                s(b.NameForForwardingRule("kops-controller-" + sn.Name)),
				Lifecycle:           b.Lifecycle,
				BackendService:      bs,
				Ports:               []string{strconv.Itoa(wellknownports.KopsControllerPort)},
				IPAddress:           ipAddress,
				IPProtocol:          "TCP",
				LoadBalancingScheme: s("INTERNAL"),
				Network:             network,
				Region:              region,
				Subnetwork:          subnet,
				Labels: map[string]string{
					clusterLabel.Key: clusterLabel.Value,
					"name":           "kops-controller-" + sn.Name,
				},
			})
		}
	}
	return b.addFirewallRules(c)
}

func (b *APILoadBalancerBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	if !b.UseLoadBalancerForAPI() {
		return nil
	}

	lbSpec := b.Cluster.Spec.API.LoadBalancer
	if lbSpec == nil {
		// Skipping API LB creation; not requested in Spec
		return nil
	}

	switch lbSpec.Type {
	case kops.LoadBalancerTypePublic:
		return b.createPublicLB(c)

	case kops.LoadBalancerTypeInternal:
		return b.createInternalLB(c)

	default:
		return fmt.Errorf("unhandled LoadBalancer type %q", lbSpec.Type)
	}
}

// subnetNotSpecified returns true if the given LB subnet is not listed in the list of cluster subnets.
func subnetNotSpecified(sn kops.LoadBalancerSubnetSpec, subnets []kops.ClusterSubnetSpec) bool {
	for _, csn := range subnets {
		if csn.Name == sn.Name || csn.ID == sn.Name {
			return false
		}
	}
	return true
}
