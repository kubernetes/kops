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
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/upup/pkg/fi"
)

const (
	namespace = "kops"
)

var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
		"kops_exporter: Duration of a collector scrape.",
		[]string{"collector"},
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_success"),
		"kops_exporter: Whether a collector succeeded.",
		[]string{"collector"},
		nil,
	)
	factories      = make(map[string]func(cluster *kops.Cluster, cloud fi.Cloud, client simple.Clientset, k8sClient *kubernetes.Clientset) (Collector, error))
	collectorState = make(map[string]bool)
)

type KopsCollector struct {
	Collectors map[string]Collector
}

type Collector interface {
	Update(ch chan<- prometheus.Metric) error
}

func registerCollector(collector string, factory func(cluster *kops.Cluster, cloud fi.Cloud, client simple.Clientset, k8sClient *kubernetes.Clientset) (Collector, error)) {
	collectorState[collector] = true
	factories[collector] = factory
}

func NewCollector(cluster *kops.Cluster, cloud fi.Cloud, client simple.Clientset, k8sClient *kubernetes.Clientset) (*KopsCollector, error) {
	collectors := make(map[string]Collector)
	for key, enabled := range collectorState {
		if enabled {
			collector, err := factories[key](cluster, cloud, client, k8sClient)
			if err != nil {
				return nil, err
			}
			collectors[key] = collector
		}
	}
	return &KopsCollector{Collectors: collectors}, nil
}

func (kc KopsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
}

func (kc KopsCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(kc.Collectors))
	for name, c := range kc.Collectors {
		go func(name string, c Collector) {
			execute(name, c, ch)
			wg.Done()
		}(name, c)
	}
	wg.Wait()
}

func execute(name string, c Collector, ch chan<- prometheus.Metric) {
	begin := time.Now()
	err := c.Update(ch)
	duration := time.Since(begin)
	var success float64

	if err != nil {
		klog.Errorf("%s collector failed after %fs: %s", name, duration.Seconds(), err.Error())
		success = 0
	}
	success = 1
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name)
}
