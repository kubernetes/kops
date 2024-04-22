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

package scaleway

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	kopsv "k8s.io/kops"
	"k8s.io/kops/pkg/bootstrap"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
)

type ScalewayVerifierOptions struct{}

type scalewayVerifier struct {
	scwClient *scw.Client
}

var _ bootstrap.Verifier = &scalewayVerifier{}

func NewScalewayVerifier(ctx context.Context, opt *ScalewayVerifierOptions) (bootstrap.Verifier, error) {
	profile, err := CreateValidScalewayProfile()
	if err != nil {
		return nil, fmt.Errorf("creating client for Scaleway Verifier: %w", err)
	}
	scwClient, err := scw.NewClient(
		scw.WithProfile(profile),
		scw.WithUserAgent(KopsUserAgentPrefix+kopsv.Version),
	)
	if err != nil {
		return nil, err
	}
	return &scalewayVerifier{
		scwClient: scwClient,
	}, nil
}

func (v scalewayVerifier) VerifyToken(ctx context.Context, rawRequest *http.Request, token string, body []byte) (*bootstrap.VerifyResult, error) {
	if !strings.HasPrefix(token, ScalewayAuthenticationTokenPrefix) {
		return nil, bootstrap.ErrNotThisVerifier
	}
	serverID := strings.TrimPrefix(token, ScalewayAuthenticationTokenPrefix)

	metadataAPI := instance.NewMetadataAPI()
	metadata, err := metadataAPI.GetMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve server metadata: %w", err)
	}
	zone, err := scw.ParseZone(metadata.Location.ZoneID)
	if err != nil {
		return nil, fmt.Errorf("unable to parse Scaleway zone %q: %w", metadata.Location.ZoneID, err)
	}
	region, err := zone.Region()
	if err != nil {
		return nil, fmt.Errorf("unable to determine region from zone %s", zone)
	}

	profile, err := CreateValidScalewayProfile()
	if err != nil {
		return nil, err
	}
	scwClient, err := scw.NewClient(
		scw.WithProfile(profile),
		scw.WithUserAgent(KopsUserAgentPrefix+kopsv.Version),
	)
	if err != nil {
		return nil, fmt.Errorf("creating client for Scaleway Verifier: %w", err)
	}

	serverResponse, err := instance.NewAPI(scwClient).GetServer(&instance.GetServerRequest{
		ServerID: serverID,
		Zone:     zone,
	}, scw.WithContext(ctx))
	if err != nil || serverResponse == nil || serverResponse.Server == nil {
		return nil, fmt.Errorf("failed to get server %s: %w", serverID, err)
	}
	server := serverResponse.Server

	ips, err := ipam.NewAPI(scwClient).ListIPs(&ipam.ListIPsRequest{
		Region:     region,
		ResourceID: fi.PtrTo(server.ID),
		IsIPv6:     fi.PtrTo(false),
		Zonal:      fi.PtrTo(zone.String()),
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("failed to get IP for server %q: %w", server.Name, err)
	}
	if ips.TotalCount == 0 {
		return nil, fmt.Errorf("no IP found for server %q: %w", server.Name, err)
	}

	addresses := []string(nil)
	challengeEndPoints := []string(nil)
	for _, ip := range ips.IPs {
		addresses = append(addresses, ip.Address.IP.String())
		challengeEndPoints = append(challengeEndPoints, net.JoinHostPort(ip.Address.IP.String(), strconv.Itoa(wellknownports.NodeupChallenge)))
	}

	result := &bootstrap.VerifyResult{
		NodeName:          server.Name,
		InstanceGroupName: InstanceGroupNameFromTags(server.Tags),
		CertificateNames:  addresses,
		ChallengeEndpoint: challengeEndPoints[0],
	}

	return result, nil
}
