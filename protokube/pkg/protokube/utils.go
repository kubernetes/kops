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

package protokube

import (
	"encoding/json"
	"fmt"
)

func DebugString(o interface{}) string {
	b, err := json.Marshal(o)
	if err != nil {
		return fmt.Sprintf("error marshaling %T: %v", o, err)
	}
	return string(b)
}
