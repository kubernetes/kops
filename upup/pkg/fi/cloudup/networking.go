package cloudup

import (
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/api"
)

func usesCNI(c *api.Cluster) bool {
	networkConfig := c.Spec.Networking
	if networkConfig == nil || networkConfig.Classic != nil {
		// classic
		return false
	}

	if networkConfig.Kubenet != nil {
		// kubenet
		return true
	}

	if networkConfig.External != nil {
		// external: assume uses CNI
		return true
	}

	// Assume other modes also use CNI
	glog.Warningf("Unknown networking mode configured")
	return true
}
