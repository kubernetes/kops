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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/wellknownpaths"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"sigs.k8s.io/yaml"
)

// healthCheckPaths are the kube-apiserver endpoints that must be reachable without
// credentials: the kubelet probes and the load balancer health checks call them.
// Anonymous requests are permitted only for these exact paths.
var healthCheckPaths = []string{"/healthz", "/livez", "/readyz"}

// The types below mirror the subset of apiserver.config.k8s.io/v1 AuthenticationConfiguration
// that kOps generates. They are defined locally rather than imported from k8s.io/apiserver,
// which would pull a large dependency into nodeup for a few fields.

type authenticationConfiguration struct {
	APIVersion string               `json:"apiVersion"`
	Kind       string               `json:"kind"`
	JWT        []jwtAuthenticator   `json:"jwt,omitempty"`
	Anonymous  *anonymousAuthConfig `json:"anonymous,omitempty"`
}

type anonymousAuthConfig struct {
	Enabled    bool                     `json:"enabled"`
	Conditions []anonymousAuthCondition `json:"conditions,omitempty"`
}

type anonymousAuthCondition struct {
	Path string `json:"path"`
}

type jwtAuthenticator struct {
	Issuer               jwtIssuer             `json:"issuer"`
	ClaimValidationRules []claimValidationRule `json:"claimValidationRules,omitempty"`
	ClaimMappings        claimMappings         `json:"claimMappings"`
}

type jwtIssuer struct {
	URL                  string   `json:"url"`
	CertificateAuthority string   `json:"certificateAuthority,omitempty"`
	Audiences            []string `json:"audiences"`
}

type claimMappings struct {
	Username prefixedClaim  `json:"username"`
	Groups   *prefixedClaim `json:"groups,omitempty"`
}

type prefixedClaim struct {
	Claim string `json:"claim"`
	// Prefix must be present whenever Claim is set; the empty string means "no prefix".
	Prefix *string `json:"prefix"`
}

type claimValidationRule struct {
	Claim         string `json:"claim"`
	RequiredValue string `json:"requiredValue"`
}

// buildAuthenticationConfiguration writes the kube-apiserver AuthenticationConfiguration file and
// points --authentication-config at it.
//
// The file always configures the anonymous authenticator so that unauthenticated requests are
// accepted only for the health check endpoints. kube-apiserver has no separate unauthenticated
// health port, so this is what lets the kubelet probes and the load balancer health checks reach
// /healthz, /livez and /readyz while anonymous access to everything else stays disabled.
//
// Because kube-apiserver rejects --anonymous-auth and the --oidc-* flags once an authentication
// configuration file is in use, the OIDC settings are translated into a JWT authenticator in the
// same file and the flags are cleared. If the cluster spec points at a user-supplied
// authentication configuration file, that file is read on the node and the anonymous section is
// added to it, so existing files keep working without changes.
func (b *KubeAPIServerBuilder) buildAuthenticationConfiguration(c *fi.NodeupModelBuilderContext, pathSrvKAPI string, kubeAPIServer *kops.KubeAPIServerConfig) error {
	anonymous := &anonymousAuthConfig{Enabled: true}
	if fi.ValueOf(kubeAPIServer.AnonymousAuth) {
		// The cluster explicitly enables anonymous auth for all requests; keep that behavior.
		klog.Warningf("spec.kubeAPIServer.anonymousAuth is true; anonymous requests are permitted for all kube-apiserver endpoints")
	} else {
		for _, path := range healthCheckPaths {
			anonymous.Conditions = append(anonymous.Conditions, anonymousAuthCondition{Path: path})
		}
	}
	// The anonymous section of the file replaces the flag, and kube-apiserver refuses to start
	// when both are set.
	kubeAPIServer.AnonymousAuth = nil

	userConfigFile := kubeAPIServer.AuthenticationConfigFile
	oidcCAFile := fi.ValueOf(kubeAPIServer.OIDCCAFile)
	jwt := buildJWTAuthenticator(kubeAPIServer)
	// The OIDC flags are mutually exclusive with --authentication-config.
	kubeAPIServer.OIDCCAFile = nil
	kubeAPIServer.OIDCClientID = nil
	kubeAPIServer.OIDCGroupsClaim = nil
	kubeAPIServer.OIDCGroupsPrefix = nil
	kubeAPIServer.OIDCIssuerURL = nil
	kubeAPIServer.OIDCRequiredClaim = nil
	kubeAPIServer.OIDCUsernameClaim = nil
	kubeAPIServer.OIDCUsernamePrefix = nil

	// Cluster validation rejects user files at this path, which would make the task below wait for itself.
	configFile := filepath.Join(pathSrvKAPI, wellknownpaths.KubeAPIServerAuthenticationConfigFileName)
	kubeAPIServer.AuthenticationConfigFile = configFile

	// The v1 API of AuthenticationConfiguration was added in Kubernetes 1.34. Earlier releases
	// only serve v1beta1, which has the same anonymous and jwt fields.
	apiVersion := "apiserver.config.k8s.io/v1"
	if b.IsKubernetesLT("1.34") {
		apiVersion = "apiserver.config.k8s.io/v1beta1"
	}

	var contents fi.Resource
	var afterFiles []string
	switch {
	case userConfigFile != "":
		// The user-supplied file may be written by nodeup in this same run (via fileAssets),
		// so read and merge it only when the task runs, after that file is in place.
		afterFiles = append(afterFiles, userConfigFile)
		contents = fi.FunctionToResource(func() ([]byte, error) {
			userConfig, err := os.ReadFile(userConfigFile)
			if err != nil {
				return nil, fmt.Errorf("reading authentication config file %q: %w", userConfigFile, err)
			}
			return mergeAnonymousAuthConfig(userConfig, anonymous)
		})

	case jwt != nil && oidcCAFile != "":
		// Same consideration for the OIDC CA file; the structured configuration wants the
		// certificate contents rather than a path.
		afterFiles = append(afterFiles, oidcCAFile)
		contents = fi.FunctionToResource(func() ([]byte, error) {
			ca, err := os.ReadFile(oidcCAFile)
			if err != nil {
				return nil, fmt.Errorf("reading OIDC CA file %q: %w", oidcCAFile, err)
			}
			jwt.Issuer.CertificateAuthority = string(ca)
			return marshalAuthenticationConfiguration(apiVersion, jwt, anonymous)
		})

	default:
		data, err := marshalAuthenticationConfiguration(apiVersion, jwt, anonymous)
		if err != nil {
			return err
		}
		contents = fi.NewBytesResource(data)
	}

	c.AddTask(&nodetasks.File{
		Path:       configFile,
		Contents:   contents,
		Type:       nodetasks.FileType_File,
		Mode:       s("0400"),
		AfterFiles: afterFiles,
	})

	return nil
}

// buildJWTAuthenticator translates the OIDC flags into a structured JWT authenticator, following
// the same rules kube-apiserver applies when converting the legacy --oidc-* flags.
func buildJWTAuthenticator(kubeAPIServer *kops.KubeAPIServerConfig) *jwtAuthenticator {
	issuerURL := fi.ValueOf(kubeAPIServer.OIDCIssuerURL)
	clientID := fi.ValueOf(kubeAPIServer.OIDCClientID)
	if issuerURL == "" || clientID == "" {
		return nil
	}

	// The flag defaults to "sub" when no claim is specified.
	usernameClaim := "sub"
	if kubeAPIServer.OIDCUsernameClaim != nil {
		usernameClaim = *kubeAPIServer.OIDCUsernameClaim
	}
	usernamePrefix := fi.ValueOf(kubeAPIServer.OIDCUsernamePrefix)
	if usernamePrefix == "" && usernameClaim != "email" {
		// Legacy flag behavior: claims other than "email" are prefixed with the issuer URL
		// unless a prefix is given. See https://github.com/kubernetes/kubernetes/issues/31380
		usernamePrefix = issuerURL + "#"
	}
	if usernamePrefix == "-" {
		// Special value indicating usernames shouldn't be prefixed.
		usernamePrefix = ""
	}

	jwt := &jwtAuthenticator{
		Issuer: jwtIssuer{
			URL:       issuerURL,
			Audiences: []string{clientID},
		},
		ClaimMappings: claimMappings{
			Username: prefixedClaim{
				Claim:  usernameClaim,
				Prefix: new(usernamePrefix),
			},
		},
	}

	if groupsClaim := fi.ValueOf(kubeAPIServer.OIDCGroupsClaim); groupsClaim != "" {
		jwt.ClaimMappings.Groups = &prefixedClaim{
			Claim:  groupsClaim,
			Prefix: new(fi.ValueOf(kubeAPIServer.OIDCGroupsPrefix)),
		}
	}

	for _, requiredClaim := range kubeAPIServer.OIDCRequiredClaim {
		claim, value, _ := strings.Cut(requiredClaim, "=")
		jwt.ClaimValidationRules = append(jwt.ClaimValidationRules, claimValidationRule{
			Claim:         claim,
			RequiredValue: value,
		})
	}

	return jwt
}

func marshalAuthenticationConfiguration(apiVersion string, jwt *jwtAuthenticator, anonymous *anonymousAuthConfig) ([]byte, error) {
	config := &authenticationConfiguration{
		APIVersion: apiVersion,
		Kind:       "AuthenticationConfiguration",
		Anonymous:  anonymous,
	}
	if jwt != nil {
		config.JWT = []jwtAuthenticator{*jwt}
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("marshaling authentication configuration: %w", err)
	}
	return data, nil
}

// mergeAnonymousAuthConfig adds the anonymous section to a user-supplied AuthenticationConfiguration.
// Everything else in the file is passed through untouched. A file that already configures
// anonymous auth is kept as-is, since --anonymous-auth is no longer passed and the file is the
// only place where that setting lives.
func mergeAnonymousAuthConfig(userConfig []byte, anonymous *anonymousAuthConfig) ([]byte, error) {
	var config map[string]interface{}
	if err := yaml.Unmarshal(userConfig, &config); err != nil {
		return nil, fmt.Errorf("parsing authentication configuration: %w", err)
	}
	if config == nil {
		return nil, fmt.Errorf("authentication configuration is empty")
	}
	if kind, _ := config["kind"].(string); kind != "AuthenticationConfiguration" {
		return nil, fmt.Errorf("authentication configuration has kind %q, expected AuthenticationConfiguration", kind)
	}

	if _, found := config["anonymous"]; found {
		klog.Warningf("authentication configuration already configures anonymous auth; kube-apiserver health checks need anonymous access to %s", strings.Join(healthCheckPaths, ", "))
	} else {
		// Round-trip through YAML so that the section has the same shape as the rest of the map.
		encoded, err := yaml.Marshal(anonymous)
		if err != nil {
			return nil, fmt.Errorf("marshaling anonymous auth configuration: %w", err)
		}
		var decoded interface{}
		if err := yaml.Unmarshal(encoded, &decoded); err != nil {
			return nil, fmt.Errorf("unmarshaling anonymous auth configuration: %w", err)
		}
		config["anonymous"] = decoded
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("marshaling authentication configuration: %w", err)
	}
	return data, nil
}
