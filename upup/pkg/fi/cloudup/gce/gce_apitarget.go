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

package gce

import (
	"k8s.io/kops/upup/pkg/fi"
)

type GCEAPITarget struct {
	Cloud GCECloud
}

var _ fi.Target = &GCEAPITarget{}

func NewGCEAPITarget(cloud GCECloud) *GCEAPITarget {
	return &GCEAPITarget{
		Cloud: cloud,
	}
}

func (t *GCEAPITarget) Finish(taskMap map[string]fi.Task) error {
	return nil
}

func (t *GCEAPITarget) ProcessDeletions() bool {
	return true
}
