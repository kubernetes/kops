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

package testutils

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/v1alpha2"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/pkg/testutils/golden"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/text"
)

type Model struct {
	Cluster        *kops.Cluster
	InstanceGroups []*kops.InstanceGroup

	// AdditionalObjects holds cluster-asssociated configuration objects, other than the Cluster and InstanceGroups.
	AdditionalObjects []*unstructured.Unstructured
}

// LoadModel loads a cluster and instancegroups from a cluster.yaml file found in basedir
func LoadModel(basedir string) (*Model, error) {
	clusterYamlPath := path.Join(basedir, "cluster.yaml")
	clusterYaml, err := os.ReadFile(clusterYamlPath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %q: %v", clusterYamlPath, err)
	}

	spec := &Model{}

	sections := text.SplitContentToSections(clusterYaml)
	for _, section := range sections {
		defaults := &schema.GroupVersionKind{
			Group:   v1alpha2.SchemeGroupVersion.Group,
			Version: v1alpha2.SchemeGroupVersion.Version,
		}
		o, gvk, err := kopscodecs.Decode(section, defaults)
		if err != nil {
			return nil, fmt.Errorf("error parsing file %v", err)
		}

		switch v := o.(type) {
		case *kops.Cluster:
			if spec.Cluster != nil {
				return nil, fmt.Errorf("found multiple clusters")
			}
			spec.Cluster = v
		case *kops.InstanceGroup:
			spec.InstanceGroups = append(spec.InstanceGroups, v)

		case *unstructured.Unstructured:
			spec.AdditionalObjects = append(spec.AdditionalObjects, v)

		default:
			return nil, fmt.Errorf("unhandled kind %T %q", o, gvk)
		}
	}

	return spec, nil
}

func ValidateTasks(t *testing.T, expectedFile string, context *fi.ModelBuilderContext) {
	var keys []string
	for key := range context.Tasks {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var yamls []string
	for _, key := range keys {
		task := context.Tasks[key]
		yaml, err := kops.ToRawYaml(task)
		if err != nil {
			t.Fatalf("error serializing task: %v", err)
		}
		yamls = append(yamls, strings.TrimSpace(string(yaml)))
	}

	actualTasksYaml := strings.Join(yamls, "\n---\n")
	actualTasksYaml = strings.TrimSpace(actualTasksYaml)

	golden.AssertMatchesFile(t, actualTasksYaml, expectedFile)

	// Asserts that FindTaskDependencies doesn't call klog.Fatalf()
	fi.FindTaskDependencies(context.Tasks)
}

// ValidateStaticFiles is used to validate generate StaticFiles.
func ValidateStaticFiles(t *testing.T, expectedDir string, assetBuilder *assets.AssetBuilder) {
	files, err := os.ReadDir(expectedDir)
	if err != nil {
		t.Fatalf("error reading directory %q: %v", expectedDir, err)
	}

	prefix := "static-"

	staticFiles := make(map[string]*assets.StaticFile)
	for _, staticFile := range assetBuilder.StaticFiles {
		k := filepath.Base(staticFile.Path)
		staticFiles[k] = staticFile
		expectedFile := filepath.Join(expectedDir, prefix+k)
		golden.AssertMatchesFile(t, staticFile.Content, expectedFile)
	}

	for _, file := range files {
		name := file.Name()
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		p := filepath.Join(expectedDir, name)
		key := strings.TrimPrefix(name, prefix)
		if _, found := staticFiles[key]; !found {
			t.Errorf("unexpected file with prefix %q: %q", prefix, p)
		}
	}
}

func ValidateCompletedCluster(t *testing.T, expectedFile string, cluster *kops.Cluster) {
	b, err := kops.ToRawYaml(cluster)
	if err != nil {
		t.Fatalf("error serializing cluster: %v", err)
	}

	yaml := strings.TrimSpace(string(b))

	golden.AssertMatchesFile(t, yaml, expectedFile)
}
