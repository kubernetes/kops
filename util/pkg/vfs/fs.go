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
	"io/ioutil"
	"os"
	"path"
	"sync"
	"syscall"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/try"
	"k8s.io/kops/util/pkg/hashing"
)

type FSPath struct {
	location string
}

var _ Path = &FSPath{}
var _ HasHash = &FSPath{}

func NewFSPath(location string) *FSPath {
	return &FSPath{location: location}
}

func (p *FSPath) Join(relativePath ...string) Path {
	args := []string{p.location}
	args = append(args, relativePath...)
	joined := path.Join(args...)
	return &FSPath{location: joined}
}

func (p *FSPath) WriteFile(data io.ReadSeeker, acl ACL) error {
	dir := path.Dir(p.location)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("error creating directories %q: %v", dir, err)
	}

	f, err := ioutil.TempFile(dir, "tmp")
	if err != nil {
		return fmt.Errorf("error creating temp file in %q: %v", dir, err)
	}

	// Note from here on in we have to close f and delete or rename the temp file
	tempfile := f.Name()

	_, err = io.Copy(f, data)

	if closeErr := f.Close(); err == nil {
		err = closeErr
	}

	if err == nil {
		err = os.Rename(tempfile, p.location)
		if err != nil {
			err = fmt.Errorf("error during file write of %q: rename failed: %v", p.location, err)
		}
	}

	if err == nil {
		return nil
	}

	// Something went wrong; try to remove the temp file
	if removeErr := os.Remove(tempfile); removeErr != nil {
		klog.Warningf("unable to remove temp file %q: %v", tempfile, removeErr)
	}

	return err
}

// To prevent concurrent creates on the same file while maintaining atomicity of writes,
// we take a process-wide lock during the operation.
// Not a great approach, but fine for a single process (with low concurrency)
// TODO: should we take a file lock or equivalent here?  Can we use RENAME_NOREPLACE ?
var createFileLock sync.Mutex

func (p *FSPath) CreateFile(data io.ReadSeeker, acl ACL) error {
	createFileLock.Lock()
	defer createFileLock.Unlock()

	// Check if exists
	_, err := os.Stat(p.location)
	if err == nil {
		return os.ErrExist
	}

	if !os.IsNotExist(err) {
		return err
	}

	return p.WriteFile(data, acl)
}

// ReadFile implements Path::ReadFile
func (p *FSPath) ReadFile() ([]byte, error) {
	file, err := ioutil.ReadFile(p.location)
	if err == syscall.ENOENT {
		err = os.ErrNotExist
	}
	return file, err
}

// WriteTo implements io.WriterTo
func (p *FSPath) WriteTo(out io.Writer) (int64, error) {
	f, err := os.Open(p.location)
	if err != nil {
		return 0, err
	}
	defer try.CloseFile(f)

	return io.Copy(out, f)
}

func (p *FSPath) ReadDir() ([]Path, error) {
	files, err := ioutil.ReadDir(p.location)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, err
	}
	var paths []Path
	for _, f := range files {
		paths = append(paths, NewFSPath(path.Join(p.location, f.Name())))
	}
	return paths, nil
}

func (p *FSPath) ReadTree() ([]Path, error) {
	var paths []Path
	err := readTree(p.location, &paths)
	if err != nil {
		return nil, err
	}
	return paths, nil
}

// readTree recursively finds files and adds them to dest
// It excludes directories.
func readTree(base string, dest *[]Path) error {
	files, err := ioutil.ReadDir(base)
	if err != nil {
		return err
	}
	for _, f := range files {
		p := path.Join(base, f.Name())
		if f.IsDir() {
			err = readTree(p, dest)
			if err != nil {
				return err
			}
		} else {
			*dest = append(*dest, NewFSPath(p))
		}
	}
	return nil
}

func (p *FSPath) Base() string {
	return path.Base(p.location)
}

func (p *FSPath) Path() string {
	return p.location
}

func (p *FSPath) String() string {
	return p.Path()
}

func (p *FSPath) Remove() error {
	return os.Remove(p.location)
}

func (p *FSPath) RemoveAllVersions() error {
	return p.Remove()
}

func (p *FSPath) PreferredHash() (*hashing.Hash, error) {
	return p.Hash(hashing.HashAlgorithmSHA256)
}

func (p *FSPath) Hash(a hashing.HashAlgorithm) (*hashing.Hash, error) {
	klog.V(2).Infof("hashing file %q", p.location)

	return a.HashFile(p.location)
}
