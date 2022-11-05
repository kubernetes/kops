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

package gce

import (
	"fmt"
	"strings"

	"google.golang.org/api/googleapi"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/truncate"
)

func IsNotFound(err error) bool {
	apiErr, ok := err.(*googleapi.Error)
	if !ok {
		return false
	}

	// We could also check for Errors[].Resource == "notFound"

	return apiErr.Code == 404
}

func IsNotReady(err error) bool {
	apiErr, ok := err.(*googleapi.Error)
	if !ok {
		return false
	}
	for _, e := range apiErr.Errors {
		if e.Reason == "resourceNotReady" {
			return true
		}
	}
	return false
}

// ClusterPrefixedName returns a cluster-prefixed name, with a maxLength
func ClusterPrefixedName(objectName string, clusterName string, maxLength int) string {
	suffix := "-" + objectName
	prefixLength := maxLength - len(suffix)
	if prefixLength < 10 {
		klog.Fatalf("cannot construct a reasonable object name of length %d with a suffix of length %d (%q)", maxLength, len(suffix), suffix)
	}

	// GCP does not support . in tags / names
	safeClusterName := strings.Replace(clusterName, ".", "-", -1)

	opt := truncate.TruncateStringOptions{
		MaxLength:     prefixLength,
		AlwaysAddHash: false,
		HashLength:    6,
	}
	prefix := truncate.TruncateString(safeClusterName, opt)

	return prefix + suffix
}

// ClusterSuffixedName returns a cluster-suffixed name, with a maxLength
func ClusterSuffixedName(objectName string, clusterName string, maxLength int) string {
	prefix := objectName + "-"
	suffixLength := maxLength - len(prefix)
	if suffixLength < 10 {
		klog.Fatalf("cannot construct a reasonable object name of length %d with a prefix of length %d (%q)", maxLength, len(prefix), prefix)
	}

	// GCP does not support . in tags / names
	safeClusterName := strings.Replace(clusterName, ".", "-", -1)

	opt := truncate.TruncateStringOptions{
		MaxLength:     suffixLength,
		AlwaysAddHash: false,
		HashLength:    6,
	}
	suffix := truncate.TruncateString(safeClusterName, opt)

	return prefix + suffix
}

// SafeClusterName returns a safe cluster name
// deprecated: prefer ClusterSuffixedName
func SafeClusterName(clusterName string) string {
	// GCP does not support . in tags / names
	safeClusterName := strings.Replace(clusterName, ".", "-", -1)
	return safeClusterName
}

// SafeTruncatedClusterName returns a safe and truncated cluster name
func SafeTruncatedClusterName(clusterName string, maxLength int) string {
	// GCP does not support . in tags / names
	safeClusterName := strings.Replace(clusterName, ".", "-", -1)

	opt := truncate.TruncateStringOptions{
		MaxLength:     maxLength,
		AlwaysAddHash: false,
		HashLength:    6,
	}
	truncatedClusterName := truncate.TruncateString(safeClusterName, opt)

	return truncatedClusterName
}

// SafeObjectName returns the object name and cluster name escaped for GCP
func SafeObjectName(name string, clusterName string) string {
	gceName := name + "-" + clusterName

	// TODO: If the cluster name > some max size (32?) we should curtail it
	return SafeClusterName(gceName)
}

// ServiceAccountName returns the cluster-suffixed service-account name
func ServiceAccountName(name string, clusterName string) string {
	return ClusterSuffixedName(name, clusterName, 30)
}

// LastComponent returns the last component of a URL, i.e. anything after the last slash
// If there is no slash, returns the whole string
func LastComponent(s string) string {
	lastSlash := strings.LastIndex(s, "/")
	if lastSlash != -1 {
		s = s[lastSlash+1:]
	}
	return s
}

// ZoneToRegion maps a GCP zone name to a GCP region name, returning an error if it cannot be mapped
func ZoneToRegion(zone string) (string, error) {
	tokens := strings.Split(zone, "-")
	if len(tokens) <= 2 {
		return "", fmt.Errorf("invalid GCP Zone: %v", zone)
	}
	region := tokens[0] + "-" + tokens[1]
	return region, nil
}
