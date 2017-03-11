package mesh

import (
	"bytes"
	"encoding/gob"
	"github.com/golang/glog"
	"github.com/weaveworks/mesh"
	"k8s.io/kops/protokube/pkg/gossip"
)

// Peer encapsulates state and implements mesh.Gossiper.
// It should be passed to mesh.Router.NewGossip,
// and the resulting Gossip registered in turn,
// before calling mesh.Router.Start.
type peer struct {
	st      *state
	send    mesh.Gossip
	actions chan<- func()
	quit    chan struct{}
}

// peer implements mesh.Gossiper.
var _ mesh.Gossiper = &peer{}

// Construct a peer with empty state.
// Be sure to register a channel, later,
// so we can make outbound communication.
func newPeer(self mesh.PeerName) *peer {
	actions := make(chan func())
	p := &peer{
		st:      newState(self),
		send:    nil, // must .register() later
		actions: actions,
		quit:    make(chan struct{}),
	}
	go p.loop(actions)
	return p
}

func (p *peer) loop(actions <-chan func()) {
	for {
		select {
		case f := <-actions:
			f()
		case <-p.quit:
			return
		}
	}
}

// register the result of a mesh.Router.NewGossip.
func (p *peer) register(send mesh.Gossip) {
	p.actions <- func() { p.send = send }
}

// Return the current value of the counter.
func (p *peer) get(key string) []byte {
	return p.st.get(key)
}

func (p *peer) snapshot() *gossip.GossipStateSnapshot {
	return p.st.snapshot()
}

// Increment the counter by one.
func (p *peer) put(key string, value []byte) (result []byte) {
	c := make(chan struct{})
	p.actions <- func() {
		defer close(c)
		p.st.put(key, value)
		if p.send != nil {
			p.send.GossipBroadcast(p.st)
		} else {
			glog.Warningf("no sender configured; not broadcasting update right now")
		}
		result = p.st.get(key)
	}
	<-c
	return result
}

func (p *peer) updateValues(removeKeys []string, putEntries map[string]string) error {
	c := make(chan struct{})
	p.actions <- func() {
		defer close(c)
		p.st.updateValues(removeKeys, putEntries)
		if p.send != nil {
			p.send.GossipBroadcast(p.st)
		} else {
			glog.Warningf("no sender configured; not broadcasting update right now")
		}
	}
	<-c
	return nil
}

func (p *peer) stop() {
	close(p.quit)
}

// Return a copy of our complete state.
func (p *peer) Gossip() (complete mesh.GossipData) {
	complete = p.st.copy()
	glog.V(4).Infof("Gossip => complete %v", complete.(*state).valueMap)
	return complete
}

// Merge the gossiped data represented by buf into our state.
// Return the state information that was modified.
func (p *peer) OnGossip(buf []byte) (delta mesh.GossipData, err error) {
	var updates map[string]record
	if err := gob.NewDecoder(bytes.NewReader(buf)).Decode(&updates); err != nil {
		glog.Warningf("error decoding OnGossip: %v", err)
		return nil, err
	}

	deltas := make(map[string]record)
	p.st.merge(updates, &deltas)

	var deltaState *state
	if len(deltas) <= 0 {
		// per OnGossip requirements
	} else {
		deltaState = &state{valueMap: deltas}
		delta = deltaState
	}

	if deltaState == nil {
		glog.V(4).Infof("OnGossip %v => delta %v", updates, deltaState)
	} else {
		glog.V(4).Infof("OnGossip %v => delta %v", updates, deltaState.valueMap)
	}
	return delta, nil
}

// Merge the gossiped data represented by buf into our state.
// Return the state information that was modified.
func (p *peer) OnGossipBroadcast(src mesh.PeerName, buf []byte) (received mesh.GossipData, err error) {
	var updates map[string]record
	if err := gob.NewDecoder(bytes.NewReader(buf)).Decode(&updates); err != nil {
		glog.Warningf("error decoding OnGossipBroadcast: %v", err)
		return nil, err
	}

	deltas := make(map[string]record)
	p.st.merge(updates, &deltas)

	deltaState := &state{valueMap: deltas}

	glog.V(4).Infof("OnGossipBroadcast %s %v => delta %v", src, updates, deltaState)

	return deltaState, nil
}

// Merge the gossiped data represented by buf into our state.
func (p *peer) OnGossipUnicast(src mesh.PeerName, buf []byte) error {
	var updates map[string]record
	if err := gob.NewDecoder(bytes.NewReader(buf)).Decode(&updates); err != nil {
		glog.Warningf("error decoding OnGossipUnicast: %v", err)
		return err
	}

	p.st.merge(updates, nil)

	glog.V(4).Infof("OnGossipUnicast %s %v => complete %v", src, updates, p.st)
	return nil
}
