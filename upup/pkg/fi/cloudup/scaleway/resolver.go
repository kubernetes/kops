package scaleway

import (
	"context"
	"fmt"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"k8s.io/klog/v2"
	//"k8s.io/kops/protokube/pkg/gossip"
)

type Resolver struct{}

//var _ gossip.SeedProvider = &Resolver{}

//func (r *Resolver) GetSeeds() ([]string, error) {
//	var seeds []string
//	klog.Infof("^^^^^^^^^^^^^^^^^^^^^^^ GET SEEDS ^^^^^^^^^^^^^^^^^^^^^^^^^^")
//
//	instances, err := r.findInstances(context.TODO(), []string{TagClusterName + "=" + r.clusterName})
//	if err != nil {
//		return nil, fmt.Errorf("could not find instances: %w", err)
//	}
//
//	for _, instance := range instances {
//		if instance.PrivateIP == nil || *instance.PrivateIP == "" {
//			klog.Warningf("failed to find private ip of the instance %s(%s)", instance.Name, instance.ID)
//			continue
//		}
//		klog.V(4).Infof("Appending gossip seed %s(%s): %q", instance.Name, instance.ID, *instance.PrivateIP)
//		seeds = append(seeds, *instance.PrivateIP)
//	}
//
//	klog.V(4).Infof("Get seeds function done now")
//	return seeds, nil
//}

//func (r *Resolver) findInstances(ctx context.Context, tags []string) ([]*instance.Server, error) {
//	//func findInstances(ctx context.Context, scwClient *scw.Client, tags []string) ([]*instance.Server, error) {
//	instanceAPI := instance.NewAPI(r.scwClient)
//	klog.Infof("^^^^^^^^^^^^^^^^^^^^^^^ FIND INSTANCES ^^^^^^^^^^^^^^^^^^^^^^^^^^")
//	zone, ok := r.scwClient.GetDefaultZone()
//	if !ok {
//		return nil, fmt.Errorf("could not determine default zone from client")
//	}
//	servers, err := instanceAPI.ListServers(&instance.ListServersRequest{
//		Zone: zone,
//		Tags: tags,
//	}, scw.WithAllPages(), scw.WithContext(ctx))
//	if err != nil {
//		return nil, fmt.Errorf("failed to get matching servers: %w", err)
//	}
//	return servers.Servers, nil
//}

func NewResolver() (*Resolver, error) {
	//func NewResolver(scwClient *scw.Client, clusterName string) (*Resolver, error) {
	klog.Infof("^^^^^^^^^^^^^^^^^^^^^^^ NEW RESOLVER ^^^^^^^^^^^^^^^^^^^^^^^^^^")

	//profile := ctx.Value(ProfileContextKey)
	//
	//scwClient, err := scw.NewClient(scw.WithProfile(profile))
	////profile, err := scaleway.CreateValidScalewayProfile()
	//if err != nil {
	//	return nil, err
	//}
	//for i, server := range bootConfig.ConfigServer.Servers {
	//	klog.Infof("server[[%d]] = %s", i, server)
	//}
	//
	//zone := "fr-par-1"
	//
	////zone := scaleway.ParseZoneFromClusterSpec()
	//
	//scwClient, err := scw.NewClient(
	//	scw.WithProfile(profile),
	//	scw.WithUserAgent(scaleway.KopsUserAgentPrefix+kopsv.Version),
	//	scw.WithDefaultZone(scw.Zone(zone)),
	//	scw.WithDefaultRegion(scw.Region(region)),
	//)
	//if err != nil {
	//	return nil, fmt.Errorf("creating client for resolver: %w", err)
	//}
	return &Resolver{
		//scwClient:   scwClient,
		//clusterName: clusterName,
	}, nil
}

func (r *Resolver) Resolve(ctx context.Context, name string) ([]string, error) {
	klog.Infof("^^^^^^^^^^^^^^^^^^^^^^^ RESOLVE ^^^^^^^^^^^^^^^^^^^^^^^^^^")
	//var records []string
	klog.Infof("trying to resolve %q using Scaleway resolver", name)

	metadataAPI := instance.NewMetadataAPI()
	lbIP, err := metadataAPI.GetUserData("")
	if err != nil {
		return nil, fmt.Errorf("could not get load-balancer's IP from instance user-data: %w", err)
	}
	return []string{string(lbIP)}, nil

	// We fetch the IPs of the control-planes instances
	//tagsToLookFor := []string{
	//	TagClusterName + "=" + r.clusterName,
	//	TagNameRolePrefix + "=" + TagRoleControlPlane,
	//}
	//instances, err := findInstances(ctx, r.scwClient, tagsToLookFor)
	//if err != nil {
	//	return nil, fmt.Errorf("could not find instances: %w", err)
	//}
	//
	//for _, instance := range instances {
	//	if instance.PrivateIP == nil || *instance.PrivateIP == "" {
	//		klog.Warningf("failed to find private IP of the instance %s(%s)", instance.Name, instance.ID)
	//		continue
	//	}
	//	klog.V(4).Infof("Appending private IP [%s] of instance %s(%s)", *instance.PrivateIP, instance.Name, instance.ID)
	//	records = append(records, *instance.PrivateIP)
	//}
	//
	//// We fetch the IPs of the load-balancers
	//zone, ok := r.scwClient.GetDefaultZone()
	//if !ok {
	//	return nil, fmt.Errorf("could not determine default zone from client")
	//}
	//
	//lbAPI := lb.NewZonedAPI(r.scwClient)
	//lbs, err := lbAPI.ListLBs(&lb.ZonedAPIListLBsRequest{
	//	Zone: zone,
	//	Name: &r.clusterName,
	//}, scw.WithAllPages(), scw.WithContext(ctx))
	//if err != nil {
	//	return nil, fmt.Errorf("could not list load-balancers: %w", err)
	//}
	//for _, lb := range lbs.LBs {
	//	for _, ip := range lb.IP {
	//		klog.V(4).Infof("Appending IP [%s] of load-balancer %s(%s)", ip.IPAddress, lb.Name, lb.ID)
	//		records = append(records, ip.IPAddress)
	//	}
	//}

	//return records, nil
}
