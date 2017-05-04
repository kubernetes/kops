package azuredns

import (
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
)

var _ dnsprovider.Zone = &Zone{}

type Zone struct {
	// Todo (kris-nova) this is actually a resource from the Azure go SDK! Wee :)
	//impl  *route53.HostedZone

	zones *Zones
}

func (zone Zone) Name() string {
	return ""
}
func (zone Zone) ID() string {
	return ""
}
func (zone Zone) ResourceRecordSets() (dnsprovider.ResourceRecordSets, bool) {
	return &ResourceRecordSets{&zone}, true
}
