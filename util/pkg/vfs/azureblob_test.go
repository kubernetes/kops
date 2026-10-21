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
	"context"
	"fmt"
	"io"
	"maps"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
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

// fakeAzureBlobTransport serves the List Blobs and Delete Blob calls of a
// single container from an in-memory set of blob names, so tests can drive a
// real azblob.Client without a storage account.
type fakeAzureBlobTransport struct {
	container string
	blobs     map[string]bool
}

func (f *fakeAzureBlobTransport) Do(req *http.Request) (*http.Response, error) {
	query := req.URL.Query()
	status, body := http.StatusBadRequest, ""
	switch {
	case req.Method == http.MethodGet && req.URL.Path == "/"+f.container && query.Get("comp") == "list":
		var b strings.Builder
		b.WriteString("<EnumerationResults><Blobs>")
		for _, name := range slices.Sorted(maps.Keys(f.blobs)) {
			if strings.HasPrefix(name, query.Get("prefix")) {
				fmt.Fprintf(&b, "<Blob><Name>%s</Name></Blob>", name)
			}
		}
		b.WriteString("</Blobs></EnumerationResults>")
		status, body = http.StatusOK, b.String()
	case req.Method == http.MethodDelete:
		delete(f.blobs, strings.TrimPrefix(req.URL.Path, "/"+f.container+"/"))
		status = http.StatusAccepted
	}
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}, nil
}

// newFakeAzureBlobPath returns the path for key in container "c" of account
// "a", backed by a fake holding blobs named after two clusters that share a
// name prefix, as in a state store shared by several CI jobs.
func newFakeAzureBlobPath(t *testing.T, key string) (*AzureBlobPath, *fakeAzureBlobTransport) {
	transport := &fakeAzureBlobTransport{
		container: "c",
		blobs: map[string]bool{
			"foo/config":        true,
			"foo/pki/ca.crt":    true,
			"foobar/config":     true,
			"foobar/pki/ca.crt": true,
		},
	}
	client, err := azblob.NewClientWithNoCredential("https://a.blob.core.windows.net/", &azblob.ClientOptions{
		ClientOptions: azcore.ClientOptions{Transport: transport},
	})
	if err != nil {
		t.Fatalf("creating client: %v", err)
	}
	vfsContext := &VFSContext{azureClients: map[string]*azblob.Client{"a": client}}
	return NewAzureBlobPath(vfsContext, "a", "c", key), transport
}

func TestAzureBlobPathReadTree(t *testing.T) {
	testCases := []struct {
		key  string
		want []string
	}{
		{
			key:  "",
			want: []string{"foo/config", "foo/pki/ca.crt", "foobar/config", "foobar/pki/ca.crt"},
		},
		{
			key:  "foo",
			want: []string{"foo/config", "foo/pki/ca.crt"},
		},
		{
			key:  "foo/",
			want: []string{"foo/config", "foo/pki/ca.crt"},
		},
		{
			// A file has no children, as with the S3 and GCS paths.
			key:  "foo/config",
			want: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%q", tc.key), func(t *testing.T) {
			p, _ := newFakeAzureBlobPath(t, tc.key)
			tree, err := p.ReadTree(context.Background())
			if err != nil {
				t.Fatalf("ReadTree: %v", err)
			}
			var got []string
			for _, f := range tree {
				got = append(got, f.(*AzureBlobPath).Key())
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("expected %q, but got %q", tc.want, got)
			}
		})
	}
}

func TestAzureBlobPathRemoveAll(t *testing.T) {
	p, transport := newFakeAzureBlobPath(t, "foo")
	if err := p.RemoveAll(context.Background()); err != nil {
		t.Fatalf("RemoveAll: %v", err)
	}
	got := slices.Sorted(maps.Keys(transport.blobs))
	want := []string{"foobar/config", "foobar/pki/ca.crt"}
	if !slices.Equal(got, want) {
		t.Errorf("expected remaining blobs %q, but got %q", want, got)
	}
}

func TestAzureBlobPathRemoveAllVersions(t *testing.T) {
	p, transport := newFakeAzureBlobPath(t, "foo/config")
	if err := p.RemoveAllVersions(context.Background()); err != nil {
		t.Fatalf("RemoveAllVersions: %v", err)
	}
	got := slices.Sorted(maps.Keys(transport.blobs))
	want := []string{"foo/pki/ca.crt", "foobar/config", "foobar/pki/ca.crt"}
	if !slices.Equal(got, want) {
		t.Errorf("expected remaining blobs %q, but got %q", want, got)
	}
}

func TestAzureBlobPathReadDir(t *testing.T) {
	p, _ := newFakeAzureBlobPath(t, "foo")
	dir, err := p.ReadDir()
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	var got []string
	for _, f := range dir {
		got = append(got, f.(*AzureBlobPath).Key())
	}
	want := []string{"foo/config"}
	if !slices.Equal(got, want) {
		t.Errorf("expected %q, but got %q", want, got)
	}
}
