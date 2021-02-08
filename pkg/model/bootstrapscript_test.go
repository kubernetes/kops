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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/testutils/golden"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/hashing"
	"k8s.io/kops/util/pkg/mirrors"
)

func Test_ProxyFunc(t *testing.T) {
	b := &BootstrapScript{}
	ps := &kops.EgressProxySpec{
		HTTPProxy: kops.HTTPProxy{
			Host: "example.com",
			Port: 80,
		},
	}

	script := b.createProxyEnv(ps)

	if script == "" {
		t.Fatalf("script cannot be empty")
	}

	if !strings.HasPrefix(script, "echo \"http_proxy=http://example.com:80\" >> /etc/environment") {
		t.Fatalf("script not setting http_proxy properly")
	}

	ps.ProxyExcludes = "www.google.com,www.kubernetes.io"

	script = b.createProxyEnv(ps)
	if !strings.Contains(script, "no_proxy="+ps.ProxyExcludes) {
		t.Fatalf("script not setting no_proxy properly")
	}
}

type nodeupConfigBuilder struct {
	cluster *kops.Cluster
}

func (n *nodeupConfigBuilder) BuildConfig(ig *kops.InstanceGroup, apiserverAdditionalIPs []string, ca fi.Resource) (*nodeup.Config, error) {
	return nodeup.NewConfig(n.cluster, ig), nil
}

func TestBootstrapUserData(t *testing.T) {
	cs := []struct {
		Role               kops.InstanceGroupRole
		ExpectedFilePath   string
		HookSpecRoles      []kops.InstanceGroupRole
		FileAssetSpecRoles []kops.InstanceGroupRole
	}{
		{
			Role:               "Master",
			ExpectedFilePath:   "tests/data/bootstrapscript_0.txt",
			HookSpecRoles:      []kops.InstanceGroupRole{""},
			FileAssetSpecRoles: []kops.InstanceGroupRole{""},
		},
		{
			Role:               "Master",
			ExpectedFilePath:   "tests/data/bootstrapscript_0.txt",
			HookSpecRoles:      []kops.InstanceGroupRole{"Node"},
			FileAssetSpecRoles: []kops.InstanceGroupRole{"Node"},
		},
		{
			Role:               "Master",
			ExpectedFilePath:   "tests/data/bootstrapscript_1.txt",
			HookSpecRoles:      []kops.InstanceGroupRole{"Master"},
			FileAssetSpecRoles: []kops.InstanceGroupRole{"Master"},
		},
		{
			Role:               "Master",
			ExpectedFilePath:   "tests/data/bootstrapscript_2.txt",
			HookSpecRoles:      []kops.InstanceGroupRole{"Master", "Node"},
			FileAssetSpecRoles: []kops.InstanceGroupRole{"Master", "Node"},
		},
		{
			Role:               "Node",
			ExpectedFilePath:   "tests/data/bootstrapscript_3.txt",
			HookSpecRoles:      []kops.InstanceGroupRole{""},
			FileAssetSpecRoles: []kops.InstanceGroupRole{""},
		},
		{
			Role:               "Node",
			ExpectedFilePath:   "tests/data/bootstrapscript_4.txt",
			HookSpecRoles:      []kops.InstanceGroupRole{"Node"},
			FileAssetSpecRoles: []kops.InstanceGroupRole{"Node"},
		},
		{
			Role:               "Node",
			ExpectedFilePath:   "tests/data/bootstrapscript_3.txt",
			HookSpecRoles:      []kops.InstanceGroupRole{"Master"},
			FileAssetSpecRoles: []kops.InstanceGroupRole{"Master"},
		},
		{
			Role:               "Node",
			ExpectedFilePath:   "tests/data/bootstrapscript_5.txt",
			HookSpecRoles:      []kops.InstanceGroupRole{"Master", "Node"},
			FileAssetSpecRoles: []kops.InstanceGroupRole{"Master", "Node"},
		},
	}

	for i, x := range cs {
		cluster := makeTestCluster(x.HookSpecRoles, x.FileAssetSpecRoles)
		group := makeTestInstanceGroup(x.Role, x.HookSpecRoles, x.FileAssetSpecRoles)
		c := &fi.ModelBuilderContext{
			Tasks: make(map[string]fi.Task),
		}

		caTask := &fitasks.Keypair{
			Name:    fi.String(fi.CertificateIDCA),
			Subject: "cn=kubernetes",
			Type:    "ca",
		}
		c.AddTask(caTask)

		bs := &BootstrapScriptBuilder{
			NodeUpConfigBuilder: &nodeupConfigBuilder{cluster: cluster},
			NodeUpAssets: map[architectures.Architecture]*mirrors.MirroredAsset{
				architectures.ArchitectureAmd64: {
					Locations: []string{"nodeup-amd64-1", "nodeup-amd64-2"},
					Hash:      hashing.MustFromString("833723369ad345a88dd85d61b1e77336d56e61b864557ded71b92b6e34158e6a"),
				},
				architectures.ArchitectureArm64: {
					Locations: []string{"nodeup-arm64-1", "nodeup-arm64-2"},
					Hash:      hashing.MustFromString("e525c28a65ff0ce4f95f9e730195b4e67fdcb15ceb1f36b5ad6921a8a4490c71"),
				},
			},
		}

		res, err := bs.ResourceNodeUp(c, group)
		if err != nil {
			t.Errorf("case %d failed to create nodeup resource. error: %s", i, err)
			continue
		}

		require.Contains(t, c.Tasks, "BootstrapScript/testIG")
		err = c.Tasks["BootstrapScript/testIG"].Run(&fi.Context{Cluster: cluster})
		require.NoError(t, err, "running task")

		actual, err := fi.ResourceAsString(res)
		if err != nil {
			t.Errorf("case %d failed to render nodeup resource. error: %s", i, err)
			continue
		}

		golden.AssertMatchesFile(t, actual, x.ExpectedFilePath)
	}
}

func makeTestCluster(hookSpecRoles []kops.InstanceGroupRole, fileAssetSpecRoles []kops.InstanceGroupRole) *kops.Cluster {
	return &kops.Cluster{
		Spec: kops.ClusterSpec{
			CloudProvider:     "aws",
			KubernetesVersion: "1.7.0",
			Subnets: []kops.ClusterSubnetSpec{
				{Name: "test", Zone: "eu-west-1a"},
			},
			NonMasqueradeCIDR: "10.100.0.0/16",
			EtcdClusters: []kops.EtcdClusterSpec{
				{
					Name: "main",
					Members: []kops.EtcdMemberSpec{
						{
							Name:          "test",
							InstanceGroup: s("ig-1"),
						},
					},
					Version: "3.1.11",
				},
				{
					Name: "events",
					Members: []kops.EtcdMemberSpec{
						{
							Name:          "test",
							InstanceGroup: s("ig-1"),
						},
					},
					Version: "3.1.11",
					Image:   "gcr.io/etcd-development/etcd:v3.1.11",
				},
			},
			NetworkCIDR: "10.79.0.0/24",
			CloudConfig: &kops.CloudConfiguration{
				NodeTags: s("something"),
			},
			ContainerRuntime: "docker",
			Containerd: &kops.ContainerdConfig{
				LogLevel: s("info"),
			},
			Docker: &kops.DockerConfig{
				LogLevel: s("INFO"),
			},
			KubeAPIServer: &kops.KubeAPIServerConfig{
				Image: "CoreOS",
			},
			KubeControllerManager: &kops.KubeControllerManagerConfig{
				CloudProvider: "aws",
			},
			KubeProxy: &kops.KubeProxyConfig{
				CPURequest:    "30m",
				CPULimit:      "30m",
				MemoryRequest: "30Mi",
				MemoryLimit:   "30Mi",
				FeatureGates: map[string]string{
					"AdvancedAuditing": "true",
				},
			},
			KubeScheduler: &kops.KubeSchedulerConfig{
				Image: "SomeImage",
			},
			Kubelet: &kops.KubeletConfigSpec{
				KubeconfigPath: "/etc/kubernetes/config.txt",
			},
			MasterKubelet: &kops.KubeletConfigSpec{
				KubeconfigPath: "/etc/kubernetes/config.cfg",
			},
			EgressProxy: &kops.EgressProxySpec{
				HTTPProxy: kops.HTTPProxy{
					Host: "example.com",
					Port: 80,
				},
			},
			Hooks: []kops.HookSpec{
				{
					ExecContainer: &kops.ExecContainerAction{
						Command: []string{
							"sh",
							"-c",
							"apt-get update",
						},
						Image: "busybox",
					},
					Roles: hookSpecRoles,
				},
			},
			FileAssets: []kops.FileAssetSpec{
				{
					Name:    "iptables-restore",
					Path:    "/var/lib/iptables/rules-save",
					Content: "blah blah",
					Roles:   fileAssetSpecRoles,
				},
			},
		},
	}
}

func makeTestInstanceGroup(role kops.InstanceGroupRole, hookSpecRoles []kops.InstanceGroupRole, fileAssetSpecRoles []kops.InstanceGroupRole) *kops.InstanceGroup {
	return &kops.InstanceGroup{
		ObjectMeta: v1.ObjectMeta{
			Name: "testIG",
		},
		Spec: kops.InstanceGroupSpec{
			Kubelet: &kops.KubeletConfigSpec{
				KubeconfigPath: "/etc/kubernetes/igconfig.txt",
			},
			NodeLabels: map[string]string{
				"labelname": "labelvalue",
				"label2":    "value2",
			},
			Role: role,
			Taints: []string{
				"key1=value1:NoSchedule",
				"key2=value2:NoExecute",
			},
			SuspendProcesses: []string{
				"AZRebalance",
			},
			Hooks: []kops.HookSpec{
				{
					Name: "disable-update-engine.service",
					Before: []string{
						"update-engine.service",
						"kubelet.service",
					},
					Manifest: "Type=oneshot\nExecStart=/usr/bin/systemctl stop update-engine.service",
					Roles:    hookSpecRoles,
				}, {
					Name:     "apply-to-all.service",
					Manifest: "Type=oneshot\nExecStart=/usr/bin/systemctl start apply-to-all.service",
				},
			},
			FileAssets: []kops.FileAssetSpec{
				{
					Name:    "iptables-restore",
					Path:    "/var/lib/iptables/rules-save",
					Content: "blah blah",
					Roles:   fileAssetSpecRoles,
				},
				{
					Name:    "tokens",
					Path:    "/kube/tokens.csv",
					Content: "user,token",
				},
			},
		},
	}
}
