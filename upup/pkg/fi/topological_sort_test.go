/*
Copyright 2026 The Kubernetes Authors.

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

package fi

import (
	"bytes"
	"slices"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
)

type testTaskWithSet struct {
	Members  sets.Set[string]
	Direct   *testTask
	Children map[string]*testTask
}

func (*testTaskWithSet) Run(*CloudupContext) error {
	return nil
}

func TestFindTaskDependenciesWithSets(t *testing.T) {
	for _, tt := range []struct {
		name    string
		members sets.Set[string]
	}{
		{name: "nil"},
		{name: "empty", members: sets.New[string]()},
		{name: "populated", members: sets.New("a", "b")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			state := klog.CaptureState()
			t.Cleanup(state.Restore)
			var warnings bytes.Buffer
			klog.LogToStderr(false)
			klog.SetOutputBySeverity("WARNING", &warnings)

			direct := &testTask{Name: new("direct")}
			child := &testTask{Name: new("child")}
			parent := &testTaskWithSet{
				Members:  tt.members,
				Direct:   direct,
				Children: map[string]*testTask{"child": child, "nil": nil},
			}
			dependencies := FindTaskDependencies(map[string]CloudupTask{
				"parent": parent,
				"direct": direct,
				"child":  child,
			})
			slices.Sort(dependencies["parent"])
			if !slices.Equal(dependencies["parent"], []string{"child", "direct"}) {
				t.Errorf("parent dependencies = %v, want [child direct]", dependencies["parent"])
			}
			for _, name := range []string{"child", "direct"} {
				if len(dependencies[name]) != 0 {
					t.Errorf("unexpected dependencies for %s: %v", name, dependencies[name])
				}
			}
			klog.Flush()
			if warnings.Len() != 0 {
				t.Errorf("unexpected warnings: %s", warnings.String())
			}
		})
	}
}
