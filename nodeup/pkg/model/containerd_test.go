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
	"os"
	"path"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi"
)

func TestContainerdPackageNames(t *testing.T) {
	for _, containerdVersion := range containerdVersions {
		if containerdVersion.PlainBinary {
			continue
		}

		sanityCheckPackageName(t, containerdVersion.Source, containerdVersion.Version, containerdVersion.Name)

		for k, p := range containerdVersion.ExtraPackages {
			sanityCheckPackageName(t, p.Source, p.Version, k)
		}
	}
}

func TestContainerdPackageHashes(t *testing.T) {
	if os.Getenv("VERIFY_HASHES") == "" {
		t.Skip("VERIFY_HASHES not set, won't download & verify docker hashes")
	}

	for _, containerdVersion := range containerdVersions {
		t.Run(containerdVersion.Source, func(t *testing.T) {
			if err := verifyPackageHash(containerdVersion.Source, containerdVersion.Hash, containerdVersion.Version); err != nil {
				t.Errorf("error verifying package %q: %v", containerdVersion.Source, err)
			}

			for _, p := range containerdVersion.ExtraPackages {
				if err := verifyPackageHash(p.Source, p.Hash, p.Version); err != nil {
					t.Errorf("error verifying package %q: %v", p.Source, err)
				}
			}
		})
	}
}

func TestContainerdBuilder_Simple(t *testing.T) {
	runContainerdBuilderTest(t, "simple")
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

func runContainerdBuilderTest(t *testing.T, key string) {
	basedir := path.Join("tests/containerdbuilder/", key)

	nodeUpModelContext, err := BuildNodeupModelContext(basedir)
	if err != nil {
		t.Fatalf("error parsing cluster yaml %q: %v", basedir, err)
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

	testutils.ValidateTasks(t, basedir, context)
}
