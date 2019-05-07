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

package nodetasks

import (
	"fmt"
	"strings"

	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/cloudinit"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
)

// Chattr performs a chattr command, in particular to set a file as immutable
type Chattr struct {
	File string `json:"file"`
	Mode string `json:"mode"`

	Deps []fi.Task `json:"-"`
}

var _ fi.Task = &Chattr{}

func (s *Chattr) String() string {
	return fmt.Sprintf("Chattr: chattr %s %s", s.Mode, s.File)
}

var _ fi.HasName = &Archive{}

func (e *Chattr) GetName() *string {
	return fi.String("Chattr-" + e.File)
}

func (e *Chattr) SetName(name string) {
	klog.Fatalf("SetName not supported for Chattr task")
}

var _ fi.HasDependencies = &Chattr{}

// GetDependencies implements HasDependencies::GetDependencies
func (e *Chattr) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return e.Deps
}

func (e *Chattr) Find(c *fi.Context) (*Chattr, error) {
	// We always re-run the chattr command
	return nil, nil
}

func (e *Chattr) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *Chattr) CheckChanges(a, e, changes *Chattr) error {
	return nil
}

func (_ *Chattr) RenderLocal(t *local.LocalTarget, a, e, changes *Chattr) error {
	return e.execute(t)
}

func (e *Chattr) execute(t Executor) error {
	chattrCommand := []string{"chattr", e.Mode, e.File}

	klog.Infof("running chattr command chattr %s", chattrCommand)
	if output, err := t.CombinedOutput(chattrCommand); err != nil {
		return fmt.Errorf("error doing %q: %v: %s", strings.Join(chattrCommand, " "), err, string(output))
	}

	return nil
}

func (_ *Chattr) RenderCloudInit(t *cloudinit.CloudInitTarget, a, e, changes *Chattr) error {
	return fmt.Errorf("Chattr::RenderCloudInit not implemented")
}
