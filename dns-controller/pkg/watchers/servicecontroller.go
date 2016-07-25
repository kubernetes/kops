package watchers

import (
	"fmt"
	"time"

	"github.com/golang/glog"

	"k8s.io/kops/dns-controller/pkg/dns"
	"k8s.io/kops/dns-controller/pkg/util"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/v1"
	client "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_3/typed/core/v1"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/watch"
	"strings"
)

// ServiceController watches for services with dns annotations
type ServiceController struct {
	util.Stoppable
	kubeClient *client.CoreClient
	scope        dns.Scope
}

// newServiceController creates a serviceController
func NewServiceController(kubeClient *client.CoreClient, dns dns.Context) (*ServiceController, error) {
	scope, err := dns.CreateScope("service")
	if err != nil {
		return nil, fmt.Errorf("error building dns scope: %v", err)
	}
	c := &ServiceController{
		kubeClient: kubeClient,
		scope:        scope,
	}

	return c, nil
}

// Run starts the ServiceController.
func (c *ServiceController) Run() {
	glog.Infof("starting service controller")

	stopCh := c.StopChannel()
	go c.runWatcher(stopCh)

	<-stopCh
	glog.Infof("shutting down service controller")
}

func (c *ServiceController) runWatcher(stopCh <-chan struct{}) {
	runOnce := func() (bool, error) {
		var listOpts api.ListOptions
		glog.Warningf("querying without label filter")
		listOpts.LabelSelector = labels.Everything()
		glog.Warningf("querying without field filter")
		listOpts.FieldSelector = fields.Everything()
		serviceList, err := c.kubeClient.Services("").List(listOpts)
		if err != nil {
			return false, fmt.Errorf("error listing services: %v", err)
		}
		for i := range serviceList.Items {
			service := &serviceList.Items[i]
			glog.V(4).Infof("found service: %v", service.Name)
			c.updateServiceRecords(service)
		}
		c.scope.MarkReady()

		glog.Warningf("querying without label filter")
		listOpts.LabelSelector = labels.Everything()
		glog.Warningf("querying without field filter")
		listOpts.FieldSelector = fields.Everything()
		listOpts.Watch = true
		listOpts.ResourceVersion = serviceList.ResourceVersion
		watcher, err := c.kubeClient.Services("").Watch(listOpts)
		if err != nil {
			return false, fmt.Errorf("error watching services: %v", err)
		}
		ch := watcher.ResultChan()
		for {
			select {
			case <-stopCh:
				glog.Infof("Got stop signal")
				return true, nil
			case event, ok := <-ch:
				if !ok {
					glog.Infof("service watch channel closed")
					return false, nil
				}

				service := event.Object.(*v1.Service)
				glog.V(4).Infof("service changed: %s %v", event.Type, service.Name)

				switch event.Type {
				case watch.Added, watch.Modified:
					c.updateServiceRecords(service)

				case watch.Deleted:
					c.scope.Replace(service.Name, nil)

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

func (c *ServiceController) updateServiceRecords(service *v1.Service) {
	var records []dns.Record

	specExternal := service.Annotations[AnnotationNameDnsExternal]
	if specExternal != "" {
		var ingresses []dns.Record
		for i := range service.Status.LoadBalancer.Ingress {
			ingress := &service.Status.LoadBalancer.Ingress[i]
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

		tokens := strings.Split(specExternal, ",")
		for _, token := range tokens {
			token = strings.TrimSpace(token)

			fqdn := dns.EnsureDotSuffix(token)
			for _, ingress := range ingresses {
				var r dns.Record
				r = ingress
				r.FQDN = fqdn
				records = append(records, r)
			}
		}
	} else {
		glog.V(4).Infof("Service %q did not have %s annotation", service.Name, AnnotationNameDnsInternal)
	}

	c.scope.Replace( service.Name, records)
}
