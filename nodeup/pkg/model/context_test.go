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

package model

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

func buildTestArchive(t *testing.T, files ...string) []byte {
	t.Helper()

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, name := range files {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(name))}); err != nil {
			t.Fatalf("writing tar header: %v", err)
		}
		if _, err := tw.Write([]byte(name)); err != nil {
			t.Fatalf("writing tar content: %v", err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("closing tar writer: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("closing gzip writer: %v", err)
	}
	return buf.Bytes()
}

func TestAddCNIBinAssets(t *testing.T) {
	files := map[string][]byte{
		"/kubelet":                             []byte("kubelet"),
		"/cni-plugins-linux-amd64-v1.9.1.tgz":  buildTestArchive(t, "bridge", "portmap", "LICENSE"),
		"/containerd-2.1.4-linux-amd64.tar.gz": buildTestArchive(t, "bin/containerd"),
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content, ok := files[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Write(content)
	}))
	defer server.Close()

	assets := fi.NewAssetStore(t.TempDir())
	for _, p := range []string{"/kubelet", "/cni-plugins-linux-amd64-v1.9.1.tgz", "/containerd-2.1.4-linux-amd64.tar.gz"} {
		sum := sha256.Sum256(files[p])
		if err := assets.Add(t.Context(), "sha256:"+hex.EncodeToString(sum[:])+"@"+server.URL+p); err != nil {
			t.Fatalf("adding asset %s: %v", p, err)
		}
	}

	b := &NodeupModelContext{Assets: assets}
	c := &fi.NodeupModelBuilderContext{Tasks: make(map[string]fi.NodeupTask)}
	if err := b.AddCNIBinAssets(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var paths []string
	for _, task := range c.Tasks {
		paths = append(paths, task.(*nodetasks.File).Path)
	}
	slices.Sort(paths)
	expected := []string{"/opt/cni/bin/LICENSE", "/opt/cni/bin/bridge", "/opt/cni/bin/portmap"}
	if !slices.Equal(paths, expected) {
		t.Errorf("unexpected files, got %v, want %v", paths, expected)
	}
}
