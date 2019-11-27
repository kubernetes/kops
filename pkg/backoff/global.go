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
	"sync"
	"time"

	"k8s.io/klog"
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
	pause := computeBackoff()

	klog.Warningf("inserting rate-limiting pause of %v after error: %v", pause, err)
	time.Sleep(pause)
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
