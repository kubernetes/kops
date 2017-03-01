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

package gcm

import (
	"fmt"
	"net/url"
	"sync"
	"time"

	gce_util "k8s.io/heapster/common/gce"
	"k8s.io/heapster/metrics/core"

	gce "cloud.google.com/go/compute/metadata"
	"github.com/golang/glog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gcm "google.golang.org/api/monitoring/v3"
)

const (
	metricDomain    = "kubernetes.io"
	customApiPrefix = "custom.googleapis.com"
	maxNumLabels    = 10
	// The largest number of timeseries we can write to per request.
	maxTimeseriesPerRequest = 200
)

type MetricFilter int8

const (
	metricsAll MetricFilter = iota
	metricsOnlyAutoscaling
)

type gcmSink struct {
	sync.RWMutex
	registered   bool
	project      string
	metricFilter MetricFilter
	gcmService   *gcm.Service
}

func (sink *gcmSink) Name() string {
	return "GCM Sink"
}

func getReq() *gcm.CreateTimeSeriesRequest {
	return &gcm.CreateTimeSeriesRequest{TimeSeries: make([]*gcm.TimeSeries, 0)}
}

func fullMetricName(project string, name string) string {
	return fmt.Sprintf("projects/%s/metricDescriptors/%s/%s/%s", project, customApiPrefix, metricDomain, name)
}

func fullMetricType(name string) string {
	return fmt.Sprintf("%s/%s/%s", customApiPrefix, metricDomain, name)
}

func createTimeSeries(timestamp time.Time, labels map[string]string, metric string, val core.MetricValue, createTime time.Time) *gcm.TimeSeries {
	point := &gcm.Point{
		Interval: &gcm.TimeInterval{
			StartTime: timestamp.Format(time.RFC3339),
			EndTime:   timestamp.Format(time.RFC3339),
		},
		Value: &gcm.TypedValue{},
	}

	var valueType string

	switch val.ValueType {
	case core.ValueInt64:
		point.Value.Int64Value = val.IntValue
		point.Value.ForceSendFields = []string{"Int64Value"}
		valueType = "INT64"
	case core.ValueFloat:
		v := float64(val.FloatValue)
		point.Value.DoubleValue = v
		point.Value.ForceSendFields = []string{"DoubleValue"}
		valueType = "DOUBLE"
	default:
		glog.Errorf("Type not supported %v in %v", val.ValueType, metric)
		return nil
	}
	// For cumulative metric use the provided start time.
	if val.MetricType == core.MetricCumulative {
		point.Interval.StartTime = createTime.Format(time.RFC3339)
	}

	return &gcm.TimeSeries{
		Points: []*gcm.Point{point},
		Metric: &gcm.Metric{
			Type:   fullMetricType(metric),
			Labels: labels,
		},
		ValueType: valueType,
	}
}

func (sink *gcmSink) getTimeSeries(timestamp time.Time, labels map[string]string, metric string, val core.MetricValue, createTime time.Time) *gcm.TimeSeries {
	finalLabels := make(map[string]string)
	if core.IsNodeAutoscalingMetric(metric) {
		// All and autoscaling. Do not populate for other filters.
		if sink.metricFilter != metricsAll &&
			sink.metricFilter != metricsOnlyAutoscaling {
			return nil
		}

		finalLabels[core.LabelHostname.Key] = labels[core.LabelHostname.Key]
		finalLabels[core.LabelGCEResourceID.Key] = labels[core.LabelHostID.Key]
		finalLabels[core.LabelGCEResourceType.Key] = "instance"
	} else {
		// Only all.
		if sink.metricFilter != metricsAll {
			return nil
		}
		supportedLables := core.GcmLabels()
		for key, value := range labels {
			if _, ok := supportedLables[key]; ok {
				finalLabels[key] = value
			}
		}
	}

	return createTimeSeries(timestamp, finalLabels, metric, val, createTime)
}

func (sink *gcmSink) getTimeSeriesForLabeledMetrics(timestamp time.Time, labels map[string]string, metric core.LabeledMetric, createTime time.Time) *gcm.TimeSeries {
	// Only all. There are no autoscaling labeled metrics.
	if sink.metricFilter != metricsAll {
		return nil
	}

	finalLabels := make(map[string]string)
	supportedLables := core.GcmLabels()
	for key, value := range labels {
		if _, ok := supportedLables[key]; ok {
			finalLabels[key] = value
		}
	}
	for key, value := range metric.Labels {
		if _, ok := supportedLables[key]; ok {
			finalLabels[key] = value
		}
	}

	return createTimeSeries(timestamp, finalLabels, metric.Name, metric.MetricValue, createTime)
}

func fullProjectName(name string) string {
	return fmt.Sprintf("projects/%s", name)
}

func (sink *gcmSink) sendRequest(req *gcm.CreateTimeSeriesRequest) {
	_, err := sink.gcmService.Projects.TimeSeries.Create(fullProjectName(sink.project), req).Do()
	if err != nil {
		glog.Errorf("Error while sending request to GCM %v", err)
	} else {
		glog.V(4).Infof("Successfully sent %v timeserieses to GCM", len(req.TimeSeries))
	}
}

func (sink *gcmSink) ExportData(dataBatch *core.DataBatch) {
	if err := sink.registerAllMetrics(); err != nil {
		glog.Warningf("Error during metrics registration: %v", err)
		return
	}

	req := getReq()
	for _, metricSet := range dataBatch.MetricSets {
		for metric, val := range metricSet.MetricValues {
			point := sink.getTimeSeries(dataBatch.Timestamp, metricSet.Labels, metric, val, metricSet.CreateTime)
			if point != nil {
				req.TimeSeries = append(req.TimeSeries, point)
			}
			if len(req.TimeSeries) >= maxTimeseriesPerRequest {
				sink.sendRequest(req)
				req = getReq()
			}
		}
		for _, metric := range metricSet.LabeledMetrics {
			point := sink.getTimeSeriesForLabeledMetrics(dataBatch.Timestamp, metricSet.Labels, metric, metricSet.CreateTime)
			if point != nil {
				req.TimeSeries = append(req.TimeSeries, point)
			}
			if len(req.TimeSeries) >= maxTimeseriesPerRequest {
				sink.sendRequest(req)
				req = getReq()
			}
		}
	}
	if len(req.TimeSeries) > 0 {
		sink.sendRequest(req)
	}
}

func (sink *gcmSink) Stop() {
	// nothing needs to be done.
}

func (sink *gcmSink) registerAllMetrics() error {
	return sink.register(core.AllMetrics)
}

// Adds the specified metrics or updates them if they already exist.
func (sink *gcmSink) register(metrics []core.Metric) error {
	sink.Lock()
	defer sink.Unlock()
	if sink.registered {
		return nil
	}

	for _, metric := range metrics {
		metricName := fullMetricName(sink.project, metric.MetricDescriptor.Name)
		metricType := fullMetricType(metric.MetricDescriptor.Name)

		if _, err := sink.gcmService.Projects.MetricDescriptors.Delete(metricName).Do(); err != nil {
			glog.Infof("[GCM] Deleting metric %v failed: %v", metricName, err)
		}
		labels := make([]*gcm.LabelDescriptor, 0)

		// Node autoscaling metrics have special labels.
		if core.IsNodeAutoscalingMetric(metric.MetricDescriptor.Name) {
			// All and autoscaling. Do not populate for other filters.
			if sink.metricFilter != metricsAll &&
				sink.metricFilter != metricsOnlyAutoscaling {
				continue
			}

			for _, l := range core.GcmNodeAutoscalingLabels() {
				labels = append(labels, &gcm.LabelDescriptor{
					Key:         l.Key,
					Description: l.Description,
				})
			}
		} else {
			// Only all.
			if sink.metricFilter != metricsAll {
				continue
			}

			for _, l := range core.GcmLabels() {
				labels = append(labels, &gcm.LabelDescriptor{
					Key:         l.Key,
					Description: l.Description,
				})
			}
		}

		var metricKind string

		switch metric.MetricDescriptor.Type {
		case core.MetricCumulative:
			metricKind = "CUMULATIVE"
		case core.MetricGauge:
			metricKind = "GAUGE"
		case core.MetricDelta:
			metricKind = "DELTA"
		}

		var valueType string

		switch metric.MetricDescriptor.ValueType {
		case core.ValueInt64:
			valueType = "INT64"
		case core.ValueFloat:
			valueType = "DOUBLE"
		}

		desc := &gcm.MetricDescriptor{
			Name:        metricName,
			Description: metric.MetricDescriptor.Description,
			Labels:      labels,
			MetricKind:  metricKind,
			ValueType:   valueType,
			Type:        metricType,
		}

		if _, err := sink.gcmService.Projects.MetricDescriptors.Create(fullProjectName(sink.project), desc).Do(); err != nil {
			glog.Errorf("Metric registration of %v failed: %v", desc.Name, err)
			return err
		}
	}
	sink.registered = true
	return nil
}

func CreateGCMSink(uri *url.URL) (core.DataSink, error) {
	if len(uri.Scheme) > 0 {
		return nil, fmt.Errorf("scheme should not be set for GCM sink")
	}
	if len(uri.Host) > 0 {
		return nil, fmt.Errorf("host should not be set for GCM sink")
	}

	opts, err := url.ParseQuery(uri.RawQuery)

	metrics := "all"
	if len(opts["metrics"]) > 0 {
		metrics = opts["metrics"][0]
	}
	var metricFilter MetricFilter = metricsAll
	switch metrics {
	case "all":
		metricFilter = metricsAll
	case "autoscaling":
		metricFilter = metricsOnlyAutoscaling
	default:
		return nil, fmt.Errorf("invalid metrics parameter: %s", metrics)
	}

	if err := gce_util.EnsureOnGCE(); err != nil {
		return nil, err
	}

	// Detect project ID
	projectId, err := gce.ProjectID()
	if err != nil {
		return nil, err
	}

	// Create Google Cloud Monitoring service.
	client := oauth2.NewClient(oauth2.NoContext, google.ComputeTokenSource(""))
	gcmService, err := gcm.New(client)
	if err != nil {
		return nil, err
	}

	sink := &gcmSink{
		registered:   false,
		project:      projectId,
		gcmService:   gcmService,
		metricFilter: metricFilter,
	}
	glog.Infof("created GCM sink")
	if err := sink.registerAllMetrics(); err != nil {
		glog.Warningf("Error during metrics registration: %v", err)
	}
	return sink, nil
}
