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
	defaultKopsMirrorBase    = "https://kubeupv2.s3.amazonaws.com/kops/%s/"
	githubKopsMirrorBase     = "https://github.com/kubernetes/kops/releases/download/v%s/"
	kubernetesKopsMirrorBase = "https://artifacts.k8s.io/binaries/kops/%s/"
)

func FindUrlMirrors(u string) []string {
	// Use the mirrors to also find hashes.
	baseURLString := fmt.Sprintf(defaultKopsMirrorBase, kops.Version)
	if !strings.HasSuffix(baseURLString, "/") {
		baseURLString += "/"
	}

	var mirrors []string
	if strings.HasPrefix(u, baseURLString) {
		suffix := strings.TrimPrefix(u, baseURLString)
		// artifacts.k8s.io is the official and preferred mirror.
		mirrors = append(mirrors, fmt.Sprintf(kubernetesKopsMirrorBase, kops.Version)+suffix)
		// GitHub artifact names are quite different, because the suffix path is collapsed.
		githubSuffix := strings.ReplaceAll(suffix, "/", "-")
		githubSuffix = strings.ReplaceAll(githubSuffix, "linux-amd64-nodeup", "nodeup-linux-amd64")
		githubSuffix = strings.ReplaceAll(githubSuffix, "linux-arm64-nodeup", "nodeup-linux-arm64")
		mirrors = append(mirrors, fmt.Sprintf(githubKopsMirrorBase, kops.Version)+githubSuffix)
	}
	// Finally append the original URL to the list of mirrored URLs.
	// In case this is a custom URL, there won't be any mirrors before it,
	// otherwise it will be the last mirror URL, because it's now a legacy location.
	mirrors = append(mirrors, u)

	return mirrors
}
