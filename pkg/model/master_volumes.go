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

	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/alitasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/aliup"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/do"
	"k8s.io/kops/upup/pkg/fi/cloudup/dotasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstacktasks"
)

const (
	DefaultEtcdVolumeSize    = 20
	DefaultAWSEtcdVolumeType = "gp2"
	DefaultAWSEtcdVolumeIops = 100
	DefaultGCEEtcdVolumeType = "pd-ssd"
	DefaultALIEtcdVolumeType = "cloud_ssd"
)

// MasterVolumeBuilder builds master EBS volumes
type MasterVolumeBuilder struct {
	*KopsModelContext
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &MasterVolumeBuilder{}

func (b *MasterVolumeBuilder) Build(c *fi.ModelBuilderContext) error {
	for _, etcd := range b.Cluster.Spec.EtcdClusters {
		for _, m := range etcd.Members {
			// EBS volume for each member of the each etcd cluster
			name := m.Name + ".etcd-" + etcd.Name + "." + b.ClusterName()

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

			switch kops.CloudProviderID(b.Cluster.Spec.CloudProvider) {
			case kops.CloudProviderAWS:
				err = b.addAWSVolume(c, name, volumeSize, zone, etcd, m, allMembers)
				if err != nil {
					return err
				}
			case kops.CloudProviderDO:
				b.addDOVolume(c, name, volumeSize, zone, etcd, m, allMembers)
			case kops.CloudProviderGCE:
				b.addGCEVolume(c, name, volumeSize, zone, etcd, m, allMembers)
			case kops.CloudProviderVSphere:
				b.addVSphereVolume(c, name, volumeSize, zone, etcd, m, allMembers)
			case kops.CloudProviderBareMetal:
				klog.Fatalf("BareMetal not implemented")
			case kops.CloudProviderOpenstack:
				err = b.addOpenstackVolume(c, name, volumeSize, zone, etcd, m, allMembers)
				if err != nil {
					return err
				}
			case kops.CloudProviderALI:
				b.addALIVolume(c, name, volumeSize, zone, etcd, m, allMembers)
			default:
				return fmt.Errorf("unknown cloudprovider %q", b.Cluster.Spec.CloudProvider)
			}
		}
	}
	return nil
}

func (b *MasterVolumeBuilder) addAWSVolume(c *fi.ModelBuilderContext, name string, volumeSize int32, zone string, etcd *kops.EtcdClusterSpec, m *kops.EtcdMemberSpec, allMembers []string) error {
	volumeType := fi.StringValue(m.VolumeType)
	volumeIops := fi.Int32Value(m.VolumeIops)
	switch volumeType {
	case "io1":
		if volumeIops <= 0 {
			volumeIops = DefaultAWSEtcdVolumeIops
		}
	default:
		volumeType = DefaultAWSEtcdVolumeType
	}

	// The tags are how protokube knows to mount the volume and use it for etcd
	tags := make(map[string]string)

	// Apply all user defined labels on the volumes
	for k, v := range b.Cluster.Spec.CloudLabels {
		tags[k] = v
	}

	//tags[awsup.TagClusterName] = b.C.cluster.Name
	// This is the configuration of the etcd cluster
	tags[awsup.TagNameEtcdClusterPrefix+etcd.Name] = m.Name + "/" + strings.Join(allMembers, ",")
	// This says "only mount on a master"
	tags[awsup.TagNameRolePrefix+"master"] = "1"

	// We always add an owned tags (these can't be shared)
	tags["kubernetes.io/cluster/"+b.Cluster.ObjectMeta.Name] = "owned"

	encrypted := fi.BoolValue(m.EncryptedVolume)

	t := &awstasks.EBSVolume{
		Name:      s(name),
		Lifecycle: b.Lifecycle,

		AvailabilityZone: s(zone),
		SizeGB:           fi.Int64(int64(volumeSize)),
		VolumeType:       s(volumeType),
		KmsKeyId:         m.KmsKeyId,
		Encrypted:        fi.Bool(encrypted),
		Tags:             tags,
	}
	if volumeType == "io1" {
		t.VolumeIops = i64(int64(volumeIops))

		// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSVolumeTypes.html
		if float64(*t.VolumeIops)/float64(*t.SizeGB) > 50.0 {
			return fmt.Errorf("volumeIops to volumeSize ratio must be lower than 50. For %s ratio is %f", *t.Name, float64(*t.VolumeIops)/float64(*t.SizeGB))
		}
	}

	c.AddTask(t)

	return nil
}

func (b *MasterVolumeBuilder) addDOVolume(c *fi.ModelBuilderContext, name string, volumeSize int32, zone string, etcd *kops.EtcdClusterSpec, m *kops.EtcdMemberSpec, allMembers []string) {
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
		Name:      s(name),
		Lifecycle: b.Lifecycle,
		SizeGB:    fi.Int64(int64(volumeSize)),
		Region:    s(zone),
		Tags:      tags,
	}

	c.AddTask(t)
}

func (b *MasterVolumeBuilder) addGCEVolume(c *fi.ModelBuilderContext, name string, volumeSize int32, zone string, etcd *kops.EtcdClusterSpec, m *kops.EtcdMemberSpec, allMembers []string) {
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
	name = strings.Replace(name, ".", "-", -1)
	if strings.IndexByte("0123456789-", name[0]) != -1 {
		name = "d" + name
	}

	t := &gcetasks.Disk{
		Name:      s(name),
		Lifecycle: b.Lifecycle,

		Zone:       s(zone),
		SizeGB:     fi.Int64(int64(volumeSize)),
		VolumeType: s(volumeType),
		Labels:     tags,
	}

	c.AddTask(t)
}

func (b *MasterVolumeBuilder) addVSphereVolume(c *fi.ModelBuilderContext, name string, volumeSize int32, zone string, etcd *kops.EtcdClusterSpec, m *kops.EtcdMemberSpec, allMembers []string) {
	fmt.Print("addVSphereVolume to be implemented")
}

func (b *MasterVolumeBuilder) addOpenstackVolume(c *fi.ModelBuilderContext, name string, volumeSize int32, zone string, etcd *kops.EtcdClusterSpec, m *kops.EtcdMemberSpec, allMembers []string) error {
	volumeType := fi.StringValue(m.VolumeType)
	if volumeType == "" {
		return fmt.Errorf("must set ETCDMemberSpec.VolumeType on Openstack platform")
	}

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
	if b.Cluster.Spec.CloudConfig.Openstack.BlockStorage != nil && b.Cluster.Spec.CloudConfig.Openstack.BlockStorage.OverrideAZ != nil {
		zone = fi.StringValue(b.Cluster.Spec.CloudConfig.Openstack.BlockStorage.OverrideAZ)
	}
	t := &openstacktasks.Volume{
		Name:             s(name),
		AvailabilityZone: s(zone),
		VolumeType:       s(volumeType),
		SizeGB:           fi.Int64(int64(volumeSize)),
		Tags:             tags,
		Lifecycle:        b.Lifecycle,
	}
	c.AddTask(t)

	return nil
}

func (b *MasterVolumeBuilder) addALIVolume(c *fi.ModelBuilderContext, name string, volumeSize int32, zone string, etcd *kops.EtcdClusterSpec, m *kops.EtcdMemberSpec, allMembers []string) {
	//Alicloud does not support volumeName starts with number
	name = "v" + name
	volumeType := fi.StringValue(m.VolumeType)
	if volumeType == "" {
		volumeType = DefaultALIEtcdVolumeType
	}

	// The tags are how protokube knows to mount the volume and use it for etcd
	tags := make(map[string]string)

	// Apply all user defined labels on the volumes
	for k, v := range b.Cluster.Spec.CloudLabels {
		tags[k] = v
	}

	// This is the configuration of the etcd cluster
	tags[aliup.TagNameEtcdClusterPrefix+etcd.Name] = m.Name + "/" + strings.Join(allMembers, ",")
	// This says "only mount on a master"
	tags[aliup.TagNameRolePrefix+"master"] = "1"
	// We always add an owned tags (these can't be shared)
	tags["kubernetes.io/cluster/"+b.Cluster.ObjectMeta.Name] = "owned"

	encrypted := fi.BoolValue(m.EncryptedVolume)

	t := &alitasks.Disk{
		Lifecycle:    b.Lifecycle,
		Name:         s(name),
		ZoneId:       s(zone),
		SizeGB:       fi.Int(int(volumeSize)),
		DiskCategory: s(volumeType),
		Encrypted:    fi.Bool(encrypted),
		Tags:         tags,
	}

	c.AddTask(t)
}
