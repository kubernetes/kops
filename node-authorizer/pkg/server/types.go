/*
Copyright 2019 The Kubernetes Authors.

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

package server

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
)

const (
	// bootstrapTokenIDKey is the id of this token. This can be transmitted in the
	// clear and encoded in the name of the secret. It must be a random 6 character
	// string that matches the regexp `^([a-z0-9]{6})$`. Required.
	bootstrapTokenIDKey = "token-id"

	// bootstrapTokenSecretKey is the actual secret. It must be a random 16 character
	// string that matches the regexp `^([a-z0-9]{16})$`. Required.
	bootstrapTokenSecretKey = "token-secret"

	// bootstrapTokenExpirationKey is when this token should be expired and no
	// longer used. A controller will delete this resource after this time. This
	// is an absolute UTC time using RFC3339. If this cannot be parsed, the token
	// should be considered invalid. Optional.
	bootstrapTokenExpirationKey = "expiration"

	// bootstrapTokenUsagePrefix is the prefix for the other usage constants that specifies different
	// functions of a bootstrap token
	bootstrapTokenUsagePrefix = "usage-bootstrap-"

	// bootstrapTokenSecretPrefix is the prefix for bootstrap token names.
	// Bootstrap tokens secrets must be named in the form
	// `bootstrap-token-<token-id>`.  This is the prefix to be used before the
	// token ID.
	bootstrapTokenSecretPrefix = "bootstrap-token-"

	// secretTypeBootstrapToken is used during the automated bootstrap process (first
	// implemented by kubeadm). It stores tokens that are used to sign well known
	// ConfigMaps. They may also eventually be used for authentication.
	secretTypeBootstrapToken v1.SecretType = "bootstrap.kubernetes.io/token"
)

// Config is the configuration for the service
type Config struct {
	// AuthorizationTimeout is the max duration for a authorization
	AuthorizationTimeout time.Duration
	// ClusterTag is the cloud tag key used to identity the cluster
	ClusterTag string
	// Features is arbitrary feature set for a authorizer
	Features []string
	// EnableVerbose indicate verbose logging
	EnableVerbose bool
	// ClientCommonName is the common name on the client certificate if mutual tls is enabled
	ClientCommonName string
	// ClusterName is the name of the kubernetes cluster
	ClusterName string
	// Listen is the interacted to bind to
	Listen string
	// TokenDuration is the expiration of a bootstrap token
	TokenDuration time.Duration
	// TLSCertPath is the path to the server TLS certificate
	TLSCertPath string
	// TLSClientCAPath is the path to a certificate authority
	TLSClientCAPath string
	// TLSPrivateKeyPath is the path to the private key
	TLSPrivateKeyPath string
}

// UseFeature indicates a feature is in use
func (c *Config) UseFeature(name string) bool {
	if len(c.Features) <= 0 {
		return false
	}

	for _, x := range c.Features {
		items := strings.Split(x, "=")
		if items[0] != name {
			continue
		}
		if len(items) == 1 {
			return true
		}

		v, err := strconv.ParseBool(items[1])
		if err != nil {
			return false
		}
		return v
	}

	return false
}

// NodeRegistration is an incoming request
type NodeRegistration struct {
	// Spec is the request specification
	Spec NodeRegistrationSpec
	// Status is the result of a admission
	Status NodeRegistrationStatus
}

// Deny marks the request as denied and adds the reason why
func (n *NodeRegistration) Deny(reason string) {
	n.Status.Allowed = false
	n.Status.Reason = reason
}

// IsAllowed checks if the request if allowed
func (n *NodeRegistration) IsAllowed() bool {
	return n.Status.Allowed
}

// Token defines a bootstrap token
type Token struct {
	// ID is the id of the token
	ID string
	// Secret is the secret of the token
	Secret string
}

// NodeRegistrationSpec is the node request specification
type NodeRegistrationSpec struct {
	// NodeName is the name of the node
	NodeName string
	// RemoteAddr is the address of the requester
	RemoteAddr string
	// Request is the request body
	Request []byte
}

// NodeRegistrationStatus is result of a authorization
type NodeRegistrationStatus struct {
	// Allowed indicates the request is permitted
	Allowed bool
	// Token is the bootstrap token
	Token string
	// Reason is the reason for the error if any
	Reason string
}

// Authorizer is the generic means to authorize the incoming node request
type Authorizer interface {
	// Admit is responsible for checking if the request is permitted
	Authorize(context.Context, *NodeRegistration) error
	// Close provides a signal to close of resources
	Close() error
	// Name returns the name of the authorizer
	Name() string
}

// Verifier is the client side of authorizer
type Verifier interface {
	// VerifyIdentity is responsible for constructing the parameters for a request
	VerifyIdentity(context.Context) ([]byte, error)
}

// IsValid checks the configuration options
func (c *Config) IsValid() error {
	if c.ClusterName == "" {
		return errors.New("no cluster name")
	}
	if c.Listen == "" {
		return errors.New("no interface to bind specified")
	}
	if c.TLSCertPath == "" {
		return errors.New("no tls certificate")
	}
	if c.TLSPrivateKeyPath == "" {
		return errors.New("no private key")
	}

	return nil
}
