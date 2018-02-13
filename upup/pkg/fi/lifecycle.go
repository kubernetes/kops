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
}

// Phases are used for validation and cli help.
var Lifecycles = sets.NewString(
	string(LifecycleSync),
	string(LifecycleIgnore),
	string(LifecycleWarnIfInsufficientAccess),
	string(LifecycleExistsAndValidates),
	string(LifecycleExistsAndWarnIfChanges),
)

var LifecycleNameMap = map[string]Lifecycle{
	"Sync":                     LifecycleSync,
	"Ignore":                   LifecycleIgnore,
	"WarnIfInsufficientAccess": LifecycleWarnIfInsufficientAccess,
	"ExistsAndValidates":       LifecycleExistsAndValidates,
	"ExistsAndWarnIfChanges":   LifecycleExistsAndWarnIfChanges,
}
