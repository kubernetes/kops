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

// openOCIBlob opens a stream for an OCI registry blob addressed by digest. The URL has the form
// oci://<registry>/<repository>; the digest is the sha256 hash of the asset. The registry must
// allow anonymous pulls of the asset.
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

	blobURL := fmt.Sprintf("https://%s/v2/%s/blobs/sha256:%s", registry, repository, hash.Hex())
	httpClient := newDownloadHTTPClient()

	response, err := getOCIBlob(ctx, httpClient, blobURL, "")
	if err != nil {
		return nil, err
	}
	if response.StatusCode == http.StatusUnauthorized {
		// Anonymous pulls may still need a token, obtained anonymously from the endpoint advertised in
		// the unauthorized response's challenge.
		challenge := response.Header.Get("WWW-Authenticate")
		response.Body.Close()
		token, err := anonymousPullToken(ctx, httpClient, challenge, repository)
		if err != nil {
			return nil, fmt.Errorf("getting anonymous pull token for registry %q: %w", registry, err)
		}
		response, err = getOCIBlob(ctx, httpClient, blobURL, "Bearer "+token)
		if err != nil {
			return nil, err
		}
	}
	if response.StatusCode < 200 || response.StatusCode > 299 {
		response.Body.Close()
		return nil, fmt.Errorf("unexpected response from %q: HTTP %s", blobURL, response.Status)
	}
	return response.Body, nil
}

func getOCIBlob(ctx context.Context, httpClient *http.Client, blobURL, auth string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, blobURL, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	response, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error doing HTTP fetch of %q: %w", blobURL, err)
	}
	return response, nil
}

// anonymousPullToken obtains a pull token for an anonymous client from the token endpoint
// advertised in a registry's WWW-Authenticate challenge.
func anonymousPullToken(ctx context.Context, httpClient *http.Client, challenge, repository string) (string, error) {
	realm, service, err := parseBearerChallenge(challenge)
	if err != nil {
		return "", err
	}

	// Merge the service and scope parameters into any query the realm already carries.
	realmURL, err := url.Parse(realm)
	if err != nil {
		return "", fmt.Errorf("cannot parse realm %q: %w", realm, err)
	}
	query := realmURL.Query()
	query.Set("service", service)
	query.Set("scope", "repository:"+repository+":pull")
	realmURL.RawQuery = query.Encode()

	tokenURL := realmURL.String()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, tokenURL, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create request: %w", err)
	}

	response, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error querying %q: %w", tokenURL, err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return "", fmt.Errorf("unexpected response from %q: HTTP %s", tokenURL, response.Status)
	}

	// Registries return the token as "token" (Docker registry auth) or "access_token" (OAuth2).
	return tokenFromJSON(response.Body, "token", "access_token")
}

// parseBearerChallenge extracts the realm and service from a Bearer WWW-Authenticate challenge,
// such as `Bearer realm="https://auth.example.com/token",service="example.com"`.
func parseBearerChallenge(challenge string) (string, string, error) {
	// The authentication scheme is case-insensitive (RFC 9110).
	const prefix = "Bearer "
	if len(challenge) < len(prefix) || !strings.EqualFold(challenge[:len(prefix)], prefix) {
		return "", "", fmt.Errorf("unsupported WWW-Authenticate challenge %q", challenge)
	}
	params := challenge[len(prefix):]

	var realm, service string
	for _, param := range strings.Split(params, ",") {
		key, value, found := strings.Cut(strings.TrimSpace(param), "=")
		if !found {
			continue
		}
		switch key {
		case "realm":
			realm = strings.Trim(value, `"`)
		case "service":
			service = strings.Trim(value, `"`)
		}
	}
	if realm == "" {
		return "", "", fmt.Errorf("WWW-Authenticate challenge %q does not contain a realm", challenge)
	}
	return realm, service, nil
}

func tokenFromJSON(r io.Reader, fields ...string) (string, error) {
	var body map[string]any
	if err := json.NewDecoder(r).Decode(&body); err != nil {
		return "", fmt.Errorf("cannot decode response: %w", err)
	}
	for _, field := range fields {
		if token, _ := body[field].(string); token != "" {
			return token, nil
		}
	}
	return "", fmt.Errorf("response did not contain %s", strings.Join(fields, " or "))
}
