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

package util

import (
	"fmt"
	"strings"

	"github.com/blang/semver"
	"k8s.io/klog"
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
		} else if strings.Contains(v, "/v1.10.") {
			sv = semver.Version{Major: 1, Minor: 10}
		} else if strings.Contains(v, "/v1.11.") {
			sv = semver.Version{Major: 1, Minor: 11}
		} else if strings.Contains(v, "/v1.12.") {
			sv = semver.Version{Major: 1, Minor: 12}
		} else if strings.Contains(v, "/v1.13.") {
			sv = semver.Version{Major: 1, Minor: 13}
		} else if strings.Contains(v, "/v1.14.") {
			sv = semver.Version{Major: 1, Minor: 14}
		} else if strings.Contains(v, "/v1.15.") {
			sv = semver.Version{Major: 1, Minor: 15}
		} else if strings.Contains(v, "/v1.16.") {
			sv = semver.Version{Major: 1, Minor: 16}
		} else if strings.Contains(v, "/v1.17.") {
			sv = semver.Version{Major: 1, Minor: 17}
		} else if strings.Contains(v, "/v1.18.") {
			sv = semver.Version{Major: 1, Minor: 18}
		} else if strings.Contains(v, "/v1.19.") {
			sv = semver.Version{Major: 1, Minor: 19}
		} else if strings.Contains(v, "/v1.20.") {
			sv = semver.Version{Major: 1, Minor: 20}
		} else if strings.Contains(v, "/v1.21.") {
			sv = semver.Version{Major: 1, Minor: 21}
		} else if strings.Contains(v, "/v1.22.") {
			sv = semver.Version{Major: 1, Minor: 22}
		} else if strings.Contains(v, "/v1.23.") {
			sv = semver.Version{Major: 1, Minor: 23}
		} else if strings.Contains(v, "/v1.24.") {
			sv = semver.Version{Major: 1, Minor: 24}
		} else if strings.Contains(v, "/v1.25.") {
			sv = semver.Version{Major: 1, Minor: 25}
		} else if strings.Contains(v, "/v1.26.") {
			sv = semver.Version{Major: 1, Minor: 26}
		} else if strings.Contains(v, "/v1.27.") {
			sv = semver.Version{Major: 1, Minor: 27}
		} else if strings.Contains(v, "/v1.28.") {
			sv = semver.Version{Major: 1, Minor: 28}
		} else if strings.Contains(v, "/v1.29.") {
			sv = semver.Version{Major: 1, Minor: 29}
		} else {
			klog.Errorf("unable to parse Kubernetes version %q", version)
			return nil, fmt.Errorf("unable to parse kubernetes version %q", version)
		}
		klog.V(1).Infof("Kubernetes version %q string matched to %v", version, sv)
	}

	return &sv, nil
}

// TODO: Convert to our own KubernetesVersion type?

func IsKubernetesGTE(version string, k8sVersion semver.Version) bool {
	parsedVersion, err := ParseKubernetesVersion(version)
	if err != nil {
		panic(fmt.Sprintf("Error parsing version %s: %v", version, err))
	}

	// Ignore Pre & Build fields
	k8sVersion.Pre = nil
	k8sVersion.Build = nil

	return k8sVersion.GTE(*parsedVersion)
}
