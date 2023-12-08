/*
Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloud

import (
	"context"
	"time"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
)

const (
	defaultCallTimeout = 1 * time.Hour
)

// ContextWithCallTimeout returns a context with a default timeout, used for
// generated client calls.
func ContextWithCallTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultCallTimeout)
}

// CallContextKey is a key identifying the most commonly used parts of an operation.
type CallContextKey struct {
	// ProjectID is the non-numeric ID of the project.
	ProjectID string
	// Operation is the specific method being invoked (e.g. "Get", "List").
	Operation string
	// Version is the API version of the call.
	Version meta.Version
	// Service is the service being invoked (e.g. "Firewalls", "BackendServices")
	Service string
}
