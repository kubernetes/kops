package nameserver

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/weaveworks/mesh"

	"github.com/weaveworks/weave/net/address"
)

func l(es Entries) Entries {
	es.addLowercase()
	return es
}

func makeEntries(values string) Entries {
	entries := make(Entries, len(values))
	for i, c := range values {
		entries[i] = Entry{Hostname: string(c)}
	}
	return l(entries)
}

func TestAdd(t *testing.T) {
	oldNow := now
	defer func() { now = oldNow }()
	now = func() int64 { return 1234 }

	entries := Entries{}
	entries.add("A", "", mesh.UnknownPeerName, address.Address(0))
	expected := l(Entries{
		Entry{Hostname: "A", Origin: mesh.UnknownPeerName, Addr: address.Address(0)},
	})
	require.Equal(t, entries, expected)

	entries.tombstone(mesh.UnknownPeerName, func(e *Entry) bool { return e.Hostname == "A" })
	expected = l(Entries{
		Entry{Hostname: "A", Origin: mesh.UnknownPeerName, Addr: address.Address(0), Version: 1, Tombstone: 1234},
	})
	require.Equal(t, entries, expected)

	entries.add("A", "", mesh.UnknownPeerName, address.Address(0))
	expected = l(Entries{
		Entry{Hostname: "A", Origin: mesh.UnknownPeerName, Addr: address.Address(0), Version: 2},
	})
	require.Equal(t, entries, expected)
}

func TestMerge(t *testing.T) {
	e1 := makeEntries("ACDF")
	e2 := makeEntries("BEF")

	diff := e1.merge(e2)

	require.Equal(t, makeEntries("BE"), diff)
	require.Equal(t, makeEntries("ABCDEF"), e1)

	diff = e1.merge(e1)
	require.Equal(t, Entries{}, diff)
}

func TestOldMerge(t *testing.T) {
	e1 := l(Entries{Entry{Hostname: "A", Version: 0}})
	diff := e1.merge(l(Entries{Entry{Hostname: "A", Version: 1}}))
	require.Equal(t, l(Entries{Entry{Hostname: "A", Version: 1}}), diff)
	require.Equal(t, l(Entries{Entry{Hostname: "A", Version: 1}}), e1)

	diff = e1.merge(l(Entries{Entry{Hostname: "A", Version: 0}}))
	require.Equal(t, Entries{}, diff)
	require.Equal(t, l(Entries{Entry{Hostname: "A", Version: 1}}), e1)
}

func TestTombstone(t *testing.T) {
	oldNow := now
	defer func() { now = oldNow }()
	now = func() int64 { return 1234 }

	es := makeEntries("AB")

	es.tombstone(mesh.UnknownPeerName, func(e *Entry) bool {
		return e.Hostname == "B"
	})
	expected := l(Entries{
		Entry{Hostname: "A"},
		Entry{Hostname: "B", Version: 1, Tombstone: 1234},
	})
	require.Equal(t, expected, es)

	// Now try a merge including two entries which differ only in tombstone
	e2 := make(Entries, len(es))
	copy(e2, es)
	e2[1].Tombstone = 5555

	diff := es.merge(e2)

	expected2 := l(Entries{
		Entry{Hostname: "A"},
		Entry{Hostname: "B", Version: 1, Tombstone: 5555},
	})
	require.Equal(t, expected2, es)
	expectedDiff := l(Entries{Entry{Hostname: "B", Version: 1, Tombstone: 5555}})
	require.Equal(t, expectedDiff, diff)
}

func TestDelete(t *testing.T) {
	es := makeEntries("AB")

	es.filter(func(e *Entry) bool {
		return e.Hostname != "A"
	})
	require.Equal(t, makeEntries("B"), es)
}

func TestLookup(t *testing.T) {
	es := l(Entries{
		Entry{Hostname: "A"},
		Entry{Hostname: "B", ContainerID: "bar"},
		Entry{Hostname: "B", ContainerID: "foo"},
		Entry{Hostname: "C"},
	})

	have := es.lookup("B")
	want := l(Entries{
		Entry{Hostname: "B", ContainerID: "bar"},
		Entry{Hostname: "B", ContainerID: "foo"},
	})
	require.Equal(t, have, want)
}

func TestGossipDataMerge(t *testing.T) {
	g1 := GossipData{Entries: makeEntries("AcDf")}
	g2 := GossipData{Entries: makeEntries("BEf")}

	g3 := g1.Merge(&g2).(*GossipData)

	require.Equal(t, GossipData{Entries: makeEntries("ABcDEf")}, *g3)
}
