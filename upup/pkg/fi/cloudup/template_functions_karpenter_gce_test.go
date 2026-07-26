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
	"os"
	"reflect"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"
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

func TestKarpenterNodePoolGCEInstanceRequirements(t *testing.T) {
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
			MixedInstancesPolicy: &kops.MixedInstancesPolicySpec{
				InstanceRequirements: &kops.InstanceRequirementsSpec{
					CPU: &kops.MinMaxSpec{
						Min: new(resource.MustParse("2")),
						Max: new(resource.MustParse("16")),
					},
					Memory: &kops.MinMaxSpec{
						Min: new(resource.MustParse("4Gi")),
						Max: new(resource.MustParse("32Gi")),
					},
					ExcludedInstanceTypes: []string{"e2.*", "n1-standard-1"},
				},
			},
		},
	}
	ig.Name = "karpenter-nodes"

	rendered, err := tf.KarpenterNodePool(ig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, expected := range []string{
		"key: karpenter.k8s.gcp/instance-cpu",
		"key: karpenter.k8s.gcp/instance-memory",
		"key: karpenter.k8s.gcp/instance-family",
		"key: node.kubernetes.io/instance-type",
	} {
		if !strings.Contains(rendered, expected) {
			t.Errorf("expected %q in rendered NodePool:\n%s", expected, rendered)
		}
	}
	if strings.Contains(rendered, "karpenter.k8s.aws") {
		t.Errorf("unexpected AWS requirement key in GCE NodePool:\n%s", rendered)
	}
}

func TestKarpenterGCENodeClassSubnetRangeNames(t *testing.T) {
	grid := []struct {
		desc       string
		networking kops.NetworkingSpec
		wantRanges []string
	}{
		{
			desc:       "ip-alias networking sets subnetRangeNames",
			networking: kops.NetworkingSpec{GCP: &kops.GCPNetworkingSpec{}},
			wantRanges: []string{"pods-test-example-com"},
		},
		{
			desc:       "non-ip-alias networking omits subnetRangeNames",
			networking: kops.NetworkingSpec{CNI: &kops.CNINetworkingSpec{}},
			wantRanges: nil,
		},
	}

	for _, g := range grid {
		t.Run(g.desc, func(t *testing.T) {
			networking := g.networking
			networking.Subnets = []kops.ClusterSubnetSpec{
				{Name: "us-central1", Region: "us-central1", Type: kops.SubnetTypePrivate},
			}
			cluster := &kops.Cluster{
				Spec: kops.ClusterSpec{
					CloudProvider: kops.CloudProviderSpec{
						GCE: &kops.GCESpec{Project: "testproject"},
					},
					Networking: networking,
				},
			}
			cluster.ObjectMeta.Name = "test.example.com"

			ig := &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Role:    kops.InstanceGroupRoleNode,
					Manager: kops.InstanceManagerKarpenter,
					Image:   "ubuntu-os-cloud/ubuntu-2404-noble-amd64-v20260615",
					Subnets: []string{"us-central1"},
				},
			}
			ig.ObjectMeta.Name = "karpenter-nodes"

			tf := &TemplateFunctions{
				KopsModelContext: model.KopsModelContext{
					IAMModelContext:   iam.IAMModelContext{Cluster: cluster},
					Region:            "us-central1",
					AllInstanceGroups: []*kops.InstanceGroup{ig},
					InstanceGroups:    []*kops.InstanceGroup{ig},
				},
				tasks: map[string]fi.CloudupTask{
					"ManagedFile/nodeupscript-karpenter-nodes": &fitasks.ManagedFile{
						Name:     new("nodeupscript-karpenter-nodes"),
						Contents: fi.NewStringResource("#!/bin/sh\necho hi\n"),
					},
				},
			}

			nodeClass, err := tf.buildKarpenterGCENodeClass(ig)
			if err != nil {
				t.Fatalf("buildKarpenterGCENodeClass() error = %v", err)
			}
			if !reflect.DeepEqual(nodeClass.Spec.SubnetRangeNames, g.wantRanges) {
				t.Errorf("SubnetRangeNames = %v, want %v", nodeClass.Spec.SubnetRangeNames, g.wantRanges)
			}
		})
	}
}

func TestRenderKarpenterGCEAddonTemplate(t *testing.T) {
	source, err := os.ReadFile("../../../models/cloudup/resources/addons/karpenter.sh/k8s-1.19-gce.yaml.template")
	if err != nil {
		t.Fatalf("reading k8s-1.19-gce.yaml.template: %v", err)
	}

	cluster := &kops.Cluster{
		Spec: kops.ClusterSpec{
			CloudProvider: kops.CloudProviderSpec{
				GCE: &kops.GCESpec{Project: "testproject"},
			},
			Karpenter: &kops.KarpenterConfig{
				Enabled:  true,
				Image:    "ghcr.io/karpenter-gcp/karpenter-provider-gcp:v0.8.3",
				LogLevel: "debug",
			},
			Networking: kops.NetworkingSpec{
				GCP: &kops.GCPNetworkingSpec{},
				Subnets: []kops.ClusterSubnetSpec{
					{Name: "us-central1", Region: "us-central1", Type: kops.SubnetTypePrivate},
				},
			},
		},
	}
	cluster.ObjectMeta.Name = "test.example.com"

	ig := &kops.InstanceGroup{
		Spec: kops.InstanceGroupSpec{
			Role:    kops.InstanceGroupRoleNode,
			Manager: kops.InstanceManagerKarpenter,
			Image:   "ubuntu-os-cloud/ubuntu-2404-noble-amd64-v20260615",
			Subnets: []string{"us-central1"},
		},
	}
	ig.ObjectMeta.Name = "karpenter-nodes"

	renderer := &addonTemplateRenderer{
		modelContext: &model.KopsModelContext{
			IAMModelContext:   iam.IAMModelContext{Cluster: cluster},
			Region:            "us-central1",
			AllInstanceGroups: []*kops.InstanceGroup{ig},
			InstanceGroups:    []*kops.InstanceGroup{ig},
		},
	}
	tasks := map[string]fi.CloudupTask{
		"ManagedFile/nodeupscript-karpenter-nodes": &fitasks.ManagedFile{
			Name:     new("nodeupscript-karpenter-nodes"),
			Contents: fi.NewStringResource("#!/bin/sh\necho hi\n"),
		},
	}

	rendered, err := renderer.RenderTemplate("k8s-1.19-gce.yaml", source, tasks)
	if err != nil {
		t.Fatalf("RenderTemplate() error = %v", err)
	}

	objects, err := kubemanifest.LoadObjectsFrom(rendered)
	if err != nil {
		t.Fatalf("LoadObjectsFrom() error = %v", err)
	}

	found := map[string]bool{}
	for _, obj := range objects {
		found[obj.Kind()+"/"+obj.GetName()] = true
	}
	for _, want := range []string{
		"Deployment/karpenter",
		"GCENodeClass/karpenter-nodes",
		"NodePool/karpenter-nodes",
	} {
		if !found[want] {
			t.Errorf("expected rendered object %q, not found", want)
		}
	}
	renderedStr := string(rendered)
	for _, want := range []string{
		"value: self-hosted",
		"value: 'testproject'",
		"value: 'us-central1'",
		"value: 'test.example.com'",
		"StaticCapacity=true",
		"pods-test-example-com",
	} {
		if !strings.Contains(renderedStr, want) {
			t.Errorf("expected %q in rendered addon manifest", want)
		}
	}
}
