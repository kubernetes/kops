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

package cloudup

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/util/pkg/vfs"
	"strings"
)

// PerformAssignments populates values that are required and immutable
// For example, it assigns stable Keys to InstanceGroups & Masters, and
// it assigns CIDRs to subnets
// We also assign KubernetesVersion, because we want it to be explicit
//
// PerformAssignments is called on create, as well as an update. In fact
// any time Run() is called in apply_cluster.go we will reach this function.
// Please do all after-market logic here.
//
func PerformAssignments(c *kops.Cluster) error {
	cloud, err := BuildCloud(c)
	if err != nil {
		return err
	}

	if c.SharedVPC() && c.Spec.NetworkCIDR == "" {
		vpcInfo, err := cloud.FindVPCInfo(c.Spec.NetworkID)
		if err != nil {
			return err
		}
		if vpcInfo == nil {
			return fmt.Errorf("unable to find VPC ID %q", c.Spec.NetworkID)
		}
		c.Spec.NetworkCIDR = vpcInfo.CIDR
		if c.Spec.NetworkCIDR == "" {
			return fmt.Errorf("Unable to infer NetworkCIDR from VPC ID, please specify --network-cidr")
		}
	}

	// Topology support
	// TODO Kris: Unsure if this needs to be here, or if the API conversion code will handle it
	if c.Spec.Topology == nil {
		c.Spec.Topology = &kops.TopologySpec{Masters: kops.TopologyPublic, Nodes: kops.TopologyPublic}
	}

	if c.Spec.NetworkCIDR == "" && !c.SharedVPC() {
		// TODO: Choose non-overlapping networking CIDRs for VPCs, using vpcInfo
		c.Spec.NetworkCIDR = "172.20.0.0/16"
	}

	if c.Spec.NonMasqueradeCIDR == "" {
		c.Spec.NonMasqueradeCIDR = "100.64.0.0/10"
	}

	// TODO: Unclear this should be here - it isn't too hard to change
	if c.Spec.MasterPublicName == "" && c.ObjectMeta.Name != "" {
		c.Spec.MasterPublicName = "api." + c.ObjectMeta.Name
	}

	// TODO: Use vpcInfo
	err = assignCIDRsToSubnets(c)
	if err != nil {
		return err
	}

	return ensureKubernetesVersion(c)
}

// ensureKubernetesVersion populates KubernetesVersion, if it is not already set
// It will be populated with the latest stable kubernetes version, or the version from the channel
func ensureKubernetesVersion(c *kops.Cluster) error {
	if c.Spec.KubernetesVersion == "" {
		if c.Spec.Channel != "" {
			channel, err := kops.LoadChannel(c.Spec.Channel)
			if err != nil {
				return err
			}
			if channel.Spec.Cluster.KubernetesVersion != "" {
				c.Spec.KubernetesVersion = channel.Spec.Cluster.KubernetesVersion
			}
		}
	}

	if c.Spec.KubernetesVersion == "" {
		latestVersion, err := FindLatestKubernetesVersion()
		if err != nil {
			return err
		}
		glog.Infof("Using kubernetes latest stable version: %s", latestVersion)
		c.Spec.KubernetesVersion = latestVersion
	}
	return nil
}

// FindLatestKubernetesVersion returns the latest kubernetes version,
// as stored at https://storage.googleapis.com/kubernetes-release/release/stable.txt
// This shouldn't be used any more; we prefer reading the stable channel
func FindLatestKubernetesVersion() (string, error) {
	stableURL := "https://storage.googleapis.com/kubernetes-release/release/stable.txt"
	glog.Warningf("Loading latest kubernetes version from %q", stableURL)
	b, err := vfs.Context.ReadFile(stableURL)
	if err != nil {
		return "", fmt.Errorf("KubernetesVersion not specified, and unable to download latest version from %q: %v", stableURL, err)
	}
	latestVersion := strings.TrimSpace(string(b))
	return latestVersion, nil
}
