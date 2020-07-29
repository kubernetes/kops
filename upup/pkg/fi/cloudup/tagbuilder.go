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
	api "k8s.io/kops/pkg/apis/kops"
)

func buildCloudupTags(cluster *api.Cluster) (sets.String, error) {
	tags := sets.NewString()

	switch api.CloudProviderID(cluster.Spec.CloudProvider) {
	case api.CloudProviderGCE:
		{
			tags.Insert("_gce")
		}

	case api.CloudProviderAWS:
		{
			tags.Insert("_aws")
		}
	case api.CloudProviderDO:
		{
			tags.Insert("_do")
		}

	case api.CloudProviderOpenstack:

	case api.CloudProviderALI:
		{
			tags.Insert("_ali")
		}
	default:
		return nil, fmt.Errorf("unknown CloudProvider %q", cluster.Spec.CloudProvider)
	}

	tags.Insert("_k8s_1_6")

	klog.V(4).Infof("tags: %s", tags.List())

	return tags, nil
}
