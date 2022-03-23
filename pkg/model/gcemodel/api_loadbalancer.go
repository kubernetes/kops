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

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
)

// APILoadBalancerBuilder builds a LoadBalancer for accessing the API
type APILoadBalancerBuilder struct {
	*GCEModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.ModelBuilder = &APILoadBalancerBuilder{}

// createPublicLB validates the existence of a target pool with the given name,
// and creates an IP address and forwarding rule pointing to that target pool.
func createPublicLB(b *APILoadBalancerBuilder, c *fi.ModelBuilderContext) error {
	// TODO: point target pool to instance group managers, as done in internal LB.
	targetPool := &gcetasks.TargetPool{
		Name:      s(b.NameForTargetPool("api")),
		Lifecycle: b.Lifecycle,
	}
	c.AddTask(targetPool)

	healthCheck := &gcetasks.HTTPHealthcheck{
		Name:      s(b.NameForHealthcheck("api")),
		Port:      i64(wellknownports.KubeAPIServerHealthCheck),
		Lifecycle: b.Lifecycle,
	}

	c.AddTask(healthCheck)

	poolHealthCheck := &gcetasks.PoolHealthCheck{
		Name:        s(b.NameForPoolHealthcheck("api")),
		Healthcheck: healthCheck,
		Pool:        targetPool,
		Lifecycle:   b.Lifecycle,
	}
	c.AddTask(poolHealthCheck)

	ipAddress := &gcetasks.Address{
		Name:      s(b.NameForIPAddress("api")),
		Lifecycle: b.Lifecycle,
	}
	c.AddTask(ipAddress)

	forwardingRule := &gcetasks.ForwardingRule{
		Name:       s(b.NameForForwardingRule("api")),
		Lifecycle:  b.Lifecycle,
		PortRange:  s("443-443"),
		TargetPool: targetPool,
		IPAddress:  ipAddress,
		IPProtocol: "TCP",
	}

	c.AddTask(forwardingRule)

	{
		// Ensure the IP address is included in our certificate
		ipAddress.ForAPIServer = true
	}

	// Allow traffic into the API (port 443) from KubernetesAPIAccess CIDRs
	{
		network, err := b.LinkToNetwork()
		if err != nil {
			return err
		}
		b.AddFirewallRulesTasks(c, "https-api", &gcetasks.FirewallRule{
			Lifecycle:    b.Lifecycle,
			Network:      network,
			SourceRanges: b.Cluster.Spec.KubernetesAPIAccess,
			TargetTags:   []string{b.GCETagForRole(kops.InstanceGroupRoleMaster)},
			Allowed:      []string{"tcp:443"},
		})
	}
	return nil

}

// createInternalLB creates an internal load balancer for the cluster.  In
// GCP this entails creating a health check, backend service, and one forwarding rule
// per specified subnet pointing to that backend service.
func createInternalLB(b *APILoadBalancerBuilder, c *fi.ModelBuilderContext) error {
	lbSpec := b.Cluster.Spec.API.LoadBalancer
	hc := &gcetasks.HealthCheck{
		Name:      s(b.NameForHealthCheck("api")),
		Port:      443,
		Lifecycle: b.Lifecycle,
	}
	c.AddTask(hc)
	var igms []*gcetasks.InstanceGroupManager
	for _, ig := range b.InstanceGroups {
		if ig.Spec.Role != kops.InstanceGroupRoleMaster {
			continue
		}
		if len(ig.Spec.Zones) > 1 {
			return fmt.Errorf("Instance group %q has %d zones, which is not yet supported for GCP.", ig.GetName(), len(ig.Spec.Zones))
		}
		if len(ig.Spec.Zones) == 0 {
			return fmt.Errorf("Instance group %q must specify exactly one zone.", ig.GetName())
		}
		zone := ig.Spec.Zones[0]
		igms = append(igms, &gcetasks.InstanceGroupManager{Name: s(gce.NameForInstanceGroupManager(b.Cluster, ig, zone)), Zone: s(zone)})
	}
	bs := &gcetasks.BackendService{
		Name:                  s(b.NameForBackendService("api")),
		Protocol:              s("TCP"),
		HealthChecks:          []*gcetasks.HealthCheck{hc},
		Lifecycle:             b.Lifecycle,
		LoadBalancingScheme:   s("INTERNAL"),
		InstanceGroupManagers: igms,
	}
	c.AddTask(bs)
	for _, sn := range lbSpec.Subnets {
		network, err := b.LinkToNetwork()
		if err != nil {
			return err
		}
		t := true
		subnet := &gcetasks.Subnet{
			Name:    s(sn.Name),
			Network: network,
			Shared:  &t,
			// Override lifecycle because these subnets are specified
			// to already exist.
			Lifecycle: fi.LifecycleExistsAndWarnIfChanges,
		}
		// TODO: automatically associate forwarding rule to subnets if no subnets are specified here.
		if subnetNotSpecified(sn, b.Cluster.Spec.Subnets) {
			c.AddTask(subnet)
		}
		c.AddTask(&gcetasks.ForwardingRule{
			Name:                s(b.NameForForwardingRule(sn.Name)),
			Lifecycle:           b.Lifecycle,
			BackendService:      bs,
			Ports:               []string{"443"},
			RuleIPAddress:       sn.PrivateIPv4Address,
			IPProtocol:          "TCP",
			LoadBalancingScheme: s("INTERNAL"),
			Network:             network,
			Subnetwork:          subnet,
		})
	}

	return nil
}

func (b *APILoadBalancerBuilder) Build(c *fi.ModelBuilderContext) error {
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
		return createPublicLB(b, c)

	case kops.LoadBalancerTypeInternal:
		return createInternalLB(b, c)

	default:
		return fmt.Errorf("unhandled LoadBalancer type %q", lbSpec.Type)
	}
}

// subnetNotSpecified returns true if the given LB subnet is not listed in the list of cluster subnets.
func subnetNotSpecified(sn kops.LoadBalancerSubnetSpec, subnets []kops.ClusterSubnetSpec) bool {
	for _, csn := range subnets {
		if csn.Name == sn.Name || csn.ProviderID == sn.Name {
			return false
		}
	}
	return true
}
