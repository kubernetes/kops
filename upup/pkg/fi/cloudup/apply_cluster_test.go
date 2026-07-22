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

package cloudup

import (
	"fmt"
	"strings"
	"testing"

	"github.com/blang/semver/v4"
	kopsbase "k8s.io/kops"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model/components"
)

func TestValidateEtcdVersionSupported(t *testing.T) {
	latest36 := components.LatestEtcd36Version
	sv := semver.MustParse(latest36)
	unsupported36 := fmt.Sprintf("%d.%d.%d", sv.Major, sv.Minor, sv.Patch+1)

	tests := []struct {
		name               string
		etcdClusters       []kopsapi.EtcdClusterSpec
		kopsVersionUpdated string
		expectError        bool
		expectErrContains  []string
	}{
		{
			name: "bundled version",
			etcdClusters: []kopsapi.EtcdClusterSpec{
				{Name: "main", Version: latest36},
			},
		},
		{
			name:         "no version recorded yet",
			etcdClusters: []kopsapi.EtcdClusterSpec{{Name: "main"}},
		},
		{
			name: "unsupported version",
			etcdClusters: []kopsapi.EtcdClusterSpec{
				{Name: "main", Version: unsupported36},
			},
			expectError:       true,
			expectErrContains: []string{"main", unsupported36},
		},
		{
			name: "unsupported version with last-updated context",
			etcdClusters: []kopsapi.EtcdClusterSpec{
				{Name: "events", Version: unsupported36},
			},
			kopsVersionUpdated: "1.34.1",
			expectError:        true,
			expectErrContains:  []string{"events", unsupported36},
		},
		{
			name: "unsupported version with custom image is not blocked",
			etcdClusters: []kopsapi.EtcdClusterSpec{
				{Name: "main", Version: unsupported36, Image: "gcr.io/etcd-development/etcd:v" + unsupported36},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &ApplyClusterCmd{
				Cluster: &kopsapi.Cluster{
					Spec: kopsapi.ClusterSpec{
						EtcdClusters: test.etcdClusters,
					},
				},
			}

			err := c.validateEtcdVersionSupported(test.kopsVersionUpdated)
			if test.expectError {
				if err == nil {
					t.Fatalf("expected an error, got none")
				}
				for _, s := range test.expectErrContains {
					if !strings.Contains(err.Error(), s) {
						t.Errorf("expected error to contain %q, got: %v", s, err)
					}
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestValidateEtcdVersionSupported_CurrentVersion sanity-checks that the running kops
// version's own name shows up in the error, since that's the version the operator would
// need to change (e.g. by not downgrading further).
func TestValidateEtcdVersionSupported_CurrentVersion(t *testing.T) {
	c := &ApplyClusterCmd{
		Cluster: &kopsapi.Cluster{
			Spec: kopsapi.ClusterSpec{
				EtcdClusters: []kopsapi.EtcdClusterSpec{
					{Name: "main", Version: "0.0.1"},
				},
			},
		},
	}

	err := c.validateEtcdVersionSupported("")
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if !strings.Contains(err.Error(), kopsbase.Version) {
		t.Errorf("expected error to contain the running kops version %q, got: %v", kopsbase.Version, err)
	}
}
