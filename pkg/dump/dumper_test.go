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
	"os"
	"path/filepath"
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

// fakeLogFilesSSHClient serves a fixed /var/log listing and file contents to the dumper.
type fakeLogFilesSSHClient struct {
	files map[string]string
}

func (fakeLogFilesSSHClient) Close() error { return nil }

func (c fakeLogFilesSSHClient) ExecPiped(ctx context.Context, command string, stdout io.Writer, stderr io.Writer) error {
	if command == "sudo find /var/log -type f -print0" {
		for name := range c.files {
			if _, err := io.WriteString(stdout, name+"\x00"); err != nil {
				return err
			}
		}
		return nil
	}
	for name, contents := range c.files {
		if command == "sudo cat '"+name+"'" {
			_, err := io.WriteString(stdout, contents)
			return err
		}
	}
	return nil
}

// TestDumpCollectsCNIKindnetLog verifies that the node dump copies the cni-kindnet log when the
// node has it, and that a node without it dumps without errors.
func TestDumpCollectsCNIKindnetLog(t *testing.T) {
	grid := []struct {
		name  string
		files map[string]string
	}{
		{
			name:  "present",
			files: map[string]string{"/var/log/cni-kindnet.log": "ADD sandbox\n", "/var/log/syslog": "unrelated\n"},
		},
		{
			name:  "missing",
			files: map[string]string{"/var/log/syslog": "unrelated\n"},
		},
	}
	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			d := NewLogDumper("", nil, nil, t.TempDir(), time.Minute)
			n := &logDumperNode{
				client: fakeLogFilesSSHClient{files: g.files},
				dumper: d,
				dir:    t.TempDir(),
			}
			if errs := n.dump(context.Background()); len(errs) != 0 {
				t.Fatalf("dump: %v", errs)
			}

			got, err := os.ReadFile(filepath.Join(n.dir, "cni-kindnet.log"))
			want, present := g.files["/var/log/cni-kindnet.log"]
			switch {
			case present && err != nil:
				t.Fatalf("reading the dumped cni-kindnet log: %v", err)
			case present && string(got) != want:
				t.Errorf("dumped cni-kindnet log: got %q, want %q", got, want)
			case !present && !os.IsNotExist(err):
				t.Errorf("expected no cni-kindnet log in the dump, got %q, %v", got, err)
			}
		})
	}
}
