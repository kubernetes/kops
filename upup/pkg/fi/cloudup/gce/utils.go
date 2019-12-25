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
	"encoding/base32"
	"fmt"
	"hash/fnv"
	"strings"

	"google.golang.org/api/googleapi"
	"k8s.io/klog"
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
// Following requirements (copied from the web UI requirements)
// * 63 chars or less
// * must begin with lower case letter
// * followed by lower case letters/numbers/hyphens
// * may not end in a hyphen (TODO Check this)
func SafeObjectName(name string, clusterName string) string {

	gceName := name + "-" + clusterName
	gceName = strings.ToLower(gceName)
	gceName = LimitedLengthName(gceName, 63)

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
