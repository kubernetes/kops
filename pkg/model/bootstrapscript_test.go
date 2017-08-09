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

package model

import (
	"io/ioutil"
	"strings"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
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

	if !strings.HasPrefix(script, "export http_proxy=http://example.com:80") {
		t.Fatalf("script not setting http_proxy properly")
	}

	ps.ProxyExcludes = "www.google.com,www.kubernetes.io"

	script = b.createProxyEnv(ps)
	t.Logf(script)
	if !strings.Contains(script, "export no_proxy="+ps.ProxyExcludes) {
		t.Fatalf("script not setting no_proxy properly")
	}
}

func TestBootstrapUserData(t *testing.T) {
	cs := []struct {
		Role             kops.InstanceGroupRole
		ExpectedFilePath string
	}{
		{
			Role:             "Master",
			ExpectedFilePath: "tests/data/bootstrapscript_0.txt",
		},
		{
			Role:             "Node",
			ExpectedFilePath: "tests/data/bootstrapscript_1.txt",
		},
	}

	for i, x := range cs {
		spec := makeTestCluster().Spec
		group := makeTestInstanceGroup(x.Role)

		renderNodeUpConfig := func(ig *kops.InstanceGroup) (*nodeup.Config, error) {
			return &nodeup.Config{}, nil
		}

		bs := &BootstrapScript{
			NodeUpSource:        "NUSource",
			NodeUpSourceHash:    "NUSHash",
			NodeUpConfigBuilder: renderNodeUpConfig,
		}

		res, err := bs.ResourceNodeUp(group, &spec)
		if err != nil {
			t.Errorf("case %d failed to create nodeup resource. error: %s", i, err)
			continue
		}

		actual, err := res.AsString()
		if err != nil {
			t.Errorf("case %d failed to render nodeup resource. error: %s", i, err)
			continue
		}

		expectedBytes, err := ioutil.ReadFile(x.ExpectedFilePath)
		if err != nil {
			t.Fatalf("unexpected error reading ExpectedFilePath %q: %v", x.ExpectedFilePath, err)
		}

		if actual != string(expectedBytes) {
			t.Errorf("case %d, expected: %s. got: %s", i, string(expectedBytes), actual)
		}
	}
}

func makeTestCluster() *kops.Cluster {
	return &kops.Cluster{
		Spec: kops.ClusterSpec{
			CloudProvider:     "aws",
			KubernetesVersion: "1.7.0",
			Subnets: []kops.ClusterSubnetSpec{
				{Name: "test", Zone: "eu-west-1a"},
			},
			NonMasqueradeCIDR: "10.100.0.0/16",
			EtcdClusters: []*kops.EtcdClusterSpec{
				{
					Name: "main",
					Members: []*kops.EtcdMemberSpec{
						{
							Name:          "test",
							InstanceGroup: s("ig-1"),
						},
					},
				},
			},
			NetworkCIDR: "10.79.0.0/24",
			CloudConfig: &kops.CloudConfiguration{
				NodeTags: s("something"),
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
				CPURequest: "30m",
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
		},
	}
}

func makeTestInstanceGroup(role kops.InstanceGroupRole) *kops.InstanceGroup {
	return &kops.InstanceGroup{
		Spec: kops.InstanceGroupSpec{
			Role: role,
			NodeLabels: map[string]string{
				"labelname": "labelvalue",
				"label2":    "value2",
			},
			Taints: []string{
				"key1=value1:NoSchedule",
				"key2=value2:NoExecute",
			},
		},
	}
}
