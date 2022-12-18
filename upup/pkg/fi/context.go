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
	"bytes"
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/util/pkg/vfs"
)

type Context[T SubContext] struct {
	ctx context.Context

	Target      Target[T]
	Cloud       Cloud
	Cluster     *kops.Cluster
	Keystore    Keystore
	SecretStore SecretStore

	CheckExisting bool

	tasks    map[string]Task[T]
	warnings []*Warning[T]

	T T
}

type CloudupContext = Context[CloudupSubContext]
type NodeupContext = Context[NodeupSubContext]

type SubContext interface {
	CloudupSubContext | NodeupSubContext
}

type CloudupSubContext struct {
	// TODO: Few places use this. They could instead get it from the cluster spec.
	ClusterConfigBase vfs.Path
}
type NodeupSubContext struct{}

func (c *Context[T]) Context() context.Context {
	return c.ctx
}

// Warning holds the details of a warning encountered during validation/creation
type Warning[T SubContext] struct {
	Task    Task[T]
	Message string
}

func NewContext[T SubContext](ctx context.Context, target Target[T], cluster *kops.Cluster, cloud Cloud, keystore Keystore, secretStore SecretStore, checkExisting bool, sub T, tasks map[string]Task[T]) (*Context[T], error) {
	c := &Context[T]{
		ctx:           ctx,
		Cloud:         cloud,
		Cluster:       cluster,
		Target:        target,
		Keystore:      keystore,
		SecretStore:   secretStore,
		CheckExisting: checkExisting,
		tasks:         tasks,
		T:             sub,
	}

	return c, nil
}

func NewNodeupContext(ctx context.Context, target NodeupTarget, cluster *kops.Cluster, cloud Cloud, keystore Keystore, secretStore SecretStore, checkExisting bool, tasks map[string]NodeupTask) (*NodeupContext, error) {
	sub := NodeupSubContext{}
	return NewContext[NodeupSubContext](ctx, target, cluster, cloud, keystore, secretStore, checkExisting, sub, tasks)
}

func NewCloudupContext(ctx context.Context, target CloudupTarget, cluster *kops.Cluster, cloud Cloud, keystore Keystore, secretStore SecretStore, clusterConfigBase vfs.Path, checkExisting bool, tasks map[string]CloudupTask) (*CloudupContext, error) {
	sub := CloudupSubContext{
		ClusterConfigBase: clusterConfigBase,
	}
	return NewContext[CloudupSubContext](ctx, target, cluster, cloud, keystore, secretStore, checkExisting, sub, tasks)
}

func (c *Context[T]) AllTasks() map[string]Task[T] {
	return c.tasks
}

func (c *Context[T]) RunTasks(options RunTasksOptions) error {
	e := &executor[T]{
		context: c,
		options: options,
	}
	return e.RunTasks(c.tasks)
}

// Render dispatches the creation of an object to the appropriate handler defined on the Task,
// it is typically called after we have checked the existing state of the Task and determined that is different
// from the desired state.
func (c *Context[T]) Render(a, e, changes Task[T]) error {
	typeContextPtr := reflect.TypeOf((*Context[T])(nil))
	var lifecycle Lifecycle
	if hl, ok := e.(HasLifecycle); ok {
		lifecycle = hl.GetLifecycle()
	}

	if lifecycle != "" {
		if reflect.ValueOf(a).IsNil() {
			switch lifecycle {
			case LifecycleExistsAndValidates:
				return fmt.Errorf("lifecycle set to ExistsAndValidates, but object was not found")
			case LifecycleExistsAndWarnIfChanges:
				return NewExistsAndWarnIfChangesError("Lifecycle set to ExistsAndWarnIfChanges and object was not found.")
			}
		} else {
			switch lifecycle {
			case LifecycleExistsAndValidates, LifecycleExistsAndWarnIfChanges:

				out := os.Stderr
				changeList, err := buildChangeList(a, e, changes)
				if err != nil {
					return err
				}

				b := &bytes.Buffer{}
				taskName := getTaskName(e)
				fmt.Fprintf(b, "Object from different phase did not match, problems possible:\n")
				fmt.Fprintf(b, "  %s/%s\n", taskName, "?")
				for _, change := range changeList {
					lines := strings.Split(change.Description, "\n")
					if len(lines) == 1 {
						fmt.Fprintf(b, "  \t%-20s\t%s\n", change.FieldName, change.Description)
					} else {
						fmt.Fprintf(b, "  \t%-20s\n", change.FieldName)
						for _, line := range lines {
							fmt.Fprintf(b, "  \t%-20s\t%s\n", "", line)
						}
					}
				}
				fmt.Fprintf(b, "\n")
				b.WriteTo(out)

				if lifecycle == LifecycleExistsAndValidates {
					return fmt.Errorf("lifecycle set to ExistsAndValidates, but object did not match")
				}
				// Warn, but then we continue
				return nil
			}
		}
	}

	if _, ok := c.Target.(*DryRunTarget[T]); ok {
		return c.Target.(*DryRunTarget[T]).Render(a, e, changes)
	}

	v := reflect.ValueOf(e)
	vType := v.Type()

	targetType := reflect.ValueOf(c.Target).Type()

	var renderer *reflect.Method
	var rendererArgs []reflect.Value

	for i := 0; i < vType.NumMethod(); i++ {
		method := vType.Method(i)
		if !strings.HasPrefix(method.Name, "Render") {
			continue
		}
		match := true

		var args []reflect.Value
		for j := 0; j < method.Type.NumIn(); j++ {
			arg := method.Type.In(j)
			if arg.ConvertibleTo(vType) {
				continue
			}
			if arg.ConvertibleTo(typeContextPtr) {
				args = append(args, reflect.ValueOf(c))
				continue
			}
			if arg.ConvertibleTo(targetType) {
				args = append(args, reflect.ValueOf(c.Target))
				continue
			}
			match = false
			break
		}
		if match {
			if renderer != nil {
				if method.Name == "Render" {
					continue
				}
				if renderer.Name != "Render" {
					return fmt.Errorf("found multiple Render methods that could be involved on %T", e)
				}
			}
			renderer = &method
			rendererArgs = args
		}

	}
	if renderer == nil {
		return fmt.Errorf("could not find Render method on type %T (target %T)", e, c.Target)
	}
	rendererArgs = append(rendererArgs, reflect.ValueOf(a))
	rendererArgs = append(rendererArgs, reflect.ValueOf(e))
	rendererArgs = append(rendererArgs, reflect.ValueOf(changes))
	klog.V(11).Infof("Calling method %s on %T", renderer.Name, e)
	m := v.MethodByName(renderer.Name)
	rv := m.Call(rendererArgs)
	var rvErr error
	if !rv[0].IsNil() {
		rvErr = rv[0].Interface().(error)
	}
	return rvErr
}

// AddWarning records a warning encountered during validation / creation.
// Typically this will be an error that we choose to ignore because of Lifecycle.
func (c *Context[T]) AddWarning(task Task[T], message string) {
	warning := &Warning[T]{
		Task:    task,
		Message: message,
	}
	// We don't actually do anything with these warnings yet, other than log them to glog below.
	// In future we might produce a structured warning report.
	c.warnings = append(c.warnings, warning)
	klog.Warningf("warning during task %s: %s", task, message)
}

// ExistsAndWarnIfChangesError is the custom error return for fi.LifecycleExistsAndWarnIfChanges.
// This error is used when an object needs to fail validation, but let the user proceed with a warning.
type ExistsAndWarnIfChangesError struct {
	msg string
}

// NewExistsAndWarnIfChangesError is a builder for ExistsAndWarnIfChangesError.
func NewExistsAndWarnIfChangesError(message string) *ExistsAndWarnIfChangesError {
	return &ExistsAndWarnIfChangesError{
		msg: message,
	}
}

// ExistsAndWarnIfChangesError implementation of the error interface.
func (e *ExistsAndWarnIfChangesError) Error() string { return e.msg }

// TryAgainLaterError is the custom used when a task needs to fail validation with a message and try again later
type TryAgainLaterError struct {
	msg string
}

// NewTryAgainLaterError is a builder for TryAgainLaterError.
func NewTryAgainLaterError(message string) *TryAgainLaterError {
	return &TryAgainLaterError{
		msg: message,
	}
}

// TryAgainLaterError implementation of the error interface.
func (e *TryAgainLaterError) Error() string { return e.msg }
