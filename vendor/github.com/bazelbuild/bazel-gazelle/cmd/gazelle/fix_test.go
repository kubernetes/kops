/* Copyright 2016 The Bazel Authors. All rights reserved.

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

package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/bazelbuild/bazel-gazelle/internal/config"
	bf "github.com/bazelbuild/buildtools/build"
)

func TestMain(m *testing.M) {
	tmpdir := os.Getenv("TEST_TMPDIR")
	flag.Set("repo_root", tmpdir)
	os.Exit(m.Run())
}

func defaultArgs(dir string) []string {
	return []string{
		"-repo_root", dir,
		"-go_prefix", "example.com/repo",
		dir,
	}
}

func TestFixFile(t *testing.T) {
	tmpdir := os.Getenv("TEST_TMPDIR")
	dir, err := ioutil.TempDir(tmpdir, "")
	if err != nil {
		t.Fatalf("ioutil.TempDir(%q, %q) failed with %v; want success", tmpdir, "", err)
	}
	defer os.RemoveAll(dir)

	stubFile := &bf.File{
		Path: filepath.Join(dir, "BUILD.bazel"),
		Stmt: []bf.Expr{
			&bf.CallExpr{
				X: &bf.LiteralExpr{Token: "foo_rule"},
				List: []bf.Expr{
					&bf.BinaryExpr{
						X:  &bf.LiteralExpr{Token: "name"},
						Op: "=",
						Y:  &bf.StringExpr{Value: "bar"},
					},
				},
			},
		},
	}
	c := &config.Config{}

	if err := fixFile(c, stubFile, stubFile.Path); err != nil {
		t.Errorf("fixFile(%#v) failed with %v; want success", stubFile, err)
		return
	}

	buf, err := ioutil.ReadFile(stubFile.Path)
	if err != nil {
		t.Errorf("ioutil.ReadFile(%q) failed with %v; want success", stubFile.Path, err)
		return
	}
	if got, want := string(buf), bf.FormatString(stubFile); got != want {
		t.Errorf("buf = %q; want %q", got, want)
	}
}

func TestCreateFile(t *testing.T) {
	// Create a directory with a simple .go file.
	tmpdir := os.Getenv("TEST_TMPDIR")
	dir, err := ioutil.TempDir(tmpdir, "")
	if err != nil {
		t.Fatalf("ioutil.TempDir(%q, %q) failed with %v; want success", tmpdir, "", err)
	}
	defer os.RemoveAll(dir)

	goFile := filepath.Join(dir, "main.go")
	if err = ioutil.WriteFile(goFile, []byte("package main"), 0600); err != nil {
		t.Fatalf("error writing file %q: %v", goFile, err)
	}

	// Check that Gazelle creates a new file named "BUILD.bazel".
	run(defaultArgs(dir))

	buildFile := filepath.Join(dir, "BUILD.bazel")
	if _, err = os.Stat(buildFile); err != nil {
		t.Errorf("could not stat BUILD.bazel: %v", err)
	}
}

func TestUpdateFile(t *testing.T) {
	// Create a directory with a simple .go file and an empty BUILD file.
	tmpdir := os.Getenv("TEST_TMPDIR")
	dir, err := ioutil.TempDir(tmpdir, "")
	if err != nil {
		t.Fatalf("ioutil.TempDir(%q, %q) failed with %v; want success", tmpdir, "", err)
	}
	defer os.RemoveAll(dir)

	goFile := filepath.Join(dir, "main.go")
	if err = ioutil.WriteFile(goFile, []byte("package main"), 0600); err != nil {
		t.Fatalf("error writing file %q: %v", goFile, err)
	}

	buildFile := filepath.Join(dir, "BUILD")
	if err = ioutil.WriteFile(buildFile, nil, 0600); err != nil {
		t.Fatalf("error writing file %q: %v", buildFile, err)
	}

	// Check that Gazelle updates the BUILD file in place.
	run(defaultArgs(dir))
	if st, err := os.Stat(buildFile); err != nil {
		t.Errorf("could not stat BUILD: %v", err)
	} else if st.Size() == 0 {
		t.Errorf("BUILD was not updated")
	}

	if _, err = os.Stat(filepath.Join(dir, "BUILD.bazel")); err == nil {
		t.Errorf("BUILD.bazel should not exist")
	}
}
