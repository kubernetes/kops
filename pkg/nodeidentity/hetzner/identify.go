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

package hetzner

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hetznercloud/hcloud-go/hcloud"
	corev1 "k8s.io/api/core/v1"
	expirationcache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/nodeidentity"
	"k8s.io/kops/pkg/nodelabels"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
)

const (
	cacheTTL = 60 * time.Minute
)

// nodeIdentifier identifies a node from Hetzner Cloud
type nodeIdentifier struct {
	client       *hcloud.Client
	cache        expirationcache.Store
	cacheEnabled bool
}

// New creates and returns a nodeidentity.Identifier for Nodes running on Hetzner Cloud
func New(CacheNodeidentityInfo bool) (nodeidentity.Identifier, error) {
	hcloudToken := os.Getenv("HCLOUD_TOKEN")
	if hcloudToken == "" {
		return nil, fmt.Errorf("%s is required", "HCLOUD_TOKEN")
	}
	opts := []hcloud.ClientOption{
		hcloud.WithToken(hcloudToken),
	}
	hcloudClient := hcloud.NewClient(opts...)

	return &nodeIdentifier{
		client:       hcloudClient,
		cache:        expirationcache.NewTTLStore(stringKeyFunc, cacheTTL),
		cacheEnabled: CacheNodeidentityInfo,
	}, nil
}

// IdentifyNode queries Hetzner Cloud for the node identity information
func (i *nodeIdentifier) IdentifyNode(ctx context.Context, node *corev1.Node) (*nodeidentity.Info, error) {
	providerID := node.Spec.ProviderID
	if providerID == "" {
		return nil, fmt.Errorf("providerID not set for node %s", node.Name)
	}
	if !strings.HasPrefix(providerID, "hcloud://") {
		return nil, fmt.Errorf("providerID %q not recognized for node %s", providerID, node.Name)
	}

	serverID := strings.TrimPrefix(providerID, "hcloud://")

	// If caching is enabled try pulling nodeidentity.Info from cache before doing a Hetzner Cloud API call.
	if i.cacheEnabled {
		obj, exists, err := i.cache.GetByKey(serverID)
		if err != nil {
			klog.Warningf("Nodeidentity info cache lookup failure: %v", err)
		}
		if exists {
			return obj.(*nodeidentity.Info), nil
		}
	}

	server, err := i.getServer(serverID)
	if err != nil {
		return nil, err
	}

	if server.Status != hcloud.ServerStatusRunning && server.Status != hcloud.ServerStatusStarting {
		return nil, fmt.Errorf("found server %s(%d) with unexpected state: %q", server.Name, server.ID, server.Status)
	}

	labels := map[string]string{}
	for key, value := range server.Labels {
		if key == hetzner.TagKubernetesInstanceRole {
			switch kops.InstanceGroupRole(value) {
			case kops.InstanceGroupRoleMaster:
				labels[nodelabels.RoleLabelControlPlane20] = ""
			case kops.InstanceGroupRoleNode:
				labels[nodelabels.RoleLabelNode16] = ""
			case kops.InstanceGroupRoleAPIServer:
				labels[nodelabels.RoleLabelAPIServer16] = ""
			default:
				klog.Warningf("Unknown node role %q for server %s(%d)", value, server.Name, server.ID)
			}
		}
	}

	info := &nodeidentity.Info{
		InstanceID: serverID,
		Labels:     labels,
	}

	// If caching is enabled add the nodeidentity.Info to cache.
	if i.cacheEnabled {
		err = i.cache.Add(info)
		if err != nil {
			klog.Warningf("Failed to add node identity info to cache: %v", err)
		}
	}

	return info, nil
}

// stringKeyFunc is a string as cache key function
func stringKeyFunc(obj interface{}) (string, error) {
	key := obj.(*nodeidentity.Info).InstanceID
	return key, nil
}

// getServer queries Hetzner Cloud for the server with the specified ID, returning an error if not found
func (i *nodeIdentifier) getServer(id string) (*hcloud.Server, error) {
	serverID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("failed to convert server ID %q to int: %w", id, err)
	}
	server, _, err := i.client.Server.GetByID(context.TODO(), serverID)
	if err != nil || server == nil {
		return nil, fmt.Errorf("failed to get info for server %q: %w", id, err)
	}

	return server, nil
}
