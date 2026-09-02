/*
Copyright 2026 The Kubernetes Authors.

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

package fi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"k8s.io/kops/util/pkg/hashing"
)

func TestOpenOCIAssetUsesDigestAndAnonymousToken(t *testing.T) {
	content := []byte("asset bytes")
	hash, err := hashing.HashAlgorithmSHA256.Hash(bytes.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}
	blobPath := "/v2/prefix/containerd/blobs/sha256:" + hash.Hex()
	var requests []string
	var server *httptest.Server
	server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		switch r.URL.Path {
		case "/v2/":
			w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer realm="%s/token?audience=assets,public", service="registry"`, server.URL))
			w.WriteHeader(http.StatusUnauthorized)
		case "/token":
			if got := r.URL.Query().Get("scope"); got != "repository:prefix/containerd:pull" {
				t.Errorf("scope = %q", got)
			}
			if got := r.URL.Query().Get("service"); got != "registry" {
				t.Errorf("service = %q", got)
			}
			if got := r.URL.Query().Get("audience"); got != "assets,public" {
				t.Errorf("audience = %q", got)
			}
			fmt.Fprint(w, `{"access_token":"pull-token"}`)
		case blobPath:
			if got := r.Header.Get("Authorization"); got != "Bearer pull-token" {
				t.Errorf("Authorization = %q", got)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_, _ = w.Write(content)
		default:
			t.Errorf("unexpected request %q", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	location, err := url.Parse("oci://" + strings.TrimPrefix(server.URL, "https://") + "/prefix/containerd:v2.2.4-amd64")
	if err != nil {
		t.Fatal(err)
	}
	reader, err := openOCIAssetWithTransport(context.Background(), server.Client().Transport, location, hash)
	if err != nil {
		t.Fatalf("openOCIAssetWithTransport() error = %v", err)
	}
	defer reader.Close()
	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, content) {
		t.Fatalf("downloaded %q, want %q", got, content)
	}
	want := []string{"GET /v2/", "GET /token", "GET " + blobPath}
	if fmt.Sprint(requests) != fmt.Sprint(want) {
		t.Fatalf("requests = %v, want %v", requests, want)
	}
}

func TestOpenOCIAssetRejectsHTTPRedirect(t *testing.T) {
	content := []byte("asset bytes")
	hash, err := hashing.HashAlgorithmSHA256.Hash(bytes.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}
	reachedHTTP := false
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reachedHTTP = true
		w.Write(content)
	}))
	defer httpServer.Close()
	blobPath := "/v2/prefix/containerd/blobs/sha256:" + hash.Hex()
	tlsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == blobPath {
			http.Redirect(w, r, httpServer.URL+"/blob", http.StatusTemporaryRedirect)
			return
		}
		http.NotFound(w, r)
	}))
	defer tlsServer.Close()

	location, err := url.Parse("oci://" + strings.TrimPrefix(tlsServer.URL, "https://") + "/prefix/containerd:v2.2.4-amd64")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := openOCIAssetWithTransport(context.Background(), tlsServer.Client().Transport, location, hash); err == nil || !strings.Contains(err.Error(), "must use HTTPS") {
		t.Fatalf("openOCIAssetWithTransport() error = %v, want HTTPS redirect error", err)
	}
	if reachedHTTP {
		t.Fatal("OCI download followed an HTTP redirect")
	}
}
