package main

import (
	"log"

	"bytes"
	"encoding/gob"

	"github.com/weaveworks/mesh"
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
	logger  *log.Logger
}

// peer implements mesh.Gossiper.
var _ mesh.Gossiper = &peer{}

// Construct a peer with empty state.
// Be sure to register a channel, later,
// so we can make outbound communication.
func newPeer(self mesh.PeerName, logger *log.Logger) *peer {
	actions := make(chan func())
	p := &peer{
		st:      newState(self),
		send:    nil, // must .register() later
		actions: actions,
		quit:    make(chan struct{}),
		logger:  logger,
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
func (p *peer) get() int {
	return p.st.get()
}

// Increment the counter by one.
func (p *peer) incr() (result int) {
	c := make(chan struct{})
	p.actions <- func() {
		defer close(c)
		st := p.st.incr()
		if p.send != nil {
			p.send.GossipBroadcast(st)
		} else {
			p.logger.Printf("no sender configured; not broadcasting update right now")
		}
		result = st.get()
	}
	<-c
	return result
}

func (p *peer) stop() {
	close(p.quit)
}

// Return a copy of our complete state.
func (p *peer) Gossip() (complete mesh.GossipData) {
	complete = p.st.copy()
	p.logger.Printf("Gossip => complete %v", complete.(*state).set)
	return complete
}

// Merge the gossiped data represented by buf into our state.
// Return the state information that was modified.
func (p *peer) OnGossip(buf []byte) (delta mesh.GossipData, err error) {
	var set map[mesh.PeerName]int
	if err := gob.NewDecoder(bytes.NewReader(buf)).Decode(&set); err != nil {
		return nil, err
	}

	delta = p.st.mergeDelta(set)
	if delta == nil {
		p.logger.Printf("OnGossip %v => delta %v", set, delta)
	} else {
		p.logger.Printf("OnGossip %v => delta %v", set, delta.(*state).set)
	}
	return delta, nil
}

// Merge the gossiped data represented by buf into our state.
// Return the state information that was modified.
func (p *peer) OnGossipBroadcast(src mesh.PeerName, buf []byte) (received mesh.GossipData, err error) {
	var set map[mesh.PeerName]int
	if err := gob.NewDecoder(bytes.NewReader(buf)).Decode(&set); err != nil {
		return nil, err
	}

	received = p.st.mergeReceived(set)
	if received == nil {
		p.logger.Printf("OnGossipBroadcast %s %v => delta %v", src, set, received)
	} else {
		p.logger.Printf("OnGossipBroadcast %s %v => delta %v", src, set, received.(*state).set)
	}
	return received, nil
}

// Merge the gossiped data represented by buf into our state.
func (p *peer) OnGossipUnicast(src mesh.PeerName, buf []byte) error {
	var set map[mesh.PeerName]int
	if err := gob.NewDecoder(bytes.NewReader(buf)).Decode(&set); err != nil {
		return err
	}

	complete := p.st.mergeComplete(set)
	p.logger.Printf("OnGossipUnicast %s %v => complete %v", src, set, complete)
	return nil
}
