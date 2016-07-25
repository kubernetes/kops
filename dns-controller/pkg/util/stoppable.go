package util

import (
	"fmt"
	"sync"

	"github.com/golang/glog"
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
	glog.Infof("shutting down controller")
	s.shutdown = true

	return nil
}

func (s *Stoppable) StopRequested() bool {
	return s.shutdown
}
