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

package spotinst

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
)

// ListGroups returns a list of all Elastigroups as Resource objects.
func ListGroups(svc Service, clusterName string) ([]*resources.Resource, error) {
	glog.V(2).Info("Listing all Elastigroups")

	groups, err := svc.List(context.Background())
	if err != nil {
		return nil, err
	}

	var resourceTrackers []*resources.Resource
	for _, group := range groups {
		if strings.HasSuffix(group.Name(), clusterName) {
			resource := &resources.Resource{
				ID:      group.Id(),
				Name:    group.Name(),
				Obj:     group,
				Deleter: deleter(svc, group),
				Dumper:  dumper,
			}
			resourceTrackers = append(resourceTrackers, resource)
		}
	}

	return resourceTrackers, nil
}

// DeleteGroup deletes an existing Elastigroup.
func DeleteGroup(svc Service, group *cloudinstances.CloudInstanceGroup) error {
	glog.V(2).Infof("Deleting Elastigroup %q", group.HumanName)

	return svc.Delete(
		context.Background(),
		group.Raw.(Elastigroup).Id())
}

// DeleteInstance removes an instance from its Elastigroup.
func DeleteInstance(svc Service, instance *cloudinstances.CloudInstanceGroupMember) error {
	glog.V(2).Infof("Detaching instance %q from Elastigroup", instance.ID)

	return svc.Detach(
		context.Background(),
		instance.CloudInstanceGroup.Raw.(Elastigroup).Id(),
		[]string{instance.ID})
}

// GetCloudGroups returns a list of Elastigroups as CloudInstanceGroup objects.
func GetCloudGroups(svc Service, cluster *kops.Cluster, instancegroups []*kops.InstanceGroup,
	warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	glog.V(2).Info("Listing all Elastigroups")

	groups, err := svc.List(context.Background())
	if err != nil {
		return nil, err
	}

	instanceGroups := make(map[string]*cloudinstances.CloudInstanceGroup)
	nodeMap := cloudinstances.GetNodeMap(nodes, cluster)

	for _, group := range groups {
		// Find matching instance group.
		var instancegroup *kops.InstanceGroup
		for _, ig := range instancegroups {
			name := getGroupNameByRole(cluster, ig)
			if name == "" {
				continue
			}
			if name == group.Name() {
				if instancegroup != nil {
					return nil, fmt.Errorf("spotinst: found multiple instance groups matching group %q", group.Name())
				}
				instancegroup = ig
			}
		}

		if instancegroup == nil {
			if warnUnmatched {
				glog.Warningf("Found group with no corresponding instance group %q", group.Name())
			}
			continue
		}

		// Build the instance group.
		ig, err := buildCloudInstanceGroup(svc, instancegroup, group, nodeMap)
		if err != nil {
			return nil, fmt.Errorf("spotinst: failed to build instance group: %v", err)
		}

		instanceGroups[instancegroup.ObjectMeta.Name] = ig
	}

	return instanceGroups, nil
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
		glog.Warningf("Ignoring InstanceGroup of unknown role %q", ig.Spec.Role)
	}

	return groupName
}

func buildCloudInstanceGroup(svc Service, ig *kops.InstanceGroup, group Elastigroup,
	nodeMap map[string]*v1.Node) (*cloudinstances.CloudInstanceGroup, error) {

	instances, err := svc.Instances(context.Background(), group.Id())
	if err != nil {
		return nil, err
	}

	instanceGroup := &cloudinstances.CloudInstanceGroup{
		HumanName:     group.Name(),
		InstanceGroup: ig,
		MinSize:       group.MinSize(),
		MaxSize:       group.MaxSize(),
		Raw:           group,
	}

	currentName := group.Name()
	newName := fmt.Sprintf("%s:%d", group.Name(), time.Now().Nanosecond())

	for _, instance := range instances {
		if instance.Id() == "" {
			glog.Warningf("Ignoring instance with no ID: %v", instance)
			continue
		}

		if err := instanceGroup.NewCloudInstanceGroupMember(
			instance.Id(), currentName, newName, nodeMap); err != nil {
			return nil, fmt.Errorf("spotinst: error creating cloud instance group member: %v", err)
		}
	}

	return instanceGroup, nil
}

func deleter(svc Service, group Elastigroup) func(fi.Cloud, *resources.Resource) error {
	return func(cloud fi.Cloud, resource *resources.Resource) error {
		glog.V(2).Infof("Deleting Elastigroup %q", group.Id())
		return svc.Delete(context.Background(), group.Id())
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
