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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseBearerChallenge(t *testing.T) {
	grid := []struct {
		challenge       string
		expectedRealm   string
		expectedService string
		expectError     bool
	}{
		{
			challenge:       `Bearer realm="https://auth.example.com/token",service="registry.example.com"`,
			expectedRealm:   "https://auth.example.com/token",
			expectedService: "registry.example.com",
		},
		{
			challenge:       `Bearer realm="https://registry.example.com/token",service="registry.example.com",scope="repository:owner/repo:pull"`,
			expectedRealm:   "https://registry.example.com/token",
			expectedService: "registry.example.com",
		},
		{
			challenge:     `Bearer realm="https://example.com/token"`,
			expectedRealm: "https://example.com/token",
		},
		{
			// The authentication scheme is case-insensitive.
			challenge:       `bearer realm="https://example.com/token",service="example.com"`,
			expectedRealm:   "https://example.com/token",
			expectedService: "example.com",
		},
		{
			challenge:   `Basic realm="registry"`,
			expectError: true,
		},
		{
			challenge:   `Bearer service="example.com"`,
			expectError: true,
		},
		{
			challenge:   "",
			expectError: true,
		},
	}

	for _, g := range grid {
		t.Run(g.challenge, func(t *testing.T) {
			realm, service, err := parseBearerChallenge(g.challenge)
			if g.expectError {
				if err == nil {
					t.Fatalf("expected an error, got realm %q, service %q", realm, service)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if realm != g.expectedRealm {
				t.Errorf("unexpected realm: expected %q, but got %q", g.expectedRealm, realm)
			}
			if service != g.expectedService {
				t.Errorf("unexpected service: expected %q, but got %q", g.expectedService, service)
			}
		})
	}
}

func TestAnonymousPullToken(t *testing.T) {
	// Registries return the token as "token" (Docker registry auth) or "access_token" (OAuth2).
	for _, field := range []string{"token", "access_token"} {
		t.Run(field, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("service"); got != "registry.example.com" {
					t.Errorf("unexpected service in token request: %q", got)
				}
				if got := r.URL.Query().Get("scope"); got != "repository:assets/nodeup:pull" {
					t.Errorf("unexpected scope in token request: %q", got)
				}
				// A query already present in the realm must be preserved.
				if got := r.URL.Query().Get("issuer"); got != "test" {
					t.Errorf("unexpected issuer in token request: %q", got)
				}
				fmt.Fprintf(w, `{"%s":"secret","expires_in":300}`, field)
			}))
			defer server.Close()

			challenge := fmt.Sprintf(`Bearer realm="%s/token?issuer=test",service="registry.example.com"`, server.URL)
			token, err := anonymousPullToken(context.Background(), server.Client(), challenge, "assets/nodeup")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if token != "secret" {
				t.Errorf("unexpected token: %q", token)
			}
		})
	}
}
