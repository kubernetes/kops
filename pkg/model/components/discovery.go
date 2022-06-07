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

package components

import (
	"fmt"
	"strings"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
	"k8s.io/kops/util/pkg/vfs"
)

// DiscoveryOptionsBuilder adds options for identity discovery to the model (mostly kube-apiserver)
type DiscoveryOptionsBuilder struct {
	*OptionsContext
}

var _ loader.OptionsBuilder = &DiscoveryOptionsBuilder{}

func (b *DiscoveryOptionsBuilder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)

	if clusterSpec.KubeAPIServer == nil {
		clusterSpec.KubeAPIServer = &kops.KubeAPIServerConfig{}
	}

	kubeAPIServer := clusterSpec.KubeAPIServer

	if len(kubeAPIServer.APIAudiences) == 0 {
		kubeAPIServer.APIAudiences = []string{"kubernetes.svc.default"}
	}

	if kubeAPIServer.ServiceAccountIssuer == nil {
		said := clusterSpec.ServiceAccountIssuerDiscovery
		var serviceAccountIssuer string
		if said != nil && said.DiscoveryStore != "" {
			store := said.DiscoveryStore
			base, err := vfs.Context.BuildVfsPath(store)
			if err != nil {
				return fmt.Errorf("error parsing locationStore=%q: %w", store, err)
			}
			switch base := base.(type) {
			case *vfs.S3Path:
				serviceAccountIssuer, err = base.GetHTTPsUrl(clusterSpec.IsIPv6Only())
				if err != nil {
					return err
				}
			case *vfs.MemFSPath:
				if !base.IsClusterReadable() {
					// If this _is_ a test, we should call MarkClusterReadable
					return fmt.Errorf("locationStore=%q is only supported in tests", store)
				}
				serviceAccountIssuer = strings.Replace(base.Path(), "memfs://", "https://", 1)
			default:
				return fmt.Errorf("locationStore=%q is of unexpected type %T", store, base)
			}
		} else {
			if supportsPublicJWKS(clusterSpec) {
				serviceAccountIssuer = "https://" + clusterSpec.MasterPublicName
			} else {
				serviceAccountIssuer = "https://" + clusterSpec.MasterInternalName
			}
		}
		kubeAPIServer.ServiceAccountIssuer = &serviceAccountIssuer
	}
	kubeAPIServer.ServiceAccountJWKSURI = fi.String(*kubeAPIServer.ServiceAccountIssuer + "/openid/v1/jwks")
	// We set apiserver ServiceAccountKey and ServiceAccountSigningKeyFile in nodeup

	return nil
}

func supportsPublicJWKS(clusterSpec *kops.ClusterSpec) bool {
	if !fi.BoolValue(clusterSpec.KubeAPIServer.AnonymousAuth) {
		return false
	}
	for _, cidr := range clusterSpec.KubernetesAPIAccess {
		if cidr == "0.0.0.0/0" || cidr == "::/0" {
			return true
		}
	}
	return false
}
