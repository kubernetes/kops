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

//go:generate fitask -type=MirrorKeystore
type MirrorKeystore struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	MirrorPath vfs.Path
}

var _ fi.HasDependencies = &MirrorKeystore{}

// GetDependencies returns the dependencies for a MirrorKeystore task - it must run after all secrets have been run
func (e *MirrorKeystore) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task
	for _, task := range tasks {
		if _, ok := task.(*Secret); ok {
			deps = append(deps, task)
		}
	}
	return deps
}

func (e *MirrorKeystore) Find(c *fi.Context) (*MirrorKeystore, error) {
	// TODO:
	return nil, nil
}

func (e *MirrorKeystore) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *MirrorKeystore) CheckChanges(a, e, changes *MirrorKeystore) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
	}
	return nil
}

func (_ *MirrorKeystore) Render(c *fi.Context, a, e, changes *MirrorKeystore) error {
	keystore := c.Keystore

	return keystore.MirrorTo(e.MirrorPath)
}
