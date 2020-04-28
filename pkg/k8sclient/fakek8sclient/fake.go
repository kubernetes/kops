/*
Copyright 2020 The Kubernetes Authors.

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

package fakek8sclient

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/kops/pkg/k8sclient"
)

type Fake struct {
	*fake.Clientset
}

var _ k8sclient.Interface = &Fake{}

func (f *Fake) RawClient() kubernetes.Interface {
	return f.Clientset
}

func (f *Fake) DeleteNode(ctx context.Context, nodeName string) error {
	return f.CoreV1().Nodes().Delete(ctx, nodeName, metav1.DeleteOptions{})
}

func (f *Fake) ListNodes(ctx context.Context) (*corev1.NodeList, error) {
	return f.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
}
