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
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/cloudinit"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
	"k8s.io/kops/upup/pkg/fi/utils"

	"k8s.io/klog"
)

const (
	// FileType_Symlink defines a symlink
	FileType_Symlink = "symlink"
	// FileType_Directory defines a directory
	FileType_Directory = "directory"
	// FileType_File is a regular file
	FileType_File = "file"
)

type File struct {
	AfterFiles      []string    `json:"afterfiles,omitempty"`
	Contents        fi.Resource `json:"contents,omitempty"`
	Group           *string     `json:"group,omitempty"`
	IfNotExists     bool        `json:"ifNotExists,omitempty"`
	Mode            *string     `json:"mode,omitempty"`
	OnChangeExecute [][]string  `json:"onChangeExecute,omitempty"`
	Owner           *string     `json:"owner,omitempty"`
	Path            string      `json:"path,omitempty"`
	Symlink         *string     `json:"symlink,omitempty"`
	Type            string      `json:"type"`
}

var _ fi.Task = &File{}
var _ fi.HasDependencies = &File{}
var _ fi.HasName = &File{}

func NewFileTask(name string, src fi.Resource, destPath string, meta string) (*File, error) {
	f := &File{
		//Name:     name,
		Contents: src,
		Path:     destPath,
	}

	if meta != "" {
		err := utils.YamlUnmarshal([]byte(meta), f)
		if err != nil {
			return nil, fmt.Errorf("error parsing meta for file %q: %v", name, err)
		}
	}

	if f.Symlink != nil && f.Type == "" {
		f.Type = FileType_Symlink
	}

	return f, nil
}

var _ fi.HasDependencies = &File{}

// GetDependencies implements HasDependencies::GetDependencies
func (e *File) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task
	if e.Owner != nil {
		ownerTask := tasks["UserTask/"+*e.Owner]
		if ownerTask == nil {
			// The user might be a pre-existing user (e.g. admin)
			klog.Warningf("Unable to find task %q", "UserTask/"+*e.Owner)
		} else {
			deps = append(deps, ownerTask)
		}
	}

	// Depend on disk mounts
	// For simplicity, we just depend on _all_ disk mounts
	// We could check the mountpath, but that feels excessive...
	for _, v := range tasks {
		if _, ok := v.(*MountDiskTask); ok {
			deps = append(deps, v)
		}
	}

	// Requires parent directories to be created
	deps = append(deps, findCreatesDirParents(e.Path, tasks)...)

	// Requires other files to be created first
	for _, f := range e.AfterFiles {
		for _, v := range tasks {
			if file, ok := v.(*File); ok {
				if file.Path == f {
					deps = append(deps, v)
				}
			}
		}
	}

	return deps
}

var _ fi.HasName = &File{}

func (f *File) GetName() *string {
	return &f.Path
}

func (f *File) SetName(name string) {
	klog.Fatalf("SetName not supported for File task")
}

func (f *File) String() string {
	return fmt.Sprintf("File: %q", f.Path)
}

var _ CreatesDir = &File{}

// Dir implements CreatesDir::Dir
func (f *File) Dir() string {
	if f.Type != FileType_Directory {
		return ""
	}
	return f.Path
}

func findFile(p string) (*File, error) {
	stat, err := os.Lstat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
	}

	actual := &File{}
	actual.Path = p
	actual.Mode = fi.String(fi.FileModeToString(stat.Mode() & os.ModePerm))

	uid := int(stat.Sys().(*syscall.Stat_t).Uid)
	owner, err := fi.LookupUserById(uid)
	if err != nil {
		return nil, err
	}
	if owner != nil {
		actual.Owner = fi.String(owner.Name)
	} else {
		actual.Owner = fi.String(strconv.Itoa(uid))
	}

	gid := int(stat.Sys().(*syscall.Stat_t).Gid)
	group, err := fi.LookupGroupById(gid)
	if err != nil {
		return nil, err
	}
	if group != nil {
		actual.Group = fi.String(group.Name)
	} else {
		actual.Group = fi.String(strconv.Itoa(gid))
	}

	if (stat.Mode() & os.ModeSymlink) != 0 {
		target, err := os.Readlink(p)
		if err != nil {
			return nil, fmt.Errorf("error reading symlink target: %v", err)
		}

		actual.Type = FileType_Symlink
		actual.Symlink = fi.String(target)
	} else if (stat.Mode() & os.ModeDir) != 0 {
		actual.Type = FileType_Directory
	} else {
		actual.Type = FileType_File
		actual.Contents = fi.NewFileResource(p)
	}

	return actual, nil
}

func (e *File) Find(c *fi.Context) (*File, error) {
	actual, err := findFile(e.Path)
	if err != nil {
		return nil, err
	}
	if actual == nil {
		return nil, nil
	}

	// To avoid spurious changes
	actual.IfNotExists = e.IfNotExists
	if e.IfNotExists {
		actual.Contents = e.Contents
	}
	actual.OnChangeExecute = e.OnChangeExecute

	return actual, nil
}

func (e *File) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *File) CheckChanges(a, e, changes *File) error {
	return nil
}

func (_ *File) RenderLocal(t *local.LocalTarget, a, e, changes *File) error {
	dirMode := os.FileMode(0755)
	fileMode, err := fi.ParseFileMode(fi.StringValue(e.Mode), 0644)
	if err != nil {
		return fmt.Errorf("invalid file mode for %q: %q", e.Path, fi.StringValue(e.Mode))
	}

	if a != nil {
		if e.IfNotExists {
			klog.V(2).Infof("file exists and IfNotExists set; skipping %q", e.Path)
			return nil
		}
	}

	changed := false
	if e.Type == FileType_Symlink {
		if changes.Symlink != nil {
			// This will currently fail if the target already exists.
			// That's probably a good thing for now ... it is hard to know what to do here!
			klog.Infof("Creating symlink %q -> %q", e.Path, *changes.Symlink)
			err := os.Symlink(*changes.Symlink, e.Path)
			if err != nil {
				return fmt.Errorf("error creating symlink %q -> %q: %v", e.Path, *changes.Symlink, err)
			}
			changed = true
		}
	} else if e.Type == FileType_Directory {
		if a == nil {
			parent := filepath.Dir(strings.TrimSuffix(e.Path, "/"))
			err := os.MkdirAll(parent, dirMode)
			if err != nil {
				return fmt.Errorf("error creating parent directories %q: %v", parent, err)
			}

			err = os.MkdirAll(e.Path, fileMode)
			if err != nil {
				return fmt.Errorf("error creating directory %q: %v", e.Path, err)
			}
			changed = true
		}
	} else if e.Type == FileType_File {
		if changes.Contents != nil {
			err = fi.WriteFile(e.Path, e.Contents, fileMode, dirMode)
			if err != nil {
				return fmt.Errorf("error copying file %q: %v", e.Path, err)
			}
			changed = true
		}
	} else {
		return fmt.Errorf("File type=%q not valid/supported", e.Type)
	}

	if changes.Mode != nil {
		modeChanged, err := fi.EnsureFileMode(e.Path, fileMode)
		if err != nil {
			return fmt.Errorf("error changing mode on %q: %v", e.Path, err)
		}
		changed = changed || modeChanged
	}

	if changes.Owner != nil || changes.Group != nil {
		ownerChanged, err := fi.EnsureFileOwner(e.Path, fi.StringValue(e.Owner), fi.StringValue(e.Group))
		if err != nil {
			return fmt.Errorf("error changing owner/group on %q: %v", e.Path, err)
		}
		changed = changed || ownerChanged
	}

	if changed && e.OnChangeExecute != nil {
		for _, args := range e.OnChangeExecute {
			human := strings.Join(args, " ")

			klog.Infof("Changed; will execute OnChangeExecute command: %q", human)

			cmd := exec.Command(args[0], args[1:]...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("error executing command %q: %v\nOutput: %s", human, err, output)
			}
		}
	}

	return nil
}

func (_ *File) RenderCloudInit(t *cloudinit.CloudInitTarget, a, e, changes *File) error {
	dirMode := os.FileMode(0755)
	fileMode, err := fi.ParseFileMode(fi.StringValue(e.Mode), 0644)
	if err != nil {
		return fmt.Errorf("invalid file mode for %s: %q", e.Path, *e.Mode)
	}

	if e.Type == FileType_Symlink {
		t.AddCommand(cloudinit.Always, "ln", "-s", fi.StringValue(e.Symlink), e.Path)
	} else if e.Type == FileType_Directory {
		parent := filepath.Dir(strings.TrimSuffix(e.Path, "/"))
		t.AddCommand(cloudinit.Once, "mkdir", "-p", "-m", fi.FileModeToString(dirMode), parent)
		t.AddCommand(cloudinit.Once, "mkdir", "-m", fi.FileModeToString(dirMode), e.Path)
	} else if e.Type == FileType_File {
		err = t.WriteFile(e.Path, e.Contents, fileMode, dirMode)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("File type=%q not valid/supported", e.Type)
	}

	if e.Owner != nil || e.Group != nil {
		t.Chown(e.Path, fi.StringValue(e.Owner), fi.StringValue(e.Group))
	}

	if e.OnChangeExecute != nil {
		return fmt.Errorf("OnChangeExecute not supported with CloudInit")
		//t.AddCommand(cloudinit.Always, e.OnChangeExecute...)
	}

	return nil
}
