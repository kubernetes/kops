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
	"k8s.io/kops/upup/pkg/fi"
)

func (s *Server) getNodeConfig(ctx context.Context, req *nodeup.BootstrapRequest, identity *fi.VerifyResult) (*nodeup.NodeConfig, error) {
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

		b, err := p.ReadFile()
		if err != nil {
			return nil, fmt.Errorf("error loading cluster config %q: %w", p, err)
		}
		nodeConfig.ClusterFullConfig = string(b)
	}

	{
		p := s.configBase.Join("instancegroup", instanceGroupName)

		b, err := p.ReadFile()
		if err != nil {
			return nil, fmt.Errorf("error loading InstanceGroup %q: %v", p, err)
		}
		nodeConfig.InstanceGroupConfig = string(b)
	}

	// We populate some certificates that we know the node will need.
	for _, name := range []string{"ca"} {
		cert, _, _, err := s.keystore.FindKeypair(name)
		if err != nil {
			return nil, fmt.Errorf("error getting certificate %q: %w", name, err)
		}

		if cert == nil {
			return nil, fmt.Errorf("certificate %q not found", name)
		}

		certData, err := cert.AsString()
		if err != nil {
			return nil, fmt.Errorf("error marshalling certificate %q: %w", name, err)
		}

		nodeConfig.Certificates = append(nodeConfig.Certificates, &nodeup.NodeConfigCertificate{
			Name: name,
			Cert: certData,
		})
	}

	return nodeConfig, nil
}
