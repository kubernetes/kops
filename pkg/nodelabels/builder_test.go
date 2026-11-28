/*
Copyright 2020 The Kubernetes Authors.

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

package nodelabels

import (
	"reflect"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
)

func TestBuildNodeLabels(t *testing.T) {
	origAPIServerNodes := featureflag.APIServerNodes.Enabled()
	featureflag.ParseFlags("-APIServerNodes")
	defer func() {
		if origAPIServerNodes {
			featureflag.ParseFlags("+APIServerNodes")
		} else {
			featureflag.ParseFlags("-APIServerNodes")
		}
	}()

	tests := []struct {
		name     string
		cluster  *kops.Cluster
		ig       *kops.InstanceGroup
		expected map[string]string

		// expectError is set for roles BuildNodeLabels should reject.
		expectError bool

		// Allow us to test labels at different feature flag levels
		featureFlags string
	}{
		{
			name: "RoleControlPlane",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "v1.32.0",
					ControlPlaneKubelet: &kops.KubeletConfigSpec{
						NodeLabels: map[string]string{
							"controlPlane1": "controlPlane1",
							"controlPlane2": "controlPlane2",
						},
					},
					Kubelet: &kops.KubeletConfigSpec{
						NodeLabels: map[string]string{
							"node1": "node1",
							"node2": "node2",
						},
					},
				},
			},
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Role: kops.InstanceGroupRoleControlPlane,
					Kubelet: &kops.KubeletConfigSpec{
						NodeLabels: map[string]string{
							"node1": "override1",
							"node3": "override3",
						},
					},
				},
			},
			expected: map[string]string{
				RoleLabelControlPlane20:                                   "",
				"node.kubernetes.io/exclude-from-external-load-balancers": "",
				"kops.k8s.io/kops-controller-pki":                         "",
				"controlPlane1":                                           "controlPlane1",
				"controlPlane2":                                           "controlPlane2",
				"node1":                                                   "override1",
				"node3":                                                   "override3",
			},
		},
		{
			name: "RoleControlPlaneWithAPIServerNodes",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "v1.31.0",
					ControlPlaneKubelet: &kops.KubeletConfigSpec{
						NodeLabels: map[string]string{
							"controlPlane1": "controlPlane1",
							"controlPlane2": "controlPlane2",
						},
					},
					Kubelet: &kops.KubeletConfigSpec{
						NodeLabels: map[string]string{
							"node1": "node1",
							"node2": "node2",
						},
					},
				},
			},
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Role: kops.InstanceGroupRoleAPIServer,
					Kubelet: &kops.KubeletConfigSpec{
						NodeLabels: map[string]string{
							"node1": "override1",
							"node3": "override3",
						},
					},
				},
			},
			expected: map[string]string{
				"node-role.kubernetes.io/api-server": "",
				"kops.k8s.io/kops-controller-pki":    "",
				"controlPlane1":                      "controlPlane1",
				"controlPlane2":                      "controlPlane2",
				"node1":                              "override1",
				"node3":                              "override3",
			},
			featureFlags: "+APIServerNodes",
		},
		{
			name: "RoleNode",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "v1.32.0",
					ControlPlaneKubelet: &kops.KubeletConfigSpec{
						NodeLabels: map[string]string{
							"controlPlane1": "controlPlane1",
							"controlPlane2": "controlPlane2",
						},
					},
					Kubelet: &kops.KubeletConfigSpec{
						NodeLabels: map[string]string{
							"node1": "node1",
							"node2": "node2",
						},
					},
				},
			},
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Role: kops.InstanceGroupRoleNode,
					Kubelet: &kops.KubeletConfigSpec{
						NodeLabels: map[string]string{
							"node1": "override1",
							"node3": "override3",
						},
					},
				},
			},
			expected: map[string]string{
				RoleLabelNode16: "",
				"node2":         "node2",
				"node1":         "override1",
				"node3":         "override3",
			},
		},
		{
			name: "RoleEtcd",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "v1.31.0",
					Kubelet: &kops.KubeletConfigSpec{
						NodeLabels: map[string]string{
							"node1": "node1",
							"node2": "node2",
						},
					},
				},
			},
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Role: kops.InstanceGroupRoleEtcd,
					Kubelet: &kops.KubeletConfigSpec{
						NodeLabels: map[string]string{
							"node1": "override1",
							"node3": "override3",
						},
					},
				},
			},
			expected: map[string]string{
				RoleLabelEtcd: "",
				"node1":       "override1",
				"node3":       "override3",
			},
		},
		{
			name: "RoleScheduler",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "v1.31.0",
					Kubelet: &kops.KubeletConfigSpec{
						NodeLabels: map[string]string{
							"node1": "node1",
							"node2": "node2",
						},
					},
				},
			},
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Role: kops.InstanceGroupRoleScheduler,
					Kubelet: &kops.KubeletConfigSpec{
						NodeLabels: map[string]string{
							"node1": "override1",
							"node3": "override3",
						},
					},
				},
			},
			expected: map[string]string{
				RoleLabelScheduler: "",
				"node1":            "override1",
				"node3":            "override3",
			},
		},
		{
			name: "RoleKubeControllerManager",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "v1.31.0",
					Kubelet: &kops.KubeletConfigSpec{
						NodeLabels: map[string]string{
							"node1": "node1",
							"node2": "node2",
						},
					},
				},
			},
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Role: kops.InstanceGroupRoleKubeControllerManager,
					Kubelet: &kops.KubeletConfigSpec{
						NodeLabels: map[string]string{
							"node1": "override1",
							"node3": "override3",
						},
					},
				},
			},
			expected: map[string]string{
				RoleLabelKubeControllerManager: "",
				"node1":                        "override1",
				"node3":                        "override3",
			},
		},
		{
			// A composite role must produce the labels of every role it carries, not just the
			// first one matched.
			name: "RoleAPIServerAndScheduler",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "v1.31.0",
					Kubelet: &kops.KubeletConfigSpec{
						NodeLabels: map[string]string{
							"node1": "node1",
						},
					},
				},
			},
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Role: "APIServer,Scheduler",
				},
			},
			expected: map[string]string{
				RoleLabelAPIServer16:              "",
				"kops.k8s.io/kops-controller-pki": "",
				RoleLabelScheduler:                "",
			},
			featureFlags: "+APIServerNodes",
		},
		{
			name: "RoleAPIServerAndKubeControllerManager",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "v1.31.0",
					Kubelet: &kops.KubeletConfigSpec{
						NodeLabels: map[string]string{
							"node1": "node1",
						},
					},
				},
			},
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Role: "APIServer,KubeControllerManager",
				},
			},
			expected: map[string]string{
				RoleLabelAPIServer16:              "",
				"kops.k8s.io/kops-controller-pki": "",
				RoleLabelKubeControllerManager:    "",
			},
			featureFlags: "+APIServerNodes",
		},
		{
			// Etcd co-located with the API server it serves.
			name: "RoleAPIServerAndEtcd",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "v1.31.0",
				},
			},
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Role: "APIServer,Etcd",
				},
			},
			expected: map[string]string{
				RoleLabelAPIServer16:              "",
				"kops.k8s.io/kops-controller-pki": "",
				RoleLabelEtcd:                     "",
			},
			featureFlags: "+APIServerNodes",
		},
		{
			name: "RoleUnknown",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "v1.31.0",
				},
			},
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Role: "Nonsense",
				},
			},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.featureFlags != "" {
				featureflag.ParseFlags(test.featureFlags)
				defer func() {
					featureflag.ParseFlags("-APIServerNodes")
				}()
			}
			out, err := BuildNodeLabels(test.cluster, test.ig)
			if test.expectError {
				if err == nil {
					t.Fatalf("expected an error from BuildNodeLabels, got labels %v", out)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error from BuildNodeLabels: %v", err)
			}
			if !reflect.DeepEqual(out, test.expected) {
				t.Fatalf("Test %s\nActual result:\n%v\nExpected result:\n%v", test.name, out, test.expected)
			}
		})
	}
}

// TestAll checks that the managed-label list covers every label BuildNodeLabels and
// kops-channels can set. Anything missing would never be pruned from a node that stops having
// that role.
func TestAll(t *testing.T) {
	all := make(map[string]bool)
	for _, label := range All() {
		if all[label] {
			t.Errorf("All() contains %q twice", label)
		}
		all[label] = true
	}

	expected := []string{
		RoleLabelAPIServer16,
		RoleLabelNode16,
		RoleLabelEtcd,
		RoleLabelScheduler,
		RoleLabelKubeControllerManager,
		RoleLabelControlPlane20,
	}
	for _, label := range expected {
		if !all[label] {
			t.Errorf("All() is missing role label %q", label)
		}
	}
	for component, label := range ClusterComponentLabel {
		if !all[label] {
			t.Errorf("All() is missing %q, the label for component %q", label, component)
		}
	}
}

// TestClusterComponentLabelCoversAllComponents makes sure every component in the API has a label,
// since a component without one could be claimed in spec.hostedComponents and then silently
// never placed.
func TestClusterComponentLabelCoversAllComponents(t *testing.T) {
	for _, component := range kops.AllClusterComponents {
		if _, ok := ClusterComponentLabel[component]; !ok {
			t.Errorf("no node label mapped for cluster component %q", component)
		}
	}
}

func TestChannelsNodeLabels(t *testing.T) {
	ig := func(name string, role kops.InstanceGroupRole, components []kops.ClusterComponent) *kops.InstanceGroup {
		g := &kops.InstanceGroup{}
		g.ObjectMeta.Name = name
		g.Spec.Role = role
		g.Spec.HostedComponents = components
		return g
	}
	gceCluster := func() *kops.Cluster {
		c := &kops.Cluster{}
		c.Spec.CloudProvider.GCE = &kops.GCESpec{}
		return c
	}
	awsCluster := func() *kops.Cluster {
		c := &kops.Cluster{}
		c.Spec.CloudProvider.AWS = &kops.AWSSpec{}
		return c
	}

	grid := []struct {
		name    string
		cluster *kops.Cluster
		groups  []*kops.InstanceGroup
		want    map[string][]string
	}{
		{
			name:    "control plane cluster uses the control-plane label",
			cluster: awsCluster(),
			groups: []*kops.InstanceGroup{
				ig("master", kops.InstanceGroupRoleControlPlane, nil),
				ig("nodes", kops.InstanceGroupRoleNode, nil),
			},
			want: map[string][]string{
				"master": {RoleLabelControlPlane20},
				"nodes":  nil,
			},
		},
		{
			name:    "split control plane claims every component on each API server",
			cluster: awsCluster(),
			groups: []*kops.InstanceGroup{
				ig("apiserver", kops.InstanceGroupRoleAPIServer, nil),
				ig("etcd", kops.InstanceGroupRoleEtcd, nil),
				ig("nodes", kops.InstanceGroupRoleNode, nil),
			},
			want: map[string][]string{
				"apiserver": {RoleLabelKopsCCM, RoleLabelKopsChannel, RoleLabelKopsController, RoleLabelCertManager},
				"etcd":      nil,
				"nodes":     nil,
			},
		},
		{
			name:    "GCE adds the capi-manager label",
			cluster: gceCluster(),
			groups: []*kops.InstanceGroup{
				ig("apiserver", kops.InstanceGroupRoleAPIServer, nil),
				ig("nodes", kops.InstanceGroupRoleNode, nil),
			},
			want: map[string][]string{
				"apiserver": {RoleLabelKopsCCM, RoleLabelKopsChannel, RoleLabelKopsController, RoleLabelCertManager, RoleLabelCAPIManager},
				"nodes":     nil,
			},
		},
		{
			// The case the field exists for: four API servers, only one of which hosts the
			// scheduled components.
			name:    "explicit placement gives the components to one group",
			cluster: gceCluster(),
			groups: []*kops.InstanceGroup{
				ig("external", kops.InstanceGroupRoleAPIServer, nil),
				ig("internal", kops.InstanceGroupRoleAPIServer, []kops.ClusterComponent{
					kops.ClusterComponentKopsChannel,
					kops.ClusterComponentKopsController,
					kops.ClusterComponentCloudControllerManager,
				}),
				ig("scheduler", "APIServer,Scheduler", nil),
				ig("kcm", "APIServer,KubeControllerManager", nil),
				ig("nodes", kops.InstanceGroupRoleNode, nil),
			},
			want: map[string][]string{
				"external": nil,
				// Sorted, not in spec order.
				"internal":  {RoleLabelKopsCCM, RoleLabelKopsChannel, RoleLabelKopsController},
				"scheduler": nil,
				"kcm":       nil,
				"nodes":     nil,
			},
		},
	}

	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			for _, instanceGroup := range g.groups {
				want, ok := g.want[instanceGroup.Name]
				if !ok {
					t.Fatalf("test case does not cover instance group %q", instanceGroup.Name)
				}
				got := ChannelsNodeLabels(g.cluster, g.groups, instanceGroup)
				if !reflect.DeepEqual(got, want) {
					t.Errorf("%s: ChannelsNodeLabels() = %v, want %v", instanceGroup.Name, got, want)
				}
			}
		})
	}
}
