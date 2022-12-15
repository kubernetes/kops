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

package fitasks

import (
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/secrets"
	"k8s.io/kops/util/pkg/vfs"
)

// +kops:fitask
type MirrorSecrets struct {
	Name      *string
	Lifecycle fi.Lifecycle

	MirrorPath vfs.Path
}

var _ fi.CloudupHasDependencies = &MirrorSecrets{}

// GetDependencies returns the dependencies for a MirrorSecrets task - it must run after all secrets have been run
func (e *MirrorSecrets) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	var deps []fi.CloudupTask
	for _, task := range tasks {
		if _, ok := task.(*Secret); ok {
			deps = append(deps, task)
		}
	}
	return deps
}

// Find implements fi.Task::Find
func (e *MirrorSecrets) Find(c *fi.CloudupContext) (*MirrorSecrets, error) {
	if vfsSecretStore, ok := c.T.SecretStore.(*secrets.VFSSecretStore); ok {
		if vfsSecretStore.VFSPath().Path() == e.MirrorPath.Path() {
			return e, nil
		}
	}

	// TODO: implement Find so that we aren't always mirroring
	klog.V(2).Infof("MirrorSecrets::Find not implemented; always copying (inefficient)")
	return nil, nil
}

// Run implements fi.Task::Run
func (e *MirrorSecrets) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
}

// CheckChanges implements fi.Task::CheckChanges
func (s *MirrorSecrets) CheckChanges(a, e, changes *MirrorSecrets) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
	}
	return nil
}

// Render implements fi.Task::Render
func (_ *MirrorSecrets) Render(c *fi.CloudupContext, a, e, changes *MirrorSecrets) error {
	ctx := c.Context()

	secrets := c.T.SecretStore
	return secrets.MirrorTo(ctx, e.MirrorPath)
}
