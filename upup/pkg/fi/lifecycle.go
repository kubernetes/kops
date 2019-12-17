/*
Copyright 2017 The Kubernetes Authors.

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

import "k8s.io/apimachinery/pkg/util/sets"

type Lifecycle string

const (
	// LifecycleSync should do the normal synchronization
	LifecycleSync Lifecycle = "Sync"

	// LifecycleIgnore will skip the task
	LifecycleIgnore Lifecycle = "Ignore"

	// LifecycleWarnIfInsufficientAccess will warn but ignore the task if there is an error during the find
	LifecycleWarnIfInsufficientAccess Lifecycle = "WarnIfInsufficientAccess"

	// LifecycleExistsAndValidates will check that the task exists and is the same
	LifecycleExistsAndValidates Lifecycle = "ExistsAndValidates"

	// LifecycleExistsAndWarnIfChanges will check that the task exists and will warn on changes, but then ignore them
	LifecycleExistsAndWarnIfChanges Lifecycle = "ExistsAndWarnIfChanges"
)

// HasLifecycle indicates that the task has a Lifecycle
type HasLifecycle interface {
	GetLifecycle() *Lifecycle
	// SetLifecycle is used to override a tasks lifecycle. If a lifecycle override exists for a specific task name, then the
	// lifecycle is modified.
	SetLifecycle(lifecycle Lifecycle)
}

// Lifecycles are used for ux validation.  When validation fails the lifecycle names are
// printed out.
var Lifecycles = sets.NewString(
	string(LifecycleSync),
	string(LifecycleIgnore),
	string(LifecycleWarnIfInsufficientAccess),
	string(LifecycleExistsAndValidates),
	string(LifecycleExistsAndWarnIfChanges),
)

// LifecycleNameMap is used to validate in the UX.  When a user provides a lifecycle name
// it then can be mapped to the actual lifecycle.
var LifecycleNameMap = map[string]Lifecycle{
	"Sync":                     LifecycleSync,
	"Ignore":                   LifecycleIgnore,
	"WarnIfInsufficientAccess": LifecycleWarnIfInsufficientAccess,
	"ExistsAndValidates":       LifecycleExistsAndValidates,
	"ExistsAndWarnIfChanges":   LifecycleExistsAndWarnIfChanges,
}
