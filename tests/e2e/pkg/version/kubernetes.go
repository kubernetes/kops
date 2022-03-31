/*
Copyright 2021 The Kubernetes Authors.

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

package version

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	"github.com/blang/semver/v4"
	"k8s.io/kops/tests/e2e/pkg/util"
)

// ParseKubernetesVersion will parse the provided k8s version
// Either a semver or marker URL is accepted
func ParseKubernetesVersion(version string) (string, error) {
	if _, err := semver.ParseTolerant(version); err == nil {
		return version, nil
	}
	if u, err := url.Parse(version); err == nil {
		var b bytes.Buffer
		err = util.HTTPGETWithHeaders(version, nil, &b)
		if err != nil {
			return "", err
		}

		// Replace the last part of the version URL path with the contents of the URL's body
		// Example:
		// https://storage.googleapis.com/k8s-release-dev/ci/latest.txt -> v1.21.0-beta.1.112+576aa2d2470b28%0A
		// becomes https://storage.googleapis.com/k8s-release-dev/ci/v1.21.0-beta.1.112+576aa2d2470b28%0A
		pathParts := strings.Split(u.Path, "/")
		pathParts[len(pathParts)-1] = strings.TrimSpace(b.String())
		u.Path = strings.Join(pathParts, "/")
		return strings.TrimSpace(u.String()), nil
	}
	return "", fmt.Errorf("unexpected kubernetes version: %v", version)
}
