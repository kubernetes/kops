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

package watch

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// WatchDelay is the time between a Watch being dropped and attempting to resume it
const WatchDelay = 30 * time.Second

func NewDynamicWatch(config rest.Config) (*dynamicWatch, chan event.GenericEvent, error) {
	dw := &dynamicWatch{events: make(chan event.GenericEvent)}

	restMapper, err := apiutil.NewDiscoveryRESTMapper(&config)
	if err != nil {
		return nil, nil, err
	}

	client, err := dynamic.NewForConfig(&config)
	if err != nil {
		return nil, nil, err
	}

	dw.restMapper = restMapper
	dw.config = config
	dw.client = client
	return dw, dw.events, nil
}

type dynamicWatch struct {
	config     rest.Config
	client     dynamic.Interface
	restMapper meta.RESTMapper
	events     chan event.GenericEvent
}

func (dw *dynamicWatch) newDynamicClient(gvk schema.GroupVersionKind) (dynamic.ResourceInterface, error) {
	mapping, err := dw.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}
	return dw.client.Resource(mapping.Resource), nil
}

// Add registers a watch for changes to 'trigger' filtered by 'options' to raise an event on 'target'
func (dw *dynamicWatch) Add(trigger schema.GroupVersionKind, options metav1.ListOptions, target metav1.ObjectMeta) error {
	client, err := dw.newDynamicClient(trigger)
	if err != nil {
		return fmt.Errorf("creating client for (%s): %v", trigger.String(), err)
	}

	go func() {
		for {
			dw.watchUntilClosed(client, trigger, options, target)

			time.Sleep(WatchDelay)
		}
	}()

	return nil
}

// A Watch will be closed when the pod loses connection to the API server.
// If a Watch is opened with no ResourceVersion then we will recieve an 'ADDED'
// event for all Watch objects[1]. This will result in 'overnotification'
// from this Watch but it will ensure we always Reconcile when needed`.
//
// [1] https://github.com/kubernetes/kubernetes/issues/54878#issuecomment-357575276
func (dw *dynamicWatch) watchUntilClosed(client dynamic.ResourceInterface, trigger schema.GroupVersionKind, options metav1.ListOptions, target metav1.ObjectMeta) {
	log := log.Log

	events, err := client.Watch(options)

	if err != nil {
		log.WithValues("kind", trigger.String()).WithValues("namespace", target.Namespace).WithValues("labels", options.LabelSelector).Error(err, "adding watch to dynamic client")
		return
	}

	log.WithValues("kind", trigger.String()).WithValues("namespace", target.Namespace).WithValues("labels", options.LabelSelector).Info("watch began")

	for clientEvent := range events.ResultChan() {
		log.WithValues("type", clientEvent.Type).WithValues("kind", trigger.String()).Info("broadcasting event")
		dw.events <- event.GenericEvent{Object: clientEvent.Object, Meta: &target}
	}

	log.WithValues("kind", trigger.String()).WithValues("namespace", target.Namespace).WithValues("labels", options.LabelSelector).Info("watch closed")

	return
}
