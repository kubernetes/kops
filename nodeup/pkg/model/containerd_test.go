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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/pkg/tomlwriter"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/distributions"
)

func TestContainerdBuilder_Simple(t *testing.T) {
	runContainerdBuilderTest(t, "simple", distributions.DistributionUbuntu2604)
}

func TestContainerdBuilder_Flatcar(t *testing.T) {
	runContainerdBuilderTest(t, "flatcar", distributions.DistributionFlatcar)
}

func TestContainerdBuilder_SkipInstall(t *testing.T) {
	runContainerdBuilderTest(t, "skipinstall", distributions.DistributionUbuntu2604)
}

func TestContainerdBuilder_Complex(t *testing.T) {
	runContainerdBuilderTest(t, "complex", distributions.DistributionUbuntu2604)
}

func TestContainerdBuilder_V3(t *testing.T) {
	runContainerdBuilderTest(t, "v3", distributions.DistributionUbuntu2604)
}

func TestRuncAssetPattern(t *testing.T) {
	for _, test := range []struct {
		name      string
		assetPath string
		want      bool
	}{
		{name: "OCI identity", assetPath: "/runc", want: true},
		{name: "amd64 release", assetPath: "https://github.com/opencontainers/runc/releases/download/v1/runc.amd64", want: true},
		{name: "arm64 release", assetPath: "https://github.com/opencontainers/runc/releases/download/v1/runc.arm64", want: true},
		{name: "bundled path", assetPath: "usr/local/sbin/runc"},
		{name: "relative path", assetPath: "docker/runc"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := runcAssetPattern.MatchString(test.assetPath); got != test.want {
				t.Errorf("runcAssetPattern.MatchString(%q) = %v, want %v", test.assetPath, got, test.want)
			}
		})
	}
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
				ConfigOverride: new("test"),
				Version:        new("test"),
			},
			"",
		},
		{
			kops.ContainerdConfig{
				Address: new("/run/containerd/containerd.sock"),
			},
			"--address=/run/containerd/containerd.sock",
		},
		{
			kops.ContainerdConfig{
				LogLevel: new("info"),
			},
			"--log-level=info",
		},
		{
			kops.ContainerdConfig{
				Root: new("/var/lib/containerd"),
			},
			"--root=/var/lib/containerd",
		},
		{
			kops.ContainerdConfig{
				State: new("/run/containerd"),
			},
			"--state=/run/containerd",
		},
		{
			kops.ContainerdConfig{
				SkipInstall:    false,
				Address:        new("/run/containerd/containerd.sock"),
				ConfigOverride: new("test"),
				LogLevel:       new("info"),
				Root:           new("/var/lib/containerd"),
				State:          new("/run/containerd"),
				Version:        new("test"),
			},
			"--address=/run/containerd/containerd.sock --log-level=info --root=/var/lib/containerd --state=/run/containerd",
		},
		{
			kops.ContainerdConfig{
				SkipInstall:    true,
				Address:        new("/run/containerd/containerd.sock"),
				ConfigOverride: new("test"),
				LogLevel:       new("info"),
				Root:           new("/var/lib/containerd"),
				State:          new("/run/containerd"),
				Version:        new("test"),
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
	nodeUpModelContext.Assets.AddForTest("containerd", "bin/containerd", "testing containerd content")
	nodeUpModelContext.Assets.AddForTest("containerd-shim", "bin/containerd-shim", "testing containerd content")
	nodeUpModelContext.Assets.AddForTest("containerd-shim-runc-v1", "bin/containerd-shim-runc-v1", "testing containerd content")
	nodeUpModelContext.Assets.AddForTest("containerd-shim-runc-v2", "bin/containerd-shim-runc-v2", "testing containerd content")
	nodeUpModelContext.Assets.AddForTest("containerd-stress", "bin/containerd-stress", "testing containerd content")
	nodeUpModelContext.Assets.AddForTest("ctr", "bin/ctr", "testing containerd content")
	nodeUpModelContext.Assets.AddForTest("runc.amd64", "https://github.com/opencontainers/runc/releases/download/v1.1.0/runc.amd64", "testing runc content")
	nodeUpModelContext.Assets.AddForTest("bundled-runc", "usr/local/sbin/runc", "testing bundled runc content")

	if err := nodeUpModelContext.Init(); err != nil {
		t.Fatalf("error from nodeupModelContext.Init(): %v", err)
		return
	}
	context := &fi.NodeupModelBuilderContext{
		Tasks: make(map[string]fi.NodeupTask),
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
	b := &ContainerdBuilder{
		NodeupModelContext: &NodeupModelContext{
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

func TestContainerdConfigVersion(t *testing.T) {
	grid := []struct {
		version      *string
		distribution distributions.Distribution
		expected     int64
	}{
		{version: new("2.1.6"), distribution: distributions.DistributionUbuntu2604, expected: 3},
		{version: new("2.2.4"), distribution: distributions.DistributionUbuntu2604, expected: 3},
		{version: new("2.3.4"), distribution: distributions.DistributionUbuntu2604, expected: 4},
		{version: new("2.3.4"), distribution: distributions.DistributionFlatcar, expected: 3},
		{version: new("2.3.4"), distribution: distributions.DistributionContainerOS, expected: 3},
		{version: nil, distribution: distributions.DistributionUbuntu2604, expected: 3},
	}

	for _, g := range grid {
		b := &ContainerdBuilder{
			NodeupModelContext: &NodeupModelContext{
				NodeupConfig: &nodeup.Config{
					ContainerdConfig: &kops.ContainerdConfig{
						Version: g.version,
					},
				},
			},
		}
		b.Distribution = g.distribution

		if actual := b.containerdConfigVersion(); actual != g.expected {
			t.Errorf("containerdConfigVersion for version %q on %v: got %d, expected %d", fi.ValueOf(g.version), g.distribution, actual, g.expected)
		}
	}
}

func TestContainerdConfigNRITimeouts(t *testing.T) {
	b := &ContainerdBuilder{
		NodeupModelContext: &NodeupModelContext{
			NodeupConfig: &nodeup.Config{
				ContainerdConfig: &kops.ContainerdConfig{
					NRI: &kops.NRIConfig{
						Enabled:                   new(true),
						PluginRequestTimeout:      &metav1.Duration{Duration: 2 * time.Second},
						PluginRegistrationTimeout: &metav1.Duration{Duration: 5 * time.Second},
					},
				},
			},
		},
	}

	config, err := b.buildContainerdConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `version = 3

[plugins]

  [plugins."io.containerd.cri.v1.runtime"]

    [plugins."io.containerd.cri.v1.runtime".containerd]
      default_runtime_name = "runc"

      [plugins."io.containerd.cri.v1.runtime".containerd.runtimes]

        [plugins."io.containerd.cri.v1.runtime".containerd.runtimes.runc]
          runtime_type = "io.containerd.runc.v2"

          [plugins."io.containerd.cri.v1.runtime".containerd.runtimes.runc.options]
            SystemdCgroup = true

  [plugins."io.containerd.nri.v1.nri"]
    disable = false
    plugin_registration_timeout = "5s"
    plugin_request_timeout = "2s"
`
	if config != expected {
		t.Error(diff.FormatDiff(expected, config))
	}
}

func TestAppendGPURuntimeContainerdConfig(t *testing.T) {
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
          runtime_type = "io.containerd.runc.v2"

          [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.nvidia.options]
            BinaryName = "/usr/bin/nvidia-container-runtime"
            SystemdCgroup = true

        [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
          runtime_type = "io.containerd.runc.v2"

          [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
            SystemdCgroup = true
`
	config := tomlwriter.NewTree()
	config.SetPath([]string{"version"}, int64(2))
	config.SetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "default_runtime_name"}, "runc")
	config.SetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "runtimes", "runc", "runtime_type"}, "io.containerd.runc.v2")
	config.SetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "runtimes", "runc", "options", "SystemdCgroup"}, true)

	appendNvidiaGPURuntimeConfig(config.Table("plugins", "io.containerd.grpc.v1.cri", "containerd", "runtimes"))

	newConfig := config.String()

	if newConfig != expectedNewConfig {
		fmt.Println(diff.FormatDiff(expectedNewConfig, newConfig))
		t.Error("new config did not match expected new config")
	}
}
