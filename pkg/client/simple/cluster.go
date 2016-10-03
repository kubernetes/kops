package simple

import (
	"k8s.io/kops/upup/pkg/api"
	k8sapi "k8s.io/kubernetes/pkg/api"
)


// ClusterInterface has methods to work with Cluster resources.
type ClusterInterface interface {
	Create(*api.Cluster) (*api.Cluster, error)
	Update(*api.Cluster) (*api.Cluster, error)
	//UpdateStatus(*api.Cluster) (*api.Cluster, error)
	//Delete(name string, options *api.DeleteOptions) error
	//DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error
	Get(name string) (*api.Cluster, error)
	List(opts k8sapi.ListOptions) (*api.ClusterList, error)
	//Watch(opts k8sapi.ListOptions) (watch.Interface, error)
	//Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *api.Cluster, err error)
	//ClusterExpansion
}