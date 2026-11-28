/*
Copyright The Kubernetes Authors.

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

package kops

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDownloadKopsMarkers(t *testing.T) {
	for _, tc := range []struct {
		name     string
		marker   string
		basePath string
		legacy   bool
		wantErr  bool
	}{
		{name: "legacy URL", basePath: "/other/build", legacy: true},
		{name: "release version", marker: "1.37.0\n", basePath: "/releases/1.37.0"},
		{name: "prefixed version", marker: "v1.37.0\n", basePath: "/releases/v1.37.0"},
		{name: "CI version", marker: "1.37.0-beta.2+v1.37.0-beta.1-52-gdeadbeef\n", basePath: "/releases/1.37.0-beta.2+v1.37.0-beta.1-52-gdeadbeef"},
		{name: "whitespace", marker: " \r\n1.37.0-beta.2\r\n ", basePath: "/releases/1.37.0-beta.2"},
		{name: "empty marker", marker: "\n", wantErr: true},
		{name: "invalid version", marker: "not-a-version", wantErr: true},
		{name: "major version only", marker: "1", wantErr: true},
		{name: "missing patch version", marker: "1.37", wantErr: true},
		{name: "leading zero", marker: "1.037.0", wantErr: true},
		{name: "current directory", marker: ".", wantErr: true},
		{name: "parent directory", marker: "..", wantErr: true},
		{name: "relative path", marker: "../1.37.0", wantErr: true},
		{name: "multiple lines", marker: "1.37.0\nBASE_URL=https://example.com", wantErr: true},
		{name: "invalid URL", marker: "https://[invalid", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const binary = "test kops binary"
			baseURL := ""
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/releases/latest.txt":
					if tc.legacy {
						fmt.Fprintln(w, baseURL)
					} else {
						fmt.Fprint(w, tc.marker)
					}
				case tc.basePath + "/" + runtime.GOOS + "/" + runtime.GOARCH + "/kops":
					fmt.Fprint(w, binary)
				default:
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			baseURL = server.URL + tc.basePath
			downloadPath := filepath.Join(t.TempDir(), "kops")
			actual, err := DownloadKops(server.URL+"/releases/latest.txt?cache=1", downloadPath, "")
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an invalid marker error")
				}
				if _, err := os.Stat(downloadPath); !os.IsNotExist(err) {
					t.Fatalf("invalid marker created a download file: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if actual != baseURL {
				t.Errorf("base URL = %q, want %q", actual, baseURL)
			}
			content, err := os.ReadFile(downloadPath)
			if err != nil {
				t.Fatal(err)
			}
			if string(content) != binary {
				t.Errorf("downloaded content = %q, want %q", content, binary)
			}
			info, err := os.Stat(downloadPath)
			if err != nil {
				t.Fatal(err)
			}
			if runtime.GOOS != "windows" && info.Mode().Perm() != 0o755 {
				t.Errorf("download mode = %o, want 755", info.Mode().Perm())
			}
		})
	}
}

func TestKopsBaseURLFromMarker(t *testing.T) {
	for _, tc := range []struct {
		name      string
		markerURL string
		contents  string
		want      string
	}{
		{
			name:      "CDN prefix",
			markerURL: "https://dl.k8s.io/ci/kops/latest-1.37.txt",
			contents:  "1.37.0-beta.2+abc123\n",
			want:      "https://dl.k8s.io/ci/kops/1.37.0-beta.2+abc123",
		},
		{
			name:      "query string and fragment",
			markerURL: "https://storage.googleapis.com/k8s-staging-kops/kops/releases/latest.txt?mirror=a/b#marker",
			contents:  "1.37.0-beta.2+abc123",
			want:      "https://storage.googleapis.com/k8s-staging-kops/kops/releases/1.37.0-beta.2+abc123",
		},
		{
			name:      "legacy URL on another host",
			markerURL: "https://dl.k8s.io/ci/kops/latest.txt",
			contents:  "https://storage.googleapis.com/k8s-staging-kops/kops/releases/1.37.0-beta.2+abc123\n",
			want:      "https://storage.googleapis.com/k8s-staging-kops/kops/releases/1.37.0-beta.2+abc123",
		},
		{
			name:      "legacy file URL",
			markerURL: "file:///tmp/latest-ci.txt",
			contents:  "file:///tmp/kops/1.37.0-beta.2+abc123\n",
			want:      "file:///tmp/kops/1.37.0-beta.2+abc123",
		},
		{
			name:      "version-only file URL",
			markerURL: "file:///tmp/kops/latest.txt",
			contents:  "1.37.0-beta.2+abc123\n",
			want:      "file:///tmp/kops/1.37.0-beta.2+abc123",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := kopsBaseURLFromMarker(tc.markerURL, tc.contents)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("base URL = %q, want %q", got, tc.want)
			}
		})
	}
}
