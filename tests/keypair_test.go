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
package tests

import (
	"math/big"
	"reflect"
	"sort"
	"testing"
	"time"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kops/util/pkg/vfs"
)

type MockTarget struct {
}

func (t *MockTarget) Finish(taskMap map[string]fi.Task) error {
	return nil
}

func (t *MockTarget) ProcessDeletions() bool {
	return false
}

var _ fi.Target = &MockTarget{}

// Verifies that we regenerate keyset.yaml if they are deleted, which covers the upgrade scenario from kops 1.8 -> kops 1.9
func TestKeypairUpgrade(t *testing.T) {
	lifecycle := fi.LifecycleSync

	runTasksOptions := fi.RunTasksOptions{}
	runTasksOptions.MaxTaskDuration = 2 * time.Second

	target := &MockTarget{}

	cluster := &kops.Cluster{}
	vfs.Context.ResetMemfsContext(true)

	basedir, err := vfs.Context.BuildVfsPath("memfs://keystore")
	if err != nil {
		t.Fatalf("error building vfs path: %v", err)
	}

	keystore := fi.NewVFSCAStore(cluster, basedir, true)

	// Generate predictable sequence numbers for testing
	var n int64
	keystore.SerialGenerator = func() *big.Int {
		n++
		return big.NewInt(n)
	}

	// We define a function so we can rebuild the tasks, because we modify in-place when running
	buildTasks := func() map[string]fi.Task {
		format := string(fi.KeysetFormatV1Alpha2)

		ca := &fitasks.Keypair{
			Name:      fi.String(fi.CertificateId_CA),
			Lifecycle: &lifecycle,
			Subject:   "cn=kubernetes",
			Type:      "ca",
			Format:    format,
		}

		kubelet := &fitasks.Keypair{
			Name:      fi.String("kubelet"),
			Lifecycle: &lifecycle,
			Subject:   "o=nodes,cn=kubelet",
			Type:      "client",
			Signer:    ca,
			Format:    format,
		}

		tasks := make(map[string]fi.Task)
		tasks["ca"] = ca
		tasks["kubelet"] = kubelet
		return tasks
	}

	t.Logf("Building some keypairs")
	{
		allTasks := buildTasks()

		context, err := fi.NewContext(target, nil, nil, keystore, nil, nil, true, allTasks)
		if err != nil {
			t.Fatalf("error building context: %v", err)
		}

		if err := context.RunTasks(runTasksOptions); err != nil {
			t.Fatalf("unexpected error during Run: %v", err)
		}
	}

	// Check that the expected files were generated
	expected := []string{
		"memfs://keystore/issued/ca/1.crt",
		"memfs://keystore/issued/ca/keyset.yaml",
		"memfs://keystore/issued/kubelet/2.crt",
		"memfs://keystore/issued/kubelet/keyset.yaml",
		"memfs://keystore/private/ca/1.key",
		"memfs://keystore/private/ca/keyset.yaml",
		"memfs://keystore/private/kubelet/2.key",
		"memfs://keystore/private/kubelet/keyset.yaml",
	}
	checkPaths(t, basedir, expected)

	// Save the contents of those files
	contents := make(map[string]string)
	for _, k := range expected {
		p, err := vfs.Context.BuildVfsPath(k)
		if err != nil {
			t.Fatalf("error building vfs path: %v", err)
		}
		b, err := p.ReadFile()
		if err != nil {
			t.Fatalf("error reading vfs path: %v", err)
		}
		contents[k] = string(b)
	}

	t.Logf("verifying that rerunning tasks does not change keys")
	{
		allTasks := buildTasks()

		context, err := fi.NewContext(target, nil, nil, keystore, nil, nil, true, allTasks)
		if err != nil {
			t.Fatalf("error building context: %v", err)
		}

		if err := context.RunTasks(runTasksOptions); err != nil {
			t.Fatalf("unexpected error during Run: %v", err)
		}
	}
	checkContents(t, basedir, contents)

	t.Logf("deleting keyset.yaml files and verifying they are recreated")
	FailOnError(t, basedir.Join("issued/ca/keyset.yaml").Remove())
	FailOnError(t, basedir.Join("issued/kubelet/keyset.yaml").Remove())
	FailOnError(t, basedir.Join("private/ca/keyset.yaml").Remove())
	FailOnError(t, basedir.Join("private/kubelet/keyset.yaml").Remove())

	{
		allTasks := buildTasks()

		context, err := fi.NewContext(target, nil, nil, keystore, nil, nil, true, allTasks)
		if err != nil {
			t.Fatalf("error building context: %v", err)
		}

		if err := context.RunTasks(runTasksOptions); err != nil {
			t.Fatalf("unexpected error during Run: %v", err)
		}
	}
	checkContents(t, basedir, contents)
}

// FailOnError calls t.Fatalf if err != nil
func FailOnError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// checkPaths verifies that the path names in the tree rooted at basedir are exactly as expected
// Unlike checkContents, it only verifies the names, not the contents
func checkPaths(t *testing.T, basedir vfs.Path, expected []string) {
	paths, err := basedir.ReadTree()
	if err != nil {
		t.Errorf("ReadTree failed: %v", err)
	}

	var actual []string
	for _, p := range paths {
		actual = append(actual, p.Path())
	}
	sort.Strings(actual)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("unexpected paths: %v", actual)
	}
}

// checkPaths verifies that the files and their contents in the tree rooted at basedir are exactly as expected
func checkContents(t *testing.T, basedir vfs.Path, expected map[string]string) {
	paths, err := basedir.ReadTree()
	if err != nil {
		t.Errorf("ReadTree failed: %v", err)
	}

	actual := make(map[string]string)
	for _, p := range paths {
		b, err := p.ReadFile()
		if err != nil {
			t.Fatalf("error reading vfs path %q: %v", p, err)
		}
		actual[p.Path()] = string(b)
	}

	var actualKeys []string
	for k := range actual {
		actualKeys = append(actualKeys, k)
	}
	sort.Strings(actualKeys)
	var expectedKeys []string
	for k := range expected {
		expectedKeys = append(expectedKeys, k)
	}
	sort.Strings(expectedKeys)

	if !reflect.DeepEqual(actualKeys, expectedKeys) {
		t.Fatalf("unexpected paths: %v", actualKeys)
	}
	if !reflect.DeepEqual(actual, expected) {
		for k := range actual {
			if actual[k] != expected[k] {
				t.Errorf("mismatch on key %q", k)
				t.Errorf("diff: %s", diff.FormatDiff(actual[k], expected[k]))
			}
		}
	}
}
