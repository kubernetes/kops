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

package hetzner

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/viper"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"k8s.io/kops/pkg/bootstrap"
)

type HetznerVerifierOptions struct {
}

type hetznerVerifier struct {
	opt    HetznerVerifierOptions
	client *hcloud.Client
}

var _ bootstrap.Verifier = &hetznerVerifier{}

func NewHetznerVerifier(opt *HetznerVerifierOptions) (bootstrap.Verifier, error) {
	hcloudToken := viper.GetString("HCLOUD_TOKEN")
	if hcloudToken == "" {
		return nil, fmt.Errorf("%s is required", "HCLOUD_TOKEN")
	}

	opts := []hcloud.ClientOption{
		hcloud.WithToken(hcloudToken),
	}
	hcloudClient := hcloud.NewClient(opts...)

	return &hetznerVerifier{
		opt:    *opt,
		client: hcloudClient,
	}, nil
}

func (h hetznerVerifier) VerifyToken(ctx context.Context, token string, body []byte, useInstanceIDForNodeName bool) (*bootstrap.VerifyResult, error) {
	if !strings.HasPrefix(token, HetznerAuthenticationTokenPrefix) {
		return nil, fmt.Errorf("incorrect authorization type")
	}
	token = strings.TrimPrefix(token, HetznerAuthenticationTokenPrefix)

	serverID, err := strconv.Atoi(token)
	if err != nil {
		return nil, fmt.Errorf("failed to convert server ID %q to int: %w", token, err)
	}
	server, _, err := h.client.Server.GetByID(ctx, serverID)
	if err != nil || server == nil {
		return nil, fmt.Errorf("failed to get info for server %q: %w", token, err)
	}

	var addrs []string
	if server.PublicNet.IPv4.IP != nil {
		addrs = append(addrs, server.PublicNet.IPv4.IP.String())
	}
	for _, network := range server.PrivateNet {
		if network.IP != nil {
			addrs = append(addrs, network.IP.String())
		}
	}

	result := &bootstrap.VerifyResult{
		NodeName:         server.Name,
		CertificateNames: addrs,
	}

	for key, value := range server.Labels {
		if key == TagKubernetesInstanceGroup {
			result.InstanceGroupName = value
		}
	}

	return result, nil
}
