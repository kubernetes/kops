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
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

type Loader struct {
	Builders []fi.ModelBuilder
}

// Build is responsible for running the build tasks for nodeup
func (l *Loader) Build() (map[string]fi.Task, error) {
	tasks := make(map[string]fi.Task)
	for _, builder := range l.Builders {
		context := &fi.ModelBuilderContext{
			Tasks: tasks,
		}
		err := builder.Build(context)
		if err != nil {
			return nil, err
		}
		tasks = context.Tasks
	}

	// If there is a package task, we need an update packages task
	for _, t := range tasks {
		if _, ok := t.(*nodetasks.Package); ok {
			klog.Infof("Package task found; adding UpdatePackages task")
			tasks["UpdatePackages"] = nodetasks.NewUpdatePackages()
			break
		}
	}
	if tasks["UpdatePackages"] == nil {
		klog.Infof("No package task found; won't update packages")
	}

	return tasks, nil
}
