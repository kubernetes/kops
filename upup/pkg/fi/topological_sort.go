package fi

import (
	"crypto/x509/pkix"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi/utils"
	"reflect"
	"sort"
)

type HasDependencies interface {
	GetDependencies(tasks map[string]Task) []string
}

func TopologicalSort(tasks map[string]Task) [][]string {
	taskToId := make(map[interface{}]string)
	for k, t := range tasks {
		taskToId[t] = k
	}

	edges := make(map[string][]string)

	for k, t := range tasks {
		task := t.(Task)
		var dependencyKeys []string

		if hd, ok := task.(HasDependencies); ok {
			dependencyKeys = hd.GetDependencies(tasks)
		} else {
			dependencyKeys = reflectForDependencies(task, taskToId)
		}

		edges[k] = dependencyKeys
	}

	glog.V(4).Infof("Dependencies:")
	for k, v := range edges {
		glog.V(4).Infof("\t%s:\t%v", k, v)
	}

	ordered := toposort(edges)
	glog.V(1).Infof("toposorted as:")
	for i, stage := range ordered {
		glog.V(1).Infof("\t%d\t%v", i, stage)
	}

	return ordered
}

func reflectForDependencies(task Task, taskToId map[interface{}]string) []string {
	v := reflect.ValueOf(task).Elem()
	dependencies := getDependencies(v)

	var dependencyKeys []string
	for _, dep := range dependencies {
		dependencyKey, found := taskToId[dep]
		if !found {
			glog.Fatalf("dependency not found: %v", dep)
		}
		dependencyKeys = append(dependencyKeys, dependencyKey)
	}

	return dependencyKeys
}

// Perform a topological sort
// Note that we group them into stages, where each stage has no dependencies on other members of that stage
// This could support parallelism but also pushes nodes with fewer dependencies earlier
//
// This is not a particularly efficient implementation, but is simple,
// and likely good enough for the sizes we will be dealing with
func toposort(edges map[string][]string) [][]string {
	var stages [][]string

	for {
		if len(edges) == 0 {
			break
		}

		var stage []string
		for k, in := range edges {
			if len(in) != 0 {
				continue
			}
			stage = append(stage, k)
		}

		// For consistency
		sort.Strings(stage)

		if len(stage) == 0 {
			glog.Fatalf("graph is circular")
		}

		stages = append(stages, stage)

		stageSet := make(map[string]bool)
		for _, k := range stage {
			delete(edges, k)
			stageSet[k] = true
		}

		for k, in := range edges {
			var after []string
			for _, v := range in {
				if !stageSet[v] {
					after = append(after, v)
				}
			}
			edges[k] = after
		}
	}

	return stages
}

func getDependencies(v reflect.Value) []Task {
	var dependencies []Task

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
			if dep, ok := intf.(Task); ok {
				dependencies = append(dependencies, dep)
			} else if _, ok := intf.(Resource); ok {
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
