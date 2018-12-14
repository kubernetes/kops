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
	"reflect"
	"testing"

	"github.com/bazelbuild/bazel-gazelle/internal/config"
)

func TestProtoRegexpGroupNames(t *testing.T) {
	names := protoRe.SubexpNames()
	nameMap := map[string]int{
		"import":     importSubexpIndex,
		"package":    packageSubexpIndex,
		"go_package": goPackageSubexpIndex,
		"service":    serviceSubexpIndex,
	}
	for name, index := range nameMap {
		if names[index] != name {
			t.Errorf("proto regexp subexp %d is %s ; want %s", index, names[index], name)
		}
	}
	if len(names)-1 != len(nameMap) {
		t.Errorf("proto regexp has %d groups ; want %d", len(names), len(nameMap))
	}
}

func TestProtoFileInfo(t *testing.T) {
	c := &config.Config{}
	dir := "."
	rel := ""
	for _, tc := range []struct {
		desc, name, proto string
		want              fileInfo
	}{
		{
			desc:  "empty",
			name:  "empty^file.proto",
			proto: "",
			want: fileInfo{
				packageName: "empty_file",
			},
		}, {
			desc:  "simple package",
			name:  "package.proto",
			proto: "package foo;",
			want: fileInfo{
				packageName: "foo",
			},
		}, {
			desc:  "full package",
			name:  "full.proto",
			proto: "package foo.bar.baz;",
			want: fileInfo{
				packageName: "foo_bar_baz",
			},
		}, {
			desc: "import simple",
			name: "imp.proto",
			proto: `import 'single.proto';
import "double.proto";`,
			want: fileInfo{
				packageName: "imp",
				imports:     []string{"double.proto", "single.proto"},
			},
		}, {
			desc: "import quote",
			name: "quote.proto",
			proto: `import '""\".proto"';
import "'.proto";`,
			want: fileInfo{
				packageName: "quote",
				imports:     []string{"\"\"\".proto\"", "'.proto"},
			},
		}, {
			desc:  "import escape",
			name:  "escape.proto",
			proto: `import '\n\012\x0a.proto';`,
			want: fileInfo{
				packageName: "escape",
				imports:     []string{"\n\n\n.proto"},
			},
		}, {
			desc: "import two",
			name: "two.proto",
			proto: `import "first.proto";
import "second.proto";`,
			want: fileInfo{
				packageName: "two",
				imports:     []string{"first.proto", "second.proto"},
			},
		}, {
			desc:  "go_package",
			name:  "gopkg.proto",
			proto: `option go_package = "github.com/example/project;projectpb";`,
			want: fileInfo{
				packageName: "projectpb",
				importPath:  "github.com/example/project",
			},
		}, {
			desc:  "go_package_simple",
			name:  "gopkg_simple.proto",
			proto: `option go_package = "bar";`,
			want: fileInfo{
				packageName: "bar",
				importPath:  "",
			},
		}, {
			desc:  "service",
			name:  "service.proto",
			proto: `service ChatService {}`,
			want: fileInfo{
				packageName: "service",
				hasServices: true,
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			if err := ioutil.WriteFile(tc.name, []byte(tc.proto), 0600); err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tc.name)

			got := protoFileInfo(c, dir, rel, tc.name)

			// Clear fields we don't care about for testing.
			got = fileInfo{
				packageName: got.packageName,
				imports:     got.imports,
				importPath:  got.importPath,
				hasServices: got.hasServices,
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %#v; want %#v", got, tc.want)
			}
		})
	}
}
