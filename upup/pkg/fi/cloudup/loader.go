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

package cloudup

import (
	"context"
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/reflectutils"
)

type Loader struct {
	Builders []fi.ModelBuilder

	tasks map[string]fi.Task
}

func (l *Loader) Init() {
	l.tasks = make(map[string]fi.Task)
}

func (l *Loader) BuildTasks(ctx context.Context, lifecycleOverrides map[string]fi.Lifecycle) (map[string]fi.Task, error) {
	for _, builder := range l.Builders {
		context := &fi.ModelBuilderContext{
			Tasks:              l.tasks,
			LifecycleOverrides: lifecycleOverrides,
		}
		context = context.WithContext(ctx)
		err := builder.Build(context)
		if err != nil {
			return nil, err
		}
		l.tasks = context.Tasks
	}

	err := l.processDeferrals()
	if err != nil {
		return nil, err
	}
	return l.tasks, nil
}

func (l *Loader) processDeferrals() error {
	for taskKey, task := range l.tasks {
		taskValue := reflect.ValueOf(task)

		visitor := func(path *reflectutils.FieldPath, f *reflect.StructField, v reflect.Value) error {
			if reflectutils.IsPrimitiveValue(v) {
				return nil
			}

			if path.IsEmpty() {
				// Don't process top-level value
				return nil
			}

			switch v.Kind() {
			case reflect.Interface, reflect.Ptr:
				if v.CanInterface() && !v.IsNil() {
					// TODO: Can we / should we use a type-switch statement
					intf := v.Interface()
					if hn, ok := intf.(fi.HasName); ok {
						name := hn.GetName()
						if name != nil {
							typeNameForTask := fi.TypeNameForTask(intf)
							primary := l.tasks[typeNameForTask+"/"+*name]
							if primary == nil {
								keys := sets.NewString()
								for k := range l.tasks {
									keys.Insert(k)
								}
								klog.Infof("Known tasks:")
								for _, k := range keys.List() {
									klog.Infof("  %s", k)
								}

								return fmt.Errorf("unable to find task %q, referenced from %s:%s", typeNameForTask+"/"+*name, taskKey, path)
							}

							klog.V(11).Infof("Replacing task %q at %s:%s", *name, taskKey, path)
							v.Set(reflect.ValueOf(primary))
						}
						return reflectutils.SkipReflection
					}
				}
			}

			return nil
		}

		err := reflectutils.ReflectRecursive(taskValue, visitor, &reflectutils.ReflectOptions{DeprecatedDoubleVisit: true})
		if err != nil {
			return fmt.Errorf("unexpected error resolving task %q: %v", taskKey, err)
		}
	}

	return nil
}

func (l *Loader) FindDeletions(cloud fi.Cloud, lifecycleOverrides map[string]fi.Lifecycle) (map[string]fi.Task, error) {
	for _, builder := range l.Builders {
		if hasDeletions, ok := builder.(fi.HasDeletions); ok {
			context := &fi.ModelBuilderContext{
				Tasks:              l.tasks,
				LifecycleOverrides: lifecycleOverrides,
			}
			if err := hasDeletions.FindDeletions(context, cloud); err != nil {
				return nil, err
			}
			l.tasks = context.Tasks
		}
	}
	return l.tasks, nil
}
