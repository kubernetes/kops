package fi

import (
	"k8s.io/kops/upup/pkg/fi/utils"
	"reflect"
)

// DefaultDeltaRunMethod implements the standard change-based run procedure:
// find the existing item; compare properties; call render with (actual, expected, changes)
func DefaultDeltaRunMethod(e Task, c *Context) error {
	var a Task
	var err error

	checkExisting := c.CheckExisting
	if hce, ok := e.(HasCheckExisting); ok {
		checkExisting = hce.CheckExisting(c)
	}

	if checkExisting {
		a, err = invokeFind(e, c)
		if err != nil {
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

		err = c.Render(a, e, changes)
		if err != nil {
			return err
		}
	}

	if producesDeletions, ok := e.(ProducesDeletions); ok {
		var deletions []Deletion
		deletions, err = producesDeletions.FindDeletions(c)
		if err != nil {
			return err
		}
		for _, deletion := range deletions {
			if _, ok := c.Target.(*DryRunTarget); ok {
				err = c.Target.(*DryRunTarget).Delete(deletion)
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
func invokeCheckChanges(a, e, changes Task) error {
	rv, err := utils.InvokeMethod(e, "CheckChanges", a, e, changes)
	if err != nil {
		return err
	}
	if !rv[0].IsNil() {
		err = rv[0].Interface().(error)
	}
	return err
}

// invokeFind calls the find method by reflection
func invokeFind(e Task, c *Context) (Task, error) {
	rv, err := utils.InvokeMethod(e, "Find", c)
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
