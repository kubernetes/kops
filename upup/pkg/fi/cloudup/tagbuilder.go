package cloudup

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi"
)

// TODO: Rename to buildCloudupTags ?
func buildClusterTags(cluster *api.Cluster) (map[string]struct{}, error) {
	// TODO: Make these configurable?
	useMasterASG := true
	useMasterLB := false

	tags := make(map[string]struct{})

	networking := cluster.Spec.Networking
	if networking == nil || networking.Classic != nil {
		tags["_networking_classic"] = struct{}{}
	} else if networking.Kubenet != nil {
		tags["_networking_kubenet"] = struct{}{}
	} else if networking.External != nil {
		// external is based on kubenet
		tags["_networking_kubenet"] = struct{}{}
		tags["_networking_external"] = struct{}{}
	} else {
		return nil, fmt.Errorf("No networking mode set")
	}

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
		tags = append(tags, "_kubernetes_pool")

		// TODO: Should we run _protokube on the nodes?
		tags = append(tags, "_protokube")

	case api.InstanceGroupRoleMaster:
		tags = append(tags, "_kubernetes_master")

		if !fi.BoolValue(cluster.Spec.IsolateMasters) {
			// Run this master as a pool node also (start kube-proxy etc)
			tags = append(tags, "_kubernetes_pool")
		}

		tags = append(tags, "_protokube")
	default:
		return nil, fmt.Errorf("Unrecognized role: %v", role)
	}

	// TODO: Replace with list of CNI plugins ?
	if usesCNI(cluster) {
		tags = append(tags, "_cni_bridge")
		tags = append(tags, "_cni_host_local")
		tags = append(tags, "_cni_loopback")
		tags = append(tags, "_cni_ptp")
		//tags = append(tags, "_cni_tuning")
	}

	switch fi.StringValue(cluster.Spec.UpdatePolicy) {
	case "": // default
		tags = append(tags, "_automatic_upgrades")
	case api.UpdatePolicyExternal:
	// Skip applying the tag
	default:
		glog.Warningf("Unrecognized value for UpdatePolicy: %v", fi.StringValue(cluster.Spec.UpdatePolicy))
	}

	if _, found := clusterTags["_gce"]; found {
		tags = append(tags, "_gce")
	}
	if _, found := clusterTags["_aws"]; found {
		tags = append(tags, "_aws")
	}

	return tags, nil
}
