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

package testcontext

import (
	"context"
	"testing"
)

// ContextForTest returns a Context for the given test scope.
func ContextForTest(t *testing.T) context.Context {
	ctx := context.TODO()
	// We might choose to bind the test to the context in future,
	// or bind the logger etc.
	return ctx
}
