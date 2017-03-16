package meters

import (
	"testing"

	th "github.com/rackspace/gophercloud/testhelper"
	"github.com/rackspace/gophercloud/testhelper/client"
)

func TestListMeters(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleMeterListSuccessfully(t)

	list, err := List(client.ServiceClient(), ListOpts{}).Extract()
	th.AssertNoErr(t, err)

	if len(list) != 2 {
		t.Fatalf("Expected 2 meters, got %d", len(list))
	}
	th.CheckDeepEquals(t, MeterHerp, (list)[0])
	th.CheckDeepEquals(t, MeterDerp, (list)[1])
}

func TestShowMeters(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleMeterShowSuccessfully(t)

	meters, err := Show(client.ServiceClient(), "instance", ShowOpts{}).Extract()
	th.AssertNoErr(t, err)

	if len(meters) != 1 {
		t.Fatalf("Expected 1 meters, got %d", len(meters))
	}
	th.CheckDeepEquals(t, ShowHerp, (meters)[0])
}

func TestMeterStatistics(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleMeterStatisticsSuccessfully(t)

	list, err := MeterStatistics(client.ServiceClient(), "memory", MeterStatisticsOpts{}).Extract()
	th.AssertNoErr(t, err)

	if len(list) != 2 {
		t.Fatalf("Expected 2 statistics, got %d", len(list))
	}
	th.CheckDeepEquals(t, StatisticsHerp, (list)[0])
	th.CheckDeepEquals(t, StatisticsDerp, (list)[1])
}
