package azuredns

import (
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	"k8s.io/kubernetes/federation/pkg/dnsprovider/rrstype"
)

var _ dnsprovider.ResourceRecordSets = ResourceRecordSets{}

type ResourceRecordSets struct {
	zone *Zone
}

func (rrsets ResourceRecordSets) List() ([]dnsprovider.ResourceRecordSet, error) {
	var rrsetsslice []dnsprovider.ResourceRecordSet
	return rrsetsslice, nil

}
func (rrsets ResourceRecordSets) Get(name string) (dnsprovider.ResourceRecordSet, error) {
	return ResourceRecordSet{}, nil
}
func (rrsets ResourceRecordSets) New(name string, rrdatas []string, ttl int64, rrstype rrstype.RrsType) dnsprovider.ResourceRecordSet {
	return ResourceRecordSet{}
}
func (rrsets ResourceRecordSets) StartChangeset() dnsprovider.ResourceRecordChangeset {
	return ResourceRecordChangeset{}
}
func (rrsets ResourceRecordSets) Zone() dnsprovider.Zone {
	return Zone{}
}
