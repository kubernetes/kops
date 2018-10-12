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

package spotinst

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/spotinst/spotinst-sdk-go/service/elastigroup/providers/aws"
	"k8s.io/api/core/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

type AWSCloud interface {
	awsup.AWSCloud

	Elastigroup() aws.Service
}

type awsCloud struct {
	awsup.AWSCloud
	svc aws.Service
}

var _ AWSCloud = &awsCloud{}

func (c *awsCloud) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderSpotinst
}

func (c *awsCloud) DeleteInstance(instance *cloudinstances.CloudInstanceGroupMember) error {
	instanceID := instance.ID
	if instanceID == "" {
		return fmt.Errorf("spotinst: unexpected instance id: %v", instanceID)
	}

	var nodeName string
	if instance.Node != nil {
		nodeName = instance.Node.Name
	}

	var groupID string
	if instance.CloudInstanceGroup != nil {
		groupID = fi.StringValue(instance.CloudInstanceGroup.Raw.(*aws.Group).ID)
	}

	glog.V(2).Infof("Stopping instance %q, node %q, in group %q", instanceID, nodeName, groupID)
	input := &aws.DetachGroupInput{
		GroupID:                       fi.String(groupID),
		InstanceIDs:                   []string{instanceID},
		ShouldDecrementTargetCapacity: fi.Bool(false),
		ShouldTerminateInstances:      fi.Bool(true),
	}
	if _, err := c.svc.Detach(context.Background(), input); err != nil {
		if nodeName != "" {
			return fmt.Errorf("spotinst: failed to delete instance %q, node %q: %v", instanceID, nodeName, err)
		}
		return fmt.Errorf("spotinst: failed to delete instance %q: %v", instanceID, err)
	}

	return nil
}

func (c *awsCloud) DeleteGroup(group *cloudinstances.CloudInstanceGroup) error {
	groupID := fi.StringValue(group.Raw.(*aws.Group).ID)
	input := &aws.DeleteGroupInput{
		GroupID: fi.String(groupID),
	}
	_, err := c.svc.Delete(context.Background(), input)
	if err != nil {
		return fmt.Errorf("spotinst: failed to delete group %q: %v", groupID, err)
	}
	return nil
}

func (c *awsCloud) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	nodeMap := cloudinstances.GetNodeMap(nodes, cluster)

	resources, err := listResourcesSpotinst(c, cluster.Name)
	if err != nil {
		return nil, fmt.Errorf("spotinst: failed to list resources: %v", err)
	}

	for _, resource := range resources {
		group, ok := resource.Obj.(*aws.Group)
		if !ok {
			continue
		}

		var instancegroup *kops.InstanceGroup
		for _, ig := range instancegroups {
			name := getGroupNameByRole(cluster, ig)
			if name == "" {
				continue
			}
			if name == resource.Name {
				if instancegroup != nil {
					return nil, fmt.Errorf("spotinst: found multiple instance groups matching group %q", name)
				}
				instancegroup = ig
			}
		}
		if instancegroup == nil {
			if warnUnmatched {
				glog.Warningf("Found group with no corresponding instance group %q", resource.Name)
			}
			continue
		}
		input := &aws.StatusGroupInput{
			GroupID: group.ID,
		}
		output, err := c.svc.Status(context.Background(), input)
		if err != nil {
			return nil, err
		}
		ig, err := buildInstanceGroup(instancegroup, group, output.Instances, nodeMap)
		if err != nil {
			return nil, fmt.Errorf("spotinst: failed to build instance group: %v", err)
		}
		groups[instancegroup.ObjectMeta.Name] = ig
	}

	return groups, nil
}

func (c *awsCloud) Elastigroup() aws.Service {
	return c.svc
}
