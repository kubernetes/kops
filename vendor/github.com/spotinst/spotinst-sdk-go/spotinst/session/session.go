package session

import (
	"github.com/spotinst/spotinst-sdk-go/spotinst"
)

// A Session provides a central location to create service clients.
//
// Sessions are safe to create service clients concurrently, but it is not safe
// to mutate the Session concurrently.
type Session struct {
	Config *spotinst.Config
}

// New creates a new instance of Session. Once the Session is created it
// can be mutated to modify the Config. The Session is safe to be read
// concurrently, but it should not be written to concurrently.
func New(cfgs ...*spotinst.Config) *Session {
	s := &Session{Config: spotinst.DefaultConfig()}
	s.Config.Merge(cfgs...)
	return s
}
