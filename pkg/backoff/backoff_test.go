/*
Copyright 2019 The Kubernetes Authors.

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

package backoff

import (
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	expected := []time.Duration{
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
		16 * time.Second,
		32 * time.Second,
		64 * time.Second,
		128 * time.Second,
		256 * time.Second,
		5 * time.Minute,
		5 * time.Minute,
		5 * time.Minute,
		5 * time.Minute,
	}

	for i := range expected {
		actual := computeBackoff()
		if actual != expected[i] {
			t.Fatalf("unexpected backoff @%d: %v", i, actual)
		}
	}
}

func TestJitterNeverShorterThanInput(t *testing.T) {
	for _, d := range []time.Duration{
		2 * time.Second,
		32 * time.Second,
		2 * time.Minute,
	} {
		ceiling := 2 * d
		if ceiling > maxGlobalBackoff {
			ceiling = maxGlobalBackoff
		}
		for i := 0; i < 500; i++ {
			got := jitter(d)
			if got < d || got > ceiling {
				t.Fatalf("jitter(%v) = %v, want within [%v, %v]", d, got, d, ceiling)
			}
		}
	}
}

func TestJitterRespectsMaxGlobalBackoff(t *testing.T) {
	for i := 0; i < 500; i++ {
		if got := jitter(maxGlobalBackoff); got != maxGlobalBackoff {
			t.Fatalf("jitter at the cap = %v, want %v", got, maxGlobalBackoff)
		}
	}
}

func TestJitterVaries(t *testing.T) {
	seen := make(map[time.Duration]struct{})
	for i := 0; i < 500; i++ {
		seen[jitter(1*time.Minute)] = struct{}{}
	}

	// A deterministic implementation yields a single value. The range here is
	// a minute wide, so one distinct value across 500 draws is not chance.
	if len(seen) <= 1 {
		t.Fatalf("jitter must vary so nodes do not retry in lockstep, got %v", seen)
	}
}

func TestJitterNonPositive(t *testing.T) {
	if got := jitter(0); got != 0 {
		t.Fatalf("jitter(0) = %v, want 0", got)
	}
	if got := jitter(-time.Second); got != -time.Second {
		t.Fatalf("jitter(-1s) = %v, want -1s", got)
	}
}
