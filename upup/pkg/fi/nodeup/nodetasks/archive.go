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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strconv"

	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/cloudinit"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
	"k8s.io/kops/util/pkg/hashing"
)

// Archive task downloads and extracts a tar file
type Archive struct {
	Name string

	// Source is the location for the archive
	Source string `json:"source,omitempty"`
	// Hash is the source tar
	Hash string `json:"hash,omitempty"`

	// TargetDir is the directory for extraction
	TargetDir string `json:"target,omitempty"`

	// StripComponents is the number of components to remove when expanding the archive
	StripComponents int `json:"stripComponents,omitempty"`
}

const (
	localArchiveDir      = "/var/cache/nodeup/archives/"
	localArchiveStateDir = "/var/cache/nodeup/archives/state/"
)

var _ fi.HasDependencies = &Archive{}

// GetDependencies implements HasDependencies::GetDependencies
func (e *Archive) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task

	// Requires parent directories to be created
	deps = append(deps, findCreatesDirParents(e.TargetDir, tasks)...)

	return deps
}

var _ fi.HasName = &Archive{}

func (e *Archive) GetName() *string {
	return &e.Name
}

func (e *Archive) SetName(name string) {
	e.Name = name
}

// String returns a string representation, implementing the Stringer interface
func (e *Archive) String() string {
	return fmt.Sprintf("Archive: %s %s->%s", e.Name, e.Source, e.TargetDir)
}

var _ CreatesDir = &Archive{}

// Dir implements CreatesDir::Dir
func (e *Archive) Dir() string {
	return e.TargetDir
}

// Find implements fi.Task::Find
func (e *Archive) Find(c *fi.Context) (*Archive, error) {
	// We write a marker file to prevent re-execution
	localStateFile := path.Join(localArchiveStateDir, e.Name)
	stateBytes, err := ioutil.ReadFile(localStateFile)
	if err != nil {
		if os.IsNotExist(err) {
			stateBytes = nil
		} else {
			klog.Warningf("error reading archive state %s: %v", localStateFile, err)
			// We can just reinstall
			return nil, nil
		}
	}

	if stateBytes == nil {
		// No marker file found, assume archive not installed
		return nil, nil
	}

	state := &Archive{}
	if err := json.Unmarshal(stateBytes, state); err != nil {
		klog.Warningf("error unmarshaling archive state %s: %v", localStateFile, err)
		// We can just reinstall
		return nil, nil
	}

	if state.Hash == e.Hash && state.TargetDir == e.TargetDir {
		return state, nil
	}

	// Existing version is different, force a reinstall
	return nil, nil
}

// Run implements fi.Task::Run
func (e *Archive) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

// CheckChanges implements fi.Task::CheckChanges
func (_ *Archive) CheckChanges(a, e, changes *Archive) error {
	return nil
}

// RenderLocal implements the fi.Task::Render functionality for a local target
func (_ *Archive) RenderLocal(t *local.LocalTarget, a, e, changes *Archive) error {
	if a == nil {
		klog.Infof("Installing archive %q", e.Name)

		localFile := path.Join(localArchiveDir, e.Name)
		if err := os.MkdirAll(localArchiveDir, 0755); err != nil {
			return fmt.Errorf("error creating directories %q: %v", localArchiveDir, err)
		}

		var hash *hashing.Hash
		if e.Hash != "" {
			parsed, err := hashing.FromString(e.Hash)
			if err != nil {
				return fmt.Errorf("error parsing hash: %v", err)
			}
			hash = parsed
		}
		if _, err := fi.DownloadURL(e.Source, localFile, hash); err != nil {
			return err
		}

		targetDir := e.TargetDir
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("error creating directories %q: %v", targetDir, err)
		}

		args := []string{"tar", "xf", localFile, "-C", targetDir}
		if e.StripComponents != 0 {
			args = append(args, "--strip-components="+strconv.Itoa(e.StripComponents))
		}

		klog.Infof("running command %s", args)
		cmd := exec.Command(args[0], args[1:]...)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("error installing archive %q: %v: %s", e.Name, err, string(output))
		}

		// We write a marker file to prevent re-execution
		localStateFile := path.Join(localArchiveStateDir, e.Name)
		if err := os.MkdirAll(localArchiveStateDir, 0755); err != nil {
			return fmt.Errorf("error creating directories %q: %v", localArchiveStateDir, err)
		}

		state, err := json.MarshalIndent(e, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshaling archive state: %v", err)
		}

		if err := ioutil.WriteFile(localStateFile, state, 0644); err != nil {
			return fmt.Errorf("error writing archive state: %v", err)
		}
	} else {
		if !reflect.DeepEqual(changes, &Archive{}) {
			klog.Warningf("cannot apply archive changes for %q: %v", e.Name, changes)
		}
	}

	return nil
}

// RenderCloudInit implements fi.Task::Render functionality for a CloudInit target
func (_ *Archive) RenderCloudInit(t *cloudinit.CloudInitTarget, a, e, changes *Archive) error {
	archiveName := e.Name

	localFile := path.Join(localArchiveDir, archiveName)
	t.AddMkdirpCommand(localArchiveDir, 0755)

	targetDir := e.TargetDir
	t.AddMkdirpCommand(targetDir, 0755)

	url := e.Source
	t.AddDownloadCommand(cloudinit.Always, url, localFile)

	t.AddCommand(cloudinit.Always, "tar", "xf", localFile, "-C", targetDir)

	return nil
}
