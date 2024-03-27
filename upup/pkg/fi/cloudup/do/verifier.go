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

package do

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
	"k8s.io/kops/pkg/bootstrap"
	"k8s.io/kops/pkg/wellknownports"
)

type DigitalOceanVerifierOptions struct {
}

type digitalOceanVerifier struct {
	doClient *godo.Client
}

var _ bootstrap.Verifier = &digitalOceanVerifier{}

func NewVerifier(ctx context.Context, opt *DigitalOceanVerifierOptions) (bootstrap.Verifier, error) {
	accessToken := os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")
	if accessToken == "" {
		return nil, errors.New("DIGITALOCEAN_ACCESS_TOKEN is required")
	}

	tokenSource := &TokenSource{
		AccessToken: accessToken,
	}

	oauthClient := oauth2.NewClient(ctx, tokenSource)
	doClient := godo.NewClient(oauthClient)

	return &digitalOceanVerifier{
		doClient: doClient,
	}, nil
}

func (o digitalOceanVerifier) VerifyToken(ctx context.Context, rawRequest *http.Request, token string, body []byte) (*bootstrap.VerifyResult, error) {
	if !strings.HasPrefix(token, DOAuthenticationTokenPrefix) {
		return nil, bootstrap.ErrNotThisVerifier
	}
	serverIDString := strings.TrimPrefix(token, DOAuthenticationTokenPrefix)

	serverID, err := strconv.Atoi(serverIDString)
	if err != nil {
		return nil, fmt.Errorf("invalid authorization token")
	}

	droplet, _, err := o.doClient.Droplets.Get(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to get info for server %v: %w", token, err)
	}

	var addresses []string
	var challengeEndpoints []string
	if droplet.Networks != nil {
		for _, nic := range droplet.Networks.V4 {
			if nic.Type == "private" {
				addresses = append(addresses, nic.IPAddress)
				challengeEndpoints = append(challengeEndpoints, net.JoinHostPort(nic.IPAddress, strconv.Itoa(wellknownports.NodeupChallenge)))
			}
		}
		for _, nic := range droplet.Networks.V6 {
			if nic.Type == "private" {
				addresses = append(addresses, nic.IPAddress)
				challengeEndpoints = append(challengeEndpoints, net.JoinHostPort(nic.IPAddress, strconv.Itoa(wellknownports.NodeupChallenge)))
			}
		}
	}

	// Note: we use TLS passthrough, so we lose the client IP address.
	// We therefore don't have a great way to verify the request.
	// We do at least prevent duplicate node registrations, preventing some attacks here.

	// The node challenge is important here though, verifying the caller has control of the IP address.

	if len(challengeEndpoints) == 0 {
		return nil, fmt.Errorf("cannot determine challenge endpoint for server %q", serverID)
	}

	result := &bootstrap.VerifyResult{
		NodeName:          strconv.Itoa(droplet.ID),
		CertificateNames:  addresses,
		ChallengeEndpoint: challengeEndpoints[0],
	}

	for _, tag := range droplet.Tags {
		if strings.HasPrefix(tag, TagKubernetesInstanceGroup+":") {
			result.InstanceGroupName = strings.TrimPrefix(tag, TagKubernetesInstanceGroup+":")
		}
	}
	return result, nil
}
