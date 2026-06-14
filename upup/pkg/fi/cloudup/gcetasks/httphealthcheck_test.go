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

package gcetasks

import (
	"context"
	"testing"

	gcemock "k8s.io/kops/cloudmock/gce"
	"k8s.io/kops/upup/pkg/fi"
)

// TestHTTPHealthcheckChange verifies that a Port or RequestPath change on an existing GCE HTTP health check is
// actually applied, instead of reappearing as an unapplied change on every run.
func TestHTTPHealthcheckChange(t *testing.T) {
	ctx := context.TODO()

	project := "testproject"
	region := "us-test1"

	cloud := gcemock.InstallMockGCECloud(region, project)

	// We define a function so we can rebuild the tasks, because we modify in-place when running
	buildTasks := func(requestPath string) map[string]fi.CloudupTask {
		healthcheck := &HTTPHealthcheck{
			Name:      fi.PtrTo("api"),
			Lifecycle: fi.LifecycleSync,

			Port:        fi.PtrTo(int64(8080)),
			RequestPath: fi.PtrTo(requestPath),
		}

		return map[string]fi.CloudupTask{
			*healthcheck.Name: healthcheck,
		}
	}

	// Create the health check with one request path (as an older kOps version would).
	runTasks(t, ctx, cloud, buildTasks("/healthz"))

	// Upgrade to a different request path.
	runTasks(t, ctx, cloud, buildTasks("/readyz"))

	// The change must have been applied, so a subsequent run sees no changes.
	checkNoChanges(t, ctx, cloud, buildTasks("/readyz"))
}
