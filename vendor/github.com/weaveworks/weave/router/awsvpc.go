package router

// A dummy overlay for the AWSVPC underlay to make `weave status` to return
// a valid information about peer connections.

import (
	"github.com/weaveworks/mesh"
)

// mesh.OverlayConnection

type AWSVPCConnection struct {
	establishedChan chan struct{}
	errorChan       chan error
}

func (conn *AWSVPCConnection) Confirm() {
	// We close the channel to notify mesh that the connection has been established.
	close(conn.establishedChan)
}

func (conn *AWSVPCConnection) EstablishedChannel() <-chan struct{} {
	return conn.establishedChan
}

func (conn *AWSVPCConnection) ErrorChannel() <-chan error {
	return conn.errorChan
}

func (conn *AWSVPCConnection) Stop() {}

func (conn *AWSVPCConnection) ControlMessage(tag byte, msg []byte) {
}

func (conn *AWSVPCConnection) Attrs() map[string]interface{} {
	return map[string]interface{}{"name": "awsvpc"}
}

// OverlayForwarder

func (conn *AWSVPCConnection) Forward(key ForwardPacketKey) FlowOp {
	return DiscardingFlowOp{}
}

type AWSVPC struct{}

func NewAWSVPC() AWSVPC {
	return AWSVPC{}
}

// mesh.Overlay

func (vpc AWSVPC) AddFeaturesTo(features map[string]string) {}

func (vpc AWSVPC) PrepareConnection(params mesh.OverlayConnectionParams) (mesh.OverlayConnection, error) {
	conn := &AWSVPCConnection{
		establishedChan: make(chan struct{}),
		errorChan:       make(chan error, 1),
	}
	return conn, nil
}

func (vpc AWSVPC) Diagnostics() interface{} {
	return nil
}

func (vpc AWSVPC) Stop() {}

// NetworkOverlay

func (vpc AWSVPC) InvalidateRoutes() {}

func (vpc AWSVPC) InvalidateShortIDs() {}

func (vpc AWSVPC) StartConsumingPackets(localPeer *mesh.Peer, peers *mesh.Peers, consumer OverlayConsumer) error {
	return nil
}
