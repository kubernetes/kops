/*
Copyright 2022 The Kubernetes Authors.

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

package applyset

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
)

// objectTrackerList is a list of objectTrackers, containing the state of the objects we are trying to apply.
// objectTrackerList is immutable (though objectTracker is mutable); we copy-on-write when the list changes.
// TODO: Given objectTracker is mutable, should we just make objectTrackerList mutable?
type objectTrackerList struct {
	items []objectTracker
}

// objectTracker tracks the state for a single object
type objectTracker struct {
	desired     ApplyableObject
	lastApplied runtime.Object

	desiredIsApplied bool
	isHealthy        bool
}

// objectKey is the key used in maps; we consider objects with the same GVKNN the same.
type objectKey struct {
	Group     string
	Version   string
	Kind      string
	Namespace string
	Name      string
}

// computeKey returns the unique key for the object.
func computeKey(u ApplyableObject) objectKey {
	gvk := u.GroupVersionKind()
	return objectKey{
		Group:     gvk.Group,
		Version:   gvk.Version,
		Kind:      gvk.Kind,
		Namespace: u.GetNamespace(),
		Name:      u.GetName(),
	}
}

// setDesiredObjects completely replaces the set of objects we are interested in.
// We aim to reuse the current state where it carries over.
// Because objectTrackerList is immutable, we copy-on-write to a new objectTrackerList and return it.
func (l *objectTrackerList) setDesiredObjects(objects []ApplyableObject) *objectTrackerList {
	existingTrackers := make(map[objectKey]*objectTracker)
	for i := range l.items {
		tracker := &l.items[i]
		key := computeKey(tracker.desired)
		existingTrackers[key] = tracker
	}

	newList := &objectTrackerList{}

	for _, obj := range objects {
		key := computeKey(obj)
		// TODO: Detect duplicate keys?
		existingTracker := existingTrackers[key]
		if existingTracker == nil {
			newList.items = append(newList.items, objectTracker{
				desired:          obj,
				lastApplied:      nil,
				desiredIsApplied: false,
				isHealthy:        false,
			})
		} else if reflect.DeepEqual(existingTracker.desired, obj) {
			newList.items = append(newList.items, objectTracker{
				desired:          obj,
				lastApplied:      existingTracker.lastApplied,
				desiredIsApplied: existingTracker.desiredIsApplied,
				isHealthy:        existingTracker.isHealthy,
			})
		} else {
			newList.items = append(newList.items, objectTracker{
				desired:          obj,
				lastApplied:      existingTracker.lastApplied,
				desiredIsApplied: false,
				isHealthy:        existingTracker.isHealthy,
			})
		}
	}

	return newList
}
