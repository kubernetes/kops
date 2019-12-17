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
	"fmt"
	"reflect"
	"strings"

	"k8s.io/klog"
)

type Task interface {
	Run(*Context) error
}

// TaskAsString renders the task for debug output
// TODO: Use reflection to make this cleaner: don't recurse into tasks - print their names instead
// also print resources in a cleaner way (use the resource source information?)
func TaskAsString(t Task) string {
	return fmt.Sprintf("%T %s", t, DebugAsJsonString(t))
}

type HasCheckExisting interface {
	CheckExisting(c *Context) bool
}

// ModelBuilder allows for plugins that configure an aspect of the model, based on the configuration
type ModelBuilder interface {
	Build(context *ModelBuilderContext) error
}

// ModelBuilderContext is a context object that holds state we want to pass to ModelBuilder
type ModelBuilderContext struct {
	Tasks              map[string]Task
	LifecycleOverrides map[string]Lifecycle
}

func (c *ModelBuilderContext) AddTask(task Task) {
	task = c.setLifecycleOverride(task)
	key := buildTaskKey(task)

	existing, found := c.Tasks[key]
	if found {
		klog.Fatalf("found duplicate tasks with name %q: %v and %v", key, task, existing)
	}
	c.Tasks[key] = task
}

// EnsureTask ensures that the specified task is configured.
// It adds the task if it does not already exist.
// If it does exist, it verifies that the existing task reflect.DeepEqual the new task,
// if they are different an error is returned.
func (c *ModelBuilderContext) EnsureTask(task Task) error {
	task = c.setLifecycleOverride(task)
	key := buildTaskKey(task)

	existing, found := c.Tasks[key]
	if found {
		if reflect.DeepEqual(task, existing) {
			klog.V(8).Infof("EnsureTask ignoring identical ")
			return nil
		}
		klog.Warningf("EnsureTask found task mismatch for %q", key)
		klog.Warningf("\tExisting: %v", existing)
		klog.Warningf("\tNew: %v", task)

		return fmt.Errorf("cannot add different task with same key %q", key)
	}
	c.Tasks[key] = task
	return nil
}

// setLifecycleOverride determines if a Lifecycle is in the LifecycleOverrides map for the current task.
// If the lifecycle exist then the task lifecycle is set to the lifecycle provides in LifecycleOverrides.
// This func allows for lifecycles to be passed in dynamically and have the task lifecycle set accordingly.
func (c *ModelBuilderContext) setLifecycleOverride(task Task) Task {
	// TODO(@chrislovecnm) - wonder if we should update the nodeup tasks to have lifecycle
	// TODO - so that we can return an error here, rather than just returning.
	// certain tasks have not implemented HasLifecycle interface
	hl, ok := task.(HasLifecycle)
	if !ok {
		klog.V(8).Infof("task %T does not implement HasLifecycle", task)
		return task
	}

	typeName := TypeNameForTask(task)
	klog.V(8).Infof("testing task %q", typeName)

	// typeName can be values like "InternetGateway"
	value, ok := c.LifecycleOverrides[typeName]
	if ok {
		klog.Warningf("overriding task %s, lifecycle %s", task, value)
		hl.SetLifecycle(value)
	}

	return task
}

func buildTaskKey(task Task) string {
	hasName, ok := task.(HasName)
	if !ok {
		klog.Fatalf("task %T does not implement HasName", task)
	}

	name := StringValue(hasName.GetName())
	if name == "" {
		klog.Fatalf("task %T (%v) did not have a Name", task, task)
	}

	typeName := TypeNameForTask(task)

	key := typeName + "/" + name

	return key
}

func TypeNameForTask(task interface{}) string {
	typeName := fmt.Sprintf("%T", task)
	lastDot := strings.LastIndex(typeName, ".")
	typeName = typeName[lastDot+1:]
	return typeName
}
