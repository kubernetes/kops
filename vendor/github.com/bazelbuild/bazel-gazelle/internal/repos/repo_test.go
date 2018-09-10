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

package repos

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	bf "github.com/bazelbuild/buildtools/build"
)

func TestGenerateRepoRules(t *testing.T) {
	repo := Repo{
		Name:     "org_golang_x_tools",
		GoPrefix: "golang.org/x/tools",
		Commit:   "123456",
	}
	got := bf.FormatString(GenerateRule(repo))
	want := `go_repository(
    name = "org_golang_x_tools",
    commit = "123456",
    importpath = "golang.org/x/tools",
)`
	if got != want {
		t.Errorf("got %s ; want %s", got, want)
	}
}

func TestFindExternalRepo(t *testing.T) {
	dir, err := ioutil.TempDir(os.Getenv("TEST_TEMPDIR"), "TestFindExternalRepo")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}

	name := "foo"
	externalPath := filepath.Join(dir, "bazel", "output-base", "external", name)
	if err := os.MkdirAll(externalPath, 0777); err != nil {
		t.Fatal(err)
	}

	bazelOutPath := filepath.Join(dir, "bazel", "output-base", "execroot", "test", "bazel-out")
	if err := os.MkdirAll(bazelOutPath, 0777); err != nil {
		t.Fatal(err)
	}

	workspacePath := filepath.Join(dir, "workspace")
	if err := os.MkdirAll(workspacePath, 0777); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(bazelOutPath, filepath.Join(workspacePath, "bazel-out")); err != nil {
		t.Fatal(err)
	}

	if got, err := FindExternalRepo(workspacePath, name); err != nil {
		t.Fatal(err)
	} else if got != externalPath {
		t.Errorf("got %q ; want %q", got, externalPath)
	}
}

func TestListRepositories(t *testing.T) {
	for _, tc := range []struct {
		desc, workspace string
		want            []Repo
	}{
		{
			desc: "empty",
			want: nil,
		}, {
			desc: "go_repository",
			workspace: `
go_repository(
    name = "custom_repo",
    commit = "123456",
    remote = "https://example.com/repo",
    importpath = "example.com/repo",
)
`,
			want: []Repo{{
				Name:     "custom_repo",
				GoPrefix: "example.com/repo",
				Remote:   "https://example.com/repo",
				Commit:   "123456",
			}},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			workspace, err := bf.Parse("WORKSPACE", []byte(tc.workspace))
			if err != nil {
				t.Fatal(err)
			}
			if got := ListRepositories(workspace); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %#v ; want %#v", got, tc.want)
			}
		})
	}
}
