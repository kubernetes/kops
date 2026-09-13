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
	"cmp"
	"context"
	"encoding/base32"
	"fmt"
	"hash/fnv"
	"slices"
	"strings"
	"time"

	compute "google.golang.org/api/compute/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
)

// PollingInterval is the polling interval to use when waiting for all MIG Instances to be deleted
var PollingInterval = 5 * time.Second

// DeleteGroup deletes a cloud of instances controlled by an Instance Group Manager
func (c *gceCloudImplementation) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	return DeleteCloudInstanceGroup(c, g)
}

// DeleteCloudInstanceGroup deletes the InstanceGroupManager and current InstanceTemplate
func DeleteCloudInstanceGroup(c GCECloud, g *cloudinstances.CloudInstanceGroup) error {
	mig := g.Raw.(*compute.InstanceGroupManager)
	err := DeleteMIGInstances(c, mig)
	if err != nil {
		return err
	}

	timeout := time.Now().Add(10 * time.Minute)

	klog.Infof("Waiting for instances in MIG %q to terminate...", mig.Name)
	for {
		if time.Now().After(timeout) {
			return fmt.Errorf("timed out waiting for instances in MIG %q to terminate", mig.Name)
		}

		instances, err := ListManagedInstances(c, mig)
		if err != nil {
			return fmt.Errorf("error listing managed instances for %q: %v", mig.Name, err)
		}
		if len(instances) == 0 {
			klog.Infof("All instances in MIG %q terminated", mig.Name)
			break
		}
		klog.Infof("%d instance(s) remaining in MIG %q, waiting...", len(instances), mig.Name)
		time.Sleep(PollingInterval)
	}

	err = DeleteInstanceGroupManager(c, mig)
	if err != nil {
		return err
	}

	return DeleteInstanceTemplate(c, mig.InstanceTemplate)
}

// DeleteInstance deletes a GCE instance
func (c *gceCloudImplementation) DeleteInstance(i *cloudinstances.CloudInstance) error {
	return recreateCloudInstance(c, i)
}

func (c *gceCloudImplementation) DeregisterInstance(i *cloudinstances.CloudInstance) error {
	klog.V(8).Info("GCE DeregisterInstance not implemented")
	return nil
}

// DetachInstance is not implemented yet. It needs to cause a cloud instance to no longer be counted against the group's size limits.
func (c *gceCloudImplementation) DetachInstance(i *cloudinstances.CloudInstance) error {
	klog.V(8).Info("gce cloud provider DetachInstance not implemented yet")
	return fmt.Errorf("gce cloud provider does not support surging")
}

// recreateCloudInstance recreates the specified instances, managed by an InstanceGroupManager
func recreateCloudInstance(c GCECloud, i *cloudinstances.CloudInstance) error {
	mig := i.CloudInstanceGroup.Raw.(*compute.InstanceGroupManager)

	klog.V(2).Infof("Recreating GCE Instance %s in MIG %s", i.ID, mig.Name)

	migURL, err := ParseGoogleCloudURL(mig.SelfLink)
	if err != nil {
		return err
	}

	op, err := c.Compute().InstanceGroupManagers().RecreateInstances(migURL.Project, migURL.Zone, migURL.Name, i.ID)
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
	return GetCloudGroups(c, cluster, instancegroups, warnUnmatched, nodes)
}

func GetCloudGroups(c GCECloud, cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
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
		migs, err := c.Compute().InstanceGroupManagers().List(ctx, project, zoneName)
		if err != nil {
			return nil, fmt.Errorf("error listing InstanceGroupManagers: %v", err)
		}
		for _, mig := range migs {
			name := mig.Name

			instanceTemplate := instanceTemplates[mig.InstanceTemplate]
			if instanceTemplate == nil {
				klog.V(2).Infof("ignoring MIG %s with unmanaged InstanceTemplate: %s", name, mig.InstanceTemplate)
				continue
			}

			ig, err := matchInstanceGroup(mig, cluster, instancegroups)
			if err != nil {
				return nil, fmt.Errorf("error getting instance group for MIG %q", name)
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
				TargetSize:    int(mig.TargetSize),
				MaxSize:       int(mig.TargetSize),
				Raw:           mig,
			}
			groups[mig.Name] = g

			latestInstanceTemplate := mig.InstanceTemplate

			instances, err := ListManagedInstances(c, mig)
			if err != nil {
				return nil, err
			}

			for _, i := range instances {
				id := i.Instance
				name := LastComponent(id)
				instance, err := c.Compute().Instances().Get(project, zoneName, name)
				if err != nil {
					if !IsNotFound(err) {
						return nil, fmt.Errorf("error getting Instance: %v", err)
					}
					klog.Warningf("Instance %s not found, it may not have been created", name)
					continue
				}
				cm := &cloudinstances.CloudInstance{
					ID:                 instance.SelfLink,
					CloudInstanceGroup: g,
				}
				addCloudInstanceData(cm, instance)

				// Try first by provider ID
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
	}

	return groups, nil
}

// NameForInstanceGroupManager builds a name for an InstanceGroupManager in the specified zone
func NameForInstanceGroupManager(clusterName, instanceGroupName, zone string) string {
	shortZone := zone
	lastDash := strings.LastIndex(shortZone, "-")
	if lastDash != -1 {
		shortZone = shortZone[lastDash+1:]
	}
	name := SafeObjectName(shortZone+"."+instanceGroupName, clusterName)
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
		name := NameForInstanceGroupManager(c.ObjectMeta.Name, ig.ObjectMeta.Name, LastComponent(mig.Zone))
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

var (
	_ cloudinstances.GroupFailureReporter        = (*gceCloudImplementation)(nil)
	_ cloudinstances.GroupInstanceStatusReporter = (*gceCloudImplementation)(nil)
)

// GetGroupFailures returns recent provisioning errors the GCE managed
// instance group has hit. To avoid surfacing errors from a previous, healthier
// incarnation of the group, results are filtered to only errors timestamped
// after the most recent successful instance creation in the group. If the
// group is empty (no instances were ever created), no filter is applied.
func (c *gceCloudImplementation) GetGroupFailures(ctx context.Context, group *cloudinstances.CloudInstanceGroup) ([]cloudinstances.GroupFailure, error) {
	return GetGroupFailures(ctx, c, group)
}

// GetGroupFailures implements the GroupFailureReporter capability for
// any GCECloud. See the method above for the filtering and aggregation rules.
func GetGroupFailures(ctx context.Context, c GCECloud, group *cloudinstances.CloudInstanceGroup) ([]cloudinstances.GroupFailure, error) {
	mig, ok := group.Raw.(*compute.InstanceGroupManager)
	if !ok || mig == nil {
		return nil, nil
	}
	u, err := ParseGoogleCloudURL(mig.SelfLink)
	if err != nil {
		return nil, fmt.Errorf("parsing MIG self link %q: %w", mig.SelfLink, err)
	}

	watermark := cloudinstances.Watermark(group)

	items, err := c.Compute().InstanceGroupManagers().ListErrors(ctx, u.Project, u.Zone, u.Name)
	if err != nil {
		return nil, fmt.Errorf("listing errors for MIG %q: %w", mig.Name, err)
	}

	type key struct{ code, message string }
	agg := map[key]*cloudinstances.GroupFailure{}
	for _, it := range items {
		if it == nil || it.Error == nil {
			continue
		}
		var ts time.Time
		if it.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339, it.Timestamp); err == nil {
				ts = t
			}
		}
		if !watermark.IsZero() && !ts.IsZero() && ts.Before(watermark) {
			continue
		}
		k := key{it.Error.Code, it.Error.Message}
		e, ok := agg[k]
		if !ok {
			e = &cloudinstances.GroupFailure{
				Code:      it.Error.Code,
				Message:   it.Error.Message,
				FirstSeen: ts,
				LastSeen:  ts,
			}
			if it.InstanceActionDetails != nil {
				e.Instance = LastComponent(it.InstanceActionDetails.Instance)
			}
			agg[k] = e
		}
		e.Count++
		if !ts.IsZero() {
			if e.FirstSeen.IsZero() || ts.Before(e.FirstSeen) {
				e.FirstSeen = ts
			}
			if ts.After(e.LastSeen) {
				e.LastSeen = ts
			}
		}
	}

	out := make([]cloudinstances.GroupFailure, 0, len(agg))
	for _, e := range agg {
		out = append(out, *e)
	}
	cloudinstances.SortFailures(out)
	return out, nil
}

// GetGroupInstanceStatuses reports every instance the MIG is managing,
// including ones it is still trying and failing to create. Those never show up
// as CloudInstances: GetCloudGroups skips a managed instance whose compute
// instance does not exist, so a MIG stuck in a create-retry loop is otherwise
// indistinguishable from one that never tried.
func (c *gceCloudImplementation) GetGroupInstanceStatuses(ctx context.Context, group *cloudinstances.CloudInstanceGroup) ([]cloudinstances.GroupInstanceStatus, error) {
	return GetGroupInstanceStatuses(ctx, c, group)
}

// GetGroupInstanceStatuses implements the GroupInstanceStatusReporter
// capability for any GCECloud. See the method above for what it reports.
func GetGroupInstanceStatuses(ctx context.Context, c GCECloud, group *cloudinstances.CloudInstanceGroup) ([]cloudinstances.GroupInstanceStatus, error) {
	mig, ok := group.Raw.(*compute.InstanceGroupManager)
	if !ok || mig == nil {
		return nil, nil
	}

	u, err := ParseGoogleCloudURL(mig.SelfLink)
	if err != nil {
		return nil, fmt.Errorf("parsing MIG self link %q: %w", mig.SelfLink, err)
	}

	// Calling the client directly rather than via ListManagedInstances, which
	// substitutes context.Background() and so cannot be cancelled.
	instances, err := c.Compute().InstanceGroupManagers().ListManagedInstances(ctx, u.Project, u.Zone, u.Name)
	if err != nil {
		return nil, fmt.Errorf("listing managed instances for MIG %q: %w", mig.Name, err)
	}

	out := make([]cloudinstances.GroupInstanceStatus, 0, len(instances))
	for _, i := range instances {
		s := cloudinstances.GroupInstanceStatus{
			Name:          LastComponent(i.Instance),
			Status:        i.InstanceStatus,
			CurrentAction: i.CurrentAction,
		}
		if i.LastAttempt != nil && i.LastAttempt.Errors != nil {
			s.Failures = make([]cloudinstances.GroupFailure, 0, len(i.LastAttempt.Errors.Errors))
			for _, e := range i.LastAttempt.Errors.Errors {
				// Count is left unset: these are the raw errors from one
				// attempt, not aggregated observations.
				s.Failures = append(s.Failures, cloudinstances.GroupFailure{
					Code:    e.Code,
					Message: e.Message,
				})
			}
		}
		out = append(out, s)
	}
	slices.SortFunc(out, func(a, b cloudinstances.GroupInstanceStatus) int {
		return cmp.Compare(a.Name, b.Name)
	})
	return out, nil
}

func addCloudInstanceData(cm *cloudinstances.CloudInstance, instance *compute.Instance) {
	cm.MachineType = LastComponent(instance.MachineType)
	cm.Status = instance.Status
	if instance.CreationTimestamp != "" {
		if t, err := time.Parse(time.RFC3339, instance.CreationTimestamp); err == nil {
			cm.CreationTimestamp = t
		}
	}
	if instance.Status == "RUNNING" {
		cm.State = cloudinstances.CloudInstanceStatusUpToDate
	}
	for k := range instance.Labels {
		if !strings.HasPrefix(k, GceLabelNameRolePrefix) {
			continue
		}
		role := strings.TrimPrefix(k, GceLabelNameRolePrefix)
		// A VM must have one network interface and at most a single AccessConfig on an network interface
		// Also kops doesn't support MultiNics
		cm.PrivateIP = fi.ValueOf(&instance.NetworkInterfaces[0].NetworkIP)
		if len(instance.NetworkInterfaces[0].AccessConfigs) == 1 {
			cm.ExternalIP = fi.ValueOf(&instance.NetworkInterfaces[0].AccessConfigs[0].NatIP)
		}
		if role == "master" || role == "control-plane" {
			cm.Roles = append(cm.Roles, "control-plane")
		} else {
			cm.Roles = append(cm.Roles, role)
		}
	}
}
