/*
Copyright 2019 The Kubernetes Authors.

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

package nodeup

import (
	"fmt"
	"reflect"

	"k8s.io/kops/upup/pkg/fi"
)

type Loader struct {
	Builders []fi.NodeupModelBuilder
}

// Build is responsible for running the build tasks for nodeup
func (l *Loader) Build() (map[string]fi.NodeupTask, error) {
	tasks := make(map[string]fi.NodeupTask)
	for _, builder := range l.Builders {
		context := &fi.NodeupModelBuilderContext{
			Tasks: tasks,
		}
		err := builder.Build(context)
		if err != nil {
			return nil, fmt.Errorf("building %s: %v", reflect.TypeOf(builder), err)
		}
		tasks = context.Tasks
	}

	return tasks, nil
}
