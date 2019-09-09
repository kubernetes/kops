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
	"os/exec"
	"strconv"
	"strings"

	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/cloudinit"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
)

// GroupTask is responsible for creating a group, by calling groupadd
type GroupTask struct {
	Name   string
	GID    *int
	System bool
}

var _ fi.Task = &GroupTask{}

func (e *GroupTask) String() string {
	return fmt.Sprintf("Group: %s", e.Name)
}

var _ fi.HasName = &File{}

func (f *GroupTask) GetName() *string {
	return &f.Name
}

func (f *GroupTask) SetName(name string) {
	klog.Fatalf("SetName not supported for Group task")
}

func (e *GroupTask) Find(c *fi.Context) (*GroupTask, error) {
	info, err := fi.LookupGroup(e.Name)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, nil
	}

	gid := info.Gid
	actual := &GroupTask{
		Name: e.Name,
		GID:  &gid,
	}

	// Avoid spurious changes
	actual.System = e.System

	return actual, nil
}

func (e *GroupTask) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *GroupTask) CheckChanges(a, e, changes *GroupTask) error {
	return nil
}

func buildGroupaddArgs(e *GroupTask) []string {
	var args []string
	if e.GID != nil {
		args = append(args, "-g", strconv.Itoa(*e.GID))
	}
	if e.System {
		args = append(args, "--system")
	}
	args = append(args, e.Name)
	return args
}

func (_ *GroupTask) RenderLocal(t *local.LocalTarget, a, e, changes *GroupTask) error {
	if a == nil {
		args := buildGroupaddArgs(e)
		klog.Infof("Creating group %q", e.Name)
		cmd := exec.Command("groupadd", args...)
		klog.V(2).Infof("running command: groupadd %s", strings.Join(args, " "))
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error creating group: %v\nOutput: %s", err, output)
		}
	} else {
		var args []string

		if changes.GID != nil {
			args = append(args, "-g", strconv.Itoa(*e.GID))
		}

		if len(args) != 0 {
			args = append(args, e.Name)
			klog.Infof("Reconfiguring group %q", e.Name)
			cmd := exec.Command("groupmod", args...)
			klog.V(2).Infof("running command: groupmod %s", strings.Join(args, " "))
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("error reconfiguring group: %v\nOutput: %s", err, output)
			}
		}
	}

	return nil
}

func (_ *GroupTask) RenderCloudInit(t *cloudinit.CloudInitTarget, a, e, changes *GroupTask) error {
	args := buildGroupaddArgs(e)
	cmd := []string{"groupadd"}
	cmd = append(cmd, args...)
	klog.Infof("Creating group %q", e.Name)
	t.AddCommand(cloudinit.Once, cmd...)

	return nil
}
