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
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/upup/pkg/kutil"
	"k8s.io/kubernetes/federation/apis/federation/v1beta1"
	"k8s.io/kubernetes/federation/client/clientset_generated/federation_release_1_5"
	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/clientset_generated/release_1_5"
)

type FederationCluster struct {
	FederationNamespace string

	ControllerKubernetesClients []release_1_5.Interface
	FederationClient            federation_release_1_5.Interface

	ClusterSecretName string

	ClusterName       string
	ApiserverHostname string
}

func (o *FederationCluster) Run(cluster *kopsapi.Cluster) error {
	keyStore, err := registry.KeyStore(cluster)
	if err != nil {
		return err
	}
	secretStore, err := registry.SecretStore(cluster)
	if err != nil {
		return err
	}

	k := kutil.CreateKubecfg{
		ContextName:  cluster.ObjectMeta.Name,
		KeyStore:     keyStore,
		SecretStore:  secretStore,
		KubeMasterIP: cluster.Spec.MasterPublicName,
	}

	kubeconfig, err := k.ExtractKubeconfig()
	if err != nil {
		return fmt.Errorf("error building connection information for cluster %q: %v", cluster.ObjectMeta.Name, err)
	}

	user := kutil.KubectlUser{
		ClientCertificateData: kubeconfig.ClientCert,
		ClientKeyData:         kubeconfig.ClientKey,
	}
	// username/password or bearer token may be set, but not both
	if kubeconfig.KubeBearerToken != "" {
		user.Token = kubeconfig.KubeBearerToken
	} else {
		user.Username = kubeconfig.KubeUser
		user.Password = kubeconfig.KubePassword
	}

	for _, k8s := range o.ControllerKubernetesClients {
		if err := o.ensureFederationSecret(k8s, kubeconfig.CACert, user); err != nil {
			return err
		}
	}

	if err := o.ensureFederationCluster(o.FederationClient); err != nil {
		return err
	}

	return nil
}

func (o *FederationCluster) ensureFederationSecret(k8s release_1_5.Interface, caCertData []byte, user kutil.KubectlUser) error {
	_, err := mutateSecret(k8s, o.FederationNamespace, o.ClusterSecretName, func(s *v1.Secret) (*v1.Secret, error) {
		var kubeconfigData []byte
		var err error

		{
			kubeconfig := &kutil.KubectlConfig{
				ApiVersion: "v1",
				Kind:       "Config",
			}

			cluster := &kutil.KubectlClusterWithName{
				Name: o.ClusterName,
				Cluster: kutil.KubectlCluster{
					Server: "https://" + o.ApiserverHostname,
				},
			}

			if caCertData != nil {
				cluster.Cluster.CertificateAuthorityData = caCertData
			}

			kubeconfig.Clusters = append(kubeconfig.Clusters, cluster)

			user := &kutil.KubectlUserWithName{
				Name: o.ClusterName,
				User: user,
			}
			kubeconfig.Users = append(kubeconfig.Users, user)

			context := &kutil.KubectlContextWithName{
				Name: o.ClusterName,
				Context: kutil.KubectlContext{
					Cluster: cluster.Name,
					User:    user.Name,
				},
			}
			kubeconfig.CurrentContext = o.ClusterName
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

func (o *FederationCluster) ensureFederationCluster(federationClient federation_release_1_5.Interface) error {
	_, err := mutateCluster(federationClient, o.ClusterName, func(c *v1beta1.Cluster) (*v1beta1.Cluster, error) {
		if c == nil {
			c = &v1beta1.Cluster{}
		}

		// How to connect to the member cluster
		c.Spec.ServerAddressByClientCIDRs = []v1beta1.ServerAddressByClientCIDR{
			{
				// The CIDR with which clients can match their IP to figure out the server address that they should use.
				ClientCIDR: "0.0.0.0/0",
				// Address of this server, suitable for a client that matches the above CIDR.
				// This can be a hostname, hostname:port, IP or IP:port.
				ServerAddress: "https://" + o.ApiserverHostname,
			},
		}

		// Secret containing credentials for connecting to cluster
		c.Spec.SecretRef = &v1.LocalObjectReference{
			Name: o.ClusterSecretName,
		}
		return c, nil
	})

	return err
}

func findCluster(k8s federation_release_1_5.Interface, name string) (*v1beta1.Cluster, error) {
	glog.V(2).Infof("querying k8s for federation cluster %s", name)
	c, err := k8s.Federation().Clusters().Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		} else {
			return nil, fmt.Errorf("error reading federation cluster %s: %v", name, err)
		}
	}
	return c, nil
}

func mutateCluster(k8s federation_release_1_5.Interface, name string, fn func(s *v1beta1.Cluster) (*v1beta1.Cluster, error)) (*v1beta1.Cluster, error) {
	existing, err := findCluster(k8s, name)
	if err != nil {
		return nil, err
	}
	createObject := existing == nil
	updated, err := fn(existing)
	if err != nil {
		return nil, err
	}

	updated.Name = name

	if createObject {
		glog.V(2).Infof("creating federation cluster %s", name)
		created, err := k8s.Federation().Clusters().Create(updated)
		if err != nil {
			return nil, fmt.Errorf("error creating federation cluster %s: %v", name, err)
		}
		return created, nil
	} else {
		glog.V(2).Infof("updating federation cluster %s", name)
		created, err := k8s.Federation().Clusters().Update(updated)
		if err != nil {
			return nil, fmt.Errorf("error updating federation cluster %s: %v", name, err)
		}
		return created, nil
	}
}
