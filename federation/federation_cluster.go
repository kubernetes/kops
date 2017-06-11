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
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kubernetes/federation/apis/federation/v1beta1"
	"k8s.io/kubernetes/federation/client/clientset_generated/federation_clientset"
	k8sapiv1 "k8s.io/kubernetes/pkg/api/v1"
)

type FederationCluster struct {
	FederationNamespace string

	ControllerKubernetesClients []kubernetes.Interface
	FederationClient            federation_clientset.Interface

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

	status := &kopsapi.NoopStatusStore{}
	conf, err := kubeconfig.BuildKubecfg(cluster, keyStore, secretStore, status)
	if err != nil {
		return fmt.Errorf("error building connection information for cluster %q: %v", cluster.ObjectMeta.Name, err)
	}

	user := kubeconfig.KubectlUser{
		ClientCertificateData: conf.ClientCert,
		ClientKeyData:         conf.ClientKey,
	}
	// username/password or bearer token may be set, but not both
	if conf.KubeBearerToken != "" {
		user.Token = conf.KubeBearerToken
	} else {
		user.Username = conf.KubeUser
		user.Password = conf.KubePassword
	}

	for _, k8s := range o.ControllerKubernetesClients {
		if err := o.ensureFederationSecret(k8s, conf.CACert, user); err != nil {
			return err
		}
	}

	if err := o.ensureFederationCluster(o.FederationClient); err != nil {
		return err
	}

	return nil
}

func (o *FederationCluster) ensureFederationSecret(k8s kubernetes.Interface, caCertData []byte, user kubeconfig.KubectlUser) error {
	_, err := mutateSecret(k8s, o.FederationNamespace, o.ClusterSecretName, func(s *v1.Secret) (*v1.Secret, error) {
		var kubeconfigData []byte
		var err error

		{
			conf := &kubeconfig.KubectlConfig{
				ApiVersion: "v1",
				Kind:       "Config",
			}

			cluster := &kubeconfig.KubectlClusterWithName{
				Name: o.ClusterName,
				Cluster: kubeconfig.KubectlCluster{
					Server: "https://" + o.ApiserverHostname,
				},
			}

			if caCertData != nil {
				cluster.Cluster.CertificateAuthorityData = caCertData
			}

			conf.Clusters = append(conf.Clusters, cluster)

			user := &kubeconfig.KubectlUserWithName{
				Name: o.ClusterName,
				User: user,
			}
			conf.Users = append(conf.Users, user)

			context := &kubeconfig.KubectlContextWithName{
				Name: o.ClusterName,
				Context: kubeconfig.KubectlContext{
					Cluster: cluster.Name,
					User:    user.Name,
				},
			}
			conf.CurrentContext = o.ClusterName
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

func (o *FederationCluster) ensureFederationCluster(federationClient federation_clientset.Interface) error {
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
		c.Spec.SecretRef = &k8sapiv1.LocalObjectReference{
			Name: o.ClusterSecretName,
		}
		return c, nil
	})

	return err
}

func findCluster(k8s federation_clientset.Interface, name string) (*v1beta1.Cluster, error) {
	glog.V(2).Infof("querying k8s for federation cluster %s", name)
	c, err := k8s.Federation().Clusters().Get(name, meta_v1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		} else {
			return nil, fmt.Errorf("error reading federation cluster %s: %v", name, err)
		}
	}
	return c, nil
}

func mutateCluster(k8s federation_clientset.Interface, name string, fn func(s *v1beta1.Cluster) (*v1beta1.Cluster, error)) (*v1beta1.Cluster, error) {
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
