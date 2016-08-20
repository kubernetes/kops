package api

// NetworkingSpec allows selection and configuration of a networking plugin
type NetworkingSpec struct {
	Classic *ClassicNetworkingSpec `json:"classic,omitempty"`
	Kubenet *KubenetNetworkingSpec `json:"kubenet,omitempty"`
}

// ClassicNetworkingSpec is the specification of classic networking mode, integrated into kubernetes
type ClassicNetworkingSpec struct {
}

// KubenetNetworkingSpec is the specification for kubenet networking, largely integrated but intended to replace classic
type KubenetNetworkingSpec struct {
}
