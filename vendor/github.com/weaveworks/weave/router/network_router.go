package router

import (
	"fmt"
	"math"
	"net"
	"os"
	"time"

	"github.com/weaveworks/mesh"

	"github.com/weaveworks/weave/common"
	"github.com/weaveworks/weave/db"
)

const (
	ChannelSize         = 16
	MaxUDPPacketSize    = 65535
	FastHeartbeat       = 500 * time.Millisecond
	SlowHeartbeat       = 10 * time.Second
	MaxMissedHeartbeats = 6
	HeartbeatTimeout    = MaxMissedHeartbeats * SlowHeartbeat
	MaxDuration         = time.Duration(math.MaxInt64)
	NameSize            = mesh.NameSize
	// should be greater than typical ARP cache expiries, i.e. > 3/2 *
	// /proc/sys/net/ipv4_neigh/*/base_reachable_time_ms on Linux
	macMaxAge = 10 * time.Minute
)

var (
	log        = common.Log
	checkFatal = common.CheckFatal
	checkWarn  = common.CheckWarn
)

type NetworkConfig struct {
	BufSz         int
	PacketLogging PacketLogging
	Bridge        Bridge
}

type PacketLogging interface {
	LogPacket(string, PacketKey)
	LogForwardPacket(string, ForwardPacketKey)
}

type NetworkRouter struct {
	*mesh.Router
	NetworkConfig
	Macs *MacCache
	db   db.DB
}

func NewNetworkRouter(config mesh.Config, networkConfig NetworkConfig, name mesh.PeerName, nickName string, overlay NetworkOverlay, db db.DB) *NetworkRouter {
	if overlay == nil {
		overlay = NullNetworkOverlay{}
	}
	if networkConfig.Bridge == nil {
		networkConfig.Bridge = NullBridge{}
	}

	router := &NetworkRouter{Router: mesh.NewRouter(config, name, nickName, overlay, common.LogLogger()), NetworkConfig: networkConfig, db: db}
	router.Peers.OnInvalidateShortIDs(overlay.InvalidateShortIDs)
	router.Routes.OnChange(overlay.InvalidateRoutes)
	router.Macs = NewMacCache(macMaxAge,
		func(mac net.HardwareAddr, peer *mesh.Peer) {
			log.Println("Expired MAC", mac, "at", peer)
		})
	router.Peers.OnGC(func(peer *mesh.Peer) { router.Macs.Delete(peer) })
	return router
}

// Start listening for TCP connections, locally captured packets, and
// forwarded packets.
func (router *NetworkRouter) Start() {
	log.Println("Sniffing traffic on", router.Bridge)
	checkFatal(router.Bridge.StartConsumingPackets(router.handleCapturedPacket))
	checkFatal(router.Overlay.(NetworkOverlay).StartConsumingPackets(router.Ourself.Peer, router.Peers, router.handleForwardedPacket))
	router.Router.Start()
}

func (router *NetworkRouter) handleCapturedPacket(key PacketKey) FlowOp {
	router.PacketLogging.LogPacket("Captured", key)
	srcMac := net.HardwareAddr(key.SrcMAC[:])
	dstMac := net.HardwareAddr(key.DstMAC[:])

	switch newSrcMac, conflictPeer := router.Macs.Add(srcMac, router.Ourself.Peer); {
	case newSrcMac:
		log.Println("Discovered local MAC", srcMac)
	case conflictPeer != nil:
		// The MAC cache has an entry for the source MAC
		// associated with another peer.  This probably means
		// we are seeing a frame we injected ourself.  That
		// shouldn't happen, but discard it just in case.
		log.Error("Captured frame from MAC (", srcMac, ") to (", dstMac, ") associated with another peer ", conflictPeer)
		return DiscardingFlowOp{}
	}

	// Discard STP broadcasts
	if key.DstMAC == [...]byte{0x01, 0x80, 0xC2, 0x00, 0x00, 0x00} {
		return DiscardingFlowOp{}
	}

	switch dstPeer := router.Macs.Lookup(dstMac); dstPeer {
	case router.Ourself.Peer:
		// The packet is destined for a local MAC.  The bridge
		// won't normally send us such packets, and if it does
		// it's likely to be broadcasting the packet to all
		// ports.  So if it happens, just drop the packet to
		// avoid warnings if we try to forward it.
		return DiscardingFlowOp{}
	case nil:
		// If we don't know which peer corresponds to the dest
		// MAC, broadcast it.
		router.PacketLogging.LogPacket("Broadcasting", key)
		return router.relayBroadcast(router.Ourself.Peer, key)
	default:
		router.PacketLogging.LogPacket("Forwarding", key)
		return router.relay(ForwardPacketKey{
			PacketKey: key,
			SrcPeer:   router.Ourself.Peer,
			DstPeer:   dstPeer})
	}
}

func (router *NetworkRouter) handleForwardedPacket(key ForwardPacketKey) FlowOp {
	srcMac := net.HardwareAddr(key.SrcMAC[:])
	dstMac := net.HardwareAddr(key.DstMAC[:])

	if key.SrcPeer == router.Ourself.Peer {
		log.Warn("Received own packet to peer ", key.DstPeer, " from MAC (", srcMac, ") to (", dstMac, ")")
	}

	if key.DstPeer != router.Ourself.Peer {
		// it's not for us, we're just relaying it
		router.PacketLogging.LogForwardPacket("Relaying", key)
		return router.relay(key)
	}

	// At this point, it's either unicast to us, or a broadcast
	// (because the DstPeer on a forwarded broadcast packet is
	// always set to the peer being forwarded to)

	switch newSrcMac, conflictPeer := router.Macs.AddForced(srcMac, key.SrcPeer); {
	case newSrcMac:
		log.Print("Discovered remote MAC ", srcMac, " at ", key.SrcPeer)
	case conflictPeer != nil:
		log.Print("Discovered remote MAC ", srcMac, " at ", key.SrcPeer, " (was at ", conflictPeer, ")")
		// We need to clear out any flows destined to the MAC
		// that forward to the old peer.
		router.Overlay.(NetworkOverlay).InvalidateRoutes()
	}

	router.PacketLogging.LogForwardPacket("Injecting", key)
	injectFop := router.Bridge.InjectPacket(key.PacketKey)
	dstPeer := router.Macs.Lookup(dstMac)
	if dstPeer == router.Ourself.Peer {
		return injectFop
	}

	router.PacketLogging.LogForwardPacket("Relaying broadcast", key)
	relayFop := router.relayBroadcast(key.SrcPeer, key.PacketKey)
	switch {
	case injectFop == nil:
		return relayFop
	case relayFop == nil:
		return injectFop
	default:
		mfop := NewMultiFlowOp(false)
		mfop.Add(injectFop)
		mfop.Add(relayFop)
		return mfop
	}
}

// Routing

func (router *NetworkRouter) relay(key ForwardPacketKey) FlowOp {
	relayPeerName, found := router.Routes.Unicast(key.DstPeer.Name)
	if !found {
		// Not necessarily an error as there could be a race with the
		// dst disappearing whilst the frame is in flight
		log.Println("Received packet for unknown destination:", key.DstPeer)
		return DiscardingFlowOp{}
	}

	conn, found := router.Ourself.ConnectionTo(relayPeerName)
	if !found {
		// Again, could just be a race, not necessarily an error
		log.Println("Unable to find connection to relay peer", relayPeerName)
		return DiscardingFlowOp{}
	}

	return conn.(*mesh.LocalConnection).OverlayConn.(OverlayForwarder).Forward(key)
}

func (router *NetworkRouter) relayBroadcast(srcPeer *mesh.Peer, key PacketKey) FlowOp {
	nextHops := router.Routes.Broadcast(srcPeer.Name)
	if len(nextHops) == 0 {
		return DiscardingFlowOp{}
	}

	op := NewMultiFlowOp(true)

	for _, conn := range router.Ourself.ConnectionsTo(nextHops) {
		op.Add(conn.(*mesh.LocalConnection).OverlayConn.(OverlayForwarder).Forward(ForwardPacketKey{
			PacketKey: key,
			SrcPeer:   srcPeer,
			DstPeer:   conn.Remote()}))
	}

	return op
}

// Persisting the set of peers we are supposed to connect to
const peersIdent = "directPeers"

func (router *NetworkRouter) persistPeers() {
	if err := router.db.Save(peersIdent, router.ConnectionMaker.Targets(false)); err != nil {
		log.Errorf("Error persisting peers: %s", err)
		return
	}
}

func (router *NetworkRouter) InitiateConnections(peers []string, replace bool) []error {
	errors := router.ConnectionMaker.InitiateConnections(peers, replace)
	router.persistPeers()
	return errors
}

func (router *NetworkRouter) ForgetConnections(peers []string) {
	router.ConnectionMaker.ForgetConnections(peers)
	router.persistPeers()
}

func (router *NetworkRouter) InitialPeers(resume bool, peers []string) ([]string, error) {
	if _, err := os.Stat("restart.sentinel"); err == nil || resume {
		var storedPeers []string
		if _, err := router.db.Load(peersIdent, &storedPeers); err != nil {
			return nil, err
		}
		log.Println("Restart/resume detected - using persisted peer list:", storedPeers)
		return storedPeers, nil
	}

	log.Println("Launch detected - using supplied peer list:", peers)
	return peers, nil
}

func (router *NetworkRouter) CreateRestartSentinel() error {
	sentinel, err := os.Create("restart.sentinel")
	if err != nil {
		return fmt.Errorf("error creating sentinel: %v", err)
	}
	sentinel.Close()

	return nil
}
