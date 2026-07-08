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

package etcdmanager

import (
	"fmt"
	"sort"
	"strings"

	"github.com/blang/semver/v4"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// EtcdManagerOptionsBuilder adds options for the etcd-manager to the model.
type EtcdManagerOptionsBuilder struct {
	*components.OptionsContext
}

var _ loader.ClusterOptionsBuilder = &EtcdManagerOptionsBuilder{}

// BuildOptions generates the configurations used to create etcd manager manifest
func (b *EtcdManagerOptionsBuilder) BuildOptions(o *kops.Cluster) error {
	clusterSpec := &o.Spec

	// Image Volumes will become GA in Kubernetes 1.35
	// https://github.com/kubernetes/enhancements/pull/5450
	if b.ControlPlaneKubernetesVersion().IsLT("1.36.0") && o.HasImageVolumesSupport() {
		if clusterSpec.ControlPlaneKubelet == nil {
			clusterSpec.ControlPlaneKubelet = &kops.KubeletConfigSpec{}
		}
		if clusterSpec.ControlPlaneKubelet.FeatureGates == nil {
			clusterSpec.ControlPlaneKubelet.FeatureGates = make(map[string]string)
		}
		if _, found := clusterSpec.ControlPlaneKubelet.FeatureGates["ImageVolume"]; !found {
			clusterSpec.ControlPlaneKubelet.FeatureGates["ImageVolume"] = "true"
		}
	}

	for i := range clusterSpec.EtcdClusters {
		etcdCluster := &clusterSpec.EtcdClusters[i]
		if etcdCluster.Backups == nil {
			etcdCluster.Backups = &kops.EtcdBackupSpec{}
		}
		if etcdCluster.Backups.BackupStore == "" {
			base := clusterSpec.ConfigStore.Base
			etcdCluster.Backups.BackupStore = join(base, "backups", "etcd", etcdCluster.Name)
		}

		if !etcdVersionIsSupported(etcdCluster.Version) {
			if etcdCluster.Image != "" {
				klog.Warningf("etcd version %q is not bundled by kOps and has not been tested; using binaries from custom image %q", etcdCluster.Version, etcdCluster.Image)
			} else {
				klog.Warningf("Unsupported etcd version %q detected; please update etcd version.", etcdCluster.Version)
				var versions []string
				for _, v := range etcdSupportedVersions() {
					versions = append(versions, v.Version)
				}
				klog.Warningf("Supported etcd versions: %s", strings.Join(versions, ", "))
				return fmt.Errorf("etcd version %q is not supported with etcd-manager, please specify a supported version, remove the value to use the recommended version, or also set the image field to run a custom version", etcdCluster.Version)
			}
		}
	}

	return nil
}

// etcdVersion describes how we want to support each etcd version.
type etcdVersion struct {
	Version          string
	Image            string
	SymlinkToVersion string
}

// etcdLatestImages lists the latest etcd patch image bundled by kops for each
// supported minor. All earlier patch versions within the same minor are
// generated as SymlinkToVersion entries by etcdSupportedVersions.
var etcdLatestImages = []etcdVersion{
	{Version: components.LatestEtcd35Version, Image: "registry.k8s.io/etcd:v" + components.LatestEtcd35Version},
	{Version: components.LatestEtcd36Version, Image: "registry.k8s.io/etcd:v" + components.LatestEtcd36Version},
	{Version: components.LatestEtcd37Version, Image: "registry.k8s.io/etcd:v" + components.LatestEtcd37Version},
}

func etcdSupportedVersions() []etcdVersion {
	var versions []etcdVersion
	for _, latest := range etcdLatestImages {
		sv := semver.MustParse(latest.Version)
		versions = append(versions, latest)
		for patch := uint64(0); patch < sv.Patch; patch++ {
			versions = append(versions, etcdVersion{
				Version:          fmt.Sprintf("%d.%d.%d", sv.Major, sv.Minor, patch),
				SymlinkToVersion: latest.Version,
			})
		}
	}
	sort.Slice(versions, func(i, j int) bool {
		return semver.MustParse(versions[i].Version).LT(semver.MustParse(versions[j].Version))
	})
	return versions
}

func etcdVersionIsSupported(version string) bool {
	version = strings.TrimPrefix(version, "v")
	for _, etcdVersion := range etcdSupportedVersions() {
		if etcdVersion.Version == version {
			return true
		}
	}
	return false
}

func join(base string, others ...string) string {
	u := base
	for _, o := range others {
		if !strings.HasSuffix(u, "/") {
			u += "/"
		}
		if strings.HasPrefix(o, "/") {
			u += o[1:]
		} else {
			u += o
		}
	}
	return u
}
