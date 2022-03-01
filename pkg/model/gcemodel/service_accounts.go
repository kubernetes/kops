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

package gcemodel

import (
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
)

// ServiceAccountsBuilder configures service accounts and grants project permissions
type ServiceAccountsBuilder struct {
	*GCEModelContext

	Lifecycle fi.Lifecycle
}

var _ fi.ModelBuilder = &ServiceAccountsBuilder{}

func (b *ServiceAccountsBuilder) Build(c *fi.ModelBuilderContext) error {
	if b.Cluster.Spec.CloudConfig.GCEServiceAccount != "" {
		serviceAccount := &gcetasks.ServiceAccount{
			Name:      s("shared"),
			Email:     &b.Cluster.Spec.CloudConfig.GCEServiceAccount,
			Shared:    fi.Bool(true),
			Lifecycle: b.Lifecycle,
		}
		c.AddTask(serviceAccount)

		return nil
	}

	doneEmails := make(map[string]bool)
	for _, ig := range b.InstanceGroups {
		link := b.LinkToServiceAccount(ig)
		if fi.BoolValue(link.Shared) {
			c.EnsureTask(link)
			continue
		}

		if doneEmails[*link.Email] {
			continue
		}
		doneEmails[*link.Email] = true

		serviceAccount := &gcetasks.ServiceAccount{
			Name:        link.Name,
			DisplayName: link.Name,
			Email:       link.Email,
			Lifecycle:   b.Lifecycle,
		}
		switch ig.Spec.Role {
		case kops.InstanceGroupRoleAPIServer, kops.InstanceGroupRoleMaster:
			serviceAccount.Description = fi.String("kubernetes control-plane instances")
		case kops.InstanceGroupRoleNode:
			serviceAccount.Description = fi.String("kubernetes worker nodes")
		case kops.InstanceGroupRoleBastion:
			serviceAccount.Description = fi.String("bastion nodes")
		default:
			klog.Warningf("unknown instance role %q", ig.Spec.Role)
		}
		c.AddTask(serviceAccount)

		role := ig.Spec.Role
		if role == kops.InstanceGroupRoleAPIServer {
			// Because these share a serviceaccount, we share a role
			role = kops.InstanceGroupRoleMaster
		}

		if err := b.addInstanceGroupServiceAccountPermissions(c, *serviceAccount.Email, role); err != nil {
			return err
		}
	}

	return nil
}

func (b *ServiceAccountsBuilder) addInstanceGroupServiceAccountPermissions(c *fi.ModelBuilderContext, serviceAccountEmail string, role kops.InstanceGroupRole) error {
	member := "serviceAccount:" + serviceAccountEmail

	// Ideally we would use a custom role here, but the deletion of a custom role takes 7 days,
	// which means we can't easily recycle cluster names.
	// If we can find a solution, we can easily switch to a custom role.

	switch role {
	case kops.InstanceGroupRoleMaster:
		// We reuse the GKE role
		c.AddTask(&gcetasks.ProjectIAMBinding{
			Name:      s("serviceaccount-control-plane"),
			Lifecycle: b.Lifecycle,

			Project: s(b.ProjectID),
			Member:  s(member),
			Role:    s("roles/container.serviceAgent"),
		})

	case kops.InstanceGroupRoleNode:
		// Known permissions:
		//  * compute.zones.list (to find out region; we could replace this with string manipulation)
		//  * compute.instances.list (for discovery; we don't need in the case of a load balancer or DNS)

		// We use the GCE viewer role

		c.AddTask(&gcetasks.ProjectIAMBinding{
			Name:      s("serviceaccount-nodes"),
			Lifecycle: b.Lifecycle,

			Project: s(b.ProjectID),
			Member:  s(member),
			Role:    s("roles/compute.viewer"),
		})
	}
	return nil
}
