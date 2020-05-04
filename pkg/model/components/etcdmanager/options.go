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
	"strings"

	"k8s.io/klog"
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

	for _, etcdCluster := range clusterSpec.EtcdClusters {
		if etcdCluster.Provider != kops.EtcdProviderTypeManager {
			continue
		}

		if etcdCluster.Manager == nil {
			etcdCluster.Manager = &kops.EtcdManagerSpec{}
		}

		if etcdCluster.Backups == nil {
			etcdCluster.Backups = &kops.EtcdBackupSpec{}
		}
		if etcdCluster.Backups.BackupStore == "" {
			base := clusterSpec.ConfigBase
			etcdCluster.Backups.BackupStore = urls.Join(base, "backups", "etcd", etcdCluster.Name)
		}

		if etcdCluster.Version == "" {
			if b.IsKubernetesGTE("1.11") {
				etcdCluster.Version = "3.2.18"
			} else {
				// Preserve existing default etcd version
				etcdCluster.Version = "2.2.1"
			}
		}

		if !etcdVersionIsSupported(etcdCluster.Version) {
			if featureflag.SkipEtcdVersionCheck.Enabled() {
				klog.Warningf("etcd version %q is not known to be supported, but ignoring because of SkipEtcdVersionCheck feature flag", etcdCluster.Version)
			} else {
				klog.Warningf("unsupported etcd version %q detected; please update etcd version.  Use export KOPS_FEATURE_FLAGS=SkipEtcdVersionCheck to override this check", etcdCluster.Version)
				return fmt.Errorf("etcd version %q is not supported with etcd-manager, please specify a supported version or remove the value to use the default version.  Supported versions: %s", etcdCluster.Version, strings.Join(supportedEtcdVersions, ", "))
			}

		}
	}

	return nil
}

var supportedEtcdVersions = []string{"2.2.1", "3.1.12", "3.2.18", "3.2.24", "3.3.10", "3.3.13", "3.3.17"}

func etcdVersionIsSupported(version string) bool {
	version = strings.TrimPrefix(version, "v")
	for _, v := range supportedEtcdVersions {
		if v == version {
			return true
		}
	}
	return false
}
