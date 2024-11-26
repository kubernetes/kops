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

package model

import (
	"fmt"
	"strings"

	"github.com/blang/semver/v4"
	"k8s.io/kops/pkg/apis/kops/util"
)

// KubernetesVersion is a wrapper over semver functionality,
// that offers some functionality particularly useful for kubernetes version semantics.
type KubernetesVersion struct {
	versionString string
	version       semver.Version
}

// ParseKubernetesVersion parses a Kubernetes version string and returns a KubernetesVersion object.
func ParseKubernetesVersion(versionString string) (*KubernetesVersion, error) {
	parsedVersion, err := util.ParseKubernetesVersion(versionString)
	if err != nil {
		return nil, fmt.Errorf("error parsing version %q: %v", versionString, err)
	}

	return &KubernetesVersion{versionString: versionString, version: *parsedVersion}, nil
}

// MustParseKubernetesVersion parses a Kubernetes version string and panics if it fails.
func MustParseKubernetesVersion(versionString string) *KubernetesVersion {
	kubernetesVersion, err := ParseKubernetesVersion(versionString)
	if err != nil || kubernetesVersion == nil {
		panic(err)
	}
	return kubernetesVersion
}

func (v KubernetesVersion) String() string {
	return v.versionString
}

// IsBaseURL checks if the version string is a URL, rather than a version identifier.
// URLs are typically used for CI builds and during development.
func IsBaseURL(kubernetesVersion string) bool {
	return strings.HasPrefix(kubernetesVersion, "http:") || strings.HasPrefix(kubernetesVersion, "https:") || strings.HasPrefix(kubernetesVersion, "memfs:")
}

// IsBaseURL checks if the version string is a URL, rather than a version identifier.
// URLs are typically used for CI builds and during development.
func (v KubernetesVersion) IsBaseURL() bool {
	return IsBaseURL(v.versionString)
}

// IsGTE checks if the version is greater than or equal (>=) to the specified version.
// It panics if the kubernetes version in the cluster is invalid, or if the version is invalid.
func (v KubernetesVersion) IsGTE(version string) bool {
	parsedVersion, err := util.ParseKubernetesVersion(version)
	if err != nil || parsedVersion == nil {
		panic(fmt.Sprintf("error parsing version %q: %v", version, err))
	}

	// Ignore Pre & Build fields
	clusterVersion := v.version
	clusterVersion.Pre = nil
	clusterVersion.Build = nil

	return clusterVersion.GTE(*parsedVersion)
}

// IsLT checks if the version is strictly less (<) than the specified version.
// It panics if the kubernetes version in the cluster is invalid, or if the version is invalid.
func (v KubernetesVersion) IsLT(version string) bool {
	return !v.IsGTE(version)
}

// Minor returns the minor version of the Kubernetes version.
func (v KubernetesVersion) Minor() int {
	return int(v.version.Minor)
}
