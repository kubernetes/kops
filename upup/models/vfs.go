package models

import (
	"k8s.io/kops/upup/pkg/fi/vfs"
	"path"
	"errors"
	"os"
	"strings"
)

var ReadOnlyError = errors.New("AssetPath is read-only")

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

func (p *AssetPath) WriteFile(data []byte) error {
	return ReadOnlyError
}

func (p *AssetPath) CreateFile(data []byte) error {
	return ReadOnlyError
}

func (p *AssetPath) ReadFile() ([]byte, error) {
	data, err := Asset(p.location)
	if err != nil {
		// Yuk
		if strings.Contains(err.Error(), "not found") {
			return nil, os.ErrNotExist
		}
	}
	return data, err
}

func (p *AssetPath) ReadDir() ([]vfs.Path, error) {
	files, err := AssetDir(p.location)
	if err != nil {
		// Yuk
		if strings.Contains(err.Error(), "not found") {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	var paths []vfs.Path
	for _, f := range files {
		paths = append(paths, NewAssetPath(path.Join(p.location, f)))
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
	files, err := AssetDir(base)
	if err != nil {
		// Yuk
		if strings.Contains(err.Error(), "not found") {
			return os.ErrNotExist
		}
		return err
	}
	for _, f := range files {
		p := path.Join(base, f)
		*dest = append(*dest, NewAssetPath(p))

		// We always assume a directory, but ignore if not found
		// This is because go-bindata doesn't support FileInfo on directories :-(
		{
			err = readTree(p, dest)
			if err != nil && !os.IsNotExist(err){
				return err
			}
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

