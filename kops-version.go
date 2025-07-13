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

package kops

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// These constants are parsed by build tooling - be careful about changing the formats
const (
	KOPS_RELEASE_VERSION = "1.33.0-alpha.1"
	KOPS_CI_VERSION      = "1.33.0-alpha.2"
)

var (
	// Version can be replaced by build tooling
	Version = KOPS_RELEASE_VERSION
	// GitVersion is semantic version.
	GitVersion = "v0.0.0-master+$Format:%h$"
	// GitTreeState state of git tree, either "clean" or "dirty".
	GitTreeState = ""
	// gitCommit sha1 from git
	gitCommit = ""
	// gitCommitDate date from git
	gitCommitDate = ""
)

const (
	commitKey     = "vcs.revision"
	commitDateKey = "vcs.time"
)

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	for _, setting := range info.Settings {
		if setting.Key == commitKey {
			gitCommit = setting.Value
		}
		if setting.Key == commitDateKey {
			gitCommitDate = setting.Value
		}
	}
}

// Info contains versioning information.
type Info struct {
	Version       string `json:"version"`
	GitVersion    string `json:"gitVersion"`
	GitCommit     string `json:"gitCommit"`
	GitCommitDate string `json:"gitCommitDate"`
	GitTreeState  string `json:"gitTreeState"`
	GoVersion     string `json:"goVersion"`
	Compiler      string `json:"compiler"`
	Platform      string `json:"platform"`
}

// Get returns the overall codebase version. It's for detecting
// what code a binary was built from.
func Get() Info {
	return Info{
		Version:       Version,
		GitVersion:    GitVersion,
		GitTreeState:  GitTreeState,
		GitCommit:     gitCommit,
		GitCommitDate: gitCommitDate,
		GoVersion:     runtime.Version(),
		Compiler:      runtime.Compiler,
		Platform:      fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
