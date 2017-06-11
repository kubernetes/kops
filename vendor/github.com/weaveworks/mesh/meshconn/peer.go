package meshconn

import (
	"errors"
	"net"
	"time"

	"github.com/weaveworks/mesh"
)

var (
	// ErrShortRead is returned by ReadFrom when the
	// passed buffer is too small for the packet.
	ErrShortRead = errors.New("short read")

	// ErrPeerClosed is returned by ReadFrom and WriteTo
	// when the peer is closed during the operation.
	ErrPeerClosed = errors.New("peer closed")

	// ErrGossipNotRegistered is returned by Write to when attempting
	// to write before a mesh.Gossip has been registered in the peer.
	ErrGossipNotRegistered = errors.New("gossip not registered")

	// ErrNotMeshAddr is returned by WriteTo when attempting
	// to write to a non-mesh address.
	ErrNotMeshAddr = errors.New("not a mesh addr")

	// ErrNotSupported is returned by methods that are not supported.
	ErrNotSupported = errors.New("not supported")
)

// Peer implements mesh.Gossiper and net.PacketConn.
type Peer struct {
	name    mesh.PeerName
	uid     mesh.PeerUID
	gossip  mesh.Gossip
	recv    chan pkt
	actions chan func()
	quit    chan struct{}
	logger  mesh.Logger
}

// NewPeer returns a Peer, which can be used as a net.PacketConn.
// Clients must Register a mesh.Gossip before calling ReadFrom or WriteTo.
// Clients should aggressively consume from ReadFrom.
func NewPeer(name mesh.PeerName, uid mesh.PeerUID, logger mesh.Logger) *Peer {
	p := &Peer{
		name:    name,
		uid:     uid,
		gossip:  nil, // initially no gossip
		recv:    make(chan pkt),
		actions: make(chan func()),
		quit:    make(chan struct{}),
		logger:  logger,
	}
	go p.loop()
	return p
}

func (p *Peer) loop() {
	for {
		select {
		case f := <-p.actions:
			f()
		case <-p.quit:
			return
		}
	}
}

// Register injects the mesh.Gossip and enables full-duplex communication.
// Clients should consume from ReadFrom without blocking.
func (p *Peer) Register(gossip mesh.Gossip) {
	p.actions <- func() { p.gossip = gossip }
}

// ReadFrom implements net.PacketConn.
// Clients should consume from ReadFrom without blocking.
func (p *Peer) ReadFrom(b []byte) (n int, remote net.Addr, err error) {
	c := make(chan struct{})
	p.actions <- func() {
		go func() { // so as not to block loop
			defer close(c)
			select {
			case pkt := <-p.recv:
				n = copy(b, pkt.Buf)
				remote = MeshAddr{PeerName: pkt.SrcName, PeerUID: pkt.SrcUID}
				if n < len(pkt.Buf) {
					err = ErrShortRead
				}
			case <-p.quit:
				err = ErrPeerClosed
			}
		}()
	}
	<-c
	return n, remote, err
}

// WriteTo implements net.PacketConn.
func (p *Peer) WriteTo(b []byte, dst net.Addr) (n int, err error) {
	c := make(chan struct{})
	p.actions <- func() {
		defer close(c)
		if p.gossip == nil {
			err = ErrGossipNotRegistered
			return
		}
		meshAddr, ok := dst.(MeshAddr)
		if !ok {
			err = ErrNotMeshAddr
			return
		}
		pkt := pkt{SrcName: p.name, SrcUID: p.uid, Buf: b}
		if meshAddr.PeerName == p.name {
			p.recv <- pkt
			return
		}
		// TODO(pb): detect and support broadcast
		buf := pkt.encode()
		n = len(buf)
		err = p.gossip.GossipUnicast(meshAddr.PeerName, buf)
	}
	<-c
	return n, err
}

// Close implements net.PacketConn.
func (p *Peer) Close() error {
	close(p.quit)
	return nil
}

// LocalAddr implements net.PacketConn.
func (p *Peer) LocalAddr() net.Addr {
	return MeshAddr{PeerName: p.name, PeerUID: p.uid}
}

// SetDeadline implements net.PacketConn.
// SetDeadline is not supported.
func (p *Peer) SetDeadline(time.Time) error {
	return ErrNotSupported
}

// SetReadDeadline implements net.PacketConn.
// SetReadDeadline is not supported.
func (p *Peer) SetReadDeadline(time.Time) error {
	return ErrNotSupported
}

// SetWriteDeadline implements net.PacketConn.
// SetWriteDeadline is not supported.
func (p *Peer) SetWriteDeadline(time.Time) error {
	return ErrNotSupported
}

// Gossip implements mesh.Gossiper.
func (p *Peer) Gossip() (complete mesh.GossipData) {
	return pktSlice{} // we're stateless
}

// OnGossip implements mesh.Gossiper.
// The buf is a single pkt.
func (p *Peer) OnGossip(buf []byte) (delta mesh.GossipData, err error) {
	return pktSlice{makePkt(buf)}, nil
}

// OnGossipBroadcast implements mesh.Gossiper.
// The buf is a single pkt
func (p *Peer) OnGossipBroadcast(_ mesh.PeerName, buf []byte) (received mesh.GossipData, err error) {
	pkt := makePkt(buf)
	p.recv <- pkt // to ReadFrom
	return pktSlice{pkt}, nil
}

// OnGossipUnicast implements mesh.Gossiper.
// The buf is a single pkt.
func (p *Peer) OnGossipUnicast(_ mesh.PeerName, buf []byte) error {
	pkt := makePkt(buf)
	p.recv <- pkt // to ReadFrom
	return nil
}
