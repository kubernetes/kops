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
	"io/ioutil"
	"path"
	"strings"
	"testing"

	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/pkg/templates"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/models"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"

	// Register our APIs
	_ "k8s.io/kops/pkg/apis/kops/install"
)

func TestBootstrapChannelBuilder_BuildTasks(t *testing.T) {
	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	h.SetupMockAWS()

	runChannelBuilderTest(t, "simple")
	runChannelBuilderTest(t, "kopeio-vxlan")
	runChannelBuilderTest(t, "weave")
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

	if err := PerformAssignments(cluster); err != nil {
		t.Fatalf("error from PerformAssignments: %v", err)
	}

	assetBuilder := assets.NewAssetBuilder()
	fullSpec, err := PopulateClusterSpec(cluster, assetBuilder)
	if err != nil {
		t.Fatalf("error from PopulateClusterSpec: %v", err)
	}
	cluster = fullSpec

	templates, err := templates.LoadTemplates(cluster, models.NewAssetPath("cloudup/resources"))
	if err != nil {
		t.Fatalf("error building templates: %v", err)
	}
	tf := &TemplateFunctions{cluster: cluster}
	tf.AddTo(templates.TemplateFunctions)

	bcb := BootstrapChannelBuilder{
		cluster:      cluster,
		templates:    templates,
		assetBuilder: assets.NewAssetBuilder(),
	}

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
