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

package api

import (
	"regexp"
)

type ApiVersion string

func (this ApiVersion) LessThan(that ApiVersion) bool {
	re := regexp.MustCompile("(v\\d+)(alpha|beta|)(\\d*)")
	thisMatches := re.FindStringSubmatch(string(this))
	thatMatches := re.FindStringSubmatch(string(that))

	a := []string{thisMatches[1]}
	if len(thisMatches) > 2 {
		a = []string{thisMatches[2], thisMatches[1], thisMatches[0]}
	}

	b := []string{thatMatches[1]}
	if len(thatMatches) > 2 {
		b = []string{thatMatches[2], thatMatches[1], thatMatches[0]}
	}

	for i := 0; i < len(a) && i < len(b); i++ {
		v1 := ""
		v2 := ""
		if i < len(a) {
			v1 = a[i]
		}
		if i < len(b) {
			v2 = b[i]
		}
		// If the "beta" or "alpha" is missing, then it is ga (empty string comes before non-empty string)
		if len(v1) == 0 || len(v2) == 0 {
			return v1 < v2
		}
		// The string with the higher number comes first (or in the case of alpha/beta, beta comes first)
		if v1 != v2 {
			//fmt.Printf("Less than %v (%s %s) this: %s %v that: %s %v\n", v1 < v2, v1, v2, this, thisMatches, that, thatMatches)
			return v1 > v2
		}
	}

	// They have the same value
	return false
}
func (a ApiVersion) String() string {
	return string(a)
}
