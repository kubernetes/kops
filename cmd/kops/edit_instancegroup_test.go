/*
Copyright 2022 The Kubernetes Authors.

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

package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/pkg/testutils/golden"
)

func TestEditInstanceGroup(t *testing.T) {
	t.Setenv("SKIP_REGION_CHECK", "1")
	var stdout bytes.Buffer

	clusterName := "test.k8s.io"

	cluster := testutils.BuildMinimalCluster(clusterName)
	nodes := testutils.BuildMinimalNodeInstanceGroup("nodes", "subnet-us-test-1a")
	nodes.Spec.Taints = []string{"e2etest:NoSchedule"}

	testutils.NewIntegrationTestHarness(t).SetupMockAWS()

	ctx := context.Background()

	factoryOptions := &util.FactoryOptions{}
	factoryOptions.RegistryPath = "memfs://tests"

	factory := util.NewFactory(factoryOptions)
	clientSet, err := factory.KopsClient()
	if err != nil {
		t.Fatalf("could not create clientset: %v", err)
	}

	cluster, err = clientSet.CreateCluster(ctx, cluster)
	if err != nil {
		t.Fatalf("could not create cluster: %v", err)
	}
	_, err = clientSet.InstanceGroupsFor(cluster).Create(ctx, &nodes, v1.CreateOptions{})
	if err != nil {
		t.Fatalf("could not create instance group: %v", err)
	}

	{
		editOptions := &EditInstanceGroupOptions{
			ClusterName: clusterName,
			GroupName:   "nodes",
			Sets:        []string{"spec.maxSize=10"},
		}
		err := RunEditInstanceGroup(ctx, factory, &stdout, editOptions)
		if err != nil {
			t.Fatalf("could not edit instance group: %v", err)
		}
	}

	storedIG, err := clientSet.InstanceGroupsFor(cluster).Get(ctx, "nodes", v1.GetOptions{})
	if err != nil {
		t.Fatalf("could not get instance group: %v", err)
	}
	storedIG.CreationTimestamp = MagicTimestamp
	actualYAMLBytes, err := kopscodecs.ToVersionedYamlWithVersion(storedIG, schema.GroupVersion{Group: "kops.k8s.io", Version: "v1alpha2"})
	if err != nil {
		t.Fatalf("unexpected error serializing Addon: %v", err)
	}

	actualYAML := strings.TrimSpace(string(actualYAMLBytes))

	golden.AssertMatchesFile(t, actualYAML, "test/edit_instance_group.yaml")
}
