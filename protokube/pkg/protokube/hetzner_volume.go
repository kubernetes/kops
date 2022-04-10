/*
Copyright 2022 The Kubernetes Authors.

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

package protokube

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/hcloud-go/hcloud/metadata"
	"k8s.io/klog/v2"
	"k8s.io/kops/protokube/pkg/gossip"
	gossiphetzner "k8s.io/kops/protokube/pkg/gossip/hetzner"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
)

// HetznerVolumes defines the Hetzner Cloud volume implementation.
type HetznerVolumes struct {
	hcloudClient *hcloud.Client
	server       *hcloud.Server
}

func (h HetznerVolumes) AttachVolume(volume *Volume) error {
	// TODO(hakman): no longer needed, remove from interface
	panic("implement me")
}

func (h HetznerVolumes) FindVolumes() ([]*Volume, error) {
	// TODO(hakman): no longer needed, remove from interface
	panic("implement me")
}

func (h HetznerVolumes) FindMountedVolume(volume *Volume) (device string, err error) {
	// TODO(hakman): no longer needed, remove from interface
	panic("implement me")
}

var _ Volumes = &HetznerVolumes{}

// NewHetznerVolumes returns a new Hetzner Cloud volume provider.
func NewHetznerVolumes() (*HetznerVolumes, error) {
	serverID, err := metadata.NewClient().InstanceID()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve server id: %s", err)
	}
	klog.V(4).Infof("Found ID of the running server: %d", serverID)

	hcloudToken := os.Getenv("HCLOUD_TOKEN")
	if hcloudToken == "" {
		return nil, fmt.Errorf("%s is required", "HCLOUD_TOKEN")
	}
	opts := []hcloud.ClientOption{
		hcloud.WithToken(hcloudToken),
	}
	hcloudClient := hcloud.NewClient(opts...)

	// TODO(hakman): Get server info from server metadata
	server, _, err := hcloudClient.Server.GetByID(context.TODO(), serverID)
	if err != nil || server == nil {
		return nil, fmt.Errorf("failed to get info for the running server: %s", err)
	}

	klog.V(4).Infof("Found name of the running server: %q", server.Name)
	if server.Datacenter != nil && server.Datacenter.Location != nil {
		klog.V(4).Infof("Found location of the running server: %q", server.Datacenter.Location.Name)
	} else {
		return nil, fmt.Errorf("failed to find location of the running server")
	}
	if len(server.PrivateNet) > 0 {
		klog.V(4).Infof("Found first private net IP of the running server: %q", server.PrivateNet[0].IP.String())
	} else {
		return nil, fmt.Errorf("failed to find private net of the running server")
	}

	h := &HetznerVolumes{
		hcloudClient: hcloudClient,
		server:       server,
	}

	return h, nil
}

func (h HetznerVolumes) InternalIP() (ip net.IP, err error) {
	if len(h.server.PrivateNet) == 0 {
		return nil, fmt.Errorf("failed to find server private ip address")
	}

	klog.V(4).Infof("Found first private IP of the running server: %s", h.server.PrivateNet[0].IP.String())

	return h.server.PrivateNet[0].IP, nil
}

func (h *HetznerVolumes) GossipSeeds() (gossip.SeedProvider, error) {
	clusterName, ok := h.server.Labels[hetzner.TagKubernetesClusterName]
	if !ok {
		return nil, fmt.Errorf("failed to find cluster name label for running server: %v", h.server.Labels)
	}
	return gossiphetzner.NewSeedProvider(h.hcloudClient, clusterName)
}

func (h *HetznerVolumes) InstanceID() string {
	return strconv.Itoa(h.server.ID)
}
