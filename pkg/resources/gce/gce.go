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

package gce

import (
	"context"
	"fmt"
	"strings"

	compute "google.golang.org/api/compute/v1"
	clouddns "google.golang.org/api/dns/v1"
	"google.golang.org/api/iam/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/pkg/truncate"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

type gceListFn func() ([]*resources.Resource, error)

const (
	typeInstance             = "Instance"
	typeInstanceTemplate     = "InstanceTemplate"
	typeDisk                 = "Disk"
	typeInstanceGroupManager = "InstanceGroupManager"
	typeTargetPool           = "TargetPool"
	typeFirewallRule         = "FirewallRule"
	typeForwardingRule       = "ForwardingRule"
	typeHTTPHealthcheck      = "HTTP HealthCheck"
	typeHealthcheck          = "HealthCheck"
	typeAddress              = "Address"
	typeRoute                = "Route"
	typeNetwork              = "Network"
	typeSubnet               = "Subnet"
	typeRouter               = "Router"
	typeDNSRecord            = "DNSRecord"
	typeServiceAccount       = "ServiceAccount"
	typeBackendService       = "BackendService"
)

// Maximum number of `-` separated tokens in a name
// Example: nodeport-external-to-node-ipv6
const maxPrefixTokens = 5

// Maximum length of a GCE route name
const maxGCERouteNameLength = 63

func ListResourcesGCE(gceCloud gce.GCECloud, clusterName string, region string) (map[string]*resources.Resource, error) {
	ctx := context.TODO()

	if region == "" {
		region = gceCloud.Region()
	}

	resources := make(map[string]*resources.Resource)

	d := &clusterDiscoveryGCE{
		cloud:       gceCloud,
		gceCloud:    gceCloud,
		clusterName: clusterName,
	}

	{
		// TODO: Only zones in api.Cluster object, if we have one?
		gceZones, err := d.gceCloud.Compute().Zones().List(ctx, d.gceCloud.Project())
		if err != nil {
			return nil, fmt.Errorf("error listing zones: %v", err)
		}
		for _, gceZone := range gceZones {
			u, err := gce.ParseGoogleCloudURL(gceZone.Region)
			if err != nil {
				return nil, err
			}
			if u.Name != region {
				continue
			}
			d.zones = append(d.zones, gceZone.Name)
		}
		if len(d.zones) == 0 {
			return nil, fmt.Errorf("unable to determine zones in region %q", region)
		}
		klog.Infof("Scanning zones: %v", d.zones)
	}

	listFunctions := []gceListFn{
		d.listGCEInstanceTemplates,
		d.listInstanceGroupManagersAndInstances,
		d.listTargetPools,
		d.listForwardingRules,
		d.listFirewallRules,
		d.listGCEDisks,
		d.listGCEDNSZone,
		// TODO: Find routes via instances (via instance groups)
		d.listAddresses,
		d.listSubnets,
		d.listRouters,
		d.listNetworks,
		d.listServiceAccounts,
		d.listBackendServices,
		d.listHealthchecks,
	}
	for _, fn := range listFunctions {
		resourceTrackers, err := fn()
		if err != nil {
			return nil, err
		}
		for _, t := range resourceTrackers {
			resources[t.Type+":"+t.ID] = t
		}
	}

	// We try to clean up orphaned routes.
	{
		resourceTrackers, err := d.listRoutes(ctx, resources)
		if err != nil {
			return nil, err
		}
		for _, t := range resourceTrackers {
			resources[t.Type+":"+t.ID] = t
		}
	}

	for k, t := range resources {
		if t.Done {
			delete(resources, k)
		}
	}
	return resources, nil
}

type clusterDiscoveryGCE struct {
	cloud       fi.Cloud
	gceCloud    gce.GCECloud
	clusterName string

	instanceTemplates []*compute.InstanceTemplate
	zones             []string
}

func (d *clusterDiscoveryGCE) findInstanceTemplates() ([]*compute.InstanceTemplate, error) {
	if d.instanceTemplates != nil {
		return d.instanceTemplates, nil
	}

	instanceTemplates, err := gce.FindInstanceTemplates(d.gceCloud, d.clusterName)
	if err != nil {
		return nil, err
	}

	d.instanceTemplates = instanceTemplates
	return d.instanceTemplates, nil
}

func (d *clusterDiscoveryGCE) listGCEInstanceTemplates() ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource

	templates, err := d.findInstanceTemplates()
	if err != nil {
		return nil, err
	}
	for _, t := range templates {
		selfLink := t.SelfLink // avoid closure-in-loop go-tcha
		resourceTracker := &resources.Resource{
			Name: t.Name,
			ID:   t.Name,
			Type: typeInstanceTemplate,
			Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
				return gce.DeleteInstanceTemplate(d.gceCloud, selfLink)
			},
			Obj: t,
		}

		for _, ni := range t.Properties.NetworkInterfaces {
			if ni.Subnetwork != "" {
				resourceTracker.Blocks = append(resourceTracker.Blocks, typeSubnet+":"+gce.LastComponent(ni.Subnetwork))
			}
		}

		klog.V(4).Infof("Found resource: %s", t.SelfLink)
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func (d *clusterDiscoveryGCE) listInstanceGroupManagersAndInstances() ([]*resources.Resource, error) {
	c := d.gceCloud
	project := c.Project()

	var resourceTrackers []*resources.Resource

	instanceTemplates := make(map[string]*compute.InstanceTemplate)
	{
		templates, err := d.findInstanceTemplates()
		if err != nil {
			return nil, err
		}
		for _, t := range templates {
			instanceTemplates[t.SelfLink] = t
		}
	}

	ctx := context.Background()

	for _, zoneName := range d.zones {
		is, err := c.Compute().InstanceGroupManagers().List(ctx, project, zoneName)
		if err != nil {
			return nil, fmt.Errorf("error listing InstanceGroupManagers: %v", err)
		}
		for i := range is {
			mig := is[i] // avoid closure-in-loop go-tcha
			instanceTemplate := instanceTemplates[mig.InstanceTemplate]
			if instanceTemplate == nil {
				klog.V(2).Infof("Ignoring MIG with unmanaged InstanceTemplate: %s", mig.InstanceTemplate)
				continue
			}

			resourceTracker := &resources.Resource{
				Name:    mig.Name,
				ID:      zoneName + "/" + mig.Name,
				Type:    typeInstanceGroupManager,
				Deleter: func(cloud fi.Cloud, r *resources.Resource) error { return gce.DeleteInstanceGroupManager(c, mig) },
				Obj:     mig,
			}

			resourceTracker.Blocks = append(resourceTracker.Blocks, typeInstanceTemplate+":"+instanceTemplate.Name)

			klog.V(4).Infof("Found resource: %s", mig.SelfLink)
			resourceTrackers = append(resourceTrackers, resourceTracker)

			instanceTrackers, err := d.listManagedInstances(mig)
			if err != nil {
				return nil, fmt.Errorf("error listing instances in InstanceGroupManager: %v", err)
			}
			resourceTrackers = append(resourceTrackers, instanceTrackers...)
		}
	}

	return resourceTrackers, nil
}

func (d *clusterDiscoveryGCE) listManagedInstances(igm *compute.InstanceGroupManager) ([]*resources.Resource, error) {
	c := d.gceCloud

	var resourceTrackers []*resources.Resource

	zoneName := gce.LastComponent(igm.Zone)

	instances, err := gce.ListManagedInstances(c, igm)
	if err != nil {
		return nil, err
	}

	for _, i := range instances {
		url := i.Instance // avoid closure-in-loop go-tcha
		name := gce.LastComponent(url)

		resourceTracker := &resources.Resource{
			Name: name,
			ID:   zoneName + "/" + name,
			Type: typeInstance,
			Deleter: func(cloud fi.Cloud, tracker *resources.Resource) error {
				return gce.DeleteInstance(c, url)
			},
			Dumper: DumpManagedInstance,
			Obj:    i,
		}

		// We don't block deletion of the instance group manager

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

// findGCEDisks finds all Disks that are associated with the current cluster
// It matches them by looking for the cluster label
func (d *clusterDiscoveryGCE) findGCEDisks() ([]*compute.Disk, error) {
	c := d.gceCloud

	clusterTag := gce.SafeClusterName(d.clusterName)

	var matches []*compute.Disk

	ctx := context.Background()

	// TODO: Push down tag filter?

	diskLists, err := c.Compute().Disks().AggregatedList(ctx, c.Project())
	if err != nil {
		return nil, fmt.Errorf("error listing disks: %v", err)
	}

	for _, list := range diskLists {
		for _, d := range list.Disks {
			match := false
			for k, v := range d.Labels {
				if k == gce.GceLabelNameKubernetesCluster {
					if v == clusterTag {
						match = true
					} else {
						match = false
						break
					}
				}
			}

			if !match {
				continue
			}

			matches = append(matches, d)
		}
	}

	return matches, nil
}

func (d *clusterDiscoveryGCE) listGCEDisks() ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource

	disks, err := d.findGCEDisks()
	if err != nil {
		return nil, err
	}
	for _, t := range disks {
		resourceTracker := &resources.Resource{
			Name:    t.Name,
			ID:      t.Name,
			Type:    typeDisk,
			Deleter: deleteGCEDisk,
			Obj:     t,
		}

		for _, u := range t.Users {
			resourceTracker.Blocked = append(resourceTracker.Blocked, typeInstance+":"+gce.LastComponent(t.Zone)+"/"+gce.LastComponent(u))
		}

		klog.V(4).Infof("Found resource: %s", t.SelfLink)
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func deleteGCEDisk(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(gce.GCECloud)
	t := r.Obj.(*compute.Disk)

	klog.V(2).Infof("Deleting GCE Disk %s", t.SelfLink)
	u, err := gce.ParseGoogleCloudURL(t.SelfLink)
	if err != nil {
		return err
	}

	op, err := c.Compute().Disks().Delete(u.Project, u.Zone, u.Name)
	if err != nil {
		if gce.IsNotFound(err) {
			klog.Infof("disk not found, assuming deleted: %q", t.SelfLink)
			return nil
		}
		return fmt.Errorf("error deleting disk %s: %v", t.SelfLink, err)
	}

	return c.WaitForOp(op)
}

func (d *clusterDiscoveryGCE) listTargetPools() ([]*resources.Resource, error) {
	c := d.gceCloud

	var resourceTrackers []*resources.Resource

	ctx := context.Background()

	tps, err := c.Compute().TargetPools().List(ctx, c.Project(), c.Region())
	if err != nil {
		return nil, fmt.Errorf("error listing TargetPools: %v", err)
	}

	for _, tp := range tps {
		if !d.matchesClusterName(tp.Name) {
			continue
		}

		resourceTracker := &resources.Resource{
			Name:    tp.Name,
			ID:      tp.Name,
			Type:    typeTargetPool,
			Deleter: deleteTargetPool,
			Obj:     tp,
		}

		klog.V(4).Infof("Found resource: %s", tp.SelfLink)
		resourceTrackers = append(resourceTrackers, resourceTracker)

		for _, healthCheckLink := range tp.HealthChecks {
			healthCheckName := gce.LastComponent(healthCheckLink)
			hc, err := c.Compute().HTTPHealthChecks().Get(c.Project(), healthCheckName)
			if err != nil {
				return nil, fmt.Errorf("error getting HTTPHealthCheck %q: %w", healthCheckName, err)
			}

			healthCheckResource := &resources.Resource{
				Name:    hc.Name,
				ID:      hc.Name,
				Type:    typeHTTPHealthcheck,
				Deleter: deleteHTTPHealthCheck,
				Obj:     hc,
			}
			healthCheckResource.Blocked = append(healthCheckResource.Blocked, resourceTracker.Type+":"+resourceTracker.ID)
			resourceTrackers = append(resourceTrackers, healthCheckResource)
		}

	}

	return resourceTrackers, nil
}

func deleteTargetPool(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(gce.GCECloud)
	t := r.Obj.(*compute.TargetPool)

	klog.V(2).Infof("Deleting GCE TargetPool %s", t.SelfLink)
	u, err := gce.ParseGoogleCloudURL(t.SelfLink)
	if err != nil {
		return err
	}

	op, err := c.Compute().TargetPools().Delete(u.Project, u.Region, u.Name)
	if err != nil {
		if gce.IsNotFound(err) {
			klog.Infof("TargetPool not found, assuming deleted: %q", t.SelfLink)
			return nil
		}
		return fmt.Errorf("error deleting TargetPool %s: %v", t.SelfLink, err)
	}

	return c.WaitForOp(op)
}

func (d *clusterDiscoveryGCE) listForwardingRules() ([]*resources.Resource, error) {
	c := d.gceCloud

	var resourceTrackers []*resources.Resource

	ctx := context.Background()

	frs, err := c.Compute().ForwardingRules().List(ctx, c.Project(), c.Region())
	if err != nil {
		return nil, fmt.Errorf("error listing ForwardingRules: %v", err)
	}

	for _, fr := range frs {
		if !d.matchesClusterName(fr.Name) {
			continue
		}

		resourceTracker := &resources.Resource{
			Name:    fr.Name,
			ID:      fr.Name,
			Type:    typeForwardingRule,
			Deleter: deleteForwardingRule,
			Obj:     fr,
		}

		if fr.Target != "" {
			resourceTracker.Blocks = append(resourceTracker.Blocks, typeTargetPool+":"+gce.LastComponent(fr.Target))
		}

		if fr.IPAddress != "" {
			resourceTracker.Blocks = append(resourceTracker.Blocks, typeAddress+":"+gce.LastComponent(fr.IPAddress))
		}

		klog.V(4).Infof("Found resource: %s", fr.SelfLink)
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func deleteForwardingRule(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(gce.GCECloud)
	t := r.Obj.(*compute.ForwardingRule)

	klog.V(2).Infof("Deleting GCE ForwardingRule %s", t.SelfLink)
	u, err := gce.ParseGoogleCloudURL(t.SelfLink)
	if err != nil {
		return err
	}

	op, err := c.Compute().ForwardingRules().Delete(u.Project, u.Region, u.Name)
	if err != nil {
		if gce.IsNotFound(err) {
			klog.Infof("ForwardingRule not found, assuming deleted: %q", t.SelfLink)
			return nil
		}
		return fmt.Errorf("error deleting ForwardingRule %s: %v", t.SelfLink, err)
	}

	return c.WaitForOp(op)
}

// listFirewallRules discovers Firewall objects for the cluster
func (d *clusterDiscoveryGCE) listFirewallRules() ([]*resources.Resource, error) {
	c := d.gceCloud

	var resourceTrackers []*resources.Resource

	ctx := context.Background()

	firewallRules, err := c.Compute().Firewalls().List(ctx, c.Project())
	if err != nil {
		return nil, fmt.Errorf("error listing FirewallRules: %v", err)
	}

nextFirewallRule:
	for _, firewallRule := range firewallRules {
		if !d.matchesClusterNameMultipart(firewallRule.Name, maxPrefixTokens) && !strings.HasPrefix(firewallRule.Name, "k8s-") {
			continue
		}

		// TODO: Check network?  (or other fields?)  No label support currently.

		// We consider only firewall rules that target our cluster tags, which include the cluster name or hash
		tagPrefix := gce.SafeClusterName(d.clusterName) + "-"
		clusterNameHash := truncate.HashString(gce.SafeClusterName(d.clusterName), 6)
		if len(firewallRule.TargetTags) != 0 {
			tagMatchCount := 0
			for _, target := range firewallRule.TargetTags {
				if strings.HasPrefix(target, tagPrefix) || strings.Contains(target, clusterNameHash) {
					tagMatchCount++
				}
			}
			if len(firewallRule.TargetTags) != tagMatchCount {
				continue nextFirewallRule
			}
		}
		// We don't have any rules that match only on source tags, but if we did we could check them here
		if len(firewallRule.TargetTags) == 0 {
			continue nextFirewallRule
		}

		firewallRuleResource := &resources.Resource{
			Name:    firewallRule.Name,
			ID:      firewallRule.Name,
			Type:    typeFirewallRule,
			Deleter: deleteFirewallRule,
			Obj:     firewallRule,
		}
		firewallRuleResource.Blocks = append(firewallRuleResource.Blocks, typeNetwork+":"+gce.LastComponent(firewallRule.Network))

		if d.matchesClusterNameMultipart(firewallRule.Name, maxPrefixTokens) {
			klog.V(4).Infof("Found resource: %s", firewallRule.SelfLink)
			resourceTrackers = append(resourceTrackers, firewallRuleResource)
		}

		// find the objects if this is a Kubernetes LoadBalancer
		if strings.HasPrefix(firewallRule.Name, "k8s-fw-") {
			// We build a list of resources if this is a k8s firewall rule,
			// but we only add them once all the checks are complete
			var k8sResources []*resources.Resource

			k8sResources = append(k8sResources, firewallRuleResource)

			// We lookup the forwarding rule by name, but we then validate that it points to one of our resources
			forwardingRuleName := strings.TrimPrefix(firewallRule.Name, "k8s-fw-")
			forwardingRule, err := c.Compute().ForwardingRules().Get(c.Project(), c.Region(), forwardingRuleName)
			if err != nil {
				if gce.IsNotFound(err) {
					// We looked it up by name, so an error isn't unlikely
					klog.Warningf("could not find forwarding rule %q, assuming firewallRule %q is not a k8s rule", forwardingRuleName, firewallRule.Name)
					continue nextFirewallRule
				}
				return nil, fmt.Errorf("error getting ForwardingRule %q: %w", forwardingRuleName, err)
			}

			forwardingRuleResource := &resources.Resource{
				Name:    forwardingRule.Name,
				ID:      forwardingRule.Name,
				Type:    typeForwardingRule,
				Deleter: deleteForwardingRule,
				Obj:     forwardingRule,
			}
			if forwardingRule.Target != "" {
				forwardingRuleResource.Blocks = append(forwardingRuleResource.Blocks, typeTargetPool+":"+gce.LastComponent(forwardingRule.Target))
			}
			k8sResources = append(k8sResources, forwardingRuleResource)

			// TODO: Can we get k8s to set labels on the ForwardingRule?

			// TODO: Check description?  It looks like e.g. description: '{"kubernetes.io/service-name":"kube-system/guestbook"}'

			if forwardingRule.Target == "" {
				klog.Warningf("forwarding rule %q did not have target, assuming firewallRule %q is not a k8s rule", forwardingRuleName, firewallRule.Name)
				continue nextFirewallRule
			}

			targetPoolName := gce.LastComponent(forwardingRule.Target)
			targetPool, err := c.Compute().TargetPools().Get(c.Project(), c.Region(), targetPoolName)
			if err != nil {
				return nil, fmt.Errorf("error getting TargetPool %q: %w", targetPoolName, err)
			}

			targetPoolResource := &resources.Resource{
				Name:    targetPool.Name,
				ID:      targetPool.Name,
				Type:    typeTargetPool,
				Deleter: deleteTargetPool,
				Obj:     targetPool,
			}
			k8sResources = append(k8sResources, targetPoolResource)

			// TODO: Check description? (looks like description: '{"kubernetes.io/service-name":"k8s-dbb09d49d9780e7e-node"}' )

			// TODO: Check instances?

			for _, healthCheckLink := range targetPool.HealthChecks {
				// l4 level healthchecks

				healthCheckName := gce.LastComponent(healthCheckLink)
				if !strings.HasPrefix(healthCheckName, "k8s-") || !strings.Contains(healthCheckLink, "/httpHealthChecks/") {
					klog.Warningf("found non-k8s healthcheck %q in targetPool %q, assuming firewallRule %q is not a k8s rule", healthCheckLink, targetPoolName, firewallRule.Name)
					continue nextFirewallRule
				}

				hc, err := c.Compute().HTTPHealthChecks().Get(c.Project(), healthCheckName)
				if err != nil {
					return nil, fmt.Errorf("error getting HTTPHealthCheck %q: %w", healthCheckName, err)
				}

				// TODO: Check description? (looks like description: '{"kubernetes.io/service-name":"k8s-dbb09d49d9780e7e-node"}' )

				healthCheckResource := &resources.Resource{
					Name:    hc.Name,
					ID:      hc.Name,
					Type:    typeHTTPHealthcheck,
					Deleter: deleteHTTPHealthCheck,
					Obj:     hc,
				}
				healthCheckResource.Blocked = append(healthCheckResource.Blocked, targetPoolResource.Type+":"+targetPoolResource.ID)

				k8sResources = append(k8sResources, healthCheckResource)

			}

			// We now have confidence that this is a k8s LoadBalancer; add the resources
			resourceTrackers = append(resourceTrackers, k8sResources...)
		}

		// find the objects if this is a Kubernetes node health check
		if strings.HasPrefix(firewallRule.Name, "k8s-") && strings.HasSuffix(firewallRule.Name, "-node-http-hc") {
			// TODO: Check port matches http health check (always 10256?)
			// TODO: Check description - looks like '{"kubernetes.io/cluster-id":"cb2e931dec561053"}'

			// We already know the target tags match
			resourceTrackers = append(resourceTrackers, firewallRuleResource)
		}
	}

	return resourceTrackers, nil
}

// deleteHTTPHealthCheck is the helper function to delete a Resource for a HTTP health check object
func deleteHTTPHealthCheck(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(gce.GCECloud)
	t := r.Obj.(*compute.HttpHealthCheck)

	klog.V(2).Infof("Deleting GCE HTTP HealthCheck %s", t.SelfLink)
	u, err := gce.ParseGoogleCloudURL(t.SelfLink)
	if err != nil {
		return err
	}

	op, err := c.Compute().HTTPHealthChecks().Delete(u.Project, u.Name)
	if err != nil {
		if gce.IsNotFound(err) {
			klog.Infof("HTTP HealthCheck not found, assuming deleted: %q", t.SelfLink)
			return nil
		}
		return fmt.Errorf("error deleting HTTP HealthCheck %s: %v", t.SelfLink, err)
	}

	return c.WaitForOp(op)
}

// deleteFirewallRule is the helper function to delete a Resource for a Firewall object
func deleteFirewallRule(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(gce.GCECloud)
	t := r.Obj.(*compute.Firewall)

	klog.V(2).Infof("Deleting GCE FirewallRule %s", t.SelfLink)
	u, err := gce.ParseGoogleCloudURL(t.SelfLink)
	if err != nil {
		return err
	}

	op, err := c.Compute().Firewalls().Delete(u.Project, u.Name)
	if err != nil {
		if gce.IsNotFound(err) {
			klog.Infof("FirewallRule not found, assuming deleted: %q", t.SelfLink)
			return nil
		}
		return fmt.Errorf("error deleting FirewallRule %s: %v", t.SelfLink, err)
	}

	return c.WaitForOp(op)
}

func (d *clusterDiscoveryGCE) listRoutes(ctx context.Context, resourceMap map[string]*resources.Resource) ([]*resources.Resource, error) {
	c := d.gceCloud

	var resourceTrackers []*resources.Resource

	instancesToDelete := make(map[string]*resources.Resource)
	for _, resource := range resourceMap {
		if resource.Type == typeInstance {
			instancesToDelete[resource.ID] = resource
		}
	}

	// TODO: Push-down prefix?
	routes, err := c.Compute().Routes().List(ctx, c.Project())
	if err != nil {
		return nil, fmt.Errorf("error listing Routes: %w", err)
	}
	for _, r := range routes {
		if !d.matchesClusterNameWithUUID(r.Name, maxGCERouteNameLength) {
			continue
		}
		remove := false
		for _, w := range r.Warnings {
			switch w.Code {
			case "NEXT_HOP_INSTANCE_NOT_FOUND":
				remove = true
			default:
				klog.Infof("Unknown warning on route %q: %q", r.Name, w.Code)
			}
		}

		if r.NextHopInstance != "" {
			u, err := gce.ParseGoogleCloudURL(r.NextHopInstance)
			if err != nil {
				klog.Warningf("error parsing URL for NextHopInstance=%q", r.NextHopInstance)
			}

			if instancesToDelete[u.Zone+"/"+u.Name] != nil {
				remove = true
			}
		}

		if remove {
			resourceTracker := &resources.Resource{
				Name:    r.Name,
				ID:      r.Name,
				Type:    typeRoute,
				Deleter: deleteRoute,
				Obj:     r,
			}

			// To avoid race conditions where the control-plane re-adds the routes, we delete routes
			// only after we have deleted all the instances.
			for _, instance := range instancesToDelete {
				resourceTracker.Blocked = append(resourceTracker.Blocked, typeInstance+":"+instance.ID)
			}

			klog.V(4).Infof("Found resource: %s", r.SelfLink)
			resourceTrackers = append(resourceTrackers, resourceTracker)
		}
	}

	return resourceTrackers, nil
}

func deleteRoute(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(gce.GCECloud)
	t := r.Obj.(*compute.Route)

	klog.V(2).Infof("Deleting GCE Route %s", t.SelfLink)
	u, err := gce.ParseGoogleCloudURL(t.SelfLink)
	if err != nil {
		return err
	}

	op, err := c.Compute().Routes().Delete(u.Project, u.Name)
	if err != nil {
		if gce.IsNotFound(err) {
			klog.Infof("Route not found, assuming deleted: %q", t.SelfLink)
			return nil
		}
		return fmt.Errorf("error deleting Route %s: %v", t.SelfLink, err)
	}

	return c.WaitForOp(op)
}

func (d *clusterDiscoveryGCE) listAddresses() ([]*resources.Resource, error) {
	c := d.gceCloud

	var resourceTrackers []*resources.Resource

	ctx := context.Background()

	addrs, err := c.Compute().Addresses().List(ctx, c.Project(), c.Region())
	if err != nil {
		return nil, fmt.Errorf("error listing Addresses: %v", err)
	}

	for _, a := range addrs {
		if !d.matchesClusterName(a.Name) {
			klog.V(8).Infof("Skipping Address with name %q", a.Name)
			continue
		}

		resourceTracker := &resources.Resource{
			Name:    a.Name,
			ID:      a.Name,
			Type:    typeAddress,
			Deleter: deleteAddress,
			Obj:     a,
		}

		klog.V(4).Infof("Found resource: %s", a.SelfLink)
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func deleteAddress(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(gce.GCECloud)
	t := r.Obj.(*compute.Address)

	klog.V(2).Infof("Deleting GCE Address %s", t.SelfLink)
	u, err := gce.ParseGoogleCloudURL(t.SelfLink)
	if err != nil {
		return err
	}

	op, err := c.Compute().Addresses().Delete(u.Project, u.Region, u.Name)
	if err != nil {
		if gce.IsNotFound(err) {
			klog.Infof("Address not found, assuming deleted: %q", t.SelfLink)
			return nil
		}
		return fmt.Errorf("error deleting Address %s: %v", t.SelfLink, err)
	}

	return c.WaitForOp(op)
}

func (d *clusterDiscoveryGCE) listSubnets() ([]*resources.Resource, error) {
	// Templates are very accurate because of the metadata, so use those as the sanity check
	templates, err := d.findInstanceTemplates()
	if err != nil {
		return nil, err
	}
	subnetworkUrls := make(map[string]bool)
	for _, t := range templates {
		for _, ni := range t.Properties.NetworkInterfaces {
			if ni.Subnetwork != "" {
				subnetworkUrls[ni.Subnetwork] = true
			}
		}
	}

	c := d.gceCloud

	var resourceTrackers []*resources.Resource
	ctx := context.Background()

	subnets, err := c.Compute().Subnetworks().List(ctx, c.Project(), c.Region())
	if err != nil {
		return nil, fmt.Errorf("error listing subnetworks: %v", err)
	}

	for _, o := range subnets {
		if !d.matchesClusterName(o.Name) {
			klog.V(8).Infof("skipping Subnet with name %q", o.Name)
			continue
		}

		if !subnetworkUrls[o.SelfLink] {
			klog.Warningf("skipping subnetwork %q because it didn't match any instance template", o.SelfLink)
			continue
		}

		resourceTracker := &resources.Resource{
			Name:    o.Name,
			ID:      o.Name,
			Type:    typeSubnet,
			Deleter: deleteSubnet,
			Obj:     o,
			Dumper:  DumpSubnetwork,
		}

		resourceTracker.Blocks = append(resourceTracker.Blocks, typeNetwork+":"+gce.LastComponent(o.Network))

		klog.V(4).Infof("found resource: %s", o.SelfLink)
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func deleteSubnet(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(gce.GCECloud)
	o := r.Obj.(*compute.Subnetwork)

	klog.V(2).Infof("deleting GCE subnetwork %s", o.SelfLink)
	u, err := gce.ParseGoogleCloudURL(o.SelfLink)
	if err != nil {
		return err
	}

	op, err := c.Compute().Subnetworks().Delete(u.Project, u.Region, u.Name)
	if err != nil {
		if gce.IsNotFound(err) {
			klog.Infof("subnetwork not found, assuming deleted: %q", o.SelfLink)
			return nil
		}
		return fmt.Errorf("error deleting subnetwork %s: %v", o.SelfLink, err)
	}

	return c.WaitForOp(op)
}

func (d *clusterDiscoveryGCE) listRouters() ([]*resources.Resource, error) {
	c := d.gceCloud

	var resourceTrackers []*resources.Resource
	ctx := context.Background()

	routers, err := c.Compute().Routers().List(ctx, c.Project(), c.Region())
	if err != nil {
		return nil, fmt.Errorf("error listing routers: %v", err)
	}

	for _, o := range routers {
		if !d.matchesClusterName(o.Name) {
			klog.V(8).Infof("skipping Router with name %q", o.Name)
			continue
		}

		resourceTracker := &resources.Resource{
			Name:    o.Name,
			ID:      o.Name,
			Type:    typeRouter,
			Deleter: deleteRouter,
			Obj:     o,
		}

		klog.V(4).Infof("found resource: %s", o.SelfLink)
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func deleteRouter(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(gce.GCECloud)
	o := r.Obj.(*compute.Router)

	klog.V(2).Infof("deleting GCE router %s", o.SelfLink)
	u, err := gce.ParseGoogleCloudURL(o.SelfLink)
	if err != nil {
		return err
	}

	op, err := c.Compute().Routers().Delete(u.Project, u.Region, u.Name)
	if err != nil {
		if gce.IsNotFound(err) {
			klog.Infof("router not found, assuming deleted: %q", o.SelfLink)
			return nil
		}
		return fmt.Errorf("error deleting router %s: %v", o.SelfLink, err)
	}

	return c.WaitForOp(op)
}

func (d *clusterDiscoveryGCE) listServiceAccounts() ([]*resources.Resource, error) {
	c := d.gceCloud
	ctx := context.Background()

	sas, err := c.IAM().ServiceAccounts().List(ctx, fmt.Sprintf("projects/%s", c.Project()))
	if err != nil {
		return nil, fmt.Errorf("error listing ServiceAccounts %w", err)
	}
	var resourceTrackers []*resources.Resource
	for _, sa := range sas {
		tokens := strings.Split(gce.LastComponent(sa.Name), "@")
		if len(tokens) != 2 {
			return nil, fmt.Errorf("Invalid service account email '%s'", gce.LastComponent(sa.Name))
		}
		accountID := tokens[0]
		names := []string{gce.ControlPlane, gce.Bastion, gce.Node}
		for _, name := range names {
			generatedName := gce.ServiceAccountName(name, d.clusterName)
			if generatedName == accountID {
				resourceTracker := &resources.Resource{
					Name:    gce.LastComponent(sa.Name),
					ID:      sa.Name,
					Type:    typeServiceAccount,
					Deleter: deleteServiceAccount,
					Obj:     sa,
				}

				klog.V(4).Infof("found resource: %s", sa.Name)
				resourceTrackers = append(resourceTrackers, resourceTracker)
				break
			}
		}
	}
	return resourceTrackers, nil
}

func deleteServiceAccount(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(gce.GCECloud)
	o := r.Obj.(*iam.ServiceAccount)

	klog.V(2).Infof("deleting GCE ServiceAccount %s", o.Name)
	_, err := c.IAM().ServiceAccounts().Delete(o.Name)
	return err
}

// containsOnlyListedIGMs returns true if all the given backend service's backends
// are contained in the provided list of IGM resources.
func containsOnlyListedIGMs(svc *compute.BackendService, igms []*resources.Resource) bool {
	for _, be := range svc.Backends {
		listed := false
		for _, igm := range igms {
			// NOTE: this should be sufficient / strict enough since IGM names include the cluster
			// that they are part of, but revisit if naming conventions change.
			if strings.HasSuffix(be.Group, "/"+igm.Name) {
				listed = true
				break
			}
		}

		if !listed {
			return false
		}
	}
	return true
}

func (d *clusterDiscoveryGCE) listBackendServices() ([]*resources.Resource, error) {
	c := d.gceCloud

	svcs, err := c.Compute().RegionBackendServices().List(context.Background(), c.Project(), c.Region())
	if err != nil {
		if gce.IsNotFound(err) {
			klog.Infof("backend services not found, assuming none exist in project: %q region: %q", c.Project(), c.Region())
			return nil, nil
		}
		return nil, fmt.Errorf("Failed to list backend services: %w", err)
	}
	// TODO: cache, for efficiency, if needed.
	// Find all relevant backend services by finding all the cluster's IGMs, and then
	// listing all backend services in the project / region, then selecting
	// the backend services which contain only the relevant IGMs.
	igms, err := d.listInstanceGroupManagersAndInstances()
	if err != nil {
		return nil, err
	}
	var bs []*resources.Resource
	for _, svc := range svcs {
		if containsOnlyListedIGMs(svc, igms) {
			bs = append(bs, &resources.Resource{
				Name: svc.Name,
				ID:   svc.Name,
				Type: typeBackendService,
				Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
					op, err := c.Compute().RegionBackendServices().Delete(c.Project(), c.Region(), svc.Name)
					if err != nil {
						return err
					}
					return c.WaitForOp(op)
				},
				Obj: svc,
			})
		}
	}

	return bs, nil
}

func (d *clusterDiscoveryGCE) listHealthchecks() ([]*resources.Resource, error) {
	c := d.gceCloud
	// TODO: cache, for efficiency, if needed.
	// Find relevant healthchecks by finding all the backend services relevant to this
	// cluster, then selecting all the healthchecks they use.
	backendServices, err := d.listBackendServices()
	if err != nil {
		return nil, err
	}
	hcs := make(map[string]struct{})
	for _, bs := range backendServices {
		bsObj, ok := bs.Obj.(*compute.BackendService)
		if !ok {
			return nil, fmt.Errorf("%T is not a *compute.BackendService", bs)
		}
		for _, hc := range bsObj.HealthChecks {
			hcs[hc] = struct{}{}
		}
	}
	var hcResources []*resources.Resource
	for hc := range hcs {
		hcResources = append(hcResources, &resources.Resource{
			Name: gce.LastComponent(hc),
			ID:   gce.LastComponent(hc),
			Type: typeHealthcheck,
			Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
				op, err := c.Compute().RegionHealthChecks().Delete(c.Project(), c.Region(), gce.LastComponent(hc))
				if err != nil {
					return err
				}
				return c.WaitForOp(op)
			},
			Obj: hc,
		})
	}

	return hcResources, nil
}

func (d *clusterDiscoveryGCE) listNetworks() ([]*resources.Resource, error) {
	// Templates are very accurate because of the metadata, so use those as the sanity check
	templates, err := d.findInstanceTemplates()
	if err != nil {
		return nil, err
	}
	networkUrls := make(map[string]bool)
	for _, t := range templates {
		for _, ni := range t.Properties.NetworkInterfaces {
			if ni.Network != "" {
				networkUrls[ni.Network] = true
			}
		}
	}

	c := d.gceCloud

	var resourceTrackers []*resources.Resource

	networks, err := c.Compute().Networks().List(c.Project())
	if err != nil {
		return nil, fmt.Errorf("error listing networks: %v", err)
	}

	for _, o := range networks.Items {
		if o.Name != gce.SafeTruncatedClusterName(d.clusterName, 63) {
			klog.V(8).Infof("skipping network with name %q", o.Name)
			continue
		}

		if !networkUrls[o.SelfLink] {
			klog.Warningf("skipping network %q because it didn't match any instance template", o.SelfLink)
			continue
		}

		resourceTracker := &resources.Resource{
			Name:    o.Name,
			ID:      o.Name,
			Type:    typeNetwork,
			Deleter: deleteNetwork,
			Obj:     o,
			Dumper:  DumpNetwork,
		}

		klog.V(4).Infof("found resource: %s", o.SelfLink)
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func deleteNetwork(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(gce.GCECloud)
	o := r.Obj.(*compute.Network)

	klog.V(2).Infof("deleting GCE network %s", o.SelfLink)
	u, err := gce.ParseGoogleCloudURL(o.SelfLink)
	if err != nil {
		return err
	}

	op, err := c.Compute().Networks().Delete(u.Project, u.Name)
	if err != nil {
		if gce.IsNotFound(err) {
			klog.Infof("network not found, assuming deleted: %q", o.SelfLink)
			return nil
		}
		return fmt.Errorf("error deleting network %s: %v", o.SelfLink, err)
	}

	return c.WaitForOp(op)
}

func (d *clusterDiscoveryGCE) matchesClusterName(name string) bool {
	// Names could have hypens in them, so really there is no limit.
	// 8 hyphens feels like enough for any "reasonable" name
	maxParts := 8

	return d.matchesClusterNameMultipart(name, maxParts)
}

// matchesClusterNameMultipart checks if the name could have been generated by our cluster
// considering all the prefixes separated by `-`.  maxParts limits the number of parts we consider.
func (d *clusterDiscoveryGCE) matchesClusterNameMultipart(name string, maxParts int) bool {
	tokens := strings.Split(name, "-")

	for i := 1; i <= maxParts; i++ {
		if i > len(tokens) {
			break
		}

		id := strings.Join(tokens[:i], "-")
		if id == "" {
			continue
		}

		safeName := gce.SafeObjectName(id, d.clusterName)
		clusterNameHash := truncate.HashString(gce.SafeClusterName(d.clusterName), 6)

		if name == safeName || strings.Contains(name, clusterNameHash) {
			return true
		}
	}
	return false
}

// matchesClusterNameWithUUID checks if the name is the clusterName with a UUID on the end.
// This is used by GCE routes (in "classic" mode)
func (d *clusterDiscoveryGCE) matchesClusterNameWithUUID(name string, maxLength int) bool {
	const uuidLength = 36 // e.g. 51a343e2-c285-4e73-b933-18a6ea44c3e4

	// Format is <cluster-name>-<uuid>
	// <cluster-name> is truncated to ensure it fits into the GCE max length
	if len(name) < uuidLength {
		return false
	}
	withoutUUID := name[:len(name)-uuidLength]

	clusterPrefix := gce.SafeClusterName(d.clusterName) + "-"
	if len(clusterPrefix) > maxLength-uuidLength {
		clusterPrefix = gce.SafeClusterName(d.clusterName)[:maxLength-uuidLength-1] + "-"
	}

	return clusterPrefix == withoutUUID
}

func (d *clusterDiscoveryGCE) clusterDNSName() string {
	return d.clusterName + "."
}

func (d *clusterDiscoveryGCE) isKopsManagedDNSName(name string) bool {
	prefix := []string{`api`, `api.internal`, `bastion`}
	for _, p := range prefix {
		if name == p+"."+d.clusterDNSName() {
			return true
		}
	}
	return false
}

func (d *clusterDiscoveryGCE) listGCEDNSZone() ([]*resources.Resource, error) {
	if dns.IsGossipHostname(d.clusterName) {
		return nil, nil
	}

	var resourceTrackers []*resources.Resource

	managedZones, err := d.gceCloud.CloudDNS().ManagedZones().List(d.gceCloud.Project())
	if err != nil {
		return nil, fmt.Errorf("error getting GCE DNS zones %v", err)
	}

	for _, zone := range managedZones {
		if !strings.HasSuffix(d.clusterDNSName(), zone.DnsName) {
			continue
		}
		rrsets, err := d.gceCloud.CloudDNS().ResourceRecordSets().List(d.gceCloud.Project(), zone.Name)
		if err != nil {
			return nil, fmt.Errorf("error getting GCE DNS zone data %v", err)
		}

		for _, record := range rrsets {
			// adapted from AWS implementation
			if record.Type != "A" {
				continue
			}

			if d.isKopsManagedDNSName(record.Name) {
				resource := resources.Resource{
					Name:         record.Name,
					ID:           record.Name,
					Type:         typeDNSRecord,
					GroupDeleter: deleteDNSRecords,
					GroupKey:     zone.Name,
					Obj:          record,
				}
				resourceTrackers = append(resourceTrackers, &resource)
			}
		}
	}

	return resourceTrackers, nil
}

func deleteDNSRecords(cloud fi.Cloud, r []*resources.Resource) error {
	c := cloud.(gce.GCECloud)
	var records []*clouddns.ResourceRecordSet
	var zoneName string

	for _, record := range r {
		r := record.Obj.(*clouddns.ResourceRecordSet)
		zoneName = record.GroupKey
		records = append(records, r)
	}

	change := clouddns.Change{Deletions: records, Kind: "dns#change", IsServing: true}
	_, err := c.CloudDNS().Changes().Create(c.Project(), zoneName, &change)
	if err != nil {
		return fmt.Errorf("error deleting GCE DNS resource record set %v", err)
	}
	return nil
}
