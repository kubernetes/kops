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
			base := clusterSpec.ConfigBase
			etcdCluster.Backups.BackupStore = urls.Join(base, "backups", "etcd", etcdCluster.Name)
		}

		if !etcdVersionIsSupported(etcdCluster.Version) {
			if featureflag.SkipEtcdVersionCheck.Enabled() {
				klog.Warningf("etcd version %q is not known to be supported, but ignoring because of SkipEtcdVersionCheck feature flag", etcdCluster.Version)
			} else {
				klog.Warningf("Unsupported etcd version %q detected; please update etcd version.", etcdCluster.Version)
				klog.Warningf("Use export KOPS_FEATURE_FLAGS=SkipEtcdVersionCheck to override this check.")
				klog.Warningf("Supported etcd versions: %s", strings.Join(etcdSupportedVersions(), ", "))
				return fmt.Errorf("etcd version %q is not supported with etcd-manager, please specify a supported version or remove the value to use the recommended version", etcdCluster.Version)
			}
		}
	}

	return nil
}

var etcdSupportedImages = map[string]string{
	"3.2.24": "registry.k8s.io/etcd:3.2.24-1",
	"3.3.10": "registry.k8s.io/etcd:3.3.10-0",
	"3.3.17": "registry.k8s.io/etcd:3.3.17-0",
	"3.4.3":  "registry.k8s.io/etcd:3.4.3-0",
	"3.4.13": "registry.k8s.io/etcd:3.4.13-0",
	"3.5.0":  "registry.k8s.io/etcd:3.5.0-0",
	"3.5.1":  "registry.k8s.io/etcd:3.5.1-0",
	"3.5.3":  "registry.k8s.io/etcd:3.5.3-0",
	"3.5.4":  "registry.k8s.io/etcd:3.5.4-0",
	"3.5.6":  "registry.k8s.io/etcd:3.5.6-0",
	"3.5.7":  "registry.k8s.io/etcd:3.5.7-0",
}

func etcdSupportedVersions() []string {
	var versions []string
	for etcdVersion := range etcdSupportedImages {
		versions = append(versions, etcdVersion)
	}
	sort.Strings(versions)
	return versions
}

func etcdVersionIsSupported(version string) bool {
	version = strings.TrimPrefix(version, "v")
	if _, ok := etcdSupportedImages[version]; ok {
		return true
	}
	return false
}
