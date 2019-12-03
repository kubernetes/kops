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
	"crypto/sha1"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
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

		sanityCheckContainerdPackageName(t, containerdVersion.Source, containerdVersion.Version, containerdVersion.Name)

		for k, p := range containerdVersion.ExtraPackages {
			sanityCheckContainerdPackageName(t, p.Source, p.Version, k)
		}
	}
}

func sanityCheckContainerdPackageName(t *testing.T, u string, version string, name string) {
	filename := u
	lastSlash := strings.LastIndex(filename, "/")
	if lastSlash != -1 {
		filename = filename[lastSlash+1:]
	}

	expectedNames := []string{}
	// Match known RPM formats
	for _, v := range []string{"-1.", "-2.", "-3.", "-3.2."} {
		for _, d := range []string{"el7", "el7.centos", "el7_6"} {
			for _, a := range []string{"noarch", "x86_64"} {
				expectedNames = append(expectedNames, name+"-"+version+v+d+"."+a+".rpm")
			}
		}
	}

	// Match known DEB formats
	for _, a := range []string{"amd64", "armhf"} {
		expectedNames = append(expectedNames, name+"_"+version+"_"+a+".deb")
	}

	found := false
	for _, s := range expectedNames {
		if s == filename {
			found = true
		}
	}
	if !found {
		t.Errorf("unexpected name=%q, version=%q for %s", name, version, u)
	}
}

func TestContainerdPackageHashes(t *testing.T) {
	if os.Getenv("VERIFY_HASHES") == "" {
		t.Skip("VERIFY_HASHES not set, won't download & verify docker hashes")
	}

	for _, containerdVersion := range containerdVersions {
		verifyContainerdPackageHash(t, containerdVersion.Source, containerdVersion.Hash)

		for _, p := range containerdVersion.ExtraPackages {
			verifyContainerdPackageHash(t, p.Source, p.Hash)
		}
	}
}

func verifyContainerdPackageHash(t *testing.T, u string, hash string) {
	resp, err := http.Get(u)
	if err != nil {
		t.Errorf("%s: error fetching: %v", u, err)
		return
	}
	defer resp.Body.Close()

	hasher := sha1.New()
	if _, err := io.Copy(hasher, resp.Body); err != nil {
		t.Errorf("%s: error reading: %v", u, err)
		return
	}

	actualHash := hex.EncodeToString(hasher.Sum(nil))
	if hash != actualHash {
		t.Errorf("%s: hash was %q", u, actualHash)
		return
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
				SkipInstall: false,
				ConfigFile:  fi.String("test"),
				Version:     fi.String("test"),
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
				SkipInstall: false,
				Address:     fi.String("/run/containerd/containerd.sock"),
				ConfigFile:  fi.String("test"),
				LogLevel:    fi.String("info"),
				Root:        fi.String("/var/lib/containerd"),
				State:       fi.String("/run/containerd"),
				Version:     fi.String("test"),
			},
			"--address=/run/containerd/containerd.sock --log-level=info --root=/var/lib/containerd --state=/run/containerd",
		},
		{
			kops.ContainerdConfig{
				SkipInstall: true,
				Address:     fi.String("/run/containerd/containerd.sock"),
				ConfigFile:  fi.String("test"),
				LogLevel:    fi.String("info"),
				Root:        fi.String("/var/lib/containerd"),
				State:       fi.String("/run/containerd"),
				Version:     fi.String("test"),
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
