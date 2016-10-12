package v1alpha1

// NetworkingSpec allows selection and configuration of a networking plugin
type NetworkingSpec struct {
	Classic  *ClassicNetworkingSpec  `json:"classic,omitempty"`
	Kubenet  *KubenetNetworkingSpec  `json:"kubenet,omitempty"`
	External *ExternalNetworkingSpec `json:"external,omitempty"`
}

// ClassicNetworkingSpec is the specification of classic networking mode, integrated into kubernetes
type ClassicNetworkingSpec struct {
}

// KubenetNetworkingSpec is the specification for kubenet networking, largely integrated but intended to replace classic
type KubenetNetworkingSpec struct {
}

// ExternalNetworkingSpec is the specification for networking that is implemented by a Daemonset
// Networking is not managed by kops - we can create options here that directly configure e.g. weave
// but this is useful for arbitrary network modes or for modes that don't need additional configuration.
type ExternalNetworkingSpec struct {
}
