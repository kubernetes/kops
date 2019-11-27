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

package validation

import (
	"net/url"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

func isValidAPIServersURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	if u.Host == "" || u.Scheme == "" {
		return false
	}
	return true
}

func IsValidValue(fldPath *field.Path, v *string, validValues []string) field.ErrorList {
	allErrs := field.ErrorList{}
	if v != nil {
		found := false
		for _, validValue := range validValues {
			if *v == validValue {
				found = true
				break
			}
		}
		if !found {
			allErrs = append(allErrs, field.NotSupported(fldPath, *v, validValues))
		}
	}
	return allErrs
}
