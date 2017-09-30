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

package instancegroups

import (
	"fmt"

	"github.com/golang/glog"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/upup/pkg/fi"
)

// DeleteInstanceGroup removes the cloud resources for an InstanceGroup
type DeleteInstanceGroup struct {
	Cluster   *api.Cluster
	Cloud     fi.Cloud
	Clientset simple.Clientset
}

// DeleteInstanceGroup deletes a cloud instance group
func (d *DeleteInstanceGroup) DeleteInstanceGroup(group *api.InstanceGroup) error {

	groups, err := d.Cloud.GetCloudGroups(d.Cluster, []*api.InstanceGroup{group}, false, nil)
	if err != nil {
		return fmt.Errorf("error finding CloudInstanceGroups: %v", err)
	}

	// TODO should we drain nodes and validate the cluster?
	cig := groups[group.ObjectMeta.Name]
	if cig == nil {
		glog.Warningf("AutoScalingGroup %q not found in cloud - skipping delete", group.ObjectMeta.Name)
	} else {
		if len(groups) != 1 {
			return fmt.Errorf("multiple InstanceGroup resources found in cloud")
		}

		glog.Infof("Deleting AutoScalingGroup %q", group.ObjectMeta.Name)

		err = d.Cloud.DeleteGroup(cig.GroupName, cig.GroupTemplateName)
		if err != nil {
			return fmt.Errorf("error deleting cloud resources for InstanceGroup: %v", err)
		}
	}

	err = d.Clientset.InstanceGroupsFor(d.Cluster).Delete(group.ObjectMeta.Name, nil)
	if err != nil {
		return err
	}

	return nil
}
