/*
Copyright 2016 The Kubernetes Authors.

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
	"strconv"
	"strings"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/protokube/pkg/protokube"

	_ "k8s.io/client-go/pkg/api/install"
)

func TestBuildEtcdManifest(t *testing.T) {
	runTest(t, "main")
}

func runTest(t *testing.T, srcDir string) {
	sourcePath := path.Join(srcDir, "cluster.yaml")
	sourceBytes, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("unexpected error reading sourcePath %q: %v", sourcePath, err)
	}

	expectedPath := path.Join(srcDir, "manifest.yaml")
	expectedBytes, err := ioutil.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("unexpected error reading expectedPath %q: %v", expectedPath, err)
	}

	cluster := &protokube.EtcdCluster{}
	err = kops.ParseRawYaml(sourceBytes, cluster)
	if err != nil {
		t.Fatalf("error parsing options yaml: %v", err)
	}

	cluster.Me = &protokube.EtcdNode{
		Name:         "node0",
		InternalName: "node0" + ".internal",
	}

	for i := 0; i <= 2; i++ {
		node := &protokube.EtcdNode{
			Name:         "node" + strconv.Itoa(i),
			InternalName: "node" + strconv.Itoa(i) + ".internal",
		}
		cluster.Nodes = append(cluster.Nodes, node)
	}

	pod := protokube.BuildEtcdManifest(cluster)
	actual, err := protokube.ToVersionedYaml(pod)
	if err != nil {
		t.Fatalf("error marshalling to yaml: %v", err)
	}

	actualString := strings.TrimSpace(string(actual))
	expectedString := strings.TrimSpace(string(expectedBytes))

	if actualString != expectedString {
		diffString := diff.FormatDiff(expectedString, actualString)
		t.Logf("diff:\n%s\n", diffString)

		t.Fatalf("manifest differed from expected")
	}
}
