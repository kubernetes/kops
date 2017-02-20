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

package gce

import (
	"google.golang.org/api/googleapi"
	"strings"
)

func IsNotFound(err error) bool {
	apiErr, ok := err.(*googleapi.Error)
	if !ok {
		return false
	}

	// We could also check for Errors[].Resource == "notFound"
	//glog.Info("apiErr: %v", apiErr)

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
