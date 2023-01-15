/*
Copyright 2023 The Kubernetes Authors.

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
	gos "github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/mitchellh/mapstructure"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/bootstrap"
)

type OpenStackVerifierOptions struct {
}

type openstackVerifier struct {
	novaClient *gophercloud.ServiceClient
}

var _ bootstrap.Verifier = &openstackVerifier{}

func NewOpenstackVerifier(opt *OpenStackVerifierOptions) (bootstrap.Verifier, error) {
	env, err := gos.AuthOptionsFromEnv()
	if err != nil {
		return nil, err
	}

	region := os.Getenv("OS_REGION_NAME")
	if region == "" {
		return nil, fmt.Errorf("unable to find region")
	}

	provider, err := gos.NewClient(env.IdentityEndpoint)
	if err != nil {
		return nil, err
	}
	ua := gophercloud.UserAgent{}
	ua.Prepend("kops/kopscontrollerverifier")
	provider.UserAgent = ua
	klog.V(4).Infof("Using user-agent %s", ua.Join())

	// node-controller should be able to renew it tokens against OpenStack API
	env.AllowReauth = true

	err = gos.Authenticate(provider, env)
	if err != nil {
		return nil, err
	}

	novaClient, err := gos.NewComputeV2(provider, gophercloud.EndpointOpts{
		Type:   "compute",
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("error building nova client: %v", err)
	}

	return &openstackVerifier{
		novaClient: novaClient,
	}, nil
}

func (o openstackVerifier) VerifyToken(ctx context.Context, token string, remoteAddr string, body []byte, useInstanceIDForNodeName bool) (*bootstrap.VerifyResult, error) {
	if !strings.HasPrefix(token, OpenstackAuthenticationTokenPrefix) {
		return nil, fmt.Errorf("incorrect authorization type")
	}
	serverID := strings.TrimPrefix(token, OpenstackAuthenticationTokenPrefix)

	instance, err := servers.Get(o.novaClient, serverID).Extract()
	if err != nil {
		return nil, fmt.Errorf("failed to get info for server %q: %w", token, err)
	}

	var addrs []string

	var addresses map[string][]Address
	err = mapstructure.Decode(instance.Addresses, &addresses)
	if err != nil {
		return nil, fmt.Errorf("unable to decode addresses: %w", err)
	}

	for _, addrList := range addresses {
		for _, props := range addrList {
			addrs = append(addrs, props.Addr)
		}
	}

	allowed := false
	for _, addr := range addrs {
		if addr == remoteAddr {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, fmt.Errorf("request is not coming from trusted sources %v (request: %s)", addrs, remoteAddr)
	}

	now := time.Now()
	diff := now.Sub(instance.Created)
	if diff > 30*time.Minute || diff < 0 {
		return nil, fmt.Errorf("instance created time diff is %s, denying request", diff)
	}

	result := &bootstrap.VerifyResult{
		NodeName:         instance.Name,
		CertificateNames: addrs,
	}
	value, ok := instance.Metadata[TagKopsInstanceGroup]
	if ok {
		result.InstanceGroupName = value
	}
	return result, nil
}
