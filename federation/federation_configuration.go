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

package federation

import (
	"fmt"
	"github.com/golang/glog"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/kops/federation/targets/kubernetestarget"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"
)

const UserAdmin = "admin"

type FederationConfiguration struct {
	Namespace string

	ApiserverKeypair     *fitasks.Keypair
	ApiserverServiceName string
	ApiserverSecretName  string

	KubeconfigSecretName string
}

func (o *FederationConfiguration) extractKubecfg(c *fi.Context, f *kopsapi.Federation) (*kubeconfig.KubeconfigBuilder, error) {
	// TODO: move this
	masterName := "api." + f.Spec.DNSName

	k := kubeconfig.NewKubeconfigBuilder()
	k.Server = "https://" + masterName
	k.Context = "federation-" + f.ObjectMeta.Name

	// CA Cert
	caCert, _, err := c.Keystore.FindKeypair(fi.CertificateId_CA)
	if err != nil {
		return nil, err
	}
	if caCert == nil {
		glog.Infof("No CA certificate in cluster %q", c)
		return nil, nil
	}
	k.CACert, err = caCert.AsBytes()
	if err != nil {
		return nil, err
	}

	k8s := c.Target.(*kubernetestarget.KubernetesTarget).KubernetesClient

	// Basic auth
	secret, err := findSecret(k8s, o.Namespace, o.ApiserverSecretName)
	if err != nil {
		return nil, err
	}
	if secret == nil {
		glog.Infof("No federation configuration in cluster %q", c)
		return nil, nil
	}

	{
		basicAuthData, err := o.findBasicAuth(secret)
		if err != nil {
			return nil, err
		}
		if basicAuthData == nil {
			glog.Infof("No auth data in cluster %q", c)
			return nil, nil
		}
		user := basicAuthData.FindUser(UserAdmin)
		if user == nil {
			glog.Infof("No auth data for user %q in cluster %q", UserAdmin, c)
			return nil, nil
		}
		k.KubeUser = user.User
		k.KubePassword = user.Secret
	}

	{
		knownTokens, err := o.findKnownTokens(secret)
		if err != nil {
			return nil, err
		}
		if knownTokens == nil {
			glog.Infof("No token data in cluster %q", c)
			return nil, nil
		}
		user := knownTokens.FindUser(UserAdmin)
		if user == nil {
			glog.Infof("No token data for user %q in cluster %q", UserAdmin, c)
			return nil, nil
		}
		k.KubeBearerToken = user.Secret
	}

	return k, nil
}

func (o *FederationConfiguration) findBasicAuth(secret *v1.Secret) (*AuthFile, error) {
	var basicAuthData *AuthFile
	var err error

	if secret == nil {
		return nil, nil
	}

	if secret.Data["basic-auth.csv"] != nil {
		basicAuthData, err = ParseAuthFile(secret.Data["basic-auth.csv"])
		if err != nil {
			return nil, fmt.Errorf("error parsing auth file basic-auth.csv in secret %s/%s: %v", secret.Namespace, secret.Name, err)
		}
	}

	return basicAuthData, nil
}

func (o *FederationConfiguration) findKnownTokens(secret *v1.Secret) (*AuthFile, error) {
	var knownTokens *AuthFile
	var err error

	if secret == nil {
		return nil, nil
	}

	if secret.Data["known-tokens.csv"] != nil {
		knownTokens, err = ParseAuthFile(secret.Data["known-tokens.csv"])
		if err != nil {
			return nil, fmt.Errorf("error parsing auth file known-tokens.csv in secret %s/%s: %v", secret.Namespace, secret.Name, err)
		}
	}

	return knownTokens, nil
}

func (o *FederationConfiguration) EnsureConfiguration(c *fi.Context) error {
	caCert, _, err := c.Keystore.FindKeypair(fi.CertificateId_CA)
	if err != nil {
		return err
	}
	if caCert == nil {
		return fmt.Errorf("cannot find CA certificate")
	}

	serverCert, serverKey, err := c.Keystore.FindKeypair(fi.StringValue(o.ApiserverKeypair.Name))
	if err != nil {
		return err
	}
	if serverCert == nil || serverKey == nil {
		return fmt.Errorf("cannot find server keypair")
	}

	k8s := c.Target.(*kubernetestarget.KubernetesTarget).KubernetesClient

	adminPassword := ""
	//adminToken := ""

	_, err = mutateSecret(k8s, o.Namespace, o.ApiserverSecretName, func(s *v1.Secret) (*v1.Secret, error) {
		basicAuthData, err := o.findBasicAuth(s)
		if err != nil {
			return nil, err
		}

		knownTokens, err := o.findKnownTokens(s)
		if err != nil {
			return nil, err
		}

		{
			if basicAuthData == nil {
				basicAuthData = &AuthFile{}
			}
			u := basicAuthData.FindUser(UserAdmin)
			if u == nil {
				s, err := fi.CreateSecret()
				if err != nil {
					return nil, err
				}
				err = basicAuthData.Add(&AuthFileLine{User: UserAdmin, Secret: string(s.Data), Role: "admin"})
				if err != nil {
					return nil, err
				}
				adminPassword = string(s.Data)
			} else {
				adminPassword = u.Secret
			}
		}

		{
			if knownTokens == nil {
				knownTokens = &AuthFile{}
			}
			u := knownTokens.FindUser(UserAdmin)
			if u == nil {
				s, err := fi.CreateSecret()
				if err != nil {
					return nil, err
				}
				err = knownTokens.Add(&AuthFileLine{User: UserAdmin, Secret: string(s.Data), Role: "admin"})
				if err != nil {
					return nil, err
				}
				//adminToken = string(s.Data)
			} else {
				//adminToken = u.Secret
			}
		}

		if s == nil {
			s = &v1.Secret{}
			s.Type = v1.SecretTypeOpaque
		}
		if s.Data == nil {
			s.Data = make(map[string][]byte)
		}

		{
			b, err := caCert.AsBytes()
			if err != nil {
				return nil, err
			}
			s.Data["ca.crt"] = b
		}
		{
			b, err := serverCert.AsBytes()
			if err != nil {
				return nil, err
			}
			s.Data["server.cert"] = b
		}
		{
			b, err := serverKey.AsBytes()
			if err != nil {
				return nil, err
			}
			s.Data["server.key"] = b
		}

		s.Data["basic-auth.csv"] = []byte(basicAuthData.Encode())
		s.Data["known-tokens.csv"] = []byte(knownTokens.Encode())

		return s, nil
	})
	if err != nil {
		return fmt.Errorf("error mutating secret: %s", err)
	}
	// TODO: Prefer username / password or token?
	user := kubeconfig.KubectlUser{
		Username: UserAdmin,
		Password: adminPassword,
		//Token: adminToken,
	}
	err = o.ensureSecretKubeconfig(c, caCert, user)
	if err != nil {
		return err
	}

	return nil
}

func (o *FederationConfiguration) ensureSecretKubeconfig(c *fi.Context, caCert *fi.Certificate, user kubeconfig.KubectlUser) error {
	k8s := c.Target.(*kubernetestarget.KubernetesTarget).KubernetesClient

	_, err := mutateSecret(k8s, o.Namespace, o.KubeconfigSecretName, func(s *v1.Secret) (*v1.Secret, error) {
		var kubeconfigData []byte
		var err error

		{
			conf := &kubeconfig.KubectlConfig{
				ApiVersion: "v1",
				Kind:       "Config",
			}

			cluster := &kubeconfig.KubectlClusterWithName{
				Name: o.ApiserverServiceName,
				Cluster: kubeconfig.KubectlCluster{
					Server: "https://" + o.ApiserverServiceName,
				},
			}

			if caCert != nil {
				caCertData, err := caCert.AsBytes()
				if err != nil {
					return nil, err
				}
				cluster.Cluster.CertificateAuthorityData = caCertData
			}

			conf.Clusters = append(conf.Clusters, cluster)

			user := &kubeconfig.KubectlUserWithName{
				Name: o.ApiserverServiceName,
				User: user,
			}
			conf.Users = append(conf.Users, user)

			context := &kubeconfig.KubectlContextWithName{
				Name: o.ApiserverServiceName,
				Context: kubeconfig.KubectlContext{
					Cluster: cluster.Name,
					User:    user.Name,
				},
			}
			conf.CurrentContext = o.ApiserverServiceName
			conf.Contexts = append(conf.Contexts, context)

			kubeconfigData, err = kopsapi.ToRawYaml(conf)
			if err != nil {
				return nil, fmt.Errorf("error building kubeconfig: %v", err)
			}
		}

		if s == nil {
			s = &v1.Secret{}
			s.Type = v1.SecretTypeOpaque
		}
		if s.Data == nil {
			s.Data = make(map[string][]byte)
		}

		s.Data["kubeconfig"] = kubeconfigData
		return s, nil
	})

	return err
}

func findSecret(k8s kubernetes.Interface, namespace, name string) (*v1.Secret, error) {
	glog.V(2).Infof("querying k8s for secret %s/%s", namespace, name)
	s, err := k8s.CoreV1().Secrets(namespace).Get(name, meta_v1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		} else {
			return nil, fmt.Errorf("error reading secret %s/%s: %v", namespace, name, err)
		}
	}
	return s, nil
}

func mutateSecret(k8s kubernetes.Interface, namespace string, name string, fn func(s *v1.Secret) (*v1.Secret, error)) (*v1.Secret, error) {
	existing, err := findSecret(k8s, namespace, name)
	if err != nil {
		return nil, err
	}
	createObject := existing == nil
	updated, err := fn(existing)
	if err != nil {
		return nil, err
	}

	updated.Namespace = namespace
	updated.Name = name

	if createObject {
		glog.V(2).Infof("creating k8s secret %s/%s", namespace, name)
		created, err := k8s.CoreV1().Secrets(namespace).Create(updated)
		if err != nil {
			return nil, fmt.Errorf("error creating secret %s/%s: %v", namespace, name, err)
		}
		return created, nil
	} else {
		// TODO: Check dirty?
		glog.V(2).Infof("updating k8s secret %s/%s", namespace, name)
		updated, err := k8s.Core().Secrets(namespace).Update(updated)
		if err != nil {
			return nil, fmt.Errorf("error updating secret %s/%s: %v", namespace, name, err)
		}
		return updated, nil
	}
}
