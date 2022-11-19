/*
Copyright 2021 The Kubernetes Authors.

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

package gcemetadata

import (
	"strings"

	"google.golang.org/api/compute/v1"
	"k8s.io/kops/upup/pkg/fi"
)

// MetadataKeyClusterName is the key used for the metadata that specifies the cluster name.
const MetadataKeyClusterName = "cluster-name"

// MetadataMatchesClusterName checks if the metadata has the specified cluster-name included.
func MetadataMatchesClusterName(findClusterName string, metadata *compute.Metadata) bool {
	if metadata == nil {
		return false
	}
	for _, item := range metadata.Items {
		if item.Key == MetadataKeyClusterName {
			value := fi.ValueOf(item.Value)
			if strings.TrimSpace(value) == findClusterName {
				return true
			} else {
				return false
			}
		}
	}
	return false
}

// InstanceMatchesClusterName checks if the instances has the specified cluster-name included.
func InstanceMatchesClusterName(findClusterName string, instance *compute.Instance) bool {
	return MetadataMatchesClusterName(findClusterName, instance.Metadata)
}
