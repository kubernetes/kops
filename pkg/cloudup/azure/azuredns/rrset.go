package azuredns

import (
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	"k8s.io/kubernetes/federation/pkg/dnsprovider/rrstype"
)

// Compile time check for interface adherence
var _ dnsprovider.ResourceRecordSet = ResourceRecordSet{}

type ResourceRecordSet struct {
	// Todo (kris-nova) this is actually a resource from the Azure go SDK! Wee :)
	//impl   *route53.ResourceRecordSet
	
	rrsets *ResourceRecordSets
}

func (rrset ResourceRecordSet) Name() string {
	return ""
}
func (rrset ResourceRecordSet) Rrdatas() []string {
	var datas []string
	return datas
}
func (rrset ResourceRecordSet) Ttl() int64 {
	return 200
}
func (rrset ResourceRecordSet) Type() rrstype.RrsType {
	return rrstype.A
}

