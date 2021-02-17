/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package collector

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/pager"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/validation"
	"k8s.io/kops/upup/pkg/fi"
)

type validateCollector struct {
	cluster                *kops.Cluster
	cloud                  fi.Cloud
	client                 simple.Clientset
	k8sClient              *kubernetes.Clientset
	readyNodeDesc          *prometheus.Desc
	notReadyNodeDesc       *prometheus.Desc
	healthyComponentDesc   *prometheus.Desc
	unhealthyComponentDesc *prometheus.Desc
	pendingPodDesc         *prometheus.Desc
	unknownPodDesc         *prometheus.Desc
	notReadyPodDesc        *prometheus.Desc
	missingPodDesc         *prometheus.Desc
}

func init() {
	registerCollector("validate", NewNodeCollector)
}

func NewNodeCollector(cluster *kops.Cluster, cloud fi.Cloud, client simple.Clientset, k8sClient *kubernetes.Clientset) (Collector, error) {
	return &validateCollector{
		cluster:   cluster,
		cloud:     cloud,
		client:    client,
		k8sClient: k8sClient,
		readyNodeDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "node", "ready"),
			"Ready nodes in the kubernetes cluster.",
			[]string{"cluster_name"},
			nil,
		),
		notReadyNodeDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "node", "notReady"),
			"Not ready nodes in the kubernetes cluster.",
			[]string{"cluster_name"},
			nil,
		),
		healthyComponentDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "component", "healthy"),
			"Healthy components of the kubernetes cluster.",
			[]string{"cluster_name"},
			nil,
		),
		unhealthyComponentDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "component", "unhealthy"),
			"Unhealthy components of the kubernetes cluster.",
			[]string{"cluster_name"},
			nil,
		),
		pendingPodDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pod", "pending"),
			"Pending pods in the kubernetes cluster.",
			[]string{"cluster_name"},
			nil,
		),
		unknownPodDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pod", "unknown"),
			"Unknown pods in the kubernetes cluster.",
			[]string{"cluster_name"},
			nil,
		),
		notReadyPodDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pod", "notReady"),
			"Not ready pods in the kubernetes cluster.",
			[]string{"cluster_name"},
			nil,
		),
		missingPodDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pod", "missing"),
			"Missing pods in the kubernetes cluster.",
			[]string{"cluster_name"},
			nil,
		),
	}, nil
}

func (c *validateCollector) Update(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	nodes, allNodes, err := c.readyNodes(ctx)
	if err != nil {
		klog.Errorf("failed to validate node: %v", err)
	}
	notReadyNodes := len(allNodes) - len(nodes)

	healthy, unhealthy, err := c.components(ctx)
	if err != nil {
		klog.Errorf("failed to validate components: %v", err)
	}

	pending, unknown, notReady, missing, err := c.pods(ctx, nodes)
	if err != nil {
		klog.Errorf("failed to validate pods: %v", err)
	}

	ch <- prometheus.MustNewConstMetric(c.readyNodeDesc, prometheus.GaugeValue, float64(len(nodes)), c.cluster.GetName())
	ch <- prometheus.MustNewConstMetric(c.notReadyNodeDesc, prometheus.GaugeValue, float64(notReadyNodes), c.cluster.GetName())
	ch <- prometheus.MustNewConstMetric(c.healthyComponentDesc, prometheus.GaugeValue, float64(len(healthy)), c.cluster.GetName())
	ch <- prometheus.MustNewConstMetric(c.unhealthyComponentDesc, prometheus.GaugeValue, float64(len(unhealthy)), c.cluster.GetName())
	ch <- prometheus.MustNewConstMetric(c.pendingPodDesc, prometheus.GaugeValue, float64(len(pending)), c.cluster.GetName())
	ch <- prometheus.MustNewConstMetric(c.unknownPodDesc, prometheus.GaugeValue, float64(len(unknown)), c.cluster.GetName())
	ch <- prometheus.MustNewConstMetric(c.notReadyPodDesc, prometheus.GaugeValue, float64(len(notReady)), c.cluster.GetName())
	ch <- prometheus.MustNewConstMetric(c.missingPodDesc, prometheus.GaugeValue, float64(len(missing)), c.cluster.GetName())
	return nil
}

func (c *validateCollector) readyNodes(ctx context.Context) ([]corev1.Node, []*cloudinstances.CloudInstance, error) {
	nodeList, err := c.k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}

	list, err := c.client.InstanceGroupsFor(c.cluster).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}

	var instanceGroups []*kops.InstanceGroup
	for i := range list.Items {
		ig := &list.Items[i]
		instanceGroups = append(instanceGroups, ig)
	}

	warnUnmatched := false
	cloudGroups, err := c.cloud.GetCloudGroups(c.cluster, instanceGroups, warnUnmatched, nodeList.Items)
	if err != nil {
		return nil, nil, err
	}
	validation := &validation.ValidationCluster{}
	ready, _ := validation.ValidateNodes(cloudGroups, instanceGroups)

	var allMember []*cloudinstances.CloudInstance
	for _, instanceGroup := range cloudGroups {
		allMember = append(allMember, instanceGroup.All...)
	}
	return ready, allMember, nil
}

func (c *validateCollector) components(ctx context.Context) ([]corev1.ComponentStatus, []corev1.ComponentStatus, error) {
	healthyComponents := []corev1.ComponentStatus{}
	unhealthyComponents := []corev1.ComponentStatus{}

	// TODO: ComponentStatuses API has been deprecated since 1.19, so we have to rewrite this API call.
	// https://github.com/kubernetes/kops/issues/10057
	componentList, err := c.k8sClient.CoreV1().ComponentStatuses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return healthyComponents, unhealthyComponents, err
	}

	for _, component := range componentList.Items {
		con := true
		for _, condition := range component.Conditions {
			if condition.Status != corev1.ConditionTrue {
				con = false
			}
		}
		if con {
			healthyComponents = append(healthyComponents, component)
		} else {
			unhealthyComponents = append(unhealthyComponents, component)
		}
	}
	return healthyComponents, unhealthyComponents, nil
}

var masterStaticPods = []string{
	"kube-apiserver",
	"kube-controller-manager",
	"kube-scheduler",
}

func (c *validateCollector) pods(ctx context.Context, nodes []corev1.Node) ([]corev1.Pod, []corev1.Pod, []corev1.Pod, map[string]string, error) {
	masterWithoutPod := map[string]map[string]bool{}
	nodeByAddress := map[string]string{}
	for _, node := range nodes {
		labels := node.GetLabels()
		if labels != nil && labels["kubernetes.io/role"] == "master" {
			masterWithoutPod[node.Name] = map[string]bool{}
			for _, pod := range masterStaticPods {
				masterWithoutPod[node.Name][pod] = true
			}
		}
		for _, nodeAddress := range node.Status.Addresses {
			nodeByAddress[nodeAddress.Address] = node.Name
		}
	}

	pendingPods := []corev1.Pod{}
	unknownPods := []corev1.Pod{}
	notReadyPods := []corev1.Pod{}
	err := pager.New(pager.SimplePageFunc(func(opts metav1.ListOptions) (runtime.Object, error) {
		return c.k8sClient.CoreV1().Pods(metav1.NamespaceAll).List(ctx, opts)
	})).EachListItem(context.TODO(), metav1.ListOptions{}, func(obj runtime.Object) error {
		pod := obj.(*corev1.Pod)

		app := pod.GetLabels()["k8s-app"]
		if pod.Namespace == "kube-system" && masterWithoutPod[nodeByAddress[pod.Status.HostIP]][app] {
			delete(masterWithoutPod[nodeByAddress[pod.Status.HostIP]], app)
		}

		priority := pod.Spec.PriorityClassName
		if priority != "system-cluster-critical" && priority != "system-node-critical" {
			return nil
		}
		if pod.Status.Phase == corev1.PodSucceeded {
			return nil
		}
		if pod.Status.Phase == corev1.PodPending {
			pendingPods = append(pendingPods, *pod)
			return nil
		}
		if pod.Status.Phase == corev1.PodUnknown {
			unknownPods = append(unknownPods, *pod)
			return nil
		}
		for _, container := range pod.Status.ContainerStatuses {
			if !container.Ready {
				notReadyPods = append(notReadyPods, *pod)
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error listing Pods: %v", err)
	}

	missingPod := map[string]string{}
	for node, nodeMap := range masterWithoutPod {
		for app := range nodeMap {
			missingPod[app] = node
		}
	}

	return pendingPods, unknownPods, notReadyPods, missingPod, nil
}
