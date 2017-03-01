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

package app

import (
	"fmt"
	"time"

	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/api/errors"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/heapster/metrics/core"
	metricsink "k8s.io/heapster/metrics/sinks/metric"
	"k8s.io/heapster/metrics/storage/util"
	"k8s.io/metrics/pkg/apis/metrics"
	_ "k8s.io/metrics/pkg/apis/metrics/install"
)

type MetricStorage struct {
	groupResource schema.GroupResource
	metricSink    *metricsink.MetricSink
	nodeLister    v1listers.NodeLister
}

var _ rest.KindProvider = &MetricStorage{}
var _ rest.Storage = &MetricStorage{}
var _ rest.Getter = &MetricStorage{}
var _ rest.Lister = &MetricStorage{}

func NewStorage(groupResource schema.GroupResource, metricSink *metricsink.MetricSink, nodeLister v1listers.NodeLister) *MetricStorage {
	return &MetricStorage{
		groupResource: groupResource,
		metricSink:    metricSink,
		nodeLister:    nodeLister,
	}
}

// Storage interface
func (m *MetricStorage) New() runtime.Object {
	return &metrics.NodeMetrics{}
}

// KindProvider interface
func (m *MetricStorage) Kind() string {
	return "NodeMetrics"
}

// Lister interface
func (m *MetricStorage) NewList() runtime.Object {
	return &metrics.NodeMetricsList{}
}

// Lister interface
func (m *MetricStorage) List(ctx genericapirequest.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	labelSelector := labels.Everything()
	if options != nil && options.LabelSelector != nil {
		labelSelector = options.LabelSelector
	}
	nodes, err := m.nodeLister.ListWithPredicate(func(node *v1.Node) bool {
		if labelSelector.Empty() {
			return true
		}
		return labelSelector.Matches(labels.Set(node.Labels))
	})
	if err != nil {
		errMsg := fmt.Errorf("Error while listing nodes: %v", err)
		glog.Error(errMsg)
		return &metrics.NodeMetricsList{}, errMsg
	}

	res := metrics.NodeMetricsList{}
	for _, node := range nodes {
		if m := m.getNodeMetrics(node.Name); m != nil {
			res.Items = append(res.Items, *m)
		}
	}
	return &res, nil
}

// Getter interface
func (m *MetricStorage) Get(ctx genericapirequest.Context, name string, opts *metav1.GetOptions) (runtime.Object, error) {
	// TODO: pay attention to get options
	nodeMetrics := m.getNodeMetrics(name)
	if nodeMetrics == nil {
		return &metrics.NodeMetrics{}, errors.NewNotFound(m.groupResource, name)
	}
	return nodeMetrics, nil
}

func (m *MetricStorage) getNodeMetrics(node string) *metrics.NodeMetrics {
	batch := m.metricSink.GetLatestDataBatch()
	if batch == nil {
		return nil
	}

	ms, found := batch.MetricSets[core.NodeKey(node)]
	if !found {
		return nil
	}

	usage, err := util.ParseResourceList(ms)
	if err != nil {
		return nil
	}

	return &metrics.NodeMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name:              node,
			CreationTimestamp: metav1.NewTime(time.Now()),
		},
		Timestamp: metav1.NewTime(batch.Timestamp),
		Window:    metav1.Duration{Duration: time.Minute},
		Usage:     usage,
	}
}
