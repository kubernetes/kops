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
	"fmt"
	"path"
	"reflect"
	"strings"
	"testing"

	"k8s.io/kops/upup/pkg/fi"
)

func TestBindMountCommands(t *testing.T) {
	containerizedMounterHome := "/containerized_mounter"

	grid := []struct {
		mount    *BindMount
		executor *MockExecutor
	}{
		{
			mount: &BindMount{
				Source:     containerizedMounterHome,
				Mountpoint: containerizedMounterHome,
				Options:    []string{"exec"},
			},
			executor: &MockExecutor{
				Commands: []*MockCommand{
					{Args: []string{"mount", "--bind", "/containerized_mounter", "/containerized_mounter"}},
					{Args: []string{"mount", "-o", "remount,exec", "/containerized_mounter"}},
				},
			},
		},
		{
			mount: &BindMount{
				Source:     "/var/lib/kubelet/",
				Mountpoint: path.Join(containerizedMounterHome, "rootfs/var/lib/kubelet"),
				Options:    []string{"rshared"},
				Recursive:  true,
			},
			executor: &MockExecutor{
				Commands: []*MockCommand{
					{Args: []string{"mount", "--rbind", "/var/lib/kubelet/", "/containerized_mounter/rootfs/var/lib/kubelet"}},
					{Args: []string{"mount", "--make-rshared", "/containerized_mounter/rootfs/var/lib/kubelet"}},
				},
			},
		},
		{
			mount: &BindMount{
				Source:     "/proc",
				Mountpoint: path.Join(containerizedMounterHome, "rootfs/proc"),
				Options:    []string{"ro"},
			},
			executor: &MockExecutor{
				Commands: []*MockCommand{
					{Args: []string{"mount", "--bind", "-o", "ro", "/proc", "/containerized_mounter/rootfs/proc"}},
				},
			},
		},
		{
			mount: &BindMount{
				Source:     "/dev",
				Mountpoint: path.Join(containerizedMounterHome, "rootfs/dev"),
				Options:    []string{"ro"},
			},
			executor: &MockExecutor{
				Commands: []*MockCommand{
					{Args: []string{"mount", "--bind", "-o", "ro", "/dev", "/containerized_mounter/rootfs/dev"}},
				},
			},
		},
		{
			mount: &BindMount{
				Source:     "/etc/resolv.conf",
				Mountpoint: path.Join(containerizedMounterHome, "rootfs/etc/resolv.conf"),
				Options:    []string{"ro"},
			},
			executor: &MockExecutor{
				Commands: []*MockCommand{
					{Args: []string{"mount", "--bind", "-o", "ro", "/etc/resolv.conf", "/containerized_mounter/rootfs/etc/resolv.conf"}},
				},
			},
		},
	}

	for _, g := range grid {
		err := g.mount.execute(g.executor)
		if err != nil {
			t.Errorf("unexpected error from %v: %v", g.mount, err)
		}
		if len(g.executor.Commands) != 0 {
			t.Errorf("not all expected commands were called: %s", g.executor.Commands)
		}
	}
}

func TestBindMountDependencies(t *testing.T) {
	containerizedMounterHome := "/containerized_mounter"

	grid := []struct {
		parent fi.Task
		child  fi.Task
	}{
		{
			parent: &MountDiskTask{
				Mountpoint: "/",
			},
			child: &BindMount{
				Source:     containerizedMounterHome,
				Mountpoint: containerizedMounterHome,
				Options:    []string{"exec"},
			},
		},
		{
			parent: &File{
				Path: containerizedMounterHome,
				Type: FileType_Directory,
			},
			child: &BindMount{
				Source:     containerizedMounterHome,
				Mountpoint: containerizedMounterHome,
				Options:    []string{"exec"},
			},
		},
		{
			parent: &Archive{
				TargetDir: containerizedMounterHome,
			},
			child: &BindMount{
				Source:     containerizedMounterHome,
				Mountpoint: containerizedMounterHome,
				Options:    []string{"exec"},
			},
		},
		{
			parent: &BindMount{
				Source:     containerizedMounterHome,
				Mountpoint: containerizedMounterHome,
				Options:    []string{"exec"},
			},
			child: &BindMount{
				Source:     "/var/lib/kubelet/",
				Mountpoint: path.Join(containerizedMounterHome, "rootfs/var/lib/kubelet"),
				Options:    []string{"rshared"},
				Recursive:  true,
			},
		},
		{
			parent: &BindMount{
				Source:     containerizedMounterHome,
				Mountpoint: containerizedMounterHome,
				Options:    []string{"exec"},
			},
			child: &BindMount{
				Source:     "/proc",
				Mountpoint: path.Join(containerizedMounterHome, "rootfs/proc"),
				Options:    []string{"ro"},
			},
		},
		{
			parent: &BindMount{
				Source:     containerizedMounterHome,
				Mountpoint: containerizedMounterHome,
				Options:    []string{"exec"},
			},
			child: &BindMount{
				Source:     "/dev",
				Mountpoint: path.Join(containerizedMounterHome, "rootfs/dev"),
				Options:    []string{"ro"},
			},
		},
		{
			parent: &BindMount{
				Source:     containerizedMounterHome,
				Mountpoint: containerizedMounterHome,
				Options:    []string{"exec"},
			},
			child: &BindMount{
				Source:     "/etc/resolv.conf",
				Mountpoint: path.Join(containerizedMounterHome, "rootfs/etc/resolv.conf"),
				Options:    []string{"ro"},
			},
		},
	}

	for _, g := range grid {
		allTasks := make(map[string]fi.Task)
		allTasks["parent"] = g.parent
		allTasks["child"] = g.child

		deps := g.parent.(fi.HasDependencies).GetDependencies(allTasks)
		if len(deps) != 0 {
			t.Errorf("found unexpected dependencies for parent: %v %v", g.parent, deps)
		}

		childDeps := g.child.(fi.HasDependencies).GetDependencies(allTasks)
		if len(childDeps) != 1 {
			t.Errorf("found unexpected dependencies for child: %v %v", g.child, childDeps)
		}
	}
}

// MockExecutor is a mock implementation of Executor
type MockExecutor struct {
	Commands []*MockCommand
}

type MockCommand struct {
	Args   []string
	Result []byte
	Error  error
}

func (c *MockCommand) String() string {
	return strings.Join(c.Args, " ")
}

var _ Executor = &MockExecutor{}

func (m *MockExecutor) Expect(args []string) *MockCommand {
	c := &MockCommand{
		Args: args,
	}
	m.Commands = append(m.Commands, c)
	return c
}

func (m *MockExecutor) CombinedOutput(args []string) ([]byte, error) {
	key := strings.Join(args, " ")
	if len(m.Commands) == 0 {
		return nil, fmt.Errorf("unexpected command %q", key)
	}
	c := m.Commands[0]
	if !reflect.DeepEqual(args, c.Args) {
		return nil, fmt.Errorf("unexpected command %q", key)
	}
	m.Commands = m.Commands[1:]
	return c.Result, c.Error
}
