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

func TestNetworkModelBuilderBuild(t *testing.T) {
	cluster := &kops.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: "example.k8s.local"},
		Spec: kops.ClusterSpec{
			Networking: kops.NetworkingSpec{
				Subnets: []kops.ClusterSubnetSpec{{Name: "us-east", Region: "us-east", CIDR: "172.16.1.0/16"}},
			},
		},
	}
	b := &NetworkModelBuilder{
		LinodeModelContext: &LinodeModelContext{KopsModelContext: &model.KopsModelContext{IAMModelContext: iam.IAMModelContext{Cluster: cluster}}},
		Lifecycle:          fi.LifecycleSync,
	}
	context := &fi.CloudupModelBuilderContext{Tasks: map[string]fi.CloudupTask{}}

	if err := b.Build(context); err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	if got, want := len(context.Tasks), 2; got != want {
		t.Fatalf("unexpected task count: got %d, want %d", got, want)
	}

	var vpcTask *linodetasks.VPC
	var subnetTasks []*linodetasks.Subnet
	for _, task := range context.Tasks {
		switch typed := task.(type) {
		case *linodetasks.VPC:
			if vpcTask != nil {
				t.Fatalf("found multiple VPC tasks")
			}
			vpcTask = typed
		case *linodetasks.Subnet:
			subnetTasks = append(subnetTasks, typed)
		default:
			t.Fatalf("unexpected task type %T", task)
		}
	}

	if vpcTask == nil {
		t.Fatalf("expected VPC task")
	}
	if got, want := fi.ValueOf(vpcTask.Name), "example-k8s-local"; got != want {
		t.Fatalf("unexpected VPC name: got %q, want %q", got, want)
	}
	if got, want := fi.ValueOf(vpcTask.Description), "kOps VPC for example.k8s.local"; got != want {
		t.Fatalf("unexpected VPC description: got %q, want %q", got, want)
	}
	if got, want := fi.ValueOf(vpcTask.Region), "us-east"; got != want {
		t.Fatalf("unexpected VPC region: got %q, want %q", got, want)
	}
	if got, want := vpcTask.Lifecycle, fi.LifecycleSync; got != want {
		t.Fatalf("unexpected lifecycle: got %q, want %q", got, want)
	}

	if got, want := len(subnetTasks), 1; got != want {
		t.Fatalf("unexpected subnet task count: got %d, want %d", got, want)
	}
	subnetTask := subnetTasks[0]
	if got, want := fi.ValueOf(subnetTask.Name), "example-k8s-local-us-east"; got != want {
		t.Fatalf("unexpected Subnet name: got %q, want %q", got, want)
	}
	if got, want := fi.ValueOf(subnetTask.IPv4), "172.16.1.0/16"; got != want {
		t.Fatalf("unexpected Subnet IPv4: got %q, want %q", got, want)
	}
	if subnetTask.VPC != vpcTask {
		t.Fatalf("expected subnet to reference shared VPC task")
	}
}

func TestNetworkModelBuilderBuildValidation(t *testing.T) {
	testCases := []struct {
		name          string
		subnets       []kops.ClusterSubnetSpec
		errorContains string
	}{
		{
			name:          "requires subnet",
			errorContains: "linode VPC requires at least one subnet",
		},
		{
			name:          "requires subnet region",
			subnets:       []kops.ClusterSubnetSpec{{Name: "us-east", CIDR: "172.16.1.0/16"}},
			errorContains: "linode subnet \"us-east\" requires a region",
		},
		{
			name:          "requires subnet cidr",
			subnets:       []kops.ClusterSubnetSpec{{Name: "us-east", Region: "us-east"}},
			errorContains: "linode subnet \"us-east\" requires a CIDR",
		},
		{
			name: "requires consistent region",
			subnets: []kops.ClusterSubnetSpec{
				{Name: "us-east-1", Region: "us-east", CIDR: "172.16.1.0/24"},
				{Name: "us-west-1", Region: "us-west", CIDR: "172.16.2.0/24"},
			},
			errorContains: "linode subnets must all use the same region",
		},
		{
			name: "rejects normalized label collisions",
			subnets: []kops.ClusterSubnetSpec{
				{Name: "us.east", Region: "us-east", CIDR: "172.16.1.0/24"},
				{Name: "us-east", Region: "us-east", CIDR: "172.16.2.0/24"},
			},
			errorContains: "normalize to the same label",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			cluster := &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{Name: "example.k8s.local"},
				Spec: kops.ClusterSpec{
					Networking: kops.NetworkingSpec{Subnets: testCase.subnets},
				},
			}
			b := &NetworkModelBuilder{
				LinodeModelContext: &LinodeModelContext{KopsModelContext: &model.KopsModelContext{IAMModelContext: iam.IAMModelContext{Cluster: cluster}}},
				Lifecycle:          fi.LifecycleSync,
			}
			context := &fi.CloudupModelBuilderContext{Tasks: map[string]fi.CloudupTask{}}

			err := b.Build(context)
			if err == nil {
				t.Fatalf("expected validation error")
			}
			if !strings.Contains(err.Error(), testCase.errorContains) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
