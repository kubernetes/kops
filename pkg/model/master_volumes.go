/*
Copyright 2016 The Kubernetes Authors.

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

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
)

const (
	DefaultEtcdVolumeSize    = 20
	DefaultAWSEtcdVolumeType = "gp2"
	DefaultGCEEtcdVolumeType = "pd-ssd"
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

			if len(ig.Spec.Subnets) == 0 {
				return fmt.Errorf("Must specify a subnet for instancegroup %q used by etcd %s/%s", igName, m.Name, etcd.Name)
			}
			if len(ig.Spec.Subnets) != 1 {
				return fmt.Errorf("Must specify a unique subnet for instancegroup %q used by etcd %s/%s", igName, m.Name, etcd.Name)
			}

			subnet := b.FindSubnet(ig.Spec.Subnets[0])
			if subnet == nil {
				return fmt.Errorf("Subnet %q not found (specified by instancegroup %q)", ig.Spec.Subnets[0], igName)
			}

			if subnet.Zone == "" {
				return fmt.Errorf("Subnet %q did not specify a zone", subnet.Name)
			}

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
				b.addAWSVolume(c, name, volumeSize, subnet, etcd, m, allMembers)
			case kops.CloudProviderGCE:
				b.addGCEVolume(c, name, volumeSize, subnet, etcd, m, allMembers)
			case kops.CloudProviderVSphere:
				b.addVSphereVolume(c, name, volumeSize, subnet, etcd, m, allMembers)
			default:
				return fmt.Errorf("unknown cloudprovider %q", b.Cluster.Spec.CloudProvider)
			}
		}
	}
	return nil
}

func (b *MasterVolumeBuilder) addAWSVolume(c *fi.ModelBuilderContext, name string, volumeSize int32, subnet *kops.ClusterSubnetSpec, etcd *kops.EtcdClusterSpec, m *kops.EtcdMemberSpec, allMembers []string) {
	volumeType := fi.StringValue(m.VolumeType)
	if volumeType == "" {
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

	encrypted := fi.BoolValue(m.EncryptedVolume)

	t := &awstasks.EBSVolume{
		Name:      s(name),
		Lifecycle: b.Lifecycle,

		AvailabilityZone: s(subnet.Zone),
		SizeGB:           fi.Int64(int64(volumeSize)),
		VolumeType:       s(volumeType),
		KmsKeyId:         m.KmsKeyId,
		Encrypted:        fi.Bool(encrypted),
		Tags:             tags,
	}

	c.AddTask(t)
}

func (b *MasterVolumeBuilder) addGCEVolume(c *fi.ModelBuilderContext, name string, volumeSize int32, subnet *kops.ClusterSubnetSpec, etcd *kops.EtcdClusterSpec, m *kops.EtcdMemberSpec, allMembers []string) {
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

	name = strings.Replace(name, ".", "-", -1)

	t := &gcetasks.Disk{
		Name:      s(name),
		Lifecycle: b.Lifecycle,

		Zone:       s(subnet.Zone),
		SizeGB:     fi.Int64(int64(volumeSize)),
		VolumeType: s(volumeType),
		Labels:     tags,
	}

	c.AddTask(t)
}

func (b *MasterVolumeBuilder) addVSphereVolume(c *fi.ModelBuilderContext, name string, volumeSize int32, subnet *kops.ClusterSubnetSpec, etcd *kops.EtcdClusterSpec, m *kops.EtcdMemberSpec, allMembers []string) {
	fmt.Print("addVSphereVolume to be implemented")
}
