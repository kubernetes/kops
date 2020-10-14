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

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/distributions"
)

func TestDockerBuilder_Simple(t *testing.T) {
	runDockerBuilderTest(t, "simple")
}

func TestDockerBuilder_18_06_3(t *testing.T) {
	runDockerBuilderTest(t, "docker_18.06.3")
}

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
			kops.DockerConfig{Bridge: fi.String("")},
			"",
		},
		{
			kops.DockerConfig{Bridge: fi.String("br0")},
			"--bridge=br0",
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
	basedir := path.Join("tests/dockerbuilder/", key)

	nodeUpModelContext, err := BuildNodeupModelContext(basedir)
	if err != nil {
		t.Fatalf("error parsing cluster yaml %q: %v", basedir, err)
		return
	}

	nodeUpModelContext.Distribution = distributions.DistributionUbuntu1604

	nodeUpModelContext.Assets = fi.NewAssetStore("")
	nodeUpModelContext.Assets.AddForTest("containerd", "docker/containerd", "testing Docker content")
	nodeUpModelContext.Assets.AddForTest("containerd-shim", "docker/containerd-shim", "testing Docker content")
	nodeUpModelContext.Assets.AddForTest("ctr", "docker/ctr", "testing Docker content")
	nodeUpModelContext.Assets.AddForTest("docker", "docker/docker", "testing Docker content")
	nodeUpModelContext.Assets.AddForTest("docker-init", "docker/docker-init", "testing Docker content")
	nodeUpModelContext.Assets.AddForTest("docker-proxy", "docker/docker-proxy", "testing Docker content")
	nodeUpModelContext.Assets.AddForTest("dockerd", "docker/dockerd", "testing Docker content")
	nodeUpModelContext.Assets.AddForTest("runc", "docker/runc", "testing Docker content")

	context := &fi.ModelBuilderContext{
		Tasks: make(map[string]fi.Task),
	}

	builder := DockerBuilder{NodeupModelContext: nodeUpModelContext}

	err = builder.Build(context)
	if err != nil {
		t.Fatalf("error from DockerBuilder Build: %v", err)
		return
	}

	testutils.ValidateTasks(t, filepath.Join(basedir, "tasks.yaml"), context)
}
