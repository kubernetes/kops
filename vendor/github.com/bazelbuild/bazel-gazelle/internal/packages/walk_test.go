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

package packages_test

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/bazelbuild/bazel-gazelle/internal/config"
	"github.com/bazelbuild/bazel-gazelle/internal/packages"
	bf "github.com/bazelbuild/buildtools/build"
)

func tempDir() (string, error) {
	return ioutil.TempDir(os.Getenv("TEST_TMPDIR"), "walk_test")
}

type fileSpec struct {
	path, content, symlink string
}

func checkFiles(t *testing.T, files []fileSpec, goPrefix string, want []*packages.Package) {
	dir, err := createFiles(files)
	if err != nil {
		t.Fatalf("createFiles() failed with %v; want success", err)
	}
	defer os.RemoveAll(dir)

	for _, p := range want {
		p.Dir = filepath.Join(dir, filepath.FromSlash(p.Rel))
	}

	c := &config.Config{
		RepoRoot:            dir,
		GoPrefix:            goPrefix,
		Dirs:                []string{dir},
		ValidBuildFileNames: config.DefaultValidBuildFileNames,
	}
	got := walkPackages(c)
	checkPackages(t, got, want)
}

func createFiles(files []fileSpec) (string, error) {
	dir, err := tempDir()
	if err != nil {
		return "", err
	}
	for _, f := range files {
		path := filepath.Join(dir, f.path)
		if strings.HasSuffix(f.path, "/") {
			if err := os.MkdirAll(path, 0700); err != nil {
				return dir, err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return "", err
		}
		if f.symlink != "" {
			if err := os.Symlink(f.symlink, path); err != nil {
				return "", err
			}
			continue
		}
		if err := ioutil.WriteFile(path, []byte(f.content), 0600); err != nil {
			return "", err
		}
	}
	return dir, nil
}

func walkPackages(c *config.Config) []*packages.Package {
	var pkgs []*packages.Package
	packages.Walk(c, c.RepoRoot, func(_, _ string, _ *config.Config, pkg *packages.Package, _ *bf.File, _ bool) {
		if pkg != nil {
			pkgs = append(pkgs, pkg)
		}
	})
	return pkgs
}

func checkPackages(t *testing.T, got []*packages.Package, want []*packages.Package) {
	if len(got) != len(want) {
		names := []string{}
		for _, p := range got {
			names = append(names, p.Name)
		}
		t.Fatalf("got %d packages %v; want %d", len(got), names, len(want))
	}
	for i := 0; i < len(got); i++ {
		checkPackage(t, got[i], want[i])
	}
}

func checkPackage(t *testing.T, got, want *packages.Package) {
	// TODO: Implement Stringer or Formatter to get more readable output.
	if !reflect.DeepEqual(got, want) {
		t.Errorf("for package %q, got %#v; want %#v", want.Name, got, want)
	}
}

func TestWalkEmpty(t *testing.T) {
	files := []fileSpec{
		{path: "a/foo.c"},
		{path: "b/BUILD"},
		{path: "c/"},
	}
	want := []*packages.Package{}
	checkFiles(t, files, "", want)
}

func TestWalkSimple(t *testing.T) {
	files := []fileSpec{{path: "lib.go", content: "package lib"}}
	want := []*packages.Package{
		{
			Name:       "lib",
			ImportPath: "example.com/repo",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"lib.go"},
				},
			},
		},
	}
	checkFiles(t, files, "example.com/repo", want)
}

func TestWalkNested(t *testing.T) {
	files := []fileSpec{
		{path: "a/foo.go", content: "package a"},
		{path: "b/c/bar.go", content: "package c"},
		{path: "b/d/baz.go", content: "package main"},
	}
	want := []*packages.Package{
		{
			Name:       "a",
			Rel:        "a",
			ImportPath: "example.com/repo/a",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"foo.go"},
				},
			},
		},
		{
			Name:       "c",
			Rel:        "b/c",
			ImportPath: "example.com/repo/b/c",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"bar.go"},
				},
			},
		},
		{
			Name:       "main",
			Rel:        "b/d",
			ImportPath: "example.com/repo/b/d",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"baz.go"},
				},
			},
		},
	}
	checkFiles(t, files, "example.com/repo", want)
}

func TestProtoOnly(t *testing.T) {
	files := []fileSpec{
		{path: "a/a.proto"},
	}
	want := []*packages.Package{
		{
			Name:       "a",
			Rel:        "a",
			ImportPath: "example.com/repo/a",
			Proto: packages.ProtoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"a.proto"},
				},
			},
		},
	}
	checkFiles(t, files, "example.com/repo", want)
}

func TestMultiplePackagesWithDefault(t *testing.T) {
	files := []fileSpec{
		{path: "a/a.go", content: "package a"},
		{path: "a/b.go", content: "package b"},
	}
	want := []*packages.Package{
		{
			Name:       "a",
			Rel:        "a",
			ImportPath: "example.com/repo/a",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"a.go"},
				},
			},
		},
	}
	checkFiles(t, files, "example.com/repo", want)
}

func TestMultiplePackagesWithoutDefault(t *testing.T) {
	files := []fileSpec{
		{path: "a/b.go", content: "package b"},
		{path: "a/c.go", content: "package c"},
	}
	dir, err := createFiles(files)
	if err != nil {
		t.Fatalf("createFiles() failed with %v; want success", err)
	}
	defer os.RemoveAll(dir)

	c := &config.Config{
		RepoRoot:            dir,
		Dirs:                []string{dir},
		ValidBuildFileNames: config.DefaultValidBuildFileNames,
	}
	got := walkPackages(c)
	if len(got) > 0 {
		t.Errorf("got %v; want empty slice", got)
	}
}

func TestMultiplePackagesWithProtoDefault(t *testing.T) {
	files := []fileSpec{
		{path: "a/a.proto", content: `syntax = "proto2";
package a;
`},
		{path: "a/b.go", content: "package b"},
	}
	want := []*packages.Package{
		{
			Name:       "a",
			Rel:        "a",
			ImportPath: "example.com/repo/a",
			Proto: packages.ProtoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"a.proto"},
				},
			},
		},
	}
	checkFiles(t, files, "example.com/repo", want)
}

func TestRootWithPrefix(t *testing.T) {
	files := []fileSpec{
		{path: "a.go", content: "package a"},
		{path: "b.go", content: "package b"},
	}
	want := []*packages.Package{
		{
			Name:       "a",
			ImportPath: "example.com/a",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"a.go"},
				},
			},
		},
	}
	checkFiles(t, files, "example.com/a", want)
}

func TestRootWithoutPrefix(t *testing.T) {
	files := []fileSpec{
		{path: "a.go", content: "package a"},
		{path: "b.go", content: "package b"},
	}
	dir, err := createFiles(files)
	if err != nil {
		t.Fatalf("createFiles() failed with %v; want success", err)
	}
	defer os.RemoveAll(dir)

	c := &config.Config{
		RepoRoot:            dir,
		Dirs:                []string{dir},
		ValidBuildFileNames: config.DefaultValidBuildFileNames,
	}
	got := walkPackages(c)
	if len(got) > 0 {
		t.Errorf("got %v; want empty slice", got)
	}
}

func TestVendorResetsPrefix(t *testing.T) {
	files := []fileSpec{
		{path: "vendor/"},
		{path: "sub/vendor/"},
	}
	dir, err := createFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	basePrefix := "example.com/repo"
	c := &config.Config{
		RepoRoot:            dir,
		Dirs:                []string{dir},
		ValidBuildFileNames: config.DefaultValidBuildFileNames,
		GoPrefix:            basePrefix,
	}
	packages.Walk(c, c.RepoRoot, func(_, rel string, c *config.Config, _ *packages.Package, _ *bf.File, _ bool) {
		if path.Base(rel) != "vendor" {
			return
		}
		if c.GoPrefix != "" {
			t.Errorf("in %q, GoPrefix not reset", rel)
		}
		if c.GoPrefixRel != rel {
			t.Errorf("in %q, GoPrefixRel not set", rel)
		}
	})
	if c.GoPrefix != basePrefix {
		t.Errorf("prefix in base configuration was modified: %q", c.GoPrefix)
	}
}

func TestPrefixEmpty(t *testing.T) {
	files := []fileSpec{
		{path: "a.go", content: "package foo"},
	}
	want := []*packages.Package{}
	checkFiles(t, files, "", want)
}

func TestProtoImportPath(t *testing.T) {
	files := []fileSpec{{
		path: "foo.proto",
		content: `syntax = "proto3";
option go_package = "example.com/repo/foo";
`,
	}}
	want := []*packages.Package{{
		Name:       "foo",
		ImportPath: "example.com/repo/foo",
		Proto: packages.ProtoTarget{
			Sources: packages.PlatformStrings{
				Generic: []string{"foo.proto"},
			},
		},
	}}
	checkFiles(t, files, "", want)
}

func TestTestdata(t *testing.T) {
	files := []fileSpec{
		{path: "raw/testdata/"},
		{path: "raw/a.go", content: "package raw"},
		{path: "with_build/testdata/BUILD"},
		{path: "with_build/a.go", content: "package with_build"},
		{path: "with_build_bazel/testdata/BUILD.bazel"},
		{path: "with_build_bazel/a.go", content: "package with_build_bazel"},
		{path: "with_build_nested/testdata/x/BUILD"},
		{path: "with_build_nested/a.go", content: "package with_build_nested"},
		{path: "with_go/testdata/a.go", content: "package testdata"},
		{path: "with_go/a.go", content: "package with_go"},
	}
	want := []*packages.Package{
		{
			Name:       "raw",
			Rel:        "raw",
			ImportPath: "example.com/repo/raw",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"a.go"},
				},
			},
			HasTestdata: true,
		},
		{
			Name:       "with_build",
			Rel:        "with_build",
			ImportPath: "example.com/repo/with_build",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"a.go"},
				},
			},
			HasTestdata: false,
		},
		{
			Name:       "with_build_bazel",
			Rel:        "with_build_bazel",
			ImportPath: "example.com/repo/with_build_bazel",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"a.go"},
				},
			},
			HasTestdata: false,
		},
		{
			Name:       "with_build_nested",
			Rel:        "with_build_nested",
			ImportPath: "example.com/repo/with_build_nested",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"a.go"},
				},
			},
			HasTestdata: false,
		},
		{
			Name:       "testdata",
			Rel:        "with_go/testdata",
			ImportPath: "example.com/repo/with_go/testdata",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"a.go"},
				},
			},
		},
		{
			Name:       "with_go",
			Rel:        "with_go",
			ImportPath: "example.com/repo/with_go",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"a.go"},
				},
			},
			HasTestdata: false,
		},
	}
	checkFiles(t, files, "example.com/repo", want)
}

func TestGenerated(t *testing.T) {
	files := []fileSpec{
		{
			path: "gen/BUILD",
			content: `
genrule(
    name = "from_genrule",
    outs = ["foo.go", "bar.go", "w.txt", "x.c", "y.s", "z.S"],
)

gen_other(
    name = "from_gen_other",
    out = "baz.go",
)
`,
		},
		{
			path: "gen/foo.go",
			content: `package foo

import "github.com/jr_hacker/stuff"
`,
		},
	}
	want := []*packages.Package{
		{
			Name:       "foo",
			Rel:        "gen",
			ImportPath: "example.com/repo/gen",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"bar.go", "baz.go", "foo.go", "y.s"},
				},
				Imports: packages.PlatformStrings{
					Generic: []string{"github.com/jr_hacker/stuff"},
				},
			},
		},
	}
	checkFiles(t, files, "example.com/repo", want)
}

func TestGeneratedCgo(t *testing.T) {
	files := []fileSpec{
		{
			path: "gen/BUILD",
			content: `
genrule(
    name = "from_genrule",
    outs = ["foo.go", "bar.go", "w.txt", "x.c", "y.s", "z.S"],
)

gen_other(
    name = "from_gen_other",
    out = "baz.go",
)
`,
		},
		{
			path: "gen/foo.go",
			content: `package foo

import "C"

import "github.com/jr_hacker/stuff"
`,
		},
	}
	want := []*packages.Package{
		{
			Name:       "foo",
			Rel:        "gen",
			ImportPath: "example.com/repo/gen",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"bar.go", "baz.go", "foo.go", "x.c", "y.s", "z.S"},
				},
				Imports: packages.PlatformStrings{
					Generic: []string{"github.com/jr_hacker/stuff"},
				},
				Cgo: true,
			},
		},
	}
	checkFiles(t, files, "example.com/repo", want)
}

func TestIgnore(t *testing.T) {
	files := []fileSpec{
		{
			path:    "BUILD",
			content: "# gazelle:ignore",
		}, {
			path:    "foo.go",
			content: "package foo",
		}, {
			path:    "bar/bar.go",
			content: "package bar",
		},
	}
	want := []*packages.Package{
		{
			Name:       "bar",
			Rel:        "bar",
			ImportPath: "example.com/repo/bar",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"bar.go"},
				},
			},
		},
	}
	checkFiles(t, files, "example.com/repo", want)
}

func TestExcluded(t *testing.T) {
	files := []fileSpec{
		{
			path:    "BUILD",
			content: "# gazelle:exclude exclude/do.go",
		}, {
			path: "exclude/BUILD",
			content: `
# gazelle:exclude not.go

# gazelle:exclude build.go

genrule(
    name = "gen_build",
    outs = ["build.go"],
)
`,
		},
		{
			path:    "exclude/do.go",
			content: "",
		},
		{
			path:    "exclude/not.go",
			content: "",
		},
		{
			path:    "exclude/build.go",
			content: "",
		},
		{
			path:    "exclude/real.go",
			content: "package exclude",
		},
	}
	want := []*packages.Package{
		{
			Name:       "exclude",
			Rel:        "exclude",
			ImportPath: "example.com/repo/exclude",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"real.go"},
				},
			},
		},
	}
	checkFiles(t, files, "example.com/repo", want)
}

func TestExcludedPbGo(t *testing.T) {
	files := []fileSpec{
		{
			path: "exclude/BUILD",
			content: `
# gazelle:exclude a.proto
`,
		},
		{
			path: "exclude/a.proto",
			content: `syntax = "proto2";
package exclude;`,
		},
		{
			path:    "exclude/a.pb.go",
			content: `package exclude`,
		},
		{
			path: "exclude/b.proto",
			content: `syntax = "proto2";
package exclude;
`,
		},
		{
			path:    "exclude/b.pb.go",
			content: `package exclude`,
		},
	}
	want := []*packages.Package{
		{
			Name:       "exclude",
			Rel:        "exclude",
			ImportPath: "example.com/repo/exclude",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"a.pb.go"},
				},
			},
			Proto: packages.ProtoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"b.proto"},
				},
				HasPbGo: true,
			},
		},
	}
	checkFiles(t, files, "example.com/repo", want)
}

func TestLegacyProtos(t *testing.T) {
	files := []fileSpec{
		{
			path:    "BUILD",
			content: `# gazelle:proto legacy`,
		}, {
			path: "have_pbgo/a.proto",
			content: `syntax = "proto2";
package have_pbgo;`,
		}, {
			path:    "have_pbgo/a.pb.go",
			content: `package have_pbgo`,
		}, {
			path: "no_pbgo/b.proto",
			content: `syntax = "proto2";
package no_pbgo;`,
		}, {
			path:    "no_pbgo/other.go",
			content: `package no_pbgo`,
		}, {
			path: "proto_only/c.proto",
			content: `syntax = "proto2";
package proto_only;`,
		},
	}
	want := []*packages.Package{
		{
			Name:       "have_pbgo",
			Rel:        "have_pbgo",
			ImportPath: "example.com/repo/have_pbgo",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"a.pb.go"},
				},
			},
			Proto: packages.ProtoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"a.proto"},
				},
				HasPbGo: true,
			},
		}, {
			Name:       "no_pbgo",
			Rel:        "no_pbgo",
			ImportPath: "example.com/repo/no_pbgo",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"other.go"},
				},
			},
			Proto: packages.ProtoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"b.proto"},
				},
				HasPbGo: false,
			},
		},
	}
	checkFiles(t, files, "example.com/repo", want)
}

func TestMalformedBuildFile(t *testing.T) {
	files := []fileSpec{
		{path: "BUILD", content: "????"},
		{path: "foo.go", content: "package foo"},
	}
	want := []*packages.Package{}
	checkFiles(t, files, "example.com/repo", want)
}

func TestMultipleBuildFiles(t *testing.T) {
	files := []fileSpec{
		{path: "BUILD"},
		{path: "BUILD.bazel"},
		{path: "foo.go", content: "package foo"},
	}
	want := []*packages.Package{}
	checkFiles(t, files, "example.com/repo", want)
}

func TestMalformedGoFile(t *testing.T) {
	files := []fileSpec{
		{path: "a.go", content: "pakcage foo"},
		{path: "b.go", content: "package foo"},
	}
	want := []*packages.Package{
		{
			Name:       "foo",
			ImportPath: "example.com/repo",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"a.go", "b.go"},
				},
			},
		},
	}
	checkFiles(t, files, "example.com/repo", want)
}

func TestSymlinksBasic(t *testing.T) {
	files := []fileSpec{
		{path: "root/a.go", content: "package a"},
		{path: "root/b", symlink: "../b"},   // symlink outside repo is followed
		{path: "root/c", symlink: "c"},      // symlink inside repo is not followed.
		{path: "root/d", symlink: "../b/d"}, // symlink under root/b not followed
		{path: "root/e", symlink: "../e"},
		{path: "c/c.go", symlink: "package c"},
		{path: "b/b.go", content: "package b"},
		{path: "b/d/d.go", content: "package d"},
		{path: "e/loop", symlink: "loop2"}, // symlink loop
		{path: "e/loop2", symlink: "loop"},
	}
	dir, err := createFiles(files)
	if err != nil {
		t.Fatalf("createFiles() failed with %v; want success", err)
	}
	want := []*packages.Package{
		{
			Name:       "d",
			Dir:        dir + "/root/b/d",
			Rel:        "b/d",
			ImportPath: "example.com/repo/b/d",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"d.go"},
				},
			},
		},
		{
			Name:       "b",
			Dir:        dir + "/root/b",
			Rel:        "b",
			ImportPath: "example.com/repo/b",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"b.go"},
				},
			},
		},
		{
			Name:       "a",
			Dir:        dir + "/root",
			ImportPath: "example.com/repo",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"a.go"},
				},
			},
		},
	}
	c := &config.Config{
		RepoRoot:            dir + "/root",
		GoPrefix:            "example.com/repo",
		Dirs:                []string{dir + "/root"},
		ValidBuildFileNames: config.DefaultValidBuildFileNames,
	}
	got := walkPackages(c)
	checkPackages(t, got, want)
}

func TestSymlinksIgnore(t *testing.T) {
	files := []fileSpec{
		{
			path:    "root/BUILD",
			content: "# gazelle:exclude b",
		},
		{path: "root/b", symlink: "../b"},
		{path: "b/b.go", content: "package b"},
	}
	dir, err := createFiles(files)
	if err != nil {
		t.Fatalf("createFiles() failed with %v; want success", err)
	}
	want := []*packages.Package{}
	c := &config.Config{
		RepoRoot:            dir + "/root",
		GoPrefix:            "example.com/repo",
		Dirs:                []string{dir + "/root"},
		ValidBuildFileNames: config.DefaultValidBuildFileNames,
	}
	got := walkPackages(c)
	checkPackages(t, got, want)
}

func TestSymlinksMixIgnoredAndNonIgnored(t *testing.T) {
	files := []fileSpec{
		{
			path:    "root/BUILD",
			content: "# gazelle:exclude b",
		},
		{path: "root/b", symlink: "../b"},  // ignored
		{path: "root/b2", symlink: "../b"}, // not ignored
		{path: "b/b.go", content: "package b"},
	}
	dir, err := createFiles(files)
	if err != nil {
		t.Fatalf("createFiles() failed with %v; want success", err)
	}
	want := []*packages.Package{
		{
			Name:       "b",
			Dir:        dir + "/root/b2",
			Rel:        "b2",
			ImportPath: "example.com/repo/b2",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"b.go"},
				},
			},
		},
	}
	c := &config.Config{
		RepoRoot:            dir + "/root",
		GoPrefix:            "example.com/repo",
		Dirs:                []string{dir + "/root"},
		ValidBuildFileNames: config.DefaultValidBuildFileNames,
	}
	got := walkPackages(c)
	checkPackages(t, got, want)
}

func TestSymlinksChained(t *testing.T) {
	files := []fileSpec{
		{path: "root/b", symlink: "../link0"},
		{path: "link0", symlink: "b"},
		{path: "root/b2", symlink: "../b"},
		{path: "b/b.go", content: "package b"},
	}
	dir, err := createFiles(files)
	if err != nil {
		t.Fatalf("createFiles() failed with %v; want success", err)
	}
	want := []*packages.Package{
		{
			Name:       "b",
			Dir:        dir + "/root/b",
			Rel:        "b",
			ImportPath: "example.com/repo/b",
			Library: packages.GoTarget{
				Sources: packages.PlatformStrings{
					Generic: []string{"b.go"},
				},
			},
		},
	}
	c := &config.Config{
		RepoRoot:            dir + "/root",
		GoPrefix:            "example.com/repo",
		Dirs:                []string{dir + "/root"},
		ValidBuildFileNames: config.DefaultValidBuildFileNames,
	}
	got := walkPackages(c)
	checkPackages(t, got, want)
}

func TestSymlinksDangling(t *testing.T) {
	files := []fileSpec{
		{path: "root/b", symlink: "../b"},
	}
	dir, err := createFiles(files)
	if err != nil {
		t.Fatalf("createFiles() failed with %v; want success", err)
	}
	want := []*packages.Package{}
	c := &config.Config{
		RepoRoot:            dir + "/root",
		GoPrefix:            "example.com/repo",
		Dirs:                []string{dir + "/root"},
		ValidBuildFileNames: config.DefaultValidBuildFileNames,
	}
	got := walkPackages(c)
	checkPackages(t, got, want)
}
