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

package cloudup

import (
	"github.com/golang/glog"
	api "k8s.io/kops/pkg/apis/kops"
)

func usesCNI(c *api.Cluster) bool {
	networkConfig := c.Spec.Networking
	if networkConfig == nil || networkConfig.Classic != nil {
		// classic
		return false
	}

	if networkConfig.Kubenet != nil {
		// kubenet
		return true
	}

	if networkConfig.External != nil {
		// external: assume uses CNI
		return true
	}

	if networkConfig.CNI != nil {
		// external: assume uses CNI
		return true
	}

	// Assume other modes also use CNI
	glog.Warningf("Unknown networking mode configured")
	return true
}
