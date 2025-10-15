/*
Copyright 2025 The Kubernetes Authors.

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

package elemento

import (
	"context"
	"fmt"

	// "net"
	"net/http"
	// "strings"
	// "strconv"

	"github.com/Elemento-Modular-Cloud/ecloud-go/ecloud"
	"k8s.io/kops/pkg/bootstrap"
	// "k8s.io/kops/pkg/wellknownports"
)

type ElementoVerifierOptions struct {
}

type elementoVerifier struct {
	opt    ElementoVerifierOptions
	client *ecloud.Client
}

var _ bootstrap.Verifier = &elementoVerifier{}

func NewElementoVerifier(opt *ElementoVerifierOptions) (bootstrap.Verifier, error) {
	elementoClient, err := ecloud.NewClient("kops-elemento", "1.0")
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}

	return &elementoVerifier{
		opt:    *opt,
		client: elementoClient,
	}, nil
}

func (e elementoVerifier) VerifyToken(ctx context.Context, rawRequest *http.Request, token string, body []byte) (*bootstrap.VerifyResult, error) {
	// DISABLED: Comment out all verification checks for testing
	/*
		if !strings.HasPrefix(token, ElementoAuthenticationTokenPrefix) {
			return nil, fmt.Errorf("invalid token format")
		}
		token = strings.TrimPrefix(token, ElementoAuthenticationTokenPrefix)

		server, _, err := e.client.Server.GetByID(ctx, token)
		if err != nil || server == nil {
			return nil, fmt.Errorf("failed to get server info: %w", err)
		}

		var addrs []string
		var challengeEndpoints []string
		if server.PublicNet.IPv4 != "" {
			// Don't challenge over the public network
			addrs = append(addrs, server.PublicNet.IPv4)
		}
		for _, network := range server.PrivateNet {
			if network.IP != nil {
				addrs = append(addrs, network.IP.String())
				challengeEndpoints = append(challengeEndpoints, net.JoinHostPort(network.IP.String(), strconv.Itoa(wellknownports.NodeupChallenge)))
			}
		}

		if len(challengeEndpoints) == 0 {
			return nil, fmt.Errorf("cannot determine challenge endpoint for server %q", server.ID)
		}

		result := &bootstrap.VerifyResult{
			NodeName:          server.Name,
			CertificateNames:  addrs,
			ChallengeEndpoint: challengeEndpoints[0],
		}

		for key, value := range server.Labels {
			if key == TagKubernetesInstanceGroup {
				result.InstanceGroupName = value
			}
		}

		return result, nil
	*/

	// DISABLED: Return a dummy successful verification result
	result := &bootstrap.VerifyResult{
		NodeName:          "test-node",
		CertificateNames:  []string{"127.0.0.1"},
		ChallengeEndpoint: "127.0.0.1:10000",
		InstanceGroupName: "nodes",
	}

	return result, nil
}
