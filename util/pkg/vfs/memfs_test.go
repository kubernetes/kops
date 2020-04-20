/*
Copyright 2020 The Kubernetes Authors.

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
	"bytes"
	"os"
	"reflect"
	"sort"
	"testing"
)

func TestMemFsCreateFile(t *testing.T) {
	tests := []struct {
		path string
		data []byte
	}{
		{
			path: "/root/subdir/test1.data",
			data: []byte("test data\nline 1\r\nline 2"),
		},
	}
	for _, test := range tests {
		memfspath := NewMemFSPath(NewMemFSContext(), test.path)
		// Create file
		err := memfspath.CreateFile(bytes.NewReader(test.data), nil)
		if err != nil {
			t.Errorf("Failed writing path %s, error: %v", test.path, err)
			continue
		}

		// Create file again should result in error
		err = memfspath.CreateFile(bytes.NewReader([]byte("data")), nil)
		if err != os.ErrExist {
			t.Errorf("Expected to get os.ErrExist, got: %v", err)
		}

		// Check file content
		data, err := memfspath.ReadFile()
		if err != nil {
			t.Errorf("Failed reading path %s, error: %v", test.path, err)
			continue
		}
		if !bytes.Equal(data, test.data) {
			t.Errorf("Expected path content %v, got %v", test.data, data)
		}
	}
}

func TestMemFsReadDir(t *testing.T) {
	tests := []struct {
		path     string
		subpaths []string
		expected []string
	}{
		{
			path: "/root/",
			subpaths: []string{
				"subdir/",
				"subdir2/",
				"subdir2/test.data",
			},
			expected: []string{
				"/root/subdir/",
				"/root/subdir2/",
			},
		},
	}
	for _, test := range tests {
		context := NewMemFSContext()
		memfspath := NewMemFSPath(context, test.path)

		// Create sub-paths
		for _, subpath := range test.subpaths {
			memfspath.Join(subpath)
		}

		// Read dir
		paths, err := memfspath.ReadDir()
		if err != nil {
			t.Errorf("Failed reading dir %s, error: %v", test.path, err)
			continue
		}

		// There is no consistent alphabetical order in the result, so we sort it
		sort.Slice(paths, func(i, j int) bool {
			return paths[i].Path() < paths[j].Path()
		})
		// Expected sub-paths
		count := len(test.expected)
		expected := make([]Path, count)
		for i := 0; i < count; i++ {
			expected[i] = NewMemFSPath(context, test.expected[i])
		}
		if !reflect.DeepEqual(paths, expected) {
			t.Errorf("Expected sub-paths %v, got %v", expected, paths)
		}
	}
}

func TestMemFsReadTree(t *testing.T) {
	tests := []struct {
		path     string
		subpaths []string
		expected []string
	}{
		{
			path: "/root/dir/",
			subpaths: []string{
				"subdir/",
				"subdir/test1.data",
				"subdir2/",
				"subdir2/test2.data",
			},
			expected: []string{
				"/root/dir/subdir/test1.data",
				"/root/dir/subdir2/test2.data",
			},
		},
	}
	for _, test := range tests {
		context := NewMemFSContext()
		memfspath := NewMemFSPath(context, test.path)

		// Create sub-paths
		for _, subpath := range test.subpaths {
			memfspath.Join(subpath)
		}

		// Read dir tree
		paths, err := memfspath.ReadTree()
		if err != nil {
			t.Errorf("Failed reading dir tree %s, error: %v", test.path, err)
			continue
		}

		// There is no consistent alphabetical order in the result, so we sort it
		sort.Slice(paths, func(i, j int) bool {
			return paths[i].Path() < paths[j].Path()
		})
		// Expected sub-paths
		count := len(test.expected)
		expected := make([]Path, count)
		for i := 0; i < count; i++ {
			expected[i] = NewMemFSPath(context, test.expected[i])
		}
		if !reflect.DeepEqual(paths, expected) {
			t.Errorf("Expected tree paths %v, got %v", expected, paths)
		}
	}
}
