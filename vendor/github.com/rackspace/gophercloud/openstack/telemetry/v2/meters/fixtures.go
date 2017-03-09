// +build fixtures

package meters

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	th "github.com/rackspace/gophercloud/testhelper"
	"github.com/rackspace/gophercloud/testhelper/client"
)

// MeterListBody contains the canned body of a meters.List response.
const MeterListBody = `
[
    {
        "meter_id": "YmQ5NDMxYzEtOGQ2OS00YWQzLTgwM2EtOGQ0YTZiODlmZDM2K2luc3RhbmNl",
        "name": "instance",
        "project_id": "35b17138-b364-4e6a-a131-8f3099c5be68",
        "resource_id": "bd9431c1-8d69-4ad3-803a-8d4a6b89fd36",
        "source": "openstack",
        "type": "gauge",
        "unit": "instance",
        "user_id": "efd87807-12d2-4b38-9c70-5f5c2ac427ff"
    },
	{
		"user_id": "7ya0f7a33717400b951037d55b929c53",
		"name": "cpu_util",
		"resource_id": "5b88239b-8ba1-44ff-a154-978cbda23479",
		"source": "region2",
		"meter_id": "NWI4ODIzOWAtOGJhMS00NGZhLWExNTQtOTc4Y2JkYTIzNDc5K2NwdV91dGls",
		"project_id": "69e6e7c4ed8b434e92feacbf3d4891fd",
		"type": "gauge",
		"unit": "%%"
	}
]
`

// MeterShowBody is the canned body of a Get request on an existing meter.
const MeterShowBody = `
[
    {
        "counter_name": "instance",
        "counter_type": "gauge",
        "counter_unit": "instance",
        "counter_volume": 1.0,
        "message_id": "5460acce-4fd6-480d-ab18-9735ec7b1996",
        "project_id": "35b17138-b364-4e6a-a131-8f3099c5be68",
        "resource_id": "bd9431c1-8d69-4ad3-803a-8d4a6b89fd36",
        "resource_metadata": {
            "name1": "value1",
            "name2": "value2"
        },
        "source": "openstack",
        "timestamp": "2013-11-21T12:33:08.323533",
        "user_id": "efd87807-12d2-4b38-9c70-5f5c2ac427ff"
    }
]
`

// MeterStatisticsBody is the canned body of a Get statistics request on an existing meter.
const MeterStatisticsBody = `
[
    {
        "avg": 4.5,
        "count": 10,
        "duration": 300.0,
        "duration_end": "2013-01-04T16:47:00",
        "duration_start": "2013-01-04T16:42:00",
        "max": 9.0,
        "min": 1.0,
        "period": 7200,
        "period_end": "2013-01-04T18:00:00",
        "period_start": "2013-01-04T16:00:00",
        "sum": 45.0,
        "unit": "GiB"
    },
    {
        "count": 28162,
        "duration_start": "2015-06-27T20:52:08",
        "min": 0.06999999999999999,
        "max": 10.06799336650083,
        "duration_end": "2015-07-27T20:47:02",
        "period": 0,
        "sum": 44655.782463977856,
        "period_end": "2015-07-17T16:43:31",
        "duration": 2591694.0,
        "period_start": "2015-07-17T16:43:31",
        "avg": 1.5856751105737468,
        "groupby": null,
        "unit": "%%"
    }
]
`

var (
	// MeterHerp is a Meter struct that should correspond to the first result in *[]Meter.
	MeterHerp = Meter{
		MeterId:    "YmQ5NDMxYzEtOGQ2OS00YWQzLTgwM2EtOGQ0YTZiODlmZDM2K2luc3RhbmNl",
		Name:       "instance",
		ProjectId:  "35b17138-b364-4e6a-a131-8f3099c5be68",
		ResourceId: "bd9431c1-8d69-4ad3-803a-8d4a6b89fd36",
		Source:     "openstack",
		Type:       "gauge",
		Unit:       "instance",
		UserId:     "efd87807-12d2-4b38-9c70-5f5c2ac427ff",
	}

	// MeterDerp is a Meter struct that should correspond to the second result in *[]Meter.
	MeterDerp = Meter{
		MeterId:    "NWI4ODIzOWAtOGJhMS00NGZhLWExNTQtOTc4Y2JkYTIzNDc5K2NwdV91dGls",
		Name:       "cpu_util",
		ProjectId:  "69e6e7c4ed8b434e92feacbf3d4891fd",
		ResourceId: "5b88239b-8ba1-44ff-a154-978cbda23479",
		Source:     "region2",
		Type:       "gauge",
		Unit:       "%",
		UserId:     "7ya0f7a33717400b951037d55b929c53",
	}

	ShowHerp = OldSample{
		Name:       "instance",
		Type:       "gauge",
		Unit:       "instance",
		Volume:     1.0,
		MessageId:  "5460acce-4fd6-480d-ab18-9735ec7b1996",
		ProjectId:  "35b17138-b364-4e6a-a131-8f3099c5be68",
		ResourceId: "bd9431c1-8d69-4ad3-803a-8d4a6b89fd36",
		ResourceMetadata: map[string]string{
			"name1": "value1",
			"name2": "value2",
		},
		Source:    "openstack",
		Timestamp: time.Date(2013, time.November, 21, 12, 33, 8, 323533000, time.UTC),
		UserId:    "efd87807-12d2-4b38-9c70-5f5c2ac427ff",
	}

	// StatisticsHerp is a Statistics struct that should correspond to the first result in *[]Statistics.
	StatisticsHerp = Statistics{
		Avg:           4.5,
		Count:         10,
		Duration:      300.0,
		DurationEnd:   "2013-01-04T16:47:00",
		DurationStart: "2013-01-04T16:42:00",
		Max:           9.0,
		Min:           1.0,
		Period:        7200,
		PeriodEnd:     "2013-01-04T18:00:00",
		PeriodStart:   "2013-01-04T16:00:00",
		Sum:           45.0,
		Unit:          "GiB",
	}

	// StatisticsDerp is a Statistics struct that should correspond to the second result in *[]Statistics.
	StatisticsDerp = Statistics{
		Avg:           1.5856751105737468,
		Count:         28162,
		Duration:      2591694.0,
		DurationEnd:   "2015-07-27T20:47:02",
		DurationStart: "2015-06-27T20:52:08",
		Max:           10.06799336650083,
		Min:           0.06999999999999999,
		Period:        0,
		PeriodEnd:     "2015-07-17T16:43:31",
		PeriodStart:   "2015-07-17T16:43:31",
		Sum:           44655.782463977856,
		Unit:          "%",
	}
)

// HandleMeterListSuccessfully sets up the test server to respond to a meters List request.
func HandleMeterListSuccessfully(t *testing.T) {
	th.Mux.HandleFunc("/v2/meters", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "GET")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)

		w.Header().Add("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, MeterListBody)
	})
}

// HandleMeterShowSuccessfully sets up the test server to respond to a show meter request.
func HandleMeterShowSuccessfully(t *testing.T) {
	th.Mux.HandleFunc("/v2/meters/instance", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "GET")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, MeterShowBody)
	})
}

// HandleMeterStatisticsSuccessfully sets up the test server to respond to a show meter request.
func HandleMeterStatisticsSuccessfully(t *testing.T) {
	th.Mux.HandleFunc("/v2/meters/memory/statistics", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "GET")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, MeterStatisticsBody)
	})
}
