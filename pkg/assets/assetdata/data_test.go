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
		{Name: "https://dl.k8s.io/release/v1.26.0/bin/linux/amd64/kubelet", Hash: "sha256:b64949fe696c77565edbe4100a315b6bf8f0e2325daeb762f7e865f16a6e54b5"},
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
		got := h.String()
		want := g.Hash
		if got != g.Hash {
			t.Errorf("unexpected hash for %q; got %q, want %q", g.Name, got, want)
		}
	}
}
