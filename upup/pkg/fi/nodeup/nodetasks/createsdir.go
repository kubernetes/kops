/*
Copyright 2017 The Kubernetes Authors.

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

package nodetasks

import (
	"strings"

	"k8s.io/kops/upup/pkg/fi"
)

// CreatesDir is a marker interface for tasks that create directories, used for dependencies
type CreatesDir interface {
	Dir() string
}

var _ CreatesDir = &File{}

// findCreatesDirParents finds the tasks which create parent directories for the given task
func findCreatesDirParents(p string, tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task
	for _, v := range tasks {
		if createsDirectory, ok := v.(CreatesDir); ok {
			dirPath := createsDirectory.Dir()
			if dirPath != "" {
				if !strings.HasSuffix(dirPath, "/") {
					dirPath += "/"
				}

				if p == dirPath {
					continue
				}

				if strings.HasPrefix(p, dirPath) {
					deps = append(deps, v)
				}
			}
		}
	}
	return deps
}

// findCreatesDirMatching finds the tasks which create the specified directory (matching, non-recursive)
func findCreatesDirMatching(p string, tasks map[string]fi.Task) []fi.Task {
	if !strings.HasSuffix(p, "/") {
		p += "/"
	}

	var deps []fi.Task
	for _, v := range tasks {
		if createsDirectory, ok := v.(CreatesDir); ok {
			dirPath := createsDirectory.Dir()
			if dirPath != "" {
				if !strings.HasSuffix(dirPath, "/") {
					dirPath += "/"
				}

				if p == dirPath {
					deps = append(deps, v)
				}
			}
		}
	}
	return deps
}
