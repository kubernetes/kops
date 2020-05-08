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

package vfsclientset

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/util/pkg/vfs"
)

func TestSSHCredentialStoreOnConfigBase(t *testing.T) {
	vfs.Context.ResetMemfsContext(true)
	configBase := "memfs://some/config/base"
	cluster := &kops.Cluster{
		Spec: kops.ClusterSpec{
			ConfigBase: configBase,
		},
	}

	p, err := pkiPath(cluster)

	if err != nil {
		t.Errorf("Failed to create ssh path: %v", err)
	}

	actual := p.Path()
	expected := configBase + "/pki"

	if actual != expected {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

func TestSSHCredentialStoreOnOwnCFS(t *testing.T) {
	vfs.Context.ResetMemfsContext(true)
	configBase := "memfs://some/config/base"
	keyPath := "memfs://keys/some/config/base/pki"
	cluster := &kops.Cluster{
		Spec: kops.ClusterSpec{
			ConfigBase: configBase,
			KeyStore:   keyPath,
		},
	}

	p, err := pkiPath(cluster)

	if err != nil {
		t.Errorf("Failed to create ssh path: %v", err)
	}

	actual := p.Path()
	expected := keyPath

	if actual != expected {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}
