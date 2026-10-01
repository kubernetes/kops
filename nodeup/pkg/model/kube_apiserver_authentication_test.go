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

package model

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/kops/pkg/apis/kops"
)

func TestBuildJWTAuthenticator(t *testing.T) {
	issuer := "https://issuer.example.com"

	grid := []struct {
		name     string
		config   kops.KubeAPIServerConfig
		expected *jwtAuthenticator
	}{
		{
			name:     "no oidc",
			config:   kops.KubeAPIServerConfig{},
			expected: nil,
		},
		{
			name: "issuer without client id is ignored",
			config: kops.KubeAPIServerConfig{
				OIDCIssuerURL: new(issuer),
			},
			expected: nil,
		},
		{
			name: "defaults: sub claim prefixed with the issuer",
			config: kops.KubeAPIServerConfig{
				OIDCIssuerURL: new(issuer),
				OIDCClientID:  new("kubernetes"),
			},
			expected: &jwtAuthenticator{
				Issuer: jwtIssuer{URL: issuer, Audiences: []string{"kubernetes"}},
				ClaimMappings: claimMappings{
					Username: prefixedClaim{Claim: "sub", Prefix: new(issuer + "#")},
				},
			},
		},
		{
			name: "email claim is not prefixed",
			config: kops.KubeAPIServerConfig{
				OIDCIssuerURL:     new(issuer),
				OIDCClientID:      new("kubernetes"),
				OIDCUsernameClaim: new("email"),
			},
			expected: &jwtAuthenticator{
				Issuer: jwtIssuer{URL: issuer, Audiences: []string{"kubernetes"}},
				ClaimMappings: claimMappings{
					Username: prefixedClaim{Claim: "email", Prefix: new("")},
				},
			},
		},
		{
			name: "dash disables the username prefix",
			config: kops.KubeAPIServerConfig{
				OIDCIssuerURL:      new(issuer),
				OIDCClientID:       new("kubernetes"),
				OIDCUsernameClaim:  new("user"),
				OIDCUsernamePrefix: new("-"),
			},
			expected: &jwtAuthenticator{
				Issuer: jwtIssuer{URL: issuer, Audiences: []string{"kubernetes"}},
				ClaimMappings: claimMappings{
					Username: prefixedClaim{Claim: "user", Prefix: new("")},
				},
			},
		},
		{
			name: "all options",
			config: kops.KubeAPIServerConfig{
				OIDCIssuerURL:      new(issuer),
				OIDCClientID:       new("kubernetes"),
				OIDCUsernameClaim:  new("user"),
				OIDCUsernamePrefix: new("oidc:"),
				OIDCGroupsClaim:    new("groups"),
				OIDCGroupsPrefix:   new("oidcgroup:"),
				OIDCRequiredClaim:  []string{"claim1=value1", "claim2=value2"},
			},
			expected: &jwtAuthenticator{
				Issuer: jwtIssuer{URL: issuer, Audiences: []string{"kubernetes"}},
				ClaimValidationRules: []claimValidationRule{
					{Claim: "claim1", RequiredValue: "value1"},
					{Claim: "claim2", RequiredValue: "value2"},
				},
				ClaimMappings: claimMappings{
					Username: prefixedClaim{Claim: "user", Prefix: new("oidc:")},
					Groups:   &prefixedClaim{Claim: "groups", Prefix: new("oidcgroup:")},
				},
			},
		},
		{
			name: "groups claim without prefix gets an empty prefix",
			config: kops.KubeAPIServerConfig{
				OIDCIssuerURL:   new(issuer),
				OIDCClientID:    new("kubernetes"),
				OIDCGroupsClaim: new("groups"),
			},
			expected: &jwtAuthenticator{
				Issuer: jwtIssuer{URL: issuer, Audiences: []string{"kubernetes"}},
				ClaimMappings: claimMappings{
					Username: prefixedClaim{Claim: "sub", Prefix: new(issuer + "#")},
					Groups:   &prefixedClaim{Claim: "groups", Prefix: new("")},
				},
			},
		},
	}

	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			actual := buildJWTAuthenticator(&g.config)
			if diff := cmp.Diff(g.expected, actual); diff != "" {
				t.Errorf("unexpected jwt authenticator (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMarshalAuthenticationConfiguration(t *testing.T) {
	anonymous := &anonymousAuthConfig{
		Enabled: true,
		Conditions: []anonymousAuthCondition{
			{Path: "/healthz"}, {Path: "/livez"}, {Path: "/readyz"},
		},
	}
	jwt := &jwtAuthenticator{
		Issuer: jwtIssuer{URL: "https://issuer.example.com", Audiences: []string{"kubernetes"}},
		ClaimMappings: claimMappings{
			Username: prefixedClaim{Claim: "email", Prefix: new("")},
		},
	}

	actual, err := marshalAuthenticationConfiguration("apiserver.config.k8s.io/v1", jwt, anonymous)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `anonymous:
  conditions:
  - path: /healthz
  - path: /livez
  - path: /readyz
  enabled: true
apiVersion: apiserver.config.k8s.io/v1
jwt:
- claimMappings:
    username:
      claim: email
      prefix: ""
  issuer:
    audiences:
    - kubernetes
    url: https://issuer.example.com
kind: AuthenticationConfiguration
`
	if diff := cmp.Diff(expected, string(actual)); diff != "" {
		t.Errorf("unexpected configuration (-want +got):\n%s", diff)
	}
}

func TestMergeAnonymousAuthConfig(t *testing.T) {
	anonymous := &anonymousAuthConfig{
		Enabled:    true,
		Conditions: []anonymousAuthCondition{{Path: "/healthz"}},
	}

	t.Run("adds anonymous section and keeps the rest", func(t *testing.T) {
		userConfig := `apiVersion: apiserver.config.k8s.io/v1
kind: AuthenticationConfiguration
jwt:
- issuer:
    url: https://issuer.example.com
    audiences: [kubernetes]
  claimMappings:
    username:
      expression: claims.email
  userValidationRules:
  - expression: "!user.username.startsWith('system:')"
    message: "username cannot use reserved system: prefix"
`
		actual, err := mergeAnonymousAuthConfig([]byte(userConfig), anonymous)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := `anonymous:
  conditions:
  - path: /healthz
  enabled: true
apiVersion: apiserver.config.k8s.io/v1
jwt:
- claimMappings:
    username:
      expression: claims.email
  issuer:
    audiences:
    - kubernetes
    url: https://issuer.example.com
  userValidationRules:
  - expression: '!user.username.startsWith(''system:'')'
    message: 'username cannot use reserved system: prefix'
kind: AuthenticationConfiguration
`
		if diff := cmp.Diff(expected, string(actual)); diff != "" {
			t.Errorf("unexpected configuration (-want +got):\n%s", diff)
		}
	})

	t.Run("keeps an existing anonymous section", func(t *testing.T) {
		userConfig := `apiVersion: apiserver.config.k8s.io/v1
kind: AuthenticationConfiguration
anonymous:
  enabled: false
`
		actual, err := mergeAnonymousAuthConfig([]byte(userConfig), anonymous)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(string(actual), "enabled: false") || strings.Contains(string(actual), "/healthz") {
			t.Errorf("expected the user's anonymous section to be kept, got:\n%s", actual)
		}
	})

	t.Run("rejects other kinds", func(t *testing.T) {
		userConfig := `apiVersion: apiserver.config.k8s.io/v1
kind: AuthorizationConfiguration
`
		if _, err := mergeAnonymousAuthConfig([]byte(userConfig), anonymous); err == nil {
			t.Errorf("expected an error for a non-AuthenticationConfiguration document")
		}
	})

	t.Run("rejects empty files", func(t *testing.T) {
		if _, err := mergeAnonymousAuthConfig([]byte(""), anonymous); err == nil {
			t.Errorf("expected an error for an empty document")
		}
	})
}
