/*
Copyright 2017 The Kubernetes Authors.

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
	"encoding/base32"
	"fmt"
	"hash/fnv"
	"strings"

	context "golang.org/x/net/context"
	compute "google.golang.org/api/compute/v0.beta"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
)

// DeleteGroup deletes a cloud of instances controlled by an Instance Group Manager
func (c *gceCloudImplementation) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	return deleteCloudInstanceGroup(c, g)
}

// deleteCloudInstanceGroup deletes the InstanceGroupManager and current InstanceTemplate
func deleteCloudInstanceGroup(c GCECloud, g *cloudinstances.CloudInstanceGroup) error {
	mig := g.Raw.(*compute.InstanceGroupManager)
	err := DeleteInstanceGroupManager(c, mig)
	if err != nil {
		return err
	}

	return DeleteInstanceTemplate(c, mig.InstanceTemplate)
}

// DeleteGroup implements fi.Cloud::DeleteGroup
func (c *mockGCECloud) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	return deleteCloudInstanceGroup(c, g)
}

// DeleteInstance deletes a GCE instance
func (c *gceCloudImplementation) DeleteInstance(i *cloudinstances.CloudInstanceGroupMember) error {
	return recreateCloudInstanceGroupMember(c, i)
}

// DeleteInstance deletes a GCE instance
func (c *mockGCECloud) DeleteInstance(i *cloudinstances.CloudInstanceGroupMember) error {
	return recreateCloudInstanceGroupMember(c, i)
}

// recreateCloudInstanceGroupMember recreates the specified instances, managed by an InstanceGroupManager
func recreateCloudInstanceGroupMember(c GCECloud, i *cloudinstances.CloudInstanceGroupMember) error {
	mig := i.CloudInstanceGroup.Raw.(*compute.InstanceGroupManager)

	klog.V(2).Infof("Recreating GCE Instance %s in MIG %s", i.ID, mig.Name)

	migURL, err := ParseGoogleCloudURL(mig.SelfLink)
	if err != nil {
		return err
	}

	req := &compute.InstanceGroupManagersRecreateInstancesRequest{
		Instances: []string{
			i.ID,
		},
	}
	op, err := c.Compute().InstanceGroupManagers.RecreateInstances(migURL.Project, migURL.Zone, migURL.Name, req).Do()
	if err != nil {
		if IsNotFound(err) {
			klog.Infof("Instance not found, assuming deleted: %q", i.ID)
			return nil
		}
		return fmt.Errorf("error recreating Instance %s: %v", i.ID, err)
	}

	return c.WaitForOp(op)
}

// GetCloudGroups returns a map of CloudGroup that backs a list of instance groups
func (c *gceCloudImplementation) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	return getCloudGroups(c, cluster, instancegroups, warnUnmatched, nodes)
}

func getCloudGroups(c GCECloud, cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	groups := make(map[string]*cloudinstances.CloudInstanceGroup)

	project := c.Project()
	ctx := context.Background()

	nodesByProviderID := make(map[string]*v1.Node)

	for i := range nodes {
		node := &nodes[i]
		nodesByProviderID[node.Spec.ProviderID] = node
	}

	// There is some code duplication with resources/gce.go here, but more in the structure than a straight copy-paste

	// The strategy:
	// * Find the InstanceTemplates, matching on tags
	// * Find InstanceGroupManagers attached to those templates
	// * Find Instances attached to those InstanceGroupManagers

	instanceTemplates := make(map[string]*compute.InstanceTemplate)
	{
		templates, err := FindInstanceTemplates(c, cluster.Name)
		if err != nil {
			return nil, err
		}

		for _, t := range templates {
			instanceTemplates[t.SelfLink] = t
		}
	}

	zones, err := c.Zones()
	if err != nil {
		return nil, err
	}

	for _, zoneName := range zones {
		err := c.Compute().InstanceGroupManagers.List(project, zoneName).Pages(ctx, func(page *compute.InstanceGroupManagerList) error {
			for _, mig := range page.Items {
				name := mig.Name

				instanceTemplate := instanceTemplates[mig.InstanceTemplate]
				if instanceTemplate == nil {
					klog.V(2).Infof("ignoring MIG %s with unmanaged InstanceTemplate: %s", name, mig.InstanceTemplate)
					continue
				}

				ig, err := matchInstanceGroup(mig, cluster, instancegroups)
				if err != nil {
					return fmt.Errorf("error getting instance group for MIG %q", name)
				}
				if ig == nil {
					if warnUnmatched {
						klog.Warningf("Found MIG with no corresponding instance group %q", name)
					}
					continue
				}

				g := &cloudinstances.CloudInstanceGroup{
					HumanName:     mig.Name,
					InstanceGroup: ig,
					MinSize:       int(mig.TargetSize),
					MaxSize:       int(mig.TargetSize),
					Raw:           mig,
				}
				groups[mig.Name] = g

				latestInstanceTemplate := mig.InstanceTemplate

				instances, err := ListManagedInstances(c, mig)
				if err != nil {
					return err
				}

				for _, i := range instances {
					id := i.Instance
					cm := &cloudinstances.CloudInstanceGroupMember{
						ID:                 id,
						CloudInstanceGroup: g,
					}

					// Try first by provider ID
					name := LastComponent(id)
					providerID := "gce://" + project + "/" + zoneName + "/" + name
					node := nodesByProviderID[providerID]

					if node != nil {
						cm.Node = node
					} else {
						klog.V(8).Infof("unable to find node for instance: %s", id)
					}

					if i.Version != nil && latestInstanceTemplate == i.Version.InstanceTemplate {
						g.Ready = append(g.Ready, cm)
					} else {
						g.NeedUpdate = append(g.NeedUpdate, cm)
					}
				}

			}
			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("error listing InstanceGroupManagers: %v", err)
		}
	}

	return groups, nil
}

// NameForInstanceGroupManager builds a name for an InstanceGroupManager in the specified zone
func NameForInstanceGroupManager(c *kops.Cluster, ig *kops.InstanceGroup, zone string) string {
	shortZone := zone
	lastDash := strings.LastIndex(shortZone, "-")
	if lastDash != -1 {
		shortZone = shortZone[lastDash+1:]
	}
	name := SafeObjectName(shortZone+"."+ig.ObjectMeta.Name, c.ObjectMeta.Name)
	name = LimitedLengthName(name, 63)
	return name
}

// LimitedLengthName returns a string subject to a maximum length
func LimitedLengthName(s string, n int) string {
	// We only use the hash if we need to
	if len(s) <= n {
		return s
	}

	h := fnv.New32a()
	if _, err := h.Write([]byte(s)); err != nil {
		klog.Fatalf("error hashing values: %v", err)
	}
	hashString := base32.HexEncoding.EncodeToString(h.Sum(nil))
	hashString = strings.ToLower(hashString)
	if len(hashString) > 6 {
		hashString = hashString[:6]
	}

	maxBaseLength := n - len(hashString) - 1
	if len(s) > maxBaseLength {
		s = s[:maxBaseLength]
	}
	s = s + "-" + hashString

	return s
}

// matchInstanceGroup filters a list of instancegroups for recognized cloud groups
func matchInstanceGroup(mig *compute.InstanceGroupManager, c *kops.Cluster, instancegroups []*kops.InstanceGroup) (*kops.InstanceGroup, error) {
	migName := LastComponent(mig.Name)
	var matches []*kops.InstanceGroup
	for _, ig := range instancegroups {
		name := NameForInstanceGroupManager(c, ig, LastComponent(mig.Zone))
		if name == migName {
			matches = append(matches, ig)
		}
	}

	if len(matches) == 0 {
		return nil, nil
	}
	if len(matches) != 1 {
		return nil, fmt.Errorf("found multiple instance groups matching MIG %q", mig.Name)
	}
	return matches[0], nil
}
