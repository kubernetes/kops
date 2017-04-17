package main

import (
	"fmt"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/status"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

// cloudDiscoveryStatusStore implements status.Store by inspecting cloud objects.
// Likely temporary until we validate our status usage
type cloudDiscoveryStatusStore struct {
}

var _ status.Store = &cloudDiscoveryStatusStore{}

func (s *cloudDiscoveryStatusStore) GetApiIngressStatus(cluster *kops.Cluster) ([]status.ApiIngressStatus, error) {
	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return nil, err
	}

	if gceCloud, ok := cloud.(*gce.GCECloud); ok {
		return gceCloud.GetApiIngressStatus(cluster)
	}

	return nil, fmt.Errorf("API Ingress Status not implemented for %T", cloud)
}
