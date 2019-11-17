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
		parent fi.Task
		child  fi.Task
	}{
		{
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
		allTasks := make(map[string]fi.Task)
		allTasks["parent"] = g.parent
		allTasks["child"] = g.child

		deps := g.parent.(fi.HasDependencies).GetDependencies(allTasks)
		if len(deps) != 0 {
			t.Errorf("found unexpected dependencies for parent: %v %v", g.parent, deps)
		}

		childDeps := g.child.(fi.HasDependencies).GetDependencies(allTasks)
		if len(childDeps) != 1 {
			t.Errorf("found unexpected dependencies for child: %v %v", g.child, childDeps)
		}
	}
}
