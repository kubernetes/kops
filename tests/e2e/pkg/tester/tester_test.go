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
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/urfave/sflags/gen/gpflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "k8s.io/kops/pkg/apis/kops/v1alpha2"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

func TestConfigureBastionFlag(t *testing.T) {
	for _, test := range []struct {
		name string
		args []string
		want bool
	}{
		{name: "default"},
		{name: "enabled", args: []string{"--configure-bastion"}, want: true},
		{name: "disabled", args: []string{"--configure-bastion=false"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			tester := NewDefaultTester()
			flags, err := gpflag.Parse(tester)
			if err != nil {
				t.Fatal(err)
			}
			if err := flags.Parse(test.args); err != nil {
				t.Fatal(err)
			}
			if tester.ConfigureBastion != test.want {
				t.Errorf("ConfigureBastion = %t, want %t", tester.ConfigureBastion, test.want)
			}
		})
	}
}

func TestSetKubeBastion(t *testing.T) {
	for _, provider := range []string{"aws", "gce"} {
		t.Run(provider, func(t *testing.T) {
			clusterName := "test.k8s.local"
			groupName := "control-plane"
			instanceGroupName := groupName
			if provider == "gce" {
				instanceGroupName = gce.NameForInstanceGroupManager(clusterName, groupName, "us-central1-b")
			}
			output, err := json.Marshal([]map[string]string{
				{"instanceGroup": "nodes", "externalIP": "192.0.2.1"},
				{"instanceGroup": instanceGroupName, "externalIP": "192.0.2.2"},
				{"instanceGroup": instanceGroupName, "externalIP": "192.0.2.3"},
			})
			if err != nil {
				t.Fatal(err)
			}
			binDir := t.TempDir()
			script := "#!/bin/sh\n[ \"$*\" = 'get instances --name test.k8s.local -ojson' ] || exit 1\nprintf '%s' \"$KOPS_TEST_INSTANCES\"\n"
			if err := os.WriteFile(filepath.Join(binDir, "kops"), []byte(script), 0755); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("KOPS_TEST_INSTANCES", string(output))
			t.Setenv("KUBE_SSH_BASTION", "")
			tester := &Tester{
				kopsCluster: &api.Cluster{
					ObjectMeta: metav1.ObjectMeta{Name: clusterName},
					Spec:       api.ClusterSpec{LegacyCloudProvider: provider},
				},
				kopsInstanceGroups: []*api.InstanceGroup{
					{
						ObjectMeta: metav1.ObjectMeta{Name: groupName},
						Spec: api.InstanceGroupSpec{
							Role:  "Master",
							Zones: []string{"us-central1-a", "us-central1-b"},
						},
					},
				},
			}
			if err := tester.setKubeBastion(); err != nil {
				t.Fatal(err)
			}
			if got := os.Getenv("KUBE_SSH_BASTION"); got != "192.0.2.2" {
				t.Errorf("KUBE_SSH_BASTION = %q, want 192.0.2.2", got)
			}
		})
	}
}

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
