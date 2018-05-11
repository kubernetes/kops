package spotinst

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/spotinst/spotinst-sdk-go/service/elastigroup/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/credentials"
	"github.com/spotinst/spotinst-sdk-go/spotinst/log"
	"k8s.io/api/core/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

func NewCredentials() *credentials.Credentials {
	return credentials.NewChainCredentials(
		new(credentials.EnvProvider),
		new(credentials.FileProvider),
	)
}

func LoadCredentials() (credentials.Value, error) {
	var creds credentials.Value
	var err error

	chain := NewCredentials()
	creds, err = chain.Get()
	if err != nil {
		return creds, fmt.Errorf("spotinst: unable to find Spotinst credentials: %s", err)
	}

	return creds, nil
}

func GuessCloudFromClusterSpec(spec *kops.ClusterSpec) kops.CloudProviderID {
	var cloudProviderID kops.CloudProviderID
	for _, subnet := range spec.Subnets {
		id, known := fi.GuessCloudForZone(subnet.Zone)
		if known {
			glog.V(2).Infof("Inferred cloud=%s from zone %q", id, subnet.Zone)
			cloudProviderID = kops.CloudProviderID(id)
			break
		}
	}
	return cloudProviderID
}

func newStdLogger() log.Logger {
	return log.LoggerFunc(func(format string, args ...interface{}) {
		glog.V(2).Infof(format, args...)
	})
}

func buildCloud(cloudProviderID kops.CloudProviderID, cluster *kops.Cluster) (fi.Cloud, error) {
	var cloud fi.Cloud
	var err error

	switch cloudProviderID {
	case kops.CloudProviderAWS:
		cloud, err = buildCloudAWS(cluster)
		if err != nil {
			return nil, err
		}
	case kops.CloudProviderGCE:
		cloud, err = buildCloudGCE(cluster)
		if err != nil {
			return nil, err
		}
	}

	return cloud, nil
}

func buildCloudAWS(cluster *kops.Cluster) (fi.Cloud, error) {
	region, err := awsup.FindRegion(cluster)
	if err != nil {
		return nil, err
	}

	tags := map[string]string{
		awsup.TagClusterName: cluster.ObjectMeta.Name,
	}

	cloud, err := awsup.NewAWSCloud(region, tags)
	if err != nil {
		return nil, err
	}

	return cloud, nil
}

func buildCloudGCE(cluster *kops.Cluster) (fi.Cloud, error) {
	var region string
	var project string

	for _, subnet := range cluster.Spec.Subnets {
		tokens := strings.Split(subnet.Zone, "-")
		if len(tokens) <= 2 {
			return nil, fmt.Errorf("spotinst: invalid GCE zone: %v", subnet.Zone)
		}
		zoneRegion := tokens[0] + "-" + tokens[1]
		if region != "" && zoneRegion != region {
			return nil, fmt.Errorf("spotinst: clusters cannot span multiple regions (found zone %q, but region is %q)", subnet.Zone, region)
		}
		region = zoneRegion
	}

	project = cluster.Spec.Project
	if project == "" {
		return nil, fmt.Errorf("spotinst: project is required for GCE")
	}

	labels := map[string]string{
		gce.GceLabelNameKubernetesCluster: gce.SafeClusterName(cluster.ObjectMeta.Name),
	}

	cloud, err := gce.NewGCECloud(region, project, labels)
	if err != nil {
		return nil, err
	}

	return cloud, nil
}

func getGroupNameByRole(cluster *kops.Cluster, ig *kops.InstanceGroup) string {
	var groupName string

	switch ig.Spec.Role {
	case kops.InstanceGroupRoleMaster:
		groupName = ig.ObjectMeta.Name + ".masters." + cluster.ObjectMeta.Name
	case kops.InstanceGroupRoleNode:
		groupName = ig.ObjectMeta.Name + "." + cluster.ObjectMeta.Name
	case kops.InstanceGroupRoleBastion:
		groupName = ig.ObjectMeta.Name + "." + cluster.ObjectMeta.Name
	default:
		glog.Warningf("Ignoring InstanceGroup of unknown role %q", ig.Spec.Role)
	}

	return groupName
}

func buildInstanceGroup(ig *kops.InstanceGroup, group *aws.Group, instances []*aws.Instance, nodeMap map[string]*v1.Node) (*cloudinstances.CloudInstanceGroup, error) {
	currentGroupName := spotinst.StringValue(group.Name)
	newGroupName := fmt.Sprintf("%s:%d", spotinst.StringValue(group.Name), time.Now().Nanosecond())

	cg := &cloudinstances.CloudInstanceGroup{
		HumanName:     spotinst.StringValue(group.Name),
		InstanceGroup: ig,
		MinSize:       spotinst.IntValue(group.Capacity.Maximum),
		MaxSize:       spotinst.IntValue(group.Capacity.Maximum),
		Raw:           group,
	}

	for _, instance := range instances {
		instanceID := spotinst.StringValue(instance.ID)
		if instanceID == "" {
			glog.Warningf("ignoring instance with no instance id: %s", instance)
			continue
		}
		err := cg.NewCloudInstanceGroupMember(instanceID, newGroupName, currentGroupName, nodeMap)
		if err != nil {
			return nil, fmt.Errorf("error creating cloud instance group member: %v", err)
		}
	}

	return cg, nil
}

func findEtcdStatusAWS(cloud awsup.AWSCloud, cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	status, err := cloud.FindClusterStatus(cluster)
	if err != nil {
		return nil, err
	}
	return status, nil
}

func findEtcdStatusGCE(cloud gce.GCECloud, cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	status, err := cloud.FindClusterStatus(cluster)
	if err != nil {
		return nil, err
	}
	return status, nil
}

func getApiIngressStatusAWS(cloud awsup.AWSCloud, cluster *kops.Cluster) ([]kops.ApiIngressStatus, error) {
	name := "api." + cluster.Name
	lb, err := awstasks.FindLoadBalancerByNameTag(cloud, name)
	if lb == nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking for AWS ELB: %v", err)
	}

	var status []kops.ApiIngressStatus
	if lb != nil {
		lbDnsName := fi.StringValue(lb.DNSName)
		if lbDnsName == "" {
			return nil, fmt.Errorf("found ELB %q, but it did not have a DNSName", name)
		}
		status = append(status, kops.ApiIngressStatus{Hostname: lbDnsName})
	}

	return status, nil
}

func getApiIngressStatusGCE(cloud gce.GCECloud, cluster *kops.Cluster) ([]kops.ApiIngressStatus, error) {
	status, err := cloud.GetApiIngressStatus(cluster)
	if err != nil {
		return nil, err
	}
	return status, nil
}
