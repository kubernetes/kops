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

package awsup

import (
	"fmt"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
)

// matchInstanceGroup filters a list of instancegroups for recognized cloud groups
func matchInstanceGroup(name string, clusterName string, instancegroups []*kops.InstanceGroup) (*kops.InstanceGroup, error) {
	var instancegroup *kops.InstanceGroup
	for _, g := range instancegroups {
		var groupName string
		switch {
		case g.Spec.Role.HasControlPlane():
			groupName = g.ObjectMeta.Name + ".masters." + clusterName
		case g.Spec.Role.HasAPIServer():
			groupName = g.ObjectMeta.Name + ".apiservers." + clusterName
		case g.Spec.Role.HasEtcd():
			groupName = g.ObjectMeta.Name + ".etcd." + clusterName
		case g.Spec.Role.HasScheduler():
			groupName = g.ObjectMeta.Name + ".scheduler." + clusterName
		case g.Spec.Role.HasCloudControllerManager():
			groupName = g.ObjectMeta.Name + ".ccm." + clusterName
		case g.Spec.Role.HasKubeControllerManager():
			groupName = g.ObjectMeta.Name + ".kcm." + clusterName
		case g.Spec.Role.HasNode():
			groupName = g.ObjectMeta.Name + "." + clusterName
		case g.Spec.Role.HasBastion():
			groupName = g.ObjectMeta.Name + "." + clusterName
		default:
			klog.Warningf("Ignoring InstanceGroup of unknown role %q", g.Spec.Role)
			continue
		}

		if name == groupName {
			if instancegroup != nil {
				return nil, fmt.Errorf("found multiple instance groups matching ASG %q", groupName)
			}
			instancegroup = g
		}
	}

	return instancegroup, nil
}
