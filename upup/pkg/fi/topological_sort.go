/*
Copyright 2016 The Kubernetes Authors.

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
	"crypto/x509/pkix"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/pkg/tasks"
	"k8s.io/kops/upup/pkg/fi/utils"
	"reflect"
)

type HasDependencies interface {
	GetDependencies(tasks map[string]tasks.Task) []tasks.Task
}

// FindTaskDependencies returns a map from each task's key to the discovered list of dependencies
func FindTaskDependencies(taskMap map[string]tasks.Task) map[string][]string {
	taskToId := make(map[interface{}]string)
	for k, t := range taskMap {
		taskToId[t] = k
	}

	edges := make(map[string][]string)

	for k, task := range taskMap {
		var dependencies []tasks.Task
		if hd, ok := task.(HasDependencies); ok {
			dependencies = hd.GetDependencies(taskMap)
		} else {
			dependencies = reflectForDependencies(taskMap, task)
		}

		var dependencyKeys []string
		for _, dep := range dependencies {
			dependencyKey, found := taskToId[dep]
			if !found {
				glog.Fatalf("dependency not found: %v (dependency of %s: %v)", dep, k, task)
			}
			dependencyKeys = append(dependencyKeys, dependencyKey)
		}

		edges[k] = dependencyKeys
	}

	glog.V(4).Infof("Dependencies:")
	for k, v := range edges {
		glog.V(4).Infof("\t%s:\t%v", k, v)
	}

	return edges
}

func reflectForDependencies(tasks map[string]tasks.Task, task tasks.Task) []tasks.Task {
	v := reflect.ValueOf(task).Elem()
	return getDependencies(tasks, v)
}

func getDependencies(taskMap map[string]tasks.Task, v reflect.Value) []tasks.Task {
	var dependencies []tasks.Task

	err := utils.ReflectRecursive(v, func(path string, f *reflect.StructField, v reflect.Value) error {
		if utils.IsPrimitiveValue(v) {
			return nil
		}

		switch v.Kind() {
		case reflect.String:
			return nil

		case reflect.Interface, reflect.Ptr, reflect.Slice, reflect.Map:
			// The recursive walk will descend into this; we can ignore here
			return nil

		case reflect.Struct:
			if path == "" {
				// Ignore self - we are a struct, but not our own dependency!
				return nil
			}

			// TODO: Can we / should we use a type-switch statement
			intf := v.Addr().Interface()
			if hd, ok := intf.(HasDependencies); ok {
				deps := hd.GetDependencies(taskMap)
				dependencies = append(dependencies, deps...)
			} else if dep, ok := intf.(tasks.Task); ok {
				dependencies = append(dependencies, dep)
			} else if _, ok := intf.(Resource); ok {
				// Ignore: not a dependency (?)
			} else if _, ok := intf.(*ResourceHolder); ok {
				// Ignore: not a dependency (?)
			} else if _, ok := intf.(*pkix.Name); ok {
				// Ignore: not a dependency
			} else {
				return fmt.Errorf("Unhandled type for %q: %T", path, v.Interface())
			}
			return utils.SkipReflection

		default:
			glog.Infof("Unhandled kind for %q: %T", path, v.Interface())
			return fmt.Errorf("Unhandled kind for %q: %v", path, v.Kind())
		}
	})

	if err != nil {
		glog.Fatalf("unexpected error finding dependencies %v", err)
	}

	return dependencies
}
