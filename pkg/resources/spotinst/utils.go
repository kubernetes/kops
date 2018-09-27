/*
Copyright 2018 The Kubernetes Authors.

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
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/spotinst/spotinst-sdk-go/service/elastigroup/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/credentials"
	"k8s.io/api/core/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
)

func GuessCloudFromClusterSpec(spec *kops.ClusterSpec) kops.CloudProviderID {
	var cloudProviderID kops.CloudProviderID

	for _, subnet := range spec.Subnets {
		id, known := fi.GuessCloudForZone(subnet.Zone)
		if known {
			glog.V(2).Infof("Inferred cloud=%s from zone %q", id, subnet.Zone)
			cloudProviderID = kops.CloudProviderID(id)
			break
		}
	}

	return cloudProviderID
}

func LoadCredentials() (credentials.Value, error) {
	var (
		chain = newChainCredentials()
		creds credentials.Value
		err   error
	)

	creds, err = chain.Get()
	if err != nil {
		return creds, fmt.Errorf("spotinst: unable to load credentials: %s", err)
	}

	return creds, nil
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

func buildInstanceGroup(ig *kops.InstanceGroup, group *aws.Group, instances []*aws.Instance, nodeMap map[string]*v1.Node) (*cloudinstances.CloudInstanceGroup, error) {
	currentGroupName := spotinst.StringValue(group.Name)
	newGroupName := fmt.Sprintf("%s:%d", spotinst.StringValue(group.Name), time.Now().Nanosecond())

	cg := &cloudinstances.CloudInstanceGroup{
		HumanName:     spotinst.StringValue(group.Name),
		InstanceGroup: ig,
		MinSize:       spotinst.IntValue(group.Capacity.Minimum),
		MaxSize:       spotinst.IntValue(group.Capacity.Maximum),
		Raw:           group,
	}

	for _, instance := range instances {
		instanceID := fi.StringValue(instance.ID)
		if instanceID == "" {
			glog.Warningf("ignoring instance with no instance id: %s", instance)
			continue
		}
		err := cg.NewCloudInstanceGroupMember(instanceID, newGroupName, currentGroupName, nodeMap)
		if err != nil {
			return nil, fmt.Errorf("spotinst: error creating cloud instance group member: %v", err)
		}
	}

	return cg, nil
}
