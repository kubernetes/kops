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

package discovery

import (
	"context"
	"sync"

	api "k8s.io/kops/discovery/apis/discovery.kops.k8s.io/v1alpha1"
)

type NamespacedName struct {
	Namespace string
	Name      string
}

// Universe represents an isolated scope defined by a CA Public Key.
type Universe struct {
	ID                 string                                    `json:"id"`
	DiscoveryEndpoints map[NamespacedName]*api.DiscoveryEndpoint `json:"endpoints"`
	mu                 sync.RWMutex
}

// Store is the interface for storage backends.
type Store interface {
	// UpsertDiscoveryEndpoint creates or updates an endpoint in a specific universe
	UpsertDiscoveryEndpoint(ctx context.Context, universeID string, ep *api.DiscoveryEndpoint) error

	// ListDiscoveryEndpoints lists all endpoints in a specific universe
	ListDiscoveryEndpoints(ctx context.Context, universeID string) ([]*api.DiscoveryEndpoint, error)

	// GetDiscoveryEndpoint retrieves a specific endpoint
	GetDiscoveryEndpoint(ctx context.Context, universeID string, ns, name string) (*api.DiscoveryEndpoint, error)
}
