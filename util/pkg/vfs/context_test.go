/*
Copyright 2026 The Kubernetes Authors.

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

package vfs

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

func TestNextBackoffDuration(t *testing.T) {
	grid := []struct {
		name     string
		duration time.Duration
		backoff  wait.Backoff
		expected time.Duration
	}{
		{
			name:     "doubles below the cap",
			duration: 4 * time.Second,
			backoff:  wait.Backoff{Factor: 2, Cap: 30 * time.Second},
			expected: 8 * time.Second,
		},
		{
			name:     "clamps at the cap",
			duration: 16 * time.Second,
			backoff:  wait.Backoff{Factor: 2, Cap: 30 * time.Second},
			expected: 30 * time.Second,
		},
		{
			name:     "stays at the cap once reached",
			duration: 30 * time.Second,
			backoff:  wait.Backoff{Factor: 2, Cap: 30 * time.Second},
			expected: 30 * time.Second,
		},
		{
			name:     "grows unbounded when no cap is set",
			duration: 16 * time.Second,
			backoff:  wait.Backoff{Factor: 2},
			expected: 32 * time.Second,
		},
	}

	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			actual := nextBackoffDuration(g.duration, g.backoff)
			if actual != g.expected {
				t.Errorf("nextBackoffDuration(%v, cap=%v) = %v, want %v", g.duration, g.backoff.Cap, actual, g.expected)
			}
		})
	}
}

// TestRetryWithBackoffRespectsCap covers the reason the cap exists: an uncapped backoff with a
// large Steps count pushes later attempts arbitrarily far apart, so a caller that has been
// failing for a while stops retrying at any useful rate.
func TestRetryWithBackoffRespectsCap(t *testing.T) {
	backoff := wait.Backoff{
		Duration: 1 * time.Millisecond,
		Factor:   2,
		Cap:      4 * time.Millisecond,
		Steps:    8,
	}

	attempts := 0
	done, err := RetryWithBackoff(backoff, func() (bool, error) {
		attempts++
		return attempts >= 8, nil
	})
	if err != nil {
		t.Fatalf("RetryWithBackoff returned error: %v", err)
	}
	if !done {
		t.Errorf("RetryWithBackoff returned done=false, want true")
	}
	if attempts != 8 {
		t.Errorf("condition called %d times, want 8", attempts)
	}
}
