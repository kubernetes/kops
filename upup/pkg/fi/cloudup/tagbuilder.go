package cloudup

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi"
)

func buildClusterTags(cluster *api.Cluster) (map[string]struct{}, error) {
	// TODO: Make these configurable?
	useMasterASG := true
	useMasterLB := false

	tags := make(map[string]struct{})

	//tags["_networking_kubenet"] = struct{}{}
	//tags["_networking_builtin"] = struct{}{}

	if useMasterASG {
		tags["_master_asg"] = struct{}{}
	} else {
		tags["_master_single"] = struct{}{}
	}

	if useMasterLB {
		tags["_master_lb"] = struct{}{}
	} else {
		tags["_not_master_lb"] = struct{}{}
	}

	if cluster.Spec.MasterPublicName != "" {
		tags["_master_dns"] = struct{}{}
	}

	if fi.BoolValue(cluster.Spec.IsolateMasters) {
		tags["_isolate_masters"] = struct{}{}
	}

	switch cluster.Spec.CloudProvider {
	case "gce":
		{
			glog.Fatalf("GCE is (probably) not working currently - please ping @justinsb for cleanup")
			tags["_gce"] = struct{}{}
		}

	case "aws":
		{
			tags["_aws"] = struct{}{}
		}

	default:
		return nil, fmt.Errorf("unknown CloudProvider %q", cluster.Spec.CloudProvider)
	}

	return tags, nil
}

func buildNodeupTags(role api.InstanceGroupRole, cluster *api.Cluster, clusterTags map[string]struct{}) ([]string, error) {
	var tags []string

	switch role {
	case api.InstanceGroupRoleNode:
		// No special tags

		// TODO: Should we run _protokube on the nodes?
		tags = append(tags, "_protokube")

	case api.InstanceGroupRoleMaster:
		if !fi.BoolValue(cluster.Spec.IsolateMasters) {
			// Run this master as a pool node also (start kube-proxy etc)
			tags = append(tags, "_kubernetes_pool")
		}
		tags = append(tags, "_protokube")
	default:
		return nil, fmt.Errorf("Unrecognized role: %v", role)
	}

	//// TODO: Replace with list of CNI plugins
	//if _, found := clusterTags["_networking_kubenet"]; found {
	//	tags = append(tags, "_cni_bridge")
	//	tags = append(tags, "_cni_host_local")
	//	tags = append(tags, "_cni_loopback")
	//	tags = append(tags, "_cni_ptp")
	//	//tags = append(tags, "_cni_tuning")
	//}

	if _, found := clusterTags["_gce"]; found {
		tags = append(tags, "_gce")
	}
	if _, found := clusterTags["_aws"]; found {
		tags = append(tags, "_aws")
	}

	return tags, nil
}
