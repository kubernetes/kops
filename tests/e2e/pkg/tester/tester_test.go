/*
Copyright 2021 The Kubernetes Authors.

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

package tester

import (
	"strings"
	"testing"

	"github.com/urfave/sflags/gen/gpflag"
	api "k8s.io/kops/pkg/apis/kops/v1alpha2"
)

func TestFlagParsing(t *testing.T) {
	tester := &Tester{}

	fs, err := gpflag.Parse(tester)
	if err != nil {
		t.Fatalf("gpflag.Parse(tester) failed: %v", err)
	}

	args := []string{"--parallel", "25"}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("fs.Parse(args) failed: %v", err)
	}

	if tester.Parallel != 25 {
		t.Errorf("unexpected value for Parallel; got %d, want %d", tester.Parallel, 25)
	}
}

func TestHasFlag(t *testing.T) {
	grid := []struct {
		Args     string
		Flag     string
		Expected bool
	}{
		{
			Args:     "--provider aws",
			Flag:     "provider",
			Expected: true,
		},
		{
			Args:     "-provider aws",
			Flag:     "provider",
			Expected: true,
		},
		{
			Args:     "provider aws",
			Flag:     "provider",
			Expected: false,
		},
		{
			Args:     "-provider=aws",
			Flag:     "provider",
			Expected: true,
		},
		{
			Args:     "--provider=aws",
			Flag:     "provider",
			Expected: true,
		},
		{
			Args:     "--foo=bar --provider aws",
			Flag:     "provider",
			Expected: true,
		},
		{
			Args:     "--foo=bar",
			Flag:     "provider",
			Expected: false,
		},
	}

	for _, g := range grid {
		t.Run(g.Args, func(t *testing.T) {
			got := hasFlag(g.Args, g.Flag)
			if got != g.Expected {
				t.Errorf("hasFlags(%q, %q) got %v, want %v", g.Args, g.Flag, got, g.Expected)
			}
		})
	}
}

func TestSetSkipRegexFlag(t *testing.T) {
	cluster := &api.Cluster{
		Spec: api.ClusterSpec{
			KubernetesVersion:   "v1.35.0",
			LegacyCloudProvider: "gce",
			LegacyNetworking:    &api.NetworkingSpec{},
			Subnets: []api.ClusterSubnetSpec{
				{
					Name: "us-east1-b",
					Zone: "us-east1-b",
					Type: api.SubnetTypePrivate,
				},
			},
		},
	}

	t.Run("default when empty", func(t *testing.T) {
		tester := NewDefaultTester()
		tester.kopsCluster = cluster
		tester.kopsInstanceGroups = []*api.InstanceGroup{}
		if err := tester.setSkipRegexFlag(); err != nil {
			t.Fatalf("setSkipRegexFlag() error: %v", err)
		}
		if !strings.Contains(tester.SkipRegex, "\\[Slow\\]") {
			t.Errorf("expected default skipRegex to contain \\[Slow\\], got: %s", tester.SkipRegex)
		}
		if !strings.Contains(tester.SkipRegex, "should.be.mountable.when.non-attachable") {
			t.Errorf("expected skipRegex to contain flexvolume skip, got: %s", tester.SkipRegex)
		}
		if !strings.Contains(tester.SkipRegex, "should.SSH.to.all.nodes.and.run.commands") {
			t.Errorf("expected skipRegex to contain ssh skip, got: %s", tester.SkipRegex)
		}
	})

	t.Run("unions caller regex with platform skips", func(t *testing.T) {
		tester := NewDefaultTester()
		tester.kopsCluster = cluster
		tester.kopsInstanceGroups = []*api.InstanceGroup{}
		tester.SkipRegex = "\\[Serial\\]|\\[Disruptive\\]"
		if err := tester.setSkipRegexFlag(); err != nil {
			t.Fatalf("setSkipRegexFlag() error: %v", err)
		}
		if strings.Contains(tester.SkipRegex, "\\[Slow\\]") {
			t.Errorf("expected caller skipRegex to preserve omission of \\[Slow\\], got: %s", tester.SkipRegex)
		}
		if !strings.HasPrefix(tester.SkipRegex, "\\[Serial\\]|\\[Disruptive\\]|") {
			t.Errorf("expected skipRegex to start with caller filter, got: %s", tester.SkipRegex)
		}
		if !strings.Contains(tester.SkipRegex, "should.be.mountable.when.non-attachable") {
			t.Errorf("expected skipRegex to contain flexvolume skip, got: %s", tester.SkipRegex)
		}
		if !strings.Contains(tester.SkipRegex, "should.SSH.to.all.nodes.and.run.commands") {
			t.Errorf("expected skipRegex to contain ssh skip, got: %s", tester.SkipRegex)
		}
	})
}
