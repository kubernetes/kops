// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pubsub

import (
	"errors"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestKeepAlive(t *testing.T) {
	tick := make(chan time.Time)
	deadline := time.Nanosecond * 15
	s := &testService{modDeadlineCalled: make(chan modDeadlineCall)}

	checkModDeadlineCall := func(ackIDs []string) {
		got := <-s.modDeadlineCalled
		sort.Strings(got.ackIDs)

		want := modDeadlineCall{
			subName:  "subname",
			deadline: deadline,
			ackIDs:   ackIDs,
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("keepalive: got:\n%v\nwant:\n%v", got, want)
		}
	}

	ka := &keepAlive{
		s:             s,
		Ctx:           context.Background(),
		Sub:           "subname",
		ExtensionTick: tick,
		Deadline:      deadline,
		MaxExtension:  time.Hour,
	}
	ka.Start()

	ka.Add("a")
	ka.Add("b")
	tick <- time.Time{}
	checkModDeadlineCall([]string{"a", "b"})
	ka.Add("c")
	ka.Remove("b")
	tick <- time.Time{}
	checkModDeadlineCall([]string{"a", "c"})
	ka.Remove("a")
	ka.Remove("c")
	ka.Add("d")
	tick <- time.Time{}
	checkModDeadlineCall([]string{"d"})

	ka.Remove("d")
	ka.Stop()
}

// TestKeepAliveStop checks that Stop blocks until all ackIDs have been removed.
func TestKeepAliveStop(t *testing.T) {
	tick := 100 * time.Microsecond
	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	s := &testService{modDeadlineCalled: make(chan modDeadlineCall, 100)}

	ka := &keepAlive{
		s:             s,
		Ctx:           context.Background(),
		ExtensionTick: ticker.C,
		MaxExtension:  time.Hour,
	}
	ka.Start()

	events := make(chan string, 10)

	// Add an ackID so that ka.Stop will not return immediately.
	ka.Add("a")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(tick * 10)
		events <- "pre-remove"
		ka.Remove("a")
		time.Sleep(tick * 10)
		events <- "post-second-sleep"
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		events <- "pre-stop"
		ka.Stop()
		events <- "stopped"

	}()

	wg.Wait()
	close(events)
	eventSequence := []string{}
	for e := range events {
		eventSequence = append(eventSequence, e)
	}

	want := []string{"pre-stop", "pre-remove", "stopped", "post-second-sleep"}
	if !reflect.DeepEqual(eventSequence, want) {
		t.Errorf("keepalive eventsequence: got:\n%v\nwant:\n%v", eventSequence, want)
	}
}

// TestMaxExtensionDeadline checks we stop extending after the configured duration.
func TestMaxExtensionDeadline(t *testing.T) {
	ticker := time.NewTicker(100 * time.Microsecond)
	defer ticker.Stop()

	s := &testService{modDeadlineCalled: make(chan modDeadlineCall, 100)}

	maxExtension := time.Millisecond
	ka := &keepAlive{
		s:             s,
		Ctx:           context.Background(),
		ExtensionTick: ticker.C,
		MaxExtension:  maxExtension,
	}
	ka.Start()

	ka.Add("a")
	stopped := make(chan struct{})

	go func() {
		ka.Stop()
		stopped <- struct{}{}
	}()

	select {
	case <-stopped:
	case <-time.After(maxExtension + 2*time.Second):
		t.Fatalf("keepalive failed to stop after maxExtension deadline")
	}
}

func TestKeepAliveStopsWhenAllAckIDsRemoved(t *testing.T) {
	s := &testService{}

	maxExtension := time.Millisecond
	ka := &keepAlive{
		s:             s,
		Ctx:           context.Background(),
		ExtensionTick: make(chan time.Time),
		MaxExtension:  maxExtension,
	}
	ka.Start()
	ka.Add("a")

	stopped := make(chan struct{})

	go func() {
		ka.Stop()
		stopped <- struct{}{}
	}()

	time.Sleep(time.Microsecond)
	// No extension tick is ever sent, but this should be enough to get ka to stop.
	ka.Remove("a")

	select {
	case <-stopped:
	case <-time.After(maxExtension / 2):
		t.Fatalf("keepalive failed to stop before maxExtension deadline")
	}
}

func TestKeepAliveStopsImmediatelyForNoAckIDs(t *testing.T) {
	ticker := time.NewTicker(100 * time.Microsecond)
	defer ticker.Stop()

	s := &testService{modDeadlineCalled: make(chan modDeadlineCall, 100)}

	maxExtension := time.Millisecond
	ka := &keepAlive{
		s:             s,
		Ctx:           context.Background(),
		ExtensionTick: ticker.C,
		MaxExtension:  maxExtension,
	}
	ka.Start()

	stopped := make(chan struct{})

	go func() {
		// There are no items in ka, so this should return immediately.
		ka.Stop()
		stopped <- struct{}{}
	}()

	select {
	case <-stopped:
	case <-time.After(maxExtension / 2):
		t.Fatalf("keepalive failed to stop before maxExtension deadline")
	}
}

// extendCallResult contains a list of ackIDs which are expected in an ackID
// extension request, along with the result that should be returned.
type extendCallResult struct {
	ackIDs []string
	err    error
}

// extendService implements modifyAckDeadline using a hard-coded list of extendCallResults.
type extendService struct {
	service

	calls []extendCallResult

	t *testing.T // used for error logging.
}

func (es *extendService) modifyAckDeadline(ctx context.Context, subName string, deadline time.Duration, ackIDs []string) error {
	if len(es.calls) == 0 {
		es.t.Fatalf("unexpected call to modifyAckDeadline: ackIDs: %v", ackIDs)
	}
	call := es.calls[0]
	es.calls = es.calls[1:]

	if got, want := ackIDs, call.ackIDs; !reflect.DeepEqual(got, want) {
		es.t.Errorf("unexpected arguments to modifyAckDeadline: got: %v ; want: %v", got, want)
	}
	return call.err
}

// Test implementation returns the first 2 elements as head, and the rest as tail.
func (es *extendService) splitAckIDs(ids []string) ([]string, []string) {
	if len(ids) < 2 {
		return ids, nil
	}
	return ids[:2], ids[2:]
}
func TestKeepAliveSplitsBatches(t *testing.T) {
	type testCase struct {
		calls []extendCallResult
	}
	for _, tc := range []testCase{
		{
			calls: []extendCallResult{
				{
					ackIDs: []string{"a", "b"},
				},
				{
					ackIDs: []string{"c", "d"},
				},
				{
					ackIDs: []string{"e", "f"},
				},
			},
		},
		{
			calls: []extendCallResult{
				{
					ackIDs: []string{"a", "b"},
					err:    errors.New("bang"),
				},
				// On error we retry once.
				{
					ackIDs: []string{"a", "b"},
					err:    errors.New("bang"),
				},
				// We give up after failing twice, so we move on to the next set, "c" and "d"
				{
					ackIDs: []string{"c", "d"},
					err:    errors.New("bang"),
				},
				// Again, we retry once.
				{
					ackIDs: []string{"c", "d"},
				},
				{
					ackIDs: []string{"e", "f"},
				},
			},
		},
	} {
		s := &extendService{
			t:     t,
			calls: tc.calls,
		}

		ka := &keepAlive{
			s:   s,
			Ctx: context.Background(),
			Sub: "subname",
		}

		ka.extendDeadlines([]string{"a", "b", "c", "d", "e", "f"})

		if len(s.calls) != 0 {
			t.Errorf("expected extend calls did not occur: %v", s.calls)
		}
	}
}
