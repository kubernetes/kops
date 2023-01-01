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
	"fmt"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/bootstrap"
)

func (s *Server) getNodeConfig(ctx context.Context, req *nodeup.BootstrapRequest, identity *bootstrap.VerifyResult) (*nodeup.NodeConfig, error) {
	klog.Infof("getting node config for %+v", req)

	instanceGroupName := identity.InstanceGroupName
	if instanceGroupName == "" {
		return nil, fmt.Errorf("did not find InstanceGroup for node %q", identity.NodeName)
	}

	nodeConfig := &nodeup.NodeConfig{}

	// Note: For now, we're assuming there is only a single cluster, and it is ours.
	// We therefore use the configured base path

	// Today we load the full cluster config from the state store (e.g. S3) every time
	// TODO: we should generate it on the fly (to allow for cluster reconfiguration)
	{
		p := s.configBase.Join(registry.PathClusterCompleted)

		b, err := p.ReadFile(ctx)
		if err != nil {
			return nil, fmt.Errorf("error loading cluster config %q: %w", p, err)
		}
		nodeConfig.ClusterFullConfig = string(b)
	}

	{
		p := s.configBase.Join("igconfig", "node", instanceGroupName, "nodeupconfig.yaml")

		b, err := p.ReadFile(ctx)
		if err != nil {
			return nil, fmt.Errorf("error loading NodeupConfig %q: %v", p, err)
		}
		nodeConfig.NodeupConfig = string(b)
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
