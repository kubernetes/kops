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

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
	corev1 "k8s.io/api/core/v1"
	expirationcache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/nodeidentity"
)

const (
	// InstanceGroupNameTag is the key of the tag used to identify an
	// instance group that VM ScaleSet belongs.
	InstanceGroupNameTag = "kops.k8s.io_instancegroup"

	// cacheTTL is the expiration time of nodeidentity.Info cache.
	cacheTTL = 60 * time.Minute
)

type vmssGetter interface {
	getVMScaleSet(ctx context.Context, vmssName string) (compute.VirtualMachineScaleSet, error)
}

var _ vmssGetter = &client{}

// nodeIdentifier identifies a node from Azure VM.
type nodeIdentifier struct {
	vmssGetter vmssGetter

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
		vmssGetter:   client,
		cache:        expirationcache.NewTTLStore(stringKeyFunc, cacheTTL),
		cacheEnabled: cacheNodeidentityInfo,
	}, nil
}

// IdentifyNode queries Azure for the node identity information.
func (i *nodeIdentifier) IdentifyNode(ctx context.Context, node *corev1.Node) (*nodeidentity.Info, error) {
	providerID := node.Spec.ProviderID
	if providerID == "" {
		return nil, fmt.Errorf("providerID was not set for node %s", node.Name)
	}
	vmssName, err := getVMSSNameFromProviderID(providerID)
	if err != nil {
		return nil, fmt.Errorf("error on extracting VM ScaleSet name: %s", err)
	}

	// If caching is enabled try pulling nodeidentity.Info from cache before
	// doing a EC2 API call.
	if i.cacheEnabled {
		obj, exists, err := i.cache.GetByKey(vmssName)
		if err != nil {
			klog.Warningf("Nodeidentity info cache lookup failure: %v", err)
		}
		if exists {
			return obj.(*nodeidentity.Info), nil
		}
	}

	vmss, err := i.vmssGetter.getVMScaleSet(ctx, vmssName)
	if err != nil {
		return nil, fmt.Errorf("error on getting VM ScaleSet: %s", err)
	}

	labels := map[string]string{}
	// TODO(kenji): Populate labels

	info := &nodeidentity.Info{
		InstanceID: vmssName,
		Labels:     labels,
	}

	for k, v := range vmss.Tags {
		if strings.HasPrefix(k, InstanceGroupNameTag) {
			info.Labels[strings.TrimPrefix(k, InstanceGroupNameTag)] = *v
		}
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

func getVMSSNameFromProviderID(providerID string) (string, error) {
	if !strings.HasPrefix(providerID, "azure://") {
		return "", fmt.Errorf("providerID %q not recognized", providerID)
	}

	l := strings.Split(strings.TrimPrefix(providerID, "azure://"), "/")
	if len(l) != 11 {
		return "", fmt.Errorf("unexpected format of providerID %q", providerID)
	}
	return l[len(l)-3], nil
}
