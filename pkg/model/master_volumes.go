/*
Copyright 2019 The Kubernetes Authors.

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

package model

import (
	"fmt"
	"sort"
	"strings"

	"k8s.io/kops/upup/pkg/fi/cloudup/yandex"
	"k8s.io/kops/upup/pkg/fi/cloudup/yandextasks"

	"github.com/aws/aws-sdk-go/service/ec2"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
	"k8s.io/kops/upup/pkg/fi/cloudup/azuretasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/do"
	"k8s.io/kops/upup/pkg/fi/cloudup/dotasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetznertasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstacktasks"
)

const (
	DefaultEtcdVolumeSize             = 20
	DefaultAWSEtcdVolumeType          = ec2.VolumeTypeGp3
	DefaultAWSEtcdVolumeIonIops       = 100
	DefaultAWSEtcdVolumeGp3Iops       = 3000
	DefaultAWSEtcdVolumeGp3Throughput = 125
	DefaultGCEEtcdVolumeType          = "pd-ssd"

	// DefaultYandexVolumeType YandexCloud disk types are network-hdd, network-ssd, network-ssd-nonreplicated
	DefaultYandexVolumeType = "network-ssd"
)

// MasterVolumeBuilder builds master EBS volumes
type MasterVolumeBuilder struct {
	*KopsModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.ModelBuilder = &MasterVolumeBuilder{}

func (b *MasterVolumeBuilder) Build(c *fi.ModelBuilderContext) error {
	for _, etcd := range b.Cluster.Spec.EtcdClusters {
		for _, m := range etcd.Members {
			// EBS volume for each member of the each etcd cluster
			prefix := m.Name + ".etcd-" + etcd.Name
			name := prefix + "." + b.ClusterName()

			igName := fi.StringValue(m.InstanceGroup)
			if igName == "" {
				return fmt.Errorf("InstanceGroup not set on etcd %s/%s", m.Name, etcd.Name)
			}
			ig := b.FindInstanceGroup(igName)
			if ig == nil {
				return fmt.Errorf("InstanceGroup not found (for etcd %s/%s): %q", m.Name, etcd.Name, igName)
			}

			zones, err := model.FindZonesForInstanceGroup(b.Cluster, ig)
			if err != nil {
				return err
			}
			if len(zones) == 0 {
				return fmt.Errorf("must specify a zone for instancegroup %q used by etcd %s/%s", igName, m.Name, etcd.Name)
			}
			if len(zones) != 1 {
				return fmt.Errorf("must specify a unique zone for instancegroup %q used by etcd %s/%s", igName, m.Name, etcd.Name)
			}
			zone := zones[0]

			volumeSize := fi.Int32Value(m.VolumeSize)
			if volumeSize == 0 {
				volumeSize = DefaultEtcdVolumeSize
			}

			var allMembers []string
			for _, m := range etcd.Members {
				allMembers = append(allMembers, m.Name)
			}
			sort.Strings(allMembers)

			switch b.Cluster.Spec.GetCloudProvider() {
			case kops.CloudProviderAWS:
				err = b.addAWSVolume(c, name, volumeSize, zone, etcd, m, allMembers)
				if err != nil {
					return err
				}
			case kops.CloudProviderDO:
				b.addDOVolume(c, name, volumeSize, zone, etcd, m, allMembers)
			case kops.CloudProviderGCE:
				b.addGCEVolume(c, prefix, volumeSize, zone, etcd, m, allMembers)
			case kops.CloudProviderHetzner:
				b.addHetznerVolume(c, name, volumeSize, zone, etcd, m, allMembers)
			case kops.CloudProviderOpenstack:
				err = b.addOpenstackVolume(c, name, volumeSize, zone, etcd, m, allMembers)
				if err != nil {
					return err
				}
			case kops.CloudProviderAzure:
				b.addAzureVolume(c, name, volumeSize, zone, etcd, m, allMembers)
			case kops.CloudProviderYandex:
				b.addYandexVolume(c, name, volumeSize, zone, etcd, m, allMembers)
			default:
				return fmt.Errorf("unknown cloudprovider %q", b.Cluster.Spec.GetCloudProvider())
			}
		}
	}
	return nil
}

func (b *MasterVolumeBuilder) addAWSVolume(c *fi.ModelBuilderContext, name string, volumeSize int32, zone string, etcd kops.EtcdClusterSpec, m kops.EtcdMemberSpec, allMembers []string) error {
	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSVolumeTypes.html
	volumeType := fi.StringValue(m.VolumeType)
	if volumeType == "" {
		volumeType = DefaultAWSEtcdVolumeType
	}
	volumeIops := fi.Int32Value(m.VolumeIOPS)
	volumeThroughput := fi.Int32Value(m.VolumeThroughput)
	switch volumeType {
	case ec2.VolumeTypeIo1, ec2.VolumeTypeIo2:
		if volumeIops < 100 {
			volumeIops = DefaultAWSEtcdVolumeIonIops
		}
	case ec2.VolumeTypeGp3:
		if volumeIops < 3000 {
			volumeIops = DefaultAWSEtcdVolumeGp3Iops
		}
		if volumeThroughput < 125 {
			volumeThroughput = DefaultAWSEtcdVolumeGp3Throughput
		}
	}

	if err := validateAWSVolume(name, volumeType, volumeSize, volumeIops, volumeThroughput); err != nil {
		return err
	}

	// The tags are how protokube knows to mount the volume and use it for etcd
	tags := make(map[string]string)

	// Apply all user defined labels on the volumes
	for k, v := range b.Cluster.Spec.CloudLabels {
		tags[k] = v
	}

	// tags[awsup.TagClusterName] = b.C.cluster.Name
	// This is the configuration of the etcd cluster
	tags[awsup.TagNameEtcdClusterPrefix+etcd.Name] = m.Name + "/" + strings.Join(allMembers, ",")
	// This says "only mount on a master"
	tags[awsup.TagNameRolePrefix+"master"] = "1"

	// We always add an owned tags (these can't be shared)
	tags["kubernetes.io/cluster/"+b.Cluster.ObjectMeta.Name] = "owned"

	encrypted := fi.BoolValue(m.EncryptedVolume)

	t := &awstasks.EBSVolume{
		Name:      fi.String(name),
		Lifecycle: b.Lifecycle,

		AvailabilityZone: fi.String(zone),
		SizeGB:           fi.Int64(int64(volumeSize)),
		VolumeType:       fi.String(volumeType),
		KmsKeyId:         m.KmsKeyID,
		Encrypted:        fi.Bool(encrypted),
		Tags:             tags,
	}
	switch volumeType {
	case ec2.VolumeTypeGp3:
		t.VolumeThroughput = fi.Int64(int64(volumeThroughput))
		fallthrough
	case ec2.VolumeTypeIo1, ec2.VolumeTypeIo2:
		t.VolumeIops = fi.Int64(int64(volumeIops))
	}

	c.AddTask(t)

	return nil
}

func validateAWSVolume(name, volumeType string, volumeSize, volumeIops, volumeThroughput int32) error {
	volumeIopsSizeRatio := float64(volumeIops) / float64(volumeSize)
	volumeThroughputIopsRatio := float64(volumeThroughput) / float64(volumeIops)
	switch volumeType {
	case ec2.VolumeTypeIo1:
		if volumeIopsSizeRatio > 50.0 {
			return fmt.Errorf("volumeIops to volumeSize ratio must be lower than 50. For %s ratio is %.02f", name, volumeIopsSizeRatio)
		}
	case ec2.VolumeTypeIo2:
		if volumeIopsSizeRatio > 500.0 {
			return fmt.Errorf("volumeIops to volumeSize ratio must be lower than 500. For %s ratio is %.02f", name, volumeIopsSizeRatio)
		}
	case ec2.VolumeTypeGp3:
		if volumeIops > 3000 && volumeIopsSizeRatio > 500.0 {
			return fmt.Errorf("volumeIops to volumeSize ratio must be lower than 500. For %s ratio is %.02f", name, volumeIopsSizeRatio)
		}
		if volumeThroughputIopsRatio > 0.25 {
			return fmt.Errorf("volumeThroughput to volumeIops ratio must be lower than 0.25. For %s ratio is %.02f", name, volumeThroughputIopsRatio)
		}
	}
	return nil
}

func (b *MasterVolumeBuilder) addDOVolume(c *fi.ModelBuilderContext, name string, volumeSize int32, zone string, etcd kops.EtcdClusterSpec, m kops.EtcdMemberSpec, allMembers []string) {
	// required that names start with a lower case and only contains letters, numbers and hyphens
	name = "kops-" + do.SafeClusterName(name)

	// DO has a 64 character limit for volume names
	if len(name) >= 64 {
		name = name[:64]
	}

	tags := make(map[string]string)
	tags[do.TagNameEtcdClusterPrefix+etcd.Name] = do.SafeClusterName(m.Name)
	tags[do.TagKubernetesClusterIndex] = do.SafeClusterName(m.Name)

	// We always add an owned tags (these can't be shared)
	tags[do.TagKubernetesClusterNamePrefix] = do.SafeClusterName(b.Cluster.ObjectMeta.Name)

	t := &dotasks.Volume{
		Name:      fi.String(name),
		Lifecycle: b.Lifecycle,
		SizeGB:    fi.Int64(int64(volumeSize)),
		Region:    fi.String(zone),
		Tags:      tags,
	}

	c.AddTask(t)
}

func (b *MasterVolumeBuilder) addGCEVolume(c *fi.ModelBuilderContext, prefix string, volumeSize int32, zone string, etcd kops.EtcdClusterSpec, m kops.EtcdMemberSpec, allMembers []string) {
	volumeType := fi.StringValue(m.VolumeType)
	if volumeType == "" {
		volumeType = DefaultGCEEtcdVolumeType
	}

	// TODO: Should no longer be needed because we trim prefixes
	//// On GCE we are close to the length limits.  So,we remove the dashes from the keys
	//// The name is normally something like "us-east1-a", and the dashes are particularly expensive
	//// because of the escaping needed (3 characters for each dash)
	//switch tf.cluster.Spec.CloudProvider {
	//case string(kops.CloudProviderGCE):
	//	// TODO: If we're still struggling for size, we don't need to put ourselves in the allmembers list
	//	for i := range allMembers {
	//		allMembers[i] = strings.Replace(allMembers[i], "-", "", -1)
	//	}
	//	meName = strings.Replace(meName, "-", "", -1)
	//}

	// This is the configuration of the etcd cluster
	clusterSpec := m.Name + "/" + strings.Join(allMembers, ",")

	// The tags are how protokube knows to mount the volume and use it for etcd
	tags := make(map[string]string)
	tags[gce.GceLabelNameKubernetesCluster] = gce.SafeClusterName(b.ClusterName())
	tags[gce.GceLabelNameRolePrefix+"master"] = "master" // Can't start with a number
	tags[gce.GceLabelNameEtcdClusterPrefix+etcd.Name] = gce.EncodeGCELabel(clusterSpec)

	// GCE disk names must match the following regular expression: '[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?'
	prefix = strings.Replace(prefix, ".", "-", -1)
	if strings.IndexByte("0123456789-", prefix[0]) != -1 {
		prefix = "d" + prefix
	}
	name := gce.ClusterSuffixedName(prefix, b.Cluster.ObjectMeta.Name, 63)

	t := &gcetasks.Disk{
		Name:      fi.String(name),
		Lifecycle: b.Lifecycle,

		Zone:       fi.String(zone),
		SizeGB:     fi.Int64(int64(volumeSize)),
		VolumeType: fi.String(volumeType),
		Labels:     tags,
	}

	c.AddTask(t)
}

func (b *MasterVolumeBuilder) addHetznerVolume(c *fi.ModelBuilderContext, name string, volumeSize int32, zone string, etcd kops.EtcdClusterSpec, m kops.EtcdMemberSpec, allMembers []string) {
	tags := make(map[string]string)
	tags[hetzner.TagKubernetesClusterName] = b.Cluster.ObjectMeta.Name
	tags[hetzner.TagKubernetesInstanceGroup] = fi.StringValue(m.InstanceGroup)
	tags[hetzner.TagKubernetesVolumeRole] = etcd.Name

	t := &hetznertasks.Volume{
		Name:      fi.String(name),
		Lifecycle: b.Lifecycle,
		Size:      int(volumeSize),
		Location:  zone,
		Labels:    tags,
	}
	c.AddTask(t)

	return
}

func (b *MasterVolumeBuilder) addOpenstackVolume(c *fi.ModelBuilderContext, name string, volumeSize int32, zone string, etcd kops.EtcdClusterSpec, m kops.EtcdMemberSpec, allMembers []string) error {
	volumeType := fi.StringValue(m.VolumeType)

	// The tags are how protokube knows to mount the volume and use it for etcd
	tags := make(map[string]string)
	// Apply all user defined labels on the volumes
	for k, v := range b.Cluster.Spec.CloudLabels {
		tags[k] = v
	}
	// This is the configuration of the etcd cluster
	tags[openstack.TagNameEtcdClusterPrefix+etcd.Name] = m.Name + "/" + strings.Join(allMembers, ",")
	// This says "only mount on a master"
	tags[openstack.TagNameRolePrefix+"master"] = "1"

	// override zone
	if b.Cluster.Spec.CloudProvider.Openstack.BlockStorage != nil && b.Cluster.Spec.CloudProvider.Openstack.BlockStorage.OverrideAZ != nil {
		zone = fi.StringValue(b.Cluster.Spec.CloudProvider.Openstack.BlockStorage.OverrideAZ)
	}
	t := &openstacktasks.Volume{
		Name:             fi.String(name),
		AvailabilityZone: fi.String(zone),
		VolumeType:       fi.String(volumeType),
		SizeGB:           fi.Int64(int64(volumeSize)),
		Tags:             tags,
		Lifecycle:        b.Lifecycle,
	}
	c.AddTask(t)

	return nil
}

func (b *MasterVolumeBuilder) addAzureVolume(
	c *fi.ModelBuilderContext,
	name string,
	volumeSize int32,
	zone string,
	etcd kops.EtcdClusterSpec,
	m kops.EtcdMemberSpec,
	allMembers []string,
) {
	// The tags are use by Protokube to mount the volume and use it for etcd.
	tags := map[string]*string{
		// This is the configuration of the etcd cluster.
		azure.TagNameEtcdClusterPrefix + etcd.Name: fi.String(m.Name + "/" + strings.Join(allMembers, ",")),
		// This says "only mount on a master".
		azure.TagNameRolePrefix + azure.TagRoleMaster: fi.String("1"),
		// We always add an owned tags (these can't be shared).
		// Use dash (_) as a splitter. Other CSPs use slash (/), but slash is not
		// allowed as a tag key in Azure.
		"kubernetes.io_cluster_" + b.Cluster.ObjectMeta.Name: fi.String("owned"),
	}

	// Apply all user defined labels on the volumes.
	for k, v := range b.Cluster.Spec.CloudLabels {
		tags[k] = fi.String(v)
	}

	// TODO(kenji): Respect zone and m.EncryptedVolume.
	t := &azuretasks.Disk{
		Name:      fi.String(name),
		Lifecycle: b.Lifecycle,
		// We cannot use AzureModelContext.LinkToResourceGroup() here because of cyclic dependency.
		ResourceGroup: &azuretasks.ResourceGroup{
			Name: fi.String(b.Cluster.AzureResourceGroupName()),
		},
		SizeGB: fi.Int32(volumeSize),
		Tags:   tags,
	}
	c.AddTask(t)
}

func (b *MasterVolumeBuilder) addYandexVolume(
	c *fi.ModelBuilderContext,
	name string,
	volumeSize int32,
	zone string,
	etcd kops.EtcdClusterSpec,
	m kops.EtcdMemberSpec,
	allMembers []string,
) {
	//TODO(YuraBeznos): yandex implement volume for etcd
	volumeType := fi.StringValue(m.VolumeType)
	if volumeType == "" {
		volumeType = DefaultYandexVolumeType
	}

	// This is the configuration of the etcd cluster
	//clusterSpec := m.Name + "/" + strings.Join(allMembers, ",")

	// The tags are how etcd-manager knows to mount the volume and use it for etcd
	tags := make(map[string]string)
	tags[yandex.TagKubernetesClusterName] = b.ClusterName()
	tags[yandex.TagKubernetesInstanceRole] = "master" // Can't start with a number
	tags[yandex.TagKubernetesVolumeRole] = etcd.Name
	//tags[yandex.TagKubernetesInstanceGroup] =
	volumeSizeBytes := int64(volumeSize) * 1024 * 1024 * 1024
	t := &yandextasks.Disk{
		Name:      fi.String(name),
		Lifecycle: b.Lifecycle,

		FolderId: &b.Cluster.Spec.Project,
		ZoneId:   fi.String(zone),
		Size:     fi.Int64(volumeSizeBytes),
		TypeId:   fi.String(volumeType),
		Labels:   tags,
	}
	c.AddTask(t)
}
