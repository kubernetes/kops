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

package nodetasks

import (
	"bytes"
	"compress/gzip"
	"io"
	"reflect"
	"testing"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/hashing"
)

func TestLoadImageTask_Deps(t *testing.T) {
	l := &LoadImageTask{}

	tasks := make(map[string]fi.NodeupTask)
	tasks["LoadImageTask1"] = &LoadImageTask{}
	tasks["FileTask1"] = &File{}
	tasks["ServiceDocker"] = &Service{Name: "docker.service"}
	tasks["Service2"] = &Service{Name: "two.service"}

	deps := l.GetDependencies(tasks)
	expected := []fi.NodeupTask{tasks["ServiceDocker"]}
	if !reflect.DeepEqual(expected, deps) {
		t.Fatalf("unexpected deps.  expected=%v, actual=%v", expected, deps)
	}
}

func TestImageImportReaderUngzipsAndHashesDownload(t *testing.T) {
	image := []byte("container image tar")
	var compressed bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressed)
	if _, err := gzipWriter.Write(image); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	expectedHash, err := hashing.HashAlgorithmSHA256.Hash(bytes.NewReader(compressed.Bytes()))
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	reader, verifyHash, closer, err := imageImportReader(bytes.NewReader(compressed.Bytes()), expectedHash)
	if err != nil {
		t.Fatalf("imageImportReader() error = %v", err)
	}

	actual, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if closer != nil {
		if err := closer.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}
	if !bytes.Equal(actual, image) {
		t.Fatalf("imageImportReader() body = %q, expected %q", actual, image)
	}
	if err := verifyHash(); err != nil {
		t.Fatalf("verifyHash() error = %v", err)
	}
}

func TestContainerImageImportArgsDoesNotUnpack(t *testing.T) {
	args := containerImageImportArgs()
	expected := []string{"ctr", "--namespace", "k8s.io", "images", "import", "--no-unpack", "-"}
	if !reflect.DeepEqual(args, expected) {
		t.Fatalf("containerImageImportArgs() = %v, expected %v", args, expected)
	}
}
