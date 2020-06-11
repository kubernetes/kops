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
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/upup/pkg/fi/loader"
	"k8s.io/kops/util/pkg/vfs"
)

// DiscoveryOptionsBuilder adds options for identity discovery to the model (mostly kube-apiserver)
type DiscoveryOptionsBuilder struct {
	Context *OptionsContext
}

var _ loader.OptionsBuilder = &DiscoveryOptionsBuilder{}

func (b *DiscoveryOptionsBuilder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)

	options := o.(*kops.ClusterSpec)

	if options.Discovery == nil || options.Discovery.Base == "" {
		return nil
	}

	base, err := vfs.Context.BuildVfsPath(clusterSpec.Discovery.Base)
	if err != nil {
		return fmt.Errorf("cannot parse VFS path %q: %v", clusterSpec.Discovery.Base, err)
	}

	p := base.Join(b.Context.ClusterName)

	if !vfs.IsClusterReadable(p) {
		return fmt.Errorf("discovery path %q is not cluster readable", p)
	}

	if options.KubeAPIServer == nil {
		options.KubeAPIServer = &kops.KubeAPIServerConfig{}
	}

	kubeAPIServer := options.KubeAPIServer

	if kubeAPIServer.FeatureGates == nil {
		kubeAPIServer.FeatureGates = make(map[string]string)
	}
	kubeAPIServer.FeatureGates["ServiceAccountIssuerDiscovery"] = "true"

	if len(kubeAPIServer.APIAudiences) == 0 {
		kubeAPIServer.APIAudiences = []string{"kubernetes.svc.default"}
	}

	if kubeAPIServer.ServiceAccountIssuer == nil {
		serviceAccountIssuer, err := iam.ServiceAccountIssuer(b.Context.ClusterName, clusterSpec)
		if err != nil {
			return err
		}
		kubeAPIServer.ServiceAccountIssuer = &serviceAccountIssuer
	}

	if kubeAPIServer.ServiceAccountJWKSURI == nil {
		jwksURL := *kubeAPIServer.ServiceAccountIssuer
		jwksURL = strings.TrimSuffix(jwksURL, "/") + "/openid/v1/jwks"

		kubeAPIServer.ServiceAccountJWKSURI = &jwksURL
	}

	if kubeAPIServer.ServiceAccountSigningKeyFile == nil {
		s := "/srv/kubernetes/server.key"
		kubeAPIServer.ServiceAccountSigningKeyFile = &s
	}

	if len(kubeAPIServer.ServiceAccountKeyFile) == 0 {
		kubeAPIServer.ServiceAccountKeyFile = []string{"/srv/kubernetes/server.key"}
	}

	if clusterSpec.ServiceOIDCProvider == nil {
		clusterSpec.ServiceOIDCProvider = &kops.ServiceOIDCProviderSpec{}
	}

	if clusterSpec.ServiceOIDCProvider.IssuerCAThumbprints == nil {
		// TODO: Don't hard code?  But we also want some protection against spoofing...  But presumably we can rely on our CA store
		s3RootCA := "a9d53002e97e00e043244f3d170d6f4c414104fd"
		clusterSpec.ServiceOIDCProvider.IssuerCAThumbprints = []string{s3RootCA}

		// TODO: Support GCS as store?

		// To obtain a certificate manually:
		//
		// BUCKET=some-bucket
		// echo | openssl s_client -servername ${BUCKET}.s3.amazonaws.com -showcerts -connect ${BUCKET}.s3.amazonaws.com:443
		// Copy root (last) certificate, write to /tmp/cert
		// cat /tmp/cert | openssl x509 -fingerprint -noout | sed -e s/://g
		// Take the fingerprint, convert to lower case
	}

	if clusterSpec.ServiceOIDCProvider.IssuerURL == "" {
		clusterSpec.ServiceOIDCProvider.IssuerURL = *kubeAPIServer.ServiceAccountIssuer
	}

	return nil
}
