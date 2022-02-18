/*
Copyright 2020 The Kubernetes Authors.

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

package util

import (
	"fmt"
	"strings"
)

// parseTaint takes a string and returns a map of its value
// it mimics the function from https://github.com/kubernetes/kubernetes/blob/master/pkg/util/taints/taints.go
// but returns a map instead of a v1.Taint
func ParseTaint(st string) (map[string]string, error) {
	taint := make(map[string]string)

	var key string
	var value string
	var effect string

	parts := strings.Split(st, ":")
	switch len(parts) {
	case 1:
		key = parts[0]
	case 2:
		effect = parts[1]

		partsKV := strings.Split(parts[0], "=")
		if len(partsKV) > 2 {
			return taint, fmt.Errorf("invalid taint spec: %v", st)
		}
		key = partsKV[0]
		if len(partsKV) == 2 {
			value = partsKV[1]
		}
	default:
		return taint, fmt.Errorf("invalid taint spec: %v", st)
	}

	taint["key"] = key
	taint["value"] = value
	taint["effect"] = effect

	return taint, nil
}
