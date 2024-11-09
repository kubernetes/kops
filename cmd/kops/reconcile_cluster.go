/*
Copyright 2024 The Kubernetes Authors.

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

package main

import (
	"context"
	"fmt"
	"io"

	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops"
)

// ReconcileCluster updates the cluster to the desired state, including rolling updates where necessary.
// To respect skew policy, it updates the control plane first, then updates the nodes.
// "update" is probably now smart enough to automatically not update the control plane if it is already at the desired version,
// but we do it explicitly here to be clearer / safer.
func ReconcileCluster(ctx context.Context, f *util.Factory, out io.Writer, c *UpdateClusterOptions) error {
	fmt.Fprintf(out, "Updating control plane configuration\n")
	{
		opt := *c
		opt.Reconcile = false // Prevent infinite loop
		opt.InstanceGroupRoles = []string{
			string(kops.InstanceGroupRoleAPIServer),
			string(kops.InstanceGroupRoleControlPlane),
		}
		if _, err := RunUpdateCluster(ctx, f, out, &opt); err != nil {
			return err
		}
	}

	fmt.Fprintf(out, "Doing rolling-update for control plane\n")
	{
		opt := &RollingUpdateOptions{}
		opt.InitDefaults()
		opt.ClusterName = c.ClusterName
		opt.InstanceGroupRoles = []string{
			string(kops.InstanceGroupRoleAPIServer),
			string(kops.InstanceGroupRoleControlPlane),
		}
		opt.Yes = c.Yes
		if err := RunRollingUpdateCluster(ctx, f, out, opt); err != nil {
			return err
		}
	}

	fmt.Fprintf(out, "Updating node configuration\n")
	{
		opt := *c
		opt.Reconcile = false // Prevent infinite loop
		// Do all roles this time, though we only expect changes to node & bastion roles
		opt.InstanceGroupRoles = nil
		if _, err := RunUpdateCluster(ctx, f, out, &opt); err != nil {
			return err
		}
	}

	fmt.Fprintf(out, "Doing rolling-update for nodes\n")
	{
		opt := &RollingUpdateOptions{}
		opt.InitDefaults()
		opt.ClusterName = c.ClusterName
		// Do all roles this time, though we only expect changes to node & bastion roles
		opt.InstanceGroupRoles = nil
		opt.Yes = c.Yes
		if err := RunRollingUpdateCluster(ctx, f, out, opt); err != nil {
			return err
		}
	}

	return nil
}
