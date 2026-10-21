/*
Copyright The Kubernetes Authors.

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

package try

import (
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

// nextBackoffDuration grows duration by backoff.Factor, clamped to backoff.Cap when one is set.
// An uncapped backoff with a large Steps count grows without bound, which strands a caller for
// far longer than the condition it is waiting on takes to become true.
func nextBackoffDuration(duration time.Duration, backoff wait.Backoff) time.Duration {
	next := time.Duration(float64(duration) * backoff.Factor)
	if backoff.Cap > 0 && next > backoff.Cap {
		return backoff.Cap
	}
	return next
}

// RetryWithBackoff runs until a condition function returns true, or until Steps attempts have been taken
// As compared to wait.ExponentialBackoff, this function returns the results from the function on the final attempt
func RetryWithBackoff(backoff wait.Backoff, condition func() (bool, error)) (bool, error) {
	duration := backoff.Duration
	i := 0
	for {
		if i != 0 {
			adjusted := duration
			if backoff.Jitter > 0.0 {
				adjusted = wait.Jitter(duration, backoff.Jitter)
			}
			time.Sleep(adjusted)
			duration = nextBackoffDuration(duration, backoff)
		}

		i++

		done, err := condition()
		if done {
			return done, err
		}
		noMoreRetries := i >= backoff.Steps
		if !noMoreRetries && err != nil {
			klog.V(2).Infof("retrying after error %v", err)
		}

		if noMoreRetries {
			klog.V(2).Infof("hit maximum retries %d with error %v", i, err)
			return done, err
		}
	}
}
