package main

import (
	"reflect"
	"testing"

	"github.com/weaveworks/mesh"
)

func TestStateMergeReceived(t *testing.T) {
	for _, testcase := range []struct {
		initial map[mesh.PeerName]int
		merge   map[mesh.PeerName]int
		want    map[mesh.PeerName]int
	}{
		{
			map[mesh.PeerName]int{},
			map[mesh.PeerName]int{123: 1, 456: 2},
			map[mesh.PeerName]int{123: 1, 456: 2},
		},
		{
			map[mesh.PeerName]int{123: 1, 456: 2},
			map[mesh.PeerName]int{123: 1, 456: 2},
			map[mesh.PeerName]int{},
		},
		{
			map[mesh.PeerName]int{123: 1, 456: 2},
			map[mesh.PeerName]int{789: 3},
			map[mesh.PeerName]int{789: 3},
		},
		{
			map[mesh.PeerName]int{456: 3},
			map[mesh.PeerName]int{123: 1, 456: 2},
			map[mesh.PeerName]int{123: 1}, // we drop keys that don't semantically merge
		},
	} {
		initial, merge := testcase.initial, testcase.merge // mergeReceived modifies arguments
		delta := newState(999).mergeComplete(initial).(*state).mergeReceived(merge)
		if want, have := testcase.want, delta.(*state).set; !reflect.DeepEqual(want, have) {
			t.Errorf("%v mergeReceived %v: want %v, have %v", testcase.initial, testcase.merge, want, have)
		}
	}
}

func TestStateMergeDelta(t *testing.T) {
	for _, testcase := range []struct {
		initial map[mesh.PeerName]int
		merge   map[mesh.PeerName]int
		want    map[mesh.PeerName]int
	}{
		{
			map[mesh.PeerName]int{},
			map[mesh.PeerName]int{123: 1, 456: 2},
			map[mesh.PeerName]int{123: 1, 456: 2},
		},
		{
			map[mesh.PeerName]int{123: 1, 456: 2},
			map[mesh.PeerName]int{123: 1, 456: 2},
			nil,
		},
		{
			map[mesh.PeerName]int{123: 1, 456: 2},
			map[mesh.PeerName]int{789: 3},
			map[mesh.PeerName]int{789: 3},
		},
		{
			map[mesh.PeerName]int{123: 1, 456: 2},
			map[mesh.PeerName]int{456: 3},
			map[mesh.PeerName]int{456: 3},
		},
	} {
		initial, merge := testcase.initial, testcase.merge // mergeDelta modifies arguments
		delta := newState(999).mergeComplete(initial).(*state).mergeDelta(merge)
		if want := testcase.want; want == nil {
			if delta != nil {
				t.Errorf("%v mergeDelta %v: want nil, have non-nil", testcase.initial, testcase.merge)
			}
		} else {
			if have := delta.(*state).set; !reflect.DeepEqual(want, have) {
				t.Errorf("%v mergeDelta %v: want %v, have %v", testcase.initial, testcase.merge, want, have)
			}
		}
	}
}

func TestStateMergeComplete(t *testing.T) {
	for _, testcase := range []struct {
		initial map[mesh.PeerName]int
		merge   map[mesh.PeerName]int
		want    map[mesh.PeerName]int
	}{
		{
			map[mesh.PeerName]int{},
			map[mesh.PeerName]int{123: 1, 456: 2},
			map[mesh.PeerName]int{123: 1, 456: 2},
		},
		{
			map[mesh.PeerName]int{123: 1, 456: 2},
			map[mesh.PeerName]int{123: 1, 456: 2},
			map[mesh.PeerName]int{123: 1, 456: 2},
		},
		{
			map[mesh.PeerName]int{123: 1, 456: 2},
			map[mesh.PeerName]int{789: 3},
			map[mesh.PeerName]int{123: 1, 456: 2, 789: 3},
		},
		{
			map[mesh.PeerName]int{123: 1, 456: 2},
			map[mesh.PeerName]int{123: 0, 456: 3},
			map[mesh.PeerName]int{123: 1, 456: 3},
		},
	} {
		st := newState(999).mergeComplete(testcase.initial).(*state).mergeComplete(testcase.merge).(*state)
		if want, have := testcase.want, st.set; !reflect.DeepEqual(want, have) {
			t.Errorf("%v mergeComplete %v: want %v, have %v", testcase.initial, testcase.merge, want, have)
		}
	}
}
