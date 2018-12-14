/* Copyright 2018 The Bazel Authors. All rights reserved.

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

package pathtools

import "testing"

func TestHasPrefix(t *testing.T) {
	for _, tc := range []struct {
		desc, path, prefix string
		want               bool
	}{
		{
			desc:   "empty prefix",
			path:   "home/jr_hacker",
			prefix: "",
			want:   true,
		}, {
			desc:   "partial prefix",
			path:   "home/jr_hacker",
			prefix: "home",
			want:   true,
		}, {
			desc:   "full prefix",
			path:   "home/jr_hacker",
			prefix: "home/jr_hacker",
			want:   true,
		}, {
			desc:   "too long",
			path:   "home",
			prefix: "home/jr_hacker",
			want:   false,
		}, {
			desc:   "partial component",
			path:   "home/jr_hacker",
			prefix: "home/jr_",
			want:   false,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			if got := HasPrefix(tc.path, tc.prefix); got != tc.want {
				t.Errorf("got %v ; want %v", got, tc.want)
			}
		})
	}
}

func TestTrimPrefix(t *testing.T) {
	for _, tc := range []struct {
		desc, path, prefix, want string
	}{
		{
			desc:   "empty prefix",
			path:   "home/jr_hacker",
			prefix: "",
			want:   "home/jr_hacker",
		}, {
			desc:   "partial prefix",
			path:   "home/jr_hacker",
			prefix: "home",
			want:   "jr_hacker",
		}, {
			desc:   "full prefix",
			path:   "home/jr_hacker",
			prefix: "home/jr_hacker",
			want:   "",
		}, {
			desc:   "partial component",
			path:   "home/jr_hacker",
			prefix: "home/jr_",
			want:   "home/jr_hacker",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			if got := TrimPrefix(tc.path, tc.prefix); got != tc.want {
				t.Errorf("got %q ; want %q", got, tc.want)
			}
		})
	}
}
