//go:build ignore
// +build ignore

/*
Copyright 2020 The Kubernetes Authors.

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

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// genLDFlags generates linker flags (-ldflags) for injecting version info into the binary.
func genLDFlags(ver string) string {
	var ldflagsStr string
	ldflagsStr = "-s -w -X k8s.io/kops.Version=" + ver + " "
	ldflagsStr = ldflagsStr + "-X k8s.io/kops.GitVersion=" + version() + " "
	ldflagsStr = ldflagsStr + "-X k8s.io/kops.GitCommit=" + commitID() + " "
	ldflagsStr = ldflagsStr + "-X k8s.io/kops.GitCommitDate=" + commitTime().Format(time.RFC3339) + " "
	ldflagsStr = ldflagsStr + "-X k8s.io/kops.GitTreeState=" + treeState() + " "
	return ldflagsStr
}

// version returns the version string from Git.
// Equivalent to: git describe --tags --always --match 'v*'
func version() string {
	var (
		tag []byte
		e   error
	)
	cmdName := "git"
	cmdArgs := []string{"describe", "--tags", "--always", "--match", "v*"}
	if tag, e = exec.Command(cmdName, cmdArgs...).Output(); e != nil {
		fmt.Fprintln(os.Stderr, "Error generating git version: ", e)
		os.Exit(1)
	}
	return strings.TrimSpace(string(tag))
}

// commitID returns the full commit hash of the last Git commit.
// Equivalent to: git log --format="%H" -n1
func commitID() string {
	var (
		commit []byte
		e      error
	)
	cmdName := "git"
	cmdArgs := []string{"log", "--format=%H", "-n1"}
	if commit, e = exec.Command(cmdName, cmdArgs...).Output(); e != nil {
		fmt.Fprintln(os.Stderr, "Error generating git commit-id: ", e)
		os.Exit(1)
	}

	return strings.TrimSpace(string(commit))
}

// commitTime returns the UTC time of the most recent Git commit.
func commitTime() time.Time {
	// git log --format=%cI -n1
	var (
		commitUnix []byte
		err        error
	)
	cmdName := "git"
	cmdArgs := []string{"log", "--format=%cI", "-n1"}
	if commitUnix, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		fmt.Fprintln(os.Stderr, "Error generating git commit-time: ", err)
		os.Exit(1)
	}

	t, err := time.Parse(time.RFC3339, strings.TrimSpace(string(commitUnix)))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error generating git commit-time: ", err)
		os.Exit(1)
	}

	return t.UTC()
}

// treeState returns the working tree state: "clean" or "dirty".
// Equivalent to: git status --porcelain
func treeState() string {
	var (
		out []byte
		e   error
	)
	cmdName := "git"
	cmdArgs := []string{"status", "--porcelain"}
	if out, e = exec.Command(cmdName, cmdArgs...).Output(); e != nil {
		fmt.Fprintln(os.Stderr, "Error generating git tree-state: ", e)
		os.Exit(1)
	}
	if strings.TrimSpace(string(out)) == "" {
		return "clean"
	}
	return "dirty"
}

func main() {
	var ver string
	if len(os.Args) > 1 {
		ver = strings.TrimSpace(os.Args[1])
	} else {
		ver = version()
	}
	fmt.Println(genLDFlags(ver))
}
