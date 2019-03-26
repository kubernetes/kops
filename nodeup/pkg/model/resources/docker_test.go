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

package resources

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestDockerPackageNames(t *testing.T) {
	for _, dockerVersion := range DockerVersions {
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
	for _, v := range []string{"-1.", "-2.", "-3."} {
		for _, d := range []string{"el7", "el7.centos"} {
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

	for _, dockerVersion := range DockerVersions {
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
