/*
Copyright The Kubernetes Authors.

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

package hetznertasks

import (
	"net"
	"reflect"
	"testing"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
)

const (
	testUserDataHash = "sha256.newhash"
	testOldHash      = "sha256.oldhash"
	testLocation     = "fsn1"
	testSize         = "cx21"
	testImage        = "ubuntu-22.04"
	testNodeGroup    = "nodes"
)

// baseGroup returns a ServerGroup with a template that a matching server satisfies.
func baseGroup(count int) *ServerGroup {
	return &ServerGroup{
		Count:      count,
		Location:   testLocation,
		Size:       testSize,
		Image:      testImage,
		EnableIPv4: true,
		EnableIPv6: false,
		Labels: map[string]string{
			hetzner.TagClusterAutoscalerNodeGroup: testNodeGroup,
			hetzner.TagKubernetesInstanceUserData: testUserDataHash,
		},
	}
}

// matchingServer returns a server that fully satisfies baseGroup's template.
func matchingServer(name string) *hcloud.Server {
	return &hcloud.Server{
		Name: name,
		Labels: map[string]string{
			hetzner.TagClusterAutoscalerNodeGroup: testNodeGroup,
			hetzner.TagKubernetesInstanceUserData: testUserDataHash,
		},
		Datacenter: &hcloud.Datacenter{
			Location: &hcloud.Location{Name: testLocation},
		},
		ServerType: &hcloud.ServerType{Name: testSize},
		Image:      &hcloud.Image{Name: testImage},
		PublicNet: hcloud.ServerPublicNet{
			IPv4: hcloud.ServerPublicNetIPv4{IP: net.ParseIP("1.2.3.4")},
		},
	}
}

func TestClassifyServers(t *testing.T) {
	tests := []struct {
		name           string
		group          *ServerGroup
		servers        []*hcloud.Server
		wantNeedUpdate []string
		wantCount      int
	}{
		{
			name:      "matching server not marked",
			group:     baseGroup(1),
			servers:   []*hcloud.Server{matchingServer("s1")},
			wantCount: 1,
		},
		{
			name:  "wrong userdata hash marks needUpdate",
			group: baseGroup(1),
			servers: func() []*hcloud.Server {
				s := matchingServer("s1")
				s.Labels[hetzner.TagKubernetesInstanceUserData] = testOldHash
				return []*hcloud.Server{s}
			}(),
			wantNeedUpdate: []string{"s1"},
			wantCount:      1,
		},
		{
			name:  "wrong location marks needUpdate",
			group: baseGroup(1),
			servers: func() []*hcloud.Server {
				s := matchingServer("s1")
				s.Datacenter.Location.Name = "nbg1"
				return []*hcloud.Server{s}
			}(),
			wantNeedUpdate: []string{"s1"},
			wantCount:      1,
		},
		{
			name:  "nil datacenter marks needUpdate",
			group: baseGroup(1),
			servers: func() []*hcloud.Server {
				s := matchingServer("s1")
				s.Datacenter = nil
				return []*hcloud.Server{s}
			}(),
			wantNeedUpdate: []string{"s1"},
			wantCount:      1,
		},
		{
			name:  "wrong server type marks needUpdate",
			group: baseGroup(1),
			servers: func() []*hcloud.Server {
				s := matchingServer("s1")
				s.ServerType.Name = "cx31"
				return []*hcloud.Server{s}
			}(),
			wantNeedUpdate: []string{"s1"},
			wantCount:      1,
		},
		{
			name:  "nil server type marks needUpdate",
			group: baseGroup(1),
			servers: func() []*hcloud.Server {
				s := matchingServer("s1")
				s.ServerType = nil
				return []*hcloud.Server{s}
			}(),
			wantNeedUpdate: []string{"s1"},
			wantCount:      1,
		},
		{
			name:  "wrong image marks needUpdate",
			group: baseGroup(1),
			servers: func() []*hcloud.Server {
				s := matchingServer("s1")
				s.Image.Name = "debian-12"
				return []*hcloud.Server{s}
			}(),
			wantNeedUpdate: []string{"s1"},
			wantCount:      1,
		},
		{
			name:  "nil image marks needUpdate",
			group: baseGroup(1),
			servers: func() []*hcloud.Server {
				s := matchingServer("s1")
				s.Image = nil
				return []*hcloud.Server{s}
			}(),
			wantNeedUpdate: []string{"s1"},
			wantCount:      1,
		},
		{
			name:  "IPv4 presence mismatch marks needUpdate",
			group: baseGroup(1),
			servers: func() []*hcloud.Server {
				s := matchingServer("s1")
				s.PublicNet.IPv4.IP = nil
				return []*hcloud.Server{s}
			}(),
			wantNeedUpdate: []string{"s1"},
			wantCount:      1,
		},
		{
			name:  "IPv6 presence mismatch marks needUpdate",
			group: baseGroup(1),
			servers: func() []*hcloud.Server {
				s := matchingServer("s1")
				s.PublicNet.IPv6.IP = net.ParseIP("2001:db8::1")
				return []*hcloud.Server{s}
			}(),
			wantNeedUpdate: []string{"s1"},
			wantCount:      1,
		},
		{
			name:  "already needs-update server is not re-marked",
			group: baseGroup(1),
			servers: func() []*hcloud.Server {
				// Mismatched (wrong image) but already labeled needs-update.
				s := matchingServer("s1")
				s.Image.Name = "debian-12"
				s.Labels[hetzner.TagKubernetesInstanceNeedsUpdate] = ""
				return []*hcloud.Server{s}
			}(),
			wantCount: 1,
		},
		{
			name:  "no index-based marking with more matching servers than count",
			group: baseGroup(1),
			servers: []*hcloud.Server{
				matchingServer("s1"),
				matchingServer("s2"),
				matchingServer("s3"),
			},
			wantCount: 1,
		},
		{
			name:      "count clamp 1 server count 3",
			group:     baseGroup(3),
			servers:   []*hcloud.Server{matchingServer("s1")},
			wantCount: 1,
		},
		{
			name:  "count clamp 5 servers count 2",
			group: baseGroup(2),
			servers: []*hcloud.Server{
				matchingServer("s1"),
				matchingServer("s2"),
				matchingServer("s3"),
				matchingServer("s4"),
				matchingServer("s5"),
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			needUpdate, actualCount := tt.group.classifyServers(tt.servers, testUserDataHash)

			if !reflect.DeepEqual(needUpdate, tt.wantNeedUpdate) {
				t.Errorf("needUpdate = %v, want %v", needUpdate, tt.wantNeedUpdate)
			}
			if actualCount != tt.wantCount {
				t.Errorf("actualCount = %d, want %d", actualCount, tt.wantCount)
			}
		})
	}
}
