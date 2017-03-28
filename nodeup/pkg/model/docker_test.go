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
	"io/ioutil"
	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/upup/pkg/fi"
	"path"
	"sort"
	"strings"
	"testing"

	// Register our APIs
	_ "k8s.io/kops/pkg/apis/kops/install"
	"k8s.io/kops/pkg/flagbuilder"
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
				LogDriver: "json-file",
			},
			"--log-driver=json-file",
		},
		{
			kops.DockerConfig{
				LogDriver: "json-file",
				LogOpt:    []string{"max-size=10m"},
			},
			"--log-driver=json-file --log-opt=max-size=10m",
		},
		{
			kops.DockerConfig{
				LogDriver: "json-file",
				LogOpt:    []string{"max-size=10m", "max-file=5"},
			},
			"--log-driver=json-file --log-opt=max-file=5 --log-opt=max-size=10m",
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

	clusterYamlPath := path.Join(basedir, "cluster.yaml")
	clusterYaml, err := ioutil.ReadFile(clusterYamlPath)
	if err != nil {
		t.Fatalf("error reading cluster yaml file %q: %v", clusterYamlPath, err)
	}
	obj, _, err := kops.ParseVersionedYaml(clusterYaml)
	if err != nil {
		t.Fatalf("error parsing cluster yaml %q: %v", clusterYamlPath, err)
	}
	cluster := obj.(*kops.Cluster)

	context := &fi.ModelBuilderContext{
		Tasks: make(map[string]fi.Task),
	}
	nodeUpModelContext := &NodeupModelContext{
		Cluster:      cluster,
		Architecture: "amd64",
		Distribution: distros.DistributionXenial,
	}

	builder := DockerBuilder{NodeupModelContext: nodeUpModelContext}

	err = builder.Build(context)
	if err != nil {
		t.Fatalf("error from DockerBuilder Build: %v", err)
	}

	var keys []string
	for key := range context.Tasks {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var yamls []string
	for _, key := range keys {
		task := context.Tasks[key]
		yaml, err := kops.ToRawYaml(task)
		if err != nil {
			t.Fatalf("error serializing task: %v", err)
		}
		yamls = append(yamls, strings.TrimSpace(string(yaml)))
	}

	actualTasksYaml := strings.Join(yamls, "\n---\n")

	tasksYamlPath := path.Join(basedir, "tasks.yaml")
	expectedTasksYamlBytes, err := ioutil.ReadFile(tasksYamlPath)
	if err != nil {
		t.Fatalf("error reading file %q: %v", tasksYamlPath, err)
	}

	actualTasksYaml = strings.TrimSpace(actualTasksYaml)
	expectedTasksYaml := strings.TrimSpace(string(expectedTasksYamlBytes))

	if expectedTasksYaml != actualTasksYaml {
		diffString := diff.FormatDiff(expectedTasksYaml, actualTasksYaml)
		t.Logf("diff:\n%s\n", diffString)

		t.Fatalf("tasks differed from expected for test %q", key)
	}
}
