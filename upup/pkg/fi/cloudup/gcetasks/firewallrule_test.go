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

package gcetasks

import (
	"bytes"
	"slices"
	"strings"
	"testing"

	"google.golang.org/api/compute/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	gcemock "k8s.io/kops/cloudmock/gce"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

func TestFirewallRuleReorderedAllowedAndSourceRanges(t *testing.T) {
	for _, tt := range []struct {
		name       string
		allowed    sets.Set[string]
		ranges     sets.Set[string]
		wantChange bool
	}{
		{name: "reordered"},
		{name: "replace allowed", allowed: sets.New("tcp:80", "udp:53"), wantChange: true},
		{name: "replace source range", ranges: sets.New("10.0.0.0/8", "192.0.2.0/24"), wantChange: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cloud := gcemock.InstallMockGCECloud("us-test1", "project")
			_, err := cloud.Compute().Firewalls().Insert("project", &compute.Firewall{
				Name:         "test-firewall",
				Network:      "projects/project/global/networks/test-network",
				Allowed:      []*compute.FirewallAllowed{{IPProtocol: "udp", Ports: []string{"53"}}, {IPProtocol: "tcp", Ports: []string{"443"}}},
				SourceRanges: []string{"172.16.0.0/12", "10.0.0.0/8"},
			})
			if err != nil {
				t.Fatal(err)
			}
			expected := &FirewallRule{
				Name:         new("test-firewall"),
				Family:       AddressFamilyIPv4,
				Network:      &Network{Name: new("test-network")},
				Allowed:      sets.New("tcp:443", "udp:53"),
				SourceRanges: sets.New("10.0.0.0/8", "172.16.0.0/12"),
			}
			if tt.allowed != nil {
				expected.Allowed = tt.allowed
			}
			if tt.ranges != nil {
				expected.SourceRanges = tt.ranges
			}
			ctx := &fi.CloudupContext{T: fi.CloudupSubContext{Cloud: cloud}}
			if err := expected.Normalize(ctx); err != nil {
				t.Fatal(err)
			}
			actual, err := expected.Find(ctx)
			if err != nil {
				t.Fatal(err)
			}
			changes := &FirewallRule{}
			if got := fi.BuildChanges(actual, expected, changes); got != tt.wantChange {
				t.Fatalf("BuildChanges = %v, want %v", got, tt.wantChange)
			}
			if tt.wantChange {
				if err := expected.RenderGCE(gce.NewGCEAPITarget(cloud), actual, expected, changes); err != nil {
					t.Fatal(err)
				}
				actual, err = expected.Find(ctx)
				if err != nil {
					t.Fatal(err)
				}
				if fi.BuildChanges(actual, expected, &FirewallRule{}) {
					t.Error("unexpected changes after applying firewall update")
				}
			}
		})
	}
}

func TestFirewallRuleTags(t *testing.T) {
	for _, field := range []string{"SourceTags", "TargetTags"} {
		for _, tt := range []struct {
			name         string
			actualTags   []string
			expectedTags []string
			wantChange   bool
		}{
			{name: "reordered", actualTags: []string{"nodes", "masters"}, expectedTags: []string{"masters", "nodes"}},
			{name: "add", actualTags: []string{"masters"}, expectedTags: []string{"nodes", "masters"}, wantChange: true},
			{name: "remove", actualTags: []string{"masters", "nodes"}, expectedTags: []string{"nodes"}, wantChange: true},
			{name: "replace", actualTags: []string{"masters"}, expectedTags: []string{"nodes"}, wantChange: true},
			{name: "clear", actualTags: []string{"masters"}, wantChange: true},
			{name: "absent"},
			{name: "empty", actualTags: []string{}},
		} {
			t.Run(field+"/"+tt.name, func(t *testing.T) {
				cloud := gcemock.InstallMockGCECloud("us-test1", "project")
				firewall := &compute.Firewall{
					Name:     "test-firewall",
					Network:  "projects/project/global/networks/test-network",
					Allowed:  []*compute.FirewallAllowed{{IPProtocol: "tcp", Ports: []string{"443"}}},
					Disabled: true,
				}
				expected := &FirewallRule{
					Name:     new("test-firewall"),
					Network:  &Network{Name: new("test-network")},
					Allowed:  sets.New("tcp:443"),
					Disabled: true,
				}
				if field == "SourceTags" {
					firewall.SourceTags = tt.actualTags
					if len(tt.expectedTags) > 0 {
						expected.SourceTags = sets.New(tt.expectedTags...)
					}
				} else {
					firewall.TargetTags = tt.actualTags
					if len(tt.expectedTags) > 0 {
						expected.TargetTags = sets.New(tt.expectedTags...)
					}
				}
				if _, err := cloud.Compute().Firewalls().Insert("project", firewall); err != nil {
					t.Fatal(err)
				}
				ctx := &fi.CloudupContext{T: fi.CloudupSubContext{Cloud: cloud}}
				actual, err := expected.Find(ctx)
				if err != nil {
					t.Fatal(err)
				}
				if got := fi.BuildChanges(actual, expected, &FirewallRule{}); got != tt.wantChange {
					t.Errorf("BuildChanges = %v, want %v", got, tt.wantChange)
				}

				// A port update must not introduce a spurious tag change in the preview.
				expected.Allowed.Insert("tcp:10250")
				changes := &FirewallRule{}
				fi.BuildChanges(actual, expected, changes)
				var out bytes.Buffer
				target := fi.NewCloudupDryRunTarget(&assets.AssetBuilder{}, true, &out)
				if err := target.Render(actual, expected, changes); err != nil {
					t.Fatal(err)
				}
				if err := target.PrintReport(map[string]fi.CloudupTask{"FirewallRule/test-firewall": expected}, &out); err != nil {
					t.Fatal(err)
				}
				if strings.Contains(out.String(), field) != tt.wantChange {
					t.Errorf("unexpected tag change in report:\n%s", out.String())
				}

				if err := expected.RenderGCE(gce.NewGCEAPITarget(cloud), actual, expected, changes); err != nil {
					t.Fatal(err)
				}
				rendered, err := cloud.Compute().Firewalls().Get("project", *expected.Name)
				if err != nil {
					t.Fatal(err)
				}
				tags := rendered.SourceTags
				if field == "TargetTags" {
					tags = rendered.TargetTags
				}
				wantTags := slices.Clone(tt.expectedTags)
				slices.Sort(wantTags)
				if !slices.Equal(tags, wantTags) {
					t.Errorf("rendered tags = %v, want %v", tags, wantTags)
				}
				actual, err = expected.Find(ctx)
				if err != nil {
					t.Fatal(err)
				}
				if fi.BuildChanges(actual, expected, &FirewallRule{}) {
					t.Error("unexpected changes after applying firewall update")
				}
			})
		}
	}
}

func TestFirewallRuleDryRunReportWithSourceTags(t *testing.T) {
	for _, tt := range []struct {
		name         string
		sourceRanges []string
	}{
		{name: "absent source ranges"},
		{name: "empty source ranges", sourceRanges: []string{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cloud := gcemock.InstallMockGCECloud("us-test1", "project")
			_, err := cloud.Compute().Firewalls().Insert("project", &compute.Firewall{
				Name:         "test-firewall",
				Network:      "projects/project/global/networks/test-network",
				SourceTags:   []string{"nodes"},
				SourceRanges: tt.sourceRanges,
				Allowed:      []*compute.FirewallAllowed{{IPProtocol: "tcp", Ports: []string{"443"}}},
			})
			if err != nil {
				t.Fatal(err)
			}
			expected := &FirewallRule{
				Name:       new("test-firewall"),
				Network:    &Network{Name: new("test-network")},
				SourceTags: sets.New("nodes"),
				Allowed:    sets.New("tcp:443", "tcp:10250"),
			}
			actual, err := expected.Find(&fi.CloudupContext{T: fi.CloudupSubContext{Cloud: cloud}})
			if err != nil {
				t.Fatal(err)
			}
			changes := &FirewallRule{}
			if !fi.BuildChanges(actual, expected, changes) {
				t.Fatal("expected firewall changes")
			}
			var out bytes.Buffer
			target := fi.NewCloudupDryRunTarget(&assets.AssetBuilder{}, true, &out)
			if err := target.Render(actual, expected, changes); err != nil {
				t.Fatal(err)
			}
			tasks := map[string]fi.CloudupTask{"FirewallRule/test-firewall": expected}
			if err := target.PrintReport(tasks, &out); err != nil {
				t.Fatal(err)
			}
			if strings.Contains(out.String(), "SourceRanges") {
				t.Errorf("unexpected source range change in report:\n%s", out.String())
			}
			if !strings.Contains(out.String(), "[tcp:443] -> [tcp:10250, tcp:443]") {
				t.Errorf("expected allowed port change in report:\n%s", out.String())
			}
		})
	}
}

func TestFirewallRuleDryRunReport(t *testing.T) {
	expected := &FirewallRule{
		Name:         new("test-firewall"),
		Network:      &Network{Name: new("test-network")},
		Allowed:      sets.New("udp:53", "tcp:443"),
		SourceRanges: sets.New("172.16.0.0/12", "10.0.0.0/8"),
	}
	actual := *expected
	actual.Allowed = sets.New("tcp:80")
	actual.SourceRanges = sets.New("192.0.2.0/24")

	for _, tt := range []struct {
		name   string
		actual *FirewallRule
		lines  []string
	}{
		{
			name: "create",
			lines: []string{
				"Will create resources:",
				"[tcp:443, udp:53]",
				"[10.0.0.0/8, 172.16.0.0/12]",
			},
		},
		{
			name:   "update",
			actual: &actual,
			lines: []string{
				"Will modify resources:",
				"[tcp:80] -> [tcp:443, udp:53]",
				"[192.0.2.0/24] -> [10.0.0.0/8, 172.16.0.0/12]",
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			changes := &FirewallRule{}
			if !fi.BuildChanges(tt.actual, expected, changes) {
				t.Fatal("expected firewall changes")
			}
			var out bytes.Buffer
			target := fi.NewCloudupDryRunTarget(&assets.AssetBuilder{}, true, &out)
			if err := target.Render(tt.actual, expected, changes); err != nil {
				t.Fatal(err)
			}
			tasks := map[string]fi.CloudupTask{"FirewallRule/test-firewall": expected}
			if err := target.PrintReport(tasks, &out); err != nil {
				t.Fatal(err)
			}
			for _, line := range tt.lines {
				if !strings.Contains(out.String(), line) {
					t.Errorf("expected report to contain %q, got:\n%s", line, out.String())
				}
			}
		})
	}
}
