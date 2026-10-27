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
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/kops/pkg/nodeidentity"
)

type recordingNodeIdentifier struct {
	calls int
}

func (i *recordingNodeIdentifier) IdentifyNode(_ context.Context, node *corev1.Node) (*nodeidentity.Info, error) {
	i.calls++
	return &nodeidentity.Info{InstanceID: node.Spec.ProviderID}, nil
}

func TestWithProviderIDDispatch(t *testing.T) {
	grid := []struct {
		cloud       string
		providerIDs []string
	}{
		{
			cloud:       "aws",
			providerIDs: []string{"aws:///zone/i-instance"},
		},
		{
			cloud:       "gce",
			providerIDs: []string{"gce://project/zone/instance"},
		},
		{
			cloud: "azure",
			providerIDs: []string{
				"azure:///subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm",
				"azure:///subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachineScaleSets/vmss/virtualMachines/0",
			},
		},
		{
			cloud:       "digitalocean",
			providerIDs: []string{"digitalocean://123"},
		},
		{
			cloud:       "hetzner",
			providerIDs: []string{"hcloud://123"},
		},
		{
			cloud:       "openstack",
			providerIDs: []string{"openstack://instance", "openstack:///instance"},
		},
		{
			cloud:       "scaleway",
			providerIDs: []string{"scaleway://instance"},
		},
		{
			cloud:       "linode",
			providerIDs: []string{"linode://101", "linode:///202", "linode://us-ord/303"},
		},
	}
	for _, g := range grid {
		t.Run(g.cloud, func(t *testing.T) {
			identifier := &recordingNodeIdentifier{}
			wrapped := withProviderIDDispatch(g.cloud, identifier)
			for _, providerID := range g.providerIDs {
				node := &corev1.Node{Spec: corev1.NodeSpec{ProviderID: providerID}}
				actual, err := wrapped.IdentifyNode(t.Context(), node)
				if err != nil {
					t.Fatalf("identifying %q: %v", providerID, err)
				}
				if actual == nil || actual.InstanceID != providerID {
					t.Errorf("expected provider ID %q to be forwarded unchanged, got %v", providerID, actual)
				}
			}

			otherProviderID := "gce://project/zone/instance"
			if g.cloud == "gce" {
				otherProviderID = "aws:///zone/i-instance"
			}
			node := &corev1.Node{Spec: corev1.NodeSpec{ProviderID: otherProviderID}}
			actual, err := wrapped.IdentifyNode(t.Context(), node)
			if err == nil {
				t.Error("expected an error for an unconfigured provider")
			}
			if actual != nil {
				t.Errorf("expected no identity for an unconfigured provider, got %v", actual)
			}
			if identifier.calls != len(g.providerIDs) {
				t.Errorf("expected %d identifier calls, got %d", len(g.providerIDs), identifier.calls)
			}
		})
	}
}

func TestWithProviderIDDispatchMetal(t *testing.T) {
	identifier := &recordingNodeIdentifier{}
	if actual := withProviderIDDispatch("metal", identifier); actual != identifier {
		t.Fatal("expected the Metal identifier to remain unchanged")
	}
}
