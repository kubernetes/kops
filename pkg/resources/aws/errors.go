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

package aws

import (
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

func IsDependencyViolation(err error) bool {
	code := awsup.AWSErrorCode(err)
	switch code {
	case "":
		return false
	case "DependencyViolation", "VolumeInUse", "InvalidIPAddress.InUse":
		return true
	default:
		klog.Infof("unexpected aws error code: %q", code)
		return false
	}
}
