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
	"k8s.io/kops/upup/pkg/fi/utils"
)

// UserTask is responsible for creating a user, by calling useradd
type UserTask struct {
	Name string

	UID   int    `json:"uid"`
	Shell string `json:"shell"`
	Home  string `json:"home"`
}

var _ fi.Task = &UserTask{}

func (e *UserTask) String() string {
	return fmt.Sprintf("User: %s", e.Name)
}

var _ fi.HasName = &File{}

func (f *UserTask) GetName() *string {
	return &f.Name
}

func (f *UserTask) SetName(name string) {
	klog.Fatalf("SetName not supported for User task")
}

func NewUserTask(name string, contents string, meta string) (fi.Task, error) {
	s := &UserTask{Name: name}

	err := utils.YamlUnmarshal([]byte(contents), s)
	if err != nil {
		return nil, fmt.Errorf("error parsing json for service %q: %v", name, err)
	}

	return s, nil
}

func (e *UserTask) Find(c *fi.Context) (*UserTask, error) {
	info, err := fi.LookupUser(e.Name)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, nil
	}

	actual := &UserTask{
		Name:  e.Name,
		UID:   info.Uid,
		Shell: info.Shell,
		Home:  info.Home,
	}

	return actual, nil
}

func (e *UserTask) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *UserTask) CheckChanges(a, e, changes *UserTask) error {
	return nil
}

func buildUseraddArgs(e *UserTask) []string {
	var args []string
	if e.UID != 0 {
		args = append(args, "-u", strconv.Itoa(e.UID))
	}
	if e.Shell != "" {
		args = append(args, "-s", e.Shell)
	}
	if e.Home != "" {
		args = append(args, "-d", e.Home)
	}
	args = append(args, e.Name)
	return args
}

func (_ *UserTask) RenderLocal(t *local.LocalTarget, a, e, changes *UserTask) error {
	if a == nil {
		args := buildUseraddArgs(e)
		klog.Infof("Creating user %q", e.Name)
		cmd := exec.Command("useradd", args...)
		klog.V(2).Infof("running command: useradd %s", strings.Join(args, " "))
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error creating user: %v\nOutput: %s", err, output)
		}
	} else {
		var args []string

		if changes.UID != 0 {
			args = append(args, "-u", strconv.Itoa(e.UID))
		}
		if changes.Shell != "" {
			args = append(args, "-s", e.Shell)
		}
		if changes.Home != "" {
			args = append(args, "-d", e.Home)
		}

		if len(args) != 0 {
			args = append(args, e.Name)
			klog.Infof("Reconfiguring user %q", e.Name)
			cmd := exec.Command("usermod", args...)
			klog.V(2).Infof("running command: usermod %s", strings.Join(args, " "))
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("error reconfiguring user: %v\nOutput: %s", err, output)
			}
		}
	}

	return nil
}

func (_ *UserTask) RenderCloudInit(t *cloudinit.CloudInitTarget, a, e, changes *UserTask) error {
	args := buildUseraddArgs(e)
	cmd := []string{"useradd"}
	cmd = append(cmd, args...)
	klog.Infof("Creating user %q", e.Name)
	t.AddCommand(cloudinit.Once, cmd...)

	return nil
}
