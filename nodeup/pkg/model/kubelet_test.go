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
	"bytes"
	"io/ioutil"
	"path"
	"sort"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/v1alpha2"
	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/upup/pkg/fi"
)

func Test_InstanceGroupKubeletMerge(t *testing.T) {
	var cluster = &kops.Cluster{}
	cluster.Spec.Kubelet = &kops.KubeletConfigSpec{}
	cluster.Spec.Kubelet.NvidiaGPUs = 0
	cluster.Spec.KubernetesVersion = "1.6.0"

	var instanceGroup = &kops.InstanceGroup{}
	instanceGroup.Spec.Kubelet = &kops.KubeletConfigSpec{}
	instanceGroup.Spec.Kubelet.NvidiaGPUs = 1
	instanceGroup.Spec.Role = kops.InstanceGroupRoleNode

	b := &KubeletBuilder{
		&NodeupModelContext{
			Cluster:       cluster,
			InstanceGroup: instanceGroup,
		},
	}
	var mergedKubeletSpec, err = b.buildKubeletConfigSpec()
	if err != nil {
		t.Error(err)
	}
	if mergedKubeletSpec == nil {
		t.Error("Returned nil kubelet spec")
	}

	if mergedKubeletSpec.NvidiaGPUs != instanceGroup.Spec.Kubelet.NvidiaGPUs {
		t.Errorf("InstanceGroup kubelet value (%d) should be reflected in merged output", instanceGroup.Spec.Kubelet.NvidiaGPUs)
	}
}

func TestTaintsAppliedAfter160(t *testing.T) {
	tests := []struct {
		version           string
		taints            []string
		expectError       bool
		expectSchedulable bool
		expectTaints      []string
	}{
		{
			version: "1.4.9",
		},
		{
			version: "1.5.2",
			taints:  []string{"foo"},
		},
		{
			version:           "1.6.0-alpha.1",
			taints:            []string{"foo"},
			expectTaints:      []string{"foo"},
			expectSchedulable: true,
		},
		{
			version:           "1.6.0",
			taints:            []string{"foo", "bar"},
			expectTaints:      []string{"foo", "bar"},
			expectSchedulable: true,
		},
		{
			version:           "1.7.0",
			taints:            []string{"foo", "bar", "baz"},
			expectTaints:      []string{"foo", "bar", "baz"},
			expectSchedulable: true,
		},
	}

	for _, g := range tests {
		cluster := &kops.Cluster{Spec: kops.ClusterSpec{KubernetesVersion: g.version}}
		ig := &kops.InstanceGroup{Spec: kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleMaster, Taints: g.taints}}

		b := &KubeletBuilder{
			&NodeupModelContext{
				Cluster:       cluster,
				InstanceGroup: ig,
			},
		}
		c, err := b.buildKubeletConfigSpec()

		if g.expectError {
			if err == nil {
				t.Fatalf("Expected error but did not get one for version %q", g.version)
			}

			continue
		} else {
			if err != nil {
				t.Fatalf("Unexpected error for version %q: %v", g.version, err)
			}
		}

		if fi.BoolValue(c.RegisterSchedulable) != g.expectSchedulable {
			t.Fatalf("Expected RegisterSchedulable == %v, got %v", g.expectSchedulable, fi.BoolValue(c.RegisterSchedulable))
		}

		if !stringSlicesEqual(g.expectTaints, c.Taints) {
			t.Fatalf("Expected taints %v, got %v", g.expectTaints, c.Taints)
		}
	}
}

func stringSlicesEqual(exp, other []string) bool {
	if exp == nil && other != nil {
		return false
	}

	if exp != nil && other == nil {
		return false
	}

	if len(exp) != len(other) {
		return false
	}

	for i, e := range exp {
		if other[i] != e {
			return false
		}
	}

	return true
}

func Test_RunKubeletBuilder(t *testing.T) {
	runKubeletBuilderTest(t, "featuregates")
}

func runKubeletBuilderTest(t *testing.T, key string) {
	basedir := path.Join("tests/kubelet/", key)

	clusterYamlPath := path.Join(basedir, "cluster.yaml")
	clusterYaml, err := ioutil.ReadFile(clusterYamlPath)
	if err != nil {
		t.Fatalf("error reading cluster yaml file %q: %v", clusterYamlPath, err)
	}

	var cluster *kops.Cluster
	var instanceGroup *kops.InstanceGroup

	// Codecs provides access to encoding and decoding for the scheme
	codecs := kops.Codecs

	codec := codecs.UniversalDecoder(kops.SchemeGroupVersion)

	sections := bytes.Split(clusterYaml, []byte("\n---\n"))
	for _, section := range sections {
		defaults := &schema.GroupVersionKind{
			Group:   v1alpha2.SchemeGroupVersion.Group,
			Version: v1alpha2.SchemeGroupVersion.Version,
		}
		o, gvk, err := codec.Decode(section, defaults, nil)
		if err != nil {
			t.Errorf("error parsing file %v", err)
		}

		switch v := o.(type) {
		case *kops.Cluster:
			cluster = v
		case *kops.InstanceGroup:
			instanceGroup = v
		default:
			t.Errorf("Unhandled kind %q", gvk)
		}
	}

	context := &fi.ModelBuilderContext{
		Tasks: make(map[string]fi.Task),
	}
	nodeUpModelContext := &NodeupModelContext{
		Cluster:       cluster,
		Architecture:  "amd64",
		Distribution:  distros.DistributionXenial,
		InstanceGroup: instanceGroup,
	}

	builder := KubeletBuilder{NodeupModelContext: nodeUpModelContext}

	kubeletConfig, err := builder.buildKubeletConfig()
	if err != nil {
		t.Errorf("error building kubelet config: %v", err)
	}

	// because of the diff we cannot test maps that include multiple values
	// as maps are not sorted and will change
	kubeletConfig.NodeLabels = make(map[string]string)
	kubeletConfig.NodeLabels["kubernetes.io/role"] = "node"

	err = builder.buildSysConfig(context, kubeletConfig)
	if err != nil {
		t.Fatalf("error from KubeletBuilder Build: %v", err)
	}

	var keys []string
	for key := range context.Tasks {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var yamls []string
	for _, key := range keys {
		task := context.Tasks[key]
		yaml, err := kops.ToRawYaml(task)
		if err != nil {
			t.Fatalf("error serializing task: %v", err)
		}
		yamls = append(yamls, strings.TrimSpace(string(yaml)))
	}

	actualTasksYaml := strings.Join(yamls, "\n---\n")

	tasksYamlPath := path.Join(basedir, "tasks.yaml")
	expectedTasksYamlBytes, err := ioutil.ReadFile(tasksYamlPath)
	if err != nil {
		t.Fatalf("error reading file %q: %v", tasksYamlPath, err)
	}

	actualTasksYaml = strings.TrimSpace(actualTasksYaml)
	expectedTasksYaml := strings.TrimSpace(string(expectedTasksYamlBytes))

	if expectedTasksYaml != actualTasksYaml {
		diffString := diff.FormatDiff(expectedTasksYaml, actualTasksYaml)
		t.Logf("diff:\n%s\n", diffString)

		t.Fatalf("tasks differed from expected for test %q", key)
	}
}
