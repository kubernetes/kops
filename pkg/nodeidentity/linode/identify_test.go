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

package linode

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/linode/linodego/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/nodelabels"
)

func TestParseInstanceIDFromProviderID(t *testing.T) {
	tests := []struct {
		name       string
		providerID string
		wantID     int
		wantString string
		wantErr    bool
	}{
		{
			name:       "simple format",
			providerID: "linode://101",
			wantID:     101,
			wantString: "101",
		},
		{
			name:       "triple slash format",
			providerID: "linode:///202",
			wantID:     202,
			wantString: "202",
		},
		{
			name:       "path-like format",
			providerID: "linode://us-ord/303",
			wantID:     303,
			wantString: "303",
		},
		{
			name:       "invalid prefix",
			providerID: "aws:///i-123",
			wantErr:    true,
		},
		{
			name:       "missing id",
			providerID: "linode://",
			wantErr:    true,
		},
		{
			name:       "non numeric id",
			providerID: "linode://abc",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotString, err := parseInstanceIDFromProviderID(tt.providerID)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotID != tt.wantID {
				t.Fatalf("expected id %d, got %d", tt.wantID, gotID)
			}
			if gotString != tt.wantString {
				t.Fatalf("expected id string %q, got %q", tt.wantString, gotString)
			}
		})
	}
}

func TestIdentifyNode(t *testing.T) {
	identifier := &nodeIdentifier{
		client: &fakeLinodeClient{
			instances: map[int]*linodego.Instance{
				101: {
					ID:     101,
					Status: linodego.InstanceRunning,
					Tags: []string{
						"kops.k8s.io/instance-role:Node",
					},
				},
			},
		},
	}

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
		Spec:       corev1.NodeSpec{ProviderID: "linode://101"},
	}

	info, err := identifier.IdentifyNode(context.Background(), node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.InstanceID != "101" {
		t.Fatalf("expected instance id 101, got %q", info.InstanceID)
	}
	if _, ok := info.Labels[nodelabels.RoleLabelNode16]; !ok {
		t.Fatalf("expected node role label %q", nodelabels.RoleLabelNode16)
	}
}

func TestIdentifyNodeStatusCheck(t *testing.T) {
	identifier := &nodeIdentifier{
		client: &fakeLinodeClient{
			instances: map[int]*linodego.Instance{
				77: {
					ID:     77,
					Status: linodego.InstanceOffline,
				},
			},
		},
	}

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-2"},
		Spec:       corev1.NodeSpec{ProviderID: "linode://77"},
	}

	if _, err := identifier.IdentifyNode(context.Background(), node); err == nil {
		t.Fatalf("expected error for offline instance")
	}
}

type fakeLinodeClient struct {
	instances map[int]*linodego.Instance
	err       error
}

func (f *fakeLinodeClient) GetInstance(_ context.Context, linodeID int) (*linodego.Instance, error) {
	if f.err != nil {
		return nil, f.err
	}
	instance, ok := f.instances[linodeID]
	if !ok {
		return nil, errors.New("instance not found")
	}
	return instance, nil
}

func TestBuildLabelsFromTags(t *testing.T) {
	tests := []struct {
		name      string
		tags      []string
		wantLabel string
	}{
		{
			name:      "control plane",
			tags:      []string{"kops.k8s.io/instance-role:ControlPlane"},
			wantLabel: nodelabels.RoleLabelControlPlane20,
		},
		{
			name:      "api server",
			tags:      []string{"kops.k8s.io/instance-role:APIServer"},
			wantLabel: nodelabels.RoleLabelAPIServer16,
		},
		{
			name:      "node",
			tags:      []string{"kops.k8s.io/instance-role:Node"},
			wantLabel: nodelabels.RoleLabelNode16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			labels := buildLabelsFromTags(tt.tags)
			if _, ok := labels[tt.wantLabel]; !ok {
				t.Fatalf("expected label %q from tags %v", tt.wantLabel, tt.tags)
			}
		})
	}
}

func TestBuildLabelsFromTagsUnknownRole(t *testing.T) {
	labels := buildLabelsFromTags([]string{"kops.k8s.io/instance-role:Unknown"})
	if len(labels) != 0 {
		t.Fatalf("expected no labels, got %v", labels)
	}
}

func TestIdentifyNodeProviderIDError(t *testing.T) {
	identifier := &nodeIdentifier{client: &fakeLinodeClient{}}
	node := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node-3"}, Spec: corev1.NodeSpec{ProviderID: "not-linode://1"}}

	_, err := identifier.IdentifyNode(context.Background(), node)
	if err == nil {
		t.Fatalf("expected providerID parse error")
	}
	if got := err.Error(); !strings.Contains(got, "providerID") {
		t.Fatalf("expected providerID context in error, got %q", got)
	}
}
