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
	"github.com/weaveworks/mesh"
	"k8s.io/klog"
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

func (p *peer) snapshot() *gossip.GossipStateSnapshot {
	return p.st.snapshot()
}

func (p *peer) updateValues(removeKeys []string, putEntries map[string]string) error {
	c := make(chan struct{})
	p.actions <- func() {
		defer close(c)
		p.st.updateValues(removeKeys, putEntries)
		if p.send != nil {
			gossipData := p.st.getData()
			p.send.GossipBroadcast(gossipData)
		} else {
			klog.Warningf("no sender configured; not broadcasting update right now")
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
	data := p.st.getData()
	klog.V(4).Infof("Gossip => complete %v", data)
	return data
}

// Merge the gossiped data represented by buf into our state.
// Return the state information that was modified.
func (p *peer) OnGossip(buf []byte) (delta mesh.GossipData, err error) {
	message, err := DecodeKVState(buf)
	if err != nil {
		klog.Warningf("error decoding OnGossip: %v", err)
		return nil, err
	}

	deltas := &KVState{}
	p.st.merge(message, deltas)

	if len(deltas.Records) == 0 {
		// per OnGossip requirements
		klog.V(4).Infof("OnGossip %v => delta empty", message)
		return nil, nil
	}
	klog.V(4).Infof("OnGossip %v => delta %v", message, deltas)
	return deltas, nil
}

// Merge the gossiped data represented by buf into our state.
// Return the state information that was modified.
func (p *peer) OnGossipBroadcast(src mesh.PeerName, buf []byte) (received mesh.GossipData, err error) {
	message, err := DecodeKVState(buf)
	if err != nil {
		klog.Warningf("error decoding OnGossipBroadcast: %v", err)
		return nil, err
	}

	deltas := &KVState{}
	p.st.merge(message, deltas)

	klog.V(4).Infof("OnGossipBroadcast %s %v => delta %v", src, message, deltas)

	return deltas, nil
}

// Merge the gossiped data represented by buf into our state.
func (p *peer) OnGossipUnicast(src mesh.PeerName, buf []byte) error {
	message, err := DecodeKVState(buf)
	if err != nil {
		klog.Warningf("error decoding OnGossipUnicast: %v", err)
		return err
	}

	p.st.merge(message, nil)

	klog.V(4).Infof("OnGossipUnicast %s %v => complete %v", src, message, p.st)
	return nil
}
