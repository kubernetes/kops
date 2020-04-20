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
	"strings"
	"time"

	"k8s.io/kops/dns-controller/pkg/dns"
	"k8s.io/kops/dns-controller/pkg/util"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

// ServiceController watches for services with dns annotations
type ServiceController struct {
	util.Stoppable
	client    kubernetes.Interface
	namespace string
	scope     dns.Scope
}

// NewServiceController creates a ServiceController
func NewServiceController(client kubernetes.Interface, dns dns.Context, namespace string) (*ServiceController, error) {
	scope, err := dns.CreateScope("service")
	if err != nil {
		return nil, fmt.Errorf("error building dns scope: %v", err)
	}
	c := &ServiceController{
		client:    client,
		namespace: namespace,
		scope:     scope,
	}

	return c, nil
}

// Run starts the ServiceController.
func (c *ServiceController) Run() {
	klog.Infof("starting service controller")

	stopCh := c.StopChannel()
	go c.runWatcher(stopCh)

	<-stopCh
	klog.Infof("shutting down service controller")
}

func (c *ServiceController) runWatcher(stopCh <-chan struct{}) {
	runOnce := func() (bool, error) {
		ctx := context.TODO()

		var listOpts metav1.ListOptions
		klog.V(4).Infof("querying without label filter")

		allKeys := c.scope.AllKeys()
		serviceList, err := c.client.CoreV1().Services(c.namespace).List(ctx, listOpts)
		if err != nil {
			return false, fmt.Errorf("error listing services: %v", err)
		}
		foundKeys := make(map[string]bool)
		for i := range serviceList.Items {
			service := &serviceList.Items[i]
			klog.V(4).Infof("found service: %v", service.Name)
			key := c.updateServiceRecords(service)
			foundKeys[key] = true
		}
		for _, key := range allKeys {
			if !foundKeys[key] {
				// The service previously existed, but no longer exists; delete it from the scope
				klog.V(2).Infof("removing service not found in list: %s", key)
				c.scope.Replace(key, nil)
			}
		}
		c.scope.MarkReady()

		listOpts.Watch = true
		listOpts.ResourceVersion = serviceList.ResourceVersion
		watcher, err := c.client.CoreV1().Services(c.namespace).Watch(ctx, listOpts)
		if err != nil {
			return false, fmt.Errorf("error watching services: %v", err)
		}
		ch := watcher.ResultChan()
		for {
			select {
			case <-stopCh:
				klog.Infof("Got stop signal")
				return true, nil
			case event, ok := <-ch:
				if !ok {
					klog.Infof("service watch channel closed")
					return false, nil
				}

				service := event.Object.(*v1.Service)
				klog.V(4).Infof("service changed: %s %v", event.Type, service.Name)

				switch event.Type {
				case watch.Added, watch.Modified:
					c.updateServiceRecords(service)

				case watch.Deleted:
					c.scope.Replace(service.Namespace+"/"+service.Name, nil)

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

// updateServiceRecords will apply the records for the specified service.
// It returns the key that was set (or "" if no key was set)
func (c *ServiceController) updateServiceRecords(service *v1.Service) string {
	var records []dns.Record

	specExternal := service.Annotations[AnnotationNameDNSExternal]
	specInternal := service.Annotations[AnnotationNameDNSInternal]
	if len(specExternal) != 0 || len(specInternal) != 0 {
		var ingresses []dns.Record

		if service.Spec.Type == v1.ServiceTypeLoadBalancer {
			for i := range service.Status.LoadBalancer.Ingress {
				ingress := &service.Status.LoadBalancer.Ingress[i]
				if ingress.Hostname != "" {
					// TODO: Support ELB aliases
					ingresses = append(ingresses, dns.Record{
						RecordType: dns.RecordTypeCNAME,
						Value:      ingress.Hostname,
					})
					klog.V(4).Infof("Found CNAME record for service %s/%s: %q", service.Namespace, service.Name, ingress.Hostname)
				}
				if ingress.IP != "" {
					ingresses = append(ingresses, dns.Record{
						RecordType: dns.RecordTypeA,
						Value:      ingress.IP,
					})
					klog.V(4).Infof("Found A record for service %s/%s: %q", service.Namespace, service.Name, ingress.IP)
				}
			}
		} else if service.Spec.Type == v1.ServiceTypeNodePort {
			var roleType string
			if len(specExternal) != 0 && len(specInternal) != 0 {
				klog.Warningln("DNS Records not possible for both Internal and Externals IPs.")
				return ""
			} else if len(specInternal) != 0 {
				roleType = dns.RoleTypeInternal
			} else {
				roleType = dns.RoleTypeExternal
			}
			ingresses = append(ingresses, dns.Record{
				RecordType: dns.RecordTypeAlias,
				Value:      dns.AliasForNodesInRole("node", roleType),
			})
			klog.V(4).Infof("Setting internal alias for NodePort service %s/%s", service.Namespace, service.Name)
		} else {
			// TODO: Emit event so that users are informed of this
			klog.V(2).Infof("Cannot expose service %s/%s of type %q", service.Namespace, service.Name, service.Spec.Type)
		}

		var tokens []string

		if len(specExternal) != 0 {
			tokens = append(tokens, strings.Split(specExternal, ",")...)
		}

		if len(specInternal) != 0 {
			tokens = append(tokens, strings.Split(specInternal, ",")...)
		}

		for _, token := range tokens {
			token = strings.TrimSpace(token)

			fqdn := dns.EnsureDotSuffix(token)
			for _, ingress := range ingresses {
				r := ingress
				r.FQDN = fqdn
				records = append(records, r)
			}
		}
	} else {
		klog.V(8).Infof("Service %s/%s did not have %s annotation", service.Namespace, service.Name, AnnotationNameDNSExternal)
	}

	key := service.Namespace + "/" + service.Name
	c.scope.Replace(key, records)
	return key
}
