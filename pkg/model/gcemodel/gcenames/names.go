/*
Copyright 2023 The Kubernetes Authors.

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
	"encoding/base32"
	"hash/fnv"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/truncate"
)

func NameForTargetPool(id string, clusterName string) string {
	return safeSuffixedObjectName(id, clusterName)
}

func NameForHealthCheck(id string, clusterName string) string {
	return SafeObjectName(id, clusterName)
}

func NameForBackendService(id string, clusterName string) string {
	return SafeObjectName(id, clusterName)
}

func NameForForwardingRule(id string, clusterName string) string {
	return safeSuffixedObjectName(id, clusterName)
}

func NameForIPAddress(id string, clusterName string) string {
	return safeSuffixedObjectName(id, clusterName)
}

func NameForPoolHealthcheck(id string, clusterName string) string {
	return SafeObjectName(id, clusterName)
}

func NameForHealthcheck(id string, clusterName string) string {
	return safeSuffixedObjectName(id, clusterName)
}

func NameForRouter(id string, clusterName string) string {
	return safeSuffixedObjectName(id, clusterName)
}

func NameForFirewallRule(id string, clusterName string) string {
	return ClusterSuffixedName(id, clusterName, 63)
}

// NameForIPAliasRange returns the name for the secondary IP range attached to a subnet
func NameForIPAliasRange(key string, clusterName string) string {
	// We include the cluster name so we could share a subnet...
	// but there's a 5 IP alias range limit per subnet anwyay, so
	// this is rather pointless and in practice we just use a
	// separate subnet per cluster
	return SafeObjectName(key, clusterName)
}

// ClusterSuffixedName returns a cluster-suffixed name, with a maxLength
func ClusterSuffixedName(objectName string, clusterName string, maxLength int) string {
	prefix := objectName + "-"
	suffixLength := maxLength - len(prefix)
	if suffixLength < 10 {
		klog.Fatalf("cannot construct a reasonable object name of length %d with a prefix of length %d (%q)", maxLength, len(prefix), prefix)
	}

	// GCE does not support . in tags / names
	safeClusterName := strings.Replace(clusterName, ".", "-", -1)

	opt := truncate.TruncateStringOptions{
		MaxLength:     suffixLength,
		AlwaysAddHash: false,
		HashLength:    6,
	}
	suffix := truncate.TruncateString(safeClusterName, opt)

	return prefix + suffix
}

// NameForInstanceGroupManager builds a name for an InstanceGroupManager in the specified zone
func NameForInstanceGroupManager(c *kops.Cluster, ig *kops.InstanceGroup, zone string) string {
	shortZone := zone
	lastDash := strings.LastIndex(shortZone, "-")
	if lastDash != -1 {
		shortZone = shortZone[lastDash+1:]
	}
	name := SafeObjectName(shortZone+"."+ig.ObjectMeta.Name, c.ObjectMeta.Name)
	name = LimitedLengthName(name, 63)
	return name
}

// ServiceAccountName returns the cluster-suffixed service-account name
func ServiceAccountName(name string, clusterName string) string {
	return ClusterSuffixedName(name, clusterName, 30)
}

// safeSuffixedObjectName returns the object name and cluster name escaped for GCE, limited to 63 chars
func safeSuffixedObjectName(name string, clusterName string) string {
	return ClusterSuffixedName(name, clusterName, 63)
}

// SafeObjectName returns the object name and cluster name escaped for GCE
func SafeObjectName(name string, clusterName string) string {
	gceName := name + "-" + clusterName

	// TODO: If the cluster name > some max size (32?) we should curtail it
	return SafeClusterName(gceName)
}

// SafeClusterName returns a safe cluster name
// deprecated: prefer ClusterSuffixedName
func SafeClusterName(clusterName string) string {
	// GCE does not support . in tags / names
	safeClusterName := strings.Replace(clusterName, ".", "-", -1)
	return safeClusterName
}

// LimitedLengthName returns a string subject to a maximum length
func LimitedLengthName(s string, n int) string {
	// We only use the hash if we need to
	if len(s) <= n {
		return s
	}

	h := fnv.New32a()
	if _, err := h.Write([]byte(s)); err != nil {
		klog.Fatalf("error hashing values: %v", err)
	}
	hashString := base32.HexEncoding.EncodeToString(h.Sum(nil))
	hashString = strings.ToLower(hashString)
	if len(hashString) > 6 {
		hashString = hashString[:6]
	}

	maxBaseLength := n - len(hashString) - 1
	if len(s) > maxBaseLength {
		s = s[:maxBaseLength]
	}
	s = s + "-" + hashString

	return s
}
