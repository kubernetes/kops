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

package cloudup

import (
	"reflect"
	"strings"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
)

func TestKarpenterGCEImageID(t *testing.T) {
	grid := []struct {
		image    string
		expected string
		error    bool
	}{
		{
			image:    "ubuntu-os-cloud/ubuntu-2404-noble-amd64-v20260615",
			expected: "projects/ubuntu-os-cloud/global/images/ubuntu-2404-noble-amd64-v20260615",
		},
		{
			image: "",
			error: true,
		},
		{
			image: "no-project-image",
			error: true,
		},
		{
			image: "too/many/parts",
			error: true,
		},
		{
			image: "/missing-project",
			error: true,
		},
		{
			image: "missing-name/",
			error: true,
		},
	}

	for _, g := range grid {
		t.Run(g.image, func(t *testing.T) {
			actual, err := karpenterGCEImageID(g.image)
			if g.error {
				if err == nil {
					t.Fatalf("expected error, got %q", actual)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if actual != g.expected {
				t.Errorf("imageID = %q, want %q", actual, g.expected)
			}
		})
	}
}

func TestKarpenterGCEBootDisk(t *testing.T) {
	grid := []struct {
		desc       string
		rootVolume *kops.InstanceRootVolumeSpec
		expected   karpenterGCEDisk
	}{
		{
			desc: "defaults",
			expected: karpenterGCEDisk{
				SizeGiB:  128,
				Category: "pd-standard",
				Boot:     true,
			},
		},
		{
			desc: "explicit size and type",
			rootVolume: &kops.InstanceRootVolumeSpec{
				Size: new(int32(200)),
				Type: new("pd-ssd"),
			},
			expected: karpenterGCEDisk{
				SizeGiB:  200,
				Category: "pd-ssd",
				Boot:     true,
			},
		},
		{
			desc: "provisioned iops and throughput",
			rootVolume: &kops.InstanceRootVolumeSpec{
				Size:       new(int32(100)),
				Type:       new("hyperdisk-balanced"),
				IOPS:       new(int32(3000)),
				Throughput: new(int32(140)),
			},
			expected: karpenterGCEDisk{
				SizeGiB:               100,
				Category:              "hyperdisk-balanced",
				Boot:                  true,
				ProvisionedIOPS:       new(int64(3000)),
				ProvisionedThroughput: new(int64(140)),
			},
		},
	}

	for _, g := range grid {
		t.Run(g.desc, func(t *testing.T) {
			ig := &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Role:       kops.InstanceGroupRoleNode,
					RootVolume: g.rootVolume,
				},
			}
			actual, err := karpenterGCEBootDisk(ig)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(*actual, g.expected) {
				t.Errorf("bootDisk = %+v, want %+v", *actual, g.expected)
			}
		})
	}
}

func TestBuildKarpenterGCEKubeletConfiguration(t *testing.T) {
	grid := []struct {
		desc            string
		kubelet         *kops.KubeletConfigSpec
		expectedMaxPods int32
	}{
		{
			desc:            "default maxPods",
			expectedMaxPods: 110,
		},
		{
			desc:            "explicit maxPods",
			kubelet:         &kops.KubeletConfigSpec{MaxPods: new(int32(58))},
			expectedMaxPods: 58,
		},
	}

	for _, g := range grid {
		t.Run(g.desc, func(t *testing.T) {
			ig := &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Role:    kops.InstanceGroupRoleNode,
					Kubelet: g.kubelet,
				},
			}
			kubelet := buildKarpenterGCEKubeletConfiguration(ig)
			if kubelet == nil || kubelet.MaxPods == nil {
				t.Fatalf("expected maxPods to be set")
			}
			if *kubelet.MaxPods != g.expectedMaxPods {
				t.Errorf("maxPods = %d, want %d", *kubelet.MaxPods, g.expectedMaxPods)
			}
		})
	}
}

func TestKarpenterNodePoolGCENodeClassRef(t *testing.T) {
	tf := &TemplateFunctions{}
	tf.Cluster = &kops.Cluster{
		Spec: kops.ClusterSpec{
			CloudProvider: kops.CloudProviderSpec{
				GCE: &kops.GCESpec{},
			},
		},
	}

	ig := &kops.InstanceGroup{
		Spec: kops.InstanceGroupSpec{
			Role: kops.InstanceGroupRoleNode,
		},
	}
	ig.Name = "karpenter-nodes"

	rendered, err := tf.KarpenterNodePool(ig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, expected := range []string{"group: karpenter.k8s.gcp", "kind: GCENodeClass", "name: karpenter-nodes"} {
		if !strings.Contains(rendered, expected) {
			t.Errorf("expected %q in rendered NodePool:\n%s", expected, rendered)
		}
	}
}

func TestKarpenterCapacityTypesGCE(t *testing.T) {
	spot := &kops.InstanceGroup{
		Spec: kops.InstanceGroupSpec{
			GCPProvisioningModel: new("SPOT"),
		},
	}
	if got := karpenterCapacityTypes(spot); !reflect.DeepEqual(got, []string{"spot"}) {
		t.Errorf("capacityTypes = %v, want [spot]", got)
	}
	onDemand := &kops.InstanceGroup{}
	if got := karpenterCapacityTypes(onDemand); !reflect.DeepEqual(got, []string{"on-demand"}) {
		t.Errorf("capacityTypes = %v, want [on-demand]", got)
	}
}
