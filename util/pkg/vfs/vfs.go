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

package vfs

import (
	"fmt"
	"io"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/util/pkg/hashing"
)

// Yet another VFS package
// If there's a "winning" VFS implementation in go, we should switch to it!

type VFS interface {
}

func IsDirectory(p Path) bool {
	_, err := p.ReadDir()
	return err == nil
}

type ACL interface {
}

type ACLOracle func(Path) (ACL, error)

// Path is a path in the VFS space, which we can read, write, list etc
type Path interface {
	io.WriterTo

	Join(relativePath ...string) Path

	// ReadFile returns the contents of the file, or an error if the file could not be read.
	// If the file did not exist, err = os.ErrNotExist
	// As this reads the entire file into memory, consider using WriteTo for bigger files
	ReadFile() ([]byte, error)

	WriteFile(data io.ReadSeeker, acl ACL) error
	// CreateFile writes the file contents, but only if the file does not already exist
	CreateFile(data io.ReadSeeker, acl ACL) error

	// Remove deletes the file
	Remove() error

	// RemoveAllVersions completely deletes the file (with all its versions and markers).
	RemoveAllVersions() error

	// Base returns the base name (last element)
	Base() string

	// Path returns a string representing the full path
	Path() string

	// ReadDir lists the files in a particular Path
	ReadDir() ([]Path, error)

	// ReadTree lists all files (recursively) in the subtree rooted at the current Path
	/// Note: returns only files, not directories
	ReadTree() ([]Path, error)
}

type HasHash interface {
	// Returns the hash of the file contents, with the preferred hash algorithm
	PreferredHash() (*hashing.Hash, error)

	// Gets the hash, or nil if the hash cannot be (easily) computed
	Hash(algorithm hashing.HashAlgorithm) (*hashing.Hash, error)
}

func RelativePath(base Path, child Path) (string, error) {
	basePath := base.Path()
	childPath := child.Path()
	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}

	if !strings.HasPrefix(childPath, basePath) {
		return "", fmt.Errorf("Path %q is not a child of %q", child, base)
	}

	relativePath := childPath[len(basePath):]
	return relativePath, nil
}

func IsClusterReadable(p Path) bool {
	if hcr, ok := p.(HasClusterReadable); ok {
		return hcr.IsClusterReadable()
	}

	switch p.(type) {
	case *S3Path, *GSPath, *SwiftPath, *OSSPath, *FSPath, *VaultPath:
		return true

	case *KubernetesPath:
		return true

	case *SSHPath:
		return false

	case *MemFSPath:
		return false

	default:
		klog.Fatalf("IsClusterReadable not implemented for type %T", p)
		return false
	}
}

type HasClusterReadable interface {
	IsClusterReadable() bool
}
