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
	"encoding/json"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ApplyableObject is implemented by objects that can be applied to the cluster.
// We don't need much, so this might allow for more efficient implementations in future.
type ApplyableObject interface {
	// GroupVersionKind returns the GroupVersionKind structure describing the type of the object
	GroupVersionKind() schema.GroupVersionKind
	// GetNamespace returns the namespace of the object
	GetNamespace() string
	// GetName returns the name of the object
	GetName() string

	// The object should implement json marshalling
	json.Marshaler
}
