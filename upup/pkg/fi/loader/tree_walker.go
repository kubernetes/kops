/*
Copyright 2016 The Kubernetes Authors.

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
	"github.com/golang/glog"
	"k8s.io/kops/util/pkg/vfs"
	"k8s.io/apimachinery/pkg/util/sets"
	"os"
	"path"
	"strings"
)

type TreeWalker struct {
	Contexts       map[string]Handler
	Extensions     map[string]Handler
	DefaultHandler Handler
	Tags           sets.String
}

type TreeWalkItem struct {
	Context      string
	Name         string
	Path         vfs.Path
	RelativePath string
	Meta         string
	Tags         []string
}

func (i *TreeWalkItem) ReadString() (string, error) {
	b, err := i.ReadBytes()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (i *TreeWalkItem) ReadBytes() ([]byte, error) {
	b, err := i.Path.ReadFile()
	if err != nil {
		return nil, fmt.Errorf("error reading file %q: %v", i.Path, err)
	}
	return b, nil
}

type Handler func(item *TreeWalkItem) error

func IsTag(name string) bool {
	return len(name) != 0 && name[0] == '_'
}

func (t *TreeWalker) Walk(basedir vfs.Path) error {
	i := &TreeWalkItem{
		Context:      "",
		Path:         basedir,
		RelativePath: "",
		Tags:         nil,
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
			Tags:         parent.Tags,
		}

		glog.V(4).Infof("visit %q", f)

		hasMeta := false
		{
			metaPath := parent.Path.Join(fileName + ".meta")
			metaBytes, err := metaPath.ReadFile()
			if err != nil {
				if !os.IsNotExist(err) {
					return fmt.Errorf("error reading file %q: %v", metaPath, err)
				}
				metaBytes = nil
			}
			if metaBytes != nil {
				hasMeta = true
				i.Meta = string(metaBytes)
			}
		}

		if _, err := f.ReadDir(); err == nil {
			if IsTag(fileName) {
				// Only descend into the tag directory if we have the tag
				_, found := t.Tags[fileName]
				if !found {
					glog.V(2).Infof("Skipping directory %q as tag %q not present", f, fileName)
					continue
				} else {
					i.Tags = append(i.Tags, fileName)
					glog.V(2).Infof("Descending into directory, as tag is present: %q", f)
					err = t.walkDirectory(i)
				}
			} else if _, found := t.Contexts[fileName]; found {
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

			// So that we can manage directories, we do not ignore directories which have a .meta file
			if hasMeta {
				glog.V(4).Infof("Found .meta file for directory %q; will process", f)
			} else {
				continue
			}
		}

		if strings.HasSuffix(fileName, ".meta") {
			// We'll read it when we see the actual file
			// But check the actual file is there
			primaryPath := strings.TrimSuffix(f.Base(), ".meta")
			if _, err := parent.Path.Join(primaryPath).ReadFile(); os.IsNotExist(err) {
				return fmt.Errorf("found .meta file without corresponding file: %q", f)
			}

			continue
		}

		var handler Handler
		if i.Context != "" {
			handler = t.Contexts[i.Context]
		} else {
			// TODO: Just remove extensions.... we barely use them!
			// (or remove default handler and replace with lots of small files?)
			extension := path.Ext(fileName)
			handler = t.Extensions[extension]
			if handler == nil {
				handler = t.DefaultHandler
			}
		}

		err = handler(i)
		if err != nil {
			return fmt.Errorf("error handling file %q: %v", f, err)
		}
	}

	return nil
}
