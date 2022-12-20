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

	"k8s.io/kops/util/pkg/reflectutils"
)

// InstallDefaultDeltaRunMethod implements the standard change-based run procedure:
// find the existing item; compare properties; call render with (actual, expected, changes)
func InstallDefaultDeltaRunMethod(e InstallTask, c *InstallContext) error {
	return defaultDeltaRunMethod[InstallSubContext](e, c)
}

// NodeupDefaultDeltaRunMethod implements the standard change-based run procedure:
// find the existing item; compare properties; call render with (actual, expected, changes)
func NodeupDefaultDeltaRunMethod(e NodeupTask, c *NodeupContext) error {
	return defaultDeltaRunMethod[NodeupSubContext](e, c)
}

// CloudupDefaultDeltaRunMethod implements the standard change-based run procedure:
// find the existing item; compare properties; call render with (actual, expected, changes)
func CloudupDefaultDeltaRunMethod(e CloudupTask, c *CloudupContext) error {
	return defaultDeltaRunMethod[CloudupSubContext](e, c)
}

// defaultDeltaRunMethod implements the standard change-based run procedure:
// find the existing item; compare properties; call render with (actual, expected, changes)
func defaultDeltaRunMethod[T SubContext](e Task[T], c *Context[T]) error {
	var a Task[T]
	var err error

	lifecycle := LifecycleSync
	if hl, ok := e.(HasLifecycle); ok {
		lifecycle = hl.GetLifecycle()
		if lifecycle == "" {
			return fmt.Errorf("task does not have a lifecycle set")
		}
	}

	if lifecycle == LifecycleIgnore {
		return nil
	}

	checkExisting := c.Target.DefaultCheckExisting()
	if hce, ok := e.(HasCheckExisting[T]); ok {
		checkExisting = hce.CheckExisting(c)
	}

	if checkExisting {
		a, err = invokeFind(e, c)
		if err != nil {
			if lifecycle == LifecycleWarnIfInsufficientAccess {
				// For now we assume all errors are permissions problems
				// TODO: bounded retry?
				c.AddWarning(e, fmt.Sprintf("error checking if task exists; assuming it is correctly configured: %v", err))
				return nil
			}
			return err
		}
	}

	if a == nil {
		// This is kind of subtle.  We want an interface pointer to a struct of the correct type...
		a = reflect.New(reflect.TypeOf(e)).Elem().Interface().(Task[T])
	}

	changes := reflect.New(reflect.TypeOf(e).Elem()).Interface().(Task[T])
	changed := BuildChanges(a, e, changes)

	if changed {
		err = invokeCheckChanges(a, e, changes)
		if err != nil {
			return err
		}

		shouldCreate, err := invokeShouldCreate(a, e, changes)
		if err != nil {
			return err
		}

		if shouldCreate {
			err = c.Render(a, e, changes)
			if err != nil {
				return err
			}
		}
	}

	if producesDeletions, ok := e.(ProducesDeletions[T]); ok && c.Target.ProcessDeletions() {
		var deletions []Deletion[T]
		deletions, err = producesDeletions.FindDeletions(c)
		if err != nil {
			return err
		}
		for _, deletion := range deletions {
			if _, ok := c.Target.(*DryRunTarget[T]); ok {
				err = c.Target.(*DryRunTarget[T]).Delete(deletion)
			} else if _, ok := c.Target.(*DryRunTarget[T]); ok {
				err = c.Target.(*DryRunTarget[T]).Delete(deletion)
			} else {
				err = deletion.Delete(c.Target)
			}
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// invokeCheckChanges calls the checkChanges method by reflection
func invokeCheckChanges[T SubContext](a, e, changes Task[T]) error {
	rv, err := reflectutils.InvokeMethod(e, "CheckChanges", a, e, changes)
	if err != nil {
		return err
	}
	if !rv[0].IsNil() {
		err = rv[0].Interface().(error)
	}
	return err
}

// invokeFind calls the find method by reflection
func invokeFind[T SubContext](e Task[T], c *Context[T]) (Task[T], error) {
	rv, err := reflectutils.InvokeMethod(e, "Find", c)
	if err != nil {
		return nil, err
	}
	var task Task[T]
	if !rv[0].IsNil() {
		task = rv[0].Interface().(Task[T])
	}
	if !rv[1].IsNil() {
		err = rv[1].Interface().(error)
	}
	return task, err
}

// invokeShouldCreate calls the ShouldCreate method by reflection, if it exists
func invokeShouldCreate[T SubContext](a, e, changes Task[T]) (bool, error) {
	rv, err := reflectutils.InvokeMethod(e, "ShouldCreate", a, e, changes)
	if err != nil {
		if reflectutils.IsMethodNotFound(err) {
			return true, nil
		}
		return false, err
	}
	shouldCreate := rv[0].Interface().(bool)
	if !rv[1].IsNil() {
		err = rv[1].Interface().(error)
	}
	return shouldCreate, err
}
