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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
	"k8s.io/kube-deploy/cluster-api/client"
)

func restClient(kubeconfigpath string) (*rest.RESTClient, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigpath)
	if err != nil {
		return nil, err
	}

	if err := clusterv1.AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}

	config := *cfg
	config.GroupVersion = &clusterv1.SchemeGroupVersion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme.Scheme)}

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func kubeClientSet(kubeconfigpath string) (*kubernetes.Clientset, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigpath)
	if err != nil {
		return nil, err
	}

	if err := clusterv1.AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}

	config := *cfg
	config.GroupVersion = &clusterv1.SchemeGroupVersion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme.Scheme)}

	client, err := kubernetes.NewForConfig(&config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func nodeClient(kubeconfig string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func machineClient(kubeconfig string) (client.MachinesInterface, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	client, err := client.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return client.Machines(), nil
}

func host(kubeconfigpath string) (string, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigpath)
	if err != nil {
		return "", err
	}

	config := *cfg
	return config.Host, nil
}
