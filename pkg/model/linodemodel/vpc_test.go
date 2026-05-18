/*
Copyright 2026 The Kubernetes Authors.

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

package linodemodel

import (
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linodetasks"
)

func TestVPCModelBuilderBuild(t *testing.T) {
	cluster := &kops.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: "example.k8s.local"},
		Spec: kops.ClusterSpec{
			Networking: kops.NetworkingSpec{
				Subnets: []kops.ClusterSubnetSpec{{Name: "us-east", Region: "us-east"}},
			},
		},
	}
	b := &VPCModelBuilder{
		LinodeModelContext: &LinodeModelContext{KopsModelContext: &model.KopsModelContext{IAMModelContext: iam.IAMModelContext{Cluster: cluster}}},
		Lifecycle:          fi.LifecycleSync,
	}
	context := &fi.CloudupModelBuilderContext{Tasks: map[string]fi.CloudupTask{}}

	if err := b.Build(context); err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	if got, want := len(context.Tasks), 1; got != want {
		t.Fatalf("unexpected task count: got %d, want %d", got, want)
	}

	for _, task := range context.Tasks {
		vpc, ok := task.(*linodetasks.VPC)
		if !ok {
			t.Fatalf("expected VPC task, got %T", task)
		}
		if got, want := fi.ValueOf(vpc.Name), "example-k8s-local"; got != want {
			t.Fatalf("unexpected VPC name: got %q, want %q", got, want)
		}
		if got, want := fi.ValueOf(vpc.Description), "kOps VPC for example.k8s.local"; got != want {
			t.Fatalf("unexpected VPC description: got %q, want %q", got, want)
		}
		if got, want := fi.ValueOf(vpc.Region), "us-east"; got != want {
			t.Fatalf("unexpected VPC region: got %q, want %q", got, want)
		}
		if got, want := vpc.Lifecycle, fi.LifecycleSync; got != want {
			t.Fatalf("unexpected lifecycle: got %q, want %q", got, want)
		}
	}
}

func TestVPCModelBuilderBuildRequiresRegion(t *testing.T) {
	cluster := &kops.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "example.k8s.local"}}
	b := &VPCModelBuilder{
		LinodeModelContext: &LinodeModelContext{KopsModelContext: &model.KopsModelContext{IAMModelContext: iam.IAMModelContext{Cluster: cluster}}},
		Lifecycle:          fi.LifecycleSync,
	}
	context := &fi.CloudupModelBuilderContext{Tasks: map[string]fi.CloudupTask{}}

	err := b.Build(context)
	if err == nil {
		t.Fatalf("expected region error")
	}
	if !strings.Contains(err.Error(), "linode VPC requires at least one subnet with a region") {
		t.Fatalf("unexpected error: %v", err)
	}
}
