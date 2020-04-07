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

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/kops/pkg/nodeidentity"
)

// nodeIdentifier identifies a node
type nodeIdentifier struct {
	novaClient *gophercloud.ServiceClient
}

// New creates and returns a nodeidentity.Identifier for Nodes running on OpenStack
func New() (nodeidentity.Identifier, error) {
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
		novaClient: novaClient,
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
	if strings.HasPrefix(instanceID, "/") {
		instanceID = strings.TrimPrefix(instanceID, "/")
	}

	kopsGroup, err := i.getInstanceGroup(instanceID)
	if err != nil {
		return nil, err
	}

	info := &nodeidentity.Info{}
	info.InstanceGroup = kopsGroup

	return info, nil
}

func (i *nodeIdentifier) getInstanceGroup(instanceID string) (string, error) {
	instance, err := servers.Get(i.novaClient, instanceID).Extract()
	if err != nil {
		return "", err
	}

	if val, ok := instance.Metadata["KopsInstanceGroup"]; ok {
		return val, nil
	}
	return "", fmt.Errorf("could not find tag 'KopsInstanceGroup' from instance metadata")
}
