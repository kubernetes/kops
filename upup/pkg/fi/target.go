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

type Target[T SubContext] interface {
	// Lifecycle methods, called by the driver
	Finish(taskMap map[string]Task[T]) error

	// ProcessDeletions returns true if we should delete resources
	// Some providers (e.g. Terraform) actively keep state, and will delete resources automatically
	ProcessDeletions() bool

	// DefaultCheckExisting returns true if DefaultDeltaRun tasks which aren't HasCheckExisting
	// should invoke Find() when running against this Target.
	DefaultCheckExisting() bool
}

type CloudupTarget = Target[CloudupSubContext]
type InstallTarget = Target[InstallSubContext]
type NodeupTarget = Target[NodeupSubContext]
