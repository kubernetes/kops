/*
Copyright 2024 The Kubernetes Authors.

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

package main

import (
	"path"
	"regexp"
	"strings"

	"k8s.io/klog/v2"
)

type globList []string

func (l *globList) String() string {
	return strings.Join(*l, ",")
}

func (l *globList) Set(value string) error {
	*l = append(*l, value)
	return nil
}

func (l *globList) Matches(p string) bool {
	for _, pattern := range *l {
		// ** is not implemented by match: https://github.com/golang/go/issues/11862
		if strings.Contains(pattern, "**") {
			asRegexp := regexp.QuoteMeta(pattern)
			asRegexp = strings.ReplaceAll(asRegexp, "\\*\\*", "(.*)")
			asRegexp = strings.ReplaceAll(asRegexp, "\\*", "([^/]*)")
			asRegexp = "^" + asRegexp + "$"
			r := regexp.MustCompile(asRegexp)
			if r.MatchString(p) {
				return true
			}
		}
		match, err := path.Match(pattern, p)
		if err != nil {
			klog.Fatalf("invalid pattern %q: %v", pattern, err)
		}
		if match {
			return true
		}
	}
	return false
}
