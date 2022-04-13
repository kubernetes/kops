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
	"os"
	"path"
	"strings"
	"sync"

	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

type MemFSPath struct {
	context  *MemFSContext
	location string
	acl      ACL

	mutex    sync.Mutex
	contents []byte
	children map[string]*MemFSPath
}

var (
	_ Path          = &MemFSPath{}
	_ TerraformPath = &MemFSPath{}
)

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
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("error reading data: %v", err)
	}
	p.contents = data
	p.acl = acl
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

func (p *MemFSPath) Location() string {
	return p.location
}

func (p *MemFSPath) IsPublic() (bool, error) {
	if p.acl == nil {
		return false, nil
	}
	s3Acl, ok := p.acl.(*S3Acl)
	if !ok {
		return false, fmt.Errorf("expected acl to be S3Acl, was %T", p.acl)
	}
	isPublic := false
	if s3Acl.RequestACL != nil {
		isPublic = *s3Acl.RequestACL == "public-read"
	}
	return isPublic, nil
}

// Terraform support for integration tests.

func (p *MemFSPath) TerraformProvider() (*TerraformProvider, error) {
	return &TerraformProvider{
		Name: "aws",
		Arguments: map[string]string{
			"region": "us-test-1",
		},
	}, nil
}

type terraformMemFSFile struct {
	Bucket   string                   `json:"bucket" cty:"bucket"`
	Key      string                   `json:"key" cty:"key"`
	Content  *terraformWriter.Literal `json:"content,omitempty" cty:"content"`
	Acl      *string                  `json:"acl,omitempty" cty:"acl"`
	SSE      string                   `json:"server_side_encryption,omitempty" cty:"server_side_encryption"`
	Provider *terraformWriter.Literal `json:"provider,omitempty" cty:"provider"`
}

func (p *MemFSPath) RenderTerraform(w *terraformWriter.TerraformWriter, name string, data io.Reader, acl ACL) error {
	bytes, err := io.ReadAll(data)
	if err != nil {
		return fmt.Errorf("reading data: %v", err)
	}

	content, err := w.AddFileBytes("aws_s3_object", name, "content", bytes, false)
	if err != nil {
		return fmt.Errorf("rendering S3 file: %v", err)
	}

	var requestAcl *string
	if acl != nil {
		s3Acl, ok := acl.(*S3Acl)
		if !ok {
			return fmt.Errorf("write to %s with ACL of unexpected type %T", p, acl)
		}
		requestAcl = s3Acl.RequestACL
	}

	tf := &terraformMemFSFile{
		Bucket:   "testingBucket",
		Key:      p.location,
		Content:  content,
		SSE:      "AES256",
		Acl:      requestAcl,
		Provider: terraformWriter.LiteralTokens("aws", "files"),
	}
	return w.RenderResource("aws_s3_object", name, tf)
}
