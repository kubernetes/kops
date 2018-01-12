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

package cloudup

import (
	"testing"

	"io/ioutil"
	"path"
	"strings"

	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"

	// Register our APIs
	_ "k8s.io/kops/pkg/apis/kops/install"
)

func TestBootstrapChannelBuilder_BuildTasks(t *testing.T) {
	runChannelBuilderTest(t, "simple")
	runChannelBuilderTest(t, "kopeio-vxlan")
}

func runChannelBuilderTest(t *testing.T, key string) {
	basedir := path.Join("tests/bootstrapchannelbuilder/", key)

	clusterYamlPath := path.Join(basedir, "cluster.yaml")
	clusterYaml, err := ioutil.ReadFile(clusterYamlPath)
	if err != nil {
		t.Fatalf("error reading cluster yaml file %q: %v", clusterYamlPath, err)
	}
	obj, _, err := api.ParseVersionedYaml(clusterYaml)
	if err != nil {
		t.Fatalf("error parsing cluster yaml %q: %v", clusterYamlPath, err)
	}
	cluster := obj.(*api.Cluster)
	bcb := BootstrapChannelBuilder{cluster: cluster}

	context := &fi.ModelBuilderContext{
		Tasks: make(map[string]fi.Task),
	}
	err = bcb.Build(context)
	if err != nil {
		t.Fatalf("error from BootstrapChannelBuilder Build: %v", err)
	}

	name := cluster.ObjectMeta.Name + "-addons-bootstrap"
	manifestTask := context.Tasks[name]
	if manifestTask == nil {
		t.Fatalf("manifest task not found (%q)", name)
	}

	manifestFileTask := manifestTask.(*fitasks.ManagedFile)
	actualManifest, err := manifestFileTask.Contents.AsString()
	if err != nil {
		t.Fatalf("error getting manifest as string: %v", err)
	}

	expectedManifestPath := path.Join(basedir, "manifest.yaml")
	expectedManifest, err := ioutil.ReadFile(expectedManifestPath)
	if err != nil {
		t.Fatalf("error reading file %q: %v", expectedManifestPath, err)
	}

	if strings.TrimSpace(string(expectedManifest)) != strings.TrimSpace(actualManifest) {
		diffString := diff.FormatDiff(string(expectedManifest), actualManifest)
		t.Logf("diff:\n%s\n", diffString)

		t.Fatalf("manifest differed from expected for test %q", key)
	}
}

func TestBootstrapChannelBuilder_buildManifest(t *testing.T) {
	c := buildDefaultCluster(t)

	c.Spec.Networking.Weave = &api.WeaveNetworkingSpec{}

	bcb := BootstrapChannelBuilder{cluster: c}
	addons, manifests, err := bcb.buildManifest()
	if err != nil {
		t.Fatalf("error building manifests: %v", err)
	}
	if addons == nil {
		t.Fatal("Addons are nil")
	}
	if manifests == nil {
		t.Fatal("Manifests are nil")
	}

	var hasLimit, hasWeave bool

	for _, value := range addons.Spec.Addons {
		if *value.Name == "networking.weave" {
			hasWeave = true
		}

		if *value.Name == "limit-range.addons.k8s.io" {
			hasLimit = true
		}

	}

	if !hasWeave {
		t.Fatal("unable to find weave")
	}

	if !hasLimit {
		t.Fatal("unable to find limit-builder")
	}
}
