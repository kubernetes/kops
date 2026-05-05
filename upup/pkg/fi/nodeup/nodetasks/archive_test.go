/*
Copyright 2017 The Kubernetes Authors.

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

package nodetasks

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/kops/upup/pkg/fi"
)

func TestArchiveDependencies(t *testing.T) {
	grid := []struct {
		parent fi.NodeupTask
		child  fi.NodeupTask
	}{
		{
			parent: &File{
				Path: "/var",
				Type: FileType_Directory,
			},
			child: &Archive{
				TargetDir: "/var/something",
			},
		},
		{
			parent: &Archive{
				TargetDir: "/var/something",
			},
			child: &File{
				Path: "/var/something/subdir",
				Type: FileType_Directory,
			},
		},
	}

	for _, g := range grid {
		allTasks := make(map[string]fi.NodeupTask)
		allTasks["parent"] = g.parent
		allTasks["child"] = g.child

		deps := g.parent.(fi.NodeupHasDependencies).GetDependencies(allTasks)
		if len(deps) != 0 {
			t.Errorf("found unexpected dependencies for parent: %v %v", g.parent, deps)
		}

		childDeps := g.child.(fi.NodeupHasDependencies).GetDependencies(allTasks)
		if len(childDeps) != 1 {
			t.Errorf("found unexpected dependencies for child: %v %v", g.child, childDeps)
		}
	}
}

func TestExtractArchive(t *testing.T) {
	archivePath := writeTestArchive(t, true, []testTarEntry{
		{
			name: "root/bin/tool",
			mode: 0o755,
			body: "hello",
		},
		{
			name: "root/etc/config",
			mode: 0o644,
			body: "config",
		},
	})

	targetDir := t.TempDir()
	if err := extractArchive(archivePath, targetDir, 1, ""); err != nil {
		t.Fatalf("extractArchive() error = %v", err)
	}

	assertFileContents(t, filepath.Join(targetDir, "bin/tool"), "hello")
	assertFileContents(t, filepath.Join(targetDir, "etc/config"), "config")
}

func TestExtractArchiveMapFiles(t *testing.T) {
	archivePath := writeTestArchive(t, false, []testTarEntry{
		{
			name: "pkg/bin/tool",
			mode: 0o755,
			body: "hello",
		},
		{
			name: "pkg/lib/ignored",
			mode: 0o644,
			body: "ignored",
		},
	})

	targetDir := t.TempDir()
	if err := extractArchive(archivePath, targetDir, 2, "pkg/bin/*"); err != nil {
		t.Fatalf("extractArchive() error = %v", err)
	}

	assertFileContents(t, filepath.Join(targetDir, "tool"), "hello")
	if _, err := os.Stat(filepath.Join(targetDir, "ignored")); !os.IsNotExist(err) {
		t.Fatalf("expected ignored file not to be extracted, stat error = %v", err)
	}
}

func TestExtractArchiveRejectsTraversal(t *testing.T) {
	baseDir := t.TempDir()
	archivePath := writeTestArchive(t, false, []testTarEntry{
		{
			name: "../evil",
			mode: 0o644,
			body: "bad",
		},
	})

	err := extractArchive(archivePath, filepath.Join(baseDir, "target"), 0, "")
	if err == nil {
		t.Fatalf("extractArchive() expected error")
	}
	if _, err := os.Stat(filepath.Join(baseDir, "evil")); !os.IsNotExist(err) {
		t.Fatalf("expected traversal target not to be created, stat error = %v", err)
	}
}

func TestExtractArchiveRejectsSymlinkEscape(t *testing.T) {
	outsideDir := t.TempDir()
	archivePath := writeTestArchive(t, false, []testTarEntry{
		{
			name:     "escape",
			typeflag: tar.TypeSymlink,
			linkname: outsideDir,
		},
		{
			name: "escape/pwned",
			mode: 0o644,
			body: "bad",
		},
	})

	err := extractArchive(archivePath, t.TempDir(), 0, "")
	if err == nil {
		t.Fatalf("extractArchive() expected error")
	}
	if _, err := os.Stat(filepath.Join(outsideDir, "pwned")); !os.IsNotExist(err) {
		t.Fatalf("expected symlink escape target not to be created, stat error = %v", err)
	}
}

type testTarEntry struct {
	name     string
	typeflag byte
	linkname string
	mode     int64
	body     string
}

func writeTestArchive(t *testing.T, gzipArchive bool, entries []testTarEntry) string {
	t.Helper()

	var buffer bytes.Buffer
	var output io.Writer = &buffer
	var gzipWriter *gzip.Writer
	if gzipArchive {
		gzipWriter = gzip.NewWriter(&buffer)
		output = gzipWriter
	}
	writer := tar.NewWriter(output)

	for _, entry := range entries {
		typeflag := entry.typeflag
		if typeflag == 0 {
			typeflag = tar.TypeReg
		}
		mode := entry.mode
		if mode == 0 {
			mode = 0o644
		}
		header := &tar.Header{
			Name:     entry.name,
			Typeflag: typeflag,
			Linkname: entry.linkname,
			Mode:     mode,
			Size:     int64(len(entry.body)),
		}
		if typeflag != tar.TypeReg {
			header.Size = 0
		}
		if err := writer.WriteHeader(header); err != nil {
			t.Fatalf("WriteHeader() error = %v", err)
		}
		if header.Size != 0 {
			if _, err := writer.Write([]byte(entry.body)); err != nil {
				t.Fatalf("Write() error = %v", err)
			}
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if gzipWriter != nil {
		if err := gzipWriter.Close(); err != nil {
			t.Fatalf("gzip Close() error = %v", err)
		}
	}

	archivePath := filepath.Join(t.TempDir(), "archive.tar")
	if err := os.WriteFile(archivePath, buffer.Bytes(), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return archivePath
}

func assertFileContents(t *testing.T, path string, expected string) {
	t.Helper()
	actual, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	if string(actual) != expected {
		t.Fatalf("ReadFile(%q) = %q, expected %q", path, actual, expected)
	}
}
