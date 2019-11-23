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

package main

import (
	"bytes"
	"io/ioutil"
	"path"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kops/pkg/apis/kops/v1alpha2"
	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/pkg/kopscodecs"
)

// TestConversionMinimal runs the test on a minimum configuration, similar to kops create cluster minimal.example.com --zones us-west-1a
func TestConversionMinimal(t *testing.T) {
	runTest(t, "minimal", "legacy-v1alpha2", "v1alpha2")
}

func runTest(t *testing.T, srcDir string, fromVersion string, toVersion string) {
	sourcePath := path.Join(srcDir, fromVersion+".yaml")
	sourceBytes, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("unexpected error reading sourcePath %q: %v", sourcePath, err)
	}

	expectedPath := path.Join(srcDir, toVersion+".yaml")
	expectedBytes, err := ioutil.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("unexpected error reading expectedPath %q: %v", expectedPath, err)
	}

	yaml, ok := runtime.SerializerInfoForMediaType(kopscodecs.Codecs.SupportedMediaTypes(), "application/yaml")
	if !ok {
		t.Fatalf("no YAML serializer registered")
	}
	var encoder runtime.Encoder

	switch toVersion {
	case "v1alpha2":
		encoder = kopscodecs.Codecs.EncoderForVersion(yaml.Serializer, v1alpha2.SchemeGroupVersion)

	default:
		t.Fatalf("unknown version %q", toVersion)
	}

	var actual []string

	for _, s := range strings.Split(string(sourceBytes), "\n---\n") {
		o, gvk, err := kopscodecs.Decode([]byte(s), nil)
		if err != nil {
			t.Fatalf("error parsing file %q: %v", sourcePath, err)
		}

		expectVersion := strings.TrimPrefix(fromVersion, "legacy-")
		if expectVersion == "v1alpha0" {
			// Our version before we had v1alpha1
			expectVersion = "v1alpha1"
		}
		if gvk.Version != expectVersion {
			t.Fatalf("unexpected version: %q vs %q", gvk.Version, expectVersion)
		}

		var b bytes.Buffer
		if err := encoder.Encode(o, &b); err != nil {
			t.Fatalf("error encoding object: %v", err)
		}

		actual = append(actual, b.String())
	}

	actualString := strings.TrimSpace(strings.Join(actual, "\n---\n\n"))
	expectedString := strings.TrimSpace(string(expectedBytes))

	if actualString != expectedString {
		diffString := diff.FormatDiff(expectedString, actualString)
		t.Logf("diff:\n%s\n", diffString)

		t.Fatalf("%s->%s converted output differed from expected", fromVersion, toVersion)
	}
}
