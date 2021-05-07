/*
Copyright 2021 The Kubernetes Authors.

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

package fluentest

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type fluentOptions struct {
	ctx        context.Context
	client     kubernetes.Interface
	restConfig *rest.Config
}

type fluentBase struct {
	obj  runtime.Object
	meta metav1.Object

	fluentOptions
}

func (o *fluentBase) Format(f fmt.State, c rune) {
	s := ""

	switch c {
	case 's', 'v':
		s = o.String()
	default:
		s = fmt.Sprintf("unsupported formatter for fluentBase %v", c)
	}

	f.Write([]byte(s))
}

func (o *fluentBase) String() string {
	return o.GetKind() + ":" + o.GetNamespace() + "/" + o.GetName()
}

func (o *fluentBase) GetName() string {
	return o.meta.GetName()
}

func (o *fluentBase) GetNamespace() string {
	return o.meta.GetNamespace()
}

func (o *fluentBase) GetKind() string {
	return o.GetObjectKind().GroupVersionKind().Kind
}

func (o *fluentBase) GetObjectKind() schema.ObjectKind {
	return o.obj.GetObjectKind()
}
