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

func (c *DeleteInstanceGroup) DeleteInstanceGroup(group *api.InstanceGroup) error {
	groups, err := FindCloudInstanceGroups(c.Cloud, c.Cluster, []*api.InstanceGroup{group}, false, nil)
	cig := groups[group.ObjectMeta.Name]
	if cig == nil {
		return fmt.Errorf("InstanceGroup not found in cloud")
	}
	if len(groups) != 1 {
		return fmt.Errorf("Multiple InstanceGroup resources found in cloud")
	}

	err = cig.Delete(c.Cloud)
	if err != nil {
		return fmt.Errorf("error deleting cloud resources for InstanceGroup: %v", err)
	}

	err = c.Clientset.InstanceGroups(c.Cluster.ObjectMeta.Name).Delete(group.ObjectMeta.Name, nil)
	if err != nil {
		return err
	}

	return nil
}
