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
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"sort"
	"strings"
)

const (
	DefaultEtcdVolumeSize = 20
	DefaultEtcdVolumeType = "gp2"
)

// MasterVolumeBuilder builds master EBS volumes
type MasterVolumeBuilder struct {
	*KopsModelContext
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

			volumeSize := int64(fi.Int32Value(m.VolumeSize))
			if volumeSize == 0 {
				volumeSize = DefaultEtcdVolumeSize
			}
			volumeType := fi.StringValue(m.VolumeType)
			if volumeType == "" {
				volumeType = DefaultEtcdVolumeType
			}

			encrypted := fi.BoolValue(m.EncryptedVolume)

			// The tags are how protokube knows to mount the volume and use it for etcd
			tags := make(map[string]string)
			{
				var allMembers []string
				for _, m := range etcd.Members {
					allMembers = append(allMembers, m.Name)
				}

				sort.Strings(allMembers)

				// This is the configuration of the etcd cluster
				tags["k8s.io/etcd/"+etcd.Name] = m.Name + "/" + strings.Join(allMembers, ",")

				// This says "only mount on a master"
				tags["k8s.io/role/master"] = "1"
			}

			t := &awstasks.EBSVolume{
				Name:             s(name),
				AvailabilityZone: s(subnet.Zone),
				SizeGB:           fi.Int64(volumeSize),
				VolumeType:       s(volumeType),
				KmsKeyId:         m.KmsKeyId,
				Encrypted:        fi.Bool(encrypted),
				Tags:             tags,
			}

			c.AddTask(t)
		}
	}
	return nil
}
