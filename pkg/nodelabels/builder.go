package nodelabels

import (
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/util/pkg/reflectutils"
)

const (
	RoleLabelName15        = "kubernetes.io/role"
	RoleLabelName16        = "kubernetes.io/role"
	RoleMasterLabelValue15 = "master"
	RoleNodeLabelValue15   = "node"

	RoleLabelMaster16 = "node-role.kubernetes.io/master"
	RoleLabelNode16   = "node-role.kubernetes.io/node"
)

// BuildNodeLabels returns the node labels for the specified instance group
// This moved from the kubelet to a central controller in kubernetes 1.16
func BuildNodeLabels(cluster *kops.Cluster, instanceGroup *kops.InstanceGroup) (map[string]string, error) {
	isMaster := instanceGroup.Spec.Role == kops.InstanceGroupRoleMaster

	// Merge KubeletConfig for NodeLabels
	c := &kops.KubeletConfigSpec{}
	if isMaster {
		reflectutils.JsonMergeStruct(c, cluster.Spec.MasterKubelet)
	} else {
		reflectutils.JsonMergeStruct(c, cluster.Spec.Kubelet)
	}

	if instanceGroup.Spec.Kubelet != nil {
		reflectutils.JsonMergeStruct(c, instanceGroup.Spec.Kubelet)
	}

	nodeLabels := c.NodeLabels

	if isMaster {
		if nodeLabels == nil {
			nodeLabels = make(map[string]string)
		}
		nodeLabels[RoleLabelMaster16] = ""
		nodeLabels[RoleLabelName15] = RoleMasterLabelValue15
	} else {
		if nodeLabels == nil {
			nodeLabels = make(map[string]string)
		}
		nodeLabels[RoleLabelNode16] = ""
		nodeLabels[RoleLabelName15] = RoleNodeLabelValue15
	}

	for k, v := range instanceGroup.Spec.NodeLabels {
		if nodeLabels == nil {
			nodeLabels = make(map[string]string)
		}
		nodeLabels[k] = v
	}

	return nodeLabels, nil
}
