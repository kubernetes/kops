package cloudup

import (
	"k8s.io/kops/upup/pkg/api"
	"github.com/golang/glog"
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

	// Assume other modes also use CNI
	glog.Warningf("Unknown networking mode configured")
	return true
}
