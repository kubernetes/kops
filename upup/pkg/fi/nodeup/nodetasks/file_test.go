/*
Copyright 2017 The Kubernetes Authors.

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

package nodetasks

import (
	"testing"

	"k8s.io/kops/upup/pkg/fi"
)

func TestFileDependencies(t *testing.T) {
	parentFileName := "/dependedon"
	childFileName := "/dependent"

	grid := []struct {
		name   string
		parent fi.Task
		child  fi.Task
	}{
		{
			name: "user",
			parent: &UserTask{
				Name:  "owner",
				UID:   3,
				Shell: "/bin/shell",
				Home:  "/home/owner",
			},
			child: &File{
				Owner:    fi.String("owner"),
				Path:     childFileName,
				Contents: fi.NewStringResource("I depend on an owner"),
				Type:     FileType_File,
			},
		},
		{
			name: "parentDir",
			parent: &File{
				Path: parentFileName,
				Type: FileType_Directory,
			},
			child: &File{
				Path:     parentFileName + "/" + childFileName,
				Contents: fi.NewStringResource("I depend on my parent directory"),
				Type:     FileType_File,
			},
		},
		{
			name: "afterFiles",
			parent: &File{
				Path:     parentFileName,
				Contents: fi.NewStringResource("I am depended on by " + childFileName),
				Type:     FileType_File,
			},
			child: &File{
				AfterFiles: []string{parentFileName},
				Path:       childFileName,
				Contents:   fi.NewStringResource("I depend on " + parentFileName),
				Type:       FileType_File,
			},
		},
	}

	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			context := &fi.ModelBuilderContext{
				Tasks: make(map[string]fi.Task),
			}
			context.AddTask(g.parent)
			context.AddTask(g.child)

			if _, ok := g.parent.(fi.HasDependencies); ok {
				deps := g.parent.(fi.HasDependencies).GetDependencies(context.Tasks)
				if len(deps) != 0 {
					t.Errorf("found unexpected dependencies for parent: %v %v", g.parent, deps)
				}
			}

			childDeps := g.child.(fi.HasDependencies).GetDependencies(context.Tasks)
			if len(childDeps) != 1 {
				t.Errorf("found unexpected dependencies for child: %v %v", g.child, childDeps)
			}
		})
	}
}
