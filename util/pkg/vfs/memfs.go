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
	"strings"
	"sync"
)

type MemFSPath struct {
	context  *MemFSContext
	location string

	mutex    sync.Mutex
	contents []byte
	children map[string]*MemFSPath
}

var _ Path = &MemFSPath{}

type MemFSContext struct {
	clusterReadable bool
	root            *MemFSPath
}

func NewMemFSContext() *MemFSContext {
	c := &MemFSContext{}
	c.root = &MemFSPath{
		context:  c,
		location: "",
	}
	return c
}

// MarkClusterReadable pretends the current memfscontext is cluster readable; this is useful for tests
func (c *MemFSContext) MarkClusterReadable() {
	c.clusterReadable = true
}

func (c *MemFSPath) HasChildren() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return len(c.children) != 0
}

func (c *MemFSPath) IsClusterReadable() bool {
	return c.context.clusterReadable
}

var _ HasClusterReadable = &MemFSPath{}

func NewMemFSPath(context *MemFSContext, location string) *MemFSPath {
	return context.root.Join(location).(*MemFSPath)
}

func (p *MemFSPath) Join(relativePath ...string) Path {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	joined := path.Join(relativePath...)
	tokens := strings.Split(joined, "/")
	current := p
	for _, token := range tokens {
		if current.children == nil {
			current.children = make(map[string]*MemFSPath)
		}
		child := current.children[token]
		if child == nil {
			child = &MemFSPath{
				context:  p.context,
				location: path.Join(current.location, token),
			}
			current.children[token] = child
		}
		current = child
		current.mutex.Lock()
		defer current.mutex.Unlock()
	}
	return current
}

func (p *MemFSPath) WriteFile(r io.ReadSeeker, acl ACL) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("error reading data: %v", err)
	}
	p.contents = data
	return nil
}

func (p *MemFSPath) CreateFile(data io.ReadSeeker, acl ACL) error {
	// Check if exists
	if p.contents != nil {
		return os.ErrExist
	}

	return p.WriteFile(data, acl)
}

// ReadFile implements Path::ReadFile
func (p *MemFSPath) ReadFile() ([]byte, error) {
	if p.contents == nil {
		return nil, os.ErrNotExist
	}
	// TODO: Copy?
	return p.contents, nil
}

// WriteTo implements io.WriterTo
func (p *MemFSPath) WriteTo(out io.Writer) (int64, error) {
	if p.contents == nil {
		return 0, os.ErrNotExist
	}
	n, err := out.Write(p.contents)
	return int64(n), err
}

func (p *MemFSPath) ReadDir() ([]Path, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var paths []Path
	for _, f := range p.children {
		paths = append(paths, f)
	}
	return paths, nil
}

func (p *MemFSPath) ReadTree() ([]Path, error) {
	var paths []Path
	p.readTree(&paths)
	return paths, nil
}

func (p *MemFSPath) readTree(dest *[]Path) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, f := range p.children {
		if !f.HasChildren() {
			*dest = append(*dest, f)
		}
		f.readTree(dest)
	}
}

func (p *MemFSPath) Base() string {
	return path.Base(p.location)
}

func (p *MemFSPath) Path() string {
	return "memfs://" + p.location
}

func (p *MemFSPath) String() string {
	return p.Path()
}

func (p *MemFSPath) Remove() error {
	p.contents = nil
	return nil
}

func (p *MemFSPath) RemoveAllVersions() error {
	return p.Remove()
}
