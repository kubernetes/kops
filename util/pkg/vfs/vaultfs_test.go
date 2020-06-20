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
	"math/rand"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	vault "github.com/hashicorp/vault/api"
)

func createClient(t *testing.T) *vault.Client {
	token := os.Getenv("VAULT_DEV_ROOT_TOKEN_ID")
	if token == "" {
		t.Skip("No vault dev token set. Skipping")
	}

	client, _ := newVaultClient("http://", "localhost", "8200")

	client.SetToken(token)
	return client
}

func Test_newVaultPath(t *testing.T) {
	client := createClient(t)
	vaultPath, err := newVaultPath(client, "http://", "/secret/foo/bar")

	if err != nil {
		t.Errorf("Failed to create vault path: %v", err)
	}
	actual := vaultPath.mountPoint
	expected := "secret"
	if actual != expected {
		t.Errorf("Expected mountPoint %v, got %v", expected, actual)
	}

	actual = vaultPath.path
	expected = "foo/bar"
	if actual != expected {
		t.Errorf("Expected path %v, got %v", expected, actual)
	}

	actual = vaultPath.metadataPath()
	expected = "secret/metadata/foo/bar"
	if actual != expected {
		t.Errorf("expected path %v, got %v", expected, actual)
	}

	actual = vaultPath.dataPath()
	expected = "secret/data/foo/bar"
	if actual != expected {
		t.Errorf("expected path %v, got %v", expected, actual)
	}

	actual = vaultPath.Path()
	expected = "vault://localhost:8200/secret/foo/bar?tls=false"
	if actual != expected {
		t.Errorf("expected path %v, got %v", expected, actual)
	}
}

func Test_newVaultPathHostOnly(t *testing.T) {
	_, err := newVaultPath(nil, "", "/")
	if err == nil {
		t.Error("Failed to return error on incorrect path")
	}
}

func Test_WriteFileReadFile(t *testing.T) {

	client := createClient(t)

	p, _ := newVaultPath(client, "http://", "/secret/createfiletest")

	secret := "my very special secret"
	err := p.WriteFile(strings.NewReader(secret), nil)
	if err != nil {
		t.Errorf("Failed to write file: %v", err)
	}

	fileBytes, err := p.ReadFile()
	if err != nil {
		t.Errorf("Failed to read file: %v", err)
	}
	if !bytes.Equal(fileBytes, []byte(secret)) {
		t.Errorf("Failed to read file. Got %v, expected %v", fileBytes, secret)
	}
}
func Test_WriteFileDeleteFile(t *testing.T) {

	client := createClient(t)

	p, _ := newVaultPath(client, "http://", "/secret/createfiletestdelete")

	secret := "my very special secret"
	err := p.WriteFile(strings.NewReader(secret), nil)
	if err != nil {
		t.Errorf("Failed to write file: %v", err)
	}

	err = p.Remove()
	if err != nil {
		t.Errorf("Failed to write file: %v", err)
	}

	_, err = p.ReadFile()
	if !os.IsNotExist(err) {
		t.Error("Failed to delete file")
	}
}

func Test_CreateFile(t *testing.T) {
	client := createClient(t)

	rand.Seed(time.Now().UnixNano())
	postfix := rand.Int()
	postfixs := strconv.Itoa(postfix)

	path := "/secret/createfiletest/" + postfixs

	p, _ := newVaultPath(client, "http://", path)

	secret := "my very special secret"
	err := p.CreateFile(strings.NewReader(secret), nil)
	if err != nil {
		t.Errorf("Failed to create file at %v: %v", path, err)
	}

	err = p.CreateFile(strings.NewReader(secret), nil)
	if err == nil {
		t.Errorf("Should have failed to create file at %v", path)
	}
}

func Test_ReadFile(t *testing.T) {
	client := createClient(t)

	p, _ := newVaultPath(client, "http://", "/secret/readfiletest")

	_, err := p.ReadFile()
	if !os.IsNotExist(err) {
		if err != nil {
			t.Errorf("Unexpected error reading file: %v", err)
		} else {
			t.Error("File should not exist")
		}
	}
}

func Test_Join(t *testing.T) {
	client := createClient(t)
	p, _ := newVaultPath(client, "http://", "/secret/joinfiletest")

	p2 := p.Join("another", "path")
	expected := "vault://localhost:8200/secret/joinfiletest/another/path?tls=false"
	if p2.Path() != expected {
		t.Errorf("Error joining path. Got: %v, Expected: %v", p2.Path(), expected)
	}
}

func Test_VaultReadDirList(t *testing.T) {

	tests := []struct {
		path     string
		subpaths []string
		expected []string
	}{
		{
			path: "/secret/readdirtest/",
			subpaths: []string{
				"subdir1/foo",
				"subdir2/foo",
				"test.data",
			},
			expected: []string{
				"subdir1",
				"subdir2",
				"test.data",
			},
		},
	}
	for _, test := range tests {
		client := createClient(t)

		vaultPath, _ := newVaultPath(client, "http://", test.path)
		// Create sub-paths
		for _, subpath := range test.subpaths {
			file := strings.NewReader("some data")
			vaultPath.Join(subpath).WriteFile(file, nil)
		}

		// Read dir
		paths, err := vaultPath.ReadDir()
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
			expected[i], _ = newVaultPath(client, "http://", test.path+test.expected[i])
		}
		if !reflect.DeepEqual(paths, expected) {
			t.Errorf("Expected sub-paths %v, got %v", expected, paths)
		}
	}

}

func Test_ReadDir(t *testing.T) {
	directory := "/secret/path"
	file := "somefile"
	client := createClient(t)
	directoryPath, _ := newVaultPath(client, "http://", directory)
	filePath := directoryPath.Join(file)
	filePath.WriteFile(strings.NewReader("foo"), nil)

	_, err := filePath.ReadDir()

	if err == nil {
		t.Error("File considered directory")
	}

	_, err = directoryPath.ReadDir()
	if err != nil {
		t.Error("Directory not considered directory")
	}

	nonExistingPath, _ := newVaultPath(client, "http://", "/secret/does/not/exist")
	_, err = nonExistingPath.ReadDir()
	if !os.IsNotExist(err) {
		t.Error("Found non-existing directory")
	}
}

func Test_IsDirectory(t *testing.T) {
	directory := "/secret/path"
	file := "somefile"
	client := createClient(t)
	directoryPath, _ := newVaultPath(client, "http://", directory)
	filePath := directoryPath.Join(file)
	filePath.WriteFile(strings.NewReader("foo"), nil)

	if IsDirectory(filePath) {
		t.Error("File considered directory")
	}

	if !IsDirectory(directoryPath) {
		t.Error("Directory not considered directory")
	}
}

func Test_VaultReadTree(t *testing.T) {
	tests := []struct {
		path     string
		subpaths []string
		expected []string
	}{
		{
			path: "/secret/dir/",
			subpaths: []string{
				"subdir/test1.data",
				"subdir2/test2.data",
			},
			expected: []string{
				"/secret/dir/subdir/test1.data",
				"/secret/dir/subdir2/test2.data",
			},
		},
	}
	for _, test := range tests {
		client := createClient(t)
		vaultPath, _ := newVaultPath(client, "http://", test.path)

		// Create sub-paths
		for _, subpath := range test.subpaths {
			vaultPath.Join(subpath).WriteFile(strings.NewReader("foo"), nil)
		}

		// Read dir tree
		paths, err := vaultPath.ReadTree()
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
			expected[i], _ = newVaultPath(client, "http://", test.expected[i])
		}
		if !reflect.DeepEqual(paths, expected) {
			t.Errorf("Expected tree paths %v, got %v", expected, paths)
		}
	}
}
