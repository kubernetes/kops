package simple

import (
	api "k8s.io/kops/pkg/apis/kops"
	k8sapi "k8s.io/kubernetes/pkg/api"
)


// FederationInterface has methods to work with Federation resources.
type FederationInterface interface {
	Create(*api.Federation) (*api.Federation, error)
	Update(*api.Federation) (*api.Federation, error)
	//UpdateStatus(*api.Federation) (*api.Federation, error)
	Delete(name string, options *k8sapi.DeleteOptions) error
	//DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error
	Get(name string) (*api.Federation, error)
	List(opts k8sapi.ListOptions) (*api.FederationList, error)
	//Watch(opts k8sapi.ListOptions) (watch.Interface, error)
	//Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *api.Federation, err error)
	//FederationExpansion
}