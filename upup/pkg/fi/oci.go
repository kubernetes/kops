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

package fi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"k8s.io/kops/util/pkg/hashing"
)

// openOCIBlob opens a stream for an OCI registry blob addressed by digest.
// The URL has the form oci://<registry>/<repository>; the digest is the sha256
// hash of the asset. The registry is expected to be an Azure Container Registry,
// authenticated with the instance's managed identity.
func openOCIBlob(ctx context.Context, u *url.URL, hash *hashing.Hash) (io.ReadCloser, error) {
	if hash == nil {
		return nil, fmt.Errorf("OCI asset %q requires a known hash", u)
	}
	if hash.Algorithm != hashing.HashAlgorithmSHA256 {
		return nil, fmt.Errorf("OCI asset %q requires a sha256 hash, got %q", u, hash.Algorithm)
	}

	registry := u.Host
	repository := strings.Trim(u.Path, "/")
	if registry == "" || repository == "" {
		return nil, fmt.Errorf("cannot parse OCI asset URL %q; expected oci://<registry>/<repository>", u)
	}

	token, err := acrPullToken(ctx, registry, repository)
	if err != nil {
		return nil, fmt.Errorf("getting pull token for registry %q: %w", registry, err)
	}

	blobURL := fmt.Sprintf("https://%s/v2/%s/blobs/sha256:%s", registry, repository, hash.Hex())

	httpClient := newDownloadHTTPClient()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, blobURL, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	response, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error doing HTTP fetch of %q: %w", blobURL, err)
	}
	if response.StatusCode < 200 || response.StatusCode > 299 {
		response.Body.Close()
		return nil, fmt.Errorf("unexpected response from %q: HTTP %s", blobURL, response.Status)
	}
	return response.Body, nil
}

// ACRDockerCredentials returns docker-login style credentials for an Azure
// Container Registry, authenticating with the instance's managed identity.
// The registry accepts its refresh token as the password for this well-known user.
func ACRDockerCredentials(ctx context.Context, registry string) (string, string, error) {
	aadToken, err := azureInstanceIdentityToken(ctx)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := acrOAuthRequest(ctx,
		fmt.Sprintf("https://%s/oauth2/exchange", registry),
		url.Values{
			"grant_type":   []string{"access_token"},
			"service":      []string{registry},
			"access_token": []string{aadToken},
		},
		"refresh_token")
	if err != nil {
		return "", "", fmt.Errorf("exchanging identity token for a registry refresh token: %w", err)
	}

	return "00000000-0000-0000-0000-000000000000", refreshToken, nil
}

// acrPullToken exchanges the instance's managed-identity token for an Azure
// Container Registry access token scoped to pulling from the given repository.
func acrPullToken(ctx context.Context, registry, repository string) (string, error) {
	aadToken, err := azureInstanceIdentityToken(ctx)
	if err != nil {
		return "", err
	}

	refreshToken, err := acrOAuthRequest(ctx,
		fmt.Sprintf("https://%s/oauth2/exchange", registry),
		url.Values{
			"grant_type":   []string{"access_token"},
			"service":      []string{registry},
			"access_token": []string{aadToken},
		},
		"refresh_token")
	if err != nil {
		return "", fmt.Errorf("exchanging identity token for a registry refresh token: %w", err)
	}

	accessToken, err := acrOAuthRequest(ctx,
		fmt.Sprintf("https://%s/oauth2/token", registry),
		url.Values{
			"grant_type":    []string{"refresh_token"},
			"service":       []string{registry},
			"scope":         []string{fmt.Sprintf("repository:%s:pull", repository)},
			"refresh_token": []string{refreshToken},
		},
		"access_token")
	if err != nil {
		return "", fmt.Errorf("getting a registry access token: %w", err)
	}

	return accessToken, nil
}

// azureInstanceIdentityToken returns a managed-identity token from the Azure
// instance metadata service.
func azureInstanceIdentityToken(ctx context.Context) (string, error) {
	imdsURL := "http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=" + url.QueryEscape("https://management.azure.com/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imdsURL, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create request: %w", err)
	}
	req.Header.Set("Metadata", "true")

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error querying the instance metadata service: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return "", fmt.Errorf("unexpected response from the instance metadata service: HTTP %s", response.Status)
	}

	return tokenFromJSON(response.Body, "access_token")
}

func acrOAuthRequest(ctx context.Context, endpoint string, values url.Values, tokenField string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return "", fmt.Errorf("cannot create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error posting to %q: %w", endpoint, err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return "", fmt.Errorf("unexpected response from %q: HTTP %s", endpoint, response.Status)
	}

	return tokenFromJSON(response.Body, tokenField)
}

func tokenFromJSON(r io.Reader, field string) (string, error) {
	var body map[string]any
	if err := json.NewDecoder(r).Decode(&body); err != nil {
		return "", fmt.Errorf("cannot decode response: %w", err)
	}
	token, _ := body[field].(string)
	if token == "" {
		return "", fmt.Errorf("response did not contain %q", field)
	}
	return token, nil
}
