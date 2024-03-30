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
	"net/url"
	"testing"
)

func TestGetHash(t *testing.T) {
	grid := []struct {
		Name string
		Hash string
	}{
		{
			Name: "https://dl.k8s.io/release/v1.26.0/bin/linux/amd64/kubelet",
			Hash: "b64949fe696c77565edbe4100a315b6bf8f0e2325daeb762f7e865f16a6e54b5",
		},
		{
			Name: "https://github.com/opencontainers/runc/releases/download/v1.1.0/runc.amd64",
			Hash: "ab1c67fbcbdddbe481e48a55cf0ef9a86b38b166b5079e0010737fd87d7454bb",
		},
		{
			Name: "https://github.com/opencontainers/runc/releases/download/v1.1.0/runc.arm64",
			Hash: "9ec8e68feabc4e7083a4cfa45ebe4d529467391e0b03ee7de7ddda5770b05e68",
		},
		{
			Name: "https://github.com/opencontainers/runc/releases/download/v1.1.12/runc.amd64",
			Hash: "aadeef400b8f05645768c1476d1023f7875b78f52c7ff1967a6dbce236b8cbd8",
		},
	}

	for _, g := range grid {
		u, err := url.Parse(g.Name)
		if err != nil {
			t.Fatalf("parsing url %q: %v", g.Name, err)
		}
		h, found, err := GetHash(u)
		if err != nil {
			t.Fatalf("getting hash for %q: %v", g.Name, err)
		}
		if !found {
			t.Fatalf("hash for %q was not found", g.Name)
		}
		got := h.Hex()
		want := g.Hash
		if got != g.Hash {
			t.Errorf("unexpected hash for %q; got %q, want %q", g.Name, got, want)
		}
	}
}
