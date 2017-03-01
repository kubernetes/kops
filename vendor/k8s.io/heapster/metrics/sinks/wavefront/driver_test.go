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

package wavefront

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/heapster/metrics/core"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

var (
	fakeNodeIp  = "192.168.1.23"
	fakePodName = "redis-test"
	fakePodUid  = "redis-test-uid"
	fakeLabel   = map[string]string{
		"name": "redis",
		"io.kubernetes.pod.name": "default/redis-test",
		"pod_id":                 fakePodUid,
		"namespace_name":         "default",
		"pod_name":               fakePodName,
		"container_name":         "redis",
		"container_base_image":   "kubernetes/redis:v1",
		"namespace_id":           "namespace-test-uid",
		"host_id":                fakeNodeIp,
		"hostname":               fakeNodeIp,
	}
)

func NewFakeWavefrontSink() *wavefrontSink {
	return &wavefrontSink{
		testMode:          true,
		ClusterName:       "testCluster",
		IncludeLabels:     false,
		IncludeContainers: true,
	}
}

func TestStoreTimeseriesEmptyInput(t *testing.T) {
	fakeSink := NewFakeWavefrontSink()
	db := core.DataBatch{}
	fakeSink.ExportData(&db)
	assert.Equal(t, 0, len(fakeSink.testReceivedLines))
}

func TestStoreTimeseriesMultipleTimeseriesInput(t *testing.T) {
	fakeSink := NewFakeWavefrontSink()
	batch := generateFakeBatch()
	fakeSink.ExportData(batch)
	assert.Equal(t, len(batch.MetricSets), len(fakeSink.testReceivedLines))
}
func TestName(t *testing.T) {
	fakeSink := NewFakeWavefrontSink()
	name := fakeSink.Name()
	assert.Equal(t, name, "Wavefront Sink")
}

func TestValidateLines(t *testing.T) {
	fakeSink := NewFakeWavefrontSink()
	batch := generateFakeBatch()
	fakeSink.ExportData(batch)

	//validate each line received from the fake batch
	for _, line := range fakeSink.testReceivedLines {
		parts := strings.Split(strings.TrimSpace(line), " ")

		//second part should always be the numeric metric value
		_, err := strconv.ParseFloat(parts[1], 64)
		assert.NoError(t, err)

		//third part should always be the epoch timestamp (a count of seconds)
		_, err = strconv.ParseInt(parts[2], 0, 64)
		assert.NoError(t, err)

		//the fourth part should be the source tag
		isSourceTag := strings.HasPrefix(parts[3], "source=")
		assert.True(t, isSourceTag)

		//all remaining parts are tags and must be key value pairs (containing "=")
		tags := parts[4:]
		for _, v := range tags {
			assert.True(t, strings.Contains(v, "="))
		}
	}
}

func TestCreateWavefrontSinkWithNoEmptyInputs(t *testing.T) {
	fakeUrl := "wavefront-proxy:2878?clusterName=testCluster&prefix=testPrefix&includeLabels=true&includeContainers=true"
	uri, _ := url.Parse(fakeUrl)
	sink, err := NewWavefrontSink(uri)
	assert.NoError(t, err)
	assert.NotNil(t, sink)
	wfSink, ok := sink.(*wavefrontSink)
	assert.Equal(t, true, ok)
	assert.Equal(t, "wavefront-proxy:2878", wfSink.ProxyAddress)
	assert.Equal(t, "testCluster", wfSink.ClusterName)
	assert.Equal(t, "testPrefix", wfSink.Prefix)
	assert.Equal(t, true, wfSink.IncludeLabels)
	assert.Equal(t, true, wfSink.IncludeContainers)
}

func generateFakeBatch() *core.DataBatch {
	batch := core.DataBatch{
		Timestamp:  time.Now(),
		MetricSets: map[string]*core.MetricSet{},
	}

	batch.MetricSets["m1"] = generateMetricSet("cpu/limit", core.MetricGauge, 1000)
	batch.MetricSets["m2"] = generateMetricSet("cpu/usage", core.MetricCumulative, 43363664)
	batch.MetricSets["m3"] = generateMetricSet("filesystem/limit", core.MetricGauge, 42241163264)
	batch.MetricSets["m4"] = generateMetricSet("filesystem/usage", core.MetricGauge, 32768)
	batch.MetricSets["m5"] = generateMetricSet("memory/limit", core.MetricGauge, -1)
	batch.MetricSets["m6"] = generateMetricSet("memory/usage", core.MetricGauge, 487424)
	batch.MetricSets["m7"] = generateMetricSet("memory/working_set", core.MetricGauge, 491520)
	batch.MetricSets["m8"] = generateMetricSet("uptime", core.MetricCumulative, 910823)
	return &batch
}

func generateMetricSet(name string, metricType core.MetricType, value int64) *core.MetricSet {
	set := &core.MetricSet{
		Labels: fakeLabel,
		MetricValues: map[string]core.MetricValue{
			name: {
				MetricType: metricType,
				ValueType:  core.ValueInt64,
				IntValue:   value,
			},
		},
	}
	return set
}
