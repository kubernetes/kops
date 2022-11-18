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
	"testing"

	gcemock "k8s.io/kops/cloudmock/gce"
	"k8s.io/kops/upup/pkg/fi"
)

func TestProjectIAMBinding(t *testing.T) {
	project := "testproject"
	region := "us-test1"

	cloud := gcemock.InstallMockGCECloud(region, project)

	// We define a function so we can rebuild the tasks, because we modify in-place when running
	buildTasks := func() map[string]fi.Task {
		binding := &ProjectIAMBinding{
			Lifecycle: fi.LifecycleSync,

			Project: fi.PtrTo("testproject"),
			Member:  fi.PtrTo("serviceAccount:foo@testproject.iam.gserviceaccount.com"),
			Role:    fi.PtrTo("roles/owner"),
		}

		return map[string]fi.Task{
			"binding": binding,
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
