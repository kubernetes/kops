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
	"reflect"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/testutils"
)

func TestRunCreateInstanceGroupGCE(t *testing.T) {
	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	clusterName := "test.k8s.io"

	cloud := h.SetupMockGCE()

	ctx := context.Background()
	f := util.NewFactory(&util.FactoryOptions{
		RegistryPath: "memfs://tests",
	})

	cluster := testutils.BuildMinimalClusterGCE(clusterName, cloud.Project())

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

	// A mixed instances policy is specific to AWS, so a list of machine types is rejected elsewhere.
	options.MachineTypes = []string{"e2-medium", "e2-standard-2"}
	if err := RunCreateInstanceGroup(ctx, f, &stdout, options); err == nil || !strings.Contains(err.Error(), "only supported on AWS") {
		t.Fatalf("Expected error about multiple machine types on GCE, got %v", err)
	}
	options.MachineTypes = nil

	if err := RunCreateInstanceGroup(ctx, f, &stdout, options); err != nil {
		t.Fatalf("could not create instance group: %v", err)
	}
}

func TestRunCreateInstanceGroupMachineTypes(t *testing.T) {
	t.Setenv("SKIP_REGION_CHECK", "1")

	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()
	h.SetupMockAWS()

	clusterName := "test.k8s.io"

	ctx := context.Background()
	f := util.NewFactory(&util.FactoryOptions{
		RegistryPath: "memfs://tests",
	})

	clientset, err := f.KopsClient()
	if err != nil {
		t.Fatalf("error getting clientset: %v", err)
	}
	cluster, err := clientset.CreateCluster(ctx, testutils.BuildMinimalClusterAWS(clusterName))
	if err != nil {
		t.Fatalf("error creating cluster: %v", err)
	}

	grid := []struct {
		name                string
		machineTypes        []string
		expectedMachineType string
		expectedInstances   []string
	}{
		{
			name:                "default",
			expectedMachineType: "t2.medium",
		},
		{
			name:                "single",
			machineTypes:        []string{"m5.large"},
			expectedMachineType: "m5.large",
		},
		{
			name:                "multiple",
			machineTypes:        []string{"m5.large", "c5.large", "m4.large"},
			expectedMachineType: "m5.large",
			expectedInstances:   []string{"m5.large", "c5.large", "m4.large"},
		},
	}
	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			var stdout bytes.Buffer
			options := &CreateInstanceGroupOptions{
				ClusterName:       clusterName,
				InstanceGroupName: "nodes-" + g.name,
				Role:              "Node",
				Subnets:           []string{"subnet-us-test-1a"},
				MachineTypes:      g.machineTypes,
			}
			if err := RunCreateInstanceGroup(ctx, f, &stdout, options); err != nil {
				t.Fatalf("could not create instance group: %v", err)
			}

			ig, err := clientset.InstanceGroupsFor(cluster).Get(ctx, options.InstanceGroupName, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("could not get instance group: %v", err)
			}
			if ig.Spec.MachineType != g.expectedMachineType {
				t.Errorf("expected machine type %q, got %q", g.expectedMachineType, ig.Spec.MachineType)
			}
			var instances []string
			if ig.Spec.MixedInstancesPolicy != nil {
				instances = ig.Spec.MixedInstancesPolicy.Instances
			}
			if !reflect.DeepEqual(instances, g.expectedInstances) {
				t.Errorf("expected mixed instances %v, got %v", g.expectedInstances, instances)
			}
		})
	}

	t.Run("empty", func(t *testing.T) {
		var stdout bytes.Buffer
		options := &CreateInstanceGroupOptions{
			ClusterName:       clusterName,
			InstanceGroupName: "nodes-empty",
			Role:              "Node",
			Subnets:           []string{"subnet-us-test-1a"},
			MachineTypes:      []string{"m5.large", ""},
		}
		if err := RunCreateInstanceGroup(ctx, f, &stdout, options); err == nil {
			t.Fatalf("Expected error for an empty machine type, got nil")
		}
	})
}
