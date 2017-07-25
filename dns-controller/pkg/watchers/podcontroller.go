/*
Copyright 2016 The Kubernetes Authors.

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

package watchers

import (
	"fmt"
	"time"

	"github.com/golang/glog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/kops/dns-controller/pkg/dns"
	"k8s.io/kops/dns-controller/pkg/util"
	"strings"
)

// PodController watches for Pods with dns annotations
type PodController struct {
	util.Stoppable
	kubeClient kubernetes.Interface
	scope      dns.Scope
}

// newPodController creates a podController
func NewPodController(kubeClient kubernetes.Interface, dns dns.Context) (*PodController, error) {
	scope, err := dns.CreateScope("pod")
	if err != nil {
		return nil, fmt.Errorf("error building dns scope: %v", err)
	}
	c := &PodController{
		kubeClient: kubeClient,
		scope:      scope,
	}

	return c, nil
}

// Run starts the PodController.
func (c *PodController) Run() {
	glog.Infof("starting pod controller")

	stopCh := c.StopChannel()
	go c.runWatcher(stopCh)

	<-stopCh
	glog.Infof("shutting down pod controller")
}

func (c *PodController) runWatcher(stopCh <-chan struct{}) {
	runOnce := func() (bool, error) {
		var listOpts metav1.ListOptions
		glog.V(4).Infof("querying without label filter")
		podList, err := c.kubeClient.CoreV1().Pods("").List(listOpts)
		if err != nil {
			return false, fmt.Errorf("error listing pods: %v", err)
		}
		for i := range podList.Items {
			pod := &podList.Items[i]
			glog.V(4).Infof("found pod: %v", pod.Name)
			c.updatePodRecords(pod)
		}
		c.scope.MarkReady()

		listOpts.Watch = true
		listOpts.ResourceVersion = podList.ResourceVersion
		watcher, err := c.kubeClient.CoreV1().Pods("").Watch(listOpts)
		if err != nil {
			return false, fmt.Errorf("error watching pods: %v", err)
		}
		ch := watcher.ResultChan()
		for {
			select {
			case <-stopCh:
				glog.Infof("Got stop signal")
				return true, nil
			case event, ok := <-ch:
				if !ok {
					glog.Infof("pod watch channel closed")
					return false, nil
				}

				pod := event.Object.(*v1.Pod)
				glog.V(4).Infof("pod changed: %s %v", event.Type, pod.Name)

				switch event.Type {
				case watch.Added, watch.Modified:
					c.updatePodRecords(pod)

				case watch.Deleted:
					c.scope.Replace(pod.Namespace+"/"+pod.Name, nil)

				default:
					glog.Warningf("Unknown event type: %v", event.Type)
				}
			}
		}
	}

	for {
		stop, err := runOnce()
		if stop {
			return
		}

		if err != nil {
			glog.Warningf("Unexpected error in event watch, will retry: %v", err)
			time.Sleep(10 * time.Second)
		}
	}
}

func (c *PodController) updatePodRecords(pod *v1.Pod) {
	var records []dns.Record

	specExternal := pod.Annotations[AnnotationNameDnsExternal]
	if specExternal != "" {
		var aliases []string
		if pod.Spec.HostNetwork {
			if pod.Spec.NodeName != "" {
				aliases = append(aliases, "node/"+pod.Spec.NodeName+"/external")
			}
		} else {
			glog.V(4).Infof("Pod %q had %s=%s, but was not HostNetwork", pod.Name, AnnotationNameDnsExternal, specExternal)
		}

		tokens := strings.Split(specExternal, ",")
		for _, token := range tokens {
			token = strings.TrimSpace(token)

			fqdn := dns.EnsureDotSuffix(token)
			for _, alias := range aliases {
				records = append(records, dns.Record{
					RecordType: dns.RecordTypeAlias,
					FQDN:       fqdn,
					Value:      alias,
				})
			}
		}
	} else {
		glog.V(4).Infof("Pod %q did not have %s annotation", pod.Name, AnnotationNameDnsExternal)
	}

	specInternal := pod.Annotations[AnnotationNameDnsInternal]
	if specInternal != "" {
		var ips []string
		if pod.Spec.HostNetwork {
			if pod.Status.PodIP != "" {
				ips = append(ips, pod.Status.PodIP)
			}
		} else {
			glog.V(4).Infof("Pod %q had %s=%s, but was not HostNetwork", pod.Name, AnnotationNameDnsInternal, specInternal)
		}

		tokens := strings.Split(specInternal, ",")
		for _, token := range tokens {
			token = strings.TrimSpace(token)

			fqdn := dns.EnsureDotSuffix(token)
			for _, ip := range ips {
				records = append(records, dns.Record{
					RecordType: dns.RecordTypeA,
					FQDN:       fqdn,
					Value:      ip,
				})
			}
		}
	} else {
		glog.V(4).Infof("Pod %q did not have %s label", pod.Name, AnnotationNameDnsInternal)
	}

	c.scope.Replace(pod.Namespace+"/"+pod.Name, records)
}
