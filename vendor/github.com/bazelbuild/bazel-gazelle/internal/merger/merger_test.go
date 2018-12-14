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

package merger

import (
	"testing"

	bf "github.com/bazelbuild/buildtools/build"
)

// should fix
// * updated srcs from new
// * data and size preserved from old
// * load stmt fixed to those in use and sorted

type testCase struct {
	desc, previous, current, empty, expected string
}

var testCases = []testCase{
	{
		desc: "basic functionality",
		previous: `
load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_prefix", "go_test")

go_prefix("github.com/jr_hacker/tools")

go_library(
    name = "go_default_library",
    srcs = [
        "lex.go",
        "print.go",
        "debug.go",
    ],
)

go_test(
    name = "go_default_test",
    size = "small",
    srcs = [
        "gen_test.go",  # keep
        "parse_test.go",
    ],
    data = glob(["testdata/*"]),
    embed = [":go_default_library"],
)
`,
		current: `
load("@io_bazel_rules_go//go:def.bzl", "go_test", "go_library")

go_prefix("")

go_library(
    name = "go_default_library",
    srcs = [
        "lex.go",
        "print.go",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "parse_test.go",
        "print_test.go",
    ],
    embed = [":go_default_library"],
)
`,
		expected: `
load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_prefix", "go_test")

go_prefix("github.com/jr_hacker/tools")

go_library(
    name = "go_default_library",
    srcs = [
        "lex.go",
        "print.go",
    ],
)

go_test(
    name = "go_default_test",
    size = "small",
    srcs = [
        "gen_test.go",  # keep
        "parse_test.go",
        "print_test.go",
    ],
    data = glob(["testdata/*"]),
    embed = [":go_default_library"],
)
`}, {
		desc: "merge dicts",
		previous: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = select({
        "darwin_amd64": [
            "foo_darwin_amd64.go", # keep
            "bar_darwin_amd64.go",
        ],
        "linux_arm": [
            "foo_linux_arm.go", # keep
            "bar_linux_arm.go",
        ],
    }),
)
`,
		current: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = select({
        "linux_arm": ["baz_linux_arm.go"],
        "darwin_amd64": ["baz_darwin_amd64.go"],
        "//conditions:default": [],
    }),
)
`,
		expected: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = select({
        "darwin_amd64": [
            "foo_darwin_amd64.go",  # keep
            "baz_darwin_amd64.go",
        ],
        "linux_arm": [
            "foo_linux_arm.go",  # keep
            "baz_linux_arm.go",
        ],
        "//conditions:default": [],
    }),
)
`,
	}, {
		desc: "merge old dict with gen list",
		previous: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = select({
        "linux_arm": [
            "foo_linux_arm.go", # keep
            "bar_linux_arm.go", # keep
        ],
        "darwin_amd64": [
            "bar_darwin_amd64.go",
        ],
    }),
)
`,
		current: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["baz.go"],
)
`,
		expected: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "baz.go",
    ] + select({
        "linux_arm": [
            "foo_linux_arm.go",  # keep
            "bar_linux_arm.go",  # keep
        ],
    }),
)
`,
	}, {
		desc: "merge old list with gen dict",
		previous: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "foo.go", # keep
        "bar.go", # keep
    ],
)
`,
		current: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = select({
        "linux_arm": [
            "foo_linux_arm.go",
            "bar_linux_arm.go",
        ],
        "darwin_amd64": [
            "bar_darwin_amd64.go",
        ],
        "//conditions:default": [],
    }),
)
`,
		expected: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "foo.go",  # keep
        "bar.go",  # keep
    ] + select({
        "linux_arm": [
            "foo_linux_arm.go",
            "bar_linux_arm.go",
        ],
        "darwin_amd64": [
            "bar_darwin_amd64.go",
        ],
        "//conditions:default": [],
    }),
)
`,
	}, {
		desc: "merge old list and dict with gen list and dict",
		previous: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "foo.go",  # keep
        "bar.go",
    ] + select({
        "linux_arm": [
            "foo_linux_arm.go",  # keep
        ],
        "//conditions:default": [],
    }),
)
`,
		current: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["baz.go"] + select({
        "linux_arm": ["bar_linux_arm.go"],
        "darwin_amd64": ["foo_darwin_amd64.go"],
        "//conditions:default": [],
    }),
)
`,
		expected: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "foo.go",  # keep
        "baz.go",
    ] + select({
        "darwin_amd64": ["foo_darwin_amd64.go"],
        "linux_arm": [
            "foo_linux_arm.go",  # keep
            "bar_linux_arm.go",
        ],
        "//conditions:default": [],
    }),
)
`,
	}, {
		desc: "os and arch",
		previous: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "generic_1.go",
    ] + select({
        "@io_bazel_rules_go//go/platform:linux": [
            "os_linux.go",  # keep
        ],
        "//conditions:default": [],
    }),
)
`,
		current: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "generic_2.go",
    ] + select({
        "@io_bazel_rules_go//go/platform:amd64": ["arch_amd64.go"],
        "//conditions:default": [],
    }),
)
`,
		expected: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "generic_2.go",
    ] + select({
        "@io_bazel_rules_go//go/platform:linux": [
            "os_linux.go",  # keep
        ],
        "//conditions:default": [],
    }) + select({
        "@io_bazel_rules_go//go/platform:amd64": ["arch_amd64.go"],
        "//conditions:default": [],
    }),
)
`,
	}, {
		desc: "merge error keeps old",
		previous: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = glob(["*.go"]),
)
`,
		current: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["foo.go"],
)
`,
		expected: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = glob(["*.go"]),
)
`,
	}, {
		desc: "delete empty list",
		previous: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["deleted.go"],
)
`,
		current: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = select({
        "linux_arm": ["foo_linux_arm.go"],
    }),
)
`,
		expected: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = select({
        "linux_arm": ["foo_linux_arm.go"],
    }),
)
`,
	}, {
		desc: "delete empty dict",
		previous: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = select({
        "linux_arm": ["foo_linux_arm.go"],
        "//conditions:default": [],
    }),
)
`,
		current: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["foo.go"],
)
`,
		expected: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["foo.go"],
)
`,
	}, {
		desc: "delete empty attr",
		previous: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["foo.go"],
    embed = ["deleted"],
)
`,
		current: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["foo.go"],
)
`,
		expected: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["foo.go"],
)
`,
	}, {
		desc: "merge comments",
		previous: `
# load
load("@io_bazel_rules_go//go:def.bzl", "go_library")

# rule
go_library(
    # unmerged attr
    name = "go_default_library",
    # merged attr
    srcs = ["foo.go"],
)
`,
		current: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["foo.go"],
)
`,
		expected: `
# load
load("@io_bazel_rules_go//go:def.bzl", "go_library")

# rule
go_library(
    # unmerged attr
    name = "go_default_library",
    # merged attr
    srcs = ["foo.go"],
)
`,
	}, {
		desc: "preserve comments",
		previous: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "a.go",  # preserve
        "b.go",  # comments
    ],
)
`,
		current: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["a.go", "b.go"],
)
`,
		expected: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "a.go",  # preserve
        "b.go",  # comments
    ],
)
`,
	}, {
		desc: "merge copts and clinkopts",
		previous: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "cgo_default_library",
    copts = [
        "-O0",
        "-g",  # keep
    ],
    clinkopts = [
        "-lX11",
    ],
)
`,
		current: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "cgo_default_library",
    cgo = True,
    copts = [
        "-O2",
    ],
    clinkopts = [
        "-lpng",
    ],
)
`,
		expected: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "cgo_default_library",
    copts = [
        "-g",  # keep
        "-O2",
    ],
    clinkopts = [
        "-lpng",
    ],
    cgo = True,
)
`,
	}, {
		desc: "keep scalar attr",
		previous: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    embed = [":lib"],  # keep
)
`,
		current: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
)
`,
		expected: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    embed = [":lib"],  # keep
)
`,
	}, {
		desc: "don't delete list with keep",
		previous: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "one.go",  # keep
    ],
)
`,
		current: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
)
`,
		expected: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "one.go",  # keep
    ],
)
`,
	}, {
		desc: "keep list multiline",
		previous: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "one.go",  # keep
        "two.go",
    ],
)
`,
		current: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
)
`,
		expected: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "one.go",  # keep
    ],
)
`,
	}, {
		desc: "keep dict list multiline",
		previous: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = select({
        "darwin_amd64": [
            "one_darwin.go",  # keep
        ],
        "linux_arm": [
            "one_linux.go",  # keep
            "two_linux.go",
        ],
    }),
)
`,
		current: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
)
`,
		expected: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = select({
        "darwin_amd64": [
            "one_darwin.go",  # keep
        ],
        "linux_arm": [
            "one_linux.go",  # keep
        ],
    }),
)
`,
	}, {
		desc: "keep prevents delete",
		previous: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

# keep
go_library(
    name = "go_default_library",
    srcs = ["lib.go"],
)
`,
		empty: `
go_library(name = "go_default_library")
`,
		expected: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

# keep
go_library(
    name = "go_default_library",
    srcs = ["lib.go"],
)
`,
	}, {
		desc: "keep prevents merge",
		previous: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

# keep
go_library(
    name = "go_default_library",
    srcs = ["old.go"],
)
`,
		current: `
go_library(
    name = "go_default_library",
    srcs = ["new.go"],
)
`,
		expected: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

# keep
go_library(
    name = "go_default_library",
    srcs = ["old.go"],
)
`,
	}, {
		desc: "delete empty rule",
		previous: `
load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["lib.go"],
)

go_binary(
    name = "old",
    srcs = ["bin.go"],
    embed = [":go_default_library"],
    importpath = "foo",
)
`,
		current: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["lib.go"],
)
`,
		empty: `
go_binary(name = "old")
`,
		expected: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["lib.go"],
)
`,
	}, {
		desc: "don't delete kept rule",
		previous: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "lib.go",  # keep
    ],
)
`,
		empty: `go_library(name = "go_default_library")`,
		expected: `
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "lib.go",  # keep
    ],
)
`,
	}, {
		desc: "match and rename",
		previous: `
load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")
load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")

go_binary(
    name = "custom_bin",
    embed = [":custom_library"],
)

go_library(
    name = "custom_library",
    embed = [":custom_proto_library"],
    importpath = "example.com/repo/foo",
)

go_proto_library(
    name = "custom_proto_library",
    importpath = "example.com/repo/foo",
    proto = ":foo_proto",
)
`,
		current: `
go_binary(
    name = "bin",
    srcs = ["bin.go"],
    embed = [":go_default_library"],
)

go_library(
    name = "go_default_library",
    srcs = ["lib.go"],
    embed = [":foo_proto_library"],
    importpath = "example.com/repo/foo",
)

go_proto_library(
    name = "foo_proto_library",
    importpath = "example.com/repo/foo",
    proto = ":foo_proto",
)
`,
		expected: `
load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")
load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")

go_binary(
    name = "custom_bin",
    embed = [":custom_library"],
    srcs = ["bin.go"],
)

go_library(
    name = "custom_library",
    embed = [":custom_proto_library"],
    importpath = "example.com/repo/foo",
    srcs = ["lib.go"],
)

go_proto_library(
    name = "custom_proto_library",
    importpath = "example.com/repo/foo",
    proto = ":foo_proto",
)
`,
	},
}

func TestMergeFile(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			genFile, err := bf.Parse("current", []byte(tc.current))
			if err != nil {
				t.Fatalf("%s: %v", tc.desc, err)
			}
			f, err := bf.Parse("previous", []byte(tc.previous))
			if err != nil {
				t.Fatalf("%s: %v", tc.desc, err)
			}
			emptyFile, err := bf.Parse("empty", []byte(tc.empty))
			if err != nil {
				t.Fatalf("%s: %v", tc.desc, err)
			}
			MergeFile(genFile.Stmt, emptyFile.Stmt, f, PreResolveAttrs)
			FixLoads(f)

			want := tc.expected
			if len(want) > 0 && want[0] == '\n' {
				want = want[1:]
			}

			if got := string(bf.Format(f)); got != want {
				t.Fatalf("%s: got %s; want %s", tc.desc, got, want)
			}
		})
	}
}

func TestMatch(t *testing.T) {
	for _, tc := range []struct {
		desc, gen, old string
		wantIndex      int
		wantError      bool
	}{
		{
			desc:      "no_match",
			gen:       `go_library(name = "lib")`,
			wantIndex: -1,
		}, {
			desc:      "name_match",
			gen:       `go_library(name = "lib")`,
			old:       `go_library(name = "lib", srcs = ["lib.go"])`,
			wantIndex: 0,
		}, {
			desc:      "name_match_kind_different",
			gen:       `go_library(name = "lib")`,
			old:       `cc_library(name = "lib")`,
			wantError: true,
		}, {
			desc: "multiple_name_match",
			gen:  `go_library(name = "lib")`,
			old: `
go_library(name = "lib")
go_library(name = "lib")
`,
			wantError: true,
		}, {
			desc:      "attr_match",
			gen:       `go_library(name = "x", importpath = "foo")`,
			old:       `go_library(name = "y", importpath = "foo")`,
			wantIndex: 0,
		}, {
			desc: "multiple_attr_match",
			gen:  `go_library(name = "x", importpath = "foo")`,
			old: `
go_library(name = "y", importpath = "foo")
go_library(name = "z", importpath = "foo")
`,
			wantError: true,
		}, {
			desc:      "any_match",
			gen:       `go_binary(name = "x")`,
			old:       `go_binary(name = "y")`,
			wantIndex: 0,
		}, {
			desc: "multiple_any_match",
			gen:  `go_binary(name = "x")`,
			old: `
go_binary(name = "y")
go_binary(name = "z")
`,
			wantError: true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			genFile, err := bf.Parse("gen", []byte(tc.gen))
			if err != nil {
				t.Fatal(err)
			}
			oldFile, err := bf.Parse("old", []byte(tc.old))
			if err != nil {
				t.Fatal(err)
			}
			if _, gotIndex, gotErr := match(oldFile.Stmt, genFile.Stmt[0].(*bf.CallExpr)); gotErr != nil {
				if !tc.wantError {
					t.Fatalf("unexpected error: %v", gotErr)
				}
			} else if tc.wantError {
				t.Fatal("unexpected success")
			} else if gotIndex != tc.wantIndex {
				t.Fatalf("got index %d ; want %d", gotIndex, tc.wantIndex)
			}
		})
	}
}
