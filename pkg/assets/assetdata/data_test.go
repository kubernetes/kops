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
			Name: "https://dl.k8s.io/release/v1.32.0/bin/linux/amd64/kubelet",
			Hash: "5ad4965598773d56a37a8e8429c3dc3d86b4c5c26d8417ab333ae345c053dae2",
		},
		{
			Name: "https://github.com/opencontainers/runc/releases/download/v1.3.0/runc.amd64",
			Hash: "028986516ab5646370edce981df2d8e8a8d12188deaf837142a02097000ae2f2",
		},
		{
			Name: "https://github.com/opencontainers/runc/releases/download/v1.3.0/runc.arm64",
			Hash: "85c5e4e4f72e442c8c17bac07527cd4f961ee48e4f2b71797f7533c94f4a52b9",
		},
		{
			Name: "https://github.com/opencontainers/runc/releases/download/v1.3.5/runc.amd64",
			Hash: "66fa8390be8fb3b23dfbb60c767368bb5b51f1acfa88692bbff1a82953d4d9e9",
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
