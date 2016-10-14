/*
Copyright 2016 The Kubernetes Authors.

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

import (
	"fmt"
)

type Task interface {
	Run(*Context) error
}

// TaskAsString renders the task for debug output
// TODO: Use reflection to make this cleaner: don't recurse into tasks - print their names instead
// also print resources in a cleaner way (use the resource source information?)
func TaskAsString(t Task) string {
	return fmt.Sprintf("%T %s", t, DebugAsJsonString(t))
}

type HasCheckExisting interface {
	CheckExisting(c *Context) bool
}
