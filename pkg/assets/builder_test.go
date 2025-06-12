/*
Copyright 2017 The Kubernetes Authors.

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

package assets

import (
	"os"
	"path/filepath"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/testutils/golden"
)

func buildAssetBuilder(t *testing.T) *AssetBuilder {
	builder := &AssetBuilder{
		AssetsLocation: &kops.AssetsSpec{},
		ImageAssets:    []*ImageAsset{},
	}
	return builder
}

func TestValidate_RemapImage_ContainerProxy_AppliesToDockerHub(t *testing.T) {
	builder := buildAssetBuilder(t)

	proxyURL := "proxy.example.com/"
	image := "weaveworks/weave-kube"
	expected := "proxy.example.com/weaveworks/weave-kube"

	builder.AssetsLocation.ContainerProxy = &proxyURL

	remapped := builder.RemapImage(image)
	if remapped != expected {
		t.Errorf("Error remapping image (Expecting: %s, got %s)", expected, remapped)
	}
}

func TestValidate_RemapImage_ContainerProxy_AppliesToSimplifiedDockerHub(t *testing.T) {
	builder := buildAssetBuilder(t)

	proxyURL := "proxy.example.com/"
	image := "debian"
	expected := "proxy.example.com/debian"

	builder.AssetsLocation.ContainerProxy = &proxyURL

	remapped := builder.RemapImage(image)
	if remapped != expected {
		t.Errorf("Error remapping image (Expecting: %s, got %s)", expected, remapped)
	}
}

func TestValidate_RemapImage_ContainerProxy_AppliesToSimplifiedKubernetesURL(t *testing.T) {
	builder := buildAssetBuilder(t)

	proxyURL := "proxy.example.com/"
	image := "registry.k8s.io/kube-apiserver"
	expected := "proxy.example.com/kube-apiserver"

	builder.AssetsLocation.ContainerProxy = &proxyURL

	remapped := builder.RemapImage(image)
	if remapped != expected {
		t.Errorf("Error remapping image (Expecting: %s, got %s)", expected, remapped)
	}
}

func TestValidate_RemapImage_ContainerProxy_AppliesToLegacyKubernetesURL(t *testing.T) {
	builder := buildAssetBuilder(t)

	proxyURL := "proxy.example.com/"
	image := "gcr.io/google_containers/kube-apiserver"
	expected := "proxy.example.com/google_containers/kube-apiserver"

	builder.AssetsLocation.ContainerProxy = &proxyURL

	remapped := builder.RemapImage(image)
	if remapped != expected {
		t.Errorf("Error remapping image (Expecting: %s, got %s)", expected, remapped)
	}
}

func TestValidate_RemapImage_ContainerProxy_AppliesToImagesWithTags(t *testing.T) {
	builder := buildAssetBuilder(t)

	proxyURL := "proxy.example.com/"
	image := "registry.k8s.io/kube-apiserver:1.2.3"
	expected := "proxy.example.com/kube-apiserver:1.2.3"

	builder.AssetsLocation.ContainerProxy = &proxyURL

	remapped := builder.RemapImage(image)
	if remapped != expected {
		t.Errorf("Error remapping image (Expecting: %s, got %s)", expected, remapped)
	}
}

func TestValidate_RemapImage_ContainerRegistry_MappingMultipleTimesConverges(t *testing.T) {
	builder := buildAssetBuilder(t)

	mirrorURL := "proxy.example.com"
	image := "kube-apiserver:1.2.3"
	expected := "proxy.example.com/kube-apiserver:1.2.3"

	builder.AssetsLocation.ContainerRegistry = &mirrorURL

	remapped := image
	iterations := make([]map[int]int, 2)
	for i := range iterations {
		remapped := builder.RemapImage(remapped)
		if remapped != expected {
			t.Errorf("Error remapping image (Expecting: %s, got %s, iteration: %d)", expected, remapped, i)
		}
	}
}

func TestRemapEmptySection(t *testing.T) {
	builder := buildAssetBuilder(t)

	testdir := filepath.Join("testdata")

	key := "emptysection"

	inputPath := filepath.Join(testdir, key+".input.yaml")
	expectedPath := filepath.Join(testdir, key+".expected.yaml")

	input, err := os.ReadFile(inputPath)
	if err != nil {
		t.Errorf("error reading file %q: %v", inputPath, err)
	}

	actual, err := builder.RemapManifest(input)
	if err != nil {
		t.Errorf("error remapping manifest %q: %v", inputPath, err)
	}

	golden.AssertMatchesFile(t, string(actual), expectedPath)
}
