// Copyright 2015 Google Inc. All Rights Reserved.
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

package util

import (
	"fmt"
	"k8s.io/apimachinery/pkg/fields"
	kube_client "k8s.io/client-go/kubernetes"
	v1listers "k8s.io/client-go/listers/core/v1"
	kube_api "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
	"sort"
	"strings"
	"time"
)

var labelSeperator string

// Concatenates a map of labels into a Seperator-seperated key:value pairs.
func LabelsToString(labels map[string]string) string {
	output := make([]string, 0, len(labels))
	for key, value := range labels {
		output = append(output, fmt.Sprintf("%s:%s", key, value))
	}

	// Sort to produce a stable output.
	sort.Strings(output)
	return strings.Join(output, labelSeperator)
}

func CopyLabels(labels map[string]string) map[string]string {
	c := make(map[string]string, len(labels))
	for key, val := range labels {
		c[key] = val
	}
	return c
}

func GetLatest(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func SetLabelSeperator(seperator string) {
	labelSeperator = seperator
}

func GetNodeLister(kubeClient *kube_client.Clientset) (v1listers.NodeLister, *cache.Reflector, error) {
	lw := cache.NewListWatchFromClient(kubeClient.Core().RESTClient(), "nodes", kube_api.NamespaceAll, fields.Everything())
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	nodeLister := v1listers.NewNodeLister(store)
	reflector := cache.NewReflector(lw, &kube_api.Node{}, store, time.Hour)
	reflector.Run()

	return nodeLister, reflector, nil
}
