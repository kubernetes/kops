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
	"path"
	"path/filepath"
	"testing"

	"github.com/blang/semver/v4"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/distributions"
)

func TestDockerBuilder_19_03_11(t *testing.T) {
	runDockerBuilderTest(t, "docker_19.03.11")
}

func TestDockerBuilder_LogFlags(t *testing.T) {
	runDockerBuilderTest(t, "logflags")
}

func TestDockerBuilder_SkipInstall(t *testing.T) {
	runDockerBuilderTest(t, "skipinstall")
}

func TestDockerBuilder_HealthCheck(t *testing.T) {
	runDockerBuilderTest(t, "healthcheck")
}

func TestDockerBuilder_BuildFlags(t *testing.T) {
	logDriver := "json-file"
	grid := []struct {
		config   kops.DockerConfig
		expected string
	}{
		{
			kops.DockerConfig{},
			"",
		},
		{
			kops.DockerConfig{
				LogDriver: &logDriver,
			},
			"--log-driver=json-file",
		},
		{
			kops.DockerConfig{
				LogDriver: &logDriver,
				LogOpt:    []string{"max-size=10m"},
			},
			"--log-driver=json-file --log-opt=max-size=10m",
		},
		{
			kops.DockerConfig{
				LogDriver: &logDriver,
				LogOpt:    []string{"max-size=10m", "max-file=5"},
			},
			"--log-driver=json-file --log-opt=max-file=5 --log-opt=max-size=10m",
		},
		// nil bridge & empty bridge are the same
		{
			kops.DockerConfig{Bridge: nil},
			"",
		},
		{
			kops.DockerConfig{Bridge: fi.PtrTo("")},
			"",
		},
		{
			kops.DockerConfig{Bridge: fi.PtrTo("br0")},
			"--bridge=br0",
		},
		{
			kops.DockerConfig{ExecOpt: []string{"native.cgroupdriver=systemd"}},
			"--exec-opt=native.cgroupdriver=systemd",
		},
		{
			kops.DockerConfig{InsecureRegistries: []string{"registry1", "registry2"}},
			"--insecure-registry=registry1 --insecure-registry=registry2",
		},
		{
			kops.DockerConfig{DNS: []string{}},
			"",
		},
		{
			kops.DockerConfig{DNS: []string{"8.8.4.4"}},
			"--dns=8.8.4.4",
		},
		{
			kops.DockerConfig{DNS: []string{"8.8.4.4", "8.8.8.8"}},
			"--dns=8.8.4.4 --dns=8.8.8.8",
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

func runDockerBuilderTest(t *testing.T, key string) {
	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	h.MockKopsVersion("1.18.0")
	h.SetupMockAWS()

	basedir := path.Join("tests/dockerbuilder/", key)

	model, err := testutils.LoadModel(basedir)
	if err != nil {
		t.Fatal(err)
	}

	nodeUpModelContext, err := BuildNodeupModelContext(model)
	if err != nil {
		t.Fatalf("error parsing cluster yaml %q: %v", basedir, err)
		return
	}

	nodeUpModelContext.Distribution = distributions.DistributionUbuntu2004

	if nodeUpModelContext.NodeupConfig.Docker.SkipInstall == false {
		if nodeUpModelContext.NodeupConfig.Docker.Version == nil {
			t.Fatalf("error finding Docker version")
			return
		}
		dv := fi.ValueOf(nodeUpModelContext.NodeupConfig.Docker.Version)
		sv, err := semver.ParseTolerant(dv)
		if err != nil {
			t.Fatalf("error parsing Docker version %q: %v", dv, err)
			return
		}
		nodeUpModelContext.Assets = fi.NewAssetStore("")
		if sv.GTE(semver.MustParse("19.3.0")) {
			nodeUpModelContext.Assets.AddForTest("containerd", "docker/containerd", "testing Docker content")
			nodeUpModelContext.Assets.AddForTest("containerd-shim", "docker/containerd-shim", "testing Docker content")
			nodeUpModelContext.Assets.AddForTest("ctr", "docker/ctr", "testing Docker content")
			nodeUpModelContext.Assets.AddForTest("docker", "docker/docker", "testing Docker content")
			nodeUpModelContext.Assets.AddForTest("docker-init", "docker/docker-init", "testing Docker content")
			nodeUpModelContext.Assets.AddForTest("docker-proxy", "docker/docker-proxy", "testing Docker content")
			nodeUpModelContext.Assets.AddForTest("dockerd", "docker/dockerd", "testing Docker content")
			nodeUpModelContext.Assets.AddForTest("runc", "docker/runc", "testing Docker content")
		} else {
			nodeUpModelContext.Assets.AddForTest("docker", "docker/docker", "testing Docker content")
			nodeUpModelContext.Assets.AddForTest("docker-containerd", "docker/docker-containerd", "testing Docker content")
			nodeUpModelContext.Assets.AddForTest("docker-containerd-ctr", "docker/docker-containerd-ctr", "testing Docker content")
			nodeUpModelContext.Assets.AddForTest("docker-containerd-shim", "docker/docker-containerd-shim", "testing Docker content")
			nodeUpModelContext.Assets.AddForTest("docker-init", "docker/docker-init", "testing Docker content")
			nodeUpModelContext.Assets.AddForTest("docker-proxy", "docker/docker-proxy", "testing Docker content")
			nodeUpModelContext.Assets.AddForTest("docker-runc", "docker/docker-runc", "testing Docker content")
			nodeUpModelContext.Assets.AddForTest("dockerd", "docker/dockerd", "testing Docker content")
		}
	}

	if err := nodeUpModelContext.Init(); err != nil {
		t.Fatalf("error from nodeUpModelContext.Init(): %v", err)
	}
	context := &fi.NodeupModelBuilderContext{
		Tasks: make(map[string]fi.NodeupTask),
	}

	builder := DockerBuilder{NodeupModelContext: nodeUpModelContext}

	err = builder.Build(context)
	if err != nil {
		t.Fatalf("error from DockerBuilder Build: %v", err)
		return
	}

	testutils.ValidateTasks(t, filepath.Join(basedir, "tasks.yaml"), context)
}
