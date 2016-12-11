package mockroute53

import (
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/golang/glog"
)

type zoneInfo struct {
}

func (m *MockRoute53) ListHostedZonesRequest(*route53.ListHostedZonesInput) (*request.Request, *route53.ListHostedZonesOutput) {
	panic("MockRoute53 function not implemented")
	return nil, nil
}

func (m *MockRoute53) ListHostedZones(*route53.ListHostedZonesInput) (*route53.ListHostedZonesOutput, error) {
	panic("MockRoute53 function not implemented")
	return nil, nil
}

func (m *MockRoute53) ListHostedZonesPages(request *route53.ListHostedZonesInput, callback func(*route53.ListHostedZonesOutput, bool) bool) error {
	glog.Infof("ListHostedZonesPages %v", request)

	page := &route53.ListHostedZonesOutput{}
	for _, zone := range m.Zones {
		copy := *zone
		page.HostedZones = append(page.HostedZones, &copy)
	}
	lastPage := true
	callback(page, lastPage)

	return nil
}

func (m *MockRoute53) ListHostedZonesByNameRequest(*route53.ListHostedZonesByNameInput) (*request.Request, *route53.ListHostedZonesByNameOutput) {
	panic("MockRoute53 function not implemented")
	return nil, nil
}

func (m *MockRoute53) ListHostedZonesByName(*route53.ListHostedZonesByNameInput) (*route53.ListHostedZonesByNameOutput, error) {
	panic("MockRoute53 function not implemented")
	return nil, nil
}
