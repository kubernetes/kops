package designate

import (
	"github.com/gophercloud/gophercloud"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
)

var _ dnsprovider.Interface = Interface{}

type Interface struct {
	sc *gophercloud.ServiceClient
}

// New builds an Interface, with a specified Designate implementation.
// This is useful for testing purposes, but also if we want an instance with custom OpenStack options.
func New(sc *gophercloud.ServiceClient) *Interface {
	return &Interface{sc}
}

func (i Interface) Zones() (zones dnsprovider.Zones, supported bool) {
	return Zones{&i}, true
}
