/*
Copyright 2019 The Kubernetes Authors.

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
	"context"
	"fmt"
	"time"

	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"k8s.io/kops/dns-controller/pkg/dns"
	"k8s.io/kops/dns-controller/pkg/util"
)

// IngressController watches for Ingress objects with dns labels
type IngressController struct {
	util.Stoppable
	client    kubernetes.Interface
	namespace string
	scope     dns.Scope
}

// NewIngressController creates a IngressController
func NewIngressController(client kubernetes.Interface, dns dns.Context, namespace string) (*IngressController, error) {
	scope, err := dns.CreateScope("ingress")
	if err != nil {
		return nil, fmt.Errorf("error building dns scope: %v", err)
	}
	c := &IngressController{
		client:    client,
		namespace: namespace,
		scope:     scope,
	}

	return c, nil
}

// Run starts the IngressController.
func (c *IngressController) Run() {
	klog.Infof("starting ingress controller")

	stopCh := c.StopChannel()
	go c.runWatcher(stopCh)

	<-stopCh
	klog.Infof("shutting down ingress controller")
}

func (c *IngressController) runWatcher(stopCh <-chan struct{}) {
	runOnce := func() (bool, error) {
		ctx := context.TODO()

		var listOpts metav1.ListOptions
		klog.V(4).Infof("querying without label filter")

		allKeys := c.scope.AllKeys()
		ingressList, err := c.client.ExtensionsV1beta1().Ingresses(c.namespace).List(ctx, listOpts)
		if err != nil {
			return false, fmt.Errorf("error listing ingresses: %v", err)
		}
		foundKeys := make(map[string]bool)
		for i := range ingressList.Items {
			ingress := &ingressList.Items[i]
			klog.V(4).Infof("found ingress: %v", ingress.Name)
			key := c.updateIngressRecords(ingress)
			foundKeys[key] = true
		}
		for _, key := range allKeys {
			if !foundKeys[key] {
				// The ingress previously existed, but no longer exists; delete it from the scope
				klog.V(2).Infof("removing ingress not found in list: %s", key)
				c.scope.Replace(key, nil)
			}
		}
		c.scope.MarkReady()

		listOpts.Watch = true
		listOpts.ResourceVersion = ingressList.ResourceVersion
		watcher, err := c.client.ExtensionsV1beta1().Ingresses(c.namespace).Watch(ctx, listOpts)
		if err != nil {
			return false, fmt.Errorf("error watching ingresses: %v", err)
		}
		ch := watcher.ResultChan()
		for {
			select {
			case <-stopCh:
				klog.Infof("Got stop signal")
				return true, nil
			case event, ok := <-ch:
				if !ok {
					klog.Infof("ingress watch channel closed")
					return false, nil
				}

				ingress := event.Object.(*v1beta1.Ingress)
				klog.V(4).Infof("ingress changed: %s %v", event.Type, ingress.Name)

				switch event.Type {
				case watch.Added, watch.Modified:
					c.updateIngressRecords(ingress)

				case watch.Deleted:
					c.scope.Replace(ingress.Namespace+"/"+ingress.Name, nil)

				default:
					klog.Warningf("Unknown event type: %v", event.Type)
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
			klog.Warningf("Unexpected error in event watch, will retry: %v", err)
			time.Sleep(10 * time.Second)
		}
	}
}

// updateIngressRecords will apply the records for the specified ingress.  It returns the key that was set.
func (c *IngressController) updateIngressRecords(ingress *v1beta1.Ingress) string {
	var records []dns.Record

	var ingresses []dns.Record
	for i := range ingress.Status.LoadBalancer.Ingress {
		ingress := &ingress.Status.LoadBalancer.Ingress[i]
		if ingress.Hostname != "" {
			// TODO: Support ELB aliases
			ingresses = append(ingresses, dns.Record{
				RecordType: dns.RecordTypeCNAME,
				Value:      ingress.Hostname,
			})
		}
		if ingress.IP != "" {
			ingresses = append(ingresses, dns.Record{
				RecordType: dns.RecordTypeA,
				Value:      ingress.IP,
			})
		}
	}

	for _, rule := range ingress.Spec.Rules {
		if rule.Host == "" {
			continue
		}

		fqdn := dns.EnsureDotSuffix(rule.Host)
		for _, ingress := range ingresses {
			r := ingress
			r.FQDN = fqdn
			records = append(records, r)
		}
	}

	key := ingress.Namespace + "/" + ingress.Name
	c.scope.Replace(key, records)
	return key
}
