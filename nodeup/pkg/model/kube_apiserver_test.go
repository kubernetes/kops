/*
Copyright 2017 The Kubernetes Authors.

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
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/upup/pkg/fi"
)

func Test_KubeAPIServer_BuildFlags(t *testing.T) {
	grid := []struct {
		config   kops.KubeAPIServerConfig
		expected string
	}{
		{
			kops.KubeAPIServerConfig{},
			"--insecure-port=0 --secure-port=0",
		},
		{
			kops.KubeAPIServerConfig{
				SecurePort: 443,
			},
			"--insecure-port=0 --secure-port=443",
		},
		{
			kops.KubeAPIServerConfig{
				MaxRequestsInflight: 1000,
			},
			"--insecure-port=0 --max-requests-inflight=1000 --secure-port=0",
		},
		{
			kops.KubeAPIServerConfig{
				MaxMutatingRequestsInflight: 900,
			},
			"--insecure-port=0 --max-mutating-requests-inflight=900 --secure-port=0",
		},
		{
			kops.KubeAPIServerConfig{
				InsecurePort: 8080,
				SecurePort:   443,
			},
			"--insecure-port=8080 --secure-port=443",
		},
		{
			kops.KubeAPIServerConfig{
				InsecurePort:        8080,
				SecurePort:          443,
				MaxRequestsInflight: 1000,
			},
			"--insecure-port=8080 --max-requests-inflight=1000 --secure-port=443",
		},
		{
			kops.KubeAPIServerConfig{
				InsecurePort:                8080,
				SecurePort:                  443,
				MaxMutatingRequestsInflight: 900,
			},
			"--insecure-port=8080 --max-mutating-requests-inflight=900 --secure-port=443",
		},
		{
			kops.KubeAPIServerConfig{
				InsecurePort: 8080,
			},
			"--insecure-port=8080 --secure-port=0",
		},
		{
			kops.KubeAPIServerConfig{
				ServiceNodePortRange: "30000-33000",
			},
			"--insecure-port=0 --secure-port=0 --service-node-port-range=30000-33000",
		},
		{
			kops.KubeAPIServerConfig{
				ExperimentalEncryptionProviderConfig: fi.String("/srv/kubernetes/encryptionconfig.yaml"),
			},
			"--experimental-encryption-provider-config=/srv/kubernetes/encryptionconfig.yaml --insecure-port=0 --secure-port=0",
		},
		{
			kops.KubeAPIServerConfig{
				TargetRamMb: 320,
			},
			"--insecure-port=0 --secure-port=0 --target-ram-mb=320",
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
