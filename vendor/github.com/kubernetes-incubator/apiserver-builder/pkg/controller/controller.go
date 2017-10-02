/*
Copyright 2017 The Kubernetes Authors.

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

package controller

import (
	"time"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/apiserver-builder/pkg/builders"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
)

// QueueingEventHandler queues the key for the object on add and update events
type QueueingEventHandler struct {
	Queue         workqueue.RateLimitingInterface
	ObjToKey      func(obj interface{}) (string, error)
	EnqueueDelete bool
}

func (c *QueueingEventHandler) enqueue(obj interface{}) {
	fn := c.ObjToKey
	if c.ObjToKey == nil {
		fn = cache.DeletionHandlingMetaNamespaceKeyFunc
	}
	key, err := fn(obj)
	if err != nil {
		glog.Errorf("Couldn't get key for object %+v: %v", obj, err)
		return
	}
	c.Queue.Add(key)
}

func (c *QueueingEventHandler) OnAdd(obj interface{}) {
	glog.V(6).Infof("Add event for %+v\n", obj)
	c.enqueue(obj)
}

func (c *QueueingEventHandler) OnUpdate(oldObj, newObj interface{}) {
	glog.V(6).Infof("Update event for %+v\n", newObj)
	c.enqueue(newObj)
}

func (c *QueueingEventHandler) OnDelete(obj interface{}) {
	glog.V(6).Infof("Delete event for %+v\n", obj)
	if c.EnqueueDelete {
		c.enqueue(obj)
	}
}

// QueueWorker continuously runs a Reconcile function against a message Queue
type QueueWorker struct {
	Queue      workqueue.RateLimitingInterface
	MaxRetries int
	Name       string
	Reconcile  func(key string) error
}

// Run schedules a routine to continuously process Queue messages
// until shutdown is closed
func (q *QueueWorker) Run(shutdown <-chan struct{}) {
	defer runtime.HandleCrash()

	// Every second, process all messages in the Queue until it is time to shutdown
	go wait.Until(q.ProcessAllMessages, time.Second, shutdown)

	go func() {
		<-shutdown

		// Stop accepting messages into the Queue
		glog.V(1).Infof("Shutting down %s Queue\n", q.Name)
		q.Queue.ShutDown()
	}()
}

// ProcessAllMessages tries to process all messages in the Queue
func (q *QueueWorker) ProcessAllMessages() {
	for done := false; !done; {
		// Process all messages in the Queue
		done = q.ProcessMessage()
	}
}

// ProcessMessage tries to process the next message in the Queue, and requeues on an error
func (q *QueueWorker) ProcessMessage() bool {
	key, quit := q.Queue.Get()
	if quit {
		// Queue is empty
		return true
	}
	defer q.Queue.Done(key)

	// Do the work
	err := q.Reconcile(key.(string))
	if err == nil {
		// Success.  Clear the requeue count for this key.
		q.Queue.Forget(key)
		return false
	}

	// Error.  Maybe retry if haven't hit the limit.
	if q.Queue.NumRequeues(key) < q.MaxRetries {
		glog.V(4).Infof("Error handling %s Queue message %v: %v", q.Name, key, err)
		q.Queue.AddRateLimited(key)
		return false
	}

	glog.V(4).Infof("Too many retries for %s Queue message %v: %v", q.Name, key, err)
	q.Queue.Forget(key)
	return false
}

func GetDefaults(c interface{}) DefaultMethods {
	i, ok := c.(DefaultMethods)
	if !ok {
		return &builders.DefaultControllerFns{}
	}
	return i
}

type DefaultMethods interface {
	Run(stopCh <-chan struct{})
}

type Controller interface {
	Run(stopCh <-chan struct{})
	GetName() string
}

// StartControllerManager
func StartControllerManager(controllers ...Controller) <-chan struct{} {
	shutdown := make(chan struct{})
	for _, c := range controllers {
		c.Run(shutdown)
	}
	return shutdown
}

func GetConfig(kubeconfig string) (*rest.Config, error) {
	if len(kubeconfig) > 0 {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		return rest.InClusterConfig()
	}
}
