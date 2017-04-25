package main

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"log"
	"reflect"
	"testing"

	"github.com/weaveworks/mesh"
)

func TestPeerOnGossip(t *testing.T) {
	for _, testcase := range []struct {
		initial map[mesh.PeerName]int
		msg     map[mesh.PeerName]int
		want    map[mesh.PeerName]int
	}{
		{
			map[mesh.PeerName]int{},
			map[mesh.PeerName]int{123: 1, 456: 2},
			map[mesh.PeerName]int{123: 1, 456: 2},
		},
		{
			map[mesh.PeerName]int{123: 1},
			map[mesh.PeerName]int{123: 0, 456: 2},
			map[mesh.PeerName]int{456: 2},
		},
		{
			map[mesh.PeerName]int{123: 9},
			map[mesh.PeerName]int{123: 8},
			nil,
		},
	} {
		p := newPeer(mesh.PeerName(999), log.New(ioutil.Discard, "", 0))
		p.st.mergeComplete(testcase.initial)
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(testcase.msg); err != nil {
			t.Fatal(err)
		}
		delta, err := p.OnGossip(buf.Bytes())
		if err != nil {
			t.Errorf("%v OnGossip %v: %v", testcase.initial, testcase.msg, err)
			continue
		}
		if want := testcase.want; want == nil {
			if delta != nil {
				t.Errorf("%v OnGossip %v: want nil, have non-nil", testcase.initial, testcase.msg)
			}
		} else {
			if have := delta.(*state).set; !reflect.DeepEqual(want, have) {
				t.Errorf("%v OnGossip %v: want %v, have %v", testcase.initial, testcase.msg, want, have)
			}
		}
	}
}

func TestPeerOnGossipBroadcast(t *testing.T) {
	for _, testcase := range []struct {
		initial map[mesh.PeerName]int
		msg     map[mesh.PeerName]int
		want    map[mesh.PeerName]int
	}{
		{
			map[mesh.PeerName]int{},
			map[mesh.PeerName]int{123: 1, 456: 2},
			map[mesh.PeerName]int{123: 1, 456: 2},
		},
		{
			map[mesh.PeerName]int{123: 1},
			map[mesh.PeerName]int{123: 0, 456: 2},
			map[mesh.PeerName]int{456: 2},
		},
		{
			map[mesh.PeerName]int{123: 9},
			map[mesh.PeerName]int{123: 8},
			map[mesh.PeerName]int{}, // OnGossipBroadcast returns received, which should never be nil
		},
	} {
		p := newPeer(999, log.New(ioutil.Discard, "", 0))
		p.st.mergeComplete(testcase.initial)
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(testcase.msg); err != nil {
			t.Fatal(err)
		}
		delta, err := p.OnGossipBroadcast(mesh.UnknownPeerName, buf.Bytes())
		if err != nil {
			t.Errorf("%v OnGossipBroadcast %v: %v", testcase.initial, testcase.msg, err)
			continue
		}
		if want, have := testcase.want, delta.(*state).set; !reflect.DeepEqual(want, have) {
			t.Errorf("%v OnGossipBroadcast %v: want %v, have %v", testcase.initial, testcase.msg, want, have)
		}
	}
}

func TestPeerOnGossipUnicast(t *testing.T) {
	for _, testcase := range []struct {
		initial map[mesh.PeerName]int
		msg     map[mesh.PeerName]int
		want    map[mesh.PeerName]int
	}{
		{
			map[mesh.PeerName]int{},
			map[mesh.PeerName]int{123: 1, 456: 2},
			map[mesh.PeerName]int{123: 1, 456: 2},
		},
		{
			map[mesh.PeerName]int{123: 1},
			map[mesh.PeerName]int{123: 0, 456: 2},
			map[mesh.PeerName]int{123: 1, 456: 2},
		},
		{
			map[mesh.PeerName]int{123: 9},
			map[mesh.PeerName]int{123: 8},
			map[mesh.PeerName]int{123: 9},
		},
	} {
		p := newPeer(999, log.New(ioutil.Discard, "", 0))
		p.st.mergeComplete(testcase.initial)
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(testcase.msg); err != nil {
			t.Fatal(err)
		}
		if err := p.OnGossipUnicast(mesh.UnknownPeerName, buf.Bytes()); err != nil {
			t.Errorf("%v OnGossipBroadcast %v: %v", testcase.initial, testcase.msg, err)
			continue
		}
		if want, have := testcase.want, p.st.set; !reflect.DeepEqual(want, have) {
			t.Errorf("%v OnGossipBroadcast %v: want %v, have %v", testcase.initial, testcase.msg, want, have)
		}
	}
}
