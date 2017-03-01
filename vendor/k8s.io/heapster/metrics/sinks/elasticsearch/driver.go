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
	"net/url"
	"sync"
	"time"

	"github.com/golang/glog"
	esCommon "k8s.io/heapster/common/elasticsearch"
	"k8s.io/heapster/metrics/core"
	"reflect"
)

// SaveDataFunc is a pluggable function to enforce limits on the object
type SaveDataFunc func(date time.Time, typeName string, sinkData []interface{}) error

type elasticSearchSink struct {
	esSvc     esCommon.ElasticSearchService
	saveData  SaveDataFunc
	flushData func() error
	sync.RWMutex
}

type EsFamilyPoints map[core.MetricFamily][]interface{}
type esPointTags map[string]string
type customTimestamp map[string]time.Time

type EsSinkPointGeneral struct {
	GeneralMetricsTimestamp time.Time
	MetricsTags             esPointTags
	MetricsName             string
	MetricsValue            interface{}
}
type EsSinkPointFamily map[string]interface{}

func (sink *elasticSearchSink) ExportData(dataBatch *core.DataBatch) {
	sink.Lock()
	defer sink.Unlock()

	for _, metricSet := range dataBatch.MetricSets {
		familyPoints := EsFamilyPoints{}

		for metricName, metricValue := range metricSet.MetricValues {
			familyPoints = addMetric(familyPoints, metricName, dataBatch.Timestamp, metricSet.Labels, metricValue.GetValue(), sink.esSvc.ClusterName)
		}
		for _, metric := range metricSet.LabeledMetrics {
			labels := make(map[string]string)
			for k, v := range metricSet.Labels {
				labels[k] = v
			}
			for k, v := range metric.Labels {
				labels[k] = v
			}

			familyPoints = addMetric(familyPoints, metric.Name, dataBatch.Timestamp, labels, metric.GetValue(), sink.esSvc.ClusterName)
		}

		for family, dataPoints := range familyPoints {
			err := sink.saveData(dataBatch.Timestamp.UTC(), string(family), dataPoints)
			if err != nil {
				glog.Warningf("Failed to export data to ElasticSearch sink: %v", err)
			}
		}
		err := sink.flushData()
		if err != nil {
			glog.Warningf("Failed to flushing data to ElasticSearch sink: %v", err)
		}
	}
}

func addMetric(points EsFamilyPoints, metricName string, date time.Time, tags esPointTags, value interface{}, clusterName string) EsFamilyPoints {
	family := core.MetricFamilyForName(metricName)

	if points[family] == nil {
		points[family] = []interface{}{}
	}

	if family == core.MetricFamilyGeneral {
		point := EsSinkPointGeneral{}
		point.MetricsTags = tags
		point.MetricsTags["cluster_name"] = clusterName
		point.GeneralMetricsTimestamp = date.UTC()
		point.MetricsName = metricName
		point.MetricsValue = EsPointValue(value)

		//add
		points[family] = append(points[family], point)
		return points
	}

	for idx, pt := range points[family] {
		if point, ok := pt.(EsSinkPointFamily); ok {
			if point[esCommon.MetricFamilyTimestamp(family)] == date.UTC() && reflect.DeepEqual(point["MetricsTags"], tags) {
				if metrics, ok := point["Metrics"].(map[string]interface{}); ok {
					metrics[metricName] = EsPointValue(value)
					point["Metrics"] = metrics
				} else {
					glog.Warningf("Failed to cast metrics to map")
				}

				if tags, ok := point["MetricsTags"].(esPointTags); ok {
					tags["cluster_name"] = clusterName
					point["MetricsTags"] = tags
				} else {
					glog.Warningf("Failed to cast metricstags to map")
				}

				//add
				points[family][idx] = point
				return points
			}
		}
	}

	point := EsSinkPointFamily{}
	point[esCommon.MetricFamilyTimestamp(family)] = date.UTC()
	tags["cluster_name"] = clusterName
	point["MetricsTags"] = tags
	metrics := make(map[string]interface{})
	metrics[metricName] = EsPointValue(value)
	point["Metrics"] = metrics

	//add
	points[family] = append(points[family], point)
	return points
}

func EsPointValue(value interface{}) interface{} {
	return map[string]interface{}{
		"value": value,
	}
}

func (sink *elasticSearchSink) Name() string {
	return "ElasticSearch Sink"
}

func (sink *elasticSearchSink) Stop() {
	// nothing needs to be done.
}

func NewElasticSearchSink(uri *url.URL) (core.DataSink, error) {
	var esSink elasticSearchSink
	esSvc, err := esCommon.CreateElasticSearchService(uri)
	if err != nil {
		glog.Warningf("Failed to config ElasticSearch: %v", err)
		return nil, err
	}

	esSink.esSvc = *esSvc
	esSink.saveData = func(date time.Time, typeName string, sinkData []interface{}) error {
		return esSvc.SaveData(date, typeName, sinkData)
	}
	esSink.flushData = func() error {
		return esSvc.FlushData()
	}

	glog.V(2).Info("ElasticSearch sink setup successfully")
	return &esSink, nil
}
