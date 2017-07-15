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

package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/glog"
	compute "google.golang.org/api/compute/v0.beta"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
)

type gceListFn func() ([]*ResourceTracker, error)

const (
	typeInstance             = "Instance"
	typeInstanceTemplate     = "InstanceTemplate"
	typeDisk                 = "Disk"
	typeInstanceGroupManager = "InstanceGroupManager"
	typeTargetPool           = "TargetPool"
	typeForwardingRule       = "ForwardingRule"
	typeAddress              = "Address"
	typeRoute                = "Route"
)

func (c *ClusterResources) listResourcesGCE() (map[string]*ResourceTracker, error) {
	gceCloud := c.Cloud.(*gce.GCECloud)
	if c.Region == "" {
		c.Region = gceCloud.Region
	}

	resources := make(map[string]*ResourceTracker)

	d := &clusterDiscoveryGCE{
		cloud:       c.Cloud,
		gceCloud:    gceCloud,
		clusterName: c.ClusterName,
	}

	{
		// TODO: Only zones in api.Cluster object, if we have one?
		gceZones, err := d.gceCloud.Compute.Zones.List(d.gceCloud.Project).Do()
		if err != nil {
			return nil, fmt.Errorf("error listing zones: %v", err)
		}
		for _, gceZone := range gceZones.Items {
			u, err := gce.ParseGoogleCloudURL(gceZone.Region)
			if err != nil {
				return nil, err
			}
			if u.Name != c.Region {
				continue
			}
			d.zones = append(d.zones, gceZone.Name)
		}
		if len(d.zones) == 0 {
			return nil, fmt.Errorf("unable to determine zones in region %q", c.Region)
		}
		glog.Infof("Scanning zones: %v", d.zones)
	}

	listFunctions := []gceListFn{
		d.listGCEInstanceTemplates,
		d.listInstanceGroupManagersAndInstances,
		d.listTargetPools,
		d.listForwardingRules,
		d.listGCEDisks,
		d.listGCEDNSZone,
		// TODO: Find routes via instances (via instance groups)
		d.listAddresses,
	}
	for _, fn := range listFunctions {
		trackers, err := fn()
		if err != nil {
			return nil, err
		}
		for _, t := range trackers {
			resources[t.Type+":"+t.ID] = t
		}
	}

	// We try to clean up orphaned routes.
	// Technically we still have a race condition here - until the master(s) are terminated, they will keep
	// creating routes.  Another option might be to have a post-destroy cleanup, and only remove routes with no target.
	{
		trackers, err := d.listRoutes(resources)
		if err != nil {
			return nil, err
		}
		for _, t := range trackers {
			resources[t.Type+":"+t.ID] = t
		}
	}

	for k, t := range resources {
		if t.done {
			delete(resources, k)
		}
	}
	return resources, nil
}

type clusterDiscoveryGCE struct {
	cloud       fi.Cloud
	gceCloud    *gce.GCECloud
	clusterName string

	instanceTemplates []*compute.InstanceTemplate
	zones             []string
}

// findInstanceTemplates finds all instance templates that are associated with the current cluster
// It matches them by looking for instance metadata with key='cluster-name' and value of our cluster name
func (d *clusterDiscoveryGCE) findInstanceTemplates() ([]*compute.InstanceTemplate, error) {
	if d.instanceTemplates != nil {
		return d.instanceTemplates, nil
	}

	c := d.gceCloud

	//clusterTag := gce.SafeClusterName(strings.TrimSpace(d.clusterName))

	findClusterName := strings.TrimSpace(d.clusterName)

	var matches []*compute.InstanceTemplate

	ctx := context.Background()

	err := c.Compute.InstanceTemplates.List(c.Project).Pages(ctx, func(page *compute.InstanceTemplateList) error {
		for _, t := range page.Items {
			match := false
			for _, item := range t.Properties.Metadata.Items {
				if item.Key == "cluster-name" {
					if strings.TrimSpace(item.Value) == findClusterName {
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

			matches = append(matches, t)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error listing instance groups: %v", err)
	}

	d.instanceTemplates = matches

	return matches, nil
}

func (d *clusterDiscoveryGCE) listGCEInstanceTemplates() ([]*ResourceTracker, error) {
	var trackers []*ResourceTracker

	templates, err := d.findInstanceTemplates()
	if err != nil {
		return nil, err
	}
	for _, t := range templates {
		tracker := &ResourceTracker{
			Name:    t.Name,
			ID:      t.Name,
			Type:    typeInstanceTemplate,
			deleter: deleteGCEInstanceTemplate,
			obj:     t,
		}

		glog.V(4).Infof("Found resource: %s", t.SelfLink)
		trackers = append(trackers, tracker)
	}

	return trackers, nil
}

func deleteGCEInstanceTemplate(cloud fi.Cloud, r *ResourceTracker) error {
	c := cloud.(*gce.GCECloud)
	t := r.obj.(*compute.InstanceTemplate)

	glog.V(2).Infof("Deleting GCE InstanceTemplate %s", t.SelfLink)
	u, err := gce.ParseGoogleCloudURL(t.SelfLink)
	if err != nil {
		return err
	}

	op, err := c.Compute.InstanceTemplates.Delete(u.Project, u.Name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			glog.Infof("instancetemplate not found, assuming deleted: %q", t.SelfLink)
			return nil
		}
		return fmt.Errorf("error deleting InstanceTemplate %s: %v", t.SelfLink, err)
	}

	return c.WaitForOp(op)
}

func (d *clusterDiscoveryGCE) listInstanceGroupManagersAndInstances() ([]*ResourceTracker, error) {
	c := d.gceCloud
	project := c.Project

	var trackers []*ResourceTracker

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
		err := c.Compute.InstanceGroupManagers.List(project, zoneName).Pages(ctx, func(page *compute.InstanceGroupManagerList) error {
			for _, mig := range page.Items {
				instanceTemplate := instanceTemplates[mig.InstanceTemplate]
				if instanceTemplate == nil {
					glog.V(2).Infof("Ignoring MIG with unmanaged InstanceTemplate: %s", mig.InstanceTemplate)
					continue
				}

				tracker := &ResourceTracker{
					Name:    mig.Name,
					ID:      zoneName + "/" + mig.Name,
					Type:    typeInstanceGroupManager,
					deleter: deleteInstanceGroupManager,
					obj:     mig,
				}

				tracker.blocks = append(tracker.blocks, typeInstanceTemplate+":"+instanceTemplate.Name)

				glog.V(4).Infof("Found resource: %s", mig.SelfLink)
				trackers = append(trackers, tracker)

				instanceTrackers, err := d.listManagedInstances(mig)
				if err != nil {
					return fmt.Errorf("error listing instances in InstanceGroupManager: %v", err)
				}
				trackers = append(trackers, instanceTrackers...)
			}
			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("error listing InstanceGroupManagers: %v", err)
		}
	}

	return trackers, nil
}

func deleteInstanceGroupManager(cloud fi.Cloud, r *ResourceTracker) error {
	c := cloud.(*gce.GCECloud)
	t := r.obj.(*compute.InstanceGroupManager)

	glog.V(2).Infof("Deleting GCE InstanceGroupManager %s", t.SelfLink)
	u, err := gce.ParseGoogleCloudURL(t.SelfLink)
	if err != nil {
		return err
	}

	//glog.Infof("MIG: %s", fi.DebugAsJsonString(t))

	op, err := c.Compute.InstanceGroupManagers.Delete(u.Project, u.Zone, u.Name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			glog.Infof("InstanceGroupManager not found, assuming deleted: %q", t.SelfLink)
			return nil
		}
		return fmt.Errorf("error deleting InstanceGroupManager %s: %v", t.SelfLink, err)
	}

	return c.WaitForOp(op)
}

func (d *clusterDiscoveryGCE) listManagedInstances(igm *compute.InstanceGroupManager) ([]*ResourceTracker, error) {
	c := d.gceCloud
	project := c.Project

	var trackers []*ResourceTracker

	zoneName := gce.LastComponent(igm.Zone)

	// This call is not paginated
	instances, err := c.Compute.InstanceGroupManagers.ListManagedInstances(project, zoneName, igm.Name).Do()
	if err != nil {
		return nil, fmt.Errorf("error listing ManagedInstances in %s: %v", igm.Name, err)
	}

	for _, i := range instances.ManagedInstances {
		name := gce.LastComponent(i.Instance)

		tracker := &ResourceTracker{
			Name:    name,
			ID:      zoneName + "/" + name,
			Type:    typeInstance,
			deleter: deleteManagedInstance,
			obj:     i.Instance,
		}

		// We don't block deletion of the instance group manager

		trackers = append(trackers, tracker)
	}

	return trackers, nil
}

// findGCEDisks finds all Disks that are associated with the current cluster
// It matches them by looking for the cluster label
func (d *clusterDiscoveryGCE) findGCEDisks() ([]*compute.Disk, error) {
	c := d.gceCloud

	clusterTag := gce.SafeClusterName(d.clusterName)

	var matches []*compute.Disk

	ctx := context.Background()

	// TODO: Push down tag filter?

	err := c.Compute.Disks.AggregatedList(c.Project).Pages(ctx, func(page *compute.DiskAggregatedList) error {
		for _, list := range page.Items {
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

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error listing disks: %v", err)
	}

	return matches, nil
}

func (d *clusterDiscoveryGCE) listGCEDisks() ([]*ResourceTracker, error) {
	var trackers []*ResourceTracker

	disks, err := d.findGCEDisks()
	if err != nil {
		return nil, err
	}
	for _, t := range disks {
		tracker := &ResourceTracker{
			Name:    t.Name,
			ID:      t.Name,
			Type:    typeDisk,
			deleter: deleteGCEDisk,
			obj:     t,
		}

		for _, u := range t.Users {
			tracker.blocked = append(tracker.blocked, typeInstance+":"+gce.LastComponent(t.Zone)+"/"+gce.LastComponent(u))
		}

		glog.V(4).Infof("Found resource: %s", t.SelfLink)
		trackers = append(trackers, tracker)
	}

	return trackers, nil
}

func deleteGCEDisk(cloud fi.Cloud, r *ResourceTracker) error {
	c := cloud.(*gce.GCECloud)
	t := r.obj.(*compute.Disk)

	glog.V(2).Infof("Deleting GCE Disk %s", t.SelfLink)
	u, err := gce.ParseGoogleCloudURL(t.SelfLink)
	if err != nil {
		return err
	}

	op, err := c.Compute.Disks.Delete(u.Project, u.Zone, u.Name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			glog.Infof("disk not found, assuming deleted: %q", t.SelfLink)
			return nil
		}
		return fmt.Errorf("error deleting disk %s: %v", t.SelfLink, err)
	}

	return c.WaitForOp(op)
}

func (d *clusterDiscoveryGCE) listTargetPools() ([]*ResourceTracker, error) {
	c := d.gceCloud

	var trackers []*ResourceTracker

	ctx := context.Background()

	err := c.Compute.TargetPools.List(c.Project, c.Region).Pages(ctx, func(page *compute.TargetPoolList) error {
		for _, tp := range page.Items {
			if !d.matchesClusterName(tp.Name) {
				continue
			}

			tracker := &ResourceTracker{
				Name:    tp.Name,
				ID:      tp.Name,
				Type:    typeTargetPool,
				deleter: deleteTargetPool,
				obj:     tp,
			}

			glog.V(4).Infof("Found resource: %s", tp.SelfLink)
			trackers = append(trackers, tracker)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error listing TargetPools: %v", err)
	}

	return trackers, nil
}

func deleteTargetPool(cloud fi.Cloud, r *ResourceTracker) error {
	c := cloud.(*gce.GCECloud)
	t := r.obj.(*compute.TargetPool)

	glog.V(2).Infof("Deleting GCE TargetPool %s", t.SelfLink)
	u, err := gce.ParseGoogleCloudURL(t.SelfLink)
	if err != nil {
		return err
	}

	glog.Infof("TargetPool: %s", fi.DebugAsJsonString(t))

	op, err := c.Compute.TargetPools.Delete(u.Project, u.Region, u.Name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			glog.Infof("TargetPool not found, assuming deleted: %q", t.SelfLink)
			return nil
		}
		return fmt.Errorf("error deleting TargetPool %s: %v", t.SelfLink, err)
	}

	return c.WaitForOp(op)
}

func (d *clusterDiscoveryGCE) listForwardingRules() ([]*ResourceTracker, error) {
	c := d.gceCloud

	var trackers []*ResourceTracker

	ctx := context.Background()

	err := c.Compute.ForwardingRules.List(c.Project, c.Region).Pages(ctx, func(page *compute.ForwardingRuleList) error {
		for _, fr := range page.Items {
			if !d.matchesClusterName(fr.Name) {
				continue
			}

			tracker := &ResourceTracker{
				Name:    fr.Name,
				ID:      fr.Name,
				Type:    typeForwardingRule,
				deleter: deleteForwardingRule,
				obj:     fr,
			}

			if fr.Target != "" {
				tracker.blocks = append(tracker.blocks, typeTargetPool+":"+gce.LastComponent(fr.Target))
			}

			if fr.IPAddress != "" {
				tracker.blocks = append(tracker.blocks, typeAddress+":"+gce.LastComponent(fr.IPAddress))
			}

			glog.V(4).Infof("Found resource: %s", fr.SelfLink)
			trackers = append(trackers, tracker)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error listing ForwardingRules: %v", err)
	}

	return trackers, nil
}

func deleteForwardingRule(cloud fi.Cloud, r *ResourceTracker) error {
	c := cloud.(*gce.GCECloud)
	t := r.obj.(*compute.ForwardingRule)

	glog.V(2).Infof("Deleting GCE ForwardingRule %s", t.SelfLink)
	u, err := gce.ParseGoogleCloudURL(t.SelfLink)
	if err != nil {
		return err
	}

	op, err := c.Compute.ForwardingRules.Delete(u.Project, u.Region, u.Name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			glog.Infof("ForwardingRule not found, assuming deleted: %q", t.SelfLink)
			return nil
		}
		return fmt.Errorf("error deleting ForwardingRule %s: %v", t.SelfLink, err)
	}

	return c.WaitForOp(op)
}

func deleteManagedInstance(cloud fi.Cloud, r *ResourceTracker) error {
	c := cloud.(*gce.GCECloud)
	selfLink := r.obj.(string)

	glog.V(2).Infof("Deleting GCE Instance %s", selfLink)
	u, err := gce.ParseGoogleCloudURL(selfLink)
	if err != nil {
		return err
	}

	op, err := c.Compute.Instances.Delete(u.Project, u.Zone, u.Name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			glog.Infof("Instance not found, assuming deleted: %q", selfLink)
			return nil
		}
		return fmt.Errorf("error deleting Instance %s: %v", selfLink, err)
	}

	return c.WaitForOp(op)
}

func (d *clusterDiscoveryGCE) listRoutes(resources map[string]*ResourceTracker) ([]*ResourceTracker, error) {
	c := d.gceCloud

	var trackers []*ResourceTracker

	instances := sets.NewString()
	for _, resource := range resources {
		if resource.Type == typeInstance {
			instances.Insert(resource.ID)
		}
	}

	prefix := gce.SafeClusterName(d.clusterName) + "-"

	ctx := context.Background()

	// TODO: Push-down prefix?
	err := c.Compute.Routes.List(c.Project).Pages(ctx, func(page *compute.RouteList) error {
		for _, r := range page.Items {
			if !strings.HasPrefix(r.Name, prefix) {
				continue
			}
			remove := false
			for _, w := range r.Warnings {
				switch w.Code {
				case "NEXT_HOP_INSTANCE_NOT_FOUND":
					remove = true
				default:
					glog.Infof("Unknown warning on route %q: %q", r.Name, w.Code)
				}
			}

			if r.NextHopInstance != "" {
				u, err := gce.ParseGoogleCloudURL(r.NextHopInstance)
				if err != nil {
					glog.Warningf("error parsing URL for NextHopInstance=%q", r.NextHopInstance)
				}

				if instances.Has(u.Zone + "/" + u.Name) {
					remove = true
				}
			}

			if remove {
				tracker := &ResourceTracker{
					Name:    r.Name,
					ID:      r.Name,
					Type:    typeRoute,
					deleter: deleteRoute,
					obj:     r,
				}

				// We don't need to block
				//if r.NextHopInstance != "" {
				//	tracker.blocked = append(tracker.blocks, typeInstance+":"+gce.LastComponent(r.NextHopInstance))
				//}

				glog.V(4).Infof("Found resource: %s", r.SelfLink)
				trackers = append(trackers, tracker)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error listing Routes: %v", err)
	}
	return trackers, nil
}

func deleteRoute(cloud fi.Cloud, r *ResourceTracker) error {
	c := cloud.(*gce.GCECloud)
	t := r.obj.(*compute.Route)

	glog.V(2).Infof("Deleting GCE Route %s", t.SelfLink)
	u, err := gce.ParseGoogleCloudURL(t.SelfLink)
	if err != nil {
		return err
	}

	op, err := c.Compute.Routes.Delete(u.Project, u.Name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			glog.Infof("Route not found, assuming deleted: %q", t.SelfLink)
			return nil
		}
		return fmt.Errorf("error deleting Route %s: %v", t.SelfLink, err)
	}

	return c.WaitForOp(op)
}

func (d *clusterDiscoveryGCE) listAddresses() ([]*ResourceTracker, error) {
	c := d.gceCloud

	var trackers []*ResourceTracker

	ctx := context.Background()

	err := c.Compute.Addresses.List(c.Project, c.Region).Pages(ctx, func(page *compute.AddressList) error {
		for _, a := range page.Items {
			if !d.matchesClusterName(a.Name) {
				glog.V(8).Infof("Skipping Address with name %q", a.Name)
				continue
			}

			tracker := &ResourceTracker{
				Name:    a.Name,
				ID:      a.Name,
				Type:    typeAddress,
				deleter: deleteAddress,
				obj:     a,
			}

			glog.V(4).Infof("Found resource: %s", a.SelfLink)
			trackers = append(trackers, tracker)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error listing Addresses: %v", err)
	}

	return trackers, nil
}

func deleteAddress(cloud fi.Cloud, r *ResourceTracker) error {
	c := cloud.(*gce.GCECloud)
	t := r.obj.(*compute.Address)

	glog.V(2).Infof("Deleting GCE Address %s", t.SelfLink)
	u, err := gce.ParseGoogleCloudURL(t.SelfLink)
	if err != nil {
		return err
	}

	op, err := c.Compute.Addresses.Delete(u.Project, u.Region, u.Name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			glog.Infof("Address not found, assuming deleted: %q", t.SelfLink)
			return nil
		}
		return fmt.Errorf("error deleting Address %s: %v", t.SelfLink, err)
	}

	return c.WaitForOp(op)
}

func (d *clusterDiscoveryGCE) matchesClusterName(name string) bool {
	firstDash := strings.Index(name, "-")
	if firstDash == -1 {
		return false
	}

	id := name[:firstDash]
	return name == gce.SafeObjectName(id, d.clusterName)
}

func (d *clusterDiscoveryGCE) listGCEDNSZone() ([]*ResourceTracker, error) {
	if dns.IsGossipHostname(d.clusterName) {
		return nil, nil
	}

	zone, err := d.findDNSZone()
	if err != nil {
		return nil, err
	}

	return []*ResourceTracker{
		{
			Name:    zone.Name(),
			ID:      zone.Name(),
			Type:    "DNS Zone",
			deleter: d.deleteDNSZone,
			obj:     zone,
		},
	}, nil
}

func (d *clusterDiscoveryGCE) findDNSZone() (dnsprovider.Zone, error) {
	dnsProvider, err := d.cloud.DNS()
	if err != nil {
		return nil, fmt.Errorf("Error getting dnsprovider: %v", err)
	}

	zonesLister, supported := dnsProvider.Zones()
	if !supported {
		return nil, fmt.Errorf("DNS provier does not support listing zones: %v", err)
	}

	allZones, err := zonesLister.List()
	if err != nil {
		return nil, fmt.Errorf("Error listing dns zones: %v", err)
	}

	for _, zone := range allZones {
		if strings.Contains(d.clusterName, strings.TrimSuffix(zone.Name(), ".")) {
			return zone, nil
		}
	}

	return nil, fmt.Errorf("DNS Zone for cluster %s could not be found", d.clusterName)
}

func (d *clusterDiscoveryGCE) deleteDNSZone(cloud fi.Cloud, r *ResourceTracker) error {
	clusterZone := r.obj.(dnsprovider.Zone)

	rrs, supported := clusterZone.ResourceRecordSets()
	if !supported {
		return fmt.Errorf("ResourceRecordSets not supported with clouddns")
	}
	records, err := rrs.List()
	if err != nil {
		return fmt.Errorf("Failed to list resource records")
	}

	changeset := rrs.StartChangeset()
	for _, record := range records {
		if record.Type() != "A" {
			continue
		}

		name := record.Name()
		name = "." + strings.TrimSuffix(name, ".")
		prefix := strings.TrimSuffix(name, "."+d.clusterName)

		remove := false
		// TODO: Compute the actual set of names?
		if prefix == ".api" || prefix == ".api.internal" {
			remove = true
		} else if strings.HasPrefix(prefix, ".etcd-") {
			remove = true
		}

		if !remove {
			continue
		}

		changeset.Remove(record)
	}

	if changeset.IsEmpty() {
		return nil
	}

	err = changeset.Apply()
	if err != nil {
		return fmt.Errorf("Error deleting cloud dns records: %v", err)
	}

	return nil
}
