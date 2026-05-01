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
	"fmt"
	"testing"
)

func TestAzureBlobPathBase(t *testing.T) {
	testCases := []struct {
		container string
		key       string
		base      string
	}{
		{
			container: "c",
			key:       "foo/bar",
			base:      "bar",
		},
		{
			container: "c/",
			key:       "/foo/bar",
			base:      "bar",
		},
		{
			container: "c",
			key:       "/foo/bar/",
			base:      "bar",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test case %d", i), func(t *testing.T) {
			p := NewAzureBlobPath(nil, "a", tc.container, tc.key)
			if a := p.Base(); a != tc.base {
				t.Errorf("expected %s, but got %s", tc.base, a)
			}
		})
	}
}

func TestAzureBlobPathPath(t *testing.T) {
	testCases := []struct {
		account   string
		container string
		key       string
		path      string
	}{
		{
			account:   "a",
			container: "c",
			key:       "foo/bar",
			path:      "azureblob://a/c/foo/bar",
		},
		{
			account:   "a",
			container: "c/",
			key:       "/foo/bar",
			path:      "azureblob://a/c/foo/bar",
		},
		{
			account:   "a",
			container: "c",
			key:       "/foo/bar/",
			path:      "azureblob://a/c/foo/bar/",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test case %d", i), func(t *testing.T) {
			p := NewAzureBlobPath(nil, tc.account, tc.container, tc.key)
			if a := p.Path(); a != tc.path {
				t.Errorf("expected %s, but got %s", tc.path, a)
			}
		})
	}
	p := NewAzureBlobPath(nil, "account", "container", "foo/bar")
	if a, e := p.Path(), "azureblob://account/container/foo/bar"; a != e {
		t.Errorf("expected %s, but got %s", e, a)
	}
}

func TestAzureBlobPathJoin(t *testing.T) {
	p := NewAzureBlobPath(nil, "a", "c", "foo/bar")
	joined := p.Join("p1", "p2")
	if a, e := joined.Path(), "azureblob://a/c/foo/bar/p1/p2"; a != e {
		t.Errorf("expected %s, but got %s", e, a)
	}
}

func TestBuildAzureBlobPath(t *testing.T) {
	testCases := []struct {
		input     string
		account   string
		container string
		key       string
		wantErr   bool
	}{
		{
			input:     "azureblob://account/container/key/path",
			account:   "account",
			container: "container",
			key:       "key/path",
		},
		{
			input:     "azureblob://account/container",
			account:   "account",
			container: "container",
			key:       "",
		},
		{
			// Old format without account is rejected.
			input:   "azureblob://container",
			wantErr: true,
		},
		{
			// Account but no container.
			input:   "azureblob://account/",
			wantErr: true,
		},
	}
	c := &VFSContext{}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			p, err := c.buildAzureBlobPath(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got %+v", tc.input, p)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.input, err)
			}
			if p.account != tc.account {
				t.Errorf("account: expected %q, got %q", tc.account, p.account)
			}
			if p.container != tc.container {
				t.Errorf("container: expected %q, got %q", tc.container, p.container)
			}
			if p.key != tc.key {
				t.Errorf("key: expected %q, got %q", tc.key, p.key)
			}
		})
	}
}
