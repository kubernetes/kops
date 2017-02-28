// Copyright 2016 Google Inc. All Rights Reserved.
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

package graphite

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"k8s.io/heapster/metrics/core"

	"github.com/golang/glog"
	"github.com/marpaia/graphite-golang"
)

const (
	DefaultHost   = "localhost"
	DefaultPort   = 2003
	DefaultPrefix = "kubernetes"
)

type graphiteClient interface {
	Connect() error
	Disconnect() error
	SendMetrics(metric []graphite.Metric) error
}

type graphiteMetric struct {
	name      string
	value     core.MetricValue
	labels    map[string]string
	timestamp int64
}

var escapeFieldReplacer = strings.NewReplacer(".", "_", "/", "_")

func escapeField(f string) string {
	return escapeFieldReplacer.Replace(f)
}

func (m *graphiteMetric) Path() string {
	var metricPath string
	if resourceId, ok := m.labels["resourceId"]; ok {
		nameParts := strings.Split(m.name, "/")
		section, parts := nameParts[0], nameParts[1:]
		metricPath = strings.Join(append([]string{section, escapeField(resourceId)}, parts...), ".")
	} else {
		metricPath = m.name
	}
	metricPath = strings.Replace(metricPath, "/", ".", -1)
	if t, ok := m.labels[core.LabelMetricSetType.Key]; ok {
		switch t {
		case core.MetricSetTypePodContainer:
			return fmt.Sprintf("nodes.%s.pods.%s.%s.containers.%s.%s",
				escapeField(m.labels[core.LabelHostname.Key]),
				m.labels[core.LabelNamespaceName.Key],
				escapeField(m.labels[core.LabelPodName.Key]),
				escapeField(m.labels[core.LabelContainerName.Key]),
				metricPath,
			)
		case core.MetricSetTypeSystemContainer:
			return fmt.Sprintf("nodes.%s.sys-containers.%s.%s",
				escapeField(m.labels[core.LabelHostname.Key]),
				escapeField(m.labels[core.LabelContainerName.Key]),
				metricPath,
			)
		case core.MetricSetTypePod:
			return fmt.Sprintf("nodes.%s.pods.%s.%s.%s",
				escapeField(m.labels[core.LabelHostname.Key]),
				m.labels[core.LabelNamespaceName.Key],
				escapeField(m.labels[core.LabelPodName.Key]),
				metricPath,
			)
		case core.MetricSetTypeNamespace:
			return fmt.Sprintf("namespaces.%s.%s",
				m.labels[core.LabelNamespaceName.Key],
				metricPath,
			)
		case core.MetricSetTypeNode:
			return fmt.Sprintf("nodes.%s.%s",
				escapeField(m.labels[core.LabelHostname.Key]),
				metricPath,
			)
		case core.MetricSetTypeCluster:
			return fmt.Sprintf("cluster.%s", metricPath)
		default:
			glog.V(6).Infof("Unknown metric type %s", t)
		}
	}
	return metricPath
}

func (m *graphiteMetric) Value() string {
	switch m.value.ValueType {
	case core.ValueInt64:
		return fmt.Sprintf("%d", m.value.IntValue)
	case core.ValueFloat:
		return fmt.Sprintf("%f", m.value.FloatValue)
	}
	return ""
}

func (m *graphiteMetric) Metric() graphite.Metric {
	return graphite.NewMetric(m.Path(), m.Value(), m.timestamp)
}

type Sink struct {
	client graphiteClient
	sync.RWMutex
}

func NewGraphiteSink(uri *url.URL) (core.DataSink, error) {
	host, portString, err := net.SplitHostPort(uri.Host)
	if err != nil {
		return nil, err
	}
	if host == "" {
		host = DefaultHost
	}
	port := DefaultPort
	if portString != "" {
		if port, err = strconv.Atoi(portString); err != nil {
			return nil, err
		}
	}

	prefix := uri.Query().Get("prefix")
	if prefix == "" {
		prefix = DefaultPrefix
	}

	client, err := graphite.GraphiteFactory(uri.Scheme, host, port, prefix)
	if err != nil {
		return nil, err
	}
	return &Sink{client: client}, nil
}

func (s *Sink) Name() string {
	return "Graphite Sink"
}

func (s *Sink) ExportData(dataBatch *core.DataBatch) {
	s.Lock()
	defer s.Unlock()
	var metrics []graphite.Metric
	for _, metricSet := range dataBatch.MetricSets {
		var m *graphiteMetric
		for metricName, metricValue := range metricSet.MetricValues {
			m = &graphiteMetric{
				name:      metricName,
				value:     metricValue,
				labels:    metricSet.Labels,
				timestamp: dataBatch.Timestamp.Unix(),
			}
			metrics = append(metrics, m.Metric())
		}
		for _, metric := range metricSet.LabeledMetrics {
			if value := metric.GetValue(); value != nil {
				labels := make(map[string]string)
				for k, v := range metricSet.Labels {
					labels[k] = v
				}
				for k, v := range metric.Labels {
					labels[k] = v
				}
				m = &graphiteMetric{
					name:      metric.Name,
					value:     metric.MetricValue,
					labels:    labels,
					timestamp: dataBatch.Timestamp.Unix(),
				}
				metrics = append(metrics, m.Metric())
			}
		}
	}
	glog.V(8).Infof("Sending %d events to graphite", len(metrics))
	if err := s.client.SendMetrics(metrics); err != nil {
		glog.V(2).Info("There were errors sending events to Graphite, reconecting")
		s.client.Disconnect()
		s.client.Connect()
	}
}

func (s *Sink) Stop() {
	s.client.Disconnect()
}
