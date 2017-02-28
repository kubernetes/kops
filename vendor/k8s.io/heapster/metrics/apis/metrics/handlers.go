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

// The file path is compatible with Kubernetes standards. This not a requirement
// right now but in the future we want to reuse apiserver code, which
// requires it.

package metrics

import (
	"fmt"
	"net/http"
	"time"

	restful "github.com/emicklei/go-restful"
	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	v1listers "k8s.io/client-go/listers/core/v1"
	kube_v1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/heapster/metrics/apis/metrics/v1alpha1"
	"k8s.io/heapster/metrics/core"
	metricsink "k8s.io/heapster/metrics/sinks/metric"
)

type Api struct {
	metricSink *metricsink.MetricSink
	podLister  v1listers.PodLister
	nodeLister v1listers.NodeLister
}

func NewApi(metricSink *metricsink.MetricSink, podLister v1listers.PodLister, nodeLister v1listers.NodeLister) *Api {
	return &Api{
		metricSink: metricSink,
		podLister:  podLister,
		nodeLister: nodeLister,
	}
}

func (a *Api) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.Path("/apis/metrics/v1alpha1").
		Doc("Root endpoint of metrics API").
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/nodes/").
		To(a.nodeMetricsList).
		Doc("Get a list of metrics for all available nodes.").
		Operation("nodeMetricsList")).
		Param(ws.QueryParameter("labelSelector", "A selector to restrict the list of returned objects by their labels. Defaults to everything.").DataType("string"))

	ws.Route(ws.GET("/nodes/{node-name}/").
		To(a.nodeMetrics).
		Doc("Get a list of all available metrics for the specified node.").
		Operation("nodeMetrics").
		Param(ws.PathParameter("node-name", "The name of the node to lookup").DataType("string")))

	ws.Route(ws.GET("/pods/").
		To(a.allPodMetricsList).
		Doc("Get metrics for all available pods.").
		Operation("allPodMetricsList"))

	ws.Route(ws.GET("/namespaces/{namespace-name}/pods/").
		To(a.podMetricsList).
		Doc("Get a list of metrics for all available pods in the specified namespace.").
		Operation("podMetricsList").
		Param(ws.PathParameter("namespace-name", "The name of the namespace to lookup").DataType("string"))).
		Param(ws.QueryParameter("labelSelector", "A selector to restrict the list of returned objects by their labels. Defaults to everything.").DataType("string"))

	ws.Route(ws.GET("/namespaces/{namespace-name}/pods/{pod-name}/").
		To(a.podMetrics).
		Doc("Get metrics for the specified pod in the specified namespace.").
		Operation("podMetrics").
		Param(ws.PathParameter("namespace-name", "The name of the namespace to lookup").DataType("string")).
		Param(ws.PathParameter("pod-name", "The name of the pod to lookup").DataType("string")))

	container.Add(ws)
}

func (a *Api) nodeMetricsList(request *restful.Request, response *restful.Response) {
	selector := request.QueryParameter("labelSelector")

	labelSelector, err := labels.Parse(selector)
	if err != nil {
		errMsg := fmt.Errorf("Error while parsing selector %v: %v", selector, err)
		glog.Error(errMsg)
		response.WriteError(http.StatusBadRequest, errMsg)
		return
	}

	nodes, err := a.nodeLister.ListWithPredicate(func(node *kube_v1.Node) bool {
		if labelSelector.Empty() {
			return true
		}
		return labelSelector.Matches(labels.Set(node.Labels))
	})
	if err != nil {
		errMsg := fmt.Errorf("Error while listing nodes: %v", err)
		glog.Error(errMsg)
		response.WriteError(http.StatusInternalServerError, errMsg)
		return
	}

	res := v1alpha1.NodeMetricsList{}
	for _, node := range nodes {
		if m := a.getNodeMetrics(node.Name); m != nil {
			res.Items = append(res.Items, *m)
		}
	}
	response.WriteEntity(&res)
}

func (a *Api) nodeMetrics(request *restful.Request, response *restful.Response) {
	node := request.PathParameter("node-name")
	m := a.getNodeMetrics(node)
	if m == nil {
		response.WriteError(http.StatusNotFound, fmt.Errorf("No metrics for node %v", node))
		return
	}
	response.WriteEntity(m)
}

func (a *Api) getNodeMetrics(node string) *v1alpha1.NodeMetrics {
	batch := a.metricSink.GetLatestDataBatch()
	if batch == nil {
		return nil
	}

	ms, found := batch.MetricSets[core.NodeKey(node)]
	if !found {
		return nil
	}

	usage, err := parseResourceList(ms)
	if err != nil {
		return nil
	}

	return &v1alpha1.NodeMetrics{
		ObjectMeta: kube_v1.ObjectMeta{
			Name:              node,
			CreationTimestamp: metav1.NewTime(time.Now()),
		},
		Timestamp: metav1.NewTime(batch.Timestamp),
		Window:    metav1.Duration{Duration: time.Minute},
		Usage:     usage,
	}
}

func parseResourceList(ms *core.MetricSet) (kube_v1.ResourceList, error) {
	cpu, found := ms.MetricValues[core.MetricCpuUsageRate.MetricDescriptor.Name]
	if !found {
		return kube_v1.ResourceList{}, fmt.Errorf("cpu not found")
	}
	mem, found := ms.MetricValues[core.MetricMemoryWorkingSet.MetricDescriptor.Name]
	if !found {
		return kube_v1.ResourceList{}, fmt.Errorf("memory not found")
	}

	return kube_v1.ResourceList{
		kube_v1.ResourceCPU: *resource.NewMilliQuantity(
			cpu.IntValue,
			resource.DecimalSI),
		kube_v1.ResourceMemory: *resource.NewQuantity(
			mem.IntValue,
			resource.BinarySI),
	}, nil
}

func (a *Api) allPodMetricsList(request *restful.Request, response *restful.Response) {
	podMetricsInNamespaceList(a, request, response, kube_v1.NamespaceAll)
}

func (a *Api) podMetricsList(request *restful.Request, response *restful.Response) {
	podMetricsInNamespaceList(a, request, response, request.PathParameter("namespace-name"))
}

func podMetricsInNamespaceList(a *Api, request *restful.Request, response *restful.Response, namespace string) {
	selector := request.QueryParameter("labelSelector")

	labelSelector, err := labels.Parse(selector)
	if err != nil {
		errMsg := fmt.Errorf("Error while parsing selector %v: %v", selector, err)
		glog.Error(errMsg)
		response.WriteError(http.StatusBadRequest, errMsg)
		return
	}

	pods, err := a.podLister.Pods(namespace).List(labelSelector)
	if err != nil {
		errMsg := fmt.Errorf("Error while listing pods for selector %v: %v", selector, err)
		glog.Error(errMsg)
		response.WriteError(http.StatusInternalServerError, errMsg)
		return
	}

	res := v1alpha1.PodMetricsList{}
	for _, pod := range pods {
		if m := a.getPodMetrics(pod); m != nil {
			res.Items = append(res.Items, *m)
		} else {
			glog.Infof("No metrics for pod %s/%s", pod.Namespace, pod.Name)
		}
	}
	response.WriteEntity(&res)
}

func (a *Api) podMetrics(request *restful.Request, response *restful.Response) {
	ns := request.PathParameter("namespace-name")
	name := request.PathParameter("pod-name")

	pod, err := a.podLister.Pods(ns).Get(name)
	if err != nil {
		errMsg := fmt.Errorf("Error while getting pod %v: %v", name, err)
		glog.Error(errMsg)
		response.WriteError(http.StatusInternalServerError, errMsg)
		return
	}
	if pod == nil {
		response.WriteError(http.StatusNotFound, fmt.Errorf("Pod %v/%v not defined", ns, name))
		return
	}

	if m := a.getPodMetrics(pod); m != nil {
		response.WriteEntity(m)
	} else {
		response.WriteError(http.StatusNotFound, fmt.Errorf("No metrics availalble for pod %v/%v", ns, name))
	}
}

func (a *Api) getPodMetrics(pod *kube_v1.Pod) *v1alpha1.PodMetrics {
	batch := a.metricSink.GetLatestDataBatch()
	if batch == nil {
		return nil
	}

	res := &v1alpha1.PodMetrics{
		ObjectMeta: kube_v1.ObjectMeta{
			Name:              pod.Name,
			Namespace:         pod.Namespace,
			CreationTimestamp: metav1.NewTime(time.Now()),
		},
		Timestamp:  metav1.NewTime(batch.Timestamp),
		Window:     metav1.Duration{Duration: time.Minute},
		Containers: make([]v1alpha1.ContainerMetrics, 0),
	}

	for _, c := range pod.Spec.Containers {
		ms, found := batch.MetricSets[core.PodContainerKey(pod.Namespace, pod.Name, c.Name)]
		if !found {
			glog.Infof("No metrics for container %s in pod %s/%s", c.Name, pod.Namespace, pod.Name)
			return nil
		}

		usage, err := parseResourceList(ms)
		if err != nil {
			return nil
		}

		res.Containers = append(res.Containers, v1alpha1.ContainerMetrics{Name: c.Name, Usage: usage})
	}

	return res
}
