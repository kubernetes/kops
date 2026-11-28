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
	"sort"

	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/util/pkg/reflectutils"
)

const (
	RoleLabelAPIServer16 = "node-role.kubernetes.io/api-server"
	RoleLabelNode16      = "node-role.kubernetes.io/node"

	// New Experimental control plane roles associated with static manifests
	RoleLabelEtcd                  = "node-role.kubernetes.io/etcd"
	RoleLabelScheduler             = "node-role.kubernetes.io/scheduler"
	RoleLabelKubeControllerManager = "node-role.kubernetes.io/kube-controller-manager"

	// New Experimental control plane roles which are dynamically allocated
	RoleLabelKopsCCM        = "node-role.kops.k8s.io/cloud-controller-manager"
	RoleLabelKopsChannel    = "node-role.kops.k8s.io/kops-channel"
	RoleLabelKopsController = "node-role.kops.k8s.io/kops-controller"
	RoleLabelCertManager    = "node-role.kops.k8s.io/cert-manager"
	RoleLabelCAPIManager    = "node-role.kops.k8s.io/capi-manager"

	RoleLabelControlPlane20 = "node-role.kubernetes.io/control-plane"
)

// ClusterComponentLabel maps each scheduled cluster component to the node label its addon
// selects on.
var ClusterComponentLabel = map[api.ClusterComponent]string{
	api.ClusterComponentCloudControllerManager: RoleLabelKopsCCM,
	api.ClusterComponentKopsController:         RoleLabelKopsController,
	api.ClusterComponentKopsChannel:            RoleLabelKopsChannel,
	api.ClusterComponentCertManager:            RoleLabelCertManager,
	api.ClusterComponentCAPIManager:            RoleLabelCAPIManager,
}

// All returns every node label kOps manages to describe a node's roles. Anything that prunes
// kOps-managed labels should use this, rather than keeping a list of its own that has to be
// remembered when a new role label is added.
func All() []string {
	labels := []string{
		RoleLabelAPIServer16,
		RoleLabelNode16,
		RoleLabelEtcd,
		RoleLabelScheduler,
		RoleLabelKubeControllerManager,
		RoleLabelControlPlane20,
	}
	for _, label := range ClusterComponentLabel {
		labels = append(labels, label)
	}
	sort.Strings(labels)
	return labels
}

// LegacyChannelsNodeLabels returns the labels kops-channels applied to its own node before
// spec.hostedComponents existed: the control-plane label when the cluster has a full
// control-plane instance group, and the per-component kOps labels when the control plane is
// split out.
func LegacyChannelsNodeLabels(cluster *api.Cluster, allInstanceGroups []*api.InstanceGroup) []string {
	for _, ig := range allInstanceGroups {
		if ig.IsControlPlane() {
			return []string{RoleLabelControlPlane20}
		}
	}

	labels := []string{
		RoleLabelKopsCCM,
		RoleLabelKopsChannel,
		RoleLabelKopsController,
		RoleLabelCertManager,
	}
	if cluster.GetCloudProvider() == api.CloudProviderGCE {
		labels = append(labels, RoleLabelCAPIManager)
	}
	return labels
}

// ChannelsNodeLabels returns the labels that kops-channels, running on this instance group,
// should apply to its own node. The components kOps schedules rather than running as static pods
// are placed by selecting on these labels, so this is what decides where they run.
//
// While no instance group in the cluster configures spec.hostedComponents, every instance group
// that runs kops-channels claims every component, as it did before the field existed. Once any
// instance group sets it, each group claims only what it lists.
func ChannelsNodeLabels(cluster *api.Cluster, allInstanceGroups []*api.InstanceGroup, ig *api.InstanceGroup) []string {
	explicit := false
	for _, other := range allInstanceGroups {
		if other.Spec.HostedComponents != nil {
			explicit = true
			break
		}
	}

	if explicit {
		var labels []string
		for _, component := range ig.Spec.HostedComponents {
			if label, ok := ClusterComponentLabel[component]; ok {
				labels = append(labels, label)
			}
		}
		sort.Strings(labels)
		return labels
	}

	// Historically kops-channels ran, and so claimed the components, on every instance group
	// with an API server.
	if !ig.IsControlPlane() && !ig.Spec.Role.HasAPIServer() {
		return nil
	}
	return LegacyChannelsNodeLabels(cluster, allInstanceGroups)
}

// BuildNodeLabels returns the node labels for the specified instance group
// This moved from the kubelet to a central controller in kubernetes 1.16
func BuildNodeLabels(cluster *api.Cluster, instanceGroup *api.InstanceGroup) (map[string]string, error) {
	// A single instance group can carry several roles, so these are independent rather than a
	// first-match dispatch: a group running kube-scheduler alongside its API server must end up
	// labelled for both.
	role := instanceGroup.Spec.Role
	hasControlPlane := role.HasControlPlane()
	hasAPIServer := role.HasAPIServer()
	hasNode := role.HasNode()
	hasEtcd := role.HasEtcd()
	hasScheduler := role.HasScheduler()
	hasKubeControllerManager := role.HasKubeControllerManager()

	// Bastions get no labels, but they are a recognised role.
	if !hasControlPlane && !hasAPIServer && !hasNode && !hasEtcd && !hasScheduler &&
		!hasKubeControllerManager && !role.HasBastion() {
		return nil, fmt.Errorf("unhandled instanceGroup role %q", role)
	}

	// Merge KubeletConfig for NodeLabels
	c := &api.KubeletConfigSpec{}
	if instanceGroup.Spec.Role.IsControlPlaneType() {
		reflectutils.JSONMergeStruct(c, cluster.Spec.ControlPlaneKubelet)
	} else {
		reflectutils.JSONMergeStruct(c, cluster.Spec.Kubelet)
	}

	if instanceGroup.Spec.Kubelet != nil {
		reflectutils.JSONMergeStruct(c, instanceGroup.Spec.Kubelet)
	}

	nodeLabels := c.NodeLabels

	if hasAPIServer || hasControlPlane {
		if nodeLabels == nil {
			nodeLabels = make(map[string]string)
		}
		// Note: featureflag is not available here - we're in kops-controller.
		// We keep the featureflag as a placeholder to change the logic;
		// when we drop the featureflag we should just always include the label, even for
		// full control-plane nodes.
		if hasAPIServer && featureflag.APIServerNodes.Enabled() {
			nodeLabels[RoleLabelAPIServer16] = ""
			nodeLabels["kops.k8s.io/kops-controller-pki"] = ""
		}
	}

	if hasNode {
		if nodeLabels == nil {
			nodeLabels = make(map[string]string)
		}
		nodeLabels[RoleLabelNode16] = ""
	}

	if hasEtcd {
		if nodeLabels == nil {
			nodeLabels = make(map[string]string)
		}
		nodeLabels[RoleLabelEtcd] = ""
	}

	if hasScheduler {
		if nodeLabels == nil {
			nodeLabels = make(map[string]string)
		}
		nodeLabels[RoleLabelScheduler] = ""
	}

	if hasKubeControllerManager {
		if nodeLabels == nil {
			nodeLabels = make(map[string]string)
		}
		nodeLabels[RoleLabelKubeControllerManager] = ""
	}

	if hasControlPlane {
		if nodeLabels == nil {
			nodeLabels = make(map[string]string)
		}
		for label, value := range BuildMandatoryControlPlaneLabels(make(map[string]string)) {
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
func BuildMandatoryControlPlaneLabels(nodeLabels map[string]string) map[string]string {
	nodeLabels[RoleLabelControlPlane20] = ""
	nodeLabels["kops.k8s.io/kops-controller-pki"] = ""
	nodeLabels["node.kubernetes.io/exclude-from-external-load-balancers"] = ""
	return nodeLabels
}
