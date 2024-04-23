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

func NewScalewayVerifier(_ context.Context, _ *ScalewayVerifierOptions) (bootstrap.Verifier, error) {
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

func (v scalewayVerifier) VerifyToken(ctx context.Context, _ *http.Request, token string, _ []byte) (*bootstrap.VerifyResult, error) {
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

	privateIP, err := GetPrivateIP(scwClient, serverID, zone)
	if err != nil {
		return nil, fmt.Errorf("failed to get private IP for server %s: %w", serverID, err)
	}

	result := &bootstrap.VerifyResult{
		NodeName:          server.Name,
		InstanceGroupName: InstanceGroupNameFromTags(server.Tags),
		CertificateNames:  []string{privateIP},
		ChallengeEndpoint: net.JoinHostPort(privateIP, strconv.Itoa(wellknownports.NodeupChallenge)),
	}

	return result, nil
}
