/*
Copyright 2016 The Kubernetes Authors.

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

package federation

import (
	"fmt"
	"github.com/golang/glog"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/federation/client/clientset_generated/federation_clientset"
	"k8s.io/kubernetes/pkg/api/v1"
)

func findNamespace(k8s federation_clientset.Interface, name string) (*v1.Namespace, error) {
	glog.V(2).Infof("querying k8s for federation Namespace %s", name)
	c, err := k8s.CoreV1().Namespaces().Get(name, meta_v1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		} else {
			return nil, fmt.Errorf("error reading federation Namespace %s: %v", name, err)
		}
	}
	return c, nil
}

func mutateNamespace(k8s federation_clientset.Interface, name string, fn func(s *v1.Namespace) (*v1.Namespace, error)) (*v1.Namespace, error) {
	existing, err := findNamespace(k8s, name)
	if err != nil {
		return nil, err
	}
	createObject := existing == nil
	updated, err := fn(existing)
	if err != nil {
		return nil, err
	}

	updated.Name = name

	if createObject {
		glog.V(2).Infof("creating federation Namespace %s", name)
		created, err := k8s.CoreV1().Namespaces().Create(updated)
		if err != nil {
			return nil, fmt.Errorf("error creating federation Namespace %s: %v", name, err)
		}
		return created, nil
	} else {
		glog.V(2).Infof("updating federation Namespace %s", name)
		created, err := k8s.CoreV1().Namespaces().Update(updated)
		if err != nil {
			return nil, fmt.Errorf("error updating federation Namespace %s: %v", name, err)
		}
		return created, nil
	}
}
