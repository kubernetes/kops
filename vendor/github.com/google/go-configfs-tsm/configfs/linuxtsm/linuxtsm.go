// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package linuxtsm defines a configfsi.Client for Linux OS operations on configfs.
package linuxtsm

import (
	"fmt"
	"os"
	"path"

	"github.com/google/go-configfs-tsm/configfs/configfsi"
)

// client provides configfsi.Client for /sys/kernel/config/tsm file operations in Linux.
type client struct{}

// MkdirTemp creates a new temporary directory in the directory dir and returns the pathname
// of the new directory. Pattern semantics follow os.MkdirTemp.
func (*client) MkdirTemp(dir, pattern string) (string, error) {
	return os.MkdirTemp(dir, pattern)
}

// ReadFile reads the named file and returns the contents.
func (*client) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

// WriteFile writes data to the named file, creating it if necessary. The permissions
// are implementation-defined.
func (*client) WriteFile(name string, contents []byte) error {
	return os.WriteFile(name, contents, 0220)
}

// RemoveAll removes path and any children it contains.
func (*client) RemoveAll(path string) error {
	return os.Remove(path)
}

// MakeClient returns a "real" client for using configfs for TSM use.
func MakeClient() (configfsi.Client, error) {
	// Linux client expects just the "report" subsystem for now.
	checkPath := path.Join(configfsi.TsmPrefix, "report")
	info, err := os.Stat(checkPath)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("expected %s to be a directory", checkPath)
	}
	return &client{}, nil
}
