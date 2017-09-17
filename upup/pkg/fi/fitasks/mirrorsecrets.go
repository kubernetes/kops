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

package fitasks

import (
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

//go:generate fitask -type=MirrorSecrets
type MirrorSecrets struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	MirrorPath vfs.Path
}

var _ fi.HasDependencies = &MirrorSecrets{}

// GetDependencies returns the dependencies for a MirrorSecrets task - it must run after all secrets have been run
func (e *MirrorSecrets) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task
	for _, task := range tasks {
		if _, ok := task.(*Secret); ok {
			deps = append(deps, task)
		}
	}
	return deps
}

func (e *MirrorSecrets) Find(c *fi.Context) (*MirrorSecrets, error) {
	// TODO:
	return nil, nil
}

func (e *MirrorSecrets) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *MirrorSecrets) CheckChanges(a, e, changes *MirrorSecrets) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
	}
	return nil
}

func (_ *MirrorSecrets) Render(c *fi.Context, a, e, changes *MirrorSecrets) error {
	secrets := c.SecretStore

	return secrets.MirrorTo(e.MirrorPath)
}
