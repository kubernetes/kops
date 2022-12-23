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

package models

import (
	"context"
	"embed"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"

	"k8s.io/kops/util/pkg/vfs"
)

var ReadOnlyError = errors.New("AssetPath is read-only")

//go:embed cloudup
var content embed.FS

type AssetPath struct {
	location string
}

var _ vfs.Path = &AssetPath{}

func NewAssetPath(location string) *AssetPath {
	a := &AssetPath{
		location: location,
	}
	return a
}

func (p *AssetPath) Join(relativePath ...string) vfs.Path {
	args := []string{p.location}
	args = append(args, relativePath...)
	joined := path.Join(args...)
	return &AssetPath{location: joined}
}

func (p *AssetPath) WriteFile(ctx context.Context, data io.ReadSeeker, acl vfs.ACL) error {
	return ReadOnlyError
}

func (p *AssetPath) CreateFile(ctx context.Context, data io.ReadSeeker, acl vfs.ACL) error {
	return ReadOnlyError
}

// WriteTo implements io.WriterTo
func (p *AssetPath) WriteTo(out io.Writer) (int64, error) {
	data, err := p.ReadFile()
	if err != nil {
		return 0, err
	}
	n, err := out.Write(data)
	return int64(n), err
}

// ReadFile implements Path::ReadFile
func (p *AssetPath) ReadFile() ([]byte, error) {
	data, err := content.ReadFile(p.location)
	if _, ok := err.(*fs.PathError); ok {
		return nil, os.ErrNotExist
	}
	return data, err
}

func (p *AssetPath) ReadDir() ([]vfs.Path, error) {
	files, err := content.ReadDir(p.location)
	if err != nil {
		if _, ok := err.(*fs.PathError); ok {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	var paths []vfs.Path
	for _, f := range files {
		paths = append(paths, NewAssetPath(path.Join(p.location, f.Name())))
	}
	return paths, nil
}

func (p *AssetPath) ReadTree() ([]vfs.Path, error) {
	var paths []vfs.Path
	err := readTree(p.location, &paths)
	if err != nil {
		return nil, err
	}
	return paths, nil
}

func readTree(base string, dest *[]vfs.Path) error {
	files, err := content.ReadDir(base)
	if err != nil {
		if _, ok := err.(*fs.PathError); ok {
			return os.ErrNotExist
		}
		return err
	}
	for _, f := range files {
		p := path.Join(base, f.Name())
		if f.IsDir() {
			childFiles, err := NewAssetPath(p).ReadTree()
			if err != nil {
				return err
			}
			*dest = append(*dest, childFiles...)
		} else {
			*dest = append(*dest, NewAssetPath(p))
		}
	}
	return nil
}

func (p *AssetPath) Base() string {
	return path.Base(p.location)
}

func (p *AssetPath) Path() string {
	return p.location
}

func (p *AssetPath) String() string {
	return p.Path()
}

func (p *AssetPath) Remove() error {
	return ReadOnlyError
}

func (p *AssetPath) RemoveAllVersions() error {
	return p.Remove()
}
