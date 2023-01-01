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

package debug

import (
	"encoding/json"
	"fmt"
)

// DeferredJSON is a helper that delays JSON formatting until/unless it is needed.
type DeferredJSON struct {
	O interface{}
}

// String is the method that is called to format the object.
func (d DeferredJSON) String() string {
	b, err := json.Marshal(d.O)
	if err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	return string(b)
}

// JSON is a helper that prints the object in JSON format.
// We use lazy-evaluation to avoid calling json.Marshal unless it is actually needed.
func JSON(o interface{}) DeferredJSON {
	return DeferredJSON{o}
}
