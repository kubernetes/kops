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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/storage/v1"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/kops/pkg/apis/kops"
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

func NewGCECloud(region string, project string, labels map[string]string) (GCECloud, error) {
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

	c.labels = labels

	return c, nil
}

func (c *gceCloudImplementation) DeleteGroup(name string, template string) error {

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

func (c *gceCloudImplementation) DeleteInstance(id *string) error {
	op, err := c.compute.Instances.Delete(c.project, c.region, *id).Do()
	if err != nil {
		/*
			if gce.IsNotFound(err) {
				return fmt.Errorf("error finding instance: %v", err)
			}*/
		return fmt.Errorf("error deleting instance: %v", err)
	}
	if err := c.WaitForOp(op); err != nil {
		return fmt.Errorf("error deleting instance: %v", err)
	}
	return nil
}

func (c *gceCloudImplementation) Compute() *compute.Service {
	return c.compute
}
func (c *gceCloudImplementation) Storage() *storage.Service {
	return c.storage
}
func (c *gceCloudImplementation) Region() string {
	return c.region
}
func (c *gceCloudImplementation) Project() string {
	return c.project
}

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
		return nil, fmt.Errorf("unable to determine zones in region %q", c.Region)
	}

	glog.Infof("Scanning zones: %v", zones)
	return zones, nil
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

func (c *gceCloudImplementation) FindCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodeMap map[string]*v1.Node) (map[string]*fi.CloudGroup, error) {
	var groups map[string]*fi.CloudGroup
	ctx := context.Background()

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

	// TODO we need to iterate through the instance groups rather can mig
	zones, err := c.Zones()
	if err != nil {
		return nil, err
	}
	for _, zoneName := range zones {
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
	}

	for _, mig := range migs {
		name := mig.Name
		var instancegroup *kops.InstanceGroup
		for _, g := range instancegroups {
			var asgName string
			switch g.Spec.Role {
			case kops.InstanceGroupRoleMaster:
				asgName = g.ObjectMeta.Name + ".masters." + cluster.ObjectMeta.Name
			case kops.InstanceGroupRoleNode:
				asgName = g.ObjectMeta.Name + "." + cluster.ObjectMeta.Name
			case kops.InstanceGroupRoleBastion:
				asgName = g.ObjectMeta.Name + "." + cluster.ObjectMeta.Name
			default:
				glog.Warningf("Ignoring InstanceGroup of unknown role %q", g.Spec.Role)
				continue
			}

			if name == asgName {
				if instancegroup != nil {
					return nil, fmt.Errorf("Found multiple instance groups matching ASG %q", asgName)
				}
				instancegroup = g
			}
		}
		if instancegroup == nil {
			if warnUnmatched {
				glog.Warningf("Found ASG with no corresponding instance group %q", name)
			}
			continue
		}
		groups[instancegroup.ObjectMeta.Name] = c.gceBuildCloudInstanceGroup(instancegroup, mig, nodeMap)
	}
	return groups, nil
}

func (c *gceCloudImplementation) gceBuildCloudInstanceGroup(ig *kops.InstanceGroup, g *compute.InstanceGroupManager, nodeMap map[string]*v1.Node) *fi.CloudGroup {
	n := &fi.CloudGroup{
		GroupName:         g.Name,
		InstanceGroup:     ig,
		GroupTemplateName: g.InstanceTemplate,
	}

	// TODO FIXME
	//readyLaunchConfigurationName := g.InstanceTemplate
	// This call is not paginated
	instances, _ := c.Compute().InstanceGroupManagers.ListManagedInstances(c.Project(), c.Region(), g.Name).Do()
	/*
		if err != nil {
			return fmt.Errorf("error listing ManagedInstances in %s: %v", igm.Name, err)
		}*/

	for _, i := range instances.ManagedInstances {
		name := LastComponent(i.Instance)

		c := &fi.CloudGroupInstance{
			// FIXME
			//ID: c.Zone() + "/" + name,
			ID: aws.String(name),
		}

		// FIXME not sure if this will work :)
		node := nodeMap[i.Instance]
		if node != nil {
			c.Node = node
		}

		n.NeedUpdate = append(n.NeedUpdate, c)

		/* not certain how to do this
		if readyLaunchConfigurationName == aws.StringValue(i.InstanceTemplate) {
			n.Ready = append(n.Ready, c)
		} else {
			n.NeedUpdate = append(n.NeedUpdate, c)
		}*/
	}

	if len(n.NeedUpdate) == 0 {
		n.Status = "Ready"
	} else {
		n.Status = "NeedsUpdate"
	}

	return n
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
