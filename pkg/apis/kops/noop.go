package status

import (
	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops"
)

// NoopStore is a stub implementation that returns empty status
// It is a temporary hackaround while we introduce status
type NoopStore struct {
}

var _ Store = &NoopStore{}

func (s *NoopStore) GetApiIngressStatus(cluster *kops.Cluster) ([]ApiIngressStatus, error) {
	glog.Warningf("GetApiIngressStatus called on NoopStore")
	return nil, nil
}
