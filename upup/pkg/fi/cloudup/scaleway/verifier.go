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
	"github.com/scaleway/scaleway-sdk-go/scw"
	kopsv "k8s.io/kops"
	"k8s.io/kops/pkg/bootstrap"
	"k8s.io/kops/pkg/wellknownports"
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

func (v scalewayVerifier) VerifyToken(ctx context.Context, rawRequest *http.Request, token string, body []byte, useInstanceIDForNodeName bool) (*bootstrap.VerifyResult, error) {
	if !strings.HasPrefix(token, ScalewayAuthenticationTokenPrefix) {
		return nil, fmt.Errorf("incorrect authorization type")
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

	instanceAPI := instance.NewAPI(scwClient)
	serverResponse, err := instanceAPI.GetServer(&instance.GetServerRequest{
		ServerID: serverID,
		Zone:     zone,
	}, scw.WithContext(ctx))
	if err != nil || serverResponse == nil {
		return nil, fmt.Errorf("failed to get server %s: %w", serverID, err)
	}
	server := serverResponse.Server

	addresses := []string(nil)
	challengeEndPoints := []string(nil)
	if server.PrivateIP != nil {
		addresses = append(addresses, *server.PrivateIP)
		challengeEndPoints = append(challengeEndPoints, net.JoinHostPort(*server.PrivateIP, strconv.Itoa(wellknownports.NodeupChallenge)))
	}
	if server.IPv6 != nil {
		addresses = append(addresses, server.IPv6.Address.String())
		challengeEndPoints = append(challengeEndPoints, net.JoinHostPort(server.IPv6.Address.String(), strconv.Itoa(wellknownports.NodeupChallenge)))
	}

	igName := ""
	for _, tag := range server.Tags {
		if strings.HasPrefix(tag, TagInstanceGroup) {
			igName = strings.TrimPrefix(tag, TagInstanceGroup+"=")
		}
	}

	result := &bootstrap.VerifyResult{
		NodeName:          server.Name,
		InstanceGroupName: igName,
		CertificateNames:  addresses,
		ChallengeEndpoint: challengeEndPoints[0],
	}

	return result, nil
}
