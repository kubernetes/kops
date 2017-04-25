package gossip

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/weaveworks/mesh"

	"github.com/weaveworks/weave/common"
)

// Router to convey gossip from one gossiper to another, for testing
type unicastMessage struct {
	sender mesh.PeerName
	buf    []byte
}
type broadcastMessage struct {
	sender mesh.PeerName
	data   mesh.GossipData
}
type gossipMessage struct {
	sender mesh.PeerName
	data   mesh.GossipData
}
type exitMessage struct {
	exitChan chan struct{}
}
type flushMessage struct {
	flushChan chan struct{}
}

type TestRouter struct {
	sync.Mutex
	gossipChans map[mesh.PeerName]chan interface{}
	loss        float32 // 0.0 means no loss
}

func NewTestRouter(loss float32) *TestRouter {
	return &TestRouter{gossipChans: make(map[mesh.PeerName]chan interface{}, 100), loss: loss}
}

// Copy so we can access outside of a lock
func (grouter *TestRouter) copyGossipChans() map[mesh.PeerName]chan interface{} {
	ret := make(map[mesh.PeerName]chan interface{})
	grouter.Lock()
	defer grouter.Unlock()
	for p, c := range grouter.gossipChans {
		ret[p] = c
	}
	return ret
}

func (grouter *TestRouter) Stop() {
	for peer := range grouter.copyGossipChans() {
		grouter.RemovePeer(peer)
	}
}

func (grouter *TestRouter) gossipBroadcast(sender mesh.PeerName, update mesh.GossipData) {
	for _, gossipChan := range grouter.copyGossipChans() {
		select {
		case gossipChan <- broadcastMessage{sender: sender, data: update}:
		default: // drop the message if we cannot send it
			common.Log.Errorf("Dropping message")
		}
	}
}

func (grouter *TestRouter) gossip(sender mesh.PeerName, update mesh.GossipData) error {
	gossipChans := grouter.copyGossipChans()
	count := int(math.Log2(float64(len(gossipChans))))
	for dest, gossipChan := range gossipChans {
		if dest == sender {
			continue
		}
		select {
		case gossipChan <- gossipMessage{sender: sender, data: update}:
		default: // drop the message if we cannot send it
			common.Log.Errorf("Dropping message")
		}
		count--
		if count <= 0 {
			break
		}
	}
	return nil
}

func (grouter *TestRouter) Flush() {
	for _, gossipChan := range grouter.copyGossipChans() {
		flushChan := make(chan struct{})
		gossipChan <- flushMessage{flushChan: flushChan}
		<-flushChan
	}
}

func (grouter *TestRouter) RemovePeer(peer mesh.PeerName) {
	grouter.Lock()
	gossipChan := grouter.gossipChans[peer]
	grouter.Unlock()
	resultChan := make(chan struct{})
	gossipChan <- exitMessage{exitChan: resultChan}
	<-resultChan
	grouter.Lock()
	delete(grouter.gossipChans, peer)
	grouter.Unlock()
}

type TestRouterClient struct {
	router *TestRouter
	sender mesh.PeerName
}

func (grouter *TestRouter) run(sender mesh.PeerName, gossiper mesh.Gossiper, gossipChan chan interface{}) {
	gossipTimer := time.Tick(2 * time.Second)
	for {
		select {
		case gossip := <-gossipChan:
			switch message := gossip.(type) {
			case exitMessage:
				close(message.exitChan)
				return

			case flushMessage:
				close(message.flushChan)

			case unicastMessage:
				if rand.Float32() > (1.0 - grouter.loss) {
					continue
				}
				if err := gossiper.OnGossipUnicast(message.sender, message.buf); err != nil {
					panic(fmt.Sprintf("Error doing gossip unicast to %s: %s", message.sender, err))
				}

			case broadcastMessage:
				if rand.Float32() > (1.0 - grouter.loss) {
					continue
				}
				for _, msg := range message.data.Encode() {
					if _, err := gossiper.OnGossipBroadcast(message.sender, msg); err != nil {
						panic(fmt.Sprintf("Error doing gossip broadcast: %s", err))
					}
				}
			case gossipMessage:
				if rand.Float32() > (1.0 - grouter.loss) {
					continue
				}
				for _, msg := range message.data.Encode() {
					diff, err := gossiper.OnGossip(msg)
					if err != nil {
						panic(fmt.Sprintf("Error doing gossip: %s", err))
					}
					if diff == nil {
						continue
					}
					// Sanity check - reconsuming the diff should yield nil
					for _, diffMsg := range diff.Encode() {
						if nextDiff, err := gossiper.OnGossip(diffMsg); err != nil {
							panic(fmt.Sprintf("Error doing gossip: %s", err))
						} else if nextDiff != nil {
							panic(fmt.Sprintf("Breach of gossip interface: %v != nil", nextDiff))
						}
					}
					grouter.gossip(message.sender, diff)
				}
			}
		case <-gossipTimer:
			grouter.gossip(sender, gossiper.Gossip())
		}
	}
}

func (grouter *TestRouter) Connect(sender mesh.PeerName, gossiper mesh.Gossiper) mesh.Gossip {
	gossipChan := make(chan interface{}, 100)

	go grouter.run(sender, gossiper, gossipChan)

	grouter.Lock()
	grouter.gossipChans[sender] = gossipChan
	grouter.Unlock()
	return TestRouterClient{grouter, sender}
}

func (client TestRouterClient) GossipUnicast(dstPeerName mesh.PeerName, buf []byte) error {
	client.router.Lock()
	gossipChan := client.router.gossipChans[dstPeerName]
	client.router.Unlock()
	select {
	case gossipChan <- unicastMessage{sender: client.sender, buf: buf}:
	default: // drop the message if we cannot send it
		common.Log.Errorf("Dropping message")
	}
	return nil
}

func (client TestRouterClient) GossipBroadcast(update mesh.GossipData) {
	client.router.gossipBroadcast(client.sender, update)
}
