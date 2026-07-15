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
	"fmt"

	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/util/pkg/reflectutils"
)

const (
	RoleLabelAPIServer16           = "node-role.kubernetes.io/api-server"
	RoleLabelNode16                = "node-role.kubernetes.io/node"
	RoleLabelEtcd                  = "node-role.kubernetes.io/etcd"
	RoleLabelScheduler             = "node-role.kubernetes.io/scheduler"
	RoleLabelKubeControllerManager = "node-role.kubernetes.io/kube-controller-manager"

	RoleLabelControlPlane20 = "node-role.kubernetes.io/control-plane"
)

// BuildNodeLabels returns the node labels for the specified instance group
// This moved from the kubelet to a central controller in kubernetes 1.16
func BuildNodeLabels(cluster *api.Cluster, instanceGroup *api.InstanceGroup) (map[string]string, error) {
	isControlPlane := false
	isAPIServer := false
	isNode := false
	isEtcd := false
	isScheduler := false
	isKubeControllerManager := false
	switch {
	case instanceGroup.Spec.Role.HasControlPlane():
		isControlPlane = true
	case instanceGroup.Spec.Role.HasAPIServer():
		isAPIServer = true
	case instanceGroup.Spec.Role.HasNode():
		isNode = true
	case instanceGroup.Spec.Role.HasBastion():
		// no labels to add
	case instanceGroup.Spec.Role.HasEtcd():
		isEtcd = true
	case instanceGroup.Spec.Role.HasScheduler():
		isScheduler = true
	case instanceGroup.Spec.Role.HasKubeControllerManager():
		isKubeControllerManager = true
	default:
		return nil, fmt.Errorf("unhandled instanceGroup role %q", instanceGroup.Spec.Role)
	}

	// Merge KubeletConfig for NodeLabels
	c := &api.KubeletConfigSpec{}
	if isControlPlane {
		reflectutils.JSONMergeStruct(c, cluster.Spec.ControlPlaneKubelet)
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
			nodeLabels["kops.k8s.io/kops-controller-pki"] = ""
		}
	}

	if isNode {
		if nodeLabels == nil {
			nodeLabels = make(map[string]string)
		}
		nodeLabels[RoleLabelNode16] = ""
	}

	if isEtcd {
		if nodeLabels == nil {
			nodeLabels = make(map[string]string)
		}
		nodeLabels[RoleLabelEtcd] = ""
	}

	if isScheduler {
		if nodeLabels == nil {
			nodeLabels = make(map[string]string)
		}
		nodeLabels[RoleLabelScheduler] = ""
	}

	if isKubeControllerManager {
		if nodeLabels == nil {
			nodeLabels = make(map[string]string)
		}
		nodeLabels[RoleLabelKubeControllerManager] = ""
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

	return nodeLabels, nil
}

// BuildMandatoryControlPlaneLabels returns the list of labels all CP nodes must have
func BuildMandatoryControlPlaneLabels() map[string]string {
	nodeLabels := make(map[string]string)
	nodeLabels[RoleLabelControlPlane20] = ""
	nodeLabels["kops.k8s.io/kops-controller-pki"] = ""
	nodeLabels["node.kubernetes.io/exclude-from-external-load-balancers"] = ""
	return nodeLabels
}
