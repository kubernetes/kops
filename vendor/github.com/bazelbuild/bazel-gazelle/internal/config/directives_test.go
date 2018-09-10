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

package config

import (
	"reflect"
	"testing"

	bf "github.com/bazelbuild/buildtools/build"
)

func TestParseDirectives(t *testing.T) {
	for _, tc := range []struct {
		desc, content string
		want          []Directive
	}{
		{
			desc: "empty file",
		}, {
			desc: "locations",
			content: `# gazelle:ignore top

#gazelle:ignore before
foo(
   "foo",  # gazelle:ignore inside
) # gazelle:ignore suffix
#gazelle:ignore after

# gazelle:ignore bottom`,
			want: []Directive{
				{"ignore", "top"},
				{"ignore", "before"},
				{"ignore", "after"},
				{"ignore", "bottom"},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			f, err := bf.Parse("test.bazel", []byte(tc.content))
			if err != nil {
				t.Fatal(err)
			}

			got := ParseDirectives(f)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %#v ; want %#v", got, tc.want)
			}
		})
	}
}

func TestApplyDirectives(t *testing.T) {
	for _, tc := range []struct {
		desc       string
		directives []Directive
		rel        string
		want       Config
	}{
		{
			desc:       "empty build_tags",
			directives: []Directive{{"build_tags", ""}},
			want:       Config{},
		}, {
			desc:       "build_tags",
			directives: []Directive{{"build_tags", "foo,bar"}},
			want:       Config{GenericTags: BuildTags{"foo": true, "bar": true}},
		}, {
			desc:       "build_file_name",
			directives: []Directive{{"build_file_name", "foo,bar"}},
			want:       Config{ValidBuildFileNames: []string{"foo", "bar"}},
		}, {
			desc:       "prefix",
			directives: []Directive{{"prefix", "example.com/repo"}},
			rel:        "sub",
			want:       Config{GoPrefix: "example.com/repo", GoPrefixRel: "sub"},
		}, {
			desc:       "importmap_prefix",
			directives: []Directive{{"importmap_prefix", "example.com/repo"}},
			rel:        "sub",
			want:       Config{GoImportMapPrefix: "example.com/repo", GoImportMapPrefixRel: "sub"},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			c := &Config{}
			c.PreprocessTags()
			got := ApplyDirectives(c, tc.directives, tc.rel)
			tc.want.PreprocessTags()
			if !reflect.DeepEqual(*got, tc.want) {
				t.Errorf("got %#v ; want %#v", *got, tc.want)
			}
		})
	}
}

func TestInferProtoMode(t *testing.T) {
	for _, tc := range []struct {
		desc, content string
		c             Config
		rel           string
		want          ProtoMode
	}{
		{
			desc: "default",
		}, {
			desc: "previous",
			c:    Config{ProtoMode: LegacyProtoMode},
			want: LegacyProtoMode,
		}, {
			desc: "explicit",
			content: `# gazelle:proto default

load("@io_bazel_rules_go//proto:go_proto_library.bzl", "go_proto_library")
`,
			want: DefaultProtoMode,
		}, {
			desc:    "explicit_no_override",
			content: `load("@io_bazel_rules_go//proto:go_proto_library.bzl", "go_proto_library")`,
			c: Config{
				ProtoMode:         DefaultProtoMode,
				ProtoModeExplicit: true,
			},
			want: DefaultProtoMode,
		}, {
			desc: "vendor",
			rel:  "vendor",
			want: DisableProtoMode,
		}, {
			desc:    "legacy",
			content: `load("@io_bazel_rules_go//proto:go_proto_library.bzl", "go_proto_library")`,
			want:    LegacyProtoMode,
		}, {
			desc:    "disable",
			content: `load("@com_example_repo//proto:go_proto_library.bzl", go_proto_library = "x")`,
			want:    DisableProtoMode,
		}, {
			desc:    "fix legacy",
			content: `load("@io_bazel_rules_go//proto:go_proto_library.bzl", "go_proto_library")`,
			c:       Config{ShouldFix: true},
		}, {
			desc:    "fix disabled",
			content: `load("@com_example_repo//proto:go_proto_library.bzl", go_proto_library = "x")`,
			c:       Config{ShouldFix: true},
			want:    DisableProtoMode,
		}, {
			desc: "well known types",
			c:    Config{GoPrefix: "github.com/golang/protobuf"},
			want: LegacyProtoMode,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			var f *bf.File
			var directives []Directive
			if tc.content != "" {
				var err error
				f, err = bf.Parse("BUILD.bazel", []byte(tc.content))
				if err != nil {
					t.Fatalf("error parsing build file: %v", err)
				}
				directives = ParseDirectives(f)
			}

			got := InferProtoMode(&tc.c, tc.rel, f, directives)
			if got.ProtoMode != tc.want {
				t.Errorf("got proto mode %v ; want %v", got.ProtoMode, tc.want)
			}
		})
	}
}
