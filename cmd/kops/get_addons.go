/*
Copyright 2019 The Kubernetes Authors.

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
	"io"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/util/pkg/tables"
)

func addonsOutputTable(cluster *api.Cluster, addons []*unstructured.Unstructured, out io.Writer) error {
	t := &tables.Table{}
	t.AddColumn("NAME", func(o *unstructured.Unstructured) string {
		return o.GetName()
	})
	t.AddColumn("KIND", func(o *unstructured.Unstructured) string {
		return o.GroupVersionKind().Kind
	})
	t.AddColumn("VERSION", func(o *unstructured.Unstructured) string {
		s, _, _ := unstructured.NestedString(o.Object, "spec", "version")
		return s
	})
	return t.Render(addons, out, "NAME", "KIND", "VERSION")
}
