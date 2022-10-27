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

package gcetasks

import (
	"bytes"
	"os"
	"testing"
	"time"

	gcemock "k8s.io/kops/cloudmock/gce"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

func TestServiceAccount(t *testing.T) {
	project := "testproject"
	region := "us-test1"

	cloud := gcemock.InstallMockGCECloud(region, project)

	// We define a function so we can rebuild the tasks, because we modify in-place when running
	buildTasks := func() map[string]fi.Task {
		serviceAccount := &ServiceAccount{
			Name:      fi.String("test"),
			Lifecycle: fi.LifecycleSync,

			Email:       fi.String("test@testproject.iam.gserviceaccount.com"),
			Description: fi.String("description of ServiceAccount"),
			DisplayName: fi.String("display name of ServiceAccount"),
		}

		return map[string]fi.Task{
			*serviceAccount.Name: serviceAccount,
		}
	}

	{
		allTasks := buildTasks()
		checkHasChanges(t, cloud, allTasks)
	}

	{
		allTasks := buildTasks()
		runTasks(t, cloud, allTasks)
	}

	{
		allTasks := buildTasks()
		checkNoChanges(t, cloud, allTasks)
	}
}

// TODO: Dedup with awstasks
var testRunTasksOptions = fi.RunTasksOptions{
	MaxTaskDuration:         2 * time.Second,
	WaitAfterAllTasksFailed: 500 * time.Millisecond,
}

// TODO: Dedup with awstasks
func checkNoChanges(t *testing.T, cloud fi.Cloud, allTasks map[string]fi.Task) {
	target := doDryRun(t, cloud, allTasks)

	if target.HasChanges() {
		var b bytes.Buffer
		if err := target.PrintReport(allTasks, &b); err != nil {
			t.Fatalf("error building report: %v", err)
		}
		t.Fatalf("Target had changes after executing: %v", b.String())
	}
}

func checkHasChanges(t *testing.T, cloud fi.Cloud, allTasks map[string]fi.Task) {
	target := doDryRun(t, cloud, allTasks)

	if !target.HasChanges() {
		t.Fatalf("expected dry-run to have changes")
	}
}

func runTasks(t *testing.T, cloud gce.GCECloud, allTasks map[string]fi.Task) {
	target := gce.NewGCEAPITarget(cloud)

	context, err := fi.NewCloudContext(target, nil, cloud, nil, nil, nil, true, allTasks)
	if err != nil {
		t.Fatalf("error building context: %v", err)
	}
	defer context.Close()

	if err := context.RunTasks(testRunTasksOptions); err != nil {
		t.Fatalf("unexpected error during Run: %v", err)
	}
}

func doDryRun(t *testing.T, cloud fi.Cloud, allTasks map[string]fi.Task) *fi.DryRunTarget {
	cluster := &kops.Cluster{
		Spec: kops.ClusterSpec{
			KubernetesVersion: "v1.23.0",
		},
	}
	assetBuilder := assets.NewAssetBuilder(cluster, false)
	target := fi.NewDryRunTarget(assetBuilder, os.Stderr)
	context, err := fi.NewCloudContext(target, nil, cloud, nil, nil, nil, true, allTasks)
	if err != nil {
		t.Fatalf("error building context: %v", err)
	}
	defer context.Close()

	if err := context.RunTasks(testRunTasksOptions); err != nil {
		t.Fatalf("unexpected error during Run: %v", err)
	}

	return target
}
