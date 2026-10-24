/*
Copyright 2026 The Kubernetes Authors.

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

package model

import (
	"slices"
	"testing"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

func TestAddCNIBinAssets(t *testing.T) {
	const cniArchive = "https://github.com/containernetworking/plugins/releases/download/v1.9.1/cni-plugins-linux-amd64-v1.9.1.tgz"

	assets := fi.NewAssetStore("")
	assets.AddForTest("kubelet", "https://dl.k8s.io/release/v1.35.0/bin/linux/amd64/kubelet", "kubelet")
	assets.AddForTest("cni-plugins-linux-amd64-v1.9.1.tgz", cniArchive, "cni archive")
	assets.AddArchiveFileForTest(cniArchive, "bridge", "bridge", "bridge")
	assets.AddArchiveFileForTest(cniArchive, "portmap", "portmap", "portmap")
	assets.AddArchiveFileForTest(cniArchive, "LICENSE", "LICENSE", "license")
	const containerdArchive = "https://github.com/containerd/containerd/releases/download/v2.1.4/containerd-2.1.4-linux-amd64.tar.gz"
	assets.AddForTest("containerd-2.1.4-linux-amd64.tar.gz", containerdArchive, "containerd archive")
	assets.AddArchiveFileForTest(containerdArchive, "containerd", "bin/containerd", "containerd")
	assets.AddArchiveFileForTest("https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.34.0/crictl-v1.34.0-linux-amd64.tar.gz", "crictl", "crictl", "crictl")

	b := &NodeupModelContext{Assets: assets}
	c := &fi.NodeupModelBuilderContext{Tasks: make(map[string]fi.NodeupTask)}
	if err := b.AddCNIBinAssets(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var paths []string
	for _, task := range c.Tasks {
		paths = append(paths, task.(*nodetasks.File).Path)
	}
	slices.Sort(paths)
	expected := []string{"/opt/cni/bin/LICENSE", "/opt/cni/bin/bridge", "/opt/cni/bin/portmap"}
	if !slices.Equal(paths, expected) {
		t.Errorf("unexpected files, got %v, want %v", paths, expected)
	}
}

func TestAddCNIBinAssetsWithoutCNIArchive(t *testing.T) {
	assets := fi.NewAssetStore("")
	assets.AddForTest("kubelet", "https://dl.k8s.io/release/v1.35.0/bin/linux/amd64/kubelet", "kubelet")

	b := &NodeupModelContext{Assets: assets}
	c := &fi.NodeupModelBuilderContext{Tasks: make(map[string]fi.NodeupTask)}
	if err := b.AddCNIBinAssets(c); err == nil {
		t.Errorf("expected an error when no cni-plugins archive was extracted, got %d tasks", len(c.Tasks))
	}
}
