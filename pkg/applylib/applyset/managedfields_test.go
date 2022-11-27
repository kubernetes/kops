/*
Copyright 2022 The Kubernetes Authors.

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

package applyset

import (
	"os"
	"path/filepath"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	"k8s.io/kops/pkg/testutils/golden"
)

func TestManagedFieldsMigrator(t *testing.T) {
	testdataBaseDir := "testdata/managedfields"
	entries, err := os.ReadDir(testdataBaseDir)
	if err != nil {
		t.Fatalf("failed to read %q: %v", testdataBaseDir, err)
	}
	for _, entry := range entries {
		testdataDir := filepath.Join(testdataBaseDir, entry.Name())
		t.Run(entry.Name(), func(t *testing.T) {
			p := filepath.Join(testdataDir, "object.yaml")
			b, err := os.ReadFile(p)
			if err != nil {
				t.Fatalf("failed to read %q: %v", p, err)
			}
			obj := &unstructured.Unstructured{}
			if err := yaml.Unmarshal(b, &obj.Object); err != nil {
				t.Fatalf("failed to parse %q: %v", p, err)
			}
			migrator := &ManagedFieldsMigrator{NewManager: "new-manager"}
			patch, err := migrator.createManagedFieldPatch(obj)
			if err != nil {
				t.Fatalf("error from createManagedFieldPatch: %v", err)
			}
			patchString := ""
			if patch != nil {
				patchString = string(patch)
			}
			expectedPatch := filepath.Join(testdataDir, "patch.json")
			golden.AssertMatchesFile(t, patchString, expectedPatch)
		})
	}
}
