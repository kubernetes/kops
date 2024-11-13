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

package kubeconfig

import (
	"context"
	"crypto/x509/pkix"
	"fmt"
	"net"
	"os/user"
	"sort"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/pkg/rbac"
	"k8s.io/kops/upup/pkg/fi"
)

const DefaultKubecfgAdminLifetime = 18 * time.Hour

func BuildKubecfg(ctx context.Context, cluster *kops.Cluster, keyStore fi.KeystoreReader, secretStore fi.SecretStore, cloud fi.Cloud, admin time.Duration, configUser string, internal bool, kopsStateStore string, useKopsAuthenticationPlugin bool) (*KubeconfigBuilder, error) {
	clusterName := cluster.ObjectMeta.Name

	var server string
	if internal {
		server = "https://" + cluster.APIInternalName()
	} else {
		if cluster.Spec.API.PublicName != "" {
			server = "https://" + wrapIPv6Address(cluster.Spec.API.PublicName)
		} else {
			server = "https://api." + clusterName
		}

		// If a load balancer exists we use it, except for when an SSL certificate is set.
		// This should avoid a lot of pain with DNS pre-creation.
		if cluster.Spec.API.LoadBalancer != nil && (cluster.Spec.API.LoadBalancer.SSLCertificate == "" || admin != 0) {
			ingresses, err := cloud.GetApiIngressStatus(cluster)
			if err != nil {
				return nil, fmt.Errorf("error getting ingress status: %v", err)
			}

			var targets []string

			for _, useInternalEndpoint := range []bool{false, true} {
				for _, ingress := range ingresses {
					if !useInternalEndpoint && ingress.InternalEndpoint {
						continue
					}
					if ingress.Hostname != "" {
						targets = append(targets, ingress.Hostname)
					}
					if ingress.IP != "" {
						targets = append(targets, ingress.IP)
					}
				}
				if len(targets) > 0 {
					// Prefer external addresses
					break
				}
				klog.Infof("no external API endpoints found; falling back to internal API endpoints")
			}

			sort.Strings(targets)
			if len(targets) == 0 {
				klog.Warningf("Did not find API endpoint; may not be able to reach cluster")
			} else {
				if len(targets) != 1 {
					klog.Warningf("Found multiple API endpoints (%v), choosing arbitrarily", targets)
				}
				server = "https://" + wrapIPv6Address(targets[0])
			}
		}
	}

	b := NewKubeconfigBuilder()

	// Use the secondary load balancer port if a certificate is on the primary listener
	if admin != 0 && cluster.Spec.API.LoadBalancer != nil && cluster.Spec.API.LoadBalancer.SSLCertificate != "" && cluster.Spec.API.LoadBalancer.Class == kops.LoadBalancerClassNetwork {
		server = server + ":8443"
	}

	b.Context = clusterName
	b.Server = server
	b.TLSServerName = cluster.APIInternalName()

	// add the CA Cert to the kubeconfig only if we didn't specify a certificate for the LB
	//  or if we're using admin credentials and the secondary port
	if cluster.Spec.API.LoadBalancer == nil || cluster.Spec.API.LoadBalancer.SSLCertificate == "" || cluster.Spec.API.LoadBalancer.Class == kops.LoadBalancerClassNetwork || internal {
		keySet, err := keyStore.FindKeyset(ctx, fi.CertificateIDCA)
		if err != nil {
			return nil, fmt.Errorf("error fetching CA keypair: %v", err)
		}
		if keySet != nil {
			b.CACerts, err = keySet.ToCertificateBytes()
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("cannot find CA certificate")
		}
	}

	if admin != 0 {
		cn := "kubecfg"
		user, err := user.Current()
		if err != nil || user == nil {
			klog.Infof("unable to get user: %v", err)
		} else {
			cn += "-" + user.Name
		}

		req := pki.IssueCertRequest{
			Signer: fi.CertificateIDCA,
			Type:   "client",
			Subject: pkix.Name{
				CommonName:   cn,
				Organization: []string{rbac.SystemPrivilegedGroup},
			},
			Validity: admin,
		}
		cert, privateKey, _, err := pki.IssueCert(ctx, &req, fi.NewPKIKeystoreAdapter(keyStore))
		if err != nil {
			return nil, err
		}
		b.ClientCert, err = cert.AsBytes()
		if err != nil {
			return nil, err
		}
		b.ClientKey, err = privateKey.AsBytes()
		if err != nil {
			return nil, err
		}
	}

	if useKopsAuthenticationPlugin {
		b.AuthenticationExec = []string{
			"kops",
			"helpers",
			"kubectl-auth",
			"--cluster=" + clusterName,
			"--state=" + kopsStateStore,
		}

		// If there's an existing client-cert / client-key, we need to clear it so it won't be used
		b.ClientCert = nil
		b.ClientKey = nil
	}

	b.Server = server

	if configUser == "" {
		b.User = cluster.ObjectMeta.Name
	} else {
		b.User = configUser
	}

	return b, nil
}

// wrapIPv6Address will wrap IPv6 addresses in square brackets,
// for use in URLs; other endpoints are unchanged.
func wrapIPv6Address(endpoint string) string {
	ip := net.ParseIP(endpoint)
	// IPv6 addresses are wrapped in square brackets in URLs
	if ip != nil && ip.To4() == nil {
		return "[" + endpoint + "]"
	}
	return endpoint
}
