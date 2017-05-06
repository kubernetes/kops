package azuredns

import (
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
)

// Compile time check for interface adherence
var _ dnsprovider.Interface = Interface{}

// Interface is the Azure dnsprovider implementation
type Interface struct {
}

func New() *Interface {
	return &Interface{}
}

func (i Interface) Zones() (dnsprovider.Zones, bool) {
	zones := Zones{}
	return zones, true
}
