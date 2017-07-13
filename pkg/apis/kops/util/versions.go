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

package util

import (
	"fmt"
	"strings"

	"github.com/blang/semver"
	"github.com/golang/glog"
)

func ParseKubernetesVersion(version string) (*semver.Version, error) {
	sv, err := semver.ParseTolerant(version)
	if err != nil {
		v := strings.Trim(version, "v")
		if strings.HasPrefix(v, "1.3.") {
			sv = semver.Version{Major: 1, Minor: 3}
		} else if strings.HasPrefix(v, "1.4.") {
			sv = semver.Version{Major: 1, Minor: 4}
		} else if strings.HasPrefix(v, "1.5.") {
			sv = semver.Version{Major: 1, Minor: 5}
		} else if strings.HasPrefix(v, "1.6.") {
			sv = semver.Version{Major: 1, Minor: 6}
		} else if strings.HasPrefix(v, "1.7.") {
			sv = semver.Version{Major: 1, Minor: 7}
		} else if strings.Contains(v, "/v1.3.") {
			sv = semver.Version{Major: 1, Minor: 3}
		} else if strings.Contains(v, "/v1.4.") {
			sv = semver.Version{Major: 1, Minor: 4}
		} else if strings.Contains(v, "/v1.5.") {
			sv = semver.Version{Major: 1, Minor: 5}
		} else if strings.Contains(v, "/v1.6.") {
			sv = semver.Version{Major: 1, Minor: 6}
		} else if strings.Contains(v, "/v1.7.") {
			sv = semver.Version{Major: 1, Minor: 7}
		} else if strings.Contains(v, "/v1.8.") {
			sv = semver.Version{Major: 1, Minor: 8}
		} else if strings.Contains(v, "/v1.9.") {
			sv = semver.Version{Major: 1, Minor: 9}
		} else {
			glog.Errorf("unable to parse Kubernetes version %q", version)
			return nil, fmt.Errorf("unable to parse kubernetes version %q", version)
		}
		glog.V(1).Infof("Kubernetes version %q string matched to %v", version, sv)
	}

	return &sv, nil
}

// TODO: Convert to our own KubernetesVersion type?

func IsKubernetesGTE(version string, k8sVersion semver.Version) bool {
	// The string-arg is a little annoying, but simplifies the calling code!
	switch version {
	case "1.2":
		return k8sVersion.Major > 1 || (k8sVersion.Major == 1 && k8sVersion.Minor >= 2)
	case "1.3":
		return k8sVersion.Major > 1 || (k8sVersion.Major == 1 && k8sVersion.Minor >= 3)
	case "1.4":
		return k8sVersion.Major > 1 || (k8sVersion.Major == 1 && k8sVersion.Minor >= 4)
	case "1.5":
		return k8sVersion.Major > 1 || (k8sVersion.Major == 1 && k8sVersion.Minor >= 5)
	case "1.6":
		return k8sVersion.Major > 1 || (k8sVersion.Major == 1 && k8sVersion.Minor >= 6)
	case "1.7":
		return k8sVersion.Major > 1 || (k8sVersion.Major == 1 && k8sVersion.Minor >= 7)
	default:
		panic(fmt.Sprintf("IsKubernetesGTE not supported with version %q", version))
	}
}
