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

package label

import (
	"reflect"
	"testing"
)

func TestLabelString(t *testing.T) {
	for _, spec := range []struct {
		l    Label
		want string
	}{
		{
			l:    Label{Name: "foo"},
			want: "//:foo",
		}, {
			l:    Label{Pkg: "foo/bar", Name: "baz"},
			want: "//foo/bar:baz",
		}, {
			l:    Label{Pkg: "foo/bar", Name: "bar"},
			want: "//foo/bar",
		}, {
			l:    Label{Repo: "com_example_repo", Pkg: "foo/bar", Name: "baz"},
			want: "@com_example_repo//foo/bar:baz",
		}, {
			l:    Label{Repo: "com_example_repo", Pkg: "foo/bar", Name: "bar"},
			want: "@com_example_repo//foo/bar",
		}, {
			l:    Label{Relative: true, Name: "foo"},
			want: ":foo",
		},
	} {
		if got, want := spec.l.String(), spec.want; got != want {
			t.Errorf("%#v.String() = %q; want %q", spec.l, got, want)
		}
	}
}

func TestParse(t *testing.T) {
	for _, tc := range []struct {
		str     string
		want    Label
		wantErr bool
	}{
		{str: "", wantErr: true},
		{str: "@//:", wantErr: true},
		{str: "@//:a", wantErr: true},
		{str: "@a:b", wantErr: true},
		{str: ":a", want: Label{Name: "a", Relative: true}},
		{str: "a", want: Label{Name: "a", Relative: true}},
		{str: "//:a", want: Label{Name: "a", Relative: false}},
		{str: "//a", want: Label{Pkg: "a", Name: "a"}},
		{str: "//a/b", want: Label{Pkg: "a/b", Name: "b"}},
		{str: "//a:b", want: Label{Pkg: "a", Name: "b"}},
		{str: "@a//b", want: Label{Repo: "a", Pkg: "b", Name: "b"}},
		{str: "@a//b:c", want: Label{Repo: "a", Pkg: "b", Name: "c"}},
		{str: "//api_proto:api.gen.pb.go_checkshtest", want: Label{Pkg: "api_proto", Name: "api.gen.pb.go_checkshtest"}},
	} {
		got, err := Parse(tc.str)
		if err != nil && !tc.wantErr {
			t.Errorf("for string %q: got error %s ; want success", tc.str, err)
			continue
		}
		if err == nil && tc.wantErr {
			t.Errorf("for string %q: got label %s ; want error", tc.str, got)
			continue
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("for string %q: got %s ; want %s", tc.str, got, tc.want)
		}
	}
}
