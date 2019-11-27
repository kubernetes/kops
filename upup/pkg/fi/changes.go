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
	"reflect"

	"k8s.io/klog"
)

// An important part of our state synchronization is to compare two tasks, to see what has changed
// Doing so means that the tasks don't have to have this logic, and we can reuse this for dry-run.
// We do so using reflection.  We have a custom notion of equality for Resources and for Tasks that implement HasID.
// A task that implements CompareWithID is compared by ID alone.

// BuildChanges compares the values of a & e, and populates differences into changes,
// except that if a value is nil in e, the corresponding value in a is ignored.
// a, e and changes must all be of the same type
// a is the actual object found, e is the expected value
// Note that the ignore-nil-in-e logic therefore implements the idea that nil value in e means "don't care"
// If a is nil, all the non-nil values in e will be copied over to changes, because every field in e must be applied
func BuildChanges(a, e, changes interface{}) bool {
	changed := false

	vc := reflect.ValueOf(changes)
	vc = vc.Elem()
	t := vc.Type()

	ve := reflect.ValueOf(e)
	ve = ve.Elem()
	if t != ve.Type() {
		panic("mismatched types in BuildChanges")
	}

	va := reflect.ValueOf(a)
	aIsNil := false
	if va.IsNil() {
		aIsNil = true
	}
	if !aIsNil {
		va = va.Elem()

		if t != va.Type() {
			panic("mismatched types in BuildChanges")
		}
	}

	for i := 0; i < ve.NumField(); i++ {
		if t.Field(i).PkgPath != "" {
			// unexported: ignore
			continue
		}

		fve := ve.Field(i)
		if fve.Kind() == reflect.Ptr && fve.IsNil() {
			// Nil expected value means 'don't care'
			continue
		}

		if !aIsNil {
			fva := va.Field(i)

			if equalFieldValues(fva, fve) {
				continue
			}

			klog.V(8).Infof("Field changed %q actual=%q expected=%q", t.Field(i).Name, DebugPrint(fva.Interface()), DebugPrint(fve.Interface()))
		}
		changed = true
		vc.Field(i).Set(fve)
	}

	return changed
}

// equalFieldValues implements our equality checking, with special cases for resources and tasks
func equalFieldValues(a, e reflect.Value) bool {
	if !a.IsValid() || !e.IsValid() {
		return a.IsValid() == e.IsValid()
	}

	if a.Kind() == reflect.Map {
		return equalMapValues(a, e)
	}
	if a.Kind() == reflect.Slice {
		return equalSlice(a, e)
	}
	if (a.Kind() == reflect.Ptr || a.Kind() == reflect.Interface) && !a.IsNil() {
		aHasID, ok := a.Interface().(CompareWithID)
		if ok && (e.Kind() == reflect.Ptr || e.Kind() == reflect.Interface) && !e.IsNil() {
			eHasID, ok := e.Interface().(CompareWithID)
			if ok {
				aID := aHasID.CompareWithID()
				eID := eHasID.CompareWithID()
				if aID != nil && eID != nil && *aID == *eID {
					return true
				}
			}
		}

		aResource, ok := a.Interface().(Resource)
		if ok && (e.Kind() == reflect.Ptr || e.Kind() == reflect.Interface) && !e.IsNil() {
			eResource, ok := e.Interface().(Resource)
			if ok {
				same, err := ResourcesMatch(aResource, eResource)
				if err != nil {
					klog.Fatalf("error while comparing resources: %v", err)
				} else {
					return same
				}
			}
		}
	}
	if reflect.DeepEqual(a.Interface(), e.Interface()) {
		return true
	}
	return false
}

// equalMapValues performs a deep-equality check on a map, but using our custom comparison logic (equalFieldValues)
func equalMapValues(a, e reflect.Value) bool {
	if a.IsNil() != e.IsNil() {
		return false
	}
	if a.IsNil() && e.IsNil() {
		return true
	}
	if a.Len() != e.Len() {
		return false
	}
	for _, k := range a.MapKeys() {
		valA := a.MapIndex(k)
		valE := e.MapIndex(k)

		klog.V(10).Infof("comparing maps: %v %v %v", k, valA, valE)

		if !equalFieldValues(valA, valE) {
			klog.V(4).Infof("unequal map value: %v %v %v", k, valA, valE)
			return false
		}
	}
	return true
}

// equalSlice performs a deep-equality check on a slice, but using our custom comparison logic (equalFieldValues)
func equalSlice(a, e reflect.Value) bool {
	if a.IsNil() != e.IsNil() {
		return false
	}
	if a.IsNil() && e.IsNil() {
		return true
	}
	if a.Len() != e.Len() {
		return false
	}
	for i := 0; i < a.Len(); i++ {
		valA := a.Index(i)
		valE := e.Index(i)

		klog.V(10).Infof("comparing slices: %d %v %v", i, valA, valE)

		if !equalFieldValues(valA, valE) {
			klog.V(4).Infof("unequal slice value: %d %v %v", i, valA, valE)
			return false
		}
	}
	return true
}
