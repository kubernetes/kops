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

package mesh

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/weaveworks/mesh"
	"k8s.io/klog"
	"k8s.io/kops/protokube/pkg/gossip"
)

func init() {
	gossip.Register("mesh", func(listen, channelName, gossipName string, gossipSecret []byte, gossipSeeds gossip.SeedProvider) (gossip.GossipState, error) {
		return NewMeshGossiper(listen, channelName, gossipName, gossipSecret, gossipSeeds)
	})
}

type MeshGossiper struct {
	seeds gossip.SeedProvider

	router *mesh.Router
	peer   *peer

	version uint64
}

func NewMeshGossiper(listen string, channelName string, nodeName string, password []byte, seeds gossip.SeedProvider) (*MeshGossiper, error) {

	connLimit := 0 // 0 means no limit
	gossipDnsConnLimit := os.Getenv("GOSSIP_DNS_CONN_LIMIT")
	if gossipDnsConnLimit != "" {
		limit, err := strconv.Atoi(gossipDnsConnLimit)
		if err != nil {
			// Continue with the default value
			klog.Warningf("cannot parse env GOSSIP_DNS_CONN_LIMIT value %q", gossipDnsConnLimit)
		} else {
			connLimit = limit
		}
	}

	klog.Infof("gossip dns connection limit is:%d", connLimit)

	meshConfig := mesh.Config{
		ProtocolMinVersion: mesh.ProtocolMinVersion,
		Password:           password,
		ConnLimit:          connLimit,
		PeerDiscovery:      true,
		//TrustedSubnets:     []*net.IPNet{},
	}

	{
		host, portString, err := net.SplitHostPort(listen)
		if err != nil {
			return nil, fmt.Errorf("cannot parse -listen flag: %v", listen)
		}
		port, err := strconv.Atoi(portString)
		if err != nil {
			return nil, fmt.Errorf("cannot parse -listen flag: %v", listen)
		}
		meshConfig.Host = host
		meshConfig.Port = port
	}

	meshName, err := mesh.PeerNameFromUserInput(nodeName)
	if err != nil {
		return nil, fmt.Errorf("error parsing peer name: %v", err)
	}

	nickname := nodeName
	logger := &glogLogger{}
	router := mesh.NewRouter(meshConfig, meshName, nickname, mesh.NullOverlay{}, logger)

	peer := newPeer(meshName)
	gossip := router.NewGossip(channelName, peer)
	peer.register(gossip)

	gossiper := &MeshGossiper{
		seeds:  seeds,
		router: router,
		peer:   peer,
	}
	return gossiper, nil
}

func (g *MeshGossiper) Start() error {
	//klog.Infof("mesh router starting (%s)", *meshListen)
	g.router.Start()

	defer func() {
		klog.Infof("mesh router stopping")
		g.router.Stop()
	}()

	g.runSeeding()

	return nil
}

func (g *MeshGossiper) runSeeding() {
	for {
		klog.V(2).Infof("Querying for seeds")

		seeds, err := g.seeds.GetSeeds()
		if err != nil {
			klog.Warningf("error getting seeds: %v", err)
			time.Sleep(1 * time.Minute)
			continue
		}

		klog.Infof("Got seeds: %s", seeds)
		// TODO: Include ourselves?  Exclude ourselves?

		removeOthers := false
		errors := g.router.ConnectionMaker.InitiateConnections(seeds, removeOthers)

		if len(errors) != 0 {
			for _, err := range errors {
				klog.Infof("error connecting to seeds: %v", err)
			}

			time.Sleep(1 * time.Minute)
			continue
		}

		klog.V(2).Infof("Seeding successful")

		// Reseed periodically, just in case of partitions
		// TODO: Make it so that only one node polls, or at least statistically get close
		time.Sleep(60 * time.Minute)
	}
}

func (g *MeshGossiper) Snapshot() *gossip.GossipStateSnapshot {
	return g.peer.snapshot()
}

func (g *MeshGossiper) UpdateValues(removeKeys []string, putEntries map[string]string) error {
	klog.V(2).Infof("UpdateValues: remove=%s, put=%s", removeKeys, putEntries)
	return g.peer.updateValues(removeKeys, putEntries)
}
