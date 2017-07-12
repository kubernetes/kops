/*
Copyright 2016 The Kubernetes Authors.

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
	"sync"
)

// InventoryTarget is a special Target that does not execute anything, but instead tracks all changes.
// By running an InventoryTarget, api.Inventory can be built without any special support from the Tasks.
type InventoryTarget struct {
	mutex sync.Mutex

	changes   []*render
	deletions []Deletion
}

var _ Target = &InventoryTarget{}

// NewInventoryTarget creates a new target.
func NewInventoryTarget() *InventoryTarget {
	return &InventoryTarget{}
}

// ProcessDeletions dummy method to process deletes.
func (t *InventoryTarget) ProcessDeletions() bool {
	return true
}

// Render tracks task changes.
func (t *InventoryTarget) Render(a, e, changes Task) error {
	valA := reflect.ValueOf(a)
	aIsNil := valA.IsNil()

	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.changes = append(t.changes, &render{
		a:       a,
		aIsNil:  aIsNil,
		e:       e,
		changes: changes,
	})
	return nil
}

// Render tracks removes tasks.
func (t *InventoryTarget) Delete(deletion Deletion) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.deletions = append(t.deletions, deletion)

	return nil
}

// Finish does not really do anything in this target besides returning nil.
func (t *InventoryTarget) Finish(taskMap map[string]Task) error {
	return nil
}

// HasChanges returns true if any changes would have been made.
func (t *InventoryTarget) HasChanges() bool {
	return len(t.changes)+len(t.deletions) != 0
}
