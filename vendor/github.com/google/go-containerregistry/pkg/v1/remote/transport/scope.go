// Copyright 2018 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package transport

// Scopes suitable to qualify each Repository
const (
	PullScope string = "pull"
	PushScope string = "push,pull"
	// DeleteScope requests "delete" in addition to push/pull so that
	// registries requiring an explicit delete action (e.g. IBM Cloud
	// Container Registry) grant the necessary access.
	DeleteScope  string = "push,pull,delete"
	CatalogScope string = "catalog"
)
