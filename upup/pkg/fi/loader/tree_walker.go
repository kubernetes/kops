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

package loader

import (
	"fmt"
	"path"

	"k8s.io/klog/v2"
	"k8s.io/kops/util/pkg/vfs"
)

type TreeWalker struct {
	Contexts map[string]Handler
}

type TreeWalkItem struct {
	Context      string
	Name         string
	Path         vfs.Path
	RelativePath string
}

func (i *TreeWalkItem) ReadBytes() ([]byte, error) {
	b, err := i.Path.ReadFile()
	if err != nil {
		return nil, fmt.Errorf("error reading file %q: %v", i.Path, err)
	}
	return b, nil
}

type Handler func(item *TreeWalkItem) error

func (t *TreeWalker) Walk(basedir vfs.Path) error {
	i := &TreeWalkItem{
		Context:      "",
		Path:         basedir,
		RelativePath: "",
	}

	return t.walkDirectory(i)
}

func (t *TreeWalker) walkDirectory(parent *TreeWalkItem) error {
	files, err := parent.Path.ReadDir()
	if err != nil {
		return fmt.Errorf("error reading directory %q: %v", parent.Path, err)
	}

	for _, f := range files {
		var err error

		fileName := f.Base()

		i := &TreeWalkItem{
			Context:      parent.Context,
			Path:         f,
			RelativePath: path.Join(parent.RelativePath, fileName),
			Name:         fileName,
		}

		klog.V(4).Infof("visit %q", f)

		if _, err := f.ReadDir(); err == nil {
			if _, found := t.Contexts[fileName]; found {
				// Entering a new context (mode of operation)
				if parent.Context != "" {
					return fmt.Errorf("found context %q inside context %q at %q", fileName, parent.Context, f)
				}
				i.Context = fileName
				i.RelativePath = ""
				err = t.walkDirectory(i)
			} else {
				// Simple directory for organization / structure
				err = t.walkDirectory(i)
			}
			if err != nil {
				return err
			}

			continue
		}

		handler := t.Contexts[i.Context]
		err = handler(i)
		if err != nil {
			return fmt.Errorf("error handling file %q: %v", f, err)
		}
	}

	return nil
}
