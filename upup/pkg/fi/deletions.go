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

type DeletionProcessingMode string

const (
	// DeletionProcessingModeIgnore will ignore all deletion tasks.
	DeletionProcessingModeIgnore DeletionProcessingMode = "Ignore"
	// TODO: implement deferred-deletion in the tasks!
	// DeletionProcessingModeDeleteIfNotDeferrred will delete resources only if they are not marked for deferred-deletion.
	DeletionProcessingModeDeleteIfNotDeferrred DeletionProcessingMode = "IfNotDeferred"
	// DeletionProcessingModeDeleteIncludingDeferrred will delete resources including those marked for deferred-deletion.
	DeletionProcessingModeDeleteIncludingDeferred DeletionProcessingMode = "DeleteIncludingDeferred"
)

type ProducesDeletions[T SubContext] interface {
	FindDeletions(*Context[T]) ([]Deletion[T], error)
}

type CloudupProducesDeletions = ProducesDeletions[CloudupSubContext]

type Deletion[T SubContext] interface {
	Delete(target Target[T]) error
	TaskName() string
	Item() string
}

type CloudupDeletion = Deletion[CloudupSubContext]
