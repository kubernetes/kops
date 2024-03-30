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

package assets

import (
	"strings"

	"k8s.io/kops"
)

const (
	// defaultKopsMirrorBase will be detected and automatically set to pull from the defaultKopsMirrors
	kopsDefaultBase      = "https://artifacts.k8s.io/binaries/kops/%s/"
	githubKopsMirrorBase = "https://github.com/kubernetes/kops/releases/download/v%s/"
)

type mirrorConfig struct {
	Base    string
	Mirrors []string
}

var wellKnownMirrors = []mirrorConfig{
	{
		Base: "https://artifacts.k8s.io/binaries/kops/{kopsVersion}/",
		Mirrors: []string{
			"https://github.com/kubernetes/kops/releases/download/v{kopsVersion}/",
		},
	},
	{
		Base: "https://dl.k8s.io/release/",
		Mirrors: []string{
			// We include this mirror in case dl.k8s.io is not directly reachable.
			"https://cdn.dl.k8s.io/release/",
		},
	},
}

func (m *mirrorConfig) findMirrors(u string) ([]string, bool) {
	baseURLString := m.Base
	baseURLString = strings.ReplaceAll(baseURLString, "{kopsVersion}", kops.Version)
	if !strings.HasSuffix(baseURLString, "/") {
		baseURLString += "/"
	}

	// Use mirrors when the URL is not a custom one
	if !strings.HasPrefix(u, baseURLString) {
		return nil, false
	}

	// Use the canonical URL as the first mirror
	mirrors := []string{u}

	for _, mirror := range m.Mirrors {
		mirror = strings.ReplaceAll(mirror, "{kopsVersion}", kops.Version)
		suffix := strings.TrimPrefix(u, baseURLString)

		if strings.HasPrefix(mirror, "https://github.com") {
			// GitHub artifact names are quite different, because the suffix path is collapsed.
			suffix = strings.ReplaceAll(suffix, "/", "-")
			suffix = strings.ReplaceAll(suffix, "linux-amd64-nodeup", "nodeup-linux-amd64")
			suffix = strings.ReplaceAll(suffix, "linux-arm64-nodeup", "nodeup-linux-arm64")
			suffix = strings.ReplaceAll(suffix, "linux-amd64-protokube", "protokube-linux-amd64")
			suffix = strings.ReplaceAll(suffix, "linux-arm64-protokube", "protokube-linux-arm64")
			suffix = strings.ReplaceAll(suffix, "linux-amd64-channels", "channels-linux-amd64")
			suffix = strings.ReplaceAll(suffix, "linux-arm64-channels", "channels-linux-arm64")
		}
		mirrors = append(mirrors, mirror+suffix)
	}
	return mirrors, true
}

// FindURLMirrors will return a list of mirrors for well-known URL locations.
func FindURLMirrors(u string) []string {
	for _, mirror := range wellKnownMirrors {
		mirrors, found := mirror.findMirrors(u)
		if found {
			return mirrors
		}
	}

	return []string{u}
}
