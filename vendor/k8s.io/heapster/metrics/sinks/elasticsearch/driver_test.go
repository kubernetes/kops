// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package elasticsearch

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/olivere/elastic.v3"
	esCommon "k8s.io/heapster/common/elasticsearch"
	"k8s.io/heapster/metrics/core"
)

type fakeESSink struct {
	core.DataSink
	savedData map[string][]string
}

var FakeESSink fakeESSink

func SaveDataIntoES_Stub(date time.Time, typeName string, sinkData []interface{}) error {
	for _, data := range sinkData {
		jsonItems, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to transform the items to json : %s", err)
		}

		if FakeESSink.savedData[typeName] == nil {
			FakeESSink.savedData[typeName] = []string{}
		}

		FakeESSink.savedData[typeName] = append(FakeESSink.savedData[typeName], string(jsonItems))
	}
	return nil
}

// Returns a fake ES sink.
func NewFakeSink() fakeESSink {
	savedData := make(map[string][]string)
	return fakeESSink{
		&elasticSearchSink{
			saveData:  SaveDataIntoES_Stub,
			flushData: func() error { return nil },
			esSvc: esCommon.ElasticSearchService{
				EsClient:    &elastic.Client{},
				ClusterName: esCommon.ESClusterName,
			},
		},
		savedData,
	}
}

func TestStoreDataEmptyInput(t *testing.T) {
	FakeESSink := NewFakeSink()
	dataBatch := core.DataBatch{}
	FakeESSink.ExportData(&dataBatch)
	assert.Equal(t, 0, len(FakeESSink.savedData))
}

func TestStoreMultipleDataInput(t *testing.T) {
	timestamp := time.Now()

	l := make(map[string]string)
	l["namespace_id"] = "123"
	l["container_name"] = "/system.slice/-.mount"
	l[core.LabelPodId.Key] = "aaaa-bbbb-cccc-dddd"

	l2 := make(map[string]string)
	l2["namespace_id"] = "123"
	l2["container_name"] = "/system.slice/dbus.service"
	l2[core.LabelPodId.Key] = "aaaa-bbbb-cccc-dddd"

	l3 := make(map[string]string)
	l3["namespace_id"] = "123"
	l3[core.LabelPodId.Key] = "aaaa-bbbb-cccc-dddd"

	l4 := make(map[string]string)
	l4["namespace_id"] = ""
	l4[core.LabelPodId.Key] = "aaaa-bbbb-cccc-dddd"

	l5 := make(map[string]string)
	l5["namespace_id"] = "123"
	l5[core.LabelPodId.Key] = "aaaa-bbbb-cccc-dddd"

	metricSet1 := core.MetricSet{
		Labels: l,
		MetricValues: map[string]core.MetricValue{
			"/system.slice/-.mount//cpu/limit": {
				ValueType:  core.ValueInt64,
				MetricType: core.MetricCumulative,
				IntValue:   123456,
			},
		},
	}

	metricSet2 := core.MetricSet{
		Labels: l2,
		MetricValues: map[string]core.MetricValue{
			"/system.slice/dbus.service//cpu/usage": {
				ValueType:  core.ValueInt64,
				MetricType: core.MetricCumulative,
				IntValue:   123456,
			},
		},
	}

	metricSet3 := core.MetricSet{
		Labels: l3,
		MetricValues: map[string]core.MetricValue{
			"test/metric/1": {
				ValueType:  core.ValueInt64,
				MetricType: core.MetricCumulative,
				IntValue:   123456,
			},
		},
	}

	metricSet4 := core.MetricSet{
		Labels: l4,
		MetricValues: map[string]core.MetricValue{
			"test/metric/1": {
				ValueType:  core.ValueInt64,
				MetricType: core.MetricCumulative,
				IntValue:   123456,
			},
		},
	}

	metricSet5 := core.MetricSet{
		Labels: l5,
		MetricValues: map[string]core.MetricValue{
			"removeme": {
				ValueType:  core.ValueInt64,
				MetricType: core.MetricCumulative,
				IntValue:   123456,
			},
		},
	}

	metricSet6 := core.MetricSet{
		Labels: l,
		MetricValues: map[string]core.MetricValue{
			"cpu/usage": {
				ValueType:  core.ValueInt64,
				MetricType: core.MetricGauge,
				IntValue:   123456,
			},
			"cpu/limit": {
				ValueType:  core.ValueInt64,
				MetricType: core.MetricGauge,
				IntValue:   223456,
			},
		},
	}

	data := core.DataBatch{
		Timestamp: timestamp,
		MetricSets: map[string]*core.MetricSet{
			"pod1": &metricSet1,
			"pod2": &metricSet2,
			"pod3": &metricSet3,
			"pod4": &metricSet4,
			"pod5": &metricSet5,
			"pod6": &metricSet6,
		},
	}

	timeStr, err := timestamp.UTC().MarshalJSON()
	assert.NoError(t, err)

	FakeESSink = NewFakeSink()
	FakeESSink.ExportData(&data)

	//expect msg string
	assert.Equal(t, 2, len(FakeESSink.savedData))

	var expectMsgTemplate = [6]string{
		`{"GeneralMetricsTimestamp":%s,"MetricsTags":{"cluster_name":"default","namespace_id":"","pod_id":"aaaa-bbbb-cccc-dddd"},"MetricsName":"test/metric/1","MetricsValue":{"value":123456}}`,
		`{"GeneralMetricsTimestamp":%s,"MetricsTags":{"cluster_name":"default","namespace_id":"123","pod_id":"aaaa-bbbb-cccc-dddd"},"MetricsName":"removeme","MetricsValue":{"value":123456}}`,
		`{"GeneralMetricsTimestamp":%s,"MetricsTags":{"cluster_name":"default","container_name":"/system.slice/-.mount","namespace_id":"123","pod_id":"aaaa-bbbb-cccc-dddd"},"MetricsName":"/system.slice/-.mount//cpu/limit","MetricsValue":{"value":123456}}`,
		`{"GeneralMetricsTimestamp":%s,"MetricsTags":{"cluster_name":"default","container_name":"/system.slice/dbus.service","namespace_id":"123","pod_id":"aaaa-bbbb-cccc-dddd"},"MetricsName":"/system.slice/dbus.service//cpu/usage","MetricsValue":{"value":123456}}`,
		`{"GeneralMetricsTimestamp":%s,"MetricsTags":{"cluster_name":"default","namespace_id":"123","pod_id":"aaaa-bbbb-cccc-dddd"},"MetricsName":"test/metric/1","MetricsValue":{"value":123456}}`,
		`{"CpuMetricsTimestamp":%s,"Metrics":{"cpu/limit":{"value":223456},"cpu/usage":{"value":123456}},"MetricsTags":{"cluster_name":"default","container_name":"/system.slice/-.mount","namespace_id":"123","pod_id":"aaaa-bbbb-cccc-dddd"}}`,
	}

	msgsString := fmt.Sprintf("%s", FakeESSink.savedData)

	for _, mgsTemplate := range expectMsgTemplate {
		expectMsg := fmt.Sprintf(mgsTemplate, timeStr)
		assert.Contains(t, msgsString, expectMsg)
	}

	FakeESSink = NewFakeSink()
}
