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

package kutil

import (
	"path"
	"testing"

	"k8s.io/kubernetes/pkg/util/homedir"
)

const (
	RecommendedHomeDir  = ".kube"
	RecommendedFileName = "config"
)

func TestGetKubectlMultiplePath(t *testing.T) {
	c := testCreateKubectlBuilder()
	path := c.getKubectlPath(c.KubeconfigPath)

	if path != "/tmp/config" {
		t.Fatalf("Wrong path got: %s, but expected /tmp/config", path)
	}
}

func TestGetKubectlSinglePath(t *testing.T) {
	c := testCreateKubectlBuilder()
	c.KubeconfigPath = "/bar/config"
	path := c.getKubectlPath(c.KubeconfigPath)

	if path != "/bar/config" {
		t.Fatalf("Wrong path got: %s, but expected /bar/config", path)
	}
}

func TestGetKubectlDefault(t *testing.T) {
	c := testCreateKubectlBuilder()
	c.KubeconfigPath = "/bar/config"
	recommendedHomeFile := path.Join(homedir.HomeDir(), RecommendedHomeDir, RecommendedFileName)
	path := c.getKubectlPath("")

	if path != recommendedHomeFile {
		t.Fatalf("Wrong path got: %s, but expected /bar/config", path)
	}
}

func testCreateKubectlBuilder() *KubeconfigBuilder {
	return &KubeconfigBuilder{
		KubectlPath:     "/usr/local/bin/kubectl",
		KubeconfigPath:  "/tmp/config:/config:path2:path3",
		KubeMasterIP:    "127.0.0.1",
		Context:         "my-context",
		Namespace:       "default",
		KubeBearerToken: "token",
		KubeUser:        "user",
		KubePassword:    "password",
	}

}
