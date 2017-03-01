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

package kafka

import (
	"net/url"
	"sync"
	"time"

	"github.com/golang/glog"
	kafka_common "k8s.io/heapster/common/kafka"
	"k8s.io/heapster/metrics/core"
)

type KafkaSinkPoint struct {
	MetricsName      string
	MetricsValue     interface{}
	MetricsTimestamp time.Time
	MetricsTags      map[string]string
}

type kafkaSink struct {
	kafka_common.KafkaClient
	sync.RWMutex
}

func (sink *kafkaSink) ExportData(dataBatch *core.DataBatch) {
	sink.Lock()
	defer sink.Unlock()

	for _, metricSet := range dataBatch.MetricSets {
		for metricName, metricValue := range metricSet.MetricValues {
			point := KafkaSinkPoint{
				MetricsName: metricName,
				MetricsTags: metricSet.Labels,
				MetricsValue: map[string]interface{}{
					"value": metricValue.GetValue(),
				},
				MetricsTimestamp: dataBatch.Timestamp.UTC(),
			}
			err := sink.ProduceKafkaMessage(point)
			if err != nil {
				glog.Errorf("Failed to produce metric message: %s", err)
			}
		}
		for _, metric := range metricSet.LabeledMetrics {
			labels := make(map[string]string)
			for k, v := range metricSet.Labels {
				labels[k] = v
			}
			for k, v := range metric.Labels {
				labels[k] = v
			}
			point := KafkaSinkPoint{
				MetricsName: metric.Name,
				MetricsTags: labels,
				MetricsValue: map[string]interface{}{
					"value": metric.GetValue(),
				},
				MetricsTimestamp: dataBatch.Timestamp.UTC(),
			}
			err := sink.ProduceKafkaMessage(point)
			if err != nil {
				glog.Errorf("Failed to produce metric message: %s", err)
			}
		}
	}
}

func NewKafkaSink(uri *url.URL) (core.DataSink, error) {
	client, err := kafka_common.NewKafkaClient(uri, kafka_common.TimeSeriesTopic)
	if err != nil {
		return nil, err
	}

	return &kafkaSink{
		KafkaClient: client,
	}, nil
}
