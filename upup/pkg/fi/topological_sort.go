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

package fi

import (
	"crypto/x509/pkix"
	"fmt"
	"reflect"

	"k8s.io/klog"

	"k8s.io/kops/util/pkg/reflectutils"
)

type HasDependencies interface {
	GetDependencies(tasks map[string]Task) []Task
}

// FindTaskDependencies returns a map from each task's key to the discovered list of dependencies
func FindTaskDependencies(tasks map[string]Task) map[string][]string {
	taskToId := make(map[interface{}]string)
	for k, t := range tasks {
		taskToId[t] = k
	}

	edges := make(map[string][]string)

	for k, t := range tasks {
		task := t.(Task)

		var dependencies []Task
		if hd, ok := task.(HasDependencies); ok {
			dependencies = hd.GetDependencies(tasks)
		} else {
			dependencies = reflectForDependencies(tasks, task)
		}

		var dependencyKeys []string
		for _, dep := range dependencies {
			dependencyKey, found := taskToId[dep]
			if !found {
				klog.Fatalf("dependency not found: %v", dep)
			}
			dependencyKeys = append(dependencyKeys, dependencyKey)
		}

		edges[k] = dependencyKeys
	}

	klog.V(4).Infof("Dependencies:")
	for k, v := range edges {
		klog.V(4).Infof("\t%s:\t%v", k, v)
	}

	return edges
}

func reflectForDependencies(tasks map[string]Task, task Task) []Task {
	v := reflect.ValueOf(task).Elem()
	return getDependencies(tasks, v)
}

func getDependencies(tasks map[string]Task, v reflect.Value) []Task {
	var dependencies []Task

	err := reflectutils.ReflectRecursive(v, func(path string, f *reflect.StructField, v reflect.Value) error {
		if reflectutils.IsPrimitiveValue(v) {
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
				deps := hd.GetDependencies(tasks)
				dependencies = append(dependencies, deps...)
			} else if dep, ok := intf.(Task); ok {
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
			return reflectutils.SkipReflection

		default:
			klog.Infof("Unhandled kind for %q: %T", path, v.Interface())
			return fmt.Errorf("Unhandled kind for %q: %v", path, v.Kind())
		}
	})

	if err != nil {
		klog.Fatalf("unexpected error finding dependencies %v", err)
	}

	return dependencies
}
