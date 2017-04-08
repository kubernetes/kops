/*
Copyright 2017 The Kubernetes Authors.

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

package kops

import "github.com/golang/glog"

// StatusStore abstracts the key status functions; and lets us introduce status gradually
type StatusStore interface {
	GetApiIngressStatus(cluster *Cluster) ([]ApiIngressStatus, error)
}

// ApiIngress represents the status of an ingress point:
// traffic intended for the service should be sent to an ingress point.
type ApiIngressStatus struct {
	// IP is set for load-balancer ingress points that are IP based
	// (typically GCE or OpenStack load-balancers)
	// +optional
	IP string `json:"ip,omitempty" protobuf:"bytes,1,opt,name=ip"`

	// Hostname is set for load-balancer ingress points that are DNS based
	// (typically AWS load-balancers)
	// +optional
	Hostname string `json:"hostname,omitempty" protobuf:"bytes,2,opt,name=hostname"`
}

// NoopStatusStore is a stub implementation that returns empty status
// It is a temporary hackaround while we introduce status
type NoopStatusStore struct {
}

var _ StatusStore = &NoopStatusStore{}

func (s *NoopStatusStore) GetApiIngressStatus(cluster *Cluster) ([]ApiIngressStatus, error) {
	glog.Warningf("GetApiIngressStatus called on NoopStore")
	return nil, nil
}
