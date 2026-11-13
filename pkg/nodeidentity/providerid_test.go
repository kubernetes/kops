/*
Copyright The Kubernetes Authors.

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

package nodeidentity

import (
	"context"
	"errors"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

type identifierFunc func(context.Context, *corev1.Node) (*Info, error)

func (f identifierFunc) IdentifyNode(ctx context.Context, node *corev1.Node) (*Info, error) {
	return f(ctx, node)
}

func TestProviderIDDispatcherSelectsIdentifier(t *testing.T) {
	grid := []struct {
		providerID string
		scheme     string
	}{
		{providerID: "gce://project/zone/instance", scheme: "gce"},
		{providerID: "other:///instance", scheme: "other"},
	}
	for _, g := range grid {
		t.Run(g.scheme, func(t *testing.T) {
			ctx := t.Context()
			node := &corev1.Node{Spec: corev1.NodeSpec{ProviderID: g.providerID}}
			expected := &Info{InstanceID: "instance", Labels: map[string]string{"role": "node"}}
			var calls []string
			newIdentifier := func(scheme string) Identifier {
				return identifierFunc(func(actualCtx context.Context, actualNode *corev1.Node) (*Info, error) {
					calls = append(calls, scheme)
					if actualCtx != ctx || actualNode != node {
						t.Error("context and node were not forwarded unchanged")
					}
					return expected, nil
				})
			}
			dispatcher := NewProviderIDDispatcher(map[string]Identifier{
				"gce":   newIdentifier("gce"),
				"other": newIdentifier("other"),
			})

			actual, err := dispatcher.IdentifyNode(ctx, node)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if actual != expected {
				t.Errorf("expected identifier result %v, got %v", expected, actual)
			}
			if len(calls) != 1 || calls[0] != g.scheme {
				t.Errorf("expected only %q to be called, got %v", g.scheme, calls)
			}
		})
	}
}

func TestProviderIDDispatcherRejectsProviderID(t *testing.T) {
	grid := []struct {
		desc       string
		providerID string
	}{
		{desc: "empty"},
		{desc: "missing delimiter", providerID: "gce:project/zone/instance"},
		{desc: "missing scheme", providerID: "://instance"},
		{desc: "unconfigured provider", providerID: "aws:///zone/instance"},
		{desc: "case-sensitive scheme", providerID: "GCE://project/zone/instance"},
		{desc: "nil identifier", providerID: "disabled://instance"},
	}
	for _, g := range grid {
		t.Run(g.desc, func(t *testing.T) {
			dispatcher := NewProviderIDDispatcher(map[string]Identifier{
				"gce": identifierFunc(func(context.Context, *corev1.Node) (*Info, error) {
					t.Fatal("identifier called for a rejected provider ID")
					return nil, nil
				}),
				"disabled": nil,
			})
			node := &corev1.Node{Spec: corev1.NodeSpec{ProviderID: g.providerID}}
			node.Name = "node"

			actual, err := dispatcher.IdentifyNode(t.Context(), node)
			if err == nil {
				t.Fatal("expected an error")
			}
			if actual != nil {
				t.Errorf("expected no identity, got %v", actual)
			}
		})
	}
}

func TestProviderIDDispatcherPreservesError(t *testing.T) {
	expected := errors.New("provider rejected the node")
	calls := 0
	dispatcher := NewProviderIDDispatcher(map[string]Identifier{
		"gce": identifierFunc(func(context.Context, *corev1.Node) (*Info, error) {
			calls++
			return nil, expected
		}),
		"other": identifierFunc(func(context.Context, *corev1.Node) (*Info, error) {
			t.Error("fell back to another identifier")
			return &Info{}, nil
		}),
	})
	node := &corev1.Node{Spec: corev1.NodeSpec{ProviderID: "gce://project/zone/instance"}}

	actual, err := dispatcher.IdentifyNode(t.Context(), node)
	if err != expected {
		t.Errorf("expected the original error %v, got %v", expected, err)
	}
	if actual != nil {
		t.Errorf("expected no identity, got %v", actual)
	}
	if calls != 1 {
		t.Errorf("expected one call to the selected identifier, got %d", calls)
	}
}
