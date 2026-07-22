/*
Copyright 2019 The Kubernetes Authors.

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
	"os"
	"strconv"
	"strings"

	"cloud.google.com/go/compute/metadata"
	compute "google.golang.org/api/compute/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/nodeidentity"
	"k8s.io/kops/pkg/nodeidentity/clusterapi"
	"k8s.io/kops/pkg/nodeidentity/clusterapi/capimanager"
	"k8s.io/kops/pkg/nodelabels"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

// MetadataKeyInstanceGroupName is the key for the metadata that specifies the instance group name
// This is used by the gce nodeidentifier to securely identify the node instancegroup
const MetadataKeyInstanceGroupName = "kops-k8s-io-instance-group-name"

// LabelKeyCAPIRoleName is the label key used by the Cluster API Provider GCP to indicate the role of the instance.
const LabelKeyCAPIRoleName = "capg-role"

// nodeIdentifier identifies a node from GCE
type nodeIdentifier struct {
	// computeService is the GCE client
	computeService *compute.Service

	// project is our GCE project; we require that instances be in this project
	project string

	// clusterName is the metadata.name of our cluster
	clusterName string

	// capiManager contains our CAPI support, if CAPI support is enabled
	capiManager *capimanager.Manager
}

// New creates and returns a nodeidentity.Identifier for Nodes running on GCE
func New(clusterName string, capiManager *capimanager.Manager) (nodeidentity.Identifier, error) {
	ctx := context.Background()

	computeService, err := compute.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("error building compute API client: %v", err)
	}

	// Project ID
	project := os.Getenv("GCP_PROJECT")
	if project != "" {
		klog.Infof("using project=%q from GCP_PROJECT env var", project)
	} else {
		project, err = metadata.ProjectID()
		if err != nil {
			return nil, fmt.Errorf("error reading project from GCE: %v", err)
		}
		project = strings.TrimSpace(project)
		if project == "" {
			return nil, fmt.Errorf("project metadata was empty")
		}
		klog.Infof("Found project=%q", project)
	}

	return &nodeIdentifier{
		computeService: computeService,
		project:        project,
		clusterName:    clusterName,
		capiManager:    capiManager,
	}, nil
}

// IdentifyNode queries GCE for the node identity information
func (i *nodeIdentifier) IdentifyNode(ctx context.Context, node *corev1.Node) (*nodeidentity.Info, error) {
	// log := klog.FromContext(ctx)

	providerID := node.Spec.ProviderID
	if providerID == "" {
		return nil, fmt.Errorf("providerID was not set for node %s", node.Name)
	}
	if !strings.HasPrefix(providerID, "gce://") {
		return nil, fmt.Errorf("providerID %q not recognized for node %s", providerID, node.Name)
	}

	tokens := strings.Split(strings.TrimPrefix(providerID, "gce://"), "/")
	if len(tokens) != 3 {
		return nil, fmt.Errorf("providerID %q not recognized for node %s", providerID, node.Name)
	}

	project := tokens[0]
	zone := tokens[1]
	instanceName := tokens[2]

	if project != i.project {
		return nil, fmt.Errorf("providerID %q did not match our project %q", providerID, i.project)
	}

	instance, err := i.getInstance(zone, instanceName)
	if err != nil {
		return nil, err
	}

	instanceStatus := instance.Status
	if instanceStatus != "RUNNING" {
		return nil, fmt.Errorf("found instance %q, but status is %q", instanceName, instanceStatus)
	}

	capgRole := instance.Labels[LabelKeyCAPIRoleName]

	var capiMachine *clusterapi.Machine

	if i.capiManager != nil && capgRole != "" {
		providerID := "gce://" + project + "/" + zone + "/" + instanceName

		m, err := i.capiManager.FindMachineByProviderID(ctx, providerID, gce.SafeClusterName(i.clusterName))
		if err != nil {
			return nil, fmt.Errorf("error finding Machine with providerID %q: %w", providerID, err)
		}
		capiMachine = m
	}

	var igName string
	if capiMachine == nil {
		instanceTemplate, err := GetInstanceTemplateForMIGMember(ctx, i.computeService, i.project, instance)
		if err != nil {
			return nil, err
		}

		igName = GetMetadataValue(instanceTemplate.Properties.Metadata, MetadataKeyInstanceGroupName)
		if igName == "" {
			return nil, fmt.Errorf("ig name not set on instance template %s", instanceTemplate.Name)
		}
	}

	info := &nodeidentity.Info{}
	// info.InstanceID TODO: InstanceID is only used by the provider?

	tagToRole := make(map[string]kops.InstanceGroupRole)
	for _, role := range kops.AllInstanceGroupRoles {
		tag := gce.TagForRole(i.clusterName, role)
		tagToRole[tag] = role
	}

	labels := make(map[string]string)
	for _, tag := range instance.Tags.Items {
		role, found := tagToRole[tag]
		if found {
			switch role {
			case kops.InstanceGroupRoleControlPlane:
				labels[nodelabels.RoleLabelControlPlane20] = ""
			case kops.InstanceGroupRoleNode:
				labels[nodelabels.RoleLabelNode16] = ""
			case kops.InstanceGroupRoleAPIServer:
				labels[nodelabels.RoleLabelAPIServer16] = ""
			default:
				klog.Warningf("unknown node role %q for server %q", role, instance.SelfLink)
			}
		}
	}
	if igName != "" {
		labels[kops.NodeLabelInstanceGroup] = igName
	}
	info.Labels = labels
	return info, nil
}

// getInstance queries GCE for the instance with the specified name, returning an error if not found
func (i *nodeIdentifier) getInstance(zone string, instanceName string) (*compute.Instance, error) {
	instance, err := i.computeService.Instances.Get(i.project, zone, instanceName).Do()
	if err != nil {
		return nil, fmt.Errorf("error fetching GCE instance: %w", err)
	}

	return instance, nil
}

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
	migName := lastComponent(createdBy)

	migMember, err := getManagedInstance(ctx, computeService, project, lastComponent(instance.Zone), migName, instance.Id)
	if err != nil {
		return nil, err
	}

	if migMember.Version == nil {
		return nil, fmt.Errorf("instance %s did not have Version set", instance.Name)
	}

	templateName := lastComponent(migMember.Version.InstanceTemplate)
	instanceTemplate, err := computeService.InstanceTemplates.Get(project, templateName).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("error fetching GCE instance group template %q: %v", templateName, err)
	}

	return instanceTemplate, nil
}

// getManagedInstance queries GCE for the instance from the MIG
func getManagedInstance(ctx context.Context, computeService *compute.Service, project string, zone string, migName string, instanceID uint64) (*compute.ManagedInstance, error) {
	var matches []*compute.ManagedInstance

	filter := "id=" + strconv.FormatUint(instanceID, 10)
	if err := computeService.InstanceGroupManagers.ListManagedInstances(project, zone, migName).Filter(filter).Pages(ctx, func(page *compute.InstanceGroupManagersListManagedInstancesResponse) error {
		// Post-filter... filters aren't implemented (b/27605549)
		for _, instance := range page.ManagedInstances {
			if instance.Id != instanceID {
				continue
			}
			matches = append(matches, instance)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("error fetching GCE managed instance group members for %q: %v", migName, err)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("instance %v not managed by mig %s", instanceID, migName)
	}
	if len(matches) > 1 {
		// Should be impossible - shows that filters / post-filters are not working
		return nil, fmt.Errorf("found multiple instances with id %v managed by mig %s", instanceID, migName)
	}

	return matches[0], nil
}

// lastComponent returns the last component of a URL, i.e. anything after the last slash
// If there is no slash, returns the whole string
func lastComponent(s string) string {
	lastSlash := strings.LastIndex(s, "/")
	if lastSlash != -1 {
		s = s[lastSlash+1:]
	}
	return s
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
