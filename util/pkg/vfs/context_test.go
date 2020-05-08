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

package vfs

import (
	"os"
	"testing"
)

func TestBuildVaultPath(t *testing.T) {
	token := os.Getenv("VAULT_DEV_ROOT_TOKEN_ID")
	if token == "" {
		t.Skip("No vault dev token set")
	}
	if os.Getenv("VAULT_TOKEN") != token {
		t.Skip("BuildVaultPath test needs VAULT_TOKEN == VAULT_DEV_ROOT_TOKEN_ID")
	}
	grid := []struct {
		url       string
		scheme    string
		vaultAddr string
	}{
		{
			url:       "vault://localhost:8200/foo/bar?tls=false",
			scheme:    "http://",
			vaultAddr: "http://localhost:8200",
		},
		{
			url:       "vault://foo.test.bar/foo/bar?tls=false",
			scheme:    "http://",
			vaultAddr: "http://foo.test.bar",
		},
		{
			url:       "vault://foo.test.bar/foo/bar",
			scheme:    "https://",
			vaultAddr: "https://foo.test.bar",
		},
	}

	for _, g := range grid {
		context := &VFSContext{}
		p, err := context.buildVaultPath(g.url)
		if err != nil {
			t.Fatalf("Unexepcted error for %q: %v", g.url, err)
		}
		if p.String() != g.url {
			t.Errorf("Unexpected path: expected %q, actual %q", g.url, p.Path())
		}

		vaultAddr := p.vaultClient.Address()
		if vaultAddr != g.vaultAddr {
			t.Errorf("Unexpected vaultAddr: expected %q, actual %q", g.vaultAddr, vaultAddr)
		}

		if p.scheme != g.scheme {
			t.Errorf("Unexpected scheme for %q: expected %q, actual %q", g.url, g.scheme, p.scheme)
		}

	}

}
