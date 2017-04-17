package status

import "k8s.io/kops/pkg/apis/kops"

// Store abstracts the key status functions; and lets us introduce status gradually
type Store interface {
	GetApiIngressStatus(cluster *kops.Cluster) ([]ApiIngressStatus, error)
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
