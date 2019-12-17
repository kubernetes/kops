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

package formatter

import (
	"strings"

	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/model"
)

// InstanceGroupRenderFunction is a render function for an InstanceGroup
type InstanceGroupRenderFunction func(ig *kops.InstanceGroup) string

//RenderInstanceGroupSubnets renders the subnet names for an InstanceGroup
func RenderInstanceGroupSubnets(cluster *kops.Cluster) InstanceGroupRenderFunction {
	return func(ig *kops.InstanceGroup) string {
		return strings.Join(ig.Spec.Subnets, ",")
	}
}

//RenderInstanceGroupZones renders the zone names for an InstanceGroup
func RenderInstanceGroupZones(cluster *kops.Cluster) InstanceGroupRenderFunction {
	return func(ig *kops.InstanceGroup) string {
		zones, err := model.FindZonesForInstanceGroup(cluster, ig)
		if err != nil {
			klog.Warningf("error fetching zones for instancegroup: %v", err)
			return ""
		}
		return strings.Join(zones, ",")
	}
}
