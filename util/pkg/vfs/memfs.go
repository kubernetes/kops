package vfs

import (
	"os"
	"path"
	"strings"
)

type MemFSPath struct {
	context  *MemFSContext
	location string

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

func (c *MemFSPath) IsClusterReadable() bool {
	return c.context.clusterReadable
}

var _ HasClusterReadable = &MemFSPath{}

func NewMemFSPath(context *MemFSContext, location string) *MemFSPath {
	return context.root.Join(location).(*MemFSPath)
}

func (p *MemFSPath) Join(relativePath ...string) Path {
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
	}
	return current
}

func (p *MemFSPath) WriteFile(data []byte) error {
	p.contents = data
	return nil
}

func (p *MemFSPath) CreateFile(data []byte) error {
	// Check if exists
	if p.contents != nil {
		return os.ErrExist
	}

	return p.WriteFile(data)
}

func (p *MemFSPath) ReadFile() ([]byte, error) {
	if p.contents == nil {
		return nil, os.ErrNotExist
	}
	// TODO: Copy?
	return p.contents, nil
}

func (p *MemFSPath) ReadDir() ([]Path, error) {
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
	for _, f := range p.children {
		*dest = append(*dest, f)
		f.readTree(dest)
	}
}

func (p *MemFSPath) Base() string {
	return path.Base(p.location)
}

func (p *MemFSPath) Path() string {
	return p.location
}

func (p *MemFSPath) String() string {
	return p.Path()
}

func (p *MemFSPath) Remove() error {
	p.contents = nil
	return nil
}
