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

// CompareWithID indicates that the value should be compared by the returned ID value (instead of a deep comparison)
// Most Tasks implement this, because typically when a Task references another task, it only is concerned with
// being linked to that task, not the values of the task.
// For example, when an instance is linked to a disk, it cares that the disk is attached to that instance,
// not the size or speed of the disk.
type CompareWithID interface {
	CompareWithID() *string
}
