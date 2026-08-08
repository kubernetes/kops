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

package gce

import (
	"context"
	"fmt"
	"strconv"

	compute "google.golang.org/api/compute/v1"
)

// GetInstanceTemplateForMIGMember returns the instance template of the MIG that manages the given
// instance. The instance metadata is potentially mutable by whoever created the instance, so we
// instead resolve the MIG from the created-by metadata and verify that the instance is indeed
// managed by it; MIG membership can't be spoofed without GCE API access.
func GetInstanceTemplateForMIGMember(ctx context.Context, computeService *compute.Service, project string, instance *compute.Instance) (*compute.InstanceTemplate, error) {
	createdBy := GetMetadataValue(instance.Metadata, "created-by")
	if createdBy == "" {
		return nil, fmt.Errorf("cannot find owner for instance %s", instance.Name)
	}

	// We need to double-check the MIG membership, in case created-by was changed
	migName := LastComponent(createdBy)

	migMember, err := getManagedInstance(ctx, computeService, project, migName, instance)
	if err != nil {
		return nil, err
	}

	if migMember.Version == nil {
		return nil, fmt.Errorf("instance %s did not have Version set", instance.Name)
	}

	templateName := LastComponent(migMember.Version.InstanceTemplate)
	instanceTemplate, err := computeService.InstanceTemplates.Get(project, templateName).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("error fetching GCE instance group template %q: %v", templateName, err)
	}

	return instanceTemplate, nil
}

// getManagedInstance queries GCE for the instance from the MIG
func getManagedInstance(ctx context.Context, computeService *compute.Service, project string, migName string, instance *compute.Instance) (*compute.ManagedInstance, error) {
	var matches []*compute.ManagedInstance

	zone := LastComponent(instance.Zone)
	filter := "id=" + strconv.FormatUint(instance.Id, 10)
	if err := computeService.InstanceGroupManagers.ListManagedInstances(project, zone, migName).Filter(filter).Pages(ctx, func(page *compute.InstanceGroupManagersListManagedInstancesResponse) error {
		// Post-filter... filters aren't implemented (b/27605549)
		for _, member := range page.ManagedInstances {
			if member.Id != instance.Id {
				continue
			}
			matches = append(matches, member)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("error fetching GCE managed instance group members for %q: %v", migName, err)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("instance %v not managed by mig %s", instance.Id, migName)
	}
	if len(matches) > 1 {
		// Should be impossible - shows that filters / post-filters are not working
		return nil, fmt.Errorf("found multiple instances with id %v managed by mig %s", instance.Id, migName)
	}

	return matches[0], nil
}

// GetMetadataValue returns the value for the given key in the metadata, or "" if not present.
func GetMetadataValue(metadata *compute.Metadata, key string) string {
	value := ""
	if metadata != nil {
		for _, item := range metadata.Items {
			if item.Key == key && item.Value != nil {
				value = *item.Value
			}
		}
	}
	return value
}
