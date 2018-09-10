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
package wspace

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestFind(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	tmp, err := ioutil.TempDir(os.Getenv("TEST_TEMPDIR"), "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	tmp, err = filepath.EvalSymlinks(tmp) // on macOS, TEST_TEMPDIR is a symlink
	if err != nil {
		t.Fatal(err)
	}
	if parent, err := Find(tmp); err == nil {
		t.Skipf("WORKSPACE visible in parent %q of tmp %q", parent, tmp)
	}

	if err := os.MkdirAll(filepath.Join(tmp, "base", "sub"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(tmp, "base", workspaceFile), nil, 0755); err != nil {
		t.Fatal(err)
	}

	tmpBase := filepath.Join(tmp, "base")
	for _, tc := range []struct {
		dir, want string // want == "" means an error is expected
	}{
		{tmp, ""},
		{tmpBase, tmpBase},
		{filepath.Join(tmpBase, "sub"), tmpBase},
	} {
		t.Run(tc.dir, func(t *testing.T) {
			if got, err := Find(tc.dir); err != nil && tc.want != "" {
				t.Errorf("in %s, Find(%q): got %v, want %q", wd, tc.dir, err, tc.want)
			} else if got != tc.want {
				t.Errorf("in %s, Find(%q): got %q, want %q", wd, tc.dir, got, tc.want)
			}
			if err := os.Chdir(tc.dir); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(wd)
			if got, err := Find("."); err != nil && tc.want != "" {
				t.Errorf(`in %s, Find("."): got %v, want %q`, tc.dir, err, tc.want)
			} else if got != tc.want {
				t.Errorf(`in %s, Find("."): got %q, want %q`, tc.dir, got, tc.want)
			}
		})
	}
}
