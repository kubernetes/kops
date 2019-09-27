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
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v0.beta"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"k8s.io/kops/pkg/nodeidentity"
)

// MetadataKeyInstanceGroupName is the key for the metadata that specifies the instance group name
// This is used by the gce nodeidentifier to securely identify the node instancegroup
const MetadataKeyInstanceGroupName = "kops-k8s-io-instance-group-name"

// nodeIdentifier identifies a node from GCE
type nodeIdentifier struct {
	// computeService is the GCE client
	computeService *compute.Service

	// project is our GCE project; we require that instances be in this project
	project string
}

// New creates and returns a nodeidentity.Identifier for Nodes running on GCE
func New() (nodeidentity.Identifier, error) {
	ctx := context.Background()

	client, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		return nil, fmt.Errorf("error building google API client: %v", err)
	}

	computeService, err := compute.New(client)
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
	}, nil
}

// IdentifyNode queries GCE for the node identity information
func (i *nodeIdentifier) IdentifyNode(ctx context.Context, node *corev1.Node) (*nodeidentity.Info, error) {
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

	// The metadata itself is potentially mutable from the instance
	// We instead look at the MIG configuration
	createdBy := getMetadataValue(instance.Metadata, "created-by")
	if createdBy == "" {
		return nil, fmt.Errorf("instance %q did not have created-by metadata label set", instanceName)
	}

	// We need to double-check the MIG configuration, in case created-by was changed
	migName := lastComponent(createdBy)

	mig, err := i.getMIG(zone, migName)
	if err != nil {
		return nil, err
	}

	// We now double check that the instance is indeed managed by the MIG
	// this can't be spoofed without GCE API access
	migMember, err := i.getManagedInstance(mig, instance.Id)
	if err != nil {
		return nil, err
	}

	if migMember.Version == nil {
		return nil, fmt.Errorf("instance %s did not have Version set", instance.Name)
	}

	instanceTemplate, err := i.getInstanceTemplate(lastComponent(migMember.Version.InstanceTemplate))
	if err != nil {
		return nil, err
	}

	igName := getMetadataValue(instanceTemplate.Properties.Metadata, MetadataKeyInstanceGroupName)
	if igName == "" {
		return nil, fmt.Errorf("ig name not set on instance template %s", instanceTemplate.Name)
	}

	info := &nodeidentity.Info{}
	info.InstanceGroup = igName
	return info, nil
}

// getInstance queries GCE for the instance with the specified name, returning an error if not found
func (i *nodeIdentifier) getInstance(zone string, instanceName string) (*compute.Instance, error) {
	instance, err := i.computeService.Instances.Get(i.project, zone, instanceName).Do()
	if err != nil {
		return nil, fmt.Errorf("error fetching GCE instance: %v", err)
	}

	return instance, nil
}

// getInstanceTemplate queries GCE for the IG Template with the specified name, returning an error if not found
func (i *nodeIdentifier) getInstanceTemplate(name string) (*compute.InstanceTemplate, error) {
	t, err := i.computeService.InstanceTemplates.Get(i.project, name).Do()
	if err != nil {
		return nil, fmt.Errorf("error fetching GCE instance group template %q: %v", name, err)
	}

	return t, nil
}

// getMIG queries GCE for the MIG with the specified name, returning an error if not found
func (i *nodeIdentifier) getMIG(zone string, migName string) (*compute.InstanceGroupManager, error) {
	mig, err := i.computeService.InstanceGroupManagers.Get(i.project, zone, migName).Do()
	if err != nil {
		return nil, fmt.Errorf("error fetching GCE managed instance group %q: %v", migName, err)
	}

	return mig, nil
}

// getMIGMember queries GCE for the instance from the MIG
func (i *nodeIdentifier) getManagedInstance(mig *compute.InstanceGroupManager, instanceID uint64) (*compute.ManagedInstance, error) {
	filter := "id=" + strconv.FormatUint(instanceID, 10)
	zone := lastComponent(mig.Zone)
	instances, err := i.computeService.InstanceGroupManagers.ListManagedInstances(i.project, zone, mig.Name).Filter(filter).Do()
	if err != nil {
		return nil, fmt.Errorf("error fetching GCE managed instance group members for %q: %v", mig.Name, err)
	}

	// Post-filter... seeing some odd results
	var matches []*compute.ManagedInstance
	for _, instance := range instances.ManagedInstances {
		if instance.Id != instanceID {
			// Should be impossible - shows that filters are not working
			klog.Warningf("found instances with mismatched id %v", instance.Id)
			continue
		}
		matches = append(matches, instance)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("instance %v not managed by mig %s", instanceID, mig.Name)
	}
	if len(matches) > 1 {
		// Should be impossible - shows that filters are not working
		return nil, fmt.Errorf("found multiple instances with id %v managed by mig %s", instanceID, mig.Name)
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

func getMetadataValue(metadata *compute.Metadata, key string) string {
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
