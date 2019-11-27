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

package components

import (
	"bytes"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/util/pkg/vfs"
)

func TestImage(t *testing.T) {
	grid := []struct {
		Component string
		Cluster   *kops.Cluster

		// File to put into VFS for the test
		VFS map[string]string

		Expected string
	}{
		{
			Component: "kube-apiserver",
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "v1.9.0",
				},
			},
			Expected: "gcr.io/google_containers/kube-apiserver:v1.9.0",
		},
		{
			Component: "kube-apiserver",
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "v1.10.0",
				},
			},
			Expected: "k8s.gcr.io/kube-apiserver:v1.10.0",
		},
		{
			Component: "kube-apiserver",
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "1.10.0",
				},
			},
			Expected: "k8s.gcr.io/kube-apiserver:v1.10.0",
		},
		{
			Component: "kube-apiserver",
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "memfs://v1.9.0-download/",
				},
			},
			VFS: map[string]string{
				"memfs://v1.9.0-download/bin/linux/amd64/kube-apiserver.docker_tag": "1-9-0dockertag",
			},
			Expected: "gcr.io/google_containers/kube-apiserver:1-9-0dockertag",
		},
		{
			Component: "kube-apiserver",
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "memfs://v1.10.0-download/",
				},
			},
			VFS: map[string]string{
				"memfs://v1.10.0-download/bin/linux/amd64/kube-apiserver.docker_tag": "1-10-0dockertag",
			},
			Expected: "k8s.gcr.io/kube-apiserver:1-10-0dockertag",
		},
		{
			Component: "kube-apiserver",
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "memfs://v1.16.0-download/",
				},
			},
			VFS: map[string]string{
				"memfs://v1.16.0-download/bin/linux/amd64/kube-apiserver.docker_tag": "1-16-0dockertag",
			},
			Expected: "k8s.gcr.io/kube-apiserver-amd64:1-16-0dockertag",
		},
		{
			Component: "kube-apiserver",
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "1.16.0",
				},
			},
			Expected: "k8s.gcr.io/kube-apiserver:v1.16.0",
		},
	}

	for _, g := range grid {
		vfs.Context.ResetMemfsContext(true)

		// Populate VFS files
		for k, v := range g.VFS {
			p, err := vfs.Context.BuildVfsPath(k)
			if err != nil {
				t.Errorf("error building vfs path for %s: %v", k, err)
				continue
			}
			if err := p.WriteFile(bytes.NewReader([]byte(v)), nil); err != nil {
				t.Errorf("error writing vfs path %s: %v", k, err)
				continue
			}
		}

		architecture := "amd64"

		assetBuilder := assets.NewAssetBuilder(g.Cluster, "")
		actual, err := Image(g.Component, architecture, &g.Cluster.Spec, assetBuilder)
		if err != nil {
			t.Errorf("unexpected error from image %q %v: %v",
				g.Component, g.Cluster.Spec.KubernetesVersion, err)
			continue
		}
		if actual != g.Expected {
			t.Errorf("unexpected result from image %q %v: actual=%q, expected=%q",
				g.Component, g.Cluster.Spec.KubernetesVersion, actual, g.Expected)
			continue
		}
	}
}
