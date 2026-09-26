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

	"google.golang.org/api/compute/v1"
	gcemock "k8s.io/kops/cloudmock/gce"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

// TestForwardingRuleMovesFromTargetPoolToBackendService verifies the upgrade of a public API
// load balancer created by kOps before 1.38: the forwarding rule is recreated to point at a
// backend service, and the target pool with its legacy HTTP health check is deleted.
func TestForwardingRuleMovesFromTargetPoolToBackendService(t *testing.T) {
	ctx := context.TODO()

	project := "testproject"
	region := "us-test1"

	cloud := gcemock.InstallMockGCECloud(region, project)

	// The resources as created by an older kOps version: a target pool with a legacy HTTP
	// health check, and a forwarding rule targeting the pool. The tasks for these no longer
	// exist, so create them through the cloud API directly.
	{
		httpHealthCheck := &compute.HttpHealthCheck{
			Name:        "api",
			Port:        3990,
			RequestPath: "/healthz",
		}
		if _, err := cloud.Compute().HTTPHealthChecks().Insert(project, httpHealthCheck); err != nil {
			t.Fatalf("creating HTTP health check: %v", err)
		}
		targetPool := &compute.TargetPool{
			Name: "api",
		}
		if _, err := cloud.Compute().TargetPools().Insert(project, region, targetPool); err != nil {
			t.Fatalf("creating target pool: %v", err)
		}
		addHealthCheck := &compute.TargetPoolsAddHealthCheckRequest{
			HealthChecks: []*compute.HealthCheckReference{{HealthCheck: httpHealthCheck.SelfLink}},
		}
		if _, err := cloud.Compute().TargetPools().AddHealthCheck(project, region, "api", addHealthCheck); err != nil {
			t.Fatalf("adding health check to target pool: %v", err)
		}
		forwardingRule := &compute.ForwardingRule{
			Name:                "api",
			PortRange:           "443-443",
			Target:              targetPool.SelfLink,
			IPProtocol:          "TCP",
			LoadBalancingScheme: "EXTERNAL",
		}
		if _, err := cloud.Compute().ForwardingRules().Insert(ctx, project, region, forwardingRule); err != nil {
			t.Fatalf("creating forwarding rule: %v", err)
		}
	}

	// The same load balancer as modeled now.
	newTasks := func() map[string]fi.CloudupTask {
		healthCheck := &HealthCheck{
			Name:        new("api-https"),
			Port:        443,
			Protocol:    HealthCheckProtocolHTTPS,
			RequestPath: new("/readyz"),
			Lifecycle:   fi.LifecycleSync,
		}
		backendService := &BackendService{
			Name:                new("api-public"),
			Protocol:            new("TCP"),
			HealthChecks:        []*HealthCheck{healthCheck},
			LoadBalancingScheme: new("EXTERNAL"),
			Lifecycle:           fi.LifecycleSync,
		}
		forwardingRule := &ForwardingRule{
			Name:                new("api"),
			Lifecycle:           fi.LifecycleSync,
			PortRange:           new("443-443"),
			BackendService:      backendService,
			IPProtocol:          "TCP",
			LoadBalancingScheme: new("EXTERNAL"),
		}
		forwardingRule.PruneTargetPoolWithName("api")
		return map[string]fi.CloudupTask{
			"healthcheck":    healthCheck,
			"backendservice": backendService,
			"forwardingrule": forwardingRule,
		}
	}

	checkHasChanges(t, ctx, cloud, newTasks())
	runTasks(t, ctx, cloud, newTasks())
	checkNoChanges(t, ctx, cloud, newTasks())

	fr, err := cloud.Compute().ForwardingRules().Get(ctx, project, region, "api")
	if err != nil {
		t.Fatalf("getting forwarding rule: %v", err)
	}
	if fr.Target != "" || fr.BackendService == "" {
		t.Errorf("expected forwarding rule to point at the backend service, got target=%q backendService=%q", fr.Target, fr.BackendService)
	}

	if _, err := cloud.Compute().TargetPools().Get(project, region, "api"); !gce.IsNotFound(err) {
		t.Errorf("expected target pool to be deleted, got err=%v", err)
	}
	if _, err := cloud.Compute().HTTPHealthChecks().Get(project, "api"); !gce.IsNotFound(err) {
		t.Errorf("expected legacy HTTP health check to be deleted, got err=%v", err)
	}
}

// TestBackendServiceReplacesHealthCheck verifies that a backend service is moved from the TCP
// health check used by kOps before 1.38 to the HTTPS one, and that the old health check is deleted.
func TestBackendServiceReplacesHealthCheck(t *testing.T) {
	ctx := context.TODO()

	project := "testproject"
	region := "us-test1"

	cloud := gcemock.InstallMockGCECloud(region, project)

	buildTasks := func(healthCheck *HealthCheck, pruneHealthCheck string) map[string]fi.CloudupTask {
		backendService := &BackendService{
			Name:                new("api"),
			Protocol:            new("TCP"),
			HealthChecks:        []*HealthCheck{healthCheck},
			LoadBalancingScheme: new("INTERNAL"),
			Lifecycle:           fi.LifecycleSync,
		}
		if pruneHealthCheck != "" {
			backendService.PruneHealthCheckWithName(pruneHealthCheck)
		}
		return map[string]fi.CloudupTask{
			"healthcheck":    healthCheck,
			"backendservice": backendService,
		}
	}

	tcpHealthCheck := func() *HealthCheck {
		return &HealthCheck{
			Name:      new("api"),
			Port:      443,
			Protocol:  HealthCheckProtocolTCP,
			Lifecycle: fi.LifecycleSync,
		}
	}
	httpsHealthCheck := func() *HealthCheck {
		return &HealthCheck{
			Name:        new("api-https"),
			Port:        443,
			Protocol:    HealthCheckProtocolHTTPS,
			RequestPath: new("/readyz"),
			Lifecycle:   fi.LifecycleSync,
		}
	}

	runTasks(t, ctx, cloud, buildTasks(tcpHealthCheck(), ""))
	checkNoChanges(t, ctx, cloud, buildTasks(tcpHealthCheck(), ""))

	checkHasChanges(t, ctx, cloud, buildTasks(httpsHealthCheck(), "api"))
	runTasks(t, ctx, cloud, buildTasks(httpsHealthCheck(), "api"))
	checkNoChanges(t, ctx, cloud, buildTasks(httpsHealthCheck(), "api"))

	bs, err := cloud.Compute().RegionBackendServices().Get(project, region, "api")
	if err != nil {
		t.Fatalf("getting backend service: %v", err)
	}
	if len(bs.HealthChecks) != 1 || lastComponent(bs.HealthChecks[0]) != "api-https" {
		t.Errorf("expected backend service to use the api-https health check, got %v", bs.HealthChecks)
	}
	if _, err := cloud.Compute().RegionHealthChecks().Get(project, region, "api"); !gce.IsNotFound(err) {
		t.Errorf("expected old health check to be deleted, got err=%v", err)
	}

	// The request path of the HTTPS health check is reconciled in place.
	changed := httpsHealthCheck()
	changed.RequestPath = new("/healthz")
	checkHasChanges(t, ctx, cloud, buildTasks(changed, ""))
	runTasks(t, ctx, cloud, buildTasks(changed, ""))
	checkNoChanges(t, ctx, cloud, buildTasks(changed, ""))
}

// TestBackendServiceUpdatesBackendsAndHealthCheck verifies that a health check replacement and a
// change of the backend instance groups in the same run are both applied.
func TestBackendServiceUpdatesBackendsAndHealthCheck(t *testing.T) {
	ctx := context.TODO()

	project := "testproject"
	region := "us-test1"

	cloud := gcemock.InstallMockGCECloud(region, project)

	// The backend service only needs the names and zones of the instance group managers;
	// the tasks are present so that the dependency resolves, but are not applied.
	igm := func(name string) *InstanceGroupManager {
		return &InstanceGroupManager{Name: new(name), Zone: new("us-test1-a"), Lifecycle: fi.LifecycleIgnore}
	}

	buildTasks := func(healthCheck *HealthCheck, igms []*InstanceGroupManager) map[string]fi.CloudupTask {
		backendService := &BackendService{
			Name:                  new("api"),
			Protocol:              new("TCP"),
			HealthChecks:          []*HealthCheck{healthCheck},
			LoadBalancingScheme:   new("INTERNAL"),
			InstanceGroupManagers: igms,
			Lifecycle:             fi.LifecycleSync,
		}
		tasks := map[string]fi.CloudupTask{
			"healthcheck":    healthCheck,
			"backendservice": backendService,
		}
		for _, igm := range igms {
			tasks[*igm.Name] = igm
		}
		return tasks
	}

	tcpHealthCheck := func() *HealthCheck {
		return &HealthCheck{Name: new("api"), Port: 443, Protocol: HealthCheckProtocolTCP, Lifecycle: fi.LifecycleSync}
	}
	httpsHealthCheck := func() *HealthCheck {
		return &HealthCheck{Name: new("api-https"), Port: 443, Protocol: HealthCheckProtocolHTTPS, RequestPath: new("/readyz"), Lifecycle: fi.LifecycleSync}
	}

	runTasks(t, ctx, cloud, buildTasks(tcpHealthCheck(), []*InstanceGroupManager{igm("igm-a")}))
	checkNoChanges(t, ctx, cloud, buildTasks(tcpHealthCheck(), []*InstanceGroupManager{igm("igm-a")}))

	// Replace the health check and add an instance group in the same run.
	runTasks(t, ctx, cloud, buildTasks(httpsHealthCheck(), []*InstanceGroupManager{igm("igm-a"), igm("igm-b")}))
	checkNoChanges(t, ctx, cloud, buildTasks(httpsHealthCheck(), []*InstanceGroupManager{igm("igm-a"), igm("igm-b")}))

	bs, err := cloud.Compute().RegionBackendServices().Get(project, region, "api")
	if err != nil {
		t.Fatalf("getting backend service: %v", err)
	}
	if len(bs.HealthChecks) != 1 || lastComponent(bs.HealthChecks[0]) != "api-https" {
		t.Errorf("expected backend service to use the api-https health check, got %v", bs.HealthChecks)
	}
	if len(bs.Backends) != 2 || lastComponent(bs.Backends[0].Group) != "igm-a" || lastComponent(bs.Backends[1].Group) != "igm-b" {
		t.Errorf("expected backends igm-a and igm-b, got %v", bs.Backends)
	}
}
