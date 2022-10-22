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
	"context"
	"fmt"
	"reflect"
	"strings"

	"k8s.io/klog/v2"
)

type Task[T SubContext] interface {
	Run(*Context[T]) error
}

type CloudupTask = Task[CloudupSubContext]
type NodeupTask = Task[NodeupSubContext]

// TaskPreRun is implemented by tasks that perform some initial validation.
type TaskPreRun[T SubContext] interface {
	Task[T]
	// PreRun will be run for all TaskPreRuns, before any Run functions are invoked.
	PreRun(*Context[T]) error
}

// TaskNormalize is implemented by tasks that perform some initial normalization.
type TaskNormalize[T SubContext] interface {
	Task[T]
	// Normalize will be run for all TaskNormalizes, before the Run function of
	// the TaskNormalize and after the Run function of any Task it is dependent on.
	Normalize(*Context[T]) error
}

type CloudupTaskNormalize = TaskNormalize[CloudupSubContext]

// TaskAsString renders the task for debug output
// TODO: Use reflection to make this cleaner: don't recurse into tasks - print their names instead
// also print resources in a cleaner way (use the resource source information?)
func TaskAsString[T SubContext](t Task[T]) string {
	return fmt.Sprintf("%T %s", t, DebugAsJsonString(t))
}

// CloudupTaskAsString renders the task for debug output
// TODO: Use reflection to make this cleaner: don't recurse into tasks - print their names instead
// also print resources in a cleaner way (use the resource source information?)
func CloudupTaskAsString(t CloudupTask) string {
	return TaskAsString(t)
}

// NodeupTaskAsString renders the task for debug output
// TODO: Use reflection to make this cleaner: don't recurse into tasks - print their names instead
// also print resources in a cleaner way (use the resource source information?)
func NodeupTaskAsString(t NodeupTask) string {
	return TaskAsString(t)
}

type HasCheckExisting[T SubContext] interface {
	Task[T]
	CheckExisting(c *Context[T]) bool
}

type NodeupHasCheckExisting = HasCheckExisting[NodeupSubContext]
type CloudupHasCheckExisting = HasCheckExisting[CloudupSubContext]

// ModelBuilder allows for plugins that configure an aspect of the model, based on the configuration
type ModelBuilder[T SubContext] interface {
	Build(context *ModelBuilderContext[T]) error
}

type CloudupModelBuilder = ModelBuilder[CloudupSubContext]
type NodeupModelBuilder = ModelBuilder[NodeupSubContext]

// HasDeletions is a ModelBuilder[CloudupContext] that creates tasks to delete cloud objects that no longer exist in the model.
type HasDeletions interface {
	ModelBuilder[CloudupSubContext]
	// FindDeletions finds cloud objects that are owned by the cluster but no longer in the model and creates tasks to delete them.
	// It is not called for the Terraform target.
	FindDeletions(context *ModelBuilderContext[CloudupSubContext], cloud Cloud) error
}

// ModelBuilderContext is a context object that holds state we want to pass to ModelBuilder
type ModelBuilderContext[T SubContext] struct {
	// ctx holds the context.Context, ideally we would pass this in to every handler,
	// but that is a fairly large refactor, and arguably ModelBuilderContext has a similar
	// lifecycle to a context.Context
	ctx context.Context

	Tasks              map[string]Task[T]
	LifecycleOverrides map[string]Lifecycle
}

func (c *ModelBuilderContext[T]) WithContext(ctx context.Context) *ModelBuilderContext[T] {
	c2 := *c
	c2.ctx = ctx
	return &c2
}

func (c *ModelBuilderContext[T]) Context() context.Context {
	ctx := c.ctx
	if ctx == nil {
		ctx = context.TODO()
	}
	return ctx
}

type NodeupModelBuilderContext = ModelBuilderContext[NodeupSubContext]
type CloudupModelBuilderContext = ModelBuilderContext[CloudupSubContext]

func (c *ModelBuilderContext[T]) AddTask(task Task[T]) {
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
func (c *ModelBuilderContext[T]) EnsureTask(task Task[T]) error {
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
func (c *ModelBuilderContext[T]) setLifecycleOverride(task Task[T]) Task[T] {
	// TODO(@chrislovecnm) - wonder if we should update the nodeup tasks to have lifecycle
	// TODO - so that we can return an error here, rather than just returning.
	// certain tasks have not implemented HasLifecycle interface
	typeName := TypeNameForTask(task)

	// typeName can be values like "InternetGateway"
	value, ok := c.LifecycleOverrides[typeName]
	if ok {
		hl, okHL := task.(HasLifecycle)
		if !okHL {
			klog.Warningf("task %T does not implement HasLifecycle", task)
			return task
		}

		klog.Infof("overriding task %s, lifecycle %s", task, value)
		hl.SetLifecycle(value)
	}

	return task
}

func buildTaskKey[T SubContext](task Task[T]) string {
	hasName, ok := task.(HasName)
	if !ok {
		klog.Fatalf("task %T does not implement HasName", task)
	}

	name := ValueOf(hasName.GetName())
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
