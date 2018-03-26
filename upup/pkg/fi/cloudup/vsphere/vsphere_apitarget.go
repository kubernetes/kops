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

package vsphere

// vsphere_target represents API execution target for vSphere.

import "k8s.io/kops/upup/pkg/fi"

// VSphereAPITarget represents target for vSphere, where cluster deployment with take place.
type VSphereAPITarget struct {
	Cloud *VSphereCloud
}

var _ fi.Target = &VSphereAPITarget{}

// NewVSphereAPITarget returns VSphereAPITarget instance for vSphere cloud provider.
func NewVSphereAPITarget(cloud *VSphereCloud) *VSphereAPITarget {
	return &VSphereAPITarget{
		Cloud: cloud,
	}
}

// Finish is no-op for vSphere cloud.
func (t *VSphereAPITarget) Finish(taskMap map[string]fi.Task) error {
	return nil
}

// ProcessDeletions is no-op for vSphere cloud.
func (t *VSphereAPITarget) ProcessDeletions() bool {
	return true
}
