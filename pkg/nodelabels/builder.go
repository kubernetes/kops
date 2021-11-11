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

package nodelabels

import (
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/util/pkg/reflectutils"
)

const (
	RoleLabelName15           = "kubernetes.io/role"
	RoleMasterLabelValue15    = "master"
	RoleAPIServerLabelValue15 = "api-server"
	RoleNodeLabelValue15      = "node"

	RoleLabelMaster16    = "node-role.kubernetes.io/master"
	RoleLabelAPIServer16 = "node-role.kubernetes.io/api-server"
	RoleLabelNode16      = "node-role.kubernetes.io/node"

	RoleLabelControlPlane20 = "node-role.kubernetes.io/control-plane"
)

// BuildNodeLabels returns the node labels for the specified instance group
// This moved from the kubelet to a central controller in kubernetes 1.16
func BuildNodeLabels(cluster *kops.Cluster, instanceGroup *kops.InstanceGroup) map[string]string {
	isControlPlane := instanceGroup.Spec.Role == kops.InstanceGroupRoleMaster

	isAPIServer := instanceGroup.Spec.Role == kops.InstanceGroupRoleAPIServer

	// Merge KubeletConfig for NodeLabels
	c := &kops.KubeletConfigSpec{}
	if isControlPlane {
		reflectutils.JSONMergeStruct(c, cluster.Spec.MasterKubelet)
	} else {
		reflectutils.JSONMergeStruct(c, cluster.Spec.Kubelet)
	}

	if instanceGroup.Spec.Kubelet != nil {
		reflectutils.JSONMergeStruct(c, instanceGroup.Spec.Kubelet)
	}

	nodeLabels := c.NodeLabels

	if isAPIServer || isControlPlane {
		if nodeLabels == nil {
			nodeLabels = make(map[string]string)
		}
		// Note: featureflag is not available here - we're in kops-controller.
		// We keep the featureflag as a placeholder to change the logic;
		// when we drop the featureflag we should just always include the label, even for
		// full control-plane nodes.
		if isAPIServer || featureflag.APIServerNodes.Enabled() {
			nodeLabels[RoleLabelAPIServer16] = ""
		}
		nodeLabels[RoleLabelName15] = RoleAPIServerLabelValue15
	} else {
		if nodeLabels == nil {
			nodeLabels = make(map[string]string)
		}
		nodeLabels[RoleLabelNode16] = ""
		nodeLabels[RoleLabelName15] = RoleNodeLabelValue15
	}

	if isControlPlane {
		if nodeLabels == nil {
			nodeLabels = make(map[string]string)
		}
		for label, value := range BuildMandatoryControlPlaneLabels() {
			nodeLabels[label] = value
		}
	}

	for k, v := range instanceGroup.Spec.NodeLabels {
		if nodeLabels == nil {
			nodeLabels = make(map[string]string)
		}
		nodeLabels[k] = v
	}

	return nodeLabels
}

// BuildMandatoryControlPlaneLabels returns the list of labels all CP nodes must have
func BuildMandatoryControlPlaneLabels() map[string]string {
	nodeLabels := make(map[string]string)
	nodeLabels[RoleLabelMaster16] = ""
	nodeLabels[RoleLabelControlPlane20] = ""
	nodeLabels[RoleLabelName15] = RoleMasterLabelValue15
	nodeLabels["kops.k8s.io/kops-controller-pki"] = ""
	nodeLabels["node.kubernetes.io/exclude-from-external-load-balancers"] = ""
	return nodeLabels
}
