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

/******************************************************************************
* The Kops Tag Builder
*
* Tags are how we manage kops functionality.
*
******************************************************************************/

package cloudup

import (
	"fmt"
	"github.com/golang/glog"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

//
//
func buildCloudupTags(cluster *api.Cluster) (map[string]struct{}, error) {
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
	} else if networking.CNI != nil {
		// external is based on cni, weave, flannel, etc
		tags["_networking_cni"] = struct{}{}
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
	} else if cluster.Spec.Topology.Masters == api.TopologyPublic {
		tags["_not_master_lb"] = struct{}{}
	}

	// Network Topologies
	if cluster.Spec.Topology == nil {
		return nil, fmt.Errorf("missing topology spec")
	}
	if cluster.Spec.Topology.Masters == api.TopologyPublic && cluster.Spec.Topology.Nodes == api.TopologyPublic {
		tags["_topology_public"] = struct{}{}
	} else if cluster.Spec.Topology.Masters == api.TopologyPrivate && cluster.Spec.Topology.Nodes == api.TopologyPrivate {
		tags["_topology_private"] = struct{}{}
	} else {
		return nil, fmt.Errorf("Unable to parse topology. Unsupported topology configuration. Masters and nodes must match!")
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

	versionTag := ""
	if cluster.Spec.KubernetesVersion != "" {
		sv, err := api.ParseKubernetesVersion(cluster.Spec.KubernetesVersion)
		if err != nil {
			return nil, fmt.Errorf("unable to determine kubernetes version from %q", cluster.Spec.KubernetesVersion)
		}

		if sv.Major == 1 && sv.Minor >= 5 {
			versionTag = "_k8s_1_5"
		} else if sv.Major == 1 && sv.Minor == 4 {
			versionTag = "_k8s_1_4"
		} else {
			// We don't differentiate between these older versions
			versionTag = "_k8s_1_3"
		}
	}
	if versionTag == "" {
		return nil, fmt.Errorf("unable to determine kubernetes version from %q", cluster.Spec.KubernetesVersion)
	} else {
		tags[versionTag] = struct{}{}
	}

	return tags, nil
}

func buildNodeupTags(role api.InstanceGroupRole, cluster *api.Cluster, clusterTags map[string]struct{}) ([]string, error) {
	var tags []string

	networking := cluster.Spec.Networking

	if networking == nil {
		return nil, fmt.Errorf("Networking is not set, and should not be nil here")
	}

	if networking.CNI != nil {
		// external is based on cni, weave, flannel, etc
		tags = append(tags, "_networking_cni")
	}

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
