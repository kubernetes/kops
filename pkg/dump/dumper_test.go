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
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
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

// failingSSHClientFactory never manages to dial.
type failingSSHClientFactory struct{}

func (failingSSHClientFactory) Dial(ctx context.Context, host string, useBastion bool) (sshClient, error) {
	return nil, errors.New("ssh: unable to authenticate")
}

func (failingSSHClientFactory) HasBastion() bool { return false }

// TestDumpAllNodesFetchesConsoleOutput verifies that unregistered and unreachable nodes get console
// output, and nodes dumped over SSH do not.
func TestDumpAllNodesFetchesConsoleOutput(t *testing.T) {
	registered := corev1.NodeList{
		Items: []corev1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "i-registered"},
				Status: corev1.NodeStatus{Addresses: []corev1.NodeAddress{
					{Type: corev1.NodeExternalIP, Address: "203.0.113.1"},
				}},
			},
		},
	}
	cloudDump := &resources.Dump{
		Instances: []*resources.Instance{
			{Name: "i-registered", PublicAddresses: []string{"203.0.113.1"}},
			{Name: "i-unregistered", PublicAddresses: []string{"203.0.113.2"}},
			{Name: "i-noaddress"},
		},
	}

	grid := []struct {
		name       string
		sshFactory sshClientFactory
		// wantConsoleDirs maps an artifact directory to the instance whose console it holds.
		wantConsoleDirs   map[string]string
		wantNoConsoleDirs []string
	}{
		{
			name:       "ssh succeeds",
			sshFactory: &fakeSSHClientFactory{},
			wantConsoleDirs: map[string]string{
				"203.0.113.2": "i-unregistered",
				"i-noaddress": "i-noaddress",
			},
			wantNoConsoleDirs: []string{"i-registered"},
		},
		{
			name:       "ssh fails",
			sshFactory: failingSSHClientFactory{},
			wantConsoleDirs: map[string]string{
				"i-registered": "i-registered",
				"203.0.113.2":  "i-unregistered",
				"i-noaddress":  "i-noaddress",
			},
		},
	}
	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			artifactsDir := t.TempDir()
			d := &logDumper{
				sshClientFactory: g.sshFactory,
				artifactsDir:     artifactsDir,
				nodeDumpTimeout:  time.Minute,
			}

			var fetched []string
			d.SetConsoleOutputFetcher(func(ctx context.Context, instanceID string) ([]byte, error) {
				fetched = append(fetched, instanceID)
				return []byte("console of " + instanceID), nil
			})

			if err := d.DumpAllNodes(t.Context(), registered, 10, cloudDump); err != nil {
				t.Fatalf("DumpAllNodes: %v", err)
			}

			var wantFetched []string
			for _, instanceID := range g.wantConsoleDirs {
				wantFetched = append(wantFetched, instanceID)
			}
			slices.Sort(fetched)
			slices.Sort(wantFetched)
			if !slices.Equal(fetched, wantFetched) {
				t.Errorf("DumpAllNodes() fetched console output for %v, want %v", fetched, wantFetched)
			}
			for dir, instanceID := range g.wantConsoleDirs {
				p := filepath.Join(artifactsDir, dir, "console-output.log")
				got, err := os.ReadFile(p)
				if err != nil {
					t.Errorf("DumpAllNodes() did not write %s: %v", p, err)
				} else if want := "console of " + instanceID; string(got) != want {
					t.Errorf("DumpAllNodes() wrote %s = %q, want %q", p, got, want)
				}
			}
			for _, dir := range g.wantNoConsoleDirs {
				p := filepath.Join(artifactsDir, dir, "console-output.log")
				if _, err := os.Stat(p); !errors.Is(err, fs.ErrNotExist) {
					t.Errorf("DumpAllNodes() wrote %s, want no console output for a node dumped over SSH (stat err = %v)", p, err)
				}
			}
		})
	}
}

// TestDumpConsoleOutputNothingToWrite verifies the cases that fetch nothing or write no file.
func TestDumpConsoleOutputNothingToWrite(t *testing.T) {
	canceled, cancel := context.WithCancel(t.Context())
	cancel()

	grid := []struct {
		name       string
		ctx        context.Context
		noFetcher  bool
		instanceID string
		output     []byte
		err        error
		dumped     int
		wantFetch  bool
	}{
		{name: "no fetcher", noFetcher: true, instanceID: "i-0123"},
		{name: "empty instance ID"},
		{name: "canceled context", ctx: canceled, instanceID: "i-0123"},
		{name: "fetch error", instanceID: "i-0123", err: errors.New("UnauthorizedOperation"), wantFetch: true},
		{name: "empty output", instanceID: "i-0123", wantFetch: true},
		{name: "limit reached", instanceID: "i-0123", dumped: maxConsoleOutputs},
	}
	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			ctx := g.ctx
			if ctx == nil {
				ctx = t.Context()
			}
			artifactsDir := t.TempDir()
			d := &logDumper{artifactsDir: artifactsDir, consoleOutputsDumped: g.dumped}
			fetched := false
			if !g.noFetcher {
				d.SetConsoleOutputFetcher(func(ctx context.Context, instanceID string) ([]byte, error) {
					fetched = true
					if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > consoleOutputTimeout {
						t.Errorf("fetcher context deadline = %v (set %t), want within %v", deadline, ok, consoleOutputTimeout)
					}
					return g.output, g.err
				})
			}

			d.dumpConsoleOutput(ctx, g.instanceID, "node", true)

			if fetched != g.wantFetch {
				t.Errorf("dumpConsoleOutput(ctx, %q, %q) fetched = %t, want %t", g.instanceID, "node", fetched, g.wantFetch)
			}
			p := filepath.Join(artifactsDir, "node")
			if _, err := os.Stat(p); !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("dumpConsoleOutput(ctx, %q, %q) created %s, want no artifacts (stat err = %v)", g.instanceID, "node", p, err)
			}
		})
	}
}

// hopRefusingSSHClient reaches the bastion, but the hop to the node is refused.
type hopRefusingSSHClient struct {
	closed *int
}

func (c hopRefusingSSHClient) Close() error {
	*c.closed++
	return nil
}

func (hopRefusingSSHClient) ExecPiped(ctx context.Context, command string, stdout io.Writer, stderr io.Writer) error {
	_, _ = io.WriteString(stderr, "Permission denied (publickey).\n")
	return errHopExit
}

// errHopExit matches the text of *ssh.ExitError.
var errHopExit = errors.New("Process exited with status 255")

// hopRefusingSSHClientFactory reaches the bastion but not the nodes behind it.
type hopRefusingSSHClientFactory struct {
	closed *int
}

func (f hopRefusingSSHClientFactory) Dial(ctx context.Context, host string, useBastion bool) (sshClient, error) {
	if !useBastion {
		return nil, errors.New("ssh: unable to authenticate")
	}
	return hopRefusingSSHClient{closed: f.closed}, nil
}

func (hopRefusingSSHClientFactory) HasBastion() bool { return true }

// TestDumpAllNodesBastionHopRefused verifies that a node refusing SSH behind a bastion is treated as
// unreachable.
func TestDumpAllNodesBastionHopRefused(t *testing.T) {
	nodes := corev1.NodeList{
		Items: []corev1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "i-private"},
				Status: corev1.NodeStatus{Addresses: []corev1.NodeAddress{
					{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
				}},
			},
		},
	}
	cloudDump := &resources.Dump{
		Instances: []*resources.Instance{{Name: "i-private", PrivateAddresses: []string{"10.0.0.1"}}},
	}

	grid := []struct {
		name        string
		sshFactory  sshClientFactory
		wantFetched []string
	}{
		{name: "hop refused", sshFactory: hopRefusingSSHClientFactory{closed: new(int)}, wantFetched: []string{"i-private"}},
		{name: "hop succeeds", sshFactory: &fakeSSHClientFactory{bastion: true}},
	}
	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			artifactsDir := t.TempDir()
			d := &logDumper{
				sshClientFactory: g.sshFactory,
				artifactsDir:     artifactsDir,
				nodeDumpTimeout:  time.Minute,
			}
			var fetched []string
			d.SetConsoleOutputFetcher(func(ctx context.Context, instanceID string) ([]byte, error) {
				fetched = append(fetched, instanceID)
				return []byte("console of " + instanceID), nil
			})

			if err := d.DumpAllNodes(t.Context(), nodes, 10, cloudDump); err != nil {
				t.Fatalf("DumpAllNodes: %v", err)
			}

			if !slices.Equal(fetched, g.wantFetched) {
				t.Errorf("DumpAllNodes() fetched console output for %v, want %v", fetched, g.wantFetched)
			}
			if g.wantFetched != nil {
				p := filepath.Join(artifactsDir, "i-private", "console-output.log")
				if _, err := os.Stat(p); err != nil {
					t.Errorf("DumpAllNodes() did not write %s: %v", p, err)
				}
				kern := filepath.Join(artifactsDir, "i-private", "kern.log")
				if _, err := os.Stat(kern); !errors.Is(err, fs.ErrNotExist) {
					t.Errorf("DumpAllNodes() wrote %s for a node whose bastion hop was refused (stat err = %v)", kern, err)
				}
			}
		})
	}
}

// TestConnectToNodeBastionHopRefused verifies that a refused hop closes the bastion client and
// returns the hop's error and stderr.
func TestConnectToNodeBastionHopRefused(t *testing.T) {
	var closed int
	d := &logDumper{sshClientFactory: hopRefusingSSHClientFactory{closed: &closed}, artifactsDir: t.TempDir()}

	_, err := d.connectToNode(t.Context(), "i-private", "10.0.0.1", true)
	if !errors.Is(err, errHopExit) {
		t.Errorf("connectToNode(ctx, %q, %q, true) error = %v, want it to wrap %v", "i-private", "10.0.0.1", err, errHopExit)
	}
	if err == nil || !strings.Contains(err.Error(), "Permission denied (publickey).") {
		t.Errorf("connectToNode(ctx, %q, %q, true) error = %v, want it to include the hop's stderr", "i-private", "10.0.0.1", err)
	}
	if closed != 1 {
		t.Errorf("connectToNode(ctx, %q, %q, true) closed the bastion client %d times, want 1", "i-private", "10.0.0.1", closed)
	}
}

// TestDumpAllNodesConsoleOutputLimits verifies that console captures stop at maxConsoleOutputs and
// continue past the --max-nodes limit.
func TestDumpAllNodesConsoleOutputLimits(t *testing.T) {
	instanceIDs := func(n int) []string {
		var ids []string
		for i := range n {
			ids = append(ids, fmt.Sprintf("i-%02d", i))
		}
		return ids
	}
	unregistered := func(n int) *resources.Dump {
		dump := &resources.Dump{}
		for i, id := range instanceIDs(n) {
			dump.Instances = append(dump.Instances, &resources.Instance{
				Name:            id,
				PublicAddresses: []string{fmt.Sprintf("203.0.113.%d", i+1)},
			})
		}
		return dump
	}

	grid := []struct {
		name           string
		sshFactory     sshClientFactory
		maxNodesToDump int
		instances      int
		fetchErr       error
		wantFetched    []string
		// Artifact directories that must, and must not, hold console-output.log.
		wantConsoleDirs   []string
		wantNoConsoleDirs []string
	}{
		{
			name:              "ssh fails everywhere",
			sshFactory:        failingSSHClientFactory{},
			maxNodesToDump:    500,
			instances:         maxConsoleOutputs + 5,
			wantFetched:       instanceIDs(maxConsoleOutputs),
			wantConsoleDirs:   []string{fmt.Sprintf("203.0.113.%d", maxConsoleOutputs)},
			wantNoConsoleDirs: []string{fmt.Sprintf("203.0.113.%d", maxConsoleOutputs+1)},
		},
		{
			name:            "ssh succeeds everywhere",
			sshFactory:      &fakeSSHClientFactory{},
			maxNodesToDump:  500,
			instances:       maxConsoleOutputs + 5,
			wantFetched:     instanceIDs(maxConsoleOutputs + 5),
			wantConsoleDirs: []string{fmt.Sprintf("203.0.113.%d", maxConsoleOutputs+5)},
		},
		{
			name:           "failed fetches count toward the limit",
			sshFactory:     failingSSHClientFactory{},
			maxNodesToDump: 500,
			instances:      maxConsoleOutputs + 5,
			fetchErr:       errors.New("UnauthorizedOperation"),
			wantFetched:    instanceIDs(maxConsoleOutputs),
		},
		{
			name:           "limit applies past the max-nodes limit",
			sshFactory:     &fakeSSHClientFactory{},
			maxNodesToDump: 1,
			instances:      maxConsoleOutputs + 3,
			// i-00 is dumped over SSH and uncapped; the console-only captures stop at the limit.
			wantFetched:       instanceIDs(maxConsoleOutputs + 1),
			wantNoConsoleDirs: []string{fmt.Sprintf("i-%02d", maxConsoleOutputs+1)},
		},
		{
			name:            "past the max-nodes limit",
			sshFactory:      &fakeSSHClientFactory{},
			maxNodesToDump:  1,
			instances:       3,
			wantFetched:     instanceIDs(3),
			wantConsoleDirs: []string{"203.0.113.1", "i-01", "i-02"},
		},
	}
	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			artifactsDir := t.TempDir()
			d := &logDumper{
				sshClientFactory: g.sshFactory,
				artifactsDir:     artifactsDir,
				nodeDumpTimeout:  time.Minute,
			}
			var fetched []string
			d.SetConsoleOutputFetcher(func(ctx context.Context, instanceID string) ([]byte, error) {
				fetched = append(fetched, instanceID)
				if g.fetchErr != nil {
					return nil, g.fetchErr
				}
				return []byte("console of " + instanceID), nil
			})

			if err := d.DumpAllNodes(t.Context(), corev1.NodeList{}, g.maxNodesToDump, unregistered(g.instances)); err != nil {
				t.Fatalf("DumpAllNodes: %v", err)
			}

			if !slices.Equal(fetched, g.wantFetched) {
				t.Errorf("DumpAllNodes(maxNodesToDump=%d) with %d instances fetched console output for %v, want %v", g.maxNodesToDump, g.instances, fetched, g.wantFetched)
			}
			for _, dir := range g.wantConsoleDirs {
				p := filepath.Join(artifactsDir, dir, "console-output.log")
				if _, err := os.Stat(p); err != nil {
					t.Errorf("DumpAllNodes() did not write %s: %v", p, err)
				}
			}
			for _, dir := range g.wantNoConsoleDirs {
				p := filepath.Join(artifactsDir, dir, "console-output.log")
				if _, err := os.Stat(p); !errors.Is(err, fs.ErrNotExist) {
					t.Errorf("DumpAllNodes() wrote %s past the console output limit (stat err = %v)", p, err)
				}
			}
		})
	}
}
