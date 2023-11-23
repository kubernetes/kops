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
	"regexp"
	"strconv"

	"github.com/blang/semver/v4"
	"k8s.io/klog/v2"
)

var versionURLPattern = regexp.MustCompile(`/v1\.([\d]+)\.`)

func ParseKubernetesVersion(version string) (*semver.Version, error) {
	sv, err := semver.ParseTolerant(version)
	if err != nil {
		if submatch := versionURLPattern.FindStringSubmatch(version); len(submatch) >= 2 {
			minor, err := strconv.Atoi(submatch[1])
			if err != nil {
				return nil, fmt.Errorf("failed to parse kubernetes version (%s): %w", version, err)
			}
			sv = semver.Version{Major: 1, Minor: uint64(minor)}
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

// Version is our helper type for semver versions.
type Version struct {
	v semver.Version
}

// ParseVersion parses the semver version string into a Version.
func ParseVersion(s string) (*Version, error) {
	v, err := semver.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("error parsing version %q: %w", s, err)
	}
	return &Version{v: v}, nil
}

// String returns a string representation of the object
func (v *Version) String() string {
	return v.v.String()
}

// IsInRange checks if we are in the provided semver range
func (v *Version) IsInRange(semverRange semver.Range) bool {
	return semverRange(v.v)
}
