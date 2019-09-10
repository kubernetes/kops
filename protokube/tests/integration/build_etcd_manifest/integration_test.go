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
	"io/ioutil"
	"path"
	"strings"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/pkg/k8scodecs"
	"k8s.io/kops/protokube/pkg/protokube"
)

func TestBuildEtcdManifest(t *testing.T) {
	cs := []struct {
		TestFile string
	}{
		{TestFile: "non_tls.yaml"},
		{TestFile: "tls.yaml"},
		{TestFile: "etcd_env_vars.yaml"},
	}
	for i, x := range cs {
		cluster, expected := loadTestIntegration(t, path.Join("main", x.TestFile))
		definition := protokube.BuildEtcdManifest(cluster)
		generated, err := k8scodecs.ToVersionedYaml(definition)
		if err != nil {
			t.Errorf("case %d, unable to convert to yaml, error: %v", i, err)
			continue
		}
		rendered := strings.TrimSpace(string(generated))
		expected = strings.TrimSpace(expected)

		if rendered != expected {
			diffString := diff.FormatDiff(expected, string(rendered))
			t.Logf("diff:\n%s\n", diffString)
			t.Errorf("case %d, failed, manifest differed from expected", i)
		}
	}
}

// loadTestIntegration is responsible for loading the integration files
func loadTestIntegration(t *testing.T, path string) (*protokube.EtcdCluster, string) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("unable to read in the integretion file: %s, error: %v", path, err)
	}
	documents := strings.Split(string(content), "---")
	if len(documents) != 2 {
		t.Fatalf("unable to find both documents in the integration file: %s, error %v:", path, err)
	}
	// read the specification into a etcd spec
	cluster := &protokube.EtcdCluster{}
	err = kops.ParseRawYaml([]byte(documents[0]), cluster)
	if err != nil {
		t.Fatalf("error parsing etcd specification in file: %s, error: %v", path, err)
	}

	return cluster, documents[1]
}
