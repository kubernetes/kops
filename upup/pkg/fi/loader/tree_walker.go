package loader

import (
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type TreeWalker struct {
	Contexts       map[string]Handler
	Extensions     map[string]Handler
	DefaultHandler Handler
	Tags           map[string]struct{}
}

type TreeWalkItem struct {
	Context      string
	Name         string
	Path         string
	RelativePath string
	Meta         string
}

func (i *TreeWalkItem) ReadString() (string, error) {
	b, err := i.ReadBytes()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (i *TreeWalkItem) ReadBytes() ([]byte, error) {
	b, err := ioutil.ReadFile(i.Path)
	if err != nil {
		return nil, fmt.Errorf("error reading file %q: %v", i.Path, err)
	}
	return b, nil
}

type Handler func(item *TreeWalkItem) error

func IsTag(name string) bool {
	return len(name) != 0 && name[0] == '_'
}

func (t *TreeWalker) Walk(basedir string) error {
	i := &TreeWalkItem{
		Context:      "",
		Path:         basedir,
		RelativePath: "",
	}

	return t.walkDirectory(i)
}

func (t *TreeWalker) walkDirectory(parent *TreeWalkItem) error {
	files, err := ioutil.ReadDir(parent.Path)
	if err != nil {
		return fmt.Errorf("error reading directory %q: %v", parent.Path, err)
	}

	for _, f := range files {
		var err error

		fileName := f.Name()

		i := &TreeWalkItem{
			Context:      parent.Context,
			Path:         path.Join(parent.Path, fileName),
			RelativePath: path.Join(parent.RelativePath, fileName),
			Name:         fileName,
		}

		glog.V(4).Infof("visit %q", i.Path)

		if f.IsDir() {
			if IsTag(fileName) {
				// Only descend into the tag directory if we have the tag
				_, found := t.Tags[fileName]
				if !found {
					glog.V(2).Infof("Skipping directory as tag not present: %q", i.Path)
					continue
				} else {
					glog.V(2).Infof("Descending into directory, as tag is present: %q", i.Path)
					err = t.walkDirectory(i)
				}
			} else if _, found := t.Contexts[fileName]; found {
				// Entering a new context (mode of operation)
				if parent.Context != "" {
					return fmt.Errorf("found context %q inside context %q at %q", fileName, parent.Context, i.Path)
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

		if strings.HasSuffix(fileName, ".meta") {
			// We'll read it when we see the actual file
			// But check the actual file is there
			primaryPath := strings.TrimSuffix(i.Path, ".meta")
			if _, err := os.Stat(primaryPath); os.IsNotExist(err) {
				return fmt.Errorf("found .meta file without corresponding file: %q", i.Path)
			}

			continue
		}

		{
			metaPath := i.Path + ".meta"
			metaBytes, err := ioutil.ReadFile(metaPath)
			if err != nil {
				if !os.IsNotExist(err) {
					return fmt.Errorf("error reading file %q: %v", metaPath, err)
				}
				metaBytes = nil
			}
			if metaBytes != nil {
				i.Meta = string(metaBytes)
			}
		}

		var handler Handler
		if i.Context != "" {
			handler = t.Contexts[i.Context]
		} else {
			extension := path.Ext(fileName)
			handler = t.Extensions[extension]
			if handler == nil {
				handler = t.DefaultHandler
			}
		}

		err = handler(i)
		if err != nil {
			return fmt.Errorf("error handling file %q: %v", i.Path, err)
		}
	}

	return nil
}
