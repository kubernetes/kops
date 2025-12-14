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
	"time"

	api "k8s.io/kops/discovery/apis/discovery.kops.k8s.io/v1alpha1"
)

type MemoryStore struct {
	universes map[string]*Universe
	mu        sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		universes: make(map[string]*Universe),
	}
}

func (s *MemoryStore) getOrCreateUniverse(id string) *Universe {
	s.mu.Lock()
	defer s.mu.Unlock()
	if u, ok := s.universes[id]; ok {
		return u
	}
	u := &Universe{
		ID:                 id,
		DiscoveryEndpoints: make(map[NamespacedName]*api.DiscoveryEndpoint),
	}
	s.universes[id] = u
	return u
}

func (s *MemoryStore) UpsertDiscoveryEndpoint(ctx context.Context, universeID string, ep *api.DiscoveryEndpoint) error {
	u := s.getOrCreateUniverse(universeID)

	u.mu.Lock()
	defer u.mu.Unlock()

	// Ensure basic metadata is consistent
	ep.TypeMeta.Kind = "DiscoveryEndpoint"
	ep.TypeMeta.APIVersion = "discovery.kops.k8s.io/v1alpha1"

	// Update LastSeen
	ep.Spec.LastSeen = time.Now().Format(time.RFC3339)

	id := NamespacedName{
		Namespace: ep.ObjectMeta.Namespace,
		Name:      ep.ObjectMeta.Name,
	}
	u.DiscoveryEndpoints[id] = ep
	return nil
}

func (s *MemoryStore) ListDiscoveryEndpoints(ctx context.Context, universeID string) ([]*api.DiscoveryEndpoint, error) {
	// For listing, we don't necessarily need to create the universe if it doesn't exist,
	// but for consistency with getOrCreate logic in memory store, checking existence is enough.
	s.mu.RLock()
	u, ok := s.universes[universeID]
	s.mu.RUnlock()

	if !ok {
		return []*api.DiscoveryEndpoint{}, nil
	}

	u.mu.RLock()
	defer u.mu.RUnlock()
	endpoints := make([]*api.DiscoveryEndpoint, 0, len(u.DiscoveryEndpoints))
	for _, n := range u.DiscoveryEndpoints {
		endpoints = append(endpoints, n)
	}
	return endpoints, nil
}

func (s *MemoryStore) GetDiscoveryEndpoint(ctx context.Context, universeID string, ns, name string) (*api.DiscoveryEndpoint, error) {
	s.mu.RLock()
	u, ok := s.universes[universeID]
	s.mu.RUnlock()

	if !ok {
		return nil, nil
	}

	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.DiscoveryEndpoints[NamespacedName{Namespace: ns, Name: name}], nil
}
