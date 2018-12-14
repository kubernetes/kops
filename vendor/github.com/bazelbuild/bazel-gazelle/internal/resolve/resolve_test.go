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

package resolve

import (
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/bazelbuild/bazel-gazelle/internal/config"
	"github.com/bazelbuild/bazel-gazelle/internal/label"
	bf "github.com/bazelbuild/buildtools/build"
)

func TestResolveGoIndex(t *testing.T) {
	c := &config.Config{
		GoPrefix: "example.com/repo",
		DepMode:  config.VendorMode,
	}
	l := label.NewLabeler(c)

	type fileSpec struct {
		rel, content string
	}
	type testCase struct {
		desc       string
		buildFiles []fileSpec
		imp        string
		from       label.Label
		wantErr    string
		want       label.Label
	}
	for _, tc := range []testCase{
		{
			desc: "no_match",
			imp:  "example.com/foo",
			// fall back to external resolver
			want: label.New("", "vendor/example.com/foo", config.DefaultLibName),
		}, {
			desc: "simple",
			buildFiles: []fileSpec{{
				rel: "foo",
				content: `
go_library(
    name = "go_default_library",
    importpath = "example.com/foo",
)
`}},
			imp:  "example.com/foo",
			want: label.New("", "foo", "go_default_library"),
		}, {
			desc: "test_and_library_not_indexed",
			buildFiles: []fileSpec{{
				rel: "foo",
				content: `
go_test(
    name = "go_default_test",
    importpath = "example.com/foo",
)

go_binary(
    name = "cmd",
    importpath = "example.com/foo",
)
`,
			}},
			imp: "example.com/foo",
			// fall back to external resolver
			want: label.New("", "vendor/example.com/foo", config.DefaultLibName),
		}, {
			desc: "multiple_rules_ambiguous",
			buildFiles: []fileSpec{{
				rel: "foo",
				content: `
go_library(
    name = "a",
    importpath = "example.com/foo",
)

go_library(
    name = "b",
    importpath = "example.com/foo",
)
`,
			}},
			imp:     "example.com/foo",
			wantErr: "multiple rules",
		}, {
			desc: "vendor_not_visible",
			buildFiles: []fileSpec{
				{
					rel: "",
					content: `
go_library(
    name = "root",
    importpath = "example.com/foo",
)
`,
				}, {
					rel: "a/vendor/foo",
					content: `
go_library(
    name = "vendored",
    importpath = "example.com/foo",
)
`,
				},
			},
			imp:  "example.com/foo",
			from: label.New("", "b", "b"),
			want: label.New("", "", "root"),
		}, {
			desc: "vendor_supercedes_nonvendor",
			buildFiles: []fileSpec{
				{
					rel: "",
					content: `
go_library(
    name = "root",
    importpath = "example.com/foo",
)
`,
				}, {
					rel: "vendor/foo",
					content: `
go_library(
    name = "vendored",
    importpath = "example.com/foo",
)
`,
				},
			},
			imp:  "example.com/foo",
			from: label.New("", "sub", "sub"),
			want: label.New("", "vendor/foo", "vendored"),
		}, {
			desc: "deep_vendor_shallow_vendor",
			buildFiles: []fileSpec{
				{
					rel: "shallow/vendor",
					content: `
go_library(
    name = "shallow",
    importpath = "example.com/foo",
)
`,
				}, {
					rel: "shallow/deep/vendor",
					content: `
go_library(
    name = "deep",
    importpath = "example.com/foo",
)
`,
				},
			},
			imp:  "example.com/foo",
			from: label.New("", "shallow/deep", "deep"),
			want: label.New("", "shallow/deep/vendor", "deep"),
		}, {
			desc: "nested_vendor",
			buildFiles: []fileSpec{
				{
					rel: "vendor/a",
					content: `
go_library(
    name = "a",
    importpath = "a",
)
`,
				}, {
					rel: "vendor/b/vendor/a",
					content: `
go_library(
    name = "a",
    importpath = "a",
)
`,
				},
			},
			imp:  "a",
			from: label.New("", "vendor/b/c", "c"),
			want: label.New("", "vendor/b/vendor/a", "a"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			ix := NewRuleIndex()
			for _, fs := range tc.buildFiles {
				f, err := bf.Parse(path.Join(fs.rel, "BUILD.bazel"), []byte(fs.content))
				if err != nil {
					t.Fatal(err)
				}
				ix.AddRulesFromFile(c, f)
			}

			ix.Finish()

			r := NewResolver(c, l, ix, nil)
			got, err := r.resolveGo(tc.imp, tc.from)
			if err != nil {
				if tc.wantErr == "" {
					t.Fatal(err)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("got %q ; want %q", err.Error(), tc.wantErr)
				}
				return
			}
			if err == nil && tc.wantErr != "" {
				t.Fatalf("got success ; want error %q", tc.wantErr)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v ; want %v", got, tc.want)
			}
		})
	}
}

func TestResolveProtoIndex(t *testing.T) {
	c := &config.Config{
		GoPrefix: "example.com/repo",
		DepMode:  config.VendorMode,
	}
	l := label.NewLabeler(c)

	buildContent := []byte(`
proto_library(
    name = "foo_proto",
    srcs = ["bar.proto"],
)

go_proto_library(
    name = "foo_go_proto",
    importpath = "example.com/foo",
    proto = ":foo_proto",
)

go_library(
    name = "embed",
    embed = [":foo_go_proto"],
    importpath = "example.com/foo",
)
`)
	f, err := bf.Parse(filepath.Join("sub", "BUILD.bazel"), buildContent)
	if err != nil {
		t.Fatal(err)
	}

	ix := NewRuleIndex()
	ix.AddRulesFromFile(c, f)
	ix.Finish()
	r := NewResolver(c, l, ix, nil)

	wantProto := label.New("", "sub", "foo_proto")
	if got, err := r.resolveProto("sub/bar.proto", label.New("", "baz", "baz")); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(got, wantProto) {
		t.Errorf("resolveProto: got %s ; want %s", got, wantProto)
	}
	_, err = r.resolveProto("sub/bar.proto", label.New("", "sub", "foo_proto"))
	if _, ok := err.(selfImportError); !ok {
		t.Errorf("resolveProto: got %v ; want selfImportError", err)
	}

	wantGoProto := label.New("", "sub", "embed")
	if got, err := r.resolveGoProto("sub/bar.proto", label.New("", "baz", "baz")); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(got, wantGoProto) {
		t.Errorf("resolveGoProto: got %s ; want %s", got, wantGoProto)
	}
	_, err = r.resolveGoProto("sub/bar.proto", label.New("", "sub", "foo_go_proto"))
	if _, ok := err.(selfImportError); !ok {
		t.Errorf("resolveGoProto: got %v ; want selfImportError", err)
	}
}

func TestResolveGoLocal(t *testing.T) {
	for _, spec := range []struct {
		importpath string
		from, want label.Label
	}{
		{
			importpath: "example.com/repo",
			want:       label.New("", "", config.DefaultLibName),
		}, {
			importpath: "example.com/repo/lib",
			want:       label.New("", "lib", config.DefaultLibName),
		}, {
			importpath: "example.com/repo/another",
			want:       label.New("", "another", config.DefaultLibName),
		}, {
			importpath: "example.com/repo",
			want:       label.New("", "", config.DefaultLibName),
		}, {
			importpath: "example.com/repo/lib/sub",
			want:       label.New("", "lib/sub", config.DefaultLibName),
		}, {
			importpath: "example.com/repo/another",
			want:       label.New("", "another", config.DefaultLibName),
		}, {
			importpath: "../y",
			from:       label.New("", "x", "x"),
			want:       label.New("", "y", config.DefaultLibName),
		},
	} {
		c := &config.Config{GoPrefix: "example.com/repo"}
		l := label.NewLabeler(c)
		ix := NewRuleIndex()
		r := NewResolver(c, l, ix, nil)
		label, err := r.resolveGo(spec.importpath, spec.from)
		if err != nil {
			t.Errorf("r.resolveGo(%q) failed with %v; want success", spec.importpath, err)
			continue
		}
		if got, want := label, spec.want; !reflect.DeepEqual(got, want) {
			t.Errorf("r.resolveGo(%q) = %s; want %s", spec.importpath, got, want)
		}
	}
}

func TestResolveGoLocalError(t *testing.T) {
	c := &config.Config{GoPrefix: "example.com/repo"}
	l := label.NewLabeler(c)
	ix := NewRuleIndex()
	rc := newStubRemoteCache(nil)
	r := NewResolver(c, l, ix, rc)

	for _, importpath := range []string{
		"fmt",
		"unknown.com/another",
		"unknown.com/another/sub",
		"unknown.com/repo_suffix",
	} {
		if l, err := r.resolveGo(importpath, label.NoLabel); err == nil {
			t.Errorf("r.resolveGo(%q) = %s; want error", importpath, l)
		}
	}

	if l, err := r.resolveGo("..", label.NoLabel); err == nil {
		t.Errorf("r.resolveGo(%q) = %s; want error", "..", l)
	}
}

func TestResolveGoEmptyPrefix(t *testing.T) {
	c := &config.Config{}
	l := label.NewLabeler(c)
	ix := NewRuleIndex()
	r := NewResolver(c, l, ix, nil)

	imp := "foo"
	want := label.New("", "foo", config.DefaultLibName)
	if got, err := r.resolveGo(imp, label.NoLabel); err != nil {
		t.Errorf("r.resolveGo(%q) failed with %v; want success", imp, err)
	} else if !reflect.DeepEqual(got, want) {
		t.Errorf("r.resolveGo(%q) = %s; want %s", imp, got, want)
	}

	imp = "fmt"
	if _, err := r.resolveGo(imp, label.NoLabel); err == nil {
		t.Errorf("r.resolveGo(%q) succeeded; want failure", imp)
	}
}

func TestResolveProto(t *testing.T) {
	prefix := "example.com/repo"
	for _, tc := range []struct {
		desc, imp              string
		from                   label.Label
		depMode                config.DependencyMode
		wantProto, wantGoProto label.Label
	}{
		{
			desc:        "root",
			imp:         "foo.proto",
			wantProto:   label.New("", "", "repo_proto"),
			wantGoProto: label.New("", "", config.DefaultLibName),
		}, {
			desc:        "sub",
			imp:         "foo/bar/bar.proto",
			wantProto:   label.New("", "foo/bar", "bar_proto"),
			wantGoProto: label.New("", "foo/bar", config.DefaultLibName),
		}, {
			desc:        "vendor",
			depMode:     config.VendorMode,
			imp:         "foo/bar/bar.proto",
			from:        label.New("", "vendor", ""),
			wantProto:   label.New("", "foo/bar", "bar_proto"),
			wantGoProto: label.New("", "vendor/foo/bar", config.DefaultLibName),
		}, {
			desc:        "well known",
			imp:         "google/protobuf/any.proto",
			wantProto:   label.New("com_google_protobuf", "", "any_proto"),
			wantGoProto: label.NoLabel,
		}, {
			desc:        "well known vendor",
			depMode:     config.VendorMode,
			imp:         "google/protobuf/any.proto",
			wantProto:   label.New("com_google_protobuf", "", "any_proto"),
			wantGoProto: label.NoLabel,
		}, {
			desc:        "descriptor",
			imp:         "google/protobuf/descriptor.proto",
			wantProto:   label.New("com_google_protobuf", "", "descriptor_proto"),
			wantGoProto: label.NoLabel,
		}, {
			desc:        "descriptor vendor",
			depMode:     config.VendorMode,
			imp:         "google/protobuf/descriptor.proto",
			wantProto:   label.New("com_google_protobuf", "", "descriptor_proto"),
			wantGoProto: label.NoLabel,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			c := &config.Config{
				GoPrefix: prefix,
				DepMode:  tc.depMode,
			}
			l := label.NewLabeler(c)
			ix := NewRuleIndex()
			r := NewResolver(c, l, ix, nil)

			got, err := r.resolveProto(tc.imp, tc.from)
			if err != nil {
				t.Errorf("resolveProto: got error %v; want success", err)
			}
			if !reflect.DeepEqual(got, tc.wantProto) {
				t.Errorf("resolveProto: got %s; want %s", got, tc.wantProto)
			}

			got, err = r.resolveGoProto(tc.imp, tc.from)
			if err != nil {
				if tc.wantGoProto != label.NoLabel {
					t.Errorf("resolveGoProto: got error %v; want %s", got, tc.wantGoProto)
				} else if _, ok := err.(standardImportError); !ok {
					t.Errorf("resolveGoProto: got error %v; want standardImportError", err)
				}
			}
			if !got.Equal(tc.wantGoProto) {
				t.Errorf("resolveGoProto: got %s; want %s", got, tc.wantGoProto)
			}
		})
	}
}

func TestResolveGoWKT(t *testing.T) {
	c := &config.Config{}
	l := label.NewLabeler(c)
	ix := NewRuleIndex()
	r := NewResolver(c, l, ix, nil)

	for _, tc := range []struct {
		imp, want string
	}{
		{"github.com/golang/protobuf/ptypes/any", "any_go_proto"},
		{"github.com/golang/protobuf/ptypes/api", "api_go_proto"},
		{"github.com/golang/protobuf/protoc-gen-go/descriptor", "descriptor_go_proto"},
		{"github.com/golang/protobuf/ptypes/duration", "duration_go_proto"},
		{"github.com/golang/protobuf/ptypes/empty", "empty_go_proto"},
		{"google.golang.org/genproto/protobuf/field_mask", "field_mask_go_proto"},
		{"google.golang.org/genproto/protobuf/source_context", "source_context_go_proto"},
		{"github.com/golang/protobuf/ptypes/struct", "struct_go_proto"},
		{"github.com/golang/protobuf/ptypes/timestamp", "timestamp_go_proto"},
		{"github.com/golang/protobuf/ptypes/wrappers", "wrappers_go_proto"},
		{"github.com/golang/protobuf/protoc-gen-go/plugin", "compiler_plugin_go_proto"},
		{"google.golang.org/genproto/protobuf/ptype", "type_go_proto"},
	} {
		t.Run(tc.want, func(t *testing.T) {
			want := label.Label{
				Repo: config.RulesGoRepoName,
				Pkg:  config.WellKnownTypesPkg,
				Name: tc.want,
			}
			if got, err := r.resolveGo(tc.imp, label.NoLabel); err != nil {
				t.Error(err)
			} else if !got.Equal(want) {
				t.Errorf("got %s; want %s", got, want)
			}
		})
	}
}

func TestResolveGoSkipEmbeds(t *testing.T) {
	c := &config.Config{}
	l := label.NewLabeler(c)
	ix := NewRuleIndex()
	r := NewResolver(c, l, ix, nil)

	f, err := bf.Parse("(test)", []byte(`
load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["lib.go"],
    importpath = "example.com/repo/lib",
)

go_test(
    name = "go_default_test",
    embed = [":go_default_library"],
    _gazelle_imports = [
        "example.com/repo/lib",
    ],
)
`))
	if err != nil {
		t.Fatal(err)
	}
	ix.AddRulesFromFile(c, f)
	ix.Finish()
	test := f.Stmt[len(f.Stmt)-1]
	test = r.ResolveRule(test, "")
	testRule := bf.Rule{Call: test.(*bf.CallExpr)}
	testDeps := testRule.Attr("deps")
	if testDeps != nil {
		t.Errorf("got deps = %s; want nil", bf.FormatString(testDeps))
	}
}
