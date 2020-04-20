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

package fi

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestWriteFile(t *testing.T) {
	var TempDir, _ = ioutil.TempDir("", "fitest")
	defer os.Remove(TempDir)
	tests := []struct {
		path     string
		data     []byte
		fileMode os.FileMode
		dirMode  os.FileMode
	}{
		{
			path:     path.Join(TempDir, "SubDir", "test1.tmp"),
			data:     []byte("test data\nline 1\r\nline 2"),
			fileMode: 0644,
			dirMode:  0755,
		},
	}
	for _, test := range tests {
		err := WriteFile(test.path, NewBytesResource(test.data), test.fileMode, test.dirMode)
		if err != nil {
			t.Errorf("Error writing file {%s}, error: {%v}", test.path, err)
			continue
		}

		// Check file content
		data, err := ioutil.ReadFile(test.path)
		if err != nil {
			t.Errorf("Error reading file {%s}, error: {%v}", test.path, err)
			continue
		}
		if !bytes.Equal(data, test.data) {
			t.Errorf("Expected file content {%v}, got {%v}", test.data, data)
			continue
		}

		// Check file mode
		stat, err := os.Lstat(test.path)
		if err != nil {
			t.Errorf("Error getting file mode of {%s}, error: {%v}", test.path, err)
			continue
		}
		fileMode := stat.Mode() & os.ModePerm
		if fileMode != test.fileMode {
			t.Errorf("Expected file mode {%v}, got {%v}", test.fileMode, fileMode)
			continue
		}

		// Check dir mode
		dirPath := path.Dir(test.path)
		stat, err = os.Lstat(dirPath)
		if err != nil {
			t.Errorf("Error getting dir mode of {%s}, error: {%v}", dirPath, err)
			continue
		}
		dirMode := stat.Mode() & os.ModePerm
		if dirMode != test.dirMode {
			t.Errorf("Expected dir mode {%v}, got {%v}", test.dirMode, dirMode)
			continue
		}
	}
}
