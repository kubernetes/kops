/* Copyright 2017 The Bazel Authors. All rights reserved.

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

// This file contains integration tests for all of Gazelle. It's meant to test
// common usage patterns and check for errors that are difficult to test in
// unit tests.

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bazelbuild/bazel-gazelle/internal/config"
	"github.com/bazelbuild/bazel-gazelle/internal/wspace"
)

type fileSpec struct {
	path, content string
}

func createFiles(files []fileSpec) (string, error) {
	dir, err := ioutil.TempDir(os.Getenv("TEST_TEMPDIR"), "integration_test")
	if err != nil {
		return "", err
	}

	for _, f := range files {
		path := filepath.Join(dir, filepath.FromSlash(f.path))
		if strings.HasSuffix(f.path, "/") {
			if err := os.MkdirAll(path, 0700); err != nil {
				os.RemoveAll(dir)
				return "", err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			os.RemoveAll(dir)
			return "", err
		}
		if err := ioutil.WriteFile(path, []byte(f.content), 0600); err != nil {
			os.RemoveAll(dir)
			return "", err
		}
	}
	return dir, nil
}

// skipIfWorkspaceVisible skips the test if the WORKSPACE file for the
// repository is visible. This happens in newer Bazel versions when tests
// are run without sandboxing, since temp directories may be inside the
// exec root.
func skipIfWorkspaceVisible(t *testing.T, dir string) {
	if parent, err := wspace.Find(dir); err == nil {
		t.Skipf("WORKSPACE visible in parent %q of tmp %q", parent, dir)
	}
}

func checkFiles(t *testing.T, dir string, files []fileSpec) {
	for _, f := range files {
		path := filepath.Join(dir, f.path)
		if strings.HasSuffix(f.path, "/") {
			if st, err := os.Stat(path); err != nil {
				t.Errorf("could not stat %s: %v", f.path, err)
			} else if !st.IsDir() {
				t.Errorf("not a directory: %s", f.path)
			}
		} else {
			want := f.content
			if len(want) > 0 && want[0] == '\n' {
				// Strip leading newline, added for readability.
				want = want[1:]
			}
			gotBytes, err := ioutil.ReadFile(filepath.Join(dir, f.path))
			if err != nil {
				t.Errorf("could not read %s: %v", f.path, err)
				continue
			}
			got := string(gotBytes)
			if got != want {
				t.Errorf("%s: got %s ; want %s", f.path, got, f.content)
			}
		}
	}
}

func runGazelle(wd string, args []string) error {
	oldWd, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := os.Chdir(wd); err != nil {
		return err
	}
	defer os.Chdir(oldWd)

	return run(args)
}

func TestNoRepoRootOrWorkspace(t *testing.T) {
	dir, err := createFiles(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	skipIfWorkspaceVisible(t, dir)
	want := "-repo_root not specified"
	if err := runGazelle(dir, nil); err == nil {
		t.Fatalf("got success; want %q", want)
	} else if !strings.Contains(err.Error(), want) {
		t.Fatalf("got %q; want %q", err, want)
	}
}

func TestNoGoPrefixArgOrRule(t *testing.T) {
	dir, err := createFiles([]fileSpec{
		{path: "WORKSPACE", content: ""},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "-go_prefix not set"
	if err := runGazelle(dir, nil); err == nil {
		t.Fatalf("got success; want %q", want)
	} else if !strings.Contains(err.Error(), want) {
		t.Fatalf("got %q; want %q", err, want)
	}
}

// TestSelectLabelsSorted checks that string lists in srcs and deps are sorted
// using buildifier order, even if they are inside select expressions.
// This applies to both new and existing lists and should preserve comments.
// buildifier does not do this yet bazelbuild/buildtools#122, so we do this
// in addition to calling build.Rewrite.
func TestSelectLabelsSorted(t *testing.T) {
	dir, err := createFiles([]fileSpec{
		{path: "WORKSPACE"},
		{
			path: "BUILD",
			content: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = select({
        "@io_bazel_rules_go//go/platform:linux": [
            # foo comment
            "foo.go",  # side comment
            # bar comment
            "bar.go",
        ],
        "//conditions:default": [],
    }),
    importpath = "example.com/foo",
)
`,
		},
		{
			path: "foo.go",
			content: `
// +build linux

package foo

import (
    _ "example.com/foo/outer"
    _ "example.com/foo/outer/inner"
    _ "github.com/jr_hacker/tools"
)
`,
		},
		{
			path: "bar.go",
			content: `// +build linux

package foo
`,
		},
		{path: "outer/outer.go", content: "package outer"},
		{path: "outer/inner/inner.go", content: "package inner"},
	})
	want := `load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        # bar comment
        "bar.go",
        # foo comment
        "foo.go",  # side comment
    ],
    importpath = "example.com/foo",
    visibility = ["//visibility:public"],
    deps = select({
        "@io_bazel_rules_go//go/platform:linux": [
            "//outer:go_default_library",
            "//outer/inner:go_default_library",
            "@com_github_jr_hacker_tools//:go_default_library",
        ],
        "//conditions:default": [],
    }),
)
`
	if err != nil {
		t.Fatal(err)
	}

	if err := runGazelle(dir, []string{"-go_prefix", "example.com/foo"}); err != nil {
		t.Fatal(err)
	}
	if got, err := ioutil.ReadFile(filepath.Join(dir, "BUILD")); err != nil {
		t.Fatal(err)
	} else if string(got) != want {
		t.Fatalf("got %s ; want %s", string(got), want)
	}
}

func TestFixAndUpdateChanges(t *testing.T) {
	files := []fileSpec{
		{path: "WORKSPACE"},
		{
			path: "BUILD",
			content: `load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_prefix")
load("@io_bazel_rules_go//go:def.bzl", "cgo_library", "go_test")

go_prefix("example.com/foo")

go_library(
    name = "go_default_library",
    srcs = [
        "extra.go",
        "pure.go",
    ],
    library = ":cgo_default_library",
    visibility = ["//visibility:default"],
)

cgo_library(
    name = "cgo_default_library",
    srcs = ["cgo.go"],
)
`,
		},
		{
			path:    "pure.go",
			content: "package foo",
		},
		{
			path: "cgo.go",
			content: `package foo

import "C"
`,
		},
	}

	cases := []struct {
		cmd, want string
	}{
		{
			cmd: "update",
			want: `load("@io_bazel_rules_go//go:def.bzl", "cgo_library", "go_library", "go_prefix")

go_prefix("example.com/foo")

go_library(
    name = "go_default_library",
    srcs = [
        "cgo.go",
        "pure.go",
    ],
    cgo = True,
    importpath = "example.com/foo",
    visibility = ["//visibility:default"],
)

cgo_library(
    name = "cgo_default_library",
    srcs = ["cgo.go"],
)
`,
		}, {
			cmd: "fix",
			want: `load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_prefix")

go_prefix("example.com/foo")

go_library(
    name = "go_default_library",
    srcs = [
        "cgo.go",
        "pure.go",
    ],
    cgo = True,
    importpath = "example.com/foo",
    visibility = ["//visibility:default"],
)
`,
		},
	}

	for _, c := range cases {
		t.Run(c.cmd, func(t *testing.T) {
			dir, err := createFiles(files)
			if err != nil {
				t.Fatal(err)
			}

			if err := runGazelle(dir, []string{c.cmd}); err != nil {
				t.Fatal(err)
			}
			if got, err := ioutil.ReadFile(filepath.Join(dir, "BUILD")); err != nil {
				t.Fatal(err)
			} else if string(got) != c.want {
				t.Fatalf("got %s ; want %s", string(got), c.want)
			}
		})
	}
}

func TestFixUnlinkedCgoLibrary(t *testing.T) {
	files := []fileSpec{
		{path: "WORKSPACE"},
		{
			path: "BUILD",
			content: `load("@io_bazel_rules_go//go:def.bzl", "cgo_library", "go_library")

cgo_library(
    name = "cgo_default_library",
    srcs = ["cgo.go"],
)

go_library(
    name = "go_default_library",
    srcs = ["pure.go"],
    importpath = "example.com/foo",
    visibility = ["//visibility:public"],
)
`,
		}, {
			path:    "pure.go",
			content: "package foo",
		},
	}

	dir, err := createFiles(files)
	if err != nil {
		t.Fatal(err)
	}

	want := `load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["pure.go"],
    importpath = "example.com/foo",
    visibility = ["//visibility:public"],
)
`
	if err := runGazelle(dir, []string{"fix", "-go_prefix", "example.com/foo"}); err != nil {
		t.Fatal(err)
	}
	if got, err := ioutil.ReadFile(filepath.Join(dir, "BUILD")); err != nil {
		t.Fatal(err)
	} else if string(got) != want {
		t.Fatalf("got %s ; want %s", string(got), want)
	}
}

// TestMultipleDirectories checks that all directories in a repository are
// indexed but only directories listed on the command line are updated.
func TestMultipleDirectories(t *testing.T) {
	files := []fileSpec{
		{path: "WORKSPACE"},
		{
			path: "a/BUILD.bazel",
			content: `# This file shouldn't be modified.
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["a.go"],
    importpath = "example.com/foo/x",
)
`,
		}, {
			path:    "a/a.go",
			content: "package a",
		}, {
			path: "b/b.go",
			content: `
package b

import _ "example.com/foo/x"
`,
		},
	}
	dir, err := createFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	args := []string{"-go_prefix", "example.com/foo", "b"}
	if err := runGazelle(dir, args); err != nil {
		t.Fatal(err)
	}

	checkFiles(t, dir, []fileSpec{
		files[1], // should not change
		{
			path: "b/BUILD.bazel",
			content: `load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["b.go"],
    importpath = "example.com/foo/b",
    visibility = ["//visibility:public"],
    deps = ["//a:go_default_library"],
)
`,
		},
	})
}

func TestErrorOutsideWorkspace(t *testing.T) {
	files := []fileSpec{
		{path: "a/"},
		{path: "b/"},
	}
	dir, err := createFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	skipIfWorkspaceVisible(t, dir)

	cases := []struct {
		name, dir, want string
		args            []string
	}{
		{
			name: "outside workspace",
			dir:  dir,
			args: nil,
			want: "WORKSPACE cannot be found",
		}, {
			name: "outside repo_root",
			dir:  filepath.Join(dir, "a"),
			args: []string{"-repo_root", filepath.Join(dir, "b")},
			want: "not a subdirectory of repo root",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := runGazelle(c.dir, c.args); err == nil {
				t.Fatalf("got success; want %q", c.want)
			} else if !strings.Contains(err.Error(), c.want) {
				t.Fatalf("got %q; want %q", err, c.want)
			}
		})
	}
}

func TestBuildFileNameIgnoresBuild(t *testing.T) {
	files := []fileSpec{
		{path: "WORKSPACE"},
		{path: "BUILD/"},
		{
			path:    "a/BUILD",
			content: "!!! parse error",
		}, {
			path:    "a.go",
			content: "package a",
		},
	}
	dir, err := createFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dir)

	args := []string{"-go_prefix", "example.com/foo", "-build_file_name", "BUILD.bazel"}
	if err := runGazelle(dir, args); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "BUILD.bazel")); err != nil {
		t.Errorf("BUILD.bazel not created: %v", err)
	}
}

func TestExternalVendor(t *testing.T) {
	files := []fileSpec{
		{
			path:    "WORKSPACE",
			content: `workspace(name = "banana")`,
		}, {
			path: "a.go",
			content: `package foo

import _ "golang.org/x/bar"
`,
		}, {
			path: "vendor/golang.org/x/bar/bar.go",
			content: `package bar

import _ "golang.org/x/baz"
`,
		}, {
			path:    "vendor/golang.org/x/baz/baz.go",
			content: "package baz",
		},
	}
	dir, err := createFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	args := []string{"-go_prefix", "example.com/foo", "-external", "vendored"}
	if err := runGazelle(dir, args); err != nil {
		t.Fatal(err)
	}

	checkFiles(t, dir, []fileSpec{
		{
			path: config.DefaultValidBuildFileNames[0],
			content: `load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["a.go"],
    importpath = "example.com/foo",
    visibility = ["//visibility:public"],
    deps = ["//vendor/golang.org/x/bar:go_default_library"],
)
`,
		}, {
			path: "vendor/golang.org/x/bar/" + config.DefaultValidBuildFileNames[0],
			content: `load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["bar.go"],
    importmap = "banana/vendor/golang.org/x/bar",
    importpath = "golang.org/x/bar",
    visibility = ["//visibility:public"],
    deps = ["//vendor/golang.org/x/baz:go_default_library"],
)
`,
		},
	})
}

func TestMigrateProtoRules(t *testing.T) {
	files := []fileSpec{
		{path: "WORKSPACE"},
		{
			path: config.DefaultValidBuildFileNames[0],
			content: `
load("@io_bazel_rules_go//proto:go_proto_library.bzl", "go_proto_library")

filegroup(
    name = "go_default_library_protos",
    srcs = ["foo.proto"],
    visibility = ["//visibility:public"],
)

go_proto_library(
    name = "go_default_library",
    srcs = [":go_default_library_protos"],
)
`,
		}, {
			path: "foo.proto",
			content: `syntax = "proto3";

option go_package = "example.com/repo";
`,
		}, {
			path:    "foo.pb.go",
			content: `package repo`,
		},
	}

	for _, tc := range []struct {
		args []string
		want string
	}{
		{
			args: []string{"update", "-go_prefix", "example.com/repo"},
			want: `
load("@io_bazel_rules_go//proto:go_proto_library.bzl", "go_proto_library")

filegroup(
    name = "go_default_library_protos",
    srcs = ["foo.proto"],
    visibility = ["//visibility:public"],
)

go_proto_library(
    name = "go_default_library",
    srcs = [":go_default_library_protos"],
)
`,
		}, {
			args: []string{"fix", "-go_prefix", "example.com/repo"},
			want: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")

proto_library(
    name = "repo_proto",
    srcs = ["foo.proto"],
    visibility = ["//visibility:public"],
)

go_proto_library(
    name = "repo_go_proto",
    importpath = "example.com/repo",
    proto = ":repo_proto",
    visibility = ["//visibility:public"],
)

go_library(
    name = "go_default_library",
    embed = [":repo_go_proto"],
    importpath = "example.com/repo",
    visibility = ["//visibility:public"],
)
`,
		},
	} {
		t.Run(tc.args[0], func(t *testing.T) {
			dir, err := createFiles(files)
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dir)

			if err := runGazelle(dir, tc.args); err != nil {
				t.Fatal(err)
			}

			checkFiles(t, dir, []fileSpec{{
				path:    config.DefaultValidBuildFileNames[0],
				content: tc.want,
			}})
		})
	}
}

func TestRemoveProtoDeletesRules(t *testing.T) {
	files := []fileSpec{
		{path: "WORKSPACE"},
		{
			path: config.DefaultValidBuildFileNames[0],
			content: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")

filegroup(
    name = "go_default_library_protos",
    srcs = ["foo.proto"],
    visibility = ["//visibility:public"],
)

proto_library(
    name = "repo_proto",
    srcs = ["foo.proto"],
    visibility = ["//visibility:public"],
)

go_proto_library(
    name = "repo_go_proto",
    importpath = "example.com/repo",
    proto = ":repo_proto",
    visibility = ["//visibility:public"],
)

go_library(
    name = "go_default_library",
    srcs = ["extra.go"],
    embed = [":repo_go_proto"],
    importpath = "example.com/repo",
    visibility = ["//visibility:public"],
)
`,
		}, {
			path:    "extra.go",
			content: `package repo`,
		},
	}
	dir, err := createFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	args := []string{"fix", "-go_prefix", "example.com/repo"}
	if err := runGazelle(dir, args); err != nil {
		t.Fatal(err)
	}

	checkFiles(t, dir, []fileSpec{{
		path: config.DefaultValidBuildFileNames[0],
		content: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["extra.go"],
    importpath = "example.com/repo",
    visibility = ["//visibility:public"],
)
`,
	}})
}

func TestAddServiceConvertsToGrpc(t *testing.T) {
	files := []fileSpec{
		{path: "WORKSPACE"},
		{
			path: config.DefaultValidBuildFileNames[0],
			content: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")

proto_library(
    name = "repo_proto",
    srcs = ["foo.proto"],
    visibility = ["//visibility:public"],
)

go_proto_library(
    name = "repo_go_proto",
    importpath = "example.com/repo",
    proto = ":repo_proto",
    visibility = ["//visibility:public"],
)

go_library(
    name = "go_default_library",
    embed = [":repo_go_proto"],
    importpath = "example.com/repo",
    visibility = ["//visibility:public"],
)
`,
		}, {
			path: "foo.proto",
			content: `syntax = "proto3";

option go_package = "example.com/repo";

service {}
`,
		},
	}

	dir, err := createFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	args := []string{"-go_prefix", "example.com/repo"}
	if err := runGazelle(dir, args); err != nil {
		t.Fatal(err)
	}

	checkFiles(t, dir, []fileSpec{{
		path: config.DefaultValidBuildFileNames[0],
		content: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")

proto_library(
    name = "repo_proto",
    srcs = ["foo.proto"],
    visibility = ["//visibility:public"],
)

go_proto_library(
    name = "repo_go_proto",
    compilers = ["@io_bazel_rules_go//proto:go_grpc"],
    importpath = "example.com/repo",
    proto = ":repo_proto",
    visibility = ["//visibility:public"],
)

go_library(
    name = "go_default_library",
    embed = [":repo_go_proto"],
    importpath = "example.com/repo",
    visibility = ["//visibility:public"],
)
`,
	}})
}

func TestEmptyGoPrefix(t *testing.T) {
	files := []fileSpec{
		{path: "WORKSPACE"},
		{
			path:    "foo/foo.go",
			content: "package foo",
		}, {
			path: "bar/bar.go",
			content: `
package bar

import (
	_ "fmt"
	_ "foo"
)
`,
		},
	}

	dir, err := createFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	args := []string{"-go_prefix", ""}
	if err := runGazelle(dir, args); err != nil {
		t.Fatal(err)
	}

	checkFiles(t, dir, []fileSpec{{
		path: filepath.Join("bar", config.DefaultValidBuildFileNames[0]),
		content: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["bar.go"],
    importpath = "bar",
    visibility = ["//visibility:public"],
    deps = ["//foo:go_default_library"],
)
`,
	}})
}

// TestResolveKeptImportpath checks that Gazelle can resolve dependencies
// against a library with a '# keep' comment on its importpath attribute
// when the importpath doesn't match what Gazelle would infer.
func TestResolveKeptImportpath(t *testing.T) {
	files := []fileSpec{
		{path: "WORKSPACE"},
		{
			path: "foo/foo.go",
			content: `
package foo

import _ "example.com/alt/baz"
`,
		}, {
			path: "bar/BUILD.bazel",
			content: `load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["bar.go"],
    importpath = "example.com/alt/baz",  # keep
    visibility = ["//visibility:public"],
)
`,
		}, {
			path:    "bar/bar.go",
			content: "package bar",
		},
	}

	dir, err := createFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	args := []string{"-go_prefix", "example.com/repo"}
	if err := runGazelle(dir, args); err != nil {
		t.Fatal(err)
	}

	checkFiles(t, dir, []fileSpec{
		{
			path: "foo/BUILD.bazel",
			content: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["foo.go"],
    importpath = "example.com/repo/foo",
    visibility = ["//visibility:public"],
    deps = ["//bar:go_default_library"],
)
`,
		}, {
			path: "bar/BUILD.bazel",
			content: `load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["bar.go"],
    importpath = "example.com/alt/baz",  # keep
    visibility = ["//visibility:public"],
)
`,
		},
	})
}

// TestResolveVendorSubdirectory checks that Gazelle can resolve libraries
// in a vendor directory which is not at the repository root.
func TestResolveVendorSubdirectory(t *testing.T) {
	files := []fileSpec{
		{path: "WORKSPACE"},
		{
			path:    "sub/vendor/example.com/foo/foo.go",
			content: "package foo",
		}, {
			path: "sub/bar/bar.go",
			content: `
package bar

import _ "example.com/foo"
`,
		},
	}
	dir, err := createFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	args := []string{"-go_prefix", "example.com/repo"}
	if err := runGazelle(dir, args); err != nil {
		t.Fatal(err)
	}

	checkFiles(t, dir, []fileSpec{
		{
			path: "sub/vendor/example.com/foo/BUILD.bazel",
			content: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["foo.go"],
    importmap = "sub/vendor/example.com/foo",
    importpath = "example.com/foo",
    visibility = ["//visibility:public"],
)
`,
		}, {
			path: "sub/bar/BUILD.bazel",
			content: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["bar.go"],
    importpath = "example.com/repo/sub/bar",
    visibility = ["//visibility:public"],
    deps = ["//sub/vendor/example.com/foo:go_default_library"],
)
`,
		},
	})
}

// TestDeleteProtoWithDeps checks that Gazelle will delete proto rules with
// dependencies after the proto sources are removed.
func TestDeleteProtoWithDeps(t *testing.T) {
	files := []fileSpec{
		{path: "WORKSPACE"},
		{
			path: "foo/BUILD.bazel",
			content: `
load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["extra.go"],
    embed = [":scratch_go_proto"],
    importpath = "example.com/repo/foo",
    visibility = ["//visibility:public"],
)

proto_library(
    name = "foo_proto",
    srcs = ["foo.proto"],
    visibility = ["//visibility:public"],
    deps = ["//foo/bar:bar_proto"],
)

go_proto_library(
    name = "foo_go_proto",
    importpath = "example.com/repo/foo",
    proto = ":foo_proto",
    visibility = ["//visibility:public"],
    deps = ["//foo/bar:go_default_library"],
)
`,
		}, {
			path:    "foo/extra.go",
			content: "package foo",
		}, {
			path: "foo/bar/bar.proto",
			content: `
syntax = "proto3";

option go_package = "example.com/repo/foo/bar";

message Bar {};
`,
		},
	}
	dir, err := createFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	args := []string{"-go_prefix", "example.com/repo"}
	if err := runGazelle(dir, args); err != nil {
		t.Fatal(err)
	}

	checkFiles(t, dir, []fileSpec{
		{
			path: "foo/BUILD.bazel",
			content: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["extra.go"],
    importpath = "example.com/repo/foo",
    visibility = ["//visibility:public"],
)
`,
		},
	})
}

func TestCustomRepoNames(t *testing.T) {
	files := []fileSpec{
		{
			path: "WORKSPACE",
			content: `
go_repository(
    name = "custom_repo",
    importpath = "example.com/bar",
    commit = "123456",
)
`,
		}, {
			path: "foo.go",
			content: `
package foo

import _ "example.com/bar"
`,
		},
	}
	dir, err := createFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	args := []string{"-go_prefix", "example.com/foo"}
	if err := runGazelle(dir, args); err != nil {
		t.Fatal(err)
	}

	checkFiles(t, dir, []fileSpec{
		{
			path: "BUILD.bazel",
			content: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["foo.go"],
    importpath = "example.com/foo",
    visibility = ["//visibility:public"],
    deps = ["@custom_repo//:go_default_library"],
)
`,
		},
	})
}

func TestImportReposFromDep(t *testing.T) {
	files := []fileSpec{
		{
			path: "WORKSPACE",
			content: `
http_archive(
    name = "io_bazel_rules_go",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.10.1/rules_go-0.10.1.tar.gz",
    sha256 = "4b14d8dd31c6dbaf3ff871adcd03f28c3274e42abc855cb8fb4d01233c0154dc",
)
http_archive(
    name = "bazel_gazelle",
    url = "https://github.com/bazelbuild/bazel-gazelle/releases/download/0.10.0/bazel-gazelle-0.10.0.tar.gz",
    sha256 = "6228d9618ab9536892aa69082c063207c91e777e51bd3c5544c9c060cafe1bd8",
)
load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains", "go_repository")
go_rules_dependencies()
go_register_toolchains()

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")
gazelle_dependencies()

go_repository(
    name = "org_golang_x_net",
    importpath = "golang.org/x/net",
    tag = "1.2",
)

# keep
go_repository(
    name = "org_golang_x_sys",
    importpath = "golang.org/x/sys",
    remote = "https://github.com/golang/sys",
)

http_archive(
    name = "com_github_go_yaml_yaml",
    urls = ["https://example.com/yaml.tar.gz"],
    sha256 = "1234",
)
`,
		}, {
			path: "Gopkg.lock",
			content: `# This file is autogenerated, do not edit; changes may be undone by the next 'dep ensure'.


[[projects]]
  name = "github.com/pkg/errors"
  packages = ["."]
  revision = "645ef00459ed84a119197bfb8d8205042c6df63d"
  version = "v0.8.0"

[[projects]]
  branch = "master"
  name = "golang.org/x/net"
  packages = ["context"]
  revision = "66aacef3dd8a676686c7ae3716979581e8b03c47"

[[projects]]
  branch = "master"
  name = "golang.org/x/sys"
  packages = ["unix"]
  revision = "bb24a47a89eac6c1227fbcb2ae37a8b9ed323366"

[[projects]]
  branch = "v2"
  name = "github.com/go-yaml/yaml"
  packages = ["."]
  revision = "cd8b52f8269e0feb286dfeef29f8fe4d5b397e0b"

[solve-meta]
  analyzer-name = "dep"
  analyzer-version = 1
  inputs-digest = "05c1cd69be2c917c0cc4b32942830c2acfa044d8200fdc94716aae48a8083702"
  solver-name = "gps-cdcl"
  solver-version = 1
`,
		},
	}
	dir, err := createFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	args := []string{"update-repos", "-from_file", "Gopkg.lock"}
	if err := runGazelle(dir, args); err != nil {
		t.Fatal(err)
	}

	checkFiles(t, dir, []fileSpec{
		{
			path: "WORKSPACE",
			content: `
http_archive(
    name = "io_bazel_rules_go",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.10.1/rules_go-0.10.1.tar.gz",
    sha256 = "4b14d8dd31c6dbaf3ff871adcd03f28c3274e42abc855cb8fb4d01233c0154dc",
)

http_archive(
    name = "bazel_gazelle",
    url = "https://github.com/bazelbuild/bazel-gazelle/releases/download/0.10.0/bazel-gazelle-0.10.0.tar.gz",
    sha256 = "6228d9618ab9536892aa69082c063207c91e777e51bd3c5544c9c060cafe1bd8",
)

load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains")

go_rules_dependencies()

go_register_toolchains()

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")

gazelle_dependencies()

go_repository(
    name = "org_golang_x_net",
    importpath = "golang.org/x/net",
    commit = "66aacef3dd8a676686c7ae3716979581e8b03c47",
)

# keep
go_repository(
    name = "org_golang_x_sys",
    importpath = "golang.org/x/sys",
    remote = "https://github.com/golang/sys",
)

http_archive(
    name = "com_github_go_yaml_yaml",
    urls = ["https://example.com/yaml.tar.gz"],
    sha256 = "1234",
)

go_repository(
    name = "com_github_pkg_errors",
    commit = "645ef00459ed84a119197bfb8d8205042c6df63d",
    importpath = "github.com/pkg/errors",
)
`,
		}})
}

func TestDeleteRulesInEmptyDir(t *testing.T) {
	files := []fileSpec{
		{path: "WORKSPACE"},
		{
			path: "BUILD.bazel",
			content: `
load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_binary")

go_library(
    name = "go_default_library",
    srcs = [
        "bar.go",
        "foo.go",
    ],
    importpath = "example.com/repo",
    visibility = ["//visibility:private"],
)

go_binary(
    name = "cmd",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)
`,
		},
	}
	dir, err := createFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	args := []string{"-go_prefix=example.com/repo"}
	if err := runGazelle(dir, args); err != nil {
		t.Fatal(err)
	}

	checkFiles(t, dir, []fileSpec{
		{
			path:    "BUILD.bazel",
			content: "",
		},
	})
}

func TestFixWorkspaceWithoutGazelle(t *testing.T) {
	files := []fileSpec{
		{
			path: "WORKSPACE",
			content: `
load("@io_bazel_rules_go//go:def.bzl", "go_repository")

go_repository(
    name = "com_example_repo",
    importpath = "example.com/repo",
    tag = "1.2.3",
)
`,
		},
	}
	dir, err := createFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if err := runGazelle(dir, []string{"fix", "-go_prefix="}); err == nil {
		t.Error("got success; want error")
	} else if want := "bazel_gazelle is not declared"; !strings.Contains(err.Error(), want) {
		t.Errorf("got error %v; want error containing %q", err, want)
	}
}

// TODO(jayconrod): more tests
//   run in fix mode in testdata directories to create new files
//   run in diff mode in testdata directories to update existing files (no change)
