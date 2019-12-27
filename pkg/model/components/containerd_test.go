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

package components

import (
	"testing"

	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/assets"
)

func buildContainerdCluster(version string) *api.Cluster {
	return &api.Cluster{
		Spec: api.ClusterSpec{
			CloudProvider:     "aws",
			KubernetesVersion: version,
		},
	}
}

func Test_Build_Containerd_Unsupported_Version(t *testing.T) {
	kubernetesVersions := []string{"1.4.8", "1.5.2", "1.9.0", "1.10.11"}
	for _, v := range kubernetesVersions {

		c := buildContainerdCluster(v)
		c.Spec.ContainerRuntime = "containerd"
		b := assets.NewAssetBuilder(c, "")

		version, err := util.ParseKubernetesVersion(v)
		if err != nil {
			t.Fatalf("unexpected error from ParseKubernetesVersion %s: %v", v, err)
		}

		ob := &ContainerdOptionsBuilder{
			&OptionsContext{
				AssetBuilder:      b,
				KubernetesVersion: *version,
			},
		}

		err = ob.BuildOptions(&c.Spec)
		if err == nil {
			t.Fatalf("expecting error from BuildOptions when Kubernetes version < 1.11: %s", v)
		}
	}
}

func Test_Build_Containerd_Untested_Version(t *testing.T) {
	kubernetesVersions := []string{"1.11.0", "1.11.2", "1.14.0", "1.16.3"}

	for _, v := range kubernetesVersions {

		c := buildContainerdCluster(v)
		c.Spec.ContainerRuntime = "containerd"
		b := assets.NewAssetBuilder(c, "")

		version, err := util.ParseKubernetesVersion(v)
		if err != nil {
			t.Fatalf("unexpected error from ParseKubernetesVersion %s: %v", v, err)
		}

		ob := &ContainerdOptionsBuilder{
			&OptionsContext{
				AssetBuilder:      b,
				KubernetesVersion: *version,
			},
		}

		err = ob.BuildOptions(&c.Spec)
		if err == nil {
			t.Fatalf("expecting error when Kubernetes version >= 1.11 and < 1.18: %s", v)
		}
	}
}

func Test_Build_Containerd_Supported_Version(t *testing.T) {
	kubernetesVersions := []string{"1.18.0", "1.18.3"}

	for _, v := range kubernetesVersions {

		c := buildContainerdCluster(v)
		c.Spec.ContainerRuntime = "containerd"
		b := assets.NewAssetBuilder(c, "")

		version, err := util.ParseKubernetesVersion(v)
		if err != nil {
			t.Fatalf("unexpected error from ParseKubernetesVersion %s: %v", v, err)
		}

		ob := &ContainerdOptionsBuilder{
			&OptionsContext{
				AssetBuilder:      b,
				KubernetesVersion: *version,
			},
		}

		err = ob.BuildOptions(&c.Spec)
		if err != nil {
			t.Fatalf("unexpected error from BuildOptions: %v", err)
		}

		if c.Spec.Containerd.SkipInstall == true {
			t.Fatalf("expecting install when Kubernetes version >= 1.11: %s", v)
		}
	}
}

func Test_Build_Containerd_Unneeded_Runtime(t *testing.T) {
	dockerVersions := []string{"1.13.1", "17.03.2", "18.06.3"}

	for _, v := range dockerVersions {

		c := buildContainerdCluster("1.11.0")
		c.Spec.ContainerRuntime = "docker"
		c.Spec.Docker = &api.DockerConfig{
			Version: &v,
		}
		b := assets.NewAssetBuilder(c, "")

		ob := &ContainerdOptionsBuilder{
			&OptionsContext{
				AssetBuilder: b,
			},
		}

		err := ob.BuildOptions(&c.Spec)
		if err != nil {
			t.Fatalf("unexpected error from BuildOptions: %v", err)
		}

		if c.Spec.Containerd.SkipInstall != true {
			t.Fatalf("unexpected install when Docker version < 19.09: %s", v)
		}
	}
}

func Test_Build_Containerd_Needed_Runtime(t *testing.T) {
	dockerVersions := []string{"18.09.3", "18.09.9", "19.03.4"}

	for _, v := range dockerVersions {

		c := buildContainerdCluster("1.11.0")
		c.Spec.ContainerRuntime = "docker"
		c.Spec.Docker = &api.DockerConfig{
			Version: &v,
		}
		b := assets.NewAssetBuilder(c, "")

		ob := &ContainerdOptionsBuilder{
			&OptionsContext{
				AssetBuilder: b,
			},
		}

		err := ob.BuildOptions(&c.Spec)
		if err != nil {
			t.Fatalf("unexpected error from BuildOptions: %v", err)
		}

		if c.Spec.Containerd.SkipInstall == true {
			t.Fatalf("expected install when Docker version >= 19.09: %s", v)
		}
	}
}
