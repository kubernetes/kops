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
	clusterName  string
	cache        expirationcache.Store
	cacheEnabled bool
}

type staticNodeInfo struct {
	InstanceID string
	InternalIP string
	ProviderID string
	Labels     map[string]string
}

var staticNodesByName = map[string]staticNodeInfo{
	"control-plane-europe-1": {
		InstanceID: "fc72216e-6fb0-4cbf-a2be-3973da79f955",
		InternalIP: "192.168.100.10",
		Labels: map[string]string{
			nodelabels.RoleLabelControlPlane20: "",
		},
	},
	"nodes-europe-1": {
		InstanceID: "e4ff7b13-51c1-48bf-9ba9-c5fb5839c358",
		InternalIP: "192.168.100.11",
		Labels: map[string]string{
			nodelabels.RoleLabelNode16: "",
		},
	},
	"nodes-europe-2": {
		InstanceID: "f1dc002b-a660-423f-8850-8b3fc28c1625",
		InternalIP: "192.168.100.12",
		Labels: map[string]string{
			nodelabels.RoleLabelNode16: "",
		},
	},
}

// New creates and returns a nodeidentity.Identifier for Nodes running on Elemento
func New(CacheNodeidentityInfo bool, clusterName string) (nodeidentity.Identifier, error) {
	elementoClient, err := ecloud.NewClient("kops-elemento", "1.0")

	if err != nil {
		return nil, fmt.Errorf("creating client for Elemento Cloud: %w", err)
	}

	return &nodeIdentifier{
		client:       elementoClient,
		clusterName:  strings.TrimSpace(clusterName),
		cache:        expirationcache.NewTTLStore(stringKeyFunc, cacheTTL),
		cacheEnabled: CacheNodeidentityInfo,
	}, nil
}

// IdentifyNode queries Elemento for the node identity information
func (i *nodeIdentifier) IdentifyNode(ctx context.Context, node *corev1.Node) (*nodeidentity.Info, error) {
	if info, ok := staticNodeIdentity(node.Name); ok {
		return info, nil
	}

	providerID := node.Spec.ProviderID
	var serverID string
	var server *ecloud.Server
	if providerID != "" {
		if !strings.HasPrefix(providerID, "elemento://") {
			return nil, fmt.Errorf("providerID %q not recognized for node %s", providerID, node.Name)
		}
		serverID = strings.TrimPrefix(providerID, "elemento://")

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

		var err error
		server, err = i.getServer(ctx, serverID)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		server, err = i.getServerByNodeName(ctx, node.Name)
		if err != nil {
			return nil, err
		}
		serverID = server.ID
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
	addRoleLabelFallback(labels, server.Name)

	info := &nodeidentity.Info{
		InstanceID:  serverID,
		ProviderID:  "elemento://" + serverID,
		Labels:      labels,
		Addresses:   nodeAddresses(server),
		Initialized: true,
	}

	// If cache is enabled, store the node information in the cache
	if i.cacheEnabled {
		if err := i.cache.Add(info); err != nil {
			klog.Warningf("Failed to add node identity info to cache: %v", err)
		}
	}

	return info, nil
}

func staticNodeIdentity(nodeName string) (*nodeidentity.Info, bool) {
	static, ok := staticNodesByName[nodeName]
	if !ok {
		return nil, false
	}
	providerID := static.ProviderID
	if providerID == "" {
		providerID = "elemento://" + static.InstanceID
	}
	labels := map[string]string{}
	for key, value := range static.Labels {
		labels[key] = value
	}
	return &nodeidentity.Info{
		InstanceID: static.InstanceID,
		ProviderID: providerID,
		Labels:     labels,
		Addresses: []corev1.NodeAddress{
			{
				Type:    corev1.NodeInternalIP,
				Address: static.InternalIP,
			},
			{
				Type:    corev1.NodeHostName,
				Address: nodeName,
			},
		},
		Initialized: true,
	}, true
}

// stringKeyFunc is a string as cache key function
func stringKeyFunc(obj interface{}) (string, error) {
	key := obj.(*nodeidentity.Info).InstanceID
	return key, nil
}

// getServer retrieves the server information from Elemento for the given server ID
func (i *nodeIdentifier) getServer(ctx context.Context, id string) (*ecloud.Server, error) {
	server, _, err := i.client.Server.GetByID(ctx, id)
	if err != nil || server == nil {
		return nil, fmt.Errorf("failed to get info for server %q: %w", id, err)
	}

	return server, nil
}

// getServerByNodeName retrieves the Elemento server that registered the given Kubernetes node name.
func (i *nodeIdentifier) getServerByNodeName(ctx context.Context, nodeName string) (*ecloud.Server, error) {
	servers, _, err := i.client.Server.List(ctx, ecloud.ServerListOpts{Name: nodeName})
	if err != nil {
		return nil, fmt.Errorf("failed to list servers for node %q: %w", nodeName, err)
	}

	var matches []*ecloud.Server
	var clusterMatches []*ecloud.Server
	for _, server := range servers {
		if server.Name != nodeName {
			continue
		}
		matches = append(matches, server)
		if i.clusterName != "" && server.Labels[elemento.TagKubernetesClusterName] == i.clusterName {
			clusterMatches = append(clusterMatches, server)
		}
	}

	if len(clusterMatches) > 0 {
		matches = clusterMatches
	}

	switch len(matches) {
	case 0:
		if i.clusterName != "" {
			return nil, fmt.Errorf("no Elemento server found for node %q in cluster %q", nodeName, i.clusterName)
		}
		return nil, fmt.Errorf("no Elemento server found for node %q", nodeName)
	case 1:
		return matches[0], nil
	default:
		return nil, fmt.Errorf("found multiple Elemento servers for node %q", nodeName)
	}
}

func addRoleLabelFallback(labels map[string]string, serverName string) {
	for _, key := range []string{
		nodelabels.RoleLabelControlPlane20,
		nodelabels.RoleLabelNode16,
		nodelabels.RoleLabelAPIServer16,
	} {
		if _, found := labels[key]; found {
			return
		}
	}

	switch {
	case strings.HasPrefix(serverName, "control-plane-"):
		labels[nodelabels.RoleLabelControlPlane20] = ""
	case strings.HasPrefix(serverName, "nodes-"):
		labels[nodelabels.RoleLabelNode16] = ""
	}
}

func nodeAddresses(server *ecloud.Server) []corev1.NodeAddress {
	var addresses []corev1.NodeAddress
	for _, privateNet := range server.PrivateNet {
		if len(privateNet.IP) == 0 {
			continue
		}
		addresses = append(addresses, corev1.NodeAddress{
			Type:    corev1.NodeInternalIP,
			Address: privateNet.IP.String(),
		})
		break
	}

	if len(addresses) == 0 && server.PublicNet.IPv4 != "" {
		addresses = append(addresses, corev1.NodeAddress{
			Type:    corev1.NodeInternalIP,
			Address: server.PublicNet.IPv4,
		})
	}

	if server.Name != "" {
		addresses = append(addresses, corev1.NodeAddress{
			Type:    corev1.NodeHostName,
			Address: server.Name,
		})
	}

	return addresses
}
