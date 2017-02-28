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

package stackdriver

import (
	"fmt"
	"net/url"
	"time"

	gce "cloud.google.com/go/compute/metadata"
	"github.com/golang/glog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	sd_api "google.golang.org/api/monitoring/v3"
	gce_util "k8s.io/heapster/common/gce"
	"k8s.io/heapster/metrics/core"
)

const (
	maxTimeseriesPerRequest = 200
)

type stackdriverSink struct {
	project           string
	zone              string
	stackdriverClient *sd_api.Service
}

type metricMetadata struct {
	MetricKind string
	ValueType  string
	Name       string
}

var (
	uptimeMD = &metricMetadata{
		MetricKind: "CUMULATIVE",
		ValueType:  "DOUBLE",
		Name:       "container.googleapis.com/container/uptime",
	}
)

func (sink *stackdriverSink) Name() string {
	return "Stackdriver Sink"
}

func (sink *stackdriverSink) Stop() {
	// nothing needs to be done
}

func (sink *stackdriverSink) ExportData(dataBatch *core.DataBatch) {
	req := getReq()

	for _, metricSet := range dataBatch.MetricSets {
		for name, value := range metricSet.MetricValues {
			point := sink.translateMetric(dataBatch.Timestamp, metricSet.Labels, name, value, metricSet.CreateTime)

			if point != nil {
				req.TimeSeries = append(req.TimeSeries, point)
			}
			if len(req.TimeSeries) >= maxTimeseriesPerRequest {
				sink.sendRequest(req)
				req = getReq()
			}
		}
	}
}

func CreateStackdriverSink(uri *url.URL) (core.DataSink, error) {
	if len(uri.Scheme) > 0 {
		return nil, fmt.Errorf("Scheme should not be set for Stackdriver sink")
	}
	if len(uri.Host) > 0 {
		return nil, fmt.Errorf("Host should not be set for Stackdriver sink")
	}

	if err := gce_util.EnsureOnGCE(); err != nil {
		return nil, err
	}

	// Detect project ID
	projectId, err := gce.ProjectID()
	if err != nil {
		return nil, err
	}

	// Detect zone
	zone, err := gce.Zone()
	if err != nil {
		return nil, err
	}

	// Create Google Cloud Monitoring service
	client := oauth2.NewClient(oauth2.NoContext, google.ComputeTokenSource(""))
	stackdriverClient, err := sd_api.New(client)
	if err != nil {
		return nil, err
	}

	sink := &stackdriverSink{
		project:           projectId,
		zone:              zone,
		stackdriverClient: stackdriverClient,
	}

	glog.Infof("Created Stackdriver sink")

	return sink, nil
}

func (sink *stackdriverSink) sendRequest(req *sd_api.CreateTimeSeriesRequest) {
	_, err := sink.stackdriverClient.Projects.TimeSeries.Create(fullProjectName(sink.project), req).Do()
	if err != nil {
		glog.Errorf("Error while sending request to Stackdriver %v", err)
	} else {
		glog.V(4).Infof("Successfully sent %v timeseries to Stackdriver", len(req.TimeSeries))
	}
}

func (sink *stackdriverSink) translateMetric(timestamp time.Time, labels map[string]string, name string, value core.MetricValue, createTime time.Time) *sd_api.TimeSeries {
	switch name {
	case core.MetricUptime.MetricDescriptor.Name:
		point := sink.uptimePoint(timestamp, createTime, value)
		resourceLabels := sink.getResourceLabels(labels)
		return createTimeSeries(resourceLabels, uptimeMD, point)
	default:
		return nil
	}
}

func (sink *stackdriverSink) getResourceLabels(labels map[string]string) map[string]string {
	return map[string]string{
		"project_id":     sink.project,
		"cluster_name":   "",
		"zone":           sink.zone,
		"instance_id":    labels[core.LabelHostID.Key],
		"namespace_id":   labels[core.LabelPodNamespaceUID.Key],
		"pod_id":         labels[core.LabelPodId.Key],
		"container_name": labels[core.LabelContainerName.Key],
	}
}

func createTimeSeries(resourceLabels map[string]string, metadata *metricMetadata, point *sd_api.Point) *sd_api.TimeSeries {
	return &sd_api.TimeSeries{
		Metric: &sd_api.Metric{
			Type: metadata.Name,
		},
		MetricKind: metadata.MetricKind,
		ValueType:  metadata.ValueType,
		Resource: &sd_api.MonitoredResource{
			Labels: resourceLabels,
			Type:   "gke_container",
		},
		Points: []*sd_api.Point{point},
	}
}

func (sink *stackdriverSink) uptimePoint(timestamp time.Time, createTime time.Time, value core.MetricValue) *sd_api.Point {
	return &sd_api.Point{
		Interval: &sd_api.TimeInterval{
			EndTime:   timestamp.Format(time.RFC3339),
			StartTime: createTime.Format(time.RFC3339),
		},
		Value: &sd_api.TypedValue{
			DoubleValue: float64(value.IntValue) / float64(time.Second/time.Millisecond),
		},
	}
}

func fullProjectName(name string) string {
	return fmt.Sprintf("projects/%s", name)
}

func getReq() *sd_api.CreateTimeSeriesRequest {
	return &sd_api.CreateTimeSeriesRequest{TimeSeries: make([]*sd_api.TimeSeries, 0)}
}
