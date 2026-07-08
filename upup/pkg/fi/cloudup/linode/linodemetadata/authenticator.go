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

// Package linodemetadata provides the node-local (metadata-service based)
// parts of the Linode support, kept separate so that nodeup does not link the
// full cloud provider implementation.
package linodemetadata

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"k8s.io/kops/pkg/bootstrap"
)

const LinodeAuthenticationTokenPrefix = "x-linode-instance-id " //nolint:gosec // This is an authentication scheme prefix, not a credential.

const (
	linodeMetadataBaseURL  = "http://169.254.169.254"
	linodeMetadataTokenTTL = "300"
)

type linodeAuthenticator struct {
	client          *http.Client
	metadataBaseURL string
}

var _ bootstrap.Authenticator = (*linodeAuthenticator)(nil)

// NewLinodeAuthenticator returns a bootstrap.Authenticator that can create
// authentication tokens for Linode instances by querying the Linode metadata
// service for the instance ID and returning it prefixed with "x-linode-instance-id".
func NewLinodeAuthenticator() (bootstrap.Authenticator, error) {
	return &linodeAuthenticator{
		client:          http.DefaultClient,
		metadataBaseURL: linodeMetadataBaseURL,
	}, nil
}

// CreateToken queries the Linode (Akamai) metadata service for the instance ID and returns
// it prefixed with "x-linode-instance-id ".
func (a *linodeAuthenticator) CreateToken(body []byte) (string, error) {
	instanceID, err := getLinodeMetadataValue(context.TODO(), a.client, a.metadataBaseURL, "id")
	if err != nil {
		return "", fmt.Errorf("unable to fetch Linode (Akamai) instance id: %w", err)
	}

	return LinodeAuthenticationTokenPrefix + instanceID, nil
}

// GetMetadataValue fetches the given field from the Linode instance metadata service
// using the standard metadata endpoint and default HTTP client.
func GetMetadataValue(ctx context.Context, key string) (string, error) {
	return getLinodeMetadataValue(ctx, http.DefaultClient, linodeMetadataBaseURL, key)
}

// getLinodeMetadataValue queries the Linode (Akamai) metadata service for the given key
// and returns the value as a string.
func getLinodeMetadataValue(ctx context.Context, client *http.Client, metadataBaseURL, key string) (string, error) {
	tokenReq, err := http.NewRequestWithContext(ctx, http.MethodPut, metadataBaseURL+"/v1/token", nil)
	if err != nil {
		return "", fmt.Errorf("building metadata token request: %w", err)
	}
	tokenReq.Header.Set("Metadata-Token-Expiry-Seconds", linodeMetadataTokenTTL)

	tokenResp, err := client.Do(tokenReq)
	if err != nil {
		return "", fmt.Errorf("fetching metadata token: %w", err)
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetching metadata token: unexpected status code %d", tokenResp.StatusCode)
	}

	tokenBytes, err := io.ReadAll(tokenResp.Body)
	if err != nil {
		return "", fmt.Errorf("reading metadata token response: %w", err)
	}

	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return "", fmt.Errorf("metadata token was empty")
	}

	instanceReq, err := http.NewRequestWithContext(ctx, http.MethodGet, metadataBaseURL+"/v1/instance", nil)
	if err != nil {
		return "", fmt.Errorf("building instance metadata request: %w", err)
	}
	instanceReq.Header.Set("Metadata-Token", token)

	instanceResp, err := client.Do(instanceReq)
	if err != nil {
		return "", fmt.Errorf("fetching instance metadata: %w", err)
	}
	defer instanceResp.Body.Close()

	if instanceResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetching instance metadata: unexpected status code %d", instanceResp.StatusCode)
	}

	instanceBytes, err := io.ReadAll(instanceResp.Body)
	if err != nil {
		return "", fmt.Errorf("reading instance metadata response: %w", err)
	}

	value := parseLinodeMetadataValue(string(instanceBytes), key)
	if value == "" {
		return "", fmt.Errorf("instance %s from Linode (Akamai) metadata was empty", key)
	}
	return value, nil
}

// parseLinodeMetadataValue parses the Linode (Akamai) metadata response for the given key
// and returns the value as a string.
func parseLinodeMetadataValue(metadata string, key string) string {
	prefix := key + ":"
	for _, line := range strings.Split(metadata, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		return strings.TrimSpace(strings.TrimPrefix(line, prefix))
	}
	return ""
}
