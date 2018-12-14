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
	"reflect"
	"testing"

	"github.com/bazelbuild/bazel-gazelle/internal/config"
)

func TestAddPlatformStrings(t *testing.T) {
	c := &config.Config{}
	for _, tc := range []struct {
		desc, filename string
		tags           []tagLine
		want           PlatformStrings
	}{
		{
			desc:     "generic",
			filename: "foo.go",
			want: PlatformStrings{
				Generic: []string{"foo.go"},
			},
		}, {
			desc:     "os",
			filename: "foo_linux.go",
			want: PlatformStrings{
				OS: map[string][]string{"linux": []string{"foo_linux.go"}},
			},
		}, {
			desc:     "arch",
			filename: "foo_amd64.go",
			want: PlatformStrings{
				Arch: map[string][]string{"amd64": []string{"foo_amd64.go"}},
			},
		}, {
			desc:     "os and arch",
			filename: "foo_linux_amd64.go",
			want: PlatformStrings{
				Platform: map[config.Platform][]string{
					config.Platform{OS: "linux", Arch: "amd64"}: []string{"foo_linux_amd64.go"},
				},
			},
		}, {
			desc:     "os not arch",
			filename: "foo.go",
			tags:     []tagLine{{{"solaris", "!arm"}}},
			want: PlatformStrings{
				Platform: map[config.Platform][]string{
					config.Platform{OS: "solaris", Arch: "amd64"}: []string{"foo.go"},
				},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			fi := fileNameInfo("", "", tc.filename)
			fi.tags = tc.tags
			var sb platformStringsBuilder
			add := getPlatformStringsAddFunction(c, fi, nil)
			add(&sb, tc.filename)
			got := sb.build()
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %#v ; want %#v", got, tc.want)
			}
		})
	}
}

func TestDuplicatePlatformStrings(t *testing.T) {
	for _, tc := range []struct {
		desc string
		add  func(sb *platformStringsBuilder)
		want PlatformStrings
	}{
		{
			desc: "both generic",
			add: func(sb *platformStringsBuilder) {
				sb.addGenericString("a")
				sb.addGenericString("a")
			},
			want: PlatformStrings{
				Generic: []string{"a"},
			},
		}, {
			desc: "os generic",
			add: func(sb *platformStringsBuilder) {
				sb.addOSString("a", []string{"linux"})
				sb.addGenericString("a")
			},
			want: PlatformStrings{
				Generic: []string{"a"},
			},
		}, {
			desc: "os arch",
			add: func(sb *platformStringsBuilder) {
				sb.addOSString("a", []string{"solaris"})
				sb.addArchString("a", []string{"mips"})
			},
			want: PlatformStrings{
				Platform: map[config.Platform][]string{
					config.Platform{OS: "solaris", Arch: "amd64"}: {"a"},
					config.Platform{OS: "linux", Arch: "mips"}:    {"a"},
				},
			},
		}, {
			desc: "platform os",
			add: func(sb *platformStringsBuilder) {
				sb.addPlatformString("a", []config.Platform{{OS: "linux", Arch: "mips"}})
				sb.addOSString("a", []string{"solaris"})
			},
			want: PlatformStrings{
				Platform: map[config.Platform][]string{
					config.Platform{OS: "solaris", Arch: "amd64"}: {"a"},
					config.Platform{OS: "linux", Arch: "mips"}:    {"a"},
				},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			var sb platformStringsBuilder
			tc.add(&sb)
			got := sb.build()
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %#v ; want %#v", got, tc.want)
			}
		})
	}
}
