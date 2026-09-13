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

package dump

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
	"sigs.k8s.io/yaml"
)

// CloudInstanceGroupsFileName is the artifact the cloud group state is written to.
const CloudInstanceGroupsFileName = "cloud-instance-groups.yaml"

// CloudInstanceGroupDump is the cloud provider's view of one instance group.
type CloudInstanceGroupDump struct {
	// Name is the cloud provider's name for the group, e.g. the ASG or MIG name.
	Name string `json:"name"`
	// InstanceGroup is the name of the kOps InstanceGroup the group backs.
	InstanceGroup string `json:"instanceGroup"`
	MinSize       int    `json:"minSize"`
	TargetSize    int    `json:"targetSize"`
	MaxSize       int    `json:"maxSize"`

	Instances []CloudInstanceDump `json:"instances,omitempty"`
	// ManagedInstances is the provider's view of the instances the group is
	// managing, which on GCE includes ones it is still failing to create. Absent
	// on providers that do not report it.
	ManagedInstances []cloudinstances.GroupInstanceStatus `json:"managedInstances,omitempty"`
	// Failures are the provisioning failures the cloud provider attributes to
	// the group. An empty group with failures here is the signature of a group
	// that could not launch, e.g. InsufficientInstanceCapacity.
	Failures []cloudinstances.GroupFailure `json:"failures,omitempty"`
}

// CloudInstanceDump is the cloud provider's view of one instance in a group.
type CloudInstanceDump struct {
	ID                string    `json:"id"`
	MachineType       string    `json:"machineType,omitempty"`
	Status            string    `json:"status,omitempty"`
	State             string    `json:"state,omitempty"`
	Roles             []string  `json:"roles,omitempty"`
	PrivateIP         string    `json:"privateIP,omitempty"`
	ExternalIP        string    `json:"externalIP,omitempty"`
	CreationTimestamp time.Time `json:"creationTimestamp,omitzero"`
	// Node is the name of the Kubernetes node the instance registered as, empty
	// if it never joined the cluster.
	Node string `json:"node,omitempty"`
}

// DumpCloudInstanceGroups writes the cloud provider's view of every instance
// group, including any provisioning errors the provider reports, to
// CloudInstanceGroupsFileName under artifactsDir.
//
// When a group launches no instances the API server never comes up, so the
// cluster itself holds no record of what went wrong. This file is the only
// place a failed run can attribute that to a citable provider error rather than
// to "no instances launched".
func DumpCloudInstanceGroups(ctx context.Context, cloud fi.Cloud, cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup, nodes []v1.Node, artifactsDir string) error {
	warnUnmatched := false
	cloudGroups, err := cloud.GetCloudGroups(cluster, instanceGroups, warnUnmatched, nodes)
	if err != nil {
		return err
	}

	failureReporter, _ := cloud.(cloudinstances.GroupFailureReporter)
	statusReporter, _ := cloud.(cloudinstances.GroupInstanceStatusReporter)

	dumps := make([]CloudInstanceGroupDump, 0, len(cloudGroups))
	for _, cloudGroup := range cloudGroups {
		d := CloudInstanceGroupDump{
			Name:       cloudGroup.HumanName,
			MinSize:    cloudGroup.MinSize,
			TargetSize: cloudGroup.TargetSize,
			MaxSize:    cloudGroup.MaxSize,
		}
		if cloudGroup.InstanceGroup != nil {
			d.InstanceGroup = cloudGroup.InstanceGroup.Name
		}

		for _, member := range slices.Concat(cloudGroup.Ready, cloudGroup.NeedUpdate) {
			i := CloudInstanceDump{
				ID:                member.ID,
				MachineType:       member.MachineType,
				Status:            member.Status,
				State:             string(member.State),
				Roles:             member.Roles,
				PrivateIP:         member.PrivateIP,
				ExternalIP:        member.ExternalIP,
				CreationTimestamp: member.CreationTimestamp,
			}
			if member.Node != nil {
				i.Node = member.Node.Name
			}
			d.Instances = append(d.Instances, i)
		}
		sort.Slice(d.Instances, func(i, j int) bool {
			return d.Instances[i].ID < d.Instances[j].ID
		})

		if statusReporter != nil {
			statuses, err := statusReporter.GetGroupInstanceStatuses(ctx, cloudGroup)
			if err != nil {
				klog.Warningf("error getting managed instances for group %q: %v", cloudGroup.HumanName, err)
			} else {
				d.ManagedInstances = statuses
			}
		}

		if failureReporter != nil {
			failures, err := failureReporter.GetGroupFailures(ctx, cloudGroup)
			if err != nil {
				klog.Warningf("error getting cloud provider errors for group %q: %v", cloudGroup.HumanName, err)
			} else {
				d.Failures = failures
			}
		}

		dumps = append(dumps, d)
	}
	sort.Slice(dumps, func(i, j int) bool {
		return dumps[i].Name < dumps[j].Name
	})

	b, err := yaml.Marshal(dumps)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(artifactsDir, CloudInstanceGroupsFileName), b, 0o600)
}
