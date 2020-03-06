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
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
)

var TempDir, _ = ioutil.TempDir("", "test")

func TestCreateFile(t *testing.T) {
	tests := []struct {
		path       string
		fileExists bool
		data       []byte
	}{
		{
			path: path.Join(TempDir, "SubDir", "test1.tmp"),
			data: []byte("test data\nline 1\r\nline 2"),
		},
	}
	defer os.Remove(TempDir)
	for _, test := range tests {
		fspath := &FSPath{test.path}
		// Create file
		err := fspath.CreateFile(bytes.NewReader(test.data), nil)
		if err != nil {
			t.Errorf("Error writing file %s, error: %v", test.path, err)
		}

		// Create file again should result in error
		err = fspath.CreateFile(bytes.NewReader([]byte("data")), nil)
		if err != os.ErrExist {
			t.Errorf("Expected to get os.ErrExist, got: %v", err)
		}

		// Check file content
		data, err := fspath.ReadFile()
		if err != nil {
			t.Errorf("Error reading file %s, error: %v", test.path, err)
		}
		if !reflect.DeepEqual(data, test.data) {
			t.Errorf("Expected file content %v, got %v", data, test.data)
		}
	}
}
