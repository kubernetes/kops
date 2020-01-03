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

/******************************************************************************
* The Kops Tag Builder
*
* Tags are how we manage kops functionality.
*
******************************************************************************/

package cloudup

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/upup/pkg/fi"
)

func buildCloudupTags(cluster *kopsapi.Cluster) (sets.String, error) {
	tags := sets.NewString()

	switch kopsapi.CloudProviderID(cluster.Spec.CloudProvider) {
	case kopsapi.CloudProviderGCE:
		{
			tags.Insert("_gce")
		}

	case kopsapi.CloudProviderAWS:
		{
			tags.Insert("_aws")
		}
	case kopsapi.CloudProviderDO:
		{
			tags.Insert("_do")
		}
	case kopsapi.CloudProviderVSphere:
		{
			tags.Insert("_vsphere")
		}

	case kopsapi.CloudProviderBareMetal:
		// No tags

	case kopsapi.CloudProviderOpenstack:

	case kopsapi.CloudProviderALI:
		{
			tags.Insert("_ali")
		}
	default:
		return nil, fmt.Errorf("unknown CloudProvider %q", cluster.Spec.CloudProvider)
	}

	versionTag := ""
	if cluster.Spec.KubernetesVersion != "" {
		sv, err := util.ParseKubernetesVersion(cluster.Spec.KubernetesVersion)
		if err != nil {
			return nil, fmt.Errorf("unable to determine kubernetes version from %q", cluster.Spec.KubernetesVersion)
		}

		if sv.Major == 1 && sv.Minor >= 6 {
			versionTag = "_k8s_1_6"
		} else if sv.Major == 1 && sv.Minor == 5 {
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
		tags.Insert(versionTag)
	}

	klog.V(4).Infof("tags: %s", tags.List())

	return tags, nil
}

func buildNodeupTags(role kopsapi.InstanceGroupRole, cluster *kopsapi.Cluster, clusterTags sets.String) (sets.String, error) {
	tags := sets.NewString()

	networking := cluster.Spec.Networking

	if networking == nil {
		return nil, fmt.Errorf("Networking is not set, and should not be nil here")
	}

	if networking.LyftVPC != nil {
		tags.Insert("_lyft_vpc_cni")
	}

	switch fi.StringValue(cluster.Spec.UpdatePolicy) {
	case "": // default
		tags.Insert("_automatic_upgrades")
	case kopsapi.UpdatePolicyExternal:
	// Skip applying the tag
	default:
		klog.Warningf("Unrecognized value for UpdatePolicy: %v", fi.StringValue(cluster.Spec.UpdatePolicy))
	}

	if clusterTags.Has("_gce") {
		tags.Insert("_gce")
	}
	if clusterTags.Has("_aws") {
		tags.Insert("_aws")
	}
	if clusterTags.Has("_do") {
		tags.Insert("_do")
	}

	return tags, nil
}
