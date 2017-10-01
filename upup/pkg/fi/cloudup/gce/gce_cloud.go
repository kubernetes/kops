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

package gce

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/storage/v1"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	"k8s.io/kubernetes/federation/pkg/dnsprovider/providers/google/clouddns"
)

type GCECloud interface {
	fi.Cloud
	Compute() *compute.Service
	Storage() *storage.Service

	Region() string
	Project() string
	WaitForOp(op *compute.Operation) error
	GetApiIngressStatus(cluster *kops.Cluster) ([]kops.ApiIngressStatus, error)
	Labels() map[string]string

	// FindClusterStatus gets the status of the cluster as it exists in GCE, inferred from volumes
	FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error)

	// FindInstanceTemplates finds all instance templates that are associated with the current cluster
	// It matches them by looking for instance metadata with key='cluster-name' and value of our cluster name
	FindInstanceTemplates(clusterName string) ([]*compute.InstanceTemplate, error)

	Zones() ([]string, error)
}

type gceCloudImplementation struct {
	compute *compute.Service
	storage *storage.Service

	region  string
	project string

	labels map[string]string
}

var _ fi.Cloud = &gceCloudImplementation{}

func (c *gceCloudImplementation) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderGCE
}

var gceCloudInstances map[string]GCECloud = make(map[string]GCECloud)

func NewGCECloud(region string, project string, labels map[string]string) (GCECloud, error) {
	i := gceCloudInstances[region+"::"+project]
	if i != nil {
		return i.(gceCloudInternal).WithLabels(labels), nil
	}

	c := &gceCloudImplementation{region: region, project: project}

	ctx := context.Background()

	client, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		return nil, fmt.Errorf("error building google API client: %v", err)
	}
	computeService, err := compute.New(client)
	if err != nil {
		return nil, fmt.Errorf("error building compute API client: %v", err)
	}
	c.compute = computeService

	storageService, err := storage.New(client)
	if err != nil {
		return nil, fmt.Errorf("error building storage API client: %v", err)
	}
	c.storage = storageService

	gceCloudInstances[region+"::"+project] = c

	return c.WithLabels(labels), nil
}

// gceCloudInternal is an interface for private functions for a gceCloudImplemention or mockGCECloud
type gceCloudInternal interface {
	// WithLabels returns a copy of the GCECloud, bound to the specified labels
	WithLabels(labels map[string]string) GCECloud
}

// WithLabels returns a copy of the GCECloud, bound to the specified labels
func (c *gceCloudImplementation) WithLabels(labels map[string]string) GCECloud {
	i := &gceCloudImplementation{}
	*i = *c
	i.labels = labels
	return i
}

// Compute returns private struct element compute.
func (c *gceCloudImplementation) Compute() *compute.Service {
	return c.compute
}

// Storage returns private struct element storage.
func (c *gceCloudImplementation) Storage() *storage.Service {
	return c.storage
}

// Region returns private struct element region.
func (c *gceCloudImplementation) Region() string {
	return c.region
}

// Project returns private struct element project.
func (c *gceCloudImplementation) Project() string {
	return c.project
}

func (c *gceCloudImplementation) DNS() (dnsprovider.Interface, error) {
	provider, err := clouddns.CreateInterface(c.project, nil)
	if err != nil {
		return nil, fmt.Errorf("Error building (k8s) DNS provider: %v", err)
	}
	return provider, nil
}

func (c *gceCloudImplementation) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	glog.Warningf("FindVPCInfo not (yet) implemented on GCE")
	return nil, nil
}

func (c *gceCloudImplementation) Labels() map[string]string {
	// Defensive copy
	tags := make(map[string]string)
	for k, v := range c.labels {
		tags[k] = v
	}
	return tags
}

// TODO refactor this out of resources
// this is needed for delete groups and other new methods

// Zones returns the zones in a region
func (c *gceCloudImplementation) Zones() ([]string, error) {

	var zones []string
	// TODO: Only zones in api.Cluster object, if we have one?
	gceZones, err := c.Compute().Zones.List(c.Project()).Do()
	if err != nil {
		return nil, fmt.Errorf("error listing zones: %v", err)
	}
	for _, gceZone := range gceZones.Items {
		u, err := ParseGoogleCloudURL(gceZone.Region)
		if err != nil {
			return nil, err
		}
		if u.Name != c.Region() {
			continue
		}
		zones = append(zones, gceZone.Name)
	}
	if len(zones) == 0 {
		return nil, fmt.Errorf("unable to determine zones in region %q", c.Region())
	}

	glog.Infof("Scanning zones: %v", zones)
	return zones, nil
}

func (c *gceCloudImplementation) WaitForOp(op *compute.Operation) error {
	return WaitForOp(c.compute, op)
}

func (c *gceCloudImplementation) GetApiIngressStatus(cluster *kops.Cluster) ([]kops.ApiIngressStatus, error) {
	var ingresses []kops.ApiIngressStatus

	// Note that this must match GCEModelContext::NameForForwardingRule
	name := SafeObjectName("api", cluster.ObjectMeta.Name)

	glog.V(2).Infof("Querying GCE to find ForwardingRules for API (%q)", name)
	forwardingRule, err := c.compute.ForwardingRules.Get(c.project, c.region, name).Do()
	if err != nil {
		if !IsNotFound(err) {
			forwardingRule = nil
		} else {
			return nil, fmt.Errorf("error getting ForwardingRule %q: %v", name, err)
		}
	}

	if forwardingRule != nil {
		if forwardingRule.IPAddress == "" {
			return nil, fmt.Errorf("Found forward rule %q, but it did not have an IPAddress", name)
		}

		ingresses = append(ingresses, kops.ApiIngressStatus{
			IP: forwardingRule.IPAddress,
		})
	}

	return ingresses, nil
}

// DeleteGroup deletes a cloud of instances controlled by an Instance Group Manager
func (c *gceCloudImplementation) DeleteGroup(name string, template string) error {
	glog.V(8).Infof("gce cloud provider DeleteGroup not tested yet")
	// TODO we need to check for order and when we can delete these.
	ctx := context.Background()
	_, err := c.compute.InstanceGroupManagers.Delete(c.project, c.region, name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("error deleting instance group manager: %v", err)
	}

	_, err = c.compute.InstanceTemplates.Delete(c.project, template).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("error deleting instance template: %v", err)
	}

	return nil
}

// DeleteInstance deletes a GCE instance
func (c *gceCloudImplementation) DeleteInstance(id *string) error {
	glog.V(8).Infof("gce cloud provider DeleteInstance not tested yet")
	instanceId := fi.StringValue(id)
	if instanceId == "" {
		return fmt.Errorf("error deleting instance no id provided")
	}

	_, err := c.compute.Instances.Delete(c.project, c.region, instanceId).Do()
	if err != nil {
		// TODO should we check googleapi.IsNotModified
		// to check whether the returned error was because
		// http.StatusNotModified was returned.
		// Not sure if it works here
		return fmt.Errorf("error deleting instance: %v", err)
	}
	return nil
}

// GetCloudGroups returns a map of CloudGroup that backs a list of instance groups
func (c *gceCloudImplementation) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	glog.V(8).Infof("gce cloud provider GetCloudGroups not implemented yet")
	ctx := context.Background()
	nodeMap := cloudinstances.GetNodeMap(nodes)
	groups := make(map[string]*cloudinstances.CloudInstanceGroup)

	instanceTemplates := make(map[string]*compute.InstanceTemplate)
	{
		templates, err := c.FindInstanceTemplates(cluster.ObjectMeta.Name)
		if err != nil {
			return nil, err
		}
		for _, t := range templates {
			instanceTemplates[t.SelfLink] = t
		}
	}

	var migs []*compute.InstanceGroupManager

	// With the zone changes I do not think we have to do this
	zones, err := c.Zones()
	if err != nil {
		return nil, err
	}
	// I am not sure how and if we need to iterate through the zones now
	// This needs to be fixed
	err := c.Compute().InstanceGroupManagers.List(c.Project(), zoneName).Pages(ctx, func(page *compute.InstanceGroupManagerList) error {
		for _, mig := range page.Items {
			instanceTemplate := instanceTemplates[mig.InstanceTemplate]
			if instanceTemplate == nil {
				glog.V(2).Infof("Ignoring MIG with unmanaged InstanceTemplate: %s", mig.InstanceTemplate)
				continue
			}

			migs = append(migs, mig)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error listing InstanceGroupManagers: %v", err)
	}

	for _, mig := range migs {
		name := LastComponent(mig.Name)
		glog.V(8).Infof("searching for %s", name)
		var instancegroup *kops.InstanceGroup
		// FIXME why can't we use the mig zone?
		// TODO I think @justinsb's zone PR will help with this so we do not have
		// to loop through the zones
		var igZoneName string
		// We do not need to loop through the zones.
		for _, zoneName := range zones {
			for _, g := range instancegroups {
				// Why is gce group name so different from aws?
				groupName := fmt.Sprintf("%s-%s-%s", zoneName, g.ObjectMeta.Name, cluster.ObjectMeta.Name)
				groupName = strings.Replace(groupName, ".", "-", -1)
				if name == groupName {
					igZoneName = zoneName
					if instancegroup != nil {
						return nil, fmt.Errorf("found multiple instance groups matching cloud groups %q", groupName)
					}
					instancegroup = g
				}
			}

			if instancegroup != nil {
				break
			}
		}
		if instancegroup == nil {
			if warnUnmatched {
				glog.Warningf("Found managed instance group with no corresponding instance group %q", name)
			}
			continue
		}
		groups[instancegroup.ObjectMeta.Name], err = c.gceBuildCloudInstanceGroup(instancegroup, mig, igZoneName, nodeMap)
		if err != nil {
			return nil, fmt.Errorf("error getting cloud instance group %q: %v", instancegroup.ObjectMeta.Name, err)
		}
	}
	return groups, nil
}

func (c *gceCloudImplementation) gceBuildCloudInstanceGroup(ig *kops.InstanceGroup, g *compute.InstanceGroupManager, zoneName string, nodeMap map[string]*v1.Node) (*cloudinstances.CloudInstanceGroup, error) {
	n, err := cloudinstances.NewCloudInstanceGroup(g.InstanceGroup, g.InstanceTemplate, ig, int(g.TargetSize), int(g.TargetSize))
	if err != nil {
		return nil, fmt.Errorf("error creating cloud instance group: %v", err)
	}

	instances, err := c.Compute().InstanceGroupManagers.ListManagedInstances(c.Project(), zoneName, g.Name).Do()
	if err != nil {
		return nil, fmt.Errorf("error listing ManagedInstances in %s: %v", g.Name, err)
	}

	for _, i := range instances.ManagedInstances {
		name := LastComponent(i.Instance)
		instance, err := c.Compute().Instances.Get(c.Project(), zoneName, name).Do()
		if err != nil {
			return nil, fmt.Errorf("error getting instance %s: %v", name, err)
		}
		if instance == nil {
			return nil, fmt.Errorf("unable to get instance %q", name)
		}

		var instanceTemplate string
		for _, item := range instance.Metadata.Items {
			if item.Key == "instance-template" {
				instanceTemplate = item.Value
				break
			}
		}
		if instanceTemplate == "" {
			glog.Warningf("unable to get template name for instance: %q", name)
		}
		nodeId := fmt.Sprintf("%v", i.Id)
		// nodeId is different between aws and gce, we need to make the mode more generic
		// I had aws pass in the same name twice, and gce pass in the nodeid
		err = n.NewCloudInstanceMember(&name, nodeId, g.InstanceTemplate, instanceTemplate, nodeMap)
		if err != nil {
			return nil, fmt.Errorf("error creating cloud instance group member: %v", err)
		}
	}

	// FIXME this was changed n.MarkIsReady()
	return n, nil
}

// FindInstanceTemplates finds all instance templates that are associated with the current cluster
// It matches them by looking for instance metadata with key='cluster-name' and value of our cluster name
func (c *gceCloudImplementation) FindInstanceTemplates(clusterName string) ([]*compute.InstanceTemplate, error) {
	findClusterName := strings.TrimSpace(clusterName)
	var matches []*compute.InstanceTemplate
	ctx := context.Background()

	err := c.Compute().InstanceTemplates.List(c.Project()).Pages(ctx, func(page *compute.InstanceTemplateList) error {
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

	return matches, nil
}


