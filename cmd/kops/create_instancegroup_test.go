/*
Copyright 2026 The Kubernetes Authors.

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
	"testing"

	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/testutils"
)

func TestRunCreateInstanceGroup(t *testing.T) {
	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	clusterName := "test.k8s.io"

	h.SetupMockGCE()

	ctx := context.Background()
	f := util.NewFactory(&util.FactoryOptions{
		RegistryPath: "memfs://tests",
	})

	cluster := testutils.BuildMinimalClusterGCE(clusterName)

	clientset, err := f.KopsClient()
	if err != nil {
		t.Fatalf("error getting clientset: %v", err)
	}
	if _, err := clientset.CreateCluster(ctx, cluster); err != nil {
		t.Fatalf("error creating cluster: %v", err)
	}

	var stdout bytes.Buffer
	options := &CreateInstanceGroupOptions{
		ClusterName:       clusterName,
		InstanceGroupName: "nodes",
		Role:              "Node",
		Subnets:           []string{"us-test-1a"},
	}

	if err := RunCreateInstanceGroup(ctx, f, &stdout, options); err == nil {
		t.Fatalf("Expected error when creating instancegroup, got nil")
	}

	options.Zones = []string{"us-test-1a"}

	if err := RunCreateInstanceGroup(ctx, f, &stdout, options); err != nil {
		t.Fatalf("could not create instance group: %v", err)
	}
}
