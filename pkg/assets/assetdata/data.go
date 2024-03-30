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

package assetdata

import (
	"embed"
	"fmt"
	"io/fs"
	"net/url"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kops/util/pkg/hashing"
	"sigs.k8s.io/yaml"
)

//go:embed *.yaml
var embeddedDataFS embed.FS

// GetHash returns the stored hash for the well-known asset, looking it up by the canonicalURL.
// If found, it returns (hash, true, nil)
// If not found, it returns (nil, false, nil)
func GetHash(canonicalURL *url.URL) (*hashing.Hash, bool, error) {
	var allMatches []*file

	if err := fs.WalkDir(embeddedDataFS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		b, err := fs.ReadFile(embeddedDataFS, p)
		if err != nil {
			return fmt.Errorf("reading embedded file %q: %w", p, err)
		}

		manifest, err := parseManifestFile(b)
		if err != nil {
			return fmt.Errorf("parsing embedded file %q: %w", p, err)
		}

		matches := manifest.Matches(canonicalURL.String())
		allMatches = append(allMatches, matches...)
		return nil
	}); err != nil {
		return nil, false, fmt.Errorf("walking embedded data: %w", err)
	}

	hashes := sets.New[string]()
	for _, match := range allMatches {
		hashes.Insert(match.SHA256)
	}
	if len(hashes) == 0 {
		return nil, false, nil
	}
	if len(hashes) > 1 {
		return nil, false, fmt.Errorf("found multiple matches for asset %q", canonicalURL)
	}
	h, err := hashing.FromString(hashes.UnsortedList()[0])
	if err != nil {
		return nil, false, err
	}
	return h, true, nil
}

type file struct {
	Name   string `json:"name,omitempty"`
	SHA256 string `json:"sha256,omitempty"`
}

type fileStore struct {
	Base string `json:"base,omitempty"`
}

type manifest struct {
	FileStores []fileStore `json:"filestores,omitempty"`
	Files      []file      `json:"files,omitempty"`
}

func parseManifestFile(b []byte) (*manifest, error) {
	m := &manifest{}
	if err := yaml.Unmarshal(b, m); err != nil {
		return nil, fmt.Errorf("parsing yaml: %w", err)
	}
	return m, nil
}

func (m *manifest) Matches(canonicalURL string) []*file {
	var matches []*file
	for _, fileStore := range m.FileStores {
		if !strings.HasPrefix(canonicalURL, fileStore.Base) {
			continue
		}
		relativePath := strings.TrimPrefix(canonicalURL, fileStore.Base)
		for i := range m.Files {
			f := &m.Files[i]
			if f.Name == relativePath {
				matches = append(matches, f)
			}
		}
	}
	return matches
}
