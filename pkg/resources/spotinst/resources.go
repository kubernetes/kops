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

package spotinst

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
)

// ListResources returns a list of all resources.
func ListResources(cloud Cloud, clusterName string) ([]*resources.Resource, error) {
	klog.V(2).Info("Listing all resources")

	fns := []func(Cloud, string) ([]*resources.Resource, error){
		ListElastigroupResources,
		ListOceanResources,
	}

	var resourceTrackers []*resources.Resource
	for _, fn := range fns {
		resources, err := fn(cloud, clusterName)
		if err != nil {
			return nil, fmt.Errorf("spotinst: error listing resources: %v", err)
		}

		resourceTrackers = append(resourceTrackers, resources...)
	}

	return resourceTrackers, nil
}

// ListElastigroupResources returns a list of all Elastigroup resources.
func ListElastigroupResources(cloud Cloud, clusterName string) ([]*resources.Resource, error) {
	klog.V(2).Info("Listing all Elastigroup resources")

	// List all Elastigroup instance groups.
	groups, err := listInstanceGroups(cloud.Elastigroup(), clusterName)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

// ListOceanResources returns a list of all Ocean resources.
func ListOceanResources(cloud Cloud, clusterName string) ([]*resources.Resource, error) {
	klog.V(2).Info("Listing all Ocean resources")
	var resourceTrackers []*resources.Resource

	// List all Ocean instance groups.
	oceans, err := listInstanceGroups(cloud.Ocean(), clusterName)
	if err != nil {
		return nil, err
	}
	resourceTrackers = append(resourceTrackers, oceans...)

	// List all Ocean launch specs.
	for _, ocean := range oceans {
		specs, err := listLaunchSpecs(cloud.LaunchSpec(), ocean.ID)
		if err != nil {
			return nil, err
		}
		resourceTrackers = append(resourceTrackers, specs...)
	}

	return resourceTrackers, nil
}

// listInstanceGroups returns a list of all instance groups.
func listInstanceGroups(svc InstanceGroupService, clusterName string) ([]*resources.Resource, error) {
	groups, err := svc.List(context.Background())
	if err != nil {
		return nil, err
	}

	var resourceTrackers []*resources.Resource
	for _, group := range groups {
		if strings.HasSuffix(group.Name(), clusterName) &&
			!strings.HasPrefix(strings.ToLower(group.Name()), "spotinst::ocean::") {
			resource := &resources.Resource{
				ID:      group.Id(),
				Name:    group.Name(),
				Type:    string(ResourceTypeInstanceGroup),
				Obj:     group,
				Deleter: instanceGroupDeleter(svc, group),
				Dumper:  dumper,
			}
			resourceTrackers = append(resourceTrackers, resource)
		}
	}

	return resourceTrackers, nil
}

// listLaunchSpecs returns a list of all launch specs.
func listLaunchSpecs(svc LaunchSpecService, oceanID string) ([]*resources.Resource, error) {
	specs, err := svc.List(context.Background(), oceanID)
	if err != nil {
		return nil, err
	}

	var resourceTrackers []*resources.Resource
	for _, spec := range specs {
		resource := &resources.Resource{
			ID:      spec.Id(),
			Name:    spec.Name(),
			Type:    string(ResourceTypeLaunchSpec),
			Obj:     spec,
			Deleter: launchSpecDeleter(svc, spec),
			Dumper:  dumper,
		}
		resourceTrackers = append(resourceTrackers, resource)
	}

	return resourceTrackers, nil
}

// DeleteInstanceGroup deletes an existing InstanceGroup.
func DeleteInstanceGroup(cloud Cloud, group *cloudinstances.CloudInstanceGroup) error {
	klog.V(2).Infof("Deleting instance group: %q", group.HumanName)

	switch obj := group.Raw.(type) {
	case InstanceGroup:
		{
			var svc InstanceGroupService
			switch obj.Type() {
			case InstanceGroupElastigroup:
				svc = cloud.Elastigroup()
			case InstanceGroupOcean:
				svc = cloud.Ocean()
			}

			return svc.Delete(context.Background(), obj.Id())
		}
	case LaunchSpec:
		{
			return cloud.LaunchSpec().Delete(context.Background(), obj.Id())
		}
	}

	return fmt.Errorf("spotinst: unexpected instance group type, got: %T", group.Raw)
}

// DeleteInstance removes an instance from its instance group.
func DeleteInstance(cloud Cloud, instance *cloudinstances.CloudInstanceGroupMember) error {
	klog.V(2).Infof("Detaching instance %q from instance group: %q",
		instance.ID, instance.CloudInstanceGroup.HumanName)

	group := instance.CloudInstanceGroup
	switch obj := group.Raw.(type) {
	case InstanceGroup:
		{
			var svc InstanceGroupService
			switch obj.Type() {
			case InstanceGroupElastigroup:
				svc = cloud.Elastigroup()
			case InstanceGroupOcean:
				svc = cloud.Ocean()
			}

			return svc.Detach(context.Background(), obj.Id(), []string{instance.ID})
		}
	case LaunchSpec:
		{
			return cloud.Ocean().Detach(context.Background(), obj.OceanId(), []string{instance.ID})
		}
	}

	return fmt.Errorf("spotinst: unexpected instance group type, got: %T", group.Raw)
}

// GetCloudGroups returns a list of InstanceGroups as CloudInstanceGroup objects.
func GetCloudGroups(cloud Cloud, cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup,
	warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {

	cloudInstanceGroups := make(map[string]*cloudinstances.CloudInstanceGroup)
	nodeMap := cloudinstances.GetNodeMap(nodes, cluster)

	// List all resources.
	resources, err := ListResources(cloud, cluster.Name)
	if err != nil {
		return nil, err
	}

	// Build all cloud instance groups.
	for _, resource := range resources {

		// Filter out the Ocean resources (they're not needed for now since
		// we fetch the instances from the launch specs).
		if ResourceType(resource.Type) == ResourceTypeInstanceGroup {
			if resource.Obj.(InstanceGroup).Type() == InstanceGroupOcean {
				continue
			}
		}

		// Build cloud instance group.
		ig, err := buildCloudInstanceGroupFromResource(cloud, cluster, instanceGroups, resource, nodeMap)
		if err != nil {
			return nil, fmt.Errorf("spotinst: error building cloud instance group: %v", err)
		}
		if ig == nil {
			if warnUnmatched {
				klog.V(2).Infof("Found group with no corresponding instance group: %q", resource.Name)
			}
			continue
		}

		cloudInstanceGroups[resource.Name] = ig
	}

	return cloudInstanceGroups, nil
}

func buildCloudInstanceGroupFromResource(cloud Cloud, cluster *kops.Cluster,
	instanceGroups []*kops.InstanceGroup, resource *resources.Resource,
	nodeMap map[string]*v1.Node) (*cloudinstances.CloudInstanceGroup, error) {

	// Find corresponding instance group.
	ig, err := findInstanceGroupFromResource(cluster, instanceGroups, resource)
	if err != nil {
		return nil, fmt.Errorf("failed to find instance group of resource %q: %v", resource.Name, err)
	}
	if ig == nil {
		return nil, nil
	}

	switch ResourceType(resource.Type) {
	case ResourceTypeInstanceGroup:
		{
			if group, ok := resource.Obj.(InstanceGroup); ok {
				return buildCloudInstanceGroupFromInstanceGroup(cloud, ig, group, nodeMap)
			}
		}

	case ResourceTypeLaunchSpec:
		{
			if spec, ok := resource.Obj.(LaunchSpec); ok {
				return buildCloudInstanceGroupFromLaunchSpec(cloud, ig, spec, nodeMap)
			}
		}
	}

	return nil, fmt.Errorf("unexpected resource type: %s", resource.Type)
}

func buildCloudInstanceGroupFromInstanceGroup(cloud Cloud, ig *kops.InstanceGroup, group InstanceGroup,
	nodeMap map[string]*v1.Node) (*cloudinstances.CloudInstanceGroup, error) {

	instanceGroup := &cloudinstances.CloudInstanceGroup{
		HumanName:     group.Name(),
		InstanceGroup: ig,
		MinSize:       group.MinSize(),
		MaxSize:       group.MaxSize(),
		Raw:           group,
	}

	var svc InstanceGroupService
	switch group.Type() {
	case InstanceGroupElastigroup:
		svc = cloud.Elastigroup()
	case InstanceGroupOcean:
		svc = cloud.Ocean()
	}

	klog.V(2).Infof("Attempting to fetch all instances of instance group: %q (id: %q)", group.Name(), group.Id())
	instances, err := svc.Instances(context.Background(), group.Id())
	if err != nil {
		return nil, err
	}

	// Register all instances as group members.
	if err := registerCloudInstanceGroupMembers(instanceGroup, nodeMap,
		instances, group.Name(), group.UpdatedAt()); err != nil {
		return nil, err
	}

	return instanceGroup, nil
}

// TODO(liran): We should fetch Ocean's instances using a query param of `?launchSpecId=foo`,
// but, since we do not support it at the moment, we should fetch all instances only once.
var fetchOceanInstances sync.Once

func buildCloudInstanceGroupFromLaunchSpec(cloud Cloud, ig *kops.InstanceGroup, spec LaunchSpec,
	nodeMap map[string]*v1.Node) (*cloudinstances.CloudInstanceGroup, error) {

	instanceGroup := &cloudinstances.CloudInstanceGroup{
		HumanName:     spec.Name(),
		InstanceGroup: ig,
		Raw:           spec,
	}

	var instances []Instance
	var err error

	fetchOceanInstances.Do(func() {
		klog.V(2).Infof("Attempting to fetch all instances of instance group: %q (id: %q)", spec.Name(), spec.Id())
		instances, err = cloud.Ocean().Instances(context.Background(), spec.OceanId())
	})
	if err != nil {
		return nil, err
	}

	// Register all instances as group members.
	if err := registerCloudInstanceGroupMembers(instanceGroup, nodeMap,
		instances, spec.Name(), spec.UpdatedAt()); err != nil {
		return nil, err
	}

	return instanceGroup, nil
}

func registerCloudInstanceGroupMembers(instanceGroup *cloudinstances.CloudInstanceGroup, nodeMap map[string]*v1.Node,
	instances []Instance, currentInstanceGroupName string, instanceGroupUpdatedAt time.Time) error {

	// The instance registration below registers all active instances with
	// their instance group. In addition, it looks for outdated instances by
	// comparing each instance creation timestamp against the modification
	// timestamp of its instance group.
	//
	// In a rolling-update operation, one or more detach operations are
	// performed to replace existing instances. This is done by updating the
	// instance group and results in updating the modification timestamp to the
	// current time.
	//
	// The update of the modification timestamp occurs only after the detach
	// operation is completed, meaning that new instances have already been
	// created, so our comparison may be incorrect.
	//
	// In order to work around this issue, we assume that the detach operation
	// will take up to two minutes, and therefore we subtract this duration from
	// the modification timestamp of the instance group before the comparison.
	instanceGroupUpdatedAt = instanceGroupUpdatedAt.Add(-2 * time.Minute)

	for _, instance := range instances {
		if instance.Id() == "" {
			klog.Warningf("Ignoring instance with no ID: %v", instance)
			continue
		}

		// If the instance was created before the last update, mark it as `NeedUpdate`.
		newInstanceGroupName := currentInstanceGroupName
		if instance.CreatedAt().Before(instanceGroupUpdatedAt) {
			newInstanceGroupName = fmt.Sprintf("%s:%d", currentInstanceGroupName, time.Now().Nanosecond())
		}

		klog.V(2).Infof("Adding instance %q (created at: %s) to instance group: %q (updated at: %s)",
			instance.Id(), instance.CreatedAt().Format(time.RFC3339),
			currentInstanceGroupName, instanceGroupUpdatedAt.Format(time.RFC3339))

		if err := instanceGroup.NewCloudInstanceGroupMember(
			instance.Id(), newInstanceGroupName, currentInstanceGroupName, nodeMap); err != nil {
			return fmt.Errorf("error creating cloud instance group member: %v", err)
		}
	}

	return nil
}

func findInstanceGroupFromResource(cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup,
	resource *resources.Resource) (*kops.InstanceGroup, error) {

	var instanceGroup *kops.InstanceGroup
	for _, ig := range instanceGroups {
		name := getGroupNameByRole(cluster, ig)
		if name == "" {
			continue
		}

		if name == resource.Name {
			if instanceGroup != nil {
				return nil, fmt.Errorf("found multiple instance groups matching group: %q", name)
			}

			klog.V(2).Infof("Found group with corresponding instance group: %q", name)
			instanceGroup = ig
		}
	}

	return instanceGroup, nil
}

func getGroupNameByRole(cluster *kops.Cluster, ig *kops.InstanceGroup) string {
	var groupName string

	switch ig.Spec.Role {
	case kops.InstanceGroupRoleMaster:
		groupName = ig.ObjectMeta.Name + ".masters." + cluster.ObjectMeta.Name
	case kops.InstanceGroupRoleNode:
		groupName = ig.ObjectMeta.Name + "." + cluster.ObjectMeta.Name
	case kops.InstanceGroupRoleBastion:
		groupName = ig.ObjectMeta.Name + "." + cluster.ObjectMeta.Name
	default:
		klog.Warningf("Ignoring InstanceGroup of unknown role %q", ig.Spec.Role)
	}

	return groupName
}

func instanceGroupDeleter(svc InstanceGroupService, group InstanceGroup) func(fi.Cloud, *resources.Resource) error {
	return func(cloud fi.Cloud, resource *resources.Resource) error {
		klog.V(2).Infof("Deleting instance group: %q", group.Id())
		return svc.Delete(context.Background(), group.Id())
	}
}

func launchSpecDeleter(svc LaunchSpecService, spec LaunchSpec) func(fi.Cloud, *resources.Resource) error {
	return func(cloud fi.Cloud, resource *resources.Resource) error {
		klog.V(2).Infof("Deleting launch spec: %q", spec.Id())
		return svc.Delete(context.Background(), spec.Id())
	}
}

func dumper(op *resources.DumpOperation, resource *resources.Resource) error {
	data := make(map[string]interface{})

	data["id"] = resource.ID
	data["type"] = resource.Type
	data["raw"] = resource.Obj

	op.Dump.Resources = append(op.Dump.Resources, data)
	return nil
}
