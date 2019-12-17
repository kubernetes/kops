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
)

func IsNotFound(err error) bool {
	apiErr, ok := err.(*googleapi.Error)
	if !ok {
		return false
	}

	// We could also check for Errors[].Resource == "notFound"
	//klog.Info("apiErr: %v", apiErr)

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

func SafeClusterName(clusterName string) string {
	// GCE does not support . in tags / names
	safeClusterName := strings.Replace(clusterName, ".", "-", -1)
	return safeClusterName
}

// SafeObjectName returns the object name and cluster name escaped for GCE
func SafeObjectName(name string, clusterName string) string {
	gceName := name + "-" + clusterName

	// TODO: If the cluster name > some max size (32?) we should curtail it
	return SafeClusterName(gceName)
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

// ZoneToRegion maps a GCE zone name to a GCE region name, returning an error if it cannot be mapped
func ZoneToRegion(zone string) (string, error) {
	tokens := strings.Split(zone, "-")
	if len(tokens) <= 2 {
		return "", fmt.Errorf("invalid GCE Zone: %v", zone)
	}
	region := tokens[0] + "-" + tokens[1]
	return region, nil
}
