package ring

import (
	"sort"

	"github.com/weaveworks/mesh"

	"github.com/weaveworks/weave/common"
	"github.com/weaveworks/weave/net/address"
)

// Entry represents entries around the ring
type entry struct {
	Token   address.Address // The start of this range
	Peer    mesh.PeerName   // Who owns this range
	Version uint32          // Version of this range
	Free    address.Count   // Number of free IPs in this range
}

func (e *entry) Equal(e2 *entry) bool {
	return e.Token == e2.Token && e.Peer == e2.Peer &&
		e.Version == e2.Version
}

func (e *entry) update(peername mesh.PeerName, free address.Count) {
	e.Peer = peername
	e.Version++
	e.Free = free
}

// For compatibility with sort.Interface
type entries []*entry

func (es entries) Len() int           { return len(es) }
func (es entries) Less(i, j int) bool { return es[i].Token < es[j].Token }
func (es entries) Swap(i, j int)      { panic("Should never be swapping entries!") }

func (es entries) entry(i int) *entry {
	i = i % len(es)
	if i < 0 {
		i += len(es)
	}
	return es[i]
}

func (es *entries) insert(e entry) {
	i := sort.Search(len(*es), func(j int) bool {
		return (*es)[j].Token >= e.Token
	})

	if i < len(*es) && (*es)[i].Token == e.Token {
		panic("Trying to insert an existing token!")
	}

	*es = append(*es, &entry{})
	copy((*es)[i+1:], (*es)[i:])
	(*es)[i] = &e
}

func (es entries) get(token address.Address) (*entry, bool) {
	i := sort.Search(len(es), func(j int) bool {
		return es[j].Token >= token
	})

	if i < len(es) && es[i].Token == token {
		return es[i], true
	}

	return nil, false
}

// Is token between entries at i and j?
// NB i and j can overflow and will wrap
// NBB this function does not work very well if there is only one
//     token on the ring; luckily an accurate answer is not needed
//     by the call sites in this case.
func (es entries) between(token address.Address, i, j int) bool {
	common.Assert(i < j)

	first := es.entry(i)
	second := es.entry(j)

	switch {
	case first.Token == second.Token:
		// This implies there is only one token
		// on the ring (i < j and i.token == j.token)
		// In which case everything is between, expect
		// this one token
		return token != first.Token

	case first.Token < second.Token:
		return first.Token <= token && token < second.Token

	case second.Token < first.Token:
		return first.Token <= token || token < second.Token
	}

	panic("Should never get here - switch covers all possibilities.")
}
