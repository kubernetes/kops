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
	"io/ioutil"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/cloudinit"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
)

// BindMount performs bind mounts
type BindMount struct {
	Source     string   `json:"source"`
	Mountpoint string   `json:"mountpoint"`
	Options    []string `json:"options,omitempty"`
	Recursive  bool     `json:"recursive"`
}

var _ fi.Task = &BindMount{}

func (s *BindMount) String() string {
	return fmt.Sprintf("BindMount: %s->%s", s.Source, s.Mountpoint)
}

var _ CreatesDir = &BindMount{}

// Dir implements CreatesDir::Dir
func (e *BindMount) Dir() string {
	return e.Mountpoint
}

var _ fi.HasName = &Archive{}

func (e *BindMount) GetName() *string {
	return fi.String("BindMount-" + e.Mountpoint)
}

func (e *BindMount) SetName(name string) {
	klog.Fatalf("SetName not supported for BindMount task")
}

var _ fi.HasDependencies = &BindMount{}

// GetDependencies implements HasDependencies::GetDependencies
func (e *BindMount) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task

	// Requires parent directories to be created
	deps = append(deps, findCreatesDirParents(e.Mountpoint, tasks)...)
	for _, v := range findCreatesDirMatching(e.Mountpoint, tasks) {
		if v != e && findTaskInSlice(deps, v) == -1 {
			deps = append(deps, v)
		}
	}

	// Requires source to be created
	for _, v := range findCreatesDirParents(e.Source, tasks) {
		if findTaskInSlice(deps, v) == -1 {
			deps = append(deps, v)
		}
	}
	for _, v := range findCreatesDirMatching(e.Source, tasks) {
		if v != e && findTaskInSlice(deps, v) == -1 {
			deps = append(deps, v)
		}
	}

	return deps
}

func findTaskInSlice(tasks []fi.Task, task fi.Task) int {
	for i, t := range tasks {
		if t == task {
			return i
		}
	}
	return -1
}

func (e *BindMount) Find(c *fi.Context) (*BindMount, error) {
	mounts, err := ioutil.ReadFile("/proc/self/mountinfo")
	if err != nil {
		return nil, fmt.Errorf("error reading /proc/self/mountinfo: %v", err)
	}
	for _, line := range strings.Split(string(mounts), "\n") {
		// See `man mount_namespaces` and `man proc`
		// 534 458 8:1 /var/lib/kubelet /home/kubernetes/containerized_mounter/rootfs/var/lib/kubelet rw,nosuid,nodev,noexec,relatime shared:19 - ext4 /dev/sda1 rw,commit=30,data=ordered
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		tokens := strings.Fields(line)
		if len(tokens) < 8 {
			klog.V(4).Infof("ignoring mountinfo line: %q", line)
		}

		mountpoint := tokens[4]
		if strings.TrimSuffix(mountpoint, "/") != strings.TrimSuffix(e.Mountpoint, "/") {
			continue
		}

		fstype := tokens[len(tokens)-3]
		source := tokens[3]
		switch fstype {
		// Some special cases
		case "devtmpfs":
			source = "/dev"
		case "proc":
			source = "/proc"
		}
		if e.Source == "/etc/resolv.conf" {
			// /etc/resolv.conf is a symlink on ContainerOS, and "magically" gets transformed to /systemd/resolve/resolv.conf
			// Special case this very odd case!
			if source == "/systemd/resolve/resolv.conf" || source == "/run/systemd/resolve/resolv.conf" {
				source = e.Source // force match
			}
		}
		if strings.TrimSuffix(source, "/") != strings.TrimSuffix(e.Source, "/") {
			continue
		}

		klog.V(8).Infof("candidate mount: %v", line)

		mountOptions := sets.NewString(strings.Split(tokens[5], ",")...)
		// exec is inferred from a lack of noexec
		if !mountOptions.Has("noexec") {
			mountOptions.Insert("exec")
		}

		// optional fields: zero or more fields of the form "tag[:value]"
		for _, token := range tokens[6:] {
			if token == "-" {
				// the end of the optional fields is marked by a single hyphen.
				break
			}
			if strings.HasPrefix(token, "shared:") {
				mountOptions.Insert("rshared")
			}
		}

		if !mountOptions.HasAll(e.Options...) {
			klog.V(2).Infof("options mismatch on mount %v", line)
			continue
		}

		klog.V(2).Infof("found matching mount %v", line)
		a := &BindMount{
			Source:     e.Source,
			Mountpoint: e.Mountpoint,
			Options:    e.Options,
			Recursive:  e.Recursive, // TODO: Validate
		}
		return a, nil
	}

	return nil, nil
}

func (e *BindMount) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *BindMount) CheckChanges(a, e, changes *BindMount) error {
	return nil
}

// Executor allows execution of a command; it makes for testing of commands
type Executor interface {
	CombinedOutput(args []string) ([]byte, error)
}

func (_ *BindMount) RenderLocal(t *local.LocalTarget, a, e, changes *BindMount) error {
	return e.execute(t)
}

func (e *BindMount) execute(t Executor) error {
	var simpleOptions []string
	var makeOptions []string
	var remountOptions []string
	for _, option := range e.Options {
		switch option {
		case "ro":
			simpleOptions = append(simpleOptions, option)

		case "rshared":
			makeOptions = append(makeOptions, "--make-rshared")

		case "exec", "noexec", "nosuid", "nodev":
			remountOptions = append(remountOptions, option)

		default:
			return fmt.Errorf("unknown option: %q", option)
		}
	}

	{
		args := []string{"mount"}
		if e.Recursive {
			args = append(args, "--rbind")
		} else {
			args = append(args, "--bind")
		}
		if len(simpleOptions) != 0 {
			args = append(args, "-o", strings.Join(simpleOptions, ","))
		}
		args = append(args, e.Source, e.Mountpoint)

		klog.Infof("running mount command %s", args)
		if output, err := t.CombinedOutput(args); err != nil {
			return fmt.Errorf("error doing mount %q: %v: %s", strings.Join(args, " "), err, string(output))
		}
	}

	if len(remountOptions) != 0 {
		args := []string{"mount", "-o", "remount," + strings.Join(remountOptions, ","), e.Mountpoint}

		klog.Infof("running mount command %s", args)
		if output, err := t.CombinedOutput(args); err != nil {
			return fmt.Errorf("error doing mount options %q: %v: %s", strings.Join(args, " "), err, string(output))
		}
	}

	if len(makeOptions) != 0 {
		args := []string{"mount"}
		args = append(args, makeOptions...)
		args = append(args, e.Mountpoint)

		klog.Infof("running mount command %s", args)
		if output, err := t.CombinedOutput(args); err != nil {
			return fmt.Errorf("error doing mount operation %q: %v: %s", strings.Join(args, " "), err, string(output))
		}
	}

	return nil
}

func (_ *BindMount) RenderCloudInit(t *cloudinit.CloudInitTarget, a, e, changes *BindMount) error {
	return fmt.Errorf("BindMount::RenderCloudInit not implemented")
}
