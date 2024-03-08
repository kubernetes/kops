/*
Copyright 2020 The Kubernetes Authors.

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

package server

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/bootstrap"
	"k8s.io/kops/pkg/commands"
	"k8s.io/kops/pkg/nodeidentity/clusterapi"
)

func (s *Server) getNodeConfig(ctx context.Context, req *nodeup.BootstrapRequest, identity *bootstrap.VerifyResult) (*nodeup.NodeConfig, error) {
	log := klog.FromContext(ctx)

	if identity == nil {
		return nil, fmt.Errorf("node identity is required")
	}

	log.Info("getting node config", "req", req, "identity", identity)

	instanceGroupName := identity.InstanceGroupName
	if instanceGroupName == "" && identity.CAPIMachine == nil {
		return nil, fmt.Errorf("did not find owner for node %q", identity.NodeName)
	}

	var nodeConfig *nodeup.NodeConfig

	configBuilder := &commands.ConfigBuilder{
		Clientset:   s.clientset,
		ClusterName: s.opt.ClusterName,
	}

	if identity.CAPIMachine != nil && instanceGroupName == "" {
		// We have a CAPI Machine (but no instance group)
		instanceGroup, err := s.buildInstanceGroupFromCAPI(ctx, identity.CAPIMachine)
		if err != nil {
			return nil, fmt.Errorf("error building InstanceGroup from CAPI Machine: %w", err)
		}
		log.Info("built InstanceGroup from CAPI Machine", "instanceGroup", instanceGroup)
		configBuilder.InstanceGroup = instanceGroup
	} else if s.opt.Cloud == "metal" {
		configBuilder.InstanceGroupName = instanceGroupName
	} else {
		// Note: For now, we're assuming there is only a single cluster, and it is ours.
		// We therefore use the configured base path

		p := s.configBase.Join("igconfig", "node", instanceGroupName, "nodeupconfig.yaml")

		b, err := p.ReadFile(ctx)
		if err != nil {
			return nil, fmt.Errorf("error loading NodeupConfig %q: %v", p, err)
		}
		nodeConfig = &nodeup.NodeConfig{}
		nodeConfig.NodeupConfig = string(b)
	}

	if nodeConfig == nil {
		bootstrapData, err := configBuilder.GetBootstrapData(ctx)
		if err != nil {
			return nil, fmt.Errorf("building nodeConfig for instanceGroup: %w", err)
		}
		nodeupConfig, err := json.Marshal(bootstrapData.NodeupConfig)
		if err != nil {
			return nil, fmt.Errorf("marshalling nodeupConfig: %w", err)
		}
		nodeConfig = &nodeup.NodeConfig{}
		nodeConfig.NodeupConfig = string(nodeupConfig)
	}

	{
		secretIDs := []string{
			"dockerconfig",
		}
		nodeConfig.NodeSecrets = make(map[string][]byte)
		for _, id := range secretIDs {
			secret, err := s.secretStore.FindSecret(id)
			if err != nil {
				return nil, fmt.Errorf("error loading secret %q: %w", id, err)
			}
			if secret != nil && secret.Data != nil {
				nodeConfig.NodeSecrets[id] = secret.Data
			}
		}
	}

	return nodeConfig, nil
}

// buildInstanceGroupFromCAPI builds an InstanceGroup from a CAPI Machine, for building bootstrap data.
// It builds a minimal instanceGroup, because many fields (e.g. image, machineType, minSize, maxSize)
// are not relevant for building the bootstrap data.
func (s *Server) buildInstanceGroupFromCAPI(ctx context.Context, capiMachine *clusterapi.Machine) (*kops.InstanceGroup, error) {
	log := klog.FromContext(ctx)

	capiDeploymentName := capiMachine.GetDeploymentName()
	if capiDeploymentName == "" {
		return nil, fmt.Errorf("CAPI Machine is missing cluster.x-k8s.io/deployment-name label")
	}
	failureDomain := capiMachine.GetFailureDomain()
	if failureDomain == "" {
		return nil, fmt.Errorf("CAPI Machine is missing spec.failureDomain")
	}

	ig := &kops.InstanceGroup{}
	ig.Labels = map[string]string{
		// kops.LabelClusterName: cluster.Name, // Should not matter
	}
	ig.Name = capiDeploymentName

	// "maxSize": 1, // Should not matter
	// "minSize": 1, // Should not matter
	// "image": "", // Should not matter
	// "machineType": "", // Should not matter
	// "subnets": // Should not matter
	ig.Spec.Zones = []string{failureDomain}
	ig.Spec.Role = "Node" // TODO: Support other roles?

	log.Info("built InstanceGroup from CAPI Machine", "instanceGroup", ig)
	return ig, nil
}
