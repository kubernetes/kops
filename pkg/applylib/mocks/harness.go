/*
Copyright 2022 The Kubernetes Authors.

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

package mocks

import (
	"bytes"
	"context"
	"flag"
	"io"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	yamlserializer "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kops/pkg/applylib/mocks/mockkubeapiserver"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var kubeconfig = flag.String("kubeconfig", "", "set to use a real kube-apiserver")

type Harness struct {
	*testing.T

	k8s        *mockkubeapiserver.MockKubeAPIServer
	restConfig *rest.Config
	restMapper *restmapper.DeferredDiscoveryRESTMapper

	Scheme *runtime.Scheme
	Ctx    context.Context
	client client.Client
}

func NewHarness(t *testing.T) *Harness {
	h := &Harness{
		T:      t,
		Scheme: runtime.NewScheme(),
		Ctx:    context.Background(),
	}
	corev1.AddToScheme(h.Scheme)

	t.Cleanup(h.Stop)
	return h
}

func (h *Harness) ParseObjects(y string) []*unstructured.Unstructured {
	t := h.T

	var objects []*unstructured.Unstructured

	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(y)), 100)
	for {
		var rawObj runtime.RawExtension
		if err := decoder.Decode(&rawObj); err != nil {
			if err != io.EOF {
				t.Fatalf("error decoding yaml: %v", err)
			}
			break
		}

		m, _, err := yamlserializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		if err != nil {
			t.Fatalf("error decoding yaml: %v", err)
		}

		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(m)
		if err != nil {
			t.Fatalf("error parsing object: %v", err)
		}
		unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}

		objects = append(objects, unstructuredObj)
	}

	return objects
}

func (h *Harness) WithObjects(initObjs ...*unstructured.Unstructured) {
	if *kubeconfig == "" {
		k8s, err := mockkubeapiserver.NewMockKubeAPIServer(":0")
		if err != nil {
			h.Fatalf("error building mock kube-apiserver: %v", err)
		}
		h.k8s = k8s

		// TODO: Discover from scheme?
		k8s.Add(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Namespace"}, "namespaces", meta.RESTScopeRoot)
		k8s.Add(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Secret"}, "secrets", meta.RESTScopeNamespace)
		k8s.Add(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"}, "configmaps", meta.RESTScopeNamespace)

		addr, err := k8s.StartServing()
		if err != nil {
			h.Errorf("error starting mock kube-apiserver: %v", err)
		}

		h.restConfig = &rest.Config{
			Host: addr.String(),
		}
	} else {
		kubeconfigPath := *kubeconfig
		if strings.HasPrefix(kubeconfigPath, "~/") {
			homeDir := homedir.HomeDir()
			kubeconfigPath = strings.Replace(kubeconfigPath, "~/", homeDir+"/", 1)
		}
		restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			h.Fatalf("error building kubeconfig for %q: %v", kubeconfigPath, err)
		}

		h.restConfig = restConfig
	}

	client, err := client.New(h.RESTConfig(), client.Options{})
	if err != nil {
		h.Fatalf("error building client: %v", err)
	}
	for _, obj := range initObjs {
		if err := client.Create(h.Ctx, obj); err != nil {
			h.Errorf("error creating object %v: %v", obj, err)
		}
	}

	h.client = client
}

func (h *Harness) Stop() {
	if h.k8s != nil {
		if err := h.k8s.Stop(); err != nil {
			h.Errorf("error closing mock kube-apiserver: %v", err)
		}
	}
}

func (h *Harness) DynamicClient() dynamic.Interface {
	dynamicClient, err := dynamic.NewForConfig(h.RESTConfig())
	if err != nil {
		h.Fatalf("error building dynamicClient: %v", err)
	}
	return dynamicClient
}

func (h *Harness) Client() client.Client {
	if h.client == nil {
		h.Fatalf("must call Start() before Client()")
	}
	return h.client
}

func (h *Harness) RESTConfig() *rest.Config {
	if h.restConfig == nil {
		h.Fatalf("cannot call RESTConfig before Start")
	}
	return h.restConfig
}

func (h *Harness) RESTMapper() *restmapper.DeferredDiscoveryRESTMapper {
	if h.restMapper == nil {
		// discoveryClient, err := discovery.NewDiscoveryClientForConfig(h.RESTConfig())
		// if err != nil {
		// 	h.Fatalf("error building discovery client: %")
		// }

		// TODO: Use memory cache or simplified rest mapper
		discoveryClient, err := disk.NewCachedDiscoveryClientForConfig(h.RESTConfig(), "/home/justinsb/tmp/discovery", "", time.Minute)
		if err != nil {
			h.Fatalf("error building discovery client: %v", err)
		}

		restMapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)

		h.restMapper = restMapper
	}

	return h.restMapper
}
