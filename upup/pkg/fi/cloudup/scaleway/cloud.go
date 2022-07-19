package scaleway

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	kopsv "k8s.io/kops"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	dns "k8s.io/kops/dnsprovider/pkg/dnsprovider/providers/scaleway"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"

	account "github.com/scaleway/scaleway-sdk-go/api/account/v2alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	TagNameEtcdClusterPrefix = "k8s.io/etcd/"
	TagNameRolePrefix        = "k8s.io/role/"
	TagClusterName           = "KubernetesCluster"
	TagRoleMaster            = "master"
)

// ScwCloud exposes all the interfaces required to operate on Scaleway resources
type ScwCloud interface {
	fi.Cloud

	GetApiIngressStatus(cluster *kops.Cluster) ([]fi.ApiIngressStatus, error)
	FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error)

	Account() *account.API
	Instance() *instance.API
	LB() *lb.API
}

// static compile time check to validate ScwCloud's fi.Cloud Interface.
var _ fi.Cloud = &scwCloudImplementation{}

// scwCloudImplementation holds the scw.Client object to interact with Scaleway resources.
type scwCloudImplementation struct {
	client *scw.Client
	dns    dnsprovider.Interface
	region string
	tags   map[string]string

	accountAPI  *account.API
	instanceAPI *instance.API
	lbAPI       *lb.API
}

// NewScwCloud returns a Cloud, using the env vars SCW_ACCESS_KEY and SCW_SECRET_KEY
func NewScwCloud(region string, tags map[string]string) (ScwCloud, error) {
	// We could either build our client this way :

	//scwAccessKey := os.Getenv("SCW_ACCESS_KEY")
	//scwSecretKey := os.Getenv("SCW_SECRET_KEY")
	//if scwAccessKey == "" {
	//	if scwSecretKey == "" {
	//		return nil, errors.New("both SCW_ACCESS_KEY and SCW_SECRET_KEY are required")
	//	}
	//	return nil, errors.New("SCW_ACCESS_KEY is required")
	//}
	//if scwSecretKey == "" {
	//	return nil, errors.New("SCW_SECRET_KEY is required")
	//}
	//
	//scwClient, err := scw.NewClient(
	//	scw.WithAuth(scwAccessKey, scwSecretKey),
	//  scw.WithUserAgent("kubernetes-kops/"+kopsv.Version),
	//)
	//if err != nil {
	//	return nil, err
	//}

	// Or we could do it this way, code is shorter :

	// Use these env variables to set or overwrite profile values
	// SCW_ACCESS_KEY
	// SCW_SECRET_KEY
	// SCW_DEFAULT_PROJECT_ID or SCW_DEFAULT_ORGANIZATION_ID

	scwClient, err := scw.NewClient(
		scw.WithUserAgent("kubernetes-kops/"+kopsv.Version),
		scw.WithEnv(),
	)
	if err != nil {
		return nil, err
		// TODO: check if error is explicit enough when credentials are missing
	}

	return &scwCloudImplementation{
		client:      scwClient,
		dns:         dns.NewProvider(scwClient, ""), //TODO: fill in domain name
		region:      region,
		tags:        tags,
		accountAPI:  account.NewAPI(scwClient),
		instanceAPI: instance.NewAPI(scwClient),
		lbAPI:       lb.NewAPI(scwClient),
	}, nil
}

func (s *scwCloudImplementation) Account() *account.API {
	return s.accountAPI
}

func (s *scwCloudImplementation) Instance() *instance.API {
	return s.instanceAPI
}

func (s *scwCloudImplementation) LB() *lb.API {
	return s.lbAPI
}

func (s *scwCloudImplementation) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderScaleway
}

func (s *scwCloudImplementation) DNS() (dnsprovider.Interface, error) {
	provider, err := dnsprovider.GetDnsProvider(dns.ProviderName, nil)
	if err != nil {
		return nil, fmt.Errorf("error building DNS provider: %v", err)
	}
	return provider, nil
}

// FindVPCInfo is not implemented yet, it's only here to satisfy the fi.Cloud interface
func (s *scwCloudImplementation) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	klog.V(8).Info("scaleway cloud provider FindVPCInfo not implemented yet")
	return nil, fmt.Errorf("scaleway cloud provider does not support vpc at this time")
}

func (s *scwCloudImplementation) DeleteInstance(i *cloudinstances.CloudInstance) error {
	zone, id, err := parseZonedID(i.ID)
	if err != nil {
		return fmt.Errorf("delete instance %s: %v", i.ID, err)
	}

	// reach stopped state
	err = reachState(s.instanceAPI, zone, id, instance.ServerStateStopped)
	if is404Error(err) {
		klog.V(8).Info("delete instance %s: instance was already deleted", id)
		return nil
	}
	if err != nil {
		return fmt.Errorf("delete instance %s: error reaching stopped state: %v", id, err)
	}

	_, err = waitForInstanceServer(s.instanceAPI, zone, id)
	if err != nil {
		return fmt.Errorf("delete instance %s: error waiting for instance: %v", id, err)
	}

	err = s.instanceAPI.DeleteServer(&instance.DeleteServerRequest{
		Zone:     zone,
		ServerID: id,
	})
	if err != nil && !is404Error(err) {
		return fmt.Errorf("error deleting instance %s: %v", id, err)
	}

	_, err = waitForInstanceServer(s.instanceAPI, zone, id)
	if err != nil && !is404Error(err) {
		return fmt.Errorf("delete instance %s: error waiting for instance: %v", id, err)
	}

	return nil
}

func (s *scwCloudImplementation) DeregisterInstance(instance *cloudinstances.CloudInstance) error {
	//TODO implement me
	panic("implement me")
}

func (s *scwCloudImplementation) DeleteGroup(group *cloudinstances.CloudInstanceGroup) error {
	//TODO implement me
	panic("implement me")
}

func (s *scwCloudImplementation) DetachInstance(instance *cloudinstances.CloudInstance) error {
	//TODO implement me
	panic("implement me")
}

func (s *scwCloudImplementation) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	//TODO implement me
	panic("implement me")
}

func (s *scwCloudImplementation) Region() string {
	return s.region
}

func (s *scwCloudImplementation) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	//TODO implement me
	panic("implement me")
}

func (s *scwCloudImplementation) GetApiIngressStatus(cluster *kops.Cluster) ([]fi.ApiIngressStatus, error) {
	var ingresses []fi.ApiIngressStatus
	name := "api." + cluster.Name

	responseLoadBalancers, err := s.lbAPI.ListLBs(&lb.ListLBsRequest{
		Region: scw.Region(s.Region()),
		Name:   &name,
	})
	if err != nil {
		return nil, fmt.Errorf("error finding load-balancers: %v", err)
	}
	if len(responseLoadBalancers.LBs) == 0 {
		// QUESTION: Is it serious ? I should probably log it
		klog.V(8).Infof("could not find any load-balancers for cluster %s", cluster.Name)
		return nil, nil
		// TODO: Careful when getting this result, not an empty tab
	}
	if len(responseLoadBalancers.LBs) > 1 {
		klog.V(4).Infof("more than 1 load-balancer with the name %s was found", name)
	}

	address := responseLoadBalancers.LBs[0].IP[0].IPAddress
	ingresses = append(ingresses, fi.ApiIngressStatus{IP: address})

	return ingresses, nil
}
