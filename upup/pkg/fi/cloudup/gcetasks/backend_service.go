/*
Copyright 2022 The Kubernetes Authors.

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
	"fmt"
	"strings"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// BackendService represents a GCE's backend service type, part of a load balancer.
// +kops:fitask
type BackendService struct {
	Name                  *string
	HealthChecks          []*HealthCheck
	LoadBalancingScheme   *string
	Protocol              *string
	InstanceGroupManagers []*InstanceGroupManager

	Lifecycle    fi.Lifecycle
	ForAPIServer bool

	// fingerprint is required by the API when updating the backend service.
	// Only set on the actual resource returned by Find.
	fingerprint string

	// pruneHealthChecks are names of health checks that this backend service no longer uses and
	// that should be deleted once it has been updated.
	pruneHealthChecks []string
}

var _ fi.CompareWithID = (*BackendService)(nil)

// PruneHealthCheckWithName deletes the named regional health check after the backend service
// has been updated. GCE cannot change the type of an existing health check, so a protocol change
// is a new health check plus removal of the old one.
func (e *BackendService) PruneHealthCheckWithName(name string) {
	e.pruneHealthChecks = append(e.pruneHealthChecks, name)
}

func (e *BackendService) CompareWithID() *string {
	return e.Name
}

func (e *BackendService) Find(c *fi.CloudupContext) (*BackendService, error) {
	actual, err := e.find(c.T.Cloud.(gce.GCECloud))
	if actual != nil && err == nil {
		// Ignore system fields
		actual.Lifecycle = e.Lifecycle
		actual.ForAPIServer = e.ForAPIServer
	}
	return actual, err
}

func (e *BackendService) find(cloud gce.GCECloud) (*BackendService, error) {
	r, err := cloud.Compute().RegionBackendServices().Get(cloud.Project(), cloud.Region(), *e.Name)
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("error listing Backend Services: %v", err)
	}

	actual := &BackendService{}
	actual.Name = &r.Name
	actual.Protocol = &r.Protocol
	actual.fingerprint = r.Fingerprint
	actual.LoadBalancingScheme = &r.LoadBalancingScheme
	var hcs []*HealthCheck
	for _, hc := range r.HealthChecks {
		nameParts := strings.Split(hc, "/")
		hcs = append(hcs, &HealthCheck{Name: &nameParts[len(nameParts)-1]})
	}
	actual.HealthChecks = hcs
	var igms []*InstanceGroupManager
	for _, be := range r.Backends {
		if be.Group == "" {
			continue
		}
		nameParts := strings.Split(be.Group, "/")
		igms = append(igms, &InstanceGroupManager{Name: &nameParts[len(nameParts)-1]})
	}
	actual.InstanceGroupManagers = igms

	return actual, nil
}

func (e *BackendService) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
}

func (_ *BackendService) CheckChanges(a, e, changes *BackendService) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
	}
	return nil
}

func (_ *BackendService) RenderGCE(t *gce.GCEAPITarget, a, e, changes *BackendService) error {
	cloud := t.Cloud
	var hcs []string
	for _, hc := range e.HealthChecks {
		hcs = append(hcs, hc.URL(cloud))
	}
	var backends []*compute.Backend
	for _, igm := range e.InstanceGroupManagers {
		backends = append(backends, &compute.Backend{
			Group: fmt.Sprintf("https://compute.googleapis.com/compute/v1/projects/%s/zones/%s/instanceGroups/%s", cloud.Project(), *igm.Zone, *igm.Name),
		})
	}
	bs := &compute.BackendService{
		Name:                *e.Name,
		Protocol:            *e.Protocol,
		HealthChecks:        hcs,
		LoadBalancingScheme: *e.LoadBalancingScheme,
		Backends:            backends,
	}

	if a == nil {
		klog.V(2).Infof("Creating BackendService: %q", bs.Name)

		op, err := cloud.Compute().RegionBackendServices().Insert(cloud.Project(), cloud.Region(), bs)
		if err != nil {
			return fmt.Errorf("error creating backend service: %v", err)
		}

		if err := cloud.WaitForOp(op); err != nil {
			return fmt.Errorf("error waiting for backend service: %v", err)
		}
	} else if changes.HealthChecks != nil {
		klog.V(2).Infof("Updating health checks of BackendService: %q", bs.Name)

		patch := &compute.BackendService{
			HealthChecks: hcs,
			Fingerprint:  a.fingerprint,
		}
		op, err := cloud.Compute().RegionBackendServices().Patch(cloud.Project(), cloud.Region(), bs.Name, patch)
		if err != nil {
			return fmt.Errorf("error updating backend service: %v", err)
		}

		if err := cloud.WaitForOp(op); err != nil {
			return fmt.Errorf("error waiting for backend service: %v", err)
		}
	} else {
		return fmt.Errorf("cannot apply changes to backend service: %v", changes)
	}

	return nil
}

var _ fi.CloudupProducesDeletions = (*BackendService)(nil)

// FindDeletions implements fi.CloudupProducesDeletions
func (e *BackendService) FindDeletions(c *fi.CloudupContext) ([]fi.CloudupDeletion, error) {
	var removals []fi.CloudupDeletion

	cloud := c.T.Cloud.(gce.GCECloud)
	for _, name := range e.pruneHealthChecks {
		hc, err := cloud.Compute().RegionHealthChecks().Get(cloud.Project(), cloud.Region(), name)
		if err != nil {
			if gce.IsNotFound(err) {
				continue
			}
			return nil, fmt.Errorf("getting health check %q: %w", name, err)
		}
		removals = append(removals, &deleteHealthCheck{healthCheck: hc})
	}

	return removals, nil
}

// deleteHealthCheck tracks a regional HealthCheck that we're going to delete
// It implements fi.CloudupDeletion
type deleteHealthCheck struct {
	healthCheck *compute.HealthCheck
}

var _ fi.CloudupDeletion = (*deleteHealthCheck)(nil)

func (d *deleteHealthCheck) TaskName() string {
	return "HealthCheck"
}

func (d *deleteHealthCheck) Item() string {
	return d.healthCheck.Name
}

func (d *deleteHealthCheck) Delete(t fi.CloudupTarget) error {
	gceTarget, ok := t.(*gce.GCEAPITarget)
	if !ok {
		return fmt.Errorf("unexpected target type for deletion: %T", t)
	}
	cloud := gceTarget.Cloud
	name := d.healthCheck.Name

	op, err := cloud.Compute().RegionHealthChecks().Delete(cloud.Project(), cloud.Region(), name)
	if err != nil {
		if gce.IsResourceInUse(err) {
			// Another backend service still references it; it will be removed on a later update.
			klog.Warningf("not deleting health check %q because it is still in use", name)
			return nil
		}
		return fmt.Errorf("deleting health check %q: %w", name, err)
	}
	if err := cloud.WaitForOp(op); err != nil {
		if gce.IsResourceInUse(err) {
			klog.Warningf("not deleting health check %q because it is still in use", name)
			return nil
		}
		return fmt.Errorf("deleting health check %q: %w", name, err)
	}
	return nil
}

func (d *deleteHealthCheck) String() string {
	return d.TaskName() + "-" + d.Item()
}

func (d *deleteHealthCheck) DeferDeletion() bool {
	return false
}

func (a *BackendService) URL(cloud gce.GCECloud) string {
	return fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/regions/%s/backendServices/%s",
		cloud.Project(),
		cloud.Region(),
		*a.Name)
}

type terraformBackend struct {
	BalancingMode *terraformWriter.Literal `cty:"balancing_mode"`
	Group         *terraformWriter.Literal `cty:"group"`
}

type terraformBackendService struct {
	Name                *string                    `cty:"name"`
	HealthChecks        []*terraformWriter.Literal `cty:"health_checks"`
	LoadBalancingScheme *string                    `cty:"load_balancing_scheme"`
	Protocol            *string                    `cty:"protocol"`
	Backend             []terraformBackend         `cty:"backend"`
}

func (_ *BackendService) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *BackendService) error {
	tf := &terraformBackendService{
		Name:                e.Name,
		LoadBalancingScheme: e.LoadBalancingScheme,
		Protocol:            e.Protocol,
	}

	var igms []terraformBackend
	for _, ig := range e.InstanceGroupManagers {
		igms = append(igms, terraformBackend{
			BalancingMode: terraformWriter.LiteralFromStringValue("CONNECTION"),
			Group:         terraformWriter.LiteralProperty("google_compute_instance_group_manager", *ig.Name, "instance_group"),
		})
	}
	tf.Backend = igms

	var hcs []*terraformWriter.Literal
	for _, hc := range e.HealthChecks {
		hcs = append(hcs, terraformWriter.LiteralProperty("google_compute_region_health_check", *hc.Name, "id"))
	}
	tf.HealthChecks = hcs

	return t.RenderResource("google_compute_region_backend_service", *e.Name, tf)
}

func (e *BackendService) TerraformAddress() *terraformWriter.Literal {
	name := fi.ValueOf(e.Name)

	return terraformWriter.LiteralProperty("google_compute_region_backend_service", name, "id")
}
