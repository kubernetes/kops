package federation

import (
	"fmt"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kops/federation/targets/kubernetes"
	"k8s.io/kubernetes/pkg/api/errors"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/kutil"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kubernetes/pkg/client/clientset_generated/release_1_3"
)

const UserAdmin = "admin"

type FederationConfiguration struct {
	Namespace            string

	ApiserverKeypair     *fitasks.Keypair
	ApiserverServiceName string
	ApiserverSecretName  string

	KubeconfigSecretName string
}

func (o*FederationConfiguration) extractKubecfg(c *fi.Context, f *kopsapi.Federation) (*kutil.KubeconfigBuilder, error) {
	// TODO: move this
	masterName := "api." + f.Spec.DNSName

	k := kutil.NewKubeconfigBuilder()
	k.KubeMasterIP = masterName
	k.Context = "federation-" + f.Name

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

	k8s := c.Target.(*kubernetes.KubernetesTarget).KubernetesClient

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

func (o*FederationConfiguration) findBasicAuth(secret *v1.Secret) (*AuthFile, error) {
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

func (o*FederationConfiguration) findKnownTokens(secret *v1.Secret) (*AuthFile, error) {
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

func (o*FederationConfiguration) EnsureConfiguration(c *fi.Context) error {
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

	k8s := c.Target.(*kubernetes.KubernetesTarget).KubernetesClient

	adminPassword := ""
	adminToken := ""

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
					return nil,  err
				}
				adminToken = string(s.Data)
			} else {
				adminToken = u.Secret
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

	// TODO: Prefer username / password or token?
	user := kutil.KubectlUser{
		Username:UserAdmin,
		Password: adminPassword,
		//Token: adminToken,
	}
	err = o.ensureSecretKubeconfig(c, caCert, user)
	if err != nil {
		return err
	}

	return nil
}

func (o*FederationConfiguration) ensureSecretKubeconfig(c *fi.Context, caCert *fi.Certificate, user kutil.KubectlUser) error {
	k8s := c.Target.(*kubernetes.KubernetesTarget).KubernetesClient

	_, err := mutateSecret(k8s, o.Namespace, o.KubeconfigSecretName, func(s *v1.Secret) (*v1.Secret, error) {
		var kubeconfigData []byte
		var err error

		{
			kubeconfig := &kutil.KubectlConfig{
				ApiVersion: "v1",
				Kind: "Config",
			}

			cluster := &kutil.KubectlClusterWithName{
				Name: o.ApiserverServiceName,
				Cluster: kutil.KubectlCluster{
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

			kubeconfig.Clusters = append(kubeconfig.Clusters, cluster)

			user := &kutil.KubectlUserWithName{
				Name: o.ApiserverServiceName,
				User: user,
			}
			kubeconfig.Users = append(kubeconfig.Users, user)

			context := &kutil.KubectlContextWithName{
				Name: o.ApiserverServiceName,
				Context: kutil.KubectlContext{
					Cluster: cluster.Name,
					User: user.Name,
				},
			}
			kubeconfig.CurrentContext = o.ApiserverServiceName
			kubeconfig.Contexts = append(kubeconfig.Contexts, context)

			kubeconfigData, err = kopsapi.ToYaml(kubeconfig)
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

func findSecret(k8s release_1_3.Interface, namespace, name string) (*v1.Secret, error) {
	glog.V(2).Infof("querying k8s for secret %s/%s", namespace, name)
	s, err := k8s.Core().Secrets(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		} else {
			return nil, fmt.Errorf("error reading secret %s/%s: %v", namespace, name, err)
		}
	}
	return s, nil
}

func mutateSecret(k8s release_1_3.Interface, namespace string, name string, fn func(s *v1.Secret) (*v1.Secret, error)) (*v1.Secret, error) {
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
		created, err := k8s.Core().Secrets(namespace).Create(updated)
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