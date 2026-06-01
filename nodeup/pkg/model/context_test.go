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

package model

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// TestNodeupContextInitFlags verifies the IsMaster / HasAPIServer / HostsEtcd flags
// are populated correctly across the role + etcd-membership combinations relevant to
// split control-plane topologies.
func TestNodeupContextInitFlags(t *testing.T) {
	cases := []struct {
		name          string
		role          kops.InstanceGroupRole
		etcdManifests []string
		wantMaster    bool
		wantAPIServer bool
		wantHostsEtcd bool
	}{
		{
			name:          "ControlPlane with etcd member",
			role:          kops.InstanceGroupRoleControlPlane,
			etcdManifests: []string{"manifests/etcd/main-master-a.yaml"},
			wantMaster:    true,
			wantAPIServer: true,
			wantHostsEtcd: true,
		},
		{
			name:          "ControlPlane without etcd member",
			role:          kops.InstanceGroupRoleControlPlane,
			etcdManifests: nil,
			wantMaster:    true,
			wantAPIServer: true,
			wantHostsEtcd: false,
		},
		{
			name:          "APIServer frontend hosting etcd",
			role:          kops.InstanceGroupRoleAPIServer,
			etcdManifests: []string{"manifests/etcd/main-apiserver-a.yaml"},
			wantMaster:    false,
			wantAPIServer: true,
			wantHostsEtcd: true,
		},
		{
			name:          "APIServer frontend without etcd",
			role:          kops.InstanceGroupRoleAPIServer,
			etcdManifests: nil,
			wantMaster:    false,
			wantAPIServer: true,
			wantHostsEtcd: false,
		},
		{
			name:          "Worker node",
			role:          kops.InstanceGroupRoleNode,
			etcdManifests: nil,
			wantMaster:    false,
			wantAPIServer: false,
			wantHostsEtcd: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := &NodeupModelContext{
				BootConfig: &nodeup.BootConfig{
					InstanceGroupRole: tc.role,
				},
				NodeupConfig: &nodeup.Config{
					KubernetesVersion: "1.32.0",
					EtcdManifests:     tc.etcdManifests,
				},
			}
			c.bootstrapCerts = map[string]*nodetasks.BootstrapCert{}
			c.bootstrapKeypairIDs = map[string]string{}
			if err := c.Init(); err != nil {
				t.Fatalf("Init returned error: %v", err)
			}
			if c.IsMaster != tc.wantMaster {
				t.Errorf("IsMaster = %v, want %v", c.IsMaster, tc.wantMaster)
			}
			if c.HasAPIServer != tc.wantAPIServer {
				t.Errorf("HasAPIServer = %v, want %v", c.HasAPIServer, tc.wantAPIServer)
			}
			if c.HostsEtcd != tc.wantHostsEtcd {
				t.Errorf("HostsEtcd = %v, want %v", c.HostsEtcd, tc.wantHostsEtcd)
			}
		})
	}
}
