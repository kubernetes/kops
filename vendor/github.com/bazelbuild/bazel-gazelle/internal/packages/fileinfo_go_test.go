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

package packages

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/bazelbuild/bazel-gazelle/internal/config"
)

func TestGoFileInfo(t *testing.T) {
	c := &config.Config{}
	dir := "."
	rel := ""
	for _, tc := range []struct {
		desc, name, source string
		want               fileInfo
	}{
		{
			"empty file",
			"foo.go",
			"package foo\n",
			fileInfo{
				packageName: "foo",
			},
		},
		{
			"xtest file",
			"foo_test.go",
			"package foo_test\n",
			fileInfo{
				packageName: "foo",
				isTest:      true,
			},
		},
		{
			"xtest suffix on non-test",
			"foo_xtest.go",
			"package foo_test\n",
			fileInfo{
				packageName: "foo_test",
				isTest:      false,
			},
		},
		{
			"single import",
			"foo.go",
			`package foo

import "github.com/foo/bar"
`,
			fileInfo{
				packageName: "foo",
				imports:     []string{"github.com/foo/bar"},
			},
		},
		{
			"multiple imports",
			"foo.go",
			`package foo

import (
	"github.com/foo/bar"
	x "github.com/local/project/y"
)
`,
			fileInfo{
				packageName: "foo",
				imports:     []string{"github.com/foo/bar", "github.com/local/project/y"},
			},
		},
		{
			"standard imports included",
			"foo.go",
			`package foo

import "fmt"
`,
			fileInfo{
				packageName: "foo",
				imports:     []string{"fmt"},
			},
		},
		{
			"cgo",
			"foo.go",
			`package foo

import "C"
`,
			fileInfo{
				packageName: "foo",
				isCgo:       true,
			},
		},
		{
			"build tags",
			"foo.go",
			`// +build linux darwin

// +build !ignore

package foo
`,
			fileInfo{
				packageName: "foo",
				tags:        []tagLine{{{"linux"}, {"darwin"}}, {{"!ignore"}}},
			},
		},
		{
			"build tags without blank line",
			"route.go",
			`// Copyright 2017

// +build darwin dragonfly freebsd netbsd openbsd

// Package route provides basic functions for the manipulation of
// packet routing facilities on BSD variants.
package route
`,
			fileInfo{
				packageName: "route",
				tags:        []tagLine{{{"darwin"}, {"dragonfly"}, {"freebsd"}, {"netbsd"}, {"openbsd"}}},
			},
		},
	} {
		if err := ioutil.WriteFile(tc.name, []byte(tc.source), 0600); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tc.name)

		got := goFileInfo(c, dir, rel, tc.name)

		// Clear fields we don't care about for testing.
		got = fileInfo{
			packageName: got.packageName,
			isTest:      got.isTest,
			imports:     got.imports,
			isCgo:       got.isCgo,
			tags:        got.tags,
		}

		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("case %q: got %#v; want %#v", tc.desc, got, tc.want)
		}
	}
}

func TestGoFileInfoFailure(t *testing.T) {
	dir := "."
	name := "foo_linux_amd64.go"
	if err := ioutil.WriteFile(name, []byte("pakcage foo"), 0600); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(name)

	c := &config.Config{}
	got := goFileInfo(c, dir, "", name)
	want := fileInfo{
		path:     filepath.Join(dir, name),
		name:     name,
		ext:      ".go",
		category: goExt,
		goos:     "linux",
		goarch:   "amd64",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v ; want %#v", got, want)
	}
}

func TestCgo(t *testing.T) {
	c := &config.Config{}
	dir := "."
	rel := ""
	for _, tc := range []struct {
		desc, source string
		want         fileInfo
	}{
		{
			"not cgo",
			"package foo\n",
			fileInfo{isCgo: false},
		},
		{
			"empty cgo",
			`package foo

import "C"
`,
			fileInfo{isCgo: true},
		},
		{
			"simple flags",
			`package foo

/*
#cgo CFLAGS: -O0
	#cgo CPPFLAGS: -O1
#cgo   CXXFLAGS:   -O2
#cgo LDFLAGS: -O3 -O4
*/
import "C"
`,
			fileInfo{
				isCgo: true,
				copts: []taggedOpts{
					{opts: "-O0"},
					{opts: "-O1"},
					{opts: "-O2"},
				},
				clinkopts: []taggedOpts{
					{opts: strings.Join([]string{"-O3", "-O4"}, OptSeparator)},
				},
			},
		},
		{
			"cflags with conditions",
			`package foo

/*
#cgo foo bar,!baz CFLAGS: -O0
*/
import "C"
`,
			fileInfo{
				isCgo: true,
				copts: []taggedOpts{
					{
						tags: tagLine{{"foo"}, {"bar", "!baz"}},
						opts: "-O0",
					},
				},
			},
		},
		{
			"slashslash comments",
			`package foo

// #cgo CFLAGS: -O0
// #cgo CFLAGS: -O1
import "C"
`,
			fileInfo{
				isCgo: true,
				copts: []taggedOpts{
					{opts: "-O0"},
					{opts: "-O1"},
				},
			},
		},
		{
			"comment above single import group",
			`package foo

/*
#cgo CFLAGS: -O0
*/
import ("C")
`,
			fileInfo{
				isCgo: true,
				copts: []taggedOpts{
					{opts: "-O0"},
				},
			},
		},
	} {
		path := "TestCgo.go"
		if err := ioutil.WriteFile(path, []byte(tc.source), 0600); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(path)

		got := goFileInfo(c, dir, rel, path)

		// Clear fields we don't care about for testing.
		got = fileInfo{isCgo: got.isCgo, copts: got.copts, clinkopts: got.clinkopts}

		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("case %q: got %#v; want %#v", tc.desc, got, tc.want)
		}
	}
}

// Copied from go/build build_test.go
var (
	expandSrcDirPath = filepath.Join(string(filepath.Separator)+"projects", "src", "add")
)

// Copied from go/build build_test.go
var expandSrcDirTests = []struct {
	input, expected string
}{
	{"-L ${SRCDIR}/libs -ladd", "-L /projects/src/add/libs -ladd"},
	{"${SRCDIR}/add_linux_386.a -pthread -lstdc++", "/projects/src/add/add_linux_386.a -pthread -lstdc++"},
	{"Nothing to expand here!", "Nothing to expand here!"},
	{"$", "$"},
	{"$$", "$$"},
	{"${", "${"},
	{"$}", "$}"},
	{"$FOO ${BAR}", "$FOO ${BAR}"},
	{"Find me the $SRCDIRECTORY.", "Find me the $SRCDIRECTORY."},
	{"$SRCDIR is missing braces", "$SRCDIR is missing braces"},
}

// Copied from go/build build_test.go
func TestExpandSrcDir(t *testing.T) {
	for _, test := range expandSrcDirTests {
		output, _ := expandSrcDir(test.input, expandSrcDirPath)
		if output != test.expected {
			t.Errorf("%q expands to %q with SRCDIR=%q when %q is expected", test.input, output, expandSrcDirPath, test.expected)
		} else {
			t.Logf("%q expands to %q with SRCDIR=%q", test.input, output, expandSrcDirPath)
		}
	}
}

func TestExpandSrcDirRepoRelative(t *testing.T) {
	repo, err := ioutil.TempDir(os.Getenv("TEST_TEMPDIR"), "repo")
	if err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(repo, "sub")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatal(err)
	}
	goFile := filepath.Join(sub, "sub.go")
	content := []byte(`package sub

/*
#cgo CFLAGS: -I${SRCDIR}/..
*/
import "C"
`)
	if err := ioutil.WriteFile(goFile, content, 0644); err != nil {
		t.Fatal(err)
	}
	c := &config.Config{
		RepoRoot: repo,
		GoPrefix: "example.com/repo",
	}
	got := buildPackage(c, sub, "sub", []string{"sub.go"}, nil, nil, false)
	want := &Package{
		Name:       "sub",
		Dir:        sub,
		Rel:        "sub",
		ImportPath: "example.com/repo/sub",
		Library: GoTarget{
			Sources: PlatformStrings{
				Generic: []string{"sub.go"},
			},
			COpts: PlatformStrings{
				Generic: []string{"-Isub/.."},
			},
			Cgo: true,
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v ; want %#v", got, want)
	}
}

// Copied from go/build build_test.go
func TestShellSafety(t *testing.T) {
	tests := []struct {
		input, srcdir, expected string
		result                  bool
	}{
		{"-I${SRCDIR}/../include", "/projects/src/issue 11868", "-I/projects/src/issue 11868/../include", true},
		{"-I${SRCDIR}", "wtf$@%", "-Iwtf$@%", true},
		{"-X${SRCDIR}/1,${SRCDIR}/2", "/projects/src/issue 11868", "-X/projects/src/issue 11868/1,/projects/src/issue 11868/2", true},
		{"-I/tmp -I/tmp", "/tmp2", "-I/tmp -I/tmp", false},
		{"-I/tmp", "/tmp/[0]", "-I/tmp", true},
		{"-I${SRCDIR}/dir", "/tmp/[0]", "-I/tmp/[0]/dir", false},
	}
	for _, test := range tests {
		output, ok := expandSrcDir(test.input, test.srcdir)
		if ok != test.result {
			t.Errorf("Expected %t while %q expands to %q with SRCDIR=%q; got %t", test.result, test.input, output, test.srcdir, ok)
		}
		if output != test.expected {
			t.Errorf("Expected %q while %q expands with SRCDIR=%q; got %q", test.expected, test.input, test.srcdir, output)
		}
	}
}
