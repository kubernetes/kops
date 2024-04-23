package scaleway

import (
	"os"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"k8s.io/kops/upup/pkg/fi"
)

func TestGetServerPrivateIP(t *testing.T) {
	err := os.Setenv("SCW_PROFILE", "default")
	if err != nil {
		t.Errorf("setting env: %v", err)
	}
	scwCloud, err := NewScwCloud(nil)
	if err != nil {
		t.Error(err)
	}
	ip, err := scwCloud.GetServerPrivateIP("scw-wizardly-mestorf", "fr-par-1")
	if err != nil {
		t.Error(err)
	}
	if ip != "10.76.76.41" {
		t.Errorf("expected 10.76.76.41, got %s", ip)
	}

	ips, err := scwCloud.IPAMService().ListIPs(&ipam.ListIPsRequest{
		Region:           "",
		Page:             nil,
		PageSize:         nil,
		OrderBy:          "",
		ProjectID:        nil,
		OrganizationID:   nil,
		Zonal:            nil,
		PrivateNetworkID: nil,
		Attached:         nil,
		ResourceID:       nil,
		ResourceType:     "",
		MacAddress:       nil,
		Tags:             nil,
		IsIPv6:           nil,
		ResourceName:     fi.PtrTo("api.paris.nonedns"),
	})
	if err != nil {
		t.Error(err)
	}
	if ips.TotalCount != 1 {

		t.Error(err)
	}
}
