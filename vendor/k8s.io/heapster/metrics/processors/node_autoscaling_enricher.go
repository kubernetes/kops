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

package processors

import (
	"net/url"

	"k8s.io/apimachinery/pkg/labels"
	kube_client "k8s.io/client-go/kubernetes"
	v1listers "k8s.io/client-go/listers/core/v1"
	kube_api "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
	kube_config "k8s.io/heapster/common/kubernetes"
	"k8s.io/heapster/metrics/core"
	"k8s.io/heapster/metrics/util"
)

type NodeAutoscalingEnricher struct {
	nodeLister v1listers.NodeLister
	reflector  *cache.Reflector
}

func (this *NodeAutoscalingEnricher) Name() string {
	return "node_autoscaling_enricher"
}

func (this *NodeAutoscalingEnricher) Process(batch *core.DataBatch) (*core.DataBatch, error) {
	nodes, err := this.nodeLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}
	for _, node := range nodes {
		if metricSet, found := batch.MetricSets[core.NodeKey(node.Name)]; found {
			metricSet.Labels[core.LabelLabels.Key] = util.LabelsToString(node.Labels)
			capacityCpu, _ := node.Status.Capacity[kube_api.ResourceCPU]
			capacityMem, _ := node.Status.Capacity[kube_api.ResourceMemory]
			allocatableCpu, _ := node.Status.Allocatable[kube_api.ResourceCPU]
			allocatableMem, _ := node.Status.Allocatable[kube_api.ResourceMemory]

			cpuRequested := getInt(metricSet, &core.MetricCpuRequest)
			cpuUsed := getInt(metricSet, &core.MetricCpuUsageRate)
			memRequested := getInt(metricSet, &core.MetricMemoryRequest)
			memUsed := getInt(metricSet, &core.MetricMemoryUsage)

			if allocatableCpu.MilliValue() != 0 {
				setFloat(metricSet, &core.MetricNodeCpuUtilization, float32(cpuUsed)/float32(allocatableCpu.MilliValue()))
				setFloat(metricSet, &core.MetricNodeCpuReservation, float32(cpuRequested)/float32(allocatableCpu.MilliValue()))
			}
			setFloat(metricSet, &core.MetricNodeCpuCapacity, float32(capacityCpu.MilliValue()))
			setFloat(metricSet, &core.MetricNodeCpuAllocatable, float32(allocatableCpu.MilliValue()))

			if allocatableMem.Value() != 0 {
				setFloat(metricSet, &core.MetricNodeMemoryUtilization, float32(memUsed)/float32(allocatableMem.Value()))
				setFloat(metricSet, &core.MetricNodeMemoryReservation, float32(memRequested)/float32(allocatableMem.Value()))
			}
			setFloat(metricSet, &core.MetricNodeMemoryCapacity, float32(capacityMem.Value()))
			setFloat(metricSet, &core.MetricNodeMemoryAllocatable, float32(allocatableMem.Value()))
		}
	}
	return batch, nil
}

func getInt(metricSet *core.MetricSet, metric *core.Metric) int64 {
	if value, found := metricSet.MetricValues[metric.MetricDescriptor.Name]; found {
		return value.IntValue
	}
	return 0
}

func setFloat(metricSet *core.MetricSet, metric *core.Metric, value float32) {
	metricSet.MetricValues[metric.MetricDescriptor.Name] = core.MetricValue{
		MetricType: core.MetricGauge,
		ValueType:  core.ValueFloat,
		FloatValue: value,
	}
}

func NewNodeAutoscalingEnricher(url *url.URL) (*NodeAutoscalingEnricher, error) {
	kubeConfig, err := kube_config.GetKubeClientConfig(url)
	if err != nil {
		return nil, err
	}
	kubeClient := kube_client.NewForConfigOrDie(kubeConfig)

	// watch nodes
	nodeLister, reflector, _ := util.GetNodeLister(kubeClient)

	return &NodeAutoscalingEnricher{
		nodeLister: nodeLister,
		reflector:  reflector,
	}, nil
}
