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

package openstack

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/gophercloud/gophercloud"
	os "github.com/gophercloud/gophercloud/openstack"
	cinder "github.com/gophercloud/gophercloud/openstack/blockstorage/v2/volumes"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/routers"
	sg "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/groups"
	sgr "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/rules"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

const TagNameEtcdClusterPrefix = "k8s.io/etcd/"
const TagNameRolePrefix = "k8s.io/role/"
const TagClusterName = "KubernetesCluster"

// readBackoff is the backoff strategy for openstack read retries.
var readBackoff = wait.Backoff{
	Duration: time.Second,
	Factor:   1.5,
	Jitter:   0.1,
	Steps:    4,
}

// writeBackoff is the backoff strategy for openstack write retries.
var writeBackoff = wait.Backoff{
	Duration: time.Second,
	Factor:   1.5,
	Jitter:   0.1,
	Steps:    5,
}

type OpenstackCloud interface {
	fi.Cloud

	// Region returns the region which cloud will run on
	Region() string

	// SetVolumeTags will set the tags for the Cinder volume
	SetVolumeTags(id string, tags map[string]string) error

	// GetCloudTags will return the tags attached on cloud
	GetCloudTags() map[string]string

	// ListVolumes will return the Cinder volumes which match the options
	ListVolumes(opt cinder.ListOpts) ([]cinder.Volume, error)

	// CreateVolume will create a new Cinder Volume
	CreateVolume(opt cinder.CreateOpts) (*cinder.Volume, error)

	//ListSecurityGroups will return the Neutron security groups which match the options
	ListSecurityGroups(opt sg.ListOpts) ([]sg.SecGroup, error)

	//CreateSecurityGroup will create a new Neutron security group
	CreateSecurityGroup(opt sg.CreateOpts) (*sg.SecGroup, error)

	//ListSecurityGroupRules will return the Neutron security group rules which match the options
	ListSecurityGroupRules(opt sgr.ListOpts) ([]sgr.SecGroupRule, error)

	//CreateSecurityGroupRule will create a new Neutron security group rule
	CreateSecurityGroupRule(opt sgr.CreateOpts) (*sgr.SecGroupRule, error)

	//ListNetworks will return the Neutron networks which match the options
	ListNetworks(opt networks.ListOptsBuilder) ([]networks.Network, error)

	//CreateNetwork will create a new Neutron network
	CreateNetwork(opt networks.CreateOptsBuilder) (*networks.Network, error)

	//ListRouters will return the Neutron routers which match the options
	ListRouters(opt routers.ListOpts) ([]routers.Router, error)

	//CreateRouter will create a new Neutron router
	CreateRouter(opt routers.CreateOptsBuilder) (*routers.Router, error)

	//ListSubnets will return the Neutron subnets which match the options
	ListSubnets(opt subnets.ListOptsBuilder) ([]subnets.Subnet, error)

	//CreateSubnet will create a new Neutron subnet
	CreateSubnet(opt subnets.CreateOptsBuilder) (*subnets.Subnet, error)

	//ListPorts will return the Neutron ports which match the options
	ListPorts(opt ports.ListOptsBuilder) ([]ports.Port, error)

	//CreateRouterInterface will create a new Neutron router interface
	CreateRouterInterface(routerID string, opt routers.AddInterfaceOptsBuilder) (*routers.InterfaceInfo, error)
}

type openstackCloud struct {
	cinderClient  *gophercloud.ServiceClient
	neutronClient *gophercloud.ServiceClient
	tags          map[string]string
	region        string
}

var _ fi.Cloud = &openstackCloud{}

func NewOpenstackCloud(tags map[string]string) (OpenstackCloud, error) {
	config := vfs.OpenstackConfig{}

	authOption, err := config.GetCredential()
	if err != nil {
		return nil, err
	}
	provider, err := os.AuthenticatedClient(authOption)
	if err != nil {
		return nil, fmt.Errorf("error building openstack authenticated client: %v", err)
	}

	endpointOpt, err := config.GetServiceConfig("Cinder")
	if err != nil {
		return nil, err
	}
	cinderClient, err := os.NewBlockStorageV2(provider, endpointOpt)
	if err != nil {
		return nil, fmt.Errorf("error building cinder client: %v", err)
	}

	endpointOpt, err = config.GetServiceConfig("Neutron")
	if err != nil {
		return nil, err
	}
	neutronClient, err := os.NewNetworkV2(provider, endpointOpt)
	if err != nil {
		return nil, fmt.Errorf("error building neutron client: %v", err)
	}
	region := endpointOpt.Region

	c := &openstackCloud{
		cinderClient:  cinderClient,
		neutronClient: neutronClient,
		tags:          tags,
		region:        region,
	}
	return c, nil
}

func (c *openstackCloud) Region() string {
	return c.region
}

func (c *openstackCloud) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderOpenstack
}

func (c *openstackCloud) DNS() (dnsprovider.Interface, error) {
	return nil, fmt.Errorf("openstackCloud::DNS not implemented")
}

func (c *openstackCloud) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return nil, fmt.Errorf("openstackCloud::FindVPCInfo not implemented")
}

func (c *openstackCloud) DeleteInstance(i *cloudinstances.CloudInstanceGroupMember) error {
	return fmt.Errorf("openstackCloud::DeleteInstance not implemented")
}

func (c *openstackCloud) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	return fmt.Errorf("openstackCloud::DeleteGroup not implemented")
}

func (c *openstackCloud) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	return nil, fmt.Errorf("openstackCloud::GetCloudGroups not implemented")
}

func (c *openstackCloud) SetVolumeTags(id string, tags map[string]string) error {
	if len(tags) == 0 {
		return nil
	}
	if id == "" {
		return fmt.Errorf("error setting tags to unknown volume")
	}
	glog.V(4).Infof("setting tags to cinder volume %q: %v", id, tags)

	opt := cinder.UpdateOpts{Metadata: tags}
	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		_, err := cinder.Update(c.cinderClient, id, opt).Extract()
		if err != nil {
			return false, fmt.Errorf("error setting tags to cinder volume %q: %v", id, err)
		}
		return true, nil
	})
	if err != nil {
		return err
	} else if done {
		return nil
	} else {
		return wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) GetCloudTags() map[string]string {
	return c.tags
}

func (c *openstackCloud) ListVolumes(opt cinder.ListOpts) ([]cinder.Volume, error) {
	var volumes []cinder.Volume

	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		allPages, err := cinder.List(c.cinderClient, opt).AllPages()
		if err != nil {
			return false, fmt.Errorf("error listing volumes %v: %v", opt, err)
		}

		vs, err := cinder.ExtractVolumes(allPages)
		if err != nil {
			return false, fmt.Errorf("error extracting volumes from pages: %v", err)
		}
		volumes = vs
		return true, nil
	})
	if err != nil {
		return volumes, err
	} else if done {
		return volumes, nil
	} else {
		return volumes, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) CreateVolume(opt cinder.CreateOpts) (*cinder.Volume, error) {
	var volume *cinder.Volume

	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		v, err := cinder.Create(c.cinderClient, opt).Extract()
		if err != nil {
			return false, fmt.Errorf("error creating volume %v: %v", opt, err)
		}
		volume = v
		return true, nil
	})
	if err != nil {
		return volume, err
	} else if done {
		return volume, nil
	} else {
		return volume, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) ListSecurityGroups(opt sg.ListOpts) ([]sg.SecGroup, error) {
	var groups []sg.SecGroup

	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		allPages, err := sg.List(c.neutronClient, opt).AllPages()
		if err != nil {
			return false, fmt.Errorf("error listing security groups %v: %v", opt, err)
		}

		gs, err := sg.ExtractGroups(allPages)
		if err != nil {
			return false, fmt.Errorf("error extracting security groups from pages: %v", err)
		}
		groups = gs
		return true, nil
	})
	if err != nil {
		return groups, err
	} else if done {
		return groups, nil
	} else {
		return groups, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) CreateSecurityGroup(opt sg.CreateOpts) (*sg.SecGroup, error) {
	var group *sg.SecGroup

	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		g, err := sg.Create(c.neutronClient, opt).Extract()
		if err != nil {
			return false, fmt.Errorf("error creating security group %v: %v", opt, err)
		}
		group = g
		return true, nil
	})
	if err != nil {
		return group, err
	} else if done {
		return group, nil
	} else {
		return group, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) ListSecurityGroupRules(opt sgr.ListOpts) ([]sgr.SecGroupRule, error) {
	var rules []sgr.SecGroupRule

	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		allPages, err := sgr.List(c.neutronClient, opt).AllPages()
		if err != nil {
			return false, fmt.Errorf("error listing security group rules %v: %v", opt, err)
		}

		rs, err := sgr.ExtractRules(allPages)
		if err != nil {
			return false, fmt.Errorf("error extracting security group rules from pages: %v", err)
		}
		rules = rs
		return true, nil
	})
	if err != nil {
		return rules, err
	} else if done {
		return rules, nil
	} else {
		return rules, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) CreateSecurityGroupRule(opt sgr.CreateOpts) (*sgr.SecGroupRule, error) {
	var rule *sgr.SecGroupRule

	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		r, err := sgr.Create(c.neutronClient, opt).Extract()
		if err != nil {
			return false, fmt.Errorf("error creating security group rule %v: %v", opt, err)
		}
		rule = r
		return true, nil
	})
	if err != nil {
		return rule, err
	} else if done {
		return rule, nil
	} else {
		return rule, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) ListNetworks(opt networks.ListOptsBuilder) ([]networks.Network, error) {
	var ns []networks.Network

	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		allPages, err := networks.List(c.neutronClient, opt).AllPages()
		if err != nil {
			return false, fmt.Errorf("error listing networks: %v", err)
		}

		r, err := networks.ExtractNetworks(allPages)
		if err != nil {
			return false, fmt.Errorf("error extracting networks from pages: %v", err)
		}
		ns = r
		return true, nil
	})
	if err != nil {
		return ns, err
	} else if done {
		return ns, nil
	} else {
		return ns, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) CreateNetwork(opt networks.CreateOptsBuilder) (*networks.Network, error) {
	var n *networks.Network

	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		r, err := networks.Create(c.neutronClient, opt).Extract()
		if err != nil {
			return false, fmt.Errorf("error creating network: %v", err)
		}
		n = r
		return true, nil
	})
	if err != nil {
		return n, err
	} else if done {
		return n, nil
	} else {
		return n, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) ListRouters(opt routers.ListOpts) ([]routers.Router, error) {
	var rs []routers.Router

	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		allPages, err := routers.List(c.neutronClient, opt).AllPages()
		if err != nil {
			return false, fmt.Errorf("error listing routers: %v", err)
		}

		r, err := routers.ExtractRouters(allPages)
		if err != nil {
			return false, fmt.Errorf("error extracting routers from pages: %v", err)
		}
		rs = r
		return true, nil
	})
	if err != nil {
		return rs, err
	} else if done {
		return rs, nil
	} else {
		return rs, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) CreateRouter(opt routers.CreateOptsBuilder) (*routers.Router, error) {
	var r *routers.Router

	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		v, err := routers.Create(c.neutronClient, opt).Extract()
		if err != nil {
			return false, fmt.Errorf("error creating router: %v", err)
		}
		r = v
		return true, nil
	})
	if err != nil {
		return r, err
	} else if done {
		return r, nil
	} else {
		return r, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) ListSubnets(opt subnets.ListOptsBuilder) ([]subnets.Subnet, error) {
	var s []subnets.Subnet

	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		allPages, err := subnets.List(c.neutronClient, opt).AllPages()
		if err != nil {
			return false, fmt.Errorf("error listing subnets: %v", err)
		}

		r, err := subnets.ExtractSubnets(allPages)
		if err != nil {
			return false, fmt.Errorf("error extracting subnets from pages: %v", err)
		}
		s = r
		return true, nil
	})
	if err != nil {
		return s, err
	} else if done {
		return s, nil
	} else {
		return s, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) CreateSubnet(opt subnets.CreateOptsBuilder) (*subnets.Subnet, error) {
	var s *subnets.Subnet

	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		v, err := subnets.Create(c.neutronClient, opt).Extract()
		if err != nil {
			return false, fmt.Errorf("error creating subnet: %v", err)
		}
		s = v
		return true, nil
	})
	if err != nil {
		return s, err
	} else if done {
		return s, nil
	} else {
		return s, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) ListPorts(opt ports.ListOptsBuilder) ([]ports.Port, error) {
	var p []ports.Port

	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		allPages, err := ports.List(c.neutronClient, opt).AllPages()
		if err != nil {
			return false, fmt.Errorf("error listing ports: %v", err)
		}

		r, err := ports.ExtractPorts(allPages)
		if err != nil {
			return false, fmt.Errorf("error extracting ports from pages: %v", err)
		}
		p = r
		return true, nil
	})
	if err != nil {
		return p, err
	} else if done {
		return p, nil
	} else {
		return p, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) CreateRouterInterface(routerID string, opt routers.AddInterfaceOptsBuilder) (*routers.InterfaceInfo, error) {
	var i *routers.InterfaceInfo

	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		v, err := routers.AddInterface(c.neutronClient, routerID, opt).Extract()
		if err != nil {
			return false, fmt.Errorf("error creating router interface: %v", err)
		}
		i = v
		return true, nil
	})
	if err != nil {
		return i, err
	} else if done {
		return i, nil
	} else {
		return i, wait.ErrWaitTimeout
	}
}
