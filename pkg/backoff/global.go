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
	"math/rand"
	"sync"
	"time"

	"k8s.io/klog/v2"
)

// globalBackoffMutex guards globalBackoff
var globalBackoffMutex sync.Mutex

// globalBackoff is the current backoff value
var globalBackoff = 1 * time.Second

// maxGlobalBackoff value is the maximum wait time
const maxGlobalBackoff = 5 * time.Minute

// DoGlobalBackoff performs a sleep with a pretty slow backoff.
// The primary use is to rate-limit repeated downloads, to prevent runaway bandwidth bills
func DoGlobalBackoff(err error) {
	pause := jitter(computeBackoff())

	klog.Warningf("inserting rate-limiting pause of %v after error: %v", pause, err)
	time.Sleep(pause)
}

// jitter spreads a pause by adding a random amount on top of it, bounded so the
// result never exceeds maxGlobalBackoff.
//
// computeBackoff is a pure doubling sequence, so every node that hits the same
// failing download computes the same delays. Nodes coming up together during a
// cluster scale-up therefore retry at the same instants, and once the backoff
// saturates at maxGlobalBackoff they continue to retry in unison every five
// minutes. That reforms the request burst this pause exists to prevent.
//
// The jitter is added rather than centred on the interval so that the pause is
// never shorter than computeBackoff intended. This function exists to limit
// download rate, so a shorter wait would work against its purpose.
//
// Once the backoff saturates there is no headroom left to add into, and that is
// the case that matters most: globalBackoff never decreases, so a node that
// keeps failing stays at maxGlobalBackoff indefinitely and the whole fleet
// settles into exactly the lockstep described above. At the ceiling the pause is
// therefore spread downwards instead. Drawing below maxGlobalBackoff breaks no
// contract, because it is an upper bound rather than a target.
func jitter(d time.Duration) time.Duration {
	if d <= 0 {
		return d
	}

	extra := d
	if headroom := maxGlobalBackoff - d; headroom < extra {
		extra = headroom
	}
	if extra > 0 {
		return d + time.Duration(rand.Int63n(int64(extra)+1))
	}

	return d - time.Duration(rand.Int63n(int64(d/2)+1))
}

// computeBackoff computes the next backoff value, by doubling the backoff value, capping it at maxGlobalBackoff
func computeBackoff() time.Duration {
	globalBackoffMutex.Lock()
	defer globalBackoffMutex.Unlock()

	v := globalBackoff
	v = v + v
	if v > maxGlobalBackoff {
		v = maxGlobalBackoff
	}
	globalBackoff = v

	return v
}
