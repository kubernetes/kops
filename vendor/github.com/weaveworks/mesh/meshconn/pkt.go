package meshconn

import (
	"bytes"
	"encoding/gob"

	"github.com/weaveworks/mesh"
)

type pkt struct {
	SrcName mesh.PeerName
	SrcUID  mesh.PeerUID
	Buf     []byte
}

func makePkt(buf []byte) pkt {
	var p pkt
	if err := gob.NewDecoder(bytes.NewBuffer(buf)).Decode(&p); err != nil {
		panic(err)
	}
	return p
}

func (p pkt) encode() []byte {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(p); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// pktSlice implements mesh.GossipData.
type pktSlice []pkt

var _ mesh.GossipData = &pktSlice{}

func (s pktSlice) Encode() [][]byte {
	bufs := make([][]byte, len(s))
	for i, pkt := range s {
		bufs[i] = pkt.encode()
	}
	return bufs
}

func (s pktSlice) Merge(other mesh.GossipData) mesh.GossipData {
	o := other.(pktSlice)
	merged := make(pktSlice, 0, len(s)+len(o))
	merged = append(merged, s...)
	merged = append(merged, o...)
	return merged
}
