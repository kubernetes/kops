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

package util

import (
	"fmt"
	"sync"

	"k8s.io/klog"
)

// Stoppable implements the standard stop / shutdown logic
type Stoppable struct {
	// mutex is used to enforce only a single call to Stop is active.
	// Needed because we allow stopping through an http endpoint and
	// allowing concurrent stoppers leads to stack traces.
	// We also use it for lazy-init
	mutex       sync.Mutex
	shutdown    bool
	stopChannel chan struct{}
}

// StopChannel gets the stopChannel, initializing it if needed
func (s *Stoppable) StopChannel() <-chan struct{} {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.stopChannel == nil {
		s.stopChannel = make(chan struct{})
	}
	return s.stopChannel
}

// Stop stops the controller.
func (s *Stoppable) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.shutdown {
		return fmt.Errorf("shutdown already in progress")
	}

	// We initialize the channel to avoid a race if we Stop before anyone is watching
	if s.stopChannel == nil {
		s.stopChannel = make(chan struct{})
	}
	close(s.stopChannel)
	klog.Infof("shutting down controller")
	s.shutdown = true

	return nil
}

func (s *Stoppable) StopRequested() bool {
	return s.shutdown
}
