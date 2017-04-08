package metcd

import (
	"net"

	"github.com/coreos/etcd/raft/raftpb"

	"github.com/weaveworks/mesh"
	"github.com/weaveworks/mesh/meshconn"
)

// packetTransport takes ownership of the net.PacketConn.
// Incoming messages are unmarshaled from the conn and send to incomingc.
// Outgoing messages are received from outgoingc and marshaled to the conn.
type packetTransport struct {
	conn         net.PacketConn
	translate    peerTranslator
	incomingc    chan<- raftpb.Message // to controller
	outgoingc    <-chan raftpb.Message // from controller
	unreachablec chan<- uint64         // to controller
	logger       mesh.Logger
}

func newPacketTransport(
	conn net.PacketConn,
	translate peerTranslator,
	incomingc chan<- raftpb.Message,
	outgoingc <-chan raftpb.Message,
	unreachablec chan<- uint64,
	logger mesh.Logger,
) *packetTransport {
	t := &packetTransport{
		conn:         conn,
		translate:    translate,
		incomingc:    incomingc,
		outgoingc:    outgoingc,
		unreachablec: unreachablec,
		logger:       logger,
	}
	go t.recvLoop()
	go t.sendLoop()
	return t
}

type peerTranslator func(uid mesh.PeerUID) (mesh.PeerName, error)

func (t *packetTransport) stop() {
	t.conn.Close()
}

func (t *packetTransport) recvLoop() {
	defer t.logger.Printf("packet transport: recv loop exit")
	const maxRecvLen = 8192
	b := make([]byte, maxRecvLen)
	for {
		n, remote, err := t.conn.ReadFrom(b)
		if err != nil {
			t.logger.Printf("packet transport: recv: %v (aborting)", err)
			return
		} else if n >= cap(b) {
			t.logger.Printf("packet transport: recv from %s: short read, %d >= %d (continuing)", remote, n, cap(b))
			continue
		}
		var msg raftpb.Message
		if err := msg.Unmarshal(b[:n]); err != nil {
			t.logger.Printf("packet transport: recv from %s (sz %d): %v (%s) (continuing)", remote, n, err, b[:n])
			continue
		}
		//t.logger.Printf("packet transport: recv from %s (sz %d/%d) OK", remote, n, msg.Size())
		t.incomingc <- msg
	}
}

func (t *packetTransport) sendLoop() {
	defer t.logger.Printf("packet transport: send loop exit")
	for msg := range t.outgoingc {
		b, err := msg.Marshal()
		if err != nil {
			t.logger.Printf("packet transport: send to Raft ID %x: %v (continuing)", msg.To, err)
			continue
		}
		peerName, err := t.translate(mesh.PeerUID(msg.To))
		if err != nil {
			select {
			case t.unreachablec <- msg.To:
				t.logger.Printf("packet transport: send to Raft ID %x: %v (unreachable; continuing) (%s)", msg.To, err, msg.Type)
			default:
				t.logger.Printf("packet transport: send to Raft ID %x: %v (unreachable, report dropped; continuing) (%s)", msg.To, err, msg.Type)
			}
			continue
		}
		dst := meshconn.MeshAddr{PeerName: peerName}
		if n, err := t.conn.WriteTo(b, dst); err != nil {
			t.logger.Printf("packet transport: send to Mesh peer %s: %v (continuing)", dst, err)
			continue
		} else if n < len(b) {
			t.logger.Printf("packet transport: send to Mesh peer %s: short write, %d < %d (continuing)", dst, n, len(b))
			continue
		}
		//t.logger.Printf("packet transport: send to %s (sz %d/%d) OK", dst, msg.Size(), len(b))
	}
}
