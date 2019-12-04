/*
Copyright 2017 The Kubernetes Authors.

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

package gossip

import (
	"fmt"
	"sync"
)

type GossipStateSnapshot struct {
	Values  map[string]string
	Version uint64
}

type GossipState interface {
	Snapshot() *GossipStateSnapshot
	UpdateValues(removeKeys []string, putKeys map[string]string) error
	Start() error
}

// MultiGossipState enables ramping between gossip mechanisms. This will replicaet
// all UpdateValue operations to the Secondary while still calling the primary for
// all Snapshot information
type MultiGossipState struct {
	Primary   GossipState
	Secondary GossipState
}

func (m *MultiGossipState) Snapshot() *GossipStateSnapshot {
	return m.Primary.Snapshot()
}

func (m *MultiGossipState) UpdateValues(removeKeys []string, putKeys map[string]string) error {
	err := m.Primary.UpdateValues(removeKeys, putKeys)
	m.Secondary.UpdateValues(removeKeys, putKeys)
	return err
}

func (m *MultiGossipState) Start() error {
	errCh := make(chan error, 2)

	go func() {
		errCh <- m.Primary.Start()
	}()

	go func() {
		errCh <- m.Secondary.Start()
	}()

	return <-errCh
}

type newGossipFunc func(listen, channelName, gossipName string, gossipSecret []byte, gossipSeeds SeedProvider) (GossipState, error)

var gossipMap = make(map[string]newGossipFunc)
var gossipMapMutex sync.Mutex

func Register(name string, f newGossipFunc) {
	gossipMapMutex.Lock()
	defer gossipMapMutex.Unlock()
	_, ok := gossipMap[name]
	if ok {
		panic("Duplicate gossip name: " + name)
	}
	gossipMap[name] = f
}

func GetGossipState(protocol, listen, channelName, gossipName string, gossipSecret []byte, gossipSeeds SeedProvider) (GossipState, error) {
	gossipMapMutex.Lock()
	f, ok := gossipMap[protocol]
	gossipMapMutex.Unlock()
	if !ok {
		return nil, fmt.Errorf("Unknown gossip protocol: %s", protocol)
	}

	return f(listen, channelName, gossipName, gossipSecret, gossipSeeds)
}
