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

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/pkg/urls"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// EtcdManagerOptionsBuilder adds options for the etcd-manager to the model.
type EtcdManagerOptionsBuilder struct {
	*components.OptionsContext
}

var _ loader.OptionsBuilder = &EtcdManagerOptionsBuilder{}

// BuildOptions generates the configurations used to create etcd manager manifest
func (b *EtcdManagerOptionsBuilder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)

	for i := range clusterSpec.EtcdClusters {
		etcdCluster := &clusterSpec.EtcdClusters[i]
		if etcdCluster.Backups == nil {
			etcdCluster.Backups = &kops.EtcdBackupSpec{}
		}
		if etcdCluster.Backups.BackupStore == "" {
			base := clusterSpec.ConfigStore.Base
			etcdCluster.Backups.BackupStore = urls.Join(base, "backups", "etcd", etcdCluster.Name)
		}

		if !etcdVersionIsSupported(etcdCluster.Version) {
			if featureflag.SkipEtcdVersionCheck.Enabled() {
				klog.Warningf("etcd version %q is not known to be supported, but ignoring because of SkipEtcdVersionCheck feature flag", etcdCluster.Version)
			} else {
				klog.Warningf("Unsupported etcd version %q detected; please update etcd version.", etcdCluster.Version)
				klog.Warningf("Use export KOPS_FEATURE_FLAGS=SkipEtcdVersionCheck to override this check.")
				var versions []string
				for _, v := range etcdSupportedVersions() {
					versions = append(versions, v.Version)
				}
				klog.Warningf("Supported etcd versions: %s", strings.Join(versions, ", "))
				return fmt.Errorf("etcd version %q is not supported with etcd-manager, please specify a supported version or remove the value to use the recommended version", etcdCluster.Version)
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

var etcdSupportedImages = []etcdVersion{
	{Version: "3.4.3", SymlinkToVersion: "3.4.13"},
	{Version: "3.4.13", Image: "registry.k8s.io/etcd:3.4.13-0"},
	{Version: "3.5.0", SymlinkToVersion: "3.5.17"},
	{Version: "3.5.1", SymlinkToVersion: "3.5.17"},
	{Version: "3.5.3", SymlinkToVersion: "3.5.17"},
	{Version: "3.5.4", SymlinkToVersion: "3.5.17"},
	{Version: "3.5.6", SymlinkToVersion: "3.5.17"},
	{Version: "3.5.7", SymlinkToVersion: "3.5.17"},
	{Version: "3.5.9", SymlinkToVersion: "3.5.17"},
	{Version: "3.5.13", SymlinkToVersion: "3.5.17"},
	{Version: "3.5.17", Image: "registry.k8s.io/etcd:3.5.17-0"},
}

func etcdSupportedVersions() []etcdVersion {
	var versions []etcdVersion
	versions = append(versions, etcdSupportedImages...)
	sort.Slice(versions, func(i, j int) bool { return versions[i].Version < versions[j].Version })
	return versions
}

func etcdVersionIsSupported(version string) bool {
	version = strings.TrimPrefix(version, "v")
	for _, etcdVersion := range etcdSupportedImages {
		if etcdVersion.Version == version {
			return true
		}
	}
	return false
}
