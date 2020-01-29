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
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/hashing"
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
		// Not entirely clear why some (docker) debian packages have this '5:' prefix
		expectedNames = append(expectedNames, name+"_"+strings.TrimPrefix(version, "5:")+"_"+a+".deb")
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
		t.Run(dockerVersion.Source, func(t *testing.T) {
			if err := verifyPackageHash(dockerVersion.Source, dockerVersion.Hash, dockerVersion.Version); err != nil {
				t.Errorf("error verifying package %q: %v", dockerVersion.Source, err)
			}

			for _, p := range dockerVersion.ExtraPackages {
				if err := verifyPackageHash(p.Source, p.Hash, p.Version); err != nil {
					t.Errorf("error verifying package %q: %v", p.Source, err)
				}
			}
		})
	}
}

func verifyPackageHash(u string, hash string, expectedVersion string) error {
	name := path.Base(u)
	p := filepath.Join("/tmp", name)

	expectedHash, err := hashing.FromString(hash)
	if err != nil {
		return err
	}

	if _, err := fi.DownloadURL(u, p, expectedHash); err != nil {
		return err
	}

	actualHash, err := hashing.HashAlgorithmSHA1.HashFile(p)
	if err != nil {
		return fmt.Errorf("error hashing file: %v", err)
	}

	if hash != actualHash.Hex() {
		return fmt.Errorf("hash was %q, expected %q", actualHash.Hex(), hash)
	}

	if strings.HasSuffix(u, ".deb") {
		cmd := exec.Command("dpkg-deb", "-I", p)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error running 'dpkg-deb -I %s': %v", p, err)
		}

		version := ""
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "Version: ") {
				continue
			}
			version += strings.TrimPrefix(line, "Version: ")
		}
		if expectedVersion != version {
			return fmt.Errorf("unexpected version, actual=%q, expected=%q", version, expectedVersion)
		}

	} else if strings.HasSuffix(u, ".rpm") {
		cmd := exec.Command("rpm", "-qp", "--queryformat", "%{VERSION}", "--nosignature", p)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error running rpm %s: %v", strings.Join(cmd.Args, " "), err)
		}

		version := strings.TrimSpace(string(out))
		if expectedVersion != version {
			return fmt.Errorf("unexpected version, actual=%q, expected=%q", version, expectedVersion)
		}
	} else if strings.HasSuffix(u, ".tgz") || strings.HasSuffix(u, ".tar.gz") {
		if expectedVersion != "" {
			return fmt.Errorf("did not expect version for tgz / tar.gz package")
		}
	} else {
		return fmt.Errorf("unexpected suffix for file (known: .rpm .deb .tar.gz .tgz)")
	}

	return nil
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
