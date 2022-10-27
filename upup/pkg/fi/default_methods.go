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

// DefaultDeltaRunMethod implements the standard change-based run procedure:
// find the existing item; compare properties; call render with (actual, expected, changes)
func DefaultDeltaRunMethod(e Task, c Context) error {
	var a Task
	var err error

	target := c.GetTarget()

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

	var contextBase *ContextBase
	switch c := c.(type) {
	case *NodeContext:
		contextBase = &c.ContextBase
	case *CloudContext:
		contextBase = &c.ContextBase
	default:
		return fmt.Errorf("unhandled context type %T", c)
	}

	checkExisting := contextBase.CheckExisting
	if hce, ok := e.(HasCheckExisting); ok {
		checkExisting = hce.CheckExisting(c.(*CloudContext))
	}

	if checkExisting {
		a, err = invokeFind(e, c)
		if err != nil {
			if lifecycle == LifecycleWarnIfInsufficientAccess {
				// For now we assume all errors are permissions problems
				// TODO: bounded retry?
				contextBase.AddWarning(e, fmt.Sprintf("error checking if task exists; assuming it is correctly configured: %v", err))
				return nil
			}
			return err
		}
	}

	if a == nil {
		// This is kind of subtle.  We want an interface pointer to a struct of the correct type...
		a = reflect.New(reflect.TypeOf(e)).Elem().Interface().(Task)
	}

	changes := reflect.New(reflect.TypeOf(e).Elem()).Interface().(Task)
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
			err = invokeRender(c, a, e, changes)
			if err != nil {
				return err
			}
		}
	}

	if producesDeletions, ok := e.(ProducesDeletions); ok && target.ProcessDeletions() {
		var deletions []Deletion
		deletions, err = producesDeletions.FindDeletions(c.(*CloudContext))
		if err != nil {
			return err
		}
		target := c.GetTarget()
		for _, deletion := range deletions {
			if _, ok := target.(*DryRunTarget); ok {
				err = target.(*DryRunTarget).Delete(deletion)
			} else if _, ok := target.(*DryRunTarget); ok {
				err = target.(*DryRunTarget).Delete(deletion)
			} else {
				err = deletion.Delete(target)
			}
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// invokeCheckChanges calls the checkChanges method by reflection
func invokeCheckChanges(a, e, changes Task) error {
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
func invokeFind(e Task, c Context) (Task, error) {
	var args []interface{}
	switch c := c.(type) {
	case *NodeContext:
		args = append(args, c)
	case *CloudContext:
		args = append(args, c)
	default:
		return nil, fmt.Errorf("unhandled context type %T", c)
	}

	rv, err := reflectutils.InvokeMethod(e, "Find", args...)
	if err != nil {
		return nil, err
	}
	var task Task
	if !rv[0].IsNil() {
		task = rv[0].Interface().(Task)
	}
	if !rv[1].IsNil() {
		err = rv[1].Interface().(error)
	}
	return task, err
}

// invokeShouldCreate calls the ShouldCreate method by reflection, if it exists
func invokeShouldCreate(a, e, changes Task) (bool, error) {
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
