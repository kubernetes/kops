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

package local

import "k8s.io/kops/upup/pkg/fi"

type LocalTarget struct {
	CacheDir string
	Tags     map[string]struct{}
}

var _ fi.Target = &LocalTarget{}

func (t *LocalTarget) Finish(taskMap map[string]fi.Task) error {
	return nil
}

func (t *LocalTarget) HasTag(tag string) bool {
	_, found := t.Tags[tag]
	return found
}
