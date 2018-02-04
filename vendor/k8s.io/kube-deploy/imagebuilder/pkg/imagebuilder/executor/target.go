/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

// TODO: This probably should be moved to / shared with nodeup, because then it would allow action on a remote target

package executor

import (
	"io"
	"os"
)

// NewTarget constructs a new target
func NewTarget(executor Executor) *Target {
	return &Target{
		executor: executor,
	}
}

// SSH holds an SSH client, and adds utilities like SCP functionality
type Target struct {
	executor Executor
}

func (t *Target) Put(dest string, length int, content io.Reader, mode os.FileMode) error {
	return t.executor.Put(dest, length, content, mode)
}

func (t *Target) Mkdir(dest string, mode os.FileMode) error {
	return t.executor.Mkdir(dest, mode)
}

// CommandExecution helps us build a command for running
type CommandExecution struct {
	Command  []string
	Cwd      string
	Env      map[string]string
	Sudo     bool
	executor Executor
}

// WithSudo indicates that the command should be executed with sudo
func (c *CommandExecution) WithSudo() *CommandExecution {
	c.Sudo = true
	return c
}

// WithCwd sets the directory in which the command will execute
func (c *CommandExecution) WithCwd(cwd string) *CommandExecution {
	c.Cwd = cwd
	return c
}

// Setenv sets an environment variable for the command execution
func (c *CommandExecution) Setenv(k, v string) *CommandExecution {
	c.Env[k] = v
	return c
}

// Run executes the command
func (c *CommandExecution) Run() error {
	return c.executor.Run(c)
}

// Command builds a CommandExecution bound to the current SSH target
func (s *Target) Command(cmd ...string) *CommandExecution {
	c := &CommandExecution{
		executor: s.executor,
		Command:  cmd,
		Env:      make(map[string]string),
	}
	return c
}

// Exec executes a command against the SSH target
func (s *Target) Exec(cmd ...string) error {
	c := s.Command(cmd...)
	return c.Run()
}
