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

package model

import (
	"encoding/json"
	"io"
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

func TestBuildAzure(t *testing.T) {
	const (
		subscriptionID    = "subID"
		tenantID          = "tenantID"
		resourceGroupName = "test-resource-group"
		routeTableName    = "test-route-table"
		vnetName          = "test-vnet"
	)
	cluster := &kops.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testcluster.test.com",
		},
		Spec: kops.ClusterSpec{
			CloudProvider: kops.CloudProviderSpec{
				Azure: &kops.AzureSpec{
					SubscriptionID:    subscriptionID,
					TenantID:          tenantID,
					ResourceGroupName: resourceGroupName,
					RouteTableName:    routeTableName,
				},
			},
			NetworkID: vnetName,
			Subnets: []kops.ClusterSubnetSpec{
				{
					Name:   "test-subnet",
					Region: "eastus",
				},
			},
		},
	}

	b := &CloudConfigBuilder{
		NodeupModelContext: &NodeupModelContext{
			CloudProvider: kops.CloudProviderAzure,
			Cluster:       cluster,
		},
	}
	ctx := &fi.ModelBuilderContext{
		Tasks: map[string]fi.Task{},
	}
	if err := b.Build(ctx); err != nil {
		t.Fatalf("unexpected error from Build(): %v", err)
	}
	var task *nodetasks.File
	for _, v := range ctx.Tasks {
		if f, ok := v.(*nodetasks.File); ok {
			task = f
			break
		}
	}
	if task == nil {
		t.Fatalf("no File task found")
	}
	r, err := task.Contents.Open()
	if err != nil {
		t.Fatalf("unexpected error from task.Contents.Open(): %v", err)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("unexpected error from io.ReadAll(): %v", err)
	}
	var actual azureCloudConfig
	if err := json.Unmarshal(data, &actual); err != nil {
		t.Fatalf("unexpected error from json.Unmarshal(%q): %v", string(data), err)
	}
	expected := azureCloudConfig{
		CloudConfigType:             "file",
		SubscriptionID:              subscriptionID,
		TenantID:                    tenantID,
		Location:                    "eastus",
		VMType:                      "vmss",
		ResourceGroup:               resourceGroupName,
		RouteTableName:              routeTableName,
		VnetName:                    vnetName,
		UseInstanceMetadata:         true,
		UseManagedIdentityExtension: true,
		DisableAvailabilitySetNodes: true,
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %+v, but got %+v", expected, actual)
	}
}

func TestBuildAWSCustomNodeIPFamilies(t *testing.T) {
	cluster := &kops.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testcluster.test.com",
		},
		Spec: kops.ClusterSpec{
			CloudProvider: kops.CloudProviderSpec{
				AWS: &kops.AWSSpec{},
			},
			CloudConfig: &kops.CloudConfiguration{
				NodeIPFamilies: []string{"ipv6"},
			},
			ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
				CloudProvider: string(kops.CloudProviderAWS),
			},
			NonMasqueradeCIDR: "::/0",
		},
	}

	b := &CloudConfigBuilder{
		NodeupModelContext: &NodeupModelContext{
			CloudProvider: kops.CloudProviderAWS,
			Cluster:       cluster,
		},
	}
	ctx := &fi.ModelBuilderContext{
		Tasks: map[string]fi.Task{},
	}
	if err := b.Build(ctx); err != nil {
		t.Fatalf("unexpected error from Build(): %v", err)
	}
	var task *nodetasks.File
	for _, v := range ctx.Tasks {
		if f, ok := v.(*nodetasks.File); ok && f.Path == CloudConfigFilePath {
			task = f
			break
		}
	}
	if task == nil {
		t.Fatalf("no File task found")
	}
	r, err := task.Contents.Open()
	if err != nil {
		t.Fatalf("unexpected error from task.Contents.Open(): %v", err)
	}
	awsCloudConfig, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("unexpected error from ReadAll(): %v", err)
	}

	actual := string(awsCloudConfig)
	expected := "[global]\nNodeIPFamilies = ipv6\n"
	if actual != expected {
		diffString := diff.FormatDiff(expected, actual)
		t.Errorf("actual did not match expected:\n%s\n", diffString)
	}
}
