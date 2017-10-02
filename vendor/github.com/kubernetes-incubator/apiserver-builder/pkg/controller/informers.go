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
	"os"
	"path/filepath"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
)

type SharedInformersDefaults struct {
	KubernetesFactory   informers.SharedInformerFactory
	KubernetesClientSet *kubernetes.Clientset

	// Extensions allows a controller-manager to define new data structures
	// shared by all of its controllers.
	// Set this by overriding the InitExtensions function on the generated *SharedInformers
	// type under the consuming projects pkg/controller/sharedinformers package
	// by in a new informers.go file
	Extensions interface{}

	WorkerQueues map[string]*QueueWorker
}

// InitKubernetesInformers initializes the Kubernetes clientset and informerfactory
// informers must still be started by overriding StartAdditionalInformers
func (si *SharedInformersDefaults) InitKubernetesInformers(config *rest.Config) {
	si.KubernetesClientSet = kubernetes.NewForConfigOrDie(config)
	si.KubernetesFactory = informers.NewSharedInformerFactory(si.KubernetesClientSet, 10*time.Minute)
}

// Init is called before the informers are started, and can be used to perform any additional
// initialization shared by multiple controllers
func (*SharedInformersDefaults) Init() {}

// StartAdditionalInformers is called to start informers for resources not defined
// in the extension apiserver.  Override this and use it to start informers for
// Kubernetes resources such as Pods
func (*SharedInformersDefaults) StartAdditionalInformers(shutdown <-chan struct{}) {}

// SetupKubernetesTypes can be overridden to initialize the Kubernetes clientset and informers
func (*SharedInformersDefaults) SetupKubernetesTypes() bool {
	return false
}

func (c *SharedInformersDefaults) Watch(
	name string, i cache.SharedIndexInformer,
	f func(interface{}) (string, error), r func(string) error) {
	q := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), name)
	queue := &QueueWorker{q, 10, name, r}
	if c.WorkerQueues == nil {
		c.WorkerQueues = map[string]*QueueWorker{}
	}
	c.WorkerQueues[name] = queue
	i.AddEventHandler(&QueueingEventHandler{q, f, true})
}

func NewConfig(kubeconfig string) (*rest.Config, error) {
	if len(kubeconfig) != 0 {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	// Try getting an in cluster config if present
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, err
	}

	// No in cluster config, look for the config in the home directory
	return clientcmd.BuildConfigFromFlags("", getHomeConfigPath())
}

// getHomeConfigPath returns the path to the kubeconfig in a user's home directory
func getHomeConfigPath() string {
	home := os.Getenv("HOME")
	if len(home) == 0 {
		home = os.Getenv("USERPROFILE") // windows
	}

	if len(home) == 0 {
		return ""
	}
	return filepath.Join(home, ".kube", "config")
}
