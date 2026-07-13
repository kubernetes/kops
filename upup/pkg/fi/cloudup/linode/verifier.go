/*
Copyright 2026 The Kubernetes Authors.

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

package linode

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/linode/linodego/v2"
	"k8s.io/kops/pkg/bootstrap"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode/linodemetadata"
)

type LinodeVerifierOptions struct{}

const (
	TagKubernetesInstanceGroup = "kops.k8s.io/instance-group"
	TagKubernetesInstanceRole  = "kops.k8s.io/instance-role"
)

type linodeVerifierClient interface {
	GetInstance(ctx context.Context, linodeID int) (*linodego.Instance, error)
}

type linodeVerifier struct {
	client linodeVerifierClient
}

var _ bootstrap.Verifier = (*linodeVerifier)(nil)

// NewLinodeVerifier returns a bootstrap.Verifier that can verify Linode (Akamai) instance tokens using the LINODE_TOKEN environment variable.
func NewLinodeVerifier(opt *LinodeVerifierOptions) (bootstrap.Verifier, error) {
	accessToken := os.Getenv("LINODE_TOKEN")
	if accessToken == "" {
		return nil, fmt.Errorf("%s is required", "LINODE_TOKEN")
	}

	client, err := linodego.NewClient(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Linode client: %w", err)
	}
	client.SetUserAgent("kops")
	client.SetToken(accessToken)

	return &linodeVerifier{client: &client}, nil
}

// VerifyToken verifies that the given token corresponds to a valid Linode (Akamai) instance.
func (v *linodeVerifier) VerifyToken(ctx context.Context, rawRequest *http.Request, token string, body []byte) (*bootstrap.VerifyResult, error) {
	if !strings.HasPrefix(token, linodemetadata.LinodeAuthenticationTokenPrefix) {
		return nil, bootstrap.ErrNotThisVerifier
	}

	instanceIDString := strings.TrimPrefix(token, linodemetadata.LinodeAuthenticationTokenPrefix)
	instanceID, err := strconv.Atoi(instanceIDString)
	if err != nil {
		return nil, fmt.Errorf("invalid authorization token")
	}

	instance, err := v.client.GetInstance(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get info for Linode (Akamai) instance %q: %w", instanceIDString, err)
	}
	if instance == nil {
		return nil, fmt.Errorf("failed to get info for Linode (Akamai) instance %q: empty response", instanceIDString)
	}

	addresses, challengeEndpoints := gatherIPv4Addresses(instance.IPv4)
	if len(challengeEndpoints) == 0 {
		return nil, fmt.Errorf("cannot determine challenge endpoint for instance id: %s", instanceIDString)
	}

	result := &bootstrap.VerifyResult{
		NodeName:          instance.Label,
		InstanceGroupName: instanceGroupNameFromTags(instance.Tags),
		CertificateNames:  addresses,
		ChallengeEndpoint: challengeEndpoints[0],
	}

	return result, nil
}

// gatherIPv4Addresses returns a list of IPv4 addresses and challenge endpoints from the given list of IPs.
func gatherIPv4Addresses(ips []net.IP) ([]string, []string) {
	addresses := make([]string, 0, len(ips))
	challengeEndpoints := make([]string, 0, len(ips))

	for _, ip := range ips {
		if ip.To4() == nil {
			continue
		}

		ipString := ip.String()
		addresses = append(addresses, ipString)

		if ip.IsPrivate() {
			challengeEndpoints = append(challengeEndpoints, net.JoinHostPort(ipString, strconv.Itoa(wellknownports.NodeupChallenge)))
		}
	}

	return addresses, challengeEndpoints
}

// instanceGroupNameFromTags returns the instance group name from the given list of Linode (Akamai) tags.
func instanceGroupNameFromTags(tags []string) string {
	for _, tag := range tags {
		if after, ok := strings.CutPrefix(tag, TagKubernetesInstanceGroup+":"); ok {
			return after
		}
	}
	return ""
}
