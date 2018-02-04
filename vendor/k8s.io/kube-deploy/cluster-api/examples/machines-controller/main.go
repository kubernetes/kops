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

package main

import (
	"context"
	"flag"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
)

// Stand-alone demo Machines controller. Right now, it just prints simple
// information whenever it sees a create/update/delete event for any Machine.
// It's meant to serve as an example / building block for implementing real
// controllers against the Machines API.

var kubeconfig = flag.String("kubeconfig", "", "path to kubeconfig file")

func main() {
	flag.Parse()

	client, _, err := restClient()
	if err != nil {
		panic(err.Error())
	}

	run(context.Background(), client)
}

func onAdd(obj interface{}) {
	machine := obj.(*clusterv1.Machine)
	fmt.Printf("object created: %s\n", machine.ObjectMeta.Name)
}

func onUpdate(oldObj, newObj interface{}) {
	oldMachine := oldObj.(*clusterv1.Machine)
	newMachine := newObj.(*clusterv1.Machine)
	fmt.Printf("object updated: %s\n", oldMachine.ObjectMeta.Name)
	fmt.Printf("  old k8s version: %s, new: %s\n", oldMachine.Spec.Versions.Kubelet, newMachine.Spec.Versions.Kubelet)
}

func onDelete(obj interface{}) {
	machine := obj.(*clusterv1.Machine)
	fmt.Printf("object deleted: %s\n", machine.ObjectMeta.Name)
}

func run(ctx context.Context, client *rest.RESTClient) error {
	source := cache.NewListWatchFromClient(client, "machines", apiv1.NamespaceAll, fields.Everything())

	_, controller := cache.NewInformer(
		source,
		&clusterv1.Machine{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    onAdd,
			UpdateFunc: onUpdate,
			DeleteFunc: onDelete,
		},
	)

	controller.Run(ctx.Done())
	// unreachable; run forever
	return nil
}
