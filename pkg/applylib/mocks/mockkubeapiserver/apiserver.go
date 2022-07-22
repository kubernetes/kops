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

package mockkubeapiserver

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

func NewMockKubeAPIServer(addr string) (*MockKubeAPIServer, error) {
	s := &MockKubeAPIServer{
		objects: make(map[schema.GroupResource]*objectList),
	}
	if addr == "" {
		addr = ":http"
	}

	s.httpServer = &http.Server{Addr: addr, Handler: s}

	return s, nil
}

type MockKubeAPIServer struct {
	httpServer *http.Server
	listener   net.Listener

	schema  mockSchema
	objects map[schema.GroupResource]*objectList
}

type mockSchema struct {
	resources []mockSchemaResource
}

type mockSchemaResource struct {
	metav1.APIResource
}

func (s *MockKubeAPIServer) StartServing() (net.Addr, error) {
	listener, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return nil, err
	}
	s.listener = listener
	addr := listener.Addr()
	go func() {
		if err := s.httpServer.Serve(s.listener); err != nil {
			if err != http.ErrServerClosed {
				klog.Errorf("error serving: %v", err)
			}
		}
	}()
	return addr, nil
}

func (s *MockKubeAPIServer) Stop() error {
	return s.httpServer.Close()
}

func (s *MockKubeAPIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	klog.Infof("kubeapiserver request: %s %s", r.Method, r.URL)

	path := r.URL.Path
	tokens := strings.Split(strings.Trim(path, "/"), "/")

	var req Request

	// matchedPath is bool if we recognized the path, but if we didn't build a req we should send StatusMethodNotAllowed instead of NotFound
	var matchedPath bool

	if len(tokens) == 2 {
		if tokens[0] == "api" && tokens[1] == "v1" {
			matchedPath = true

			switch r.Method {
			case http.MethodGet:
				req = &apiResourceList{
					Group:   "",
					Version: "v1",
				}
			}
		}
	}
	if len(tokens) == 1 {
		if tokens[0] == "api" {
			matchedPath = true

			switch r.Method {
			case http.MethodGet:
				req = &apiVersionsRequest{}
			}
		}

		if tokens[0] == "apis" {
			matchedPath = true
			switch r.Method {
			case http.MethodGet:
				req = &apiGroupList{}
			}
		}
	}

	if len(tokens) == 3 {
		if tokens[0] == "apis" {
			matchedPath = true
			switch r.Method {
			case http.MethodGet:
				req = &apiResourceList{
					Group:   tokens[1],
					Version: tokens[2],
				}
			}
		}
	}

	buildObjectRequest := func(common resourceRequestBase) {
		switch r.Method {
		case http.MethodGet:
			req = &getResource{
				resourceRequestBase: common,
			}
		case http.MethodPatch:
			req = &patchResource{
				resourceRequestBase: common,
			}
		case http.MethodPut:
			req = &putResource{
				resourceRequestBase: common,
			}
		}
	}

	if len(tokens) == 4 {
		if tokens[0] == "api" {
			buildObjectRequest(resourceRequestBase{
				Group:    "",
				Version:  tokens[1],
				Resource: tokens[2],
				Name:     tokens[3],
			})
			matchedPath = true
		}
	}
	if len(tokens) == 6 {
		if tokens[0] == "api" && tokens[2] == "namespaces" {
			buildObjectRequest(resourceRequestBase{
				Group:     "",
				Version:   tokens[1],
				Resource:  tokens[4],
				Namespace: tokens[3],
				Name:      tokens[5],
			})
			matchedPath = true
		}
	}
	if len(tokens) == 7 {
		if tokens[0] == "apis" && tokens[3] == "namespaces" {
			buildObjectRequest(resourceRequestBase{
				Group:     tokens[1],
				Version:   tokens[2],
				Namespace: tokens[4],
				Resource:  tokens[5],
				Name:      tokens[6],
			})
			matchedPath = true
		}
	}
	if len(tokens) == 8 {
		if tokens[0] == "apis" && tokens[3] == "namespaces" {
			buildObjectRequest(resourceRequestBase{
				Group:       tokens[1],
				Version:     tokens[2],
				Namespace:   tokens[4],
				Resource:    tokens[5],
				Name:        tokens[6],
				SubResource: tokens[7],
			})
			matchedPath = true
		}
	}

	if req == nil {
		if matchedPath {
			klog.Warningf("method not allowed for %s %s", r.Method, r.URL)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		} else {
			klog.Warningf("404 for %s %s tokens=%#v", r.Method, r.URL, tokens)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		}
		return
	}

	req.Init(w, r)

	err := req.Run(s)
	if err != nil {
		klog.Warningf("internal error for %s %s: %v", r.Method, r.URL, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

}

type Request interface {
	Run(s *MockKubeAPIServer) error
	Init(w http.ResponseWriter, r *http.Request)
}

// baseRequest is the base for our higher-level http requests
type baseRequest struct {
	w http.ResponseWriter
	r *http.Request
}

func (b *baseRequest) Init(w http.ResponseWriter, r *http.Request) {
	b.w = w
	b.r = r
}

func (r *baseRequest) writeResponse(obj interface{}) error {
	b, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("error from json.Marshal on %T: %w", obj, err)
	}
	r.w.Header().Add("Content-Type", "application/json")
	r.w.Header().Add("Cache-Control", "no-cache, private")

	if _, err := r.w.Write(b); err != nil {
		// Too late to send error response
		klog.Warningf("error writing http response: %w", err)
		return nil
	}
	return nil
}

func (r *baseRequest) writeErrorResponse(statusCode int) error {
	klog.Warningf("404 for %s %s", r.r.Method, r.r.URL)
	http.Error(r.w, http.StatusText(statusCode), statusCode)

	return nil
}

// Add registers a type with the schema for the mock kubeapiserver
func (s *MockKubeAPIServer) Add(gvk schema.GroupVersionKind, resource string, scope meta.RESTScope) {
	r := mockSchemaResource{
		APIResource: metav1.APIResource{
			Name:    resource,
			Group:   gvk.Group,
			Version: gvk.Version,
			Kind:    gvk.Kind,
		},
	}
	if scope.Name() == meta.RESTScopeNameNamespace {
		r.Namespaced = true
	}

	s.schema.resources = append(s.schema.resources, r)
}

// AddObject pre-creates an object
func (s *MockKubeAPIServer) AddObject(obj *unstructured.Unstructured) error {
	gv, err := schema.ParseGroupVersion(obj.GetAPIVersion())
	if err != nil {
		return fmt.Errorf("cannot parse apiVersion %q: %w", obj.GetAPIVersion(), err)
	}
	kind := obj.GetKind()

	id := types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}

	for _, resource := range s.schema.resources {
		if resource.Group != gv.Group || resource.Version != gv.Version {
			continue
		}
		if resource.Kind != kind {
			continue
		}

		gr := schema.GroupResource{Group: resource.Group, Resource: resource.Name}
		objects := s.objects[gr]
		if objects == nil {
			objects = &objectList{
				GroupResource: gr,
				Objects:       make(map[types.NamespacedName]*unstructured.Unstructured),
			}
			s.objects[gr] = objects
		}

		objects.Objects[id] = obj
		return nil
	}
	gvk := gv.WithKind(kind)
	return fmt.Errorf("object group/version/kind %v not known", gvk)
}

type objectList struct {
	GroupResource schema.GroupResource
	Objects       map[types.NamespacedName]*unstructured.Unstructured
}
