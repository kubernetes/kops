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

package dns

// Context represents a state of the world for DNS.
// It is grouped by scopes & named keys, and controllers will replace those groups
// The DNS controller will then merge all those record sets, resolve aliases etc,
// and then call into a dns backend to match the desired state of the world.
type Context interface {
	// CreateScope creates a new scope, which holds a set of records.
	// MarkReady must be called on every scope before any changes will be applied.
	// Records from all the scopes will be merged
	CreateScope(name string) (Scope, error)
}

type Scope interface {
	// Replace sets the records for recordName to the provided set of records.
	Replace(recordName string, records []Record)

	// MarkReady should be called when a controller has populated all the records for a particular scope
	MarkReady()

	// AllKeys gets the set of all keys currently in the scope
	AllKeys() []string
}
