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

package version

import "testing"

func TestCompare(t *testing.T) {
	for _, tc := range []struct {
		x, y Version
		want int
	}{
		{
			x:    Version{1},
			y:    Version{1},
			want: 0,
		}, {
			x:    Version{1},
			y:    Version{2},
			want: -1,
		}, {
			x:    Version{2},
			y:    Version{1},
			want: 1,
		}, {
			x:    Version{1},
			y:    Version{1, 1},
			want: -1,
		}, {
			x:    Version{1, 1},
			y:    Version{1},
			want: 1,
		},
	} {
		if got := tc.x.Compare(tc.y); got != tc.want {
			t.Errorf("Compare(%s, %s): got %v, want %v", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestParseVersion(t *testing.T) {
	for _, tc := range []struct {
		str     string
		want    Version
		wantErr bool
	}{
		{
			str:     "",
			wantErr: true,
		}, {
			str:  "1",
			want: Version{1},
		}, {
			str:     "-1",
			wantErr: true,
		}, {
			str:  "0.1.2",
			want: Version{0, 1, 2},
		}, {
			str:  "0-suffix",
			want: Version{0},
		},
	} {
		if got, err := ParseVersion(tc.str); tc.wantErr && err == nil {
			t.Errorf("ParseVersion(%q): got %s, want error", tc.str, got)
		} else if !tc.wantErr && err != nil {
			t.Errorf("ParseVersion(%q): got %v, want success", tc.str, err)
		} else if got.Compare(tc.want) != 0 {
			t.Errorf("ParseVersion(%q): got %s, want %s", tc.str, got, tc.want)
		}
	}
}
