/*
Copyright 2024 The Kubernetes Authors.

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
	"errors"

	"google.golang.org/api/googleapi"
	"k8s.io/klog/v2"
)

// AsGoogleAPIError returns a googleapi.Error in the error chain, or false if none is found.
func AsGoogleAPIError(err error) (*googleapi.Error, bool) {
	var googleAPIError *googleapi.Error
	if errors.As(err, &googleAPIError) {
		return googleAPIError, true
	} else {
		return nil, false
	}
}

// IsDependencyViolation checks if the error is a dependency violation.
func IsDependencyViolation(err error) bool {
	apiErr, ok := AsGoogleAPIError(err)
	if !ok {
		return false
	}

	for _, e := range apiErr.Errors {
		switch e.Reason {
		case "resourceInUseByAnotherResource":
			return true
		default:
			klog.Infof("unexpected gce error code: %+v", e)
		}
	}

	return false
}
