/*
Copyright 2025 The Kubernetes Authors.

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

package elemento

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Elemento-Modular-Cloud/ecloud-go/ecloud"
	corev1 "k8s.io/api/core/v1"
	expirationcache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"

	// "k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/nodeidentity"
	"k8s.io/kops/pkg/nodelabels"
	"k8s.io/kops/upup/pkg/fi/cloudup/elemento"
)

const (
	cacheTTL = 60 * time.Minute
)

// nodeIdentifier identifies a node from Elemento
type nodeIdentifier struct {
	client       *ecloud.Client
	cache        expirationcache.Store
	cacheEnabled bool
}

// New creates and returns a nodeidentity.Identifier for Nodes running on Elemento
func New(CacheNodeidentityInfo bool) (nodeidentity.Identifier, error) {
	elementoClient, err := ecloud.NewClient("kops-elemento", "1.0")

	if err != nil {
		return nil, fmt.Errorf("creating client for Elemento Cloud: %w", err)
	}

	return &nodeIdentifier{
		client:       elementoClient,
		cache:        expirationcache.NewTTLStore(stringKeyFunc, cacheTTL),
		cacheEnabled: CacheNodeidentityInfo,
	}, nil
}

// IdentifyNode queries Elemento for the node identity information
func (i *nodeIdentifier) IdentifyNode(ctx context.Context, node *corev1.Node) (*nodeidentity.Info, error) {
	providerID := node.Spec.ProviderID
	if providerID == "" {
		return nil, fmt.Errorf("providerID not set for node %s", node.Name)
	}
	if !strings.HasPrefix(providerID, "elemento://") {
		return nil, fmt.Errorf("providerID %q not recognized for node %s", providerID, node.Name)
	}

	serverID := strings.TrimPrefix(providerID, "elemento://")

	// If cache is enabled, check if the node information is already cached
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

	if server.Status != "running" {
		return nil, fmt.Errorf("server %s is not running", serverID)
	}

	labels := map[string]string{}
	for key, value := range server.Labels {
		switch {
		case key == elemento.TagKubernetesInstanceRole:
			switch kops.InstanceGroupRole(value) {
			case kops.InstanceGroupRoleControlPlane:
				labels[nodelabels.RoleLabelControlPlane20] = ""
			case kops.InstanceGroupRoleNode:
				labels[nodelabels.RoleLabelNode16] = ""
			case kops.InstanceGroupRoleAPIServer:
				labels[nodelabels.RoleLabelAPIServer16] = ""
			default:
				klog.Warningf("Unknown node role %q for server %s(%s)", value, server.Name, server.ID)
			}
		case strings.HasPrefix(key, elemento.TagKubernetesNodeLabelPrefix):
			labels[strings.TrimPrefix(key, elemento.TagKubernetesNodeLabelPrefix)] = value
		}
	}

	info := &nodeidentity.Info{
		InstanceID: serverID,
		Labels:     labels,
	}

	// If cache is enabled, store the node information in the cache
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

// getServer retrieves the server information from Elemento for the given server ID
func (i *nodeIdentifier) getServer(id string) (*ecloud.Server, error) {
	server, _, err := i.client.Server.GetByID(context.TODO(), id)
	if err != nil || server == nil {
		return nil, fmt.Errorf("failed to get info for server %q: %w", id, err)
	}

	return server, nil
}
