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

package model

import (
	"path"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/upup/pkg/fi"
)

func TestDockerBuilder_Simple(t *testing.T) {
	runDockerBuilderTest(t, "simple")
}

func TestDockerBuilder_1_12_1(t *testing.T) {
	runDockerBuilderTest(t, "docker_1.12.1")
}

func TestDockerBuilder_LogFlags(t *testing.T) {
	runDockerBuilderTest(t, "logflags")
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

	nodeUpModelContext, err := LoadModel(basedir)
	if err != nil {
		t.Fatalf("error parsing cluster yaml %q: %v", basedir, err)
		return
	}

	context := &fi.ModelBuilderContext{
		Tasks: make(map[string]fi.Task),
	}

	builder := DockerBuilder{NodeupModelContext: nodeUpModelContext}

	err = builder.Build(context)
	if err != nil {
		t.Fatalf("error from DockerBuilder Build: %v", err)
		return
	}

	ValidateTasks(t, basedir, context)
}
