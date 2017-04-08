package meshconn

import (
	"fmt"
	"net"

	"github.com/weaveworks/mesh"
)

// MeshAddr implements net.Addr for mesh peers.
type MeshAddr struct {
	mesh.PeerName // stable across invocations
	mesh.PeerUID  // new with each invocation
}

var _ net.Addr = MeshAddr{}

// Network returns weavemesh.
func (a MeshAddr) Network() string { return "weavemesh" }

// String returns weavemesh://<PeerName>.
func (a MeshAddr) String() string { return fmt.Sprintf("%s://%s", a.Network(), a.PeerName.String()) }
