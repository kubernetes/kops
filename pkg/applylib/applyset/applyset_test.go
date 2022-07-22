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
	"path/filepath"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kops/pkg/applylib/mocks"
	"k8s.io/kops/pkg/testutils/golden"
	"sigs.k8s.io/yaml"
)

func TestApplySet(t *testing.T) {
	h := mocks.NewHarness(t)

	existing := ``

	apply := `
apiVersion: v1
kind: Namespace
metadata:
  name: test-applyset

---

apiVersion: v1
kind: ConfigMap
metadata:
  name: foo
  namespace: test-applyset
data:
  foo: bar
`

	h.WithObjects(h.ParseObjects(existing)...)

	applyObjects := h.ParseObjects(apply)

	patchOptions := metav1.PatchOptions{
		FieldManager: "kops",
	}

	force := true
	patchOptions.Force = &force

	s, err := New(Options{
		RESTMapper:   h.RESTMapper(),
		Client:       h.DynamicClient(),
		PatchOptions: patchOptions,
	})
	if err != nil {
		h.Fatalf("error building applyset object: %v", err)
	}

	var applyableObjects []ApplyableObject
	for _, object := range applyObjects {
		applyableObjects = append(applyableObjects, object)
	}
	if err := s.SetDesiredObjects(applyableObjects); err != nil {
		h.Fatalf("failed to set desired objects: %v", err)
	}

	results, err := s.ApplyOnce(h.Ctx)
	if err != nil {
		h.Fatalf("failed to apply objects: %v", err)
	}

	// TODO: Implement pruning

	if !results.AllApplied() {
		h.Fatalf("not all objects were applied")
	}

	// TODO: Check object health status
	if !results.AllHealthy() {
		h.Fatalf("not all objects were healthy")
	}

	var actual []string

	for _, object := range applyObjects {
		id := types.NamespacedName{
			Namespace: object.GetNamespace(),
			Name:      object.GetName(),
		}

		u := &unstructured.Unstructured{}
		u.SetAPIVersion(object.GetAPIVersion())
		u.SetKind(object.GetKind())

		if err := h.Client().Get(h.Ctx, id, u); err != nil {
			h.Fatalf("failed to get object %v: %v", id, err)
		}

		metadata := u.Object["metadata"].(map[string]interface{})
		delete(metadata, "creationTimestamp")
		delete(metadata, "managedFields")
		delete(metadata, "resourceVersion")
		delete(metadata, "uid")

		y, err := yaml.Marshal(u)
		if err != nil {
			h.Fatalf("failed to marshal object %v: %v", id, err)
		}
		actual = append(actual, string(y))
	}
	testDir := filepath.Join("testdata", strings.ToLower(t.Name()))
	golden.AssertMatchesFile(t, strings.Join(actual, "\n---\n"), filepath.Join(testDir, "expected.yaml"))
}
