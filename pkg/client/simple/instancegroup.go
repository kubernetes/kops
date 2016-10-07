package simple

import (
	"k8s.io/kops/upup/pkg/api"
	k8sapi "k8s.io/kubernetes/pkg/api"
)


// InstanceGroupInterface has methods to work with InstanceGroup resources.
type InstanceGroupInterface interface {
	Create(*api.InstanceGroup) (*api.InstanceGroup, error)
	Update(*api.InstanceGroup) (*api.InstanceGroup, error)
	//UpdateStatus(*api.InstanceGroup) (*api.InstanceGroup, error)
	Delete(name string, options *k8sapi.DeleteOptions) error
	//DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error
	Get(name string) (*api.InstanceGroup, error)
	List(opts k8sapi.ListOptions) (*api.InstanceGroupList, error)
	//Watch(opts k8sapi.ListOptions) (watch.Interface, error)
	//Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *api.InstanceGroup, err error)
	//InstanceGroupExpansion
}