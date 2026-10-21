/*
Copyright The Kubernetes Authors.

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

package gcenames

import (
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/truncate"
)

const (
	GceLabelNameRolePrefix = "k8s-io-role-"
)

// SafeClusterName returns a safe cluster name
// Deprecated: prefer ClusterSuffixedName
func SafeClusterName(clusterName string) string {
	// GCE does not support . in tags / names
	safeClusterName := strings.ReplaceAll(clusterName, ".", "-")
	return safeClusterName
}

// ClusterPrefixedName returns a cluster-prefixed name, with a maxLength
func ClusterPrefixedName(objectName string, clusterName string, maxLength int) string {
	suffix := "-" + objectName
	prefixLength := maxLength - len(suffix)
	if prefixLength < 10 {
		klog.Fatalf("cannot construct a reasonable object name of length %d with a suffix of length %d (%q)", maxLength, len(suffix), suffix)
	}

	// GCE does not support . in tags / names
	safeClusterName := strings.ReplaceAll(clusterName, ".", "-")

	opt := truncate.TruncateStringOptions{
		MaxLength:     prefixLength,
		AlwaysAddHash: false,
		HashLength:    6,
	}
	prefix := truncate.TruncateString(safeClusterName, opt)

	return prefix + suffix
}

// TagForRole return the instance (network) tag used for instances with the given role.
func TagForRole(clusterName string, role kops.InstanceGroupRole) string {
	return ClusterPrefixedName(GceLabelNameRolePrefix+role.ToLowerString(), clusterName, 63)
}

// UsesIPAliases checks if the cluster uses IP aliases for network connectivity
func UsesIPAliases(c *kops.Cluster) bool {
	// "GCP" networking mode is called "ip-alias" or "vpc-native" on GKE.
	if c.Spec.Networking.GCP != nil {
		return true
	}

	if c.Spec.Networking.Kindnet != nil {
		// TODO: Are we _always_ using ipalias - should we at least check the cloud is GCP?
		return true
	}

	return false
}
