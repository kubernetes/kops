/*
Copyright 2020 The Kubernetes Authors.

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

package azure

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	corev1 "k8s.io/api/core/v1"
	expirationcache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/nodeidentity"
	"k8s.io/kops/pkg/nodelabels"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

const (
	// InstanceGroupNameTag is the key of the tag used to identify an
	// instance group that VM ScaleSet belongs.
	InstanceGroupNameTag = "kops.k8s.io_instancegroup"
	// ClusterNodeTemplateLabel is the prefix used on node labels when copying to cloud tags.
	ClusterNodeTemplateLabel = "k8s.io_cluster_node-template_label_"
	// cacheTTL is the expiration time of nodeidentity.Info cache.
	cacheTTL = 60 * time.Minute
)

// nodeIdentifier identifies a node from Azure VM.
type nodeIdentifier struct {
	azureClient *client
	// cache is a cache of nodeidentity.Info
	cache expirationcache.Store
	// cacheEnabled indicates if caching should be used
	cacheEnabled bool
}

var _ nodeidentity.Identifier = &nodeIdentifier{}

// New creates and returns a a node identifier for Nodes running on Azure.
func New(cacheNodeidentityInfo bool) (nodeidentity.Identifier, error) {
	client, err := newClient()
	if err != nil {
		return nil, err
	}

	return &nodeIdentifier{
		azureClient:  client,
		cache:        expirationcache.NewTTLStore(stringKeyFunc, cacheTTL),
		cacheEnabled: cacheNodeidentityInfo,
	}, nil
}

// IdentifyNode queries Azure for the node identity information.
func (i *nodeIdentifier) IdentifyNode(ctx context.Context, node *corev1.Node) (*nodeidentity.Info, error) {
	providerID := node.Spec.ProviderID
	if providerID == "" {
		return nil, fmt.Errorf("providerID not set for node %q", node.Name)
	}
	if !strings.HasPrefix(providerID, "azure://") {
		return nil, fmt.Errorf("providerID %q not recognized for node %q", providerID, node.Name)
	}

	vmName, err := getVMNameFromProviderID(providerID)
	if err != nil {
		return nil, err
	}

	// If caching is enabled, try pulling nodeidentity.Info from the cache before doing an API call.
	if i.cacheEnabled {
		obj, exists, err := i.cache.GetByKey(vmName)
		if err != nil {
			klog.Warningf("Nodeidentity info cache lookup failure: %v", err)
		}
		if exists {
			return obj.(*nodeidentity.Info), nil
		}
	}

	tags, err := i.azureClient.getVMTags(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("error on getting VM ScaleSet: %s", err)
	}

	labels := map[string]string{}
	for k, v := range tags {
		if k == azure.TagClusterName && v != nil {
			labels[kops.LabelClusterName] = *v
		}
		if k == InstanceGroupNameTag && v != nil {
			labels[kops.NodeLabelInstanceGroup] = *v
		}
		if strings.HasPrefix(k, azure.TagNameRolePrefix) {
			role := strings.TrimPrefix(k, azure.TagNameRolePrefix)
			switch role {
			case kops.InstanceGroupRoleControlPlane.ToLowerString():
				labels[nodelabels.RoleLabelControlPlane20] = ""
			case "master":
				labels[nodelabels.RoleLabelControlPlane20] = ""
			case kops.InstanceGroupRoleNode.ToLowerString():
				labels[nodelabels.RoleLabelNode16] = ""
			default:
				klog.Warningf("Unknown or unsupported node role tag %q for VM %q", k, vmName)
			}
		}
		if strings.HasPrefix(k, ClusterNodeTemplateLabel) && v != nil {
			l := strings.SplitN(*v, "=", 2)
			if len(l) <= 1 {
				klog.Warningf("Malformed cloud label tag %q=%q for VM %q", k, *v, vmName)
			} else {
				labels[l[0]] = l[1]
			}
		}
	}

	info := &nodeidentity.Info{
		InstanceID: vmName,
		Labels:     labels,
	}

	// If caching is enabled, add the nodeidentity.Info to the cache.
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

func getVMNameFromProviderID(providerID string) (string, error) {
	if !strings.HasPrefix(providerID, "azure://") {
		return "", fmt.Errorf("providerID %q not recognized", providerID)
	}

	res, err := arm.ParseResourceID(strings.TrimPrefix(providerID, "azure://"))
	if err != nil {
		return "", fmt.Errorf("error parsing providerID: %v", err)
	}

	switch res.ResourceType.String() {
	case "Microsoft.Compute/virtualMachines":
		return res.Name, nil
	case "Microsoft.Compute/virtualMachineScaleSets/virtualMachines":
		return res.Parent.Name + "_" + res.Name, nil
	default:
		return "", fmt.Errorf("unsupported resource type %q for providerID %q", res.ResourceType, providerID)
	}
}
