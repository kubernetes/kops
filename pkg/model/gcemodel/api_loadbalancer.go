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
	"slices"
	"strconv"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/model"
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

// apiHealthCheck returns the health check shared by the API load balancers.
// It requests /readyz over HTTPS rather than opening a TCP connection: kube-apiserver accepts
// connections while it is still starting up, has lost etcd, or is draining before shutdown, and
// /readyz reports all of those. GCE does not verify the serving certificate, and kube-apiserver
// allows this path without credentials.
func (b *APILoadBalancerBuilder) apiHealthCheck() *gcetasks.HealthCheck {
	return &gcetasks.HealthCheck{
		Name:        s(b.NameForHealthCheck("api-https")),
		Port:        wellknownports.KubeAPIServer,
		Protocol:    gcetasks.HealthCheckProtocolHTTPS,
		RequestPath: s("/readyz"),
		Lifecycle:   b.Lifecycle,
	}
}

// linkToInstanceGroupManager returns a reference to the instance group manager of an instance group.
func (b *APILoadBalancerBuilder) linkToInstanceGroupManager(ig *kops.InstanceGroup) (*gcetasks.InstanceGroupManager, error) {
	if len(ig.Spec.Zones) > 1 {
		return nil, fmt.Errorf("instance group %q has %d zones, which is not yet supported for GCP", ig.GetName(), len(ig.Spec.Zones))
	}
	if len(ig.Spec.Zones) == 0 {
		return nil, fmt.Errorf("instance group %q must specify exactly one zone", ig.GetName())
	}
	zone := ig.Spec.Zones[0]
	return &gcetasks.InstanceGroupManager{Name: s(gce.NameForInstanceGroupManager(b.Cluster.ObjectMeta.Name, ig.ObjectMeta.Name, zone)), Zone: s(zone)}, nil
}

// createPublicLB creates an external passthrough load balancer for the API: a backend service
// pointing at the instance groups that serve the API, an IP address and a forwarding rule.
func (b *APILoadBalancerBuilder) createPublicLB(c *fi.CloudupModelBuilderContext, healthCheck *gcetasks.HealthCheck) error {
	// The API server only instance groups front the public endpoint when the cluster has any;
	// otherwise the control plane instance groups do.
	clusterHasAPIServerOnly := b.HasAPIServerOnlyInstanceGroups()
	var igms []*gcetasks.InstanceGroupManager
	for _, ig := range b.InstanceGroups {
		if !ig.IsAPIServerOnly() && (clusterHasAPIServerOnly || !ig.IsControlPlane()) {
			continue
		}
		igm, err := b.linkToInstanceGroupManager(ig)
		if err != nil {
			return err
		}
		igms = append(igms, igm)
	}

	backendService := &gcetasks.BackendService{
		Name:                  s(b.NameForBackendService("api-public")),
		Protocol:              s("TCP"),
		HealthChecks:          []*gcetasks.HealthCheck{healthCheck},
		Lifecycle:             b.Lifecycle,
		LoadBalancingScheme:   s("EXTERNAL"),
		InstanceGroupManagers: igms,
	}
	c.AddTask(backendService)

	ipAddress := &gcetasks.Address{
		Name: s(b.NameForIPAddress("api")),

		Lifecycle:         b.Lifecycle,
		WellKnownServices: []wellknownservices.WellKnownService{wellknownservices.KubeAPIServer},
	}
	c.AddTask(ipAddress)

	clusterLabel := gce.LabelForCluster(b.ClusterName())

	forwardingRule := &gcetasks.ForwardingRule{
		Name:                s(b.NameForForwardingRule("api")),
		Lifecycle:           b.Lifecycle,
		PortRange:           s(strconv.Itoa(wellknownports.KubeAPIServer) + "-" + strconv.Itoa(wellknownports.KubeAPIServer)),
		BackendService:      backendService,
		IPAddress:           ipAddress,
		IPProtocol:          "TCP",
		LoadBalancingScheme: s("EXTERNAL"),
		Labels: map[string]string{
			clusterLabel.Key: clusterLabel.Value,
			"name":           "api",
		},
	}
	// Clusters created before kOps 1.38 used a target pool, whose legacy HTTP health check
	// went through the (since removed) kube-apiserver-healthcheck sidecar.
	forwardingRule.PruneTargetPoolWithName(b.NameForTargetPool("api"))
	c.AddTask(forwardingRule)

	return nil
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
			TargetTags:   b.GCETagsForAPIServerTargets(),
			Allowed:      []string{"tcp:" + strconv.Itoa(wellknownports.KubeAPIServer)},
		})

		if b.NetworkingIsIPAlias() {
			c.AddTask(&gcetasks.FirewallRule{
				Name:         s(b.NameForFirewallRule("pod-cidrs-to-https-api")),
				Lifecycle:    b.Lifecycle,
				Network:      network,
				Family:       gcetasks.AddressFamilyIPv4, // ip alias is always ipv4
				SourceRanges: []string{b.Cluster.Spec.Networking.PodCIDR},
				TargetTags:   b.GCETagsForAPIServerTargets(),
				Allowed:      []string{"tcp:" + strconv.Itoa(wellknownports.KubeAPIServer)},
			})
		}

		if b.Cluster.UsesLoadBalancerForKopsController() {
			b.AddFirewallRulesTasks(c, "kops-controller", &gcetasks.FirewallRule{
				Lifecycle:    b.Lifecycle,
				Network:      network,
				SourceRanges: b.Cluster.Spec.API.Access,
				TargetTags:   []string{b.GCETagForRole(kops.InstanceGroupRoleControlPlane)},
				Allowed:      []string{"tcp:" + strconv.Itoa(wellknownports.KopsControllerPort)},
			})
		}

		if model.UseCiliumEtcd(b.Cluster) {
			b.AddFirewallRulesTasks(c, "cilium-etcd", &gcetasks.FirewallRule{
				Lifecycle:    b.Lifecycle,
				Network:      network,
				SourceRanges: b.Cluster.Spec.API.Access,
				TargetTags:   []string{b.GCETagForRole(kops.InstanceGroupRoleControlPlane)},
				Allowed:      []string{"tcp:" + strconv.Itoa(wellknownports.EtcdCiliumClientPort)},
			})
		}
	}
	return nil

}

// createInternalLB creates an internal load balancer for the cluster.  In
// GCP this entails creating a health check, backend service, and one forwarding rule
// per specified subnet pointing to that backend service.
func (b *APILoadBalancerBuilder) createInternalLB(c *fi.CloudupModelBuilderContext, hc *gcetasks.HealthCheck) error {
	clusterLabel := gce.LabelForCluster(b.ClusterName())

	// Collect ControlPlane and APIServer MIGs separately. The API backend service
	// includes both (both serve the kube-apiserver), while the kops-controller and
	// etcd backend services only include ControlPlane MIGs.
	var apiIGMs []*gcetasks.InstanceGroupManager
	var etcdIGMs []*gcetasks.InstanceGroupManager
	var kopsControllerIGMs []*gcetasks.InstanceGroupManager
	requireEtcdLB := false
	for _, ig := range b.InstanceGroups {
		if !ig.RunsAPIServer() && !ig.RunsEtcd() {
			continue
		}
		if len(ig.Spec.Zones) > 1 {
			return fmt.Errorf("instance group %q has %d zones, which is not yet supported for GCP", ig.GetName(), len(ig.Spec.Zones))
		}
		if len(ig.Spec.Zones) == 0 {
			return fmt.Errorf("instance group %q must specify exactly one zone", ig.GetName())
		}
		zone := ig.Spec.Zones[0]
		igm := &gcetasks.InstanceGroupManager{Name: s(gce.NameForInstanceGroupManager(b.Cluster.ObjectMeta.Name, ig.ObjectMeta.Name, zone)), Zone: s(zone)}
		if ig.RunsAPIServer() {
			apiIGMs = append(apiIGMs, igm)
		}
		if ig.RunsEtcd() {
			etcdIGMs = append(etcdIGMs, igm)
		}
		if ig.IsControlPlane() || ig.RunsAPIServer() /* && Check for no control plane */ {
			kopsControllerIGMs = append(kopsControllerIGMs, igm)
		}
		if ig.IsAPIServerOnly() {
			requireEtcdLB = b.Cluster.UsesNoneDNS()
		}
		if ig.IsEtcdOnly() {
			requireEtcdLB = true
		}
	}
	backendService := &gcetasks.BackendService{
		Name:                  s(b.NameForBackendService("api")),
		Protocol:              s("TCP"),
		HealthChecks:          []*gcetasks.HealthCheck{hc},
		Lifecycle:             b.Lifecycle,
		LoadBalancingScheme:   s("INTERNAL"),
		InstanceGroupManagers: apiIGMs,
	}
	// Clusters created before kOps 1.38 used a TCP health check, which cannot be changed in place.
	backendService.PruneHealthCheckWithName(b.NameForHealthCheck("api"))
	c.AddTask(backendService)

	// kopsControllerBS is a backend service that targets ControlPlane or MIGs.
	kopsControllerBS := backendService
	if b.HasAPIServerOnlyInstanceGroups() || b.HasEtcdOnlyInstanceGroups() {
		controlPlaneHC := &gcetasks.HealthCheck{
			Name:      s(b.NameForHealthCheck("kops-controller")),
			Port:      wellknownports.KopsControllerPort,
			Protocol:  gcetasks.HealthCheckProtocolSSL,
			Lifecycle: b.Lifecycle,
		}
		c.AddTask(controlPlaneHC)
		kopsControllerBS = &gcetasks.BackendService{
			Name:                  s(b.NameForBackendService("kops-controller")),
			Protocol:              s("TCP"),
			HealthChecks:          []*gcetasks.HealthCheck{controlPlaneHC},
			Lifecycle:             b.Lifecycle,
			LoadBalancingScheme:   s("INTERNAL"),
			InstanceGroupManagers: kopsControllerIGMs,
		}
		c.AddTask(kopsControllerBS)
	}

	network, err := b.LinkToNetwork()
	if err != nil {
		return err
	}

	for _, sn := range b.Cluster.Spec.Networking.Subnets {
		var subnet *gcetasks.Subnet
		for _, ig := range b.InstanceGroups {
			if ig.RunsAPIServer() && slices.Contains(ig.Spec.Subnets, sn.Name) {
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
			Lifecycle:         b.Lifecycle,
		}
		c.AddTask(ipAddress)

		c.AddTask(&gcetasks.ForwardingRule{
			Name:                s(b.NameForForwardingRule("api-" + sn.Name)),
			Lifecycle:           b.Lifecycle,
			BackendService:      backendService,
			Ports:               []string{strconv.Itoa(wellknownports.KubeAPIServer)},
			IPAddress:           ipAddress,
			IPProtocol:          "TCP",
			LoadBalancingScheme: s("INTERNAL"),
			Network:             network,
			Subnetwork:          subnet,
			Labels: map[string]string{
				clusterLabel.Key: clusterLabel.Value,
				"name":           "api-" + sn.Name,
			},
		})
		if b.Cluster.UsesLoadBalancerForKopsController() {
			ipAddress.WellKnownServices = append(ipAddress.WellKnownServices, wellknownservices.KopsController)

			fr := &gcetasks.ForwardingRule{
				Name:                s(b.NameForForwardingRule("kops-controller-" + sn.Name)),
				Lifecycle:           b.Lifecycle,
				BackendService:      kopsControllerBS,
				Ports:               []string{strconv.Itoa(wellknownports.KopsControllerPort)},
				IPAddress:           ipAddress,
				IPProtocol:          "TCP",
				LoadBalancingScheme: s("INTERNAL"),
				Network:             network,
				Subnetwork:          subnet,
				Labels: map[string]string{
					clusterLabel.Key: clusterLabel.Value,
					"name":           "kops-controller-" + sn.Name,
				},
			}
			// We previously created a forwarding rule which was external; prune it
			fr.PruneForwardingRulesWithName(b.NameForForwardingRule("kops-controller")) // , "Removing legacy external load balancer for kops-controller")

			c.AddTask(fr)
		}

		if model.UseCiliumEtcd(b.Cluster) {
			etcdBS := kopsControllerBS
			if b.HasEtcdOnlyInstanceGroups() {
				etcdBS = &gcetasks.BackendService{
					Name:                  s(b.NameForBackendService("cilium-etcd")),
					Protocol:              s("TCP"),
					HealthChecks:          []*gcetasks.HealthCheck{hc},
					Lifecycle:             b.Lifecycle,
					LoadBalancingScheme:   s("INTERNAL"),
					InstanceGroupManagers: etcdIGMs,
				}
				c.AddTask(etcdBS)
			}
			c.AddTask(&gcetasks.ForwardingRule{
				Name:                s(b.NameForForwardingRule("cilium-etcd-" + sn.Name)),
				Lifecycle:           b.Lifecycle,
				BackendService:      etcdBS,
				Ports:               []string{strconv.Itoa(wellknownports.EtcdCiliumClientPort)},
				IPAddress:           ipAddress,
				IPProtocol:          "TCP",
				LoadBalancingScheme: s("INTERNAL"),
				Network:             network,
				Subnetwork:          subnet,
				Labels: map[string]string{
					clusterLabel.Key: clusterLabel.Value,
					"name":           "cilium-etcd-" + sn.Name,
				},
			})
		}
	}

	if requireEtcdLB {
		if err := b.createEtcdInternalLB(c, etcdIGMs); err != nil {
			return err
		}
	}

	return nil
}

func (b *APILoadBalancerBuilder) createEtcdInternalLB(c *fi.CloudupModelBuilderContext, etcdIGMs []*gcetasks.InstanceGroupManager) error {
	clusterLabel := gce.LabelForCluster(b.ClusterName())
	main_hc := &gcetasks.HealthCheck{
		Name:      s(b.NameForHealthCheck("etcd-main")),
		Port:      wellknownports.EtcdMainClientPort,
		Protocol:  gcetasks.HealthCheckProtocolTCP,
		Lifecycle: b.Lifecycle,
	}
	c.AddTask(main_hc)
	// "Value for field 'resource.healthChecks' is too large: maximum size 1 element(s); actual size 2."
	// Skipping event health check till this is supported.
	/*
	   event_hc := &gcetasks.HealthCheck{
	           Name:      s(b.NameForHealthCheck("etcd-event")),
	           Port:      wellknownports.EtcdEventsClientPort,
	           Protocol:  gcetasks.HealthCheckProtocolTCP,
	           Lifecycle: b.Lifecycle,
	   }
	   c.AddTask(event_hc)
	*/
	bs := &gcetasks.BackendService{
		Name:                  s(b.NameForBackendService("etcd")),
		Protocol:              s("TCP"),
		HealthChecks:          []*gcetasks.HealthCheck{main_hc /*event_hc,*/},
		Lifecycle:             b.Lifecycle,
		LoadBalancingScheme:   s("INTERNAL"),
		InstanceGroupManagers: etcdIGMs,
	}
	c.AddTask(bs)
	network, err := b.LinkToNetwork()
	if err != nil {
		return err
	}
	for _, sn := range b.Cluster.Spec.Networking.Subnets {
		var subnet *gcetasks.Subnet
		for _, ig := range b.InstanceGroups {
			if ig.RunsAPIServer() && slices.Contains(ig.Spec.Subnets, sn.Name) {
				subnet = b.LinkToSubnet(&sn)
				break
			}
		}
		if subnet == nil {
			continue
		}

		ipAddress := &gcetasks.Address{
			Name:          s(b.NameForIPAddress("etcd-" + sn.Name)),
			IPAddressType: s("INTERNAL"),
			Purpose:       s("SHARED_LOADBALANCER_VIP"),
			Subnetwork:    subnet,

			WellKnownServices: []wellknownservices.WellKnownService{wellknownservices.EtcdMain},
			Lifecycle:         b.Lifecycle,
		}
		c.AddTask(ipAddress)
		c.AddTask(&gcetasks.ForwardingRule{
			Name:                s(b.NameForForwardingRule("etcd-" + sn.Name)),
			Lifecycle:           b.Lifecycle,
			BackendService:      bs,
			Ports:               []string{strconv.Itoa(wellknownports.EtcdMainClientPort), strconv.Itoa(wellknownports.EtcdEventsClientPort)},
			IPAddress:           ipAddress,
			IPProtocol:          "TCP",
			LoadBalancingScheme: s("INTERNAL"),
			Network:             network,
			Subnetwork:          subnet,
			Labels: map[string]string{
				clusterLabel.Key: clusterLabel.Value,
				"name":           "etcd-" + sn.Name,
			},
		})
	}
	return nil
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

	healthCheck := b.apiHealthCheck()
	c.AddTask(healthCheck)

	switch lbSpec.Type {
	case kops.LoadBalancerTypePublic:
		if err := b.createPublicLB(c, healthCheck); err != nil {
			return err
		}
		// We always create the internal load balancer also;
		// it allows us to restrict access to only the nodes.
		if err := b.createInternalLB(c, healthCheck); err != nil {
			return err
		}

		return b.addFirewallRules(c)

	case kops.LoadBalancerTypeInternal:
		if err := b.createInternalLB(c, healthCheck); err != nil {
			return err
		}

		return b.addFirewallRules(c)

	default:
		return fmt.Errorf("unhandled LoadBalancer type %q", lbSpec.Type)
	}
}
