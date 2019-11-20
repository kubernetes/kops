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

func TestDockerPackageNames(t *testing.T) {
	for _, dockerVersion := range dockerVersions {
		if dockerVersion.PlainBinary {
			continue
		}

		sanityCheckPackageName(t, dockerVersion.Source, dockerVersion.Version, dockerVersion.Name)

		for k, p := range dockerVersion.ExtraPackages {
			sanityCheckPackageName(t, p.Source, p.Version, k)
		}
	}
}

func sanityCheckPackageName(t *testing.T, u string, version string, name string) {
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

func TestDockerPackageHashes(t *testing.T) {
	if os.Getenv("VERIFY_HASHES") == "" {
		t.Skip("VERIFY_HASHES not set, won't download & verify docker hashes")
	}

	for _, dockerVersion := range dockerVersions {
		verifyPackageHash(t, dockerVersion.Source, dockerVersion.Hash)

		for _, p := range dockerVersion.ExtraPackages {
			verifyPackageHash(t, p.Source, p.Hash)
		}
	}
}

func verifyPackageHash(t *testing.T, u string, hash string) {
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

func TestDockerBuilder_Simple(t *testing.T) {
	runDockerBuilderTest(t, "simple")
}

func TestDockerBuilder_1_12_1(t *testing.T) {
	runDockerBuilderTest(t, "docker_1.12.1")
}

func TestDockerBuilder_LogFlags(t *testing.T) {
	runDockerBuilderTest(t, "logflags")
}

func TestDockerBuilder_SkipInstall(t *testing.T) {
	runDockerBuilderTest(t, "skipinstall")
}

func TestDockerBuilder_StaticBinary(t *testing.T) {
	runDockerBuilderTest(t, "staticbinary")
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

	context := &fi.ModelBuilderContext{
		Tasks: make(map[string]fi.Task),
	}

	builder := DockerBuilder{NodeupModelContext: nodeUpModelContext}

	err = builder.Build(context)
	if err != nil {
		t.Fatalf("error from DockerBuilder Build: %v", err)
		return
	}

	testutils.ValidateTasks(t, basedir, context)
}
