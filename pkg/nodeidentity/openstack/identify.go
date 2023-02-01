/*
Copyright 2019 The Kubernetes Authors.

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

package openstack

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	corev1 "k8s.io/api/core/v1"
	expirationcache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/nodeidentity"
	"k8s.io/kops/pkg/nodelabels"
	kos "k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

const (
	cacheTTL                           = 60 * time.Minute
	ClusterAutoscalerNodeTemplateLabel = "k8s.io_cluster-autoscaler_node-template_label_"
)

// nodeIdentifier identifies a node
type nodeIdentifier struct {
	novaClient   *gophercloud.ServiceClient
	cache        expirationcache.Store
	cacheEnabled bool
}

// New creates and returns a nodeidentity.Identifier for Nodes running on OpenStack
func New(CacheNodeidentityInfo bool) (nodeidentity.Identifier, error) {
	env, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return nil, err
	}

	region := os.Getenv("OS_REGION_NAME")
	if region == "" {
		return nil, fmt.Errorf("unable to find region")
	}

	provider, err := openstack.NewClient(env.IdentityEndpoint)
	if err != nil {
		return nil, err
	}
	ua := gophercloud.UserAgent{}
	ua.Prepend("kops/nodeidentity")
	provider.UserAgent = ua
	klog.V(4).Infof("Using user-agent %s", ua.Join())

	// node-controller should be able to renew it tokens against OpenStack API
	env.AllowReauth = true

	err = openstack.Authenticate(provider, env)
	if err != nil {
		return nil, err
	}

	novaClient, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
		Type:   "compute",
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("error building nova client: %v", err)
	}

	return &nodeIdentifier{
		novaClient:   novaClient,
		cache:        expirationcache.NewTTLStore(stringKeyFunc, cacheTTL),
		cacheEnabled: CacheNodeidentityInfo,
	}, nil
}

// IdentifyNode queries OpenStack for the node identity information
func (i *nodeIdentifier) IdentifyNode(ctx context.Context, node *corev1.Node) (*nodeidentity.Info, error) {
	providerID := node.Spec.ProviderID
	if providerID == "" {
		return nil, fmt.Errorf("providerID was not set for node %s", node.Name)
	}
	if !strings.HasPrefix(providerID, "openstack://") {
		return nil, fmt.Errorf("providerID %q not recognized for node %s", providerID, node.Name)
	}

	instanceID := strings.TrimPrefix(providerID, "openstack://")
	// instanceid looks like its openstack:/// but no idea is that really correct like that?
	// this supports now both openstack:// and openstack:/// format

	instanceID = strings.TrimPrefix(instanceID, "/")

	// If caching is enabled try pulling nodeidentity.Info from cache before doing a Hetzner Cloud API call.
	if i.cacheEnabled {
		obj, exists, err := i.cache.GetByKey(instanceID)
		if err != nil {
			klog.Warningf("Nodeidentity info cache lookup failure: %v", err)
		}
		if exists {
			return obj.(*nodeidentity.Info), nil
		}
	}

	server, err := servers.Get(i.novaClient, instanceID).Extract()
	if err != nil {
		return nil, err
	}

	labels := map[string]string{}
	value, ok := server.Metadata[kos.TagKopsRole]
	if ok {
		switch kops.InstanceGroupRole(value) {
		case kops.InstanceGroupRoleControlPlane:
			labels[nodelabels.RoleLabelControlPlane20] = ""
		case kops.InstanceGroupRoleNode:
			labels[nodelabels.RoleLabelNode16] = ""
		case kops.InstanceGroupRoleAPIServer:
			labels[nodelabels.RoleLabelAPIServer16] = ""
		default:
			klog.Warningf("Unknown node role %q for server %s(%d)", value, server.Name, server.ID)
		}
	}

	for key, value := range server.Metadata {
		if strings.HasPrefix(key, ClusterAutoscalerNodeTemplateLabel) {
			trimKey := strings.ReplaceAll(strings.TrimPrefix(key, ClusterAutoscalerNodeTemplateLabel), "_", "/")
			labels[trimKey] = value
		}
	}

	info := &nodeidentity.Info{
		InstanceID: instanceID,
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
