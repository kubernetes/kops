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

// Package configfsi defines an interface for interaction with the TSM configfs subsystem.
package configfsi

// Client abstracts the filesystem operations for interacting with configfs files.
type Client interface {
	// MkdirTemp creates a new temporary directory in the directory dir and returns the pathname
	// of the new directory. Pattern semantics follow os.MkdirTemp.
	MkdirTemp(dir, pattern string) (string, error)
	// ReadFile reads the named file and returns the contents.
	ReadFile(name string) ([]byte, error)
	// WriteFile writes data to the named file, creating it if necessary. The permissions
	// are implementation-defined.
	WriteFile(name string, contents []byte) error
	// RemoveAll removes path and any children it contains.
	RemoveAll(path string) error
}
