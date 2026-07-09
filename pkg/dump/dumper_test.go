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

package dump

import (
	"context"
	"io"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/resources"
)

// fakeSSHClient is a no-op sshClient used to exercise the dumper without a real connection.
type fakeSSHClient struct{}

func (fakeSSHClient) Close() error { return nil }

func (fakeSSHClient) ExecPiped(ctx context.Context, command string, stdout io.Writer, stderr io.Writer) error {
	return nil
}

// fakeSSHClientFactory returns no-op clients and never fails to dial.
type fakeSSHClientFactory struct {
	bastion bool
}

func (f *fakeSSHClientFactory) Dial(ctx context.Context, host string, useBastion bool) (sshClient, error) {
	return fakeSSHClient{}, nil
}

func (f *fakeSSHClientFactory) HasBastion() bool { return f.bastion }

// TestDumpAllNodesGrantsSSHAccess verifies that the SSH access granter is invoked once
// per cloud instance before the dumper connects, covering control-plane, regular, and
// unregistered nodes. This is the hook AWS uses to authorize EC2 Instance Connect.
func TestDumpAllNodesGrantsSSHAccess(t *testing.T) {
	d := &logDumper{
		sshClientFactory: &fakeSSHClientFactory{},
		artifactsDir:     t.TempDir(),
		nodeDumpTimeout:  time.Minute,
	}

	var granted []string
	d.SetSSHAccessGranter(func(ctx context.Context, instanceID string) error {
		granted = append(granted, instanceID)
		return nil
	})

	cloudDump := &resources.Dump{
		Instances: []*resources.Instance{
			{Name: "i-controlplane", PublicAddresses: []string{"203.0.113.1"}},
			{Name: "i-worker", PublicAddresses: []string{"203.0.113.2"}},
			{Name: "i-unregistered", PublicAddresses: []string{"203.0.113.3"}},
		},
	}

	nodes := corev1.NodeList{
		Items: []corev1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "i-controlplane",
					Labels: map[string]string{"node-role.kubernetes.io/control-plane": ""},
				},
				Status: corev1.NodeStatus{Addresses: []corev1.NodeAddress{
					{Type: corev1.NodeExternalIP, Address: "203.0.113.1"},
				}},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "i-worker"},
				Status: corev1.NodeStatus{Addresses: []corev1.NodeAddress{
					{Type: corev1.NodeExternalIP, Address: "203.0.113.2"},
				}},
			},
		},
	}

	if err := d.DumpAllNodes(context.Background(), nodes, 10, cloudDump); err != nil {
		t.Fatalf("DumpAllNodes: %v", err)
	}

	want := map[string]bool{"i-controlplane": true, "i-worker": true, "i-unregistered": true}
	got := map[string]bool{}
	for _, id := range granted {
		got[id] = true
	}
	for id := range want {
		if !got[id] {
			t.Errorf("expected SSH access to be granted for %q; granted=%v", id, granted)
		}
	}
	if len(granted) != len(want) {
		t.Errorf("expected %d grants, got %d: %v", len(want), len(granted), granted)
	}
}

// TestDumpAllNodesNilGranter verifies the dumper works when no granter is set (non-AWS).
func TestDumpAllNodesNilGranter(t *testing.T) {
	d := &logDumper{
		sshClientFactory: &fakeSSHClientFactory{},
		artifactsDir:     t.TempDir(),
		nodeDumpTimeout:  time.Minute,
	}

	cloudDump := &resources.Dump{
		Instances: []*resources.Instance{
			{Name: "i-worker", PublicAddresses: []string{"203.0.113.2"}},
		},
	}
	nodes := corev1.NodeList{
		Items: []corev1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "i-worker"},
				Status: corev1.NodeStatus{Addresses: []corev1.NodeAddress{
					{Type: corev1.NodeExternalIP, Address: "203.0.113.2"},
				}},
			},
		},
	}

	if err := d.DumpAllNodes(context.Background(), nodes, 10, cloudDump); err != nil {
		t.Fatalf("DumpAllNodes with nil granter: %v", err)
	}
}
