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

package k8sversion

import (
	"github.com/blang/semver"

	"k8s.io/kops/pkg/apis/kops/util"
)

// KubernetesVersion holds a semver-version of kubernetes
type KubernetesVersion struct {
	semver semver.Version
}

// Parse parses the string to determine the KubernetesVersion.
// The version may be a semver version, or it may be a URL with the kubernetes version in the path
func Parse(version string) (*KubernetesVersion, error) {
	sv, err := util.ParseKubernetesVersion(version)
	if err != nil {
		return nil, err
	}

	return &KubernetesVersion{semver: *sv}, nil
}

// IsGTE checks if the version is greater than or equal to the passed version.  Pre and Build fields are ignored.
// Panic if version is not valid, so version should only be used with static strings like "1.10"
func (k *KubernetesVersion) IsGTE(version string) bool {
	return util.IsKubernetesGTE(version, k.semver)
}

// String returns a string representation of the semver, like 1.10.1.  It does not include a leading 'v'
func (k *KubernetesVersion) String() string {
	return k.semver.String()
}
