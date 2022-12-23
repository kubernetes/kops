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
	"k8s.io/kops/util/pkg/architectures"
)

func Test_KubeAPIServer_BuildFlags(t *testing.T) {
	grid := []struct {
		config   kops.KubeAPIServerConfig
		expected string
	}{
		{
			kops.KubeAPIServerConfig{},
			"--secure-port=0",
		},
		{
			kops.KubeAPIServerConfig{
				SecurePort: 443,
			},
			"--secure-port=443",
		},
		{
			kops.KubeAPIServerConfig{
				MaxRequestsInflight: 1000,
			},
			"--max-requests-inflight=1000 --secure-port=0",
		},
		{
			kops.KubeAPIServerConfig{
				MaxMutatingRequestsInflight: 900,
			},
			"--max-mutating-requests-inflight=900 --secure-port=0",
		},
		{
			kops.KubeAPIServerConfig{
				SecurePort: 443,
			},
			"--secure-port=443",
		},
		{
			kops.KubeAPIServerConfig{
				SecurePort:          443,
				MaxRequestsInflight: 1000,
			},
			"--max-requests-inflight=1000 --secure-port=443",
		},
		{
			kops.KubeAPIServerConfig{
				SecurePort:                  443,
				MaxMutatingRequestsInflight: 900,
			},
			"--max-mutating-requests-inflight=900 --secure-port=443",
		},
		{
			kops.KubeAPIServerConfig{
				ServiceNodePortRange: "30000-33000",
			},
			"--secure-port=0 --service-node-port-range=30000-33000",
		},
		{
			kops.KubeAPIServerConfig{
				ExperimentalEncryptionProviderConfig: fi.PtrTo("/srv/kubernetes/encryptionconfig.yaml"),
			},
			"--experimental-encryption-provider-config=/srv/kubernetes/encryptionconfig.yaml --secure-port=0",
		},
		{
			kops.KubeAPIServerConfig{
				EncryptionProviderConfig: fi.PtrTo("/srv/kubernetes/encryptionconfig.yaml"),
			},
			"--encryption-provider-config=/srv/kubernetes/encryptionconfig.yaml --secure-port=0",
		},
		{
			kops.KubeAPIServerConfig{
				TargetRamMB: 320,
			},
			"--secure-port=0 --target-ram-mb=320",
		},
		{
			kops.KubeAPIServerConfig{
				AuditDynamicConfiguration: &[]bool{true}[0],
				ServiceAccountKeyFile:     []string{"/srv/kubernetes/server.key", "/srv/kubernetes/service-account.key"},
			},
			"--audit-dynamic-configuration=true --secure-port=0 --service-account-key-file=/srv/kubernetes/server.key --service-account-key-file=/srv/kubernetes/service-account.key",
		},
		{
			kops.KubeAPIServerConfig{
				AuditDynamicConfiguration: &[]bool{false}[0],
			},
			"--audit-dynamic-configuration=false --secure-port=0",
		},
		{
			kops.KubeAPIServerConfig{
				AuditDynamicConfiguration: &[]bool{true}[0],
			},
			"--audit-dynamic-configuration=true --secure-port=0",
		},
		{
			kops.KubeAPIServerConfig{
				EnableProfiling: &[]bool{false}[0],
			},
			"--profiling=false --secure-port=0",
		},
		{
			kops.KubeAPIServerConfig{
				ClientCAFile: "client-ca.crt",
			},
			"--client-ca-file=client-ca.crt --secure-port=0",
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

func TestKubeAPIServerBuilder(t *testing.T) {
	RunGoldenTest(t, "tests/golden/minimal", "kube-apiserver", func(nodeupModelContext *NodeupModelContext, target *fi.NodeupModelBuilderContext) error {
		builder := KubeAPIServerBuilder{NodeupModelContext: nodeupModelContext}
		return builder.Build(target)
	})
}

func TestAuditConfigAPIServerBuilder(t *testing.T) {
	RunGoldenTest(t, "tests/golden/audit", "kube-apiserver", func(nodeupModelContext *NodeupModelContext, target *fi.NodeupModelBuilderContext) error {
		builder := KubeAPIServerBuilder{NodeupModelContext: nodeupModelContext}
		return builder.Build(target)
	})
}

func TestDeddicatedAPIServerBuilder(t *testing.T) {
	RunGoldenTest(t, "tests/golden/dedicated-apiserver", "kube-apiserver", func(nodeupModelContext *NodeupModelContext, target *fi.NodeupModelBuilderContext) error {
		builder := KubeAPIServerBuilder{NodeupModelContext: nodeupModelContext}
		return builder.Build(target)
	})
}

func TestWithoutEtcdEventsAPIServerBuilder(t *testing.T) {
	RunGoldenTest(t, "tests/golden/without-etcd-events", "kube-apiserver", func(nodeupModelContext *NodeupModelContext, target *fi.NodeupModelBuilderContext) error {
		builder := KubeAPIServerBuilder{NodeupModelContext: nodeupModelContext}
		return builder.Build(target)
	})
}

func TestAwsIamAuthenticator(t *testing.T) {
	RunGoldenTest(t, "tests/golden/awsiam", "kube-apiserver", func(nodeupModelContext *NodeupModelContext, target *fi.NodeupModelBuilderContext) error {
		builder := KubeAPIServerBuilder{NodeupModelContext: nodeupModelContext}
		return builder.Build(target)
	})
}

func TestKubeAPIServerBuilderAMD64(t *testing.T) {
	RunGoldenTest(t, "tests/golden/side-loading", "kube-apiserver-amd64", func(nodeupModelContext *NodeupModelContext, target *fi.NodeupModelBuilderContext) error {
		builder := KubeAPIServerBuilder{NodeupModelContext: nodeupModelContext}
		return builder.Build(target)
	})
}

func TestKubeAPIServerBuilderARM64(t *testing.T) {
	RunGoldenTest(t, "tests/golden/side-loading", "kube-apiserver-arm64", func(nodeupModelContext *NodeupModelContext, target *fi.NodeupModelBuilderContext) error {
		builder := KubeAPIServerBuilder{NodeupModelContext: nodeupModelContext}
		builder.Architecture = architectures.ArchitectureArm64
		return builder.Build(target)
	})
}
