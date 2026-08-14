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
	"strings"

	"cloud.google.com/go/compute/metadata"
	compute "google.golang.org/api/compute/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

		capiCluster := types.NamespacedName{
			Namespace: metav1.NamespaceSystem,
			Name:      gce.SafeClusterName(i.clusterName),
		}
		m, err := i.capiManager.FindMachineByProviderID(ctx, providerID, capiCluster)
		if err != nil {
			return nil, fmt.Errorf("error finding Machine with providerID %q: %w", providerID, err)
		}
		capiMachine = m
	}

	var igName string
	if capiMachine == nil {
		instanceTemplate, err := gce.GetInstanceTemplateForMIGMember(ctx, i.computeService, i.project, instance)
		if err != nil {
			return nil, err
		}

		igName = gce.GetMetadataValue(instanceTemplate.Properties.Metadata, MetadataKeyInstanceGroupName)
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
