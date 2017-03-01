// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/golang/glog"
	"github.com/stretchr/testify/assert"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/heapster/metrics/cmd/heapster-apiserver/app"
	"k8s.io/heapster/metrics/options"
	metricsink "k8s.io/heapster/metrics/sinks/metric"
	"k8s.io/metrics/pkg/apis/metrics/v1alpha1"
)

var metricsGroupVersion = v1alpha1.SchemeGroupVersion

var metricsGroupVersionForDiscovery = metav1.GroupVersionForDiscovery{
	GroupVersion: metricsGroupVersion.String(),
	Version:      metricsGroupVersion.Version,
}

var emptyMetricSink = &metricsink.MetricSink{}
var emptyNodeLister = v1listers.NewNodeLister(nil)
var emptyPodLister = v1listers.NewPodLister(nil)

var testSecurePort = 6443

func TestRunSecureServer(t *testing.T) {
	serverIP := fmt.Sprintf("https://localhost:%d", testSecurePort)

	go func() {
		opt := getServerOptions()
		opt.SecureServing.ServingOptions.BindPort = testSecurePort
		opt.SecureServing.ServerCert.CertDirectory = "/tmp"
		opt.DisableAuthForTesting = true

		server, err := app.NewHeapsterApiServer(opt, emptyMetricSink, emptyNodeLister, emptyPodLister)
		if err != nil {
			t.Fatalf("Could not create the API server: %v", err)
		}
		if err := server.RunServer(); err != nil {
			t.Fatalf("Error in bringing up the server: %v", err)
		}
	}()
	if err := waitForApiserverUp(serverIP); err != nil {
		t.Fatalf("%v", err)
	}
	testSwaggerSpec(t, serverIP)
	testAPIGroupList(t, serverIP)
	testAPIGroup(t, serverIP)
	testAPIResourceList(t, serverIP)
}

func getServerOptions() *options.HeapsterRunOptions {
	opt := options.NewHeapsterRunOptions()
	return opt
}

func waitForApiserverUp(serverIP string) error {
	for start := time.Now(); time.Since(start) < time.Minute; time.Sleep(5 * time.Second) {
		glog.Errorf("Waiting for : %#v", serverIP)
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		_, err := client.Get(serverIP)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("waiting for apiserver timed out")
}

func readResponse(serverURL string) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	response, err := client.Get(serverURL)
	if err != nil {
		glog.Errorf("http get err code : %#v", err)
		return nil, fmt.Errorf("Error in fetching %s: %v", serverURL, err)
	}
	defer response.Body.Close()
	glog.Errorf("http get response code : %#v", response.StatusCode)
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d for URL: %s, expected status: %d", response.StatusCode, serverURL, http.StatusOK)
	}
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response from %s: %v", serverURL, err)
	}
	return contents, nil
}

func testSwaggerSpec(t *testing.T, serverIP string) {
	serverURL := serverIP + "/swaggerapi"
	_, err := readResponse(serverURL)
	if err != nil {
		t.Fatalf("%v", err)
	}
}

func testAPIGroupList(t *testing.T, serverIP string) {
	serverURL := serverIP + "/apis"
	contents, err := readResponse(serverURL)
	if err != nil {
		t.Fatalf("%v", err)
	}
	var apiGroupList metav1.APIGroupList
	err = json.Unmarshal(contents, &apiGroupList)
	if err != nil {
		t.Fatalf("Error in unmarshalling response from server %s: %v", serverURL, err)
	}
	assert.Equal(t, 1, len(apiGroupList.Groups))
	assert.Equal(t, apiGroupList.Groups[0].Name, metricsGroupVersion.Group)
	assert.Equal(t, 1, len(apiGroupList.Groups[0].Versions))
	assert.Equal(t, apiGroupList.Groups[0].Versions[0], metricsGroupVersionForDiscovery)
	assert.Equal(t, apiGroupList.Groups[0].PreferredVersion, metricsGroupVersionForDiscovery)
}

func testAPIGroup(t *testing.T, serverIP string) {
	serverURL := serverIP + "/apis/metrics"
	contents, err := readResponse(serverURL)
	if err != nil {
		t.Fatalf("%v", err)
	}
	var apiGroup metav1.APIGroup
	err = json.Unmarshal(contents, &apiGroup)
	if err != nil {
		t.Fatalf("Error in unmarshalling response from server %s: %v", serverURL, err)
	}
	assert.Equal(t, apiGroup.APIVersion, "v1")
	assert.Equal(t, apiGroup.Name, metricsGroupVersion.Group)
	assert.Equal(t, 1, len(apiGroup.Versions))
	assert.Equal(t, apiGroup.Versions[0].GroupVersion, metricsGroupVersion.String())
	assert.Equal(t, apiGroup.Versions[0].Version, metricsGroupVersion.Version)
	assert.Equal(t, apiGroup.Versions[0], apiGroup.PreferredVersion)
}

func testAPIResourceList(t *testing.T, serverIP string) {
	serverURL := serverIP + "/apis/metrics/v1alpha1"
	contents, err := readResponse(serverURL)
	if err != nil {
		t.Fatalf("%v", err)
	}
	var apiResourceList metav1.APIResourceList
	err = json.Unmarshal(contents, &apiResourceList)
	if err != nil {
		t.Fatalf("Error in unmarshalling response from server %s: %v", serverURL, err)
	}
	assert.Equal(t, apiResourceList.APIVersion, "v1")
	assert.Equal(t, apiResourceList.GroupVersion, metricsGroupVersion.String())
	assert.Equal(t, 2, len(apiResourceList.APIResources))
	assert.Equal(t, apiResourceList.APIResources[0].Name, "nodes")
	assert.False(t, apiResourceList.APIResources[0].Namespaced)
	assert.Equal(t, apiResourceList.APIResources[0].Kind, "NodeMetrics")
	assert.Equal(t, apiResourceList.APIResources[1].Name, "pods")
	assert.True(t, apiResourceList.APIResources[1].Namespaced)
	assert.Equal(t, apiResourceList.APIResources[1].Kind, "PodMetrics")
}
