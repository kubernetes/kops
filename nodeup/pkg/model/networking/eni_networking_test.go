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

package networking

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindPhysicalInterfaceByMAC(t *testing.T) {
	// If physical is true, the test makes the "device" entry. The scan uses this entry to know
	// that the interface is physical, not virtual (veths, bridges, VLANs).
	type iface struct {
		name     string
		mac      string
		physical bool
	}

	tests := []struct {
		name        string
		interfaces  []iface
		mac         string
		expected    string
		expectError bool
	}{
		{
			name: "primary ens34 on Graviton4",
			interfaces: []iface{
				{name: "lo", mac: "00:00:00:00:00:00", physical: false},
				{name: "ens34", mac: "02:2b:4d:6e:68:dd", physical: true},
				{name: "ens35", mac: "02:1c:00:1d:07:d3", physical: true},
			},
			mac:      "02:2b:4d:6e:68:dd",
			expected: "ens34",
		},
		{
			name: "case-insensitive match",
			interfaces: []iface{
				{name: "enp39s0", mac: "02:de:80:e0:00:c9", physical: true},
			},
			mac:      "02:DE:80:E0:00:C9",
			expected: "enp39s0",
		},
		{
			name: "virtual interface with cloned MAC is skipped",
			interfaces: []iface{
				{name: "ens5", mac: "02:a0:ac:d3:05:3b", physical: true},
				{name: "vlan5", mac: "02:a0:ac:d3:05:3b", physical: false},
			},
			mac:      "02:a0:ac:d3:05:3b",
			expected: "ens5",
		},
		{
			name: "no match",
			interfaces: []iface{
				{name: "ens5", mac: "02:a0:ac:d3:05:3b", physical: true},
			},
			mac:         "02:ff:ff:ff:ff:ff",
			expectError: true,
		},
		{
			name: "duplicate MAC on physical interfaces",
			interfaces: []iface{
				{name: "ens5", mac: "02:a0:ac:d3:05:3b", physical: true},
				{name: "ens6", mac: "02:a0:ac:d3:05:3b", physical: true},
			},
			mac:         "02:a0:ac:d3:05:3b",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sysClassNet := t.TempDir()
			for _, i := range test.interfaces {
				dir := filepath.Join(sysClassNet, i.name)
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "address"), []byte(i.mac+"\n"), 0o600); err != nil {
					t.Fatal(err)
				}
				if i.physical {
					if err := os.Mkdir(filepath.Join(dir, "device"), 0o755); err != nil {
						t.Fatal(err)
					}
				}
			}

			actual, err := findPhysicalInterfaceByMAC(sysClassNet, test.mac)
			if test.expectError {
				if err == nil {
					t.Fatalf("expected error, got interface %q", actual)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if actual != test.expected {
				t.Fatalf("expected interface %q, got %q", test.expected, actual)
			}
		})
	}
}
