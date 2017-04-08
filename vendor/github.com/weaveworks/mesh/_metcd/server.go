package metcd

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/etcd/raft/raftpb"
	"google.golang.org/grpc"

	"github.com/weaveworks/mesh"
	"github.com/weaveworks/mesh/meshconn"
)

// Server collects the etcd V3 server interfaces that we implement.
type Server interface {
	//etcdserverpb.AuthServer
	//etcdserverpb.ClusterServer
	etcdserverpb.KVServer
	//etcdserverpb.LeaseServer
	//etcdserverpb.MaintenanceServer
	//etcdserverpb.WatchServer
}

// GRPCServer converts a metcd.Server to a *grpc.Server.
func GRPCServer(s Server, options ...grpc.ServerOption) *grpc.Server {
	srv := grpc.NewServer(options...)
	//etcdserverpb.RegisterAuthServer(srv, s)
	//etcdserverpb.RegisterClusterServer(srv, s)
	etcdserverpb.RegisterKVServer(srv, s)
	//etcdserverpb.RegisterLeaseServer(srv, s)
	//etcdserverpb.RegisterMaintenanceServer(srv, s)
	//etcdserverpb.RegisterWatchServer(srv, s)
	return srv
}

// NewServer returns a Server that (partially) implements the etcd V3 API.
// It uses the passed mesh components to act as the Raft transport.
// For the moment, it blocks until the mesh has minPeerCount peers.
// (This responsibility should rather be given to the caller.)
// The server can be terminated by certain conditions in the cluster.
// If that happens, terminatedc signaled, and the server is invalid.
func NewServer(
	router *mesh.Router,
	peer *meshconn.Peer,
	minPeerCount int,
	terminatec <-chan struct{},
	terminatedc chan<- error,
	logger mesh.Logger,
) Server {
	c := make(chan Server)
	go serverManager(router, peer, minPeerCount, terminatec, terminatedc, logger, c)
	return <-c
}

// NewDefaultServer is like NewServer, but we take care of creating a
// mesh.Router and meshconn.Peer for you, with sane defaults. If you need more
// fine-grained control, create the components yourself and use NewServer.
func NewDefaultServer(
	minPeerCount int,
	terminatec <-chan struct{},
	terminatedc chan<- error,
	logger mesh.Logger,
) Server {
	var (
		peerName = mustPeerName()
		nickName = mustHostname()
		host     = "0.0.0.0"
		port     = 6379
		password = ""
		channel  = "metcd"
	)
	router := mesh.NewRouter(mesh.Config{
		Host:               host,
		Port:               port,
		ProtocolMinVersion: mesh.ProtocolMinVersion,
		Password:           []byte(password),
		ConnLimit:          64,
		PeerDiscovery:      true,
		TrustedSubnets:     []*net.IPNet{},
	}, peerName, nickName, mesh.NullOverlay{}, logger)

	// Create a meshconn.Peer and connect it to a channel.
	peer := meshconn.NewPeer(router.Ourself.Peer.Name, router.Ourself.UID, logger)
	gossip := router.NewGossip(channel, peer)
	peer.Register(gossip)

	// Start the router and join the mesh.
	// Note that we don't ever stop the router.
	// This may or may not be a problem.
	// TODO(pb): determine if this is a super huge problem
	router.Start()

	return NewServer(router, peer, minPeerCount, terminatec, terminatedc, logger)
}

func serverManager(
	router *mesh.Router,
	peer *meshconn.Peer,
	minPeerCount int,
	terminatec <-chan struct{},
	terminatedc chan<- error,
	logger mesh.Logger,
	out chan<- Server,
) {
	// Identify mesh peers to either create or join a cluster.
	// This algorithm is presently completely insufficient.
	// It suffers from timing failures, and doesn't understand channels.
	// TODO(pb): use gossip to agree on better starting conditions
	var (
		self   = meshconn.MeshAddr{PeerName: router.Ourself.Peer.Name, PeerUID: router.Ourself.UID}
		others = []net.Addr{}
	)
	for {
		others = others[:0]
		for _, desc := range router.Peers.Descriptions() {
			others = append(others, meshconn.MeshAddr{PeerName: desc.Name, PeerUID: desc.UID})
		}
		if len(others) == minPeerCount {
			logger.Printf("detected %d peers; creating", len(others))
			break
		} else if len(others) > minPeerCount {
			logger.Printf("detected %d peers; joining", len(others))
			others = others[:0] // empty others slice means join
			break
		}
		logger.Printf("detected %d peers; waiting...", len(others))
		time.Sleep(time.Second)
	}

	var (
		incomingc    = make(chan raftpb.Message)    // from meshconn to ctrl
		outgoingc    = make(chan raftpb.Message)    // from ctrl to meshconn
		unreachablec = make(chan uint64, 10000)     // from meshconn to ctrl
		confchangec  = make(chan raftpb.ConfChange) // from meshconn to ctrl
		snapshotc    = make(chan raftpb.Snapshot)   // from ctrl to state machine
		entryc       = make(chan raftpb.Entry)      // from ctrl to state
		confentryc   = make(chan raftpb.Entry)      // from state to configurator
		proposalc    = make(chan []byte)            // from state machine to ctrl
		removedc     = make(chan struct{})          // from ctrl to us
		shrunkc      = make(chan struct{})          // from membership to us
	)

	// Create the thing that watches the cluster membership via the router. It
	// signals conf changes, and closes shrunkc when the cluster is too small.
	var (
		addc = make(chan uint64)
		remc = make(chan uint64)
	)
	m := newMembership(router, membershipSet(router), minPeerCount, addc, remc, shrunkc, logger)
	defer m.stop()

	// Create the thing that converts mesh membership changes to Raft ConfChange
	// proposals.
	c := newConfigurator(addc, remc, confchangec, confentryc, logger)
	defer c.stop()

	// Create a packet transport, wrapping the meshconn.Peer.
	transport := newPacketTransport(peer, translateVia(router), incomingc, outgoingc, unreachablec, logger)
	defer transport.stop()

	// Create the API server. store.stop must go on the defer stack before
	// ctrl.stop so that the ctrl stops first. Otherwise, ctrl can deadlock
	// processing the last tick.
	store := newEtcdStore(proposalc, snapshotc, entryc, confentryc, logger)
	defer store.stop()

	// Create the controller, which drives the Raft node internally.
	ctrl := newCtrl(self, others, minPeerCount, incomingc, outgoingc, unreachablec, confchangec, snapshotc, entryc, proposalc, removedc, logger)
	defer ctrl.stop()

	// Return the store to the client.
	out <- store

	errc := make(chan error)
	go func() {
		<-terminatec
		errc <- fmt.Errorf("metcd server terminated by user request")
	}()
	go func() {
		<-removedc
		errc <- fmt.Errorf("the Raft peer was removed from the cluster")
	}()
	go func() {
		<-shrunkc
		errc <- fmt.Errorf("the Raft cluster got too small")
	}()
	terminatedc <- <-errc
}

func translateVia(router *mesh.Router) peerTranslator {
	return func(uid mesh.PeerUID) (mesh.PeerName, error) {
		for _, d := range router.Peers.Descriptions() {
			if d.UID == uid {
				return d.Name, nil
			}
		}
		return 0, fmt.Errorf("peer UID %x not known", uid)
	}
}

func mustPeerName() mesh.PeerName {
	peerName, err := mesh.PeerNameFromString(mustHardwareAddr())
	if err != nil {
		panic(err)
	}
	return peerName
}

func mustHardwareAddr() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	for _, iface := range ifaces {
		if s := iface.HardwareAddr.String(); s != "" {
			return s
		}
	}
	panic("no valid network interfaces")
}

func mustHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return hostname
}
