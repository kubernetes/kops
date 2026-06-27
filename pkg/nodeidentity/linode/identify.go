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

package linode

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/linode/linodego"
	corev1 "k8s.io/api/core/v1"
	expirationcache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/nodeidentity"
	"k8s.io/kops/pkg/nodelabels"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

const (
	cacheTTL         = 60 * time.Minute
	providerIDPrefix = "linode://"
)

type linodeClient interface {
	GetInstance(ctx context.Context, linodeID int) (*linodego.Instance, error)
}

// nodeIdentifier identifies a node from Linode (Akamai).
type nodeIdentifier struct {
	client       linodeClient
	cache        expirationcache.Store
	cacheEnabled bool
}

// New creates and returns a nodeidentity.Identifier for Nodes running on Linode (Akamai).
func New(cacheNodeidentityInfo bool) (nodeidentity.Identifier, error) {
	accessToken := os.Getenv("LINODE_TOKEN")
	if accessToken == "" {
		return nil, fmt.Errorf("%s is required", "LINODE_TOKEN")
	}

	client := linodego.NewClient(nil)
	client.SetUserAgent("kops/nodeidentity")
	client.SetToken(accessToken)

	return &nodeIdentifier{
		client:       &client,
		cache:        expirationcache.NewTTLStore(stringKeyFunc, cacheTTL),
		cacheEnabled: cacheNodeidentityInfo,
	}, nil
}

// IdentifyNode queries Linode (Akamai) for the node identity information.
func (i *nodeIdentifier) IdentifyNode(ctx context.Context, node *corev1.Node) (*nodeidentity.Info, error) {
	providerID := node.Spec.ProviderID
	if providerID == "" {
		return nil, fmt.Errorf("providerID not set for node %s", node.Name)
	}

	instanceNumericID, instanceID, err := parseInstanceIDFromProviderID(providerID)
	if err != nil {
		return nil, fmt.Errorf("providerID %q not recognized for node %s: %w", providerID, node.Name, err)
	}

	if i.cacheEnabled {
		obj, exists, err := i.cache.GetByKey(instanceID)
		if err != nil {
			klog.Warningf("Nodeidentity info cache lookup failure: %v", err)
		}
		if exists {
			return obj.(*nodeidentity.Info), nil
		}
	}

	instance, err := i.client.GetInstance(ctx, instanceNumericID)
	if err != nil {
		return nil, fmt.Errorf("failed to get info for Linode (Akamai) instance %q: %w", instanceID, err)
	}
	if instance == nil {
		return nil, fmt.Errorf("failed to get info for Linode (Akamai) instance %q: empty response", instanceID)
	}

	if !isExpectedInstanceStatus(instance.Status) {
		return nil, fmt.Errorf("found instance %q with unexpected status: %q", instanceID, instance.Status)
	}

	info := &nodeidentity.Info{
		InstanceID: instanceID,
		Labels:     buildLabelsFromTags(instance.Tags),
	}

	if i.cacheEnabled {
		if err := i.cache.Add(info); err != nil {
			klog.Warningf("Failed to add node identity info to cache: %v", err)
		}
	}

	return info, nil
}

// parseInstanceIDFromProviderID extracts the numeric instance ID and string ID from a provider ID.
// It supports formats like "linode://123", "linode:///123", and "linode://region/123".
func parseInstanceIDFromProviderID(providerID string) (int, string, error) {
	if !strings.HasPrefix(providerID, providerIDPrefix) {
		return 0, "", fmt.Errorf("missing prefix %q", providerIDPrefix)
	}

	raw := strings.TrimPrefix(providerID, providerIDPrefix)
	raw = strings.Trim(raw, "/")
	if raw == "" {
		return 0, "", fmt.Errorf("missing instance id")
	}

	parts := strings.Split(raw, "/")
	instanceID := parts[len(parts)-1]
	if instanceID == "" {
		return 0, "", fmt.Errorf("missing instance id")
	}

	linodeID, err := strconv.Atoi(instanceID)
	if err != nil {
		return 0, "", fmt.Errorf("invalid instance id %q: %w", instanceID, err)
	}

	return linodeID, instanceID, nil
}

// isExpectedInstanceStatus returns true if the instance is in a state where it can be identified.
func isExpectedInstanceStatus(status linodego.InstanceStatus) bool {
	switch status {
	case linodego.InstanceRunning, linodego.InstanceBooting, linodego.InstanceProvisioning:
		return true
	default:
		return false
	}
}

// buildLabelsFromTags converts Linode instance tags into Kubernetes node labels.
// It handles role tags (kops.k8s.io/instance-role) and direct label tags (kops.k8s.io/* and node-role.kubernetes.io/*).
func buildLabelsFromTags(tags []string) map[string]string {
	labels := map[string]string{}

	for _, tag := range tags {
		key, value, found := strings.Cut(tag, ":")
		if !found {
			continue
		}

		// Derive role labels from instance role tag (must be checked before direct label handling)
		if key == linode.TagKubernetesInstanceRole {
			switch kops.InstanceGroupRole(value) {
			case kops.InstanceGroupSubRoleControlPlane.Role():
				labels[nodelabels.RoleLabelControlPlane20] = ""
			case kops.InstanceGroupSubRoleNode.Role():
				labels[nodelabels.RoleLabelNode16] = ""
			case kops.InstanceGroupSubRoleAPIServer.Role():
				labels[nodelabels.RoleLabelAPIServer16] = ""
			default:
				klog.Warningf("Unknown node role %q for Linode (Akamai) instance", value)
			}
			continue
		}

		// Handle direct label tags (for critical labels like kops.k8s.io/instancegroup)
		if strings.HasPrefix(key, "kops.k8s.io/") || strings.HasPrefix(key, "node-role.kubernetes.io/") {
			labels[key] = value
		}
	}

	return labels
}

func stringKeyFunc(obj interface{}) (string, error) {
	return obj.(*nodeidentity.Info).InstanceID, nil
}
