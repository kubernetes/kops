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

package scaleway

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	expirationcache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	kopsv "k8s.io/kops"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/nodeidentity"
	"k8s.io/kops/pkg/nodelabels"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	cacheTTL = 60 * time.Minute
)

// nodeIdentifier identifies a node from Scaleway
type nodeIdentifier struct {
	client       *scw.Client
	cache        expirationcache.Store
	cacheEnabled bool
}

// New creates and returns a nodeidentity.Identifier for Nodes running on Scaleway
func New(CacheNodeidentityInfo bool) (nodeidentity.Identifier, error) {
	scwClient, err := scw.NewClient(
		scw.WithUserAgent("kubernetes-kops/"+kopsv.Version),
		scw.WithEnv(),
	)
	if err != nil {
		return nil, err
	}

	return &nodeIdentifier{
		client:       scwClient,
		cache:        expirationcache.NewTTLStore(stringKeyFunc, cacheTTL),
		cacheEnabled: CacheNodeidentityInfo,
	}, nil
}

// IdentifyNode queries Scaleway for the node identify information
func (i *nodeIdentifier) IdentifyNode(ctx context.Context, node *corev1.Node) (*nodeidentity.Info, error) {
	providerID := node.Spec.ProviderID
	if providerID == "" {
		return nil, fmt.Errorf("providerID not set for node %s", node.Name)
	}
	if !strings.HasPrefix(providerID, "scaleway://") {
		return nil, fmt.Errorf("providerID %q not recognized for node %s", providerID, node.Name)
	}

	serverID := strings.TrimPrefix(providerID, "scaleway://")

	// If caching is enabled try pulling nodeidentity.Info from cache before doing a Scaleway API call.
	if i.cacheEnabled {
		obj, exists, err := i.cache.GetByKey(serverID)
		if err != nil {
			klog.Warningf("Nodeidentity info cache lookup failure: %v", err)
		}
		if exists {
			return obj.(*nodeidentity.Info), nil
		}
	}

	server, err := i.getServer(ctx, serverID)
	if err != nil {
		return nil, err
	}

	if server.State != instance.ServerStateRunning && server.State != instance.ServerStateStarting {
		return nil, fmt.Errorf("found server %s (%s) with unexpected state: %q", server.Name, server.ID, server.State)
	}

	labels := map[string]string{}
	for _, tag := range server.Tags {
		if strings.HasPrefix(tag, scaleway.TagNameRolePrefix) {
			role := strings.TrimPrefix(tag, scaleway.TagNameRolePrefix+"=")
			switch kops.InstanceGroupRole(role) {
			case kops.InstanceGroupRoleControlPlane:
				labels[nodelabels.RoleLabelControlPlane20] = ""
			case kops.InstanceGroupRoleNode:
				labels[nodelabels.RoleLabelNode16] = ""
			case kops.InstanceGroupRoleAPIServer:
				labels[nodelabels.RoleLabelAPIServer16] = ""
			default:
				klog.Warningf("Unknown node role %q for server %s(%d)", role, server.Name, server.ID)
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

// getServer queries Scaleway for the server with the specified ID, returning an error if not found
func (i *nodeIdentifier) getServer(ctx context.Context, id string) (*instance.Server, error) {
	api := instance.NewAPI(i.client)
	uuid := strings.Split(id, "/")
	if len(uuid) != 3 {
		return nil, fmt.Errorf("unexpected format for server id %s", id)
	}
	server, err := api.GetServer(&instance.GetServerRequest{
		ServerID: uuid[2],
		Zone:     scw.Zone(uuid[1]),
	}, scw.WithContext(ctx))
	if err != nil || server == nil {
		return nil, fmt.Errorf("failed to get server %s: %w", id, err)
	}

	return server.Server, nil
}
