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

package kutil

import (
	"fmt"
	"github.com/golang/glog"
	compute "google.golang.org/api/compute/v0.beta"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"strings"
)

type gceListFn func() ([]*ResourceTracker, error)

const (
	typeInstanceTemplate = "instancetemplate"
	typeDisk             = "disk"
)

func (c *AwsCluster) listResourcesGCE() (map[string]*ResourceTracker, error) {
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
		d.listGCEDisks,
		d.listGCEInstanceGroupManagers,

		// TODO: Find routes via instances (via instance groups)
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

	templates, err := c.Compute.InstanceTemplates.List(c.Project).Do()
	if err != nil {
		return nil, fmt.Errorf("error listing instance groups: %v", err)
	}

	for _, t := range templates.Items {
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

	return c.WaitForGlobalOp(op)
}

func (d *clusterDiscoveryGCE) listGCEInstanceGroupManagers() ([]*ResourceTracker, error) {
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

	for _, zoneName := range d.zones {
		migs, err := c.Compute.InstanceGroupManagers.List(project, zoneName).Do()
		if err != nil {
			return nil, fmt.Errorf("error listing InstanceGroupManagers: %v", err)
		}

		for _, mig := range migs.Items {
			instanceTemplate := instanceTemplates[mig.InstanceTemplate]
			if instanceTemplate == nil {
				glog.V(2).Infof("Ignoring MIG with unmanaged InstanceTemplate: %s", mig.InstanceTemplate)
				continue
			}

			tracker := &ResourceTracker{
				Name:    mig.Name,
				ID:      zoneName + "/" + mig.Name,
				Type:    "instancegroupmanager",
				deleter: deleteInstanceGroupManager,
				obj:     mig,
			}

			tracker.blocks = append(tracker.blocks, typeInstanceTemplate+":"+instanceTemplate.Name)

			trackers = append(trackers, tracker)
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

	glog.Infof("MIG: %s", fi.DebugAsJsonString(t))

	op, err := c.Compute.InstanceGroupManagers.Delete(u.Project, u.Zone, u.Name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			glog.Infof("InstanceGroupManager not found, assuming deleted: %q", t.SelfLink)
			return nil
		}
		return fmt.Errorf("error deleting InstanceGroupManager %s: %v", t.SelfLink, err)
	}

	return c.WaitForZoneOp(op, u.Zone)
}

// findGCEDisks finds all Disks that are associated with the current cluster
// It matches them by looking for the cluster label
func (d *clusterDiscoveryGCE) findGCEDisks() ([]*compute.Disk, error) {
	c := d.gceCloud

	clusterTag := gce.SafeClusterName(d.clusterName)

	var matches []*compute.Disk

	// TODO: Push down tag filter?

	disks, err := c.Compute.Disks.AggregatedList(c.Project).Do()
	if err != nil {
		return nil, fmt.Errorf("error listing disks: %v", err)
	}

	for _, list := range disks.Items {
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

	return c.WaitForZoneOp(op, u.Zone)
}
