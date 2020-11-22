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
			p := NewAzureBlobPath(nil, tc.container, tc.key)
			if a := p.Base(); a != tc.base {
				t.Errorf("expected %s, but got %s", tc.base, a)
			}
		})
	}

}

func TestAzureBlobPathPath(t *testing.T) {
	testCases := []struct {
		container string
		key       string
		path      string
	}{
		{
			container: "c",
			key:       "foo/bar",
			path:      "azureblob://c/foo/bar",
		},
		{
			container: "c/",
			key:       "/foo/bar",
			path:      "azureblob://c/foo/bar",
		},
		{
			container: "c",
			key:       "/foo/bar/",
			path:      "azureblob://c/foo/bar/",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test case %d", i), func(t *testing.T) {
			p := NewAzureBlobPath(nil, tc.container, tc.key)
			if a := p.Path(); a != tc.path {
				t.Errorf("expected %s, but got %s", tc.path, a)
			}
		})
	}
	p := NewAzureBlobPath(nil, "container", "foo/bar")
	if a, e := p.Path(), "azureblob://container/foo/bar"; a != e {
		t.Errorf("expected %s, but got %s", e, a)
	}
}

func TestAzureBlobPathJoin(t *testing.T) {
	p := NewAzureBlobPath(nil, "c", "foo/bar")
	joined := p.Join("p1", "p2")
	if a, e := joined.Path(), "azureblob://c/foo/bar/p1/p2"; a != e {
		t.Errorf("expected %s, but got %s", e, a)
	}
}
