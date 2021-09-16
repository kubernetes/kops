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

package model

import (
	"fmt"
	"path"
	"path/filepath"
	"testing"

	"github.com/pelletier/go-toml"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/distributions"
)

func TestContainerdBuilder_Docker_19_03_13(t *testing.T) {
	runContainerdBuilderTest(t, "from_docker_19.03.11", distributions.DistributionUbuntu2004)
}

func TestContainerdBuilder_Docker_19_03_14(t *testing.T) {
	runContainerdBuilderTest(t, "from_docker_19.03.14", distributions.DistributionUbuntu2004)
}

func TestContainerdBuilder_Simple(t *testing.T) {
	runContainerdBuilderTest(t, "simple", distributions.DistributionUbuntu2004)
}

func TestContainerdBuilder_Flatcar(t *testing.T) {
	runContainerdBuilderTest(t, "flatcar", distributions.DistributionFlatcar)
}

func TestContainerdBuilder_SkipInstall(t *testing.T) {
	runDockerBuilderTest(t, "skipinstall")
}

func TestContainerdBuilder_BuildFlags(t *testing.T) {
	grid := []struct {
		config   kops.ContainerdConfig
		expected string
	}{
		{
			kops.ContainerdConfig{},
			"",
		},
		{
			kops.ContainerdConfig{
				SkipInstall:    false,
				ConfigOverride: fi.String("test"),
				Version:        fi.String("test"),
			},
			"",
		},
		{
			kops.ContainerdConfig{
				Address: fi.String("/run/containerd/containerd.sock"),
			},
			"--address=/run/containerd/containerd.sock",
		},
		{
			kops.ContainerdConfig{
				LogLevel: fi.String("info"),
			},
			"--log-level=info",
		},
		{
			kops.ContainerdConfig{
				Root: fi.String("/var/lib/containerd"),
			},
			"--root=/var/lib/containerd",
		},
		{
			kops.ContainerdConfig{
				State: fi.String("/run/containerd"),
			},
			"--state=/run/containerd",
		},
		{
			kops.ContainerdConfig{
				SkipInstall:    false,
				Address:        fi.String("/run/containerd/containerd.sock"),
				ConfigOverride: fi.String("test"),
				LogLevel:       fi.String("info"),
				Root:           fi.String("/var/lib/containerd"),
				State:          fi.String("/run/containerd"),
				Version:        fi.String("test"),
			},
			"--address=/run/containerd/containerd.sock --log-level=info --root=/var/lib/containerd --state=/run/containerd",
		},
		{
			kops.ContainerdConfig{
				SkipInstall:    true,
				Address:        fi.String("/run/containerd/containerd.sock"),
				ConfigOverride: fi.String("test"),
				LogLevel:       fi.String("info"),
				Root:           fi.String("/var/lib/containerd"),
				State:          fi.String("/run/containerd"),
				Version:        fi.String("test"),
			},
			"--address=/run/containerd/containerd.sock --log-level=info --root=/var/lib/containerd --state=/run/containerd",
		},
	}

	for _, g := range grid {
		actual, err := flagbuilder.BuildFlags(&g.config)
		if err != nil {
			t.Errorf("error building flags for %v: %v", g.config, err)
			continue
		}
		if actual != g.expected {
			t.Errorf("flags did not match.  actual=%q expected=%q", actual, g.expected)
		}
	}
}

func runContainerdBuilderTest(t *testing.T, key string, distro distributions.Distribution) {
	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	h.MockKopsVersion("1.18.0")
	h.SetupMockAWS()

	basedir := path.Join("tests/containerdbuilder/", key)

	model, err := testutils.LoadModel(basedir)
	if err != nil {
		t.Fatal(err)
	}

	nodeUpModelContext, err := BuildNodeupModelContext(model)
	if err != nil {
		t.Fatalf("error parsing cluster yaml %q: %v", basedir, err)
		return
	}

	nodeUpModelContext.Distribution = distro

	nodeUpModelContext.Assets = fi.NewAssetStore("")
	nodeUpModelContext.Assets.AddForTest("containerd", "usr/local/bin/containerd", "testing containerd content")
	nodeUpModelContext.Assets.AddForTest("containerd-shim", "usr/local/bin/containerd-shim", "testing containerd content")
	nodeUpModelContext.Assets.AddForTest("containerd-shim-runc-v1", "usr/local/bin/containerd-shim-runc-v1", "testing containerd content")
	nodeUpModelContext.Assets.AddForTest("containerd-shim-runc-v2", "usr/local/bin/containerd-shim-runc-v2", "testing containerd content")
	nodeUpModelContext.Assets.AddForTest("crictl", "usr/local/bin/crictl", "testing containerd content")
	nodeUpModelContext.Assets.AddForTest("critest", "usr/local/bin/critest", "testing containerd content")
	nodeUpModelContext.Assets.AddForTest("ctr", "usr/local/bin/ctr", "testing containerd content")
	nodeUpModelContext.Assets.AddForTest("runc", "usr/local/sbin/runc", "testing containerd content")

	if err := nodeUpModelContext.Init(); err != nil {
		t.Fatalf("error from nodeupModelContext.Init(): %v", err)
		return
	}
	context := &fi.ModelBuilderContext{
		Tasks: make(map[string]fi.Task),
	}

	builder := ContainerdBuilder{NodeupModelContext: nodeUpModelContext}

	err = builder.Build(context)
	if err != nil {
		t.Fatalf("error from ContainerdBuilder Build: %v", err)
		return
	}

	testutils.ValidateTasks(t, filepath.Join(basedir, "tasks.yaml"), context)
}

func TestContainerdConfig(t *testing.T) {
	cluster := &kops.Cluster{
		Spec: kops.ClusterSpec{
			ContainerRuntime:  "containerd",
			Containerd:        &kops.ContainerdConfig{},
			KubernetesVersion: "1.21.0",
			Networking: &kops.NetworkingSpec{
				Kubenet: &kops.KubenetNetworkingSpec{},
			},
		},
	}

	b := &ContainerdBuilder{
		NodeupModelContext: &NodeupModelContext{
			Cluster: cluster,
			NodeupConfig: &nodeup.Config{
				ContainerdConfig: &kops.ContainerdConfig{},
			},
		},
	}

	config, err := b.buildContainerdConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config == "" {
		t.Errorf("got unexpected empty containerd config")
	}

}

func TestAppendGPURuntimeContainerdConfig(t *testing.T) {
	originalConfig := `version = 2
[plugins]
  [plugins."io.containerd.grpc.v1.cri"]
	[plugins."io.containerd.grpc.v1.cri".containerd]
	  [plugins."io.containerd.grpc.v1.cri".containerd.runtimes]
		[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
		  runtime_type = "io.containerd.runc.v2"
		  [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
			SystemdCgroup = true
`

	expectedNewConfig := `version = 2

[plugins]

  [plugins."io.containerd.grpc.v1.cri"]

    [plugins."io.containerd.grpc.v1.cri".containerd]
      default_runtime_name = "runc"

      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes]

        [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.nvidia]
          privileged_without_host_devices = false
          runtime_engine = ""
          runtime_root = ""
          runtime_type = "io.containerd.runc.v1"

          [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.nvidia.options]
            BinaryName = "/usr/bin/nvidia-container-runtime"
            SystemdCgroup = true

        [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
          runtime_type = "io.containerd.runc.v2"

          [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
            SystemdCgroup = true
`
	config, err := toml.Load(originalConfig)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := appendNvidiaGPURuntimeConfig(config); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	newConfig, err := config.ToTomlString()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if newConfig != expectedNewConfig {
		fmt.Println(diff.FormatDiff(expectedNewConfig, newConfig))
		t.Error("new config did not match expected new config")
	}

}
