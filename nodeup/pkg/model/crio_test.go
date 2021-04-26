/*
Copyright 2020 The Kubernetes Authors.

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

func TestCrioBuilder_Simple(t *testing.T) {
	runCrioBuilderTest(t, "simple")
}

func TestCrioBuilder_SkipInstall(t *testing.T) {
	runCrioBuilderTest(t, "skipinstall")
}

func TestCrioBuildFlags(t *testing.T) {
	grid := []struct {
		config   kops.CrioConfig
		expected string
	}{
		{
			config:   kops.CrioConfig{},
			expected: "",
		},
		{
			config: kops.CrioConfig{
				SkipInstall: false,
				LogLevel:    fi.String("warn"),
			},
			expected: "--log-level=warn",
		},
		{
			config: kops.CrioConfig{
				SkipInstall:    false,
				ConfigOverride: nil,
			},
			expected: "",
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

func runCrioBuilderTest(t *testing.T, key string) {
	basedir := path.Join("tests/criobuilder/", key)

	nodeUpModelContext, err := BuildNodeupModelContext(basedir)
	if err != nil {
		t.Fatalf("error parsing cluster yaml %q: %v", basedir, err)
		return
	}

	nodeUpModelContext.Distribution = distributions.DistributionUbuntu1604

	nodeUpModelContext.Assets = fi.NewAssetStore("")
	nodeUpModelContext.Assets.AddForTest("crio", "bin/containerd", "testing crio content")
	nodeUpModelContext.Assets.AddForTest("crictl", "bin/crictl", "testing crio content")
	nodeUpModelContext.Assets.AddForTest("conmon", "bin/conmon", "testing crio content")
	nodeUpModelContext.Assets.AddForTest("runc", "bin/runc", "testing crio content")
	nodeUpModelContext.Assets.AddForTest("pinns", "bin/pinns", "testing crio content")
	nodeUpModelContext.Assets.AddForTest("crio-status", "bin/crio-status", "testing crio content")
	nodeUpModelContext.Assets.AddForTest("crun", "bin/crun", "testing crio content")

	context := &fi.ModelBuilderContext{
		Tasks: make(map[string]fi.Task),
	}

	builder := CrioBuilder{NodeupModelContext: nodeUpModelContext}

	err = builder.Build(context)
	if err != nil {
		t.Fatalf("error from Criobuilder Build: %v", err)
		return
	}

	testutils.ValidateTasks(t, filepath.Join(basedir, "tasks.yaml"), context)
}
