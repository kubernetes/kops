package azuredns

import (
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
)

// Compile time check for interface adherence
var _ dnsprovider.Zones = Zones{}

type Zones struct {
	interface_ *Interface
}

func (zones Zones) List() ([]dnsprovider.Zone, error) {
	var z []dnsprovider.Zone
	return z, nil
}

func (zones Zones) Add(zone dnsprovider.Zone) (dnsprovider.Zone, error) {
	return zone, nil
}

func (zones Zones) Remove(dnsprovider.Zone) error {
	return nil
}

func (zones Zones) New(name string) (dnsprovider.Zone, error) {
	return Zone{}, nil
}
