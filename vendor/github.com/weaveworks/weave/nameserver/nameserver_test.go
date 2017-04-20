package nameserver

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/weaveworks/mesh"

	"github.com/weaveworks/weave/net/address"
	"github.com/weaveworks/weave/testing/gossip"
)

func makeNameserver(name mesh.PeerName) *Nameserver {
	return New(name, "", func(mesh.PeerName) bool { return true })
}

func makeNetwork(size int) ([]*Nameserver, *gossip.TestRouter) {
	gossipRouter := gossip.NewTestRouter(0.0)
	nameservers := make([]*Nameserver, size)

	for i := 0; i < size; i++ {
		name, _ := mesh.PeerNameFromString(fmt.Sprintf("%02d:00:00:02:00:00", i))
		nameserver := makeNameserver(name)
		nameserver.SetGossip(gossipRouter.Connect(nameserver.ourName, nameserver))
		nameserver.Start()
		nameservers[i] = nameserver
	}

	return nameservers, gossipRouter
}

func stopNetwork(nameservers []*Nameserver, grouter *gossip.TestRouter) {
	for _, nameserver := range nameservers {
		nameserver.Stop()
	}
	grouter.Stop()
}

type pair struct {
	origin mesh.PeerName
	addr   address.Address
}

type mapping struct {
	hostname string
	addrs    []pair
}

type addrs []address.Address

func (a addrs) Len() int           { return len(a) }
func (a addrs) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a addrs) Less(i, j int) bool { return a[i] < a[j] }
func (a addrs) String() string {
	ss := []string{}
	for _, addr := range a {
		ss = append(ss, addr.String())
	}
	return strings.Join(ss, " ")
}

func (m mapping) Addrs() []address.Address {
	want := addrs{}
	for _, p := range m.addrs {
		want = append(want, p.addr)
	}
	sort.Sort(want)
	return want
}

func TestNameservers(t *testing.T) {
	//common.SetLogLevel("debug")

	lookupTimeout := 10 // ms
	nameservers, grouter := makeNetwork(30)
	defer stopNetwork(nameservers, grouter)
	// This subset will sometimes lose touch with some of the others
	badNameservers := nameservers[25:]
	// This subset will remain well-connected, and we will deal mainly with them
	nameservers = nameservers[:25]
	nameserversByName := map[mesh.PeerName]*Nameserver{}
	for _, n := range nameservers {
		nameserversByName[n.ourName] = n
	}
	mappings := []mapping{}

	check := func(nameserver *Nameserver, expected mapping) {
		have := []address.Address{}
		for i := 0; i < lookupTimeout; i++ {
			have = nameserver.Lookup(expected.hostname)
			sort.Sort(addrs(have))
			if reflect.DeepEqual(have, expected.Addrs()) {
				return
			}
			time.Sleep(1 * time.Millisecond)
		}
		want := expected.Addrs()
		require.Equal(t, addrs(want).String(), addrs(have).String())
	}

	addMapping := func() {
		nameserver := nameservers[rand.Intn(len(nameservers))]
		addr := address.Address(rand.Int31())
		// Create a hostname which has some upper and lowercase letters,
		// and a unique number so we don't have to check if we allocated it already
		randomBits := rand.Int63()
		firstLetter := 'H' + (randomBits&1)*32
		secondLetter := 'O' + (randomBits&2)*16
		randomBits = randomBits >> 2
		hostname := fmt.Sprintf("%c%cstname%d", firstLetter, secondLetter, randomBits)
		mapping := mapping{hostname, []pair{{nameserver.ourName, addr}}}
		mappings = append(mappings, mapping)

		nameserver.AddEntry(hostname, "", nameserver.ourName, addr)
		check(nameserver, mapping)
	}

	addExtraMapping := func() {
		if len(mappings) <= 0 {
			return
		}
		nameserver := nameservers[rand.Intn(len(nameservers))]
		i := rand.Intn(len(mappings))
		mapping := mappings[i]
		addr := address.Address(rand.Int31())
		mapping.addrs = append(mapping.addrs, pair{nameserver.ourName, addr})
		mappings[i] = mapping

		nameserver.AddEntry(mapping.hostname, "", nameserver.ourName, addr)
		check(nameserver, mapping)
	}

	loseConnection := func() {
		nameserver1 := badNameservers[rand.Intn(len(badNameservers))]
		nameserver2 := nameservers[rand.Intn(len(nameservers))]
		nameserver1.PeerGone(nameserver2.ourName)
	}

	deleteMapping := func() {
		if len(mappings) <= 0 {
			return
		}
		i := rand.Intn(len(mappings))
		mapping := mappings[i]
		if len(mapping.addrs) <= 0 {
			return
		}
		j := rand.Intn(len(mapping.addrs))
		pair := mapping.addrs[j]
		mapping.addrs = append(mapping.addrs[:j], mapping.addrs[j+1:]...)
		mappings[i] = mapping
		nameserver := nameserversByName[pair.origin]

		nameserver.Delete(mapping.hostname, "*", pair.addr.String(), pair.addr)
		check(nameserver, mapping)
	}

	doLookup := func() {
		if len(mappings) <= 0 {
			return
		}
		mapping := mappings[rand.Intn(len(mappings))]
		nameserver := nameservers[rand.Intn(len(nameservers))]
		check(nameserver, mapping)
	}

	doReverseLookup := func() {
		if len(mappings) <= 0 {
			return
		}
		mapping := mappings[rand.Intn(len(mappings))]
		if len(mapping.addrs) <= 0 {
			return
		}
		nameserver := nameservers[rand.Intn(len(nameservers))]
		hostname := ""
		var err error
		for i := 0; i < lookupTimeout; i++ {
			hostname, err = nameserver.ReverseLookup(mapping.addrs[0].addr)
			if err != nil && mapping.hostname == hostname {
				return
			}
			time.Sleep(1 * time.Millisecond)
		}
		require.Nil(t, err)
		require.Equal(t, mapping.hostname, hostname)
	}

	for i := 0; i < 800; i++ {
		r := rand.Float32()
		switch {
		case r < 0.1:
			addMapping()

		case 0.1 <= r && r < 0.2:
			addExtraMapping()

		case 0.2 <= r && r < 0.3:
			deleteMapping()

		case 0.3 <= r && r < 0.35:
			loseConnection()

		case 0.35 <= r && r < 0.9:
			doLookup()

		case 0.9 <= r:
			doReverseLookup()
		}

		grouter.Flush()
	}
}

func TestContainerAndPeerDeath(t *testing.T) {
	peername, err := mesh.PeerNameFromString("00:00:00:02:00:00")
	require.Nil(t, err)
	nameserver := makeNameserver(peername)

	nameserver.AddEntry("hostname", "containerid", peername, address.Address(0))
	require.Equal(t, []address.Address{0}, nameserver.Lookup("hostname"))

	nameserver.ContainerDied("containerid")
	require.Equal(t, []address.Address{}, nameserver.Lookup("hostname"))

	nameserver.AddEntry("hostname", "containerid", peername, address.Address(0))
	require.Equal(t, []address.Address{0}, nameserver.Lookup("hostname"))

	nameserver.PeerGone(peername)
	require.Equal(t, []address.Address{}, nameserver.Lookup("hostname"))
}

func TestTombstoneDeletion(t *testing.T) {
	oldNow := now
	defer func() { now = oldNow }()
	now = func() int64 { return 1234 }

	peername, err := mesh.PeerNameFromString("00:00:00:02:00:00")
	require.Nil(t, err)
	nameserver := makeNameserver(peername)

	nameserver.AddEntry("hostname", "containerid", peername, address.Address(0))
	require.Equal(t, []address.Address{0}, nameserver.Lookup("hostname"))

	nameserver.deleteTombstones()
	require.Equal(t, []address.Address{0}, nameserver.Lookup("hostname"))

	nameserver.Delete("hostname", "containerid", "", address.Address(0))
	require.Equal(t, []address.Address{}, nameserver.Lookup("hostname"))
	require.Equal(t, l(Entries{Entry{
		ContainerID: "containerid",
		Origin:      peername,
		Addr:        address.Address(0),
		Hostname:    "hostname",
		Version:     1,
		Tombstone:   1234,
	}}), nameserver.entries)

	now = func() int64 { return 1234 + int64(tombstoneTimeout/time.Second) + 1 }
	nameserver.deleteTombstones()
	require.Equal(t, Entries{}, nameserver.entries)
}
