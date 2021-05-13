/*
Copyright 2020 The Kubernetes Authors.

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

package mirrors

import (
	"fmt"
	"strings"

	"k8s.io/kops"
)

const (
	// defaultKopsMirrorBase will be detected and automatically set to pull from the defaultKopsMirrors
	defaultKopsMirrorBase = "https://artifacts.k8s.io/binaries/kops/%s/"
	githubKopsMirrorBase  = "https://github.com/kubernetes/kops/releases/download/v%s/"
)

func FindUrlMirrors(u string) []string {
	// Use the canonical URL as the first mirror
	mirrors := []string{u}

	// Use the mirrors to also find hashes.
	baseURLString := fmt.Sprintf(defaultKopsMirrorBase, kops.Version)
	if !strings.HasSuffix(baseURLString, "/") {
		baseURLString += "/"
	}

	// Use mirrors when the URL is not a custom one
	if strings.HasPrefix(u, baseURLString) {
		suffix := strings.TrimPrefix(u, baseURLString)
		// GitHub artifact names are quite different, because the suffix path is collapsed.
		githubSuffix := strings.ReplaceAll(suffix, "/", "-")
		githubSuffix = strings.ReplaceAll(githubSuffix, "linux-amd64-nodeup", "nodeup-linux-amd64")
		githubSuffix = strings.ReplaceAll(githubSuffix, "linux-arm64-nodeup", "nodeup-linux-arm64")
		githubSuffix = strings.ReplaceAll(githubSuffix, "linux-amd64-protokube", "protokube-linux-amd64")
		githubSuffix = strings.ReplaceAll(githubSuffix, "linux-arm64-protokube", "protokube-linux-arm64")
		githubSuffix = strings.ReplaceAll(githubSuffix, "linux-amd64-channels", "channels-linux-amd64")
		githubSuffix = strings.ReplaceAll(githubSuffix, "linux-arm64-channels", "channels-linux-arm64")
		mirrors = append(mirrors, fmt.Sprintf(githubKopsMirrorBase, kops.Version)+githubSuffix)
	}

	return mirrors
}
