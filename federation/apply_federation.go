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
	"bytes"
	crypto_rand "crypto/rand"
	"crypto/rsa"
	"fmt"
	"strings"
	"text/template"

	"github.com/golang/glog"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/kops/federation/model"
	"k8s.io/kops/federation/targets/kubernetestarget"
	"k8s.io/kops/federation/tasks"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kops/upup/pkg/fi/k8sapi"
	"k8s.io/kubernetes/federation/client/clientset_generated/federation_clientset"
	k8sapiv1 "k8s.io/kubernetes/pkg/api/v1"
)

type ApplyFederationOperation struct {
	Federation *kopsapi.Federation
	KopsClient simple.Clientset

	namespace string
	name      string

	apiserverDeploymentName string
	apiserverServiceName    string
	apiserverHostName       string
	dnsZoneName             string
	apiserverSecretName     string
}

func (o *ApplyFederationOperation) FindKubecfg() (*kubeconfig.KubeconfigBuilder, error) {
	// TODO: Only if not yet set?
	//	hasKubecfg, err := hasKubecfg(f.Name)
	//	if err != nil {
	//		glog.Warningf("error reading kubecfg: %v", err)
	//		hasKubecfg = true
	//	}

	// Loop through looking for a configured cluster
	for _, controller := range o.Federation.Spec.Controllers {
		cluster, err := o.KopsClient.GetCluster(controller)
		if err != nil {
			return nil, fmt.Errorf("error reading cluster %q: %v", controller, err)
		}
		if cluster == nil {
			return nil, fmt.Errorf("cluster %q not found", controller)
		}

		context, err := o.federationContextForCluster(cluster)
		if err != nil {
			return nil, err
		}

		apiserverKeypair := o.buildApiserverKeypair()

		federationConfiguration := &FederationConfiguration{
			Namespace:            o.namespace,
			ApiserverSecretName:  o.apiserverSecretName,
			ApiserverServiceName: o.apiserverServiceName,
			ApiserverKeypair:     apiserverKeypair,
			KubeconfigSecretName: "federation-apiserver-kubeconfig",
		}
		k, err := federationConfiguration.extractKubecfg(context, o.Federation)
		if err != nil {
			return nil, err
		}
		if k == nil {
			continue
		}

		return k, nil
	}

	return nil, nil
}

func (o *ApplyFederationOperation) Run() error {
	o.namespace = "federation"
	o.name = "federation"

	o.apiserverDeploymentName = "federation-apiserver"
	o.apiserverServiceName = o.apiserverDeploymentName
	o.apiserverSecretName = "federation-apiserver-secrets"

	o.dnsZoneName = o.Federation.Spec.DNSName

	o.apiserverHostName = "api." + o.dnsZoneName

	// TODO: sync clusters

	var controllerKubernetesClients []kubernetes.Interface
	for _, controller := range o.Federation.Spec.Controllers {
		cluster, err := o.KopsClient.GetCluster(controller)
		if err != nil {
			return fmt.Errorf("error reading cluster %q: %v", controller, err)
		}
		if cluster == nil {
			return fmt.Errorf("cluster %q not found", controller)
		}

		context, err := o.federationContextForCluster(cluster)
		if err != nil {
			return err
		}

		err = o.runOnCluster(context, cluster)
		if err != nil {
			return err
		}

		k8s := context.Target.(*kubernetestarget.KubernetesTarget).KubernetesClient
		controllerKubernetesClients = append(controllerKubernetesClients, k8s)
	}

	federationKubecfg, err := o.FindKubecfg()
	if err != nil {
		return err
	}
	federationRestConfig, err := federationKubecfg.BuildRestConfig()
	if err != nil {
		return err
	}
	federationControllerClient, err := federation_clientset.NewForConfig(federationRestConfig)
	if err != nil {
		return err
	}

	for _, member := range o.Federation.Spec.Members {
		glog.V(2).Infof("configuring member cluster %q", member)
		cluster, err := o.KopsClient.GetCluster(member)
		if err != nil {
			return fmt.Errorf("error reading cluster %q: %v", member, err)
		}
		if cluster == nil {
			return fmt.Errorf("cluster %q not found", member)
		}

		clusterName := strings.Replace(cluster.Name, ".", "-", -1)

		a := &FederationCluster{
			FederationNamespace: o.namespace,

			ControllerKubernetesClients: controllerKubernetesClients,
			FederationClient:            federationControllerClient,

			ClusterSecretName: "secret-" + cluster.Name,
			ClusterName:       clusterName,
			ApiserverHostname: cluster.Spec.MasterPublicName,
		}
		err = a.Run(cluster)
		if err != nil {
			return err
		}
	}

	// Create default namespace
	glog.V(2).Infof("Ensuring default namespace exists")
	if _, err := o.ensureFederationNamespace(federationControllerClient, "default"); err != nil {
		return err
	}

	return nil
}

// Builds a fi.Context applying to the federation namespace in the specified cluster
// Note that this operates inside the cluster, for example the KeyStore is backed by secrets in the namespace
func (o *ApplyFederationOperation) federationContextForCluster(cluster *kopsapi.Cluster) (*fi.Context, error) {
	clusterKeystore, err := registry.KeyStore(cluster)
	if err != nil {
		return nil, err
	}

	target, err := kubernetestarget.NewKubernetesTarget(o.KopsClient, clusterKeystore, cluster)
	if err != nil {
		return nil, err
	}

	federationKeystore := k8sapi.NewKubernetesKeystore(target.KubernetesClient, o.namespace)

	checkExisting := true
	context, err := fi.NewContext(target, nil, federationKeystore, nil, nil, checkExisting, nil)
	if err != nil {
		return nil, err
	}
	return context, nil
}

func (o *ApplyFederationOperation) buildApiserverKeypair() *fitasks.Keypair {
	keypairName := "secret-" + o.apiserverHostName
	keypair := &fitasks.Keypair{
		Name:    fi.String(keypairName),
		Subject: "cn=" + o.Federation.Name,
		Type:    "server",
	}

	// So it has a valid cert inside the cluster
	if o.apiserverServiceName != "" {
		keypair.AlternateNames = append(keypair.AlternateNames, o.apiserverServiceName)
	}

	// So it has a valid cert outside the cluster
	if o.apiserverHostName != "" {
		keypair.AlternateNames = append(keypair.AlternateNames, o.apiserverHostName)
	}

	return keypair
}

func (o *ApplyFederationOperation) runOnCluster(context *fi.Context, cluster *kopsapi.Cluster) error {
	_, _, err := EnsureCASecret(context.Keystore)
	if err != nil {
		return err
	}

	apiserverKeypair := o.buildApiserverKeypair()

	err = apiserverKeypair.Run(context)
	if err != nil {
		return err
	}

	err = o.EnsureNamespace(context)
	if err != nil {
		return err
	}

	federationConfiguration := &FederationConfiguration{
		ApiserverServiceName: o.apiserverServiceName,
		Namespace:            o.namespace,
		ApiserverSecretName:  o.apiserverSecretName,
		ApiserverKeypair:     apiserverKeypair,
		KubeconfigSecretName: "federation-apiserver-kubeconfig",
	}
	err = federationConfiguration.EnsureConfiguration(context)
	if err != nil {
		return err
	}

	templateData, err := model.Asset("manifest.yaml")
	if err != nil {
		return fmt.Errorf("error loading manifest: %v", err)
	}
	manifest, err := o.executeTemplate("manifest", string(templateData))
	if err != nil {
		return fmt.Errorf("error expanding manifest template: %v", err)
	}

	applyManifestTask := tasks.KubernetesResource{
		Name:     fi.String(o.name),
		Manifest: fi.WrapResource(fi.NewStringResource(manifest)),
	}
	err = applyManifestTask.Run(context)
	if err != nil {
		return err
	}

	return nil
}

func (o *ApplyFederationOperation) buildTemplateData() map[string]string {
	namespace := o.namespace
	name := o.name

	dnsZoneName := o.dnsZoneName

	apiserverHostname := o.apiserverHostName

	// The names of the k8s apiserver & controller-manager objects
	apiserverDeploymentName := "federation-apiserver"
	controllerDeploymentName := "federation-controller-manager"

	imageRepo := "gcr.io/google_containers/hyperkube-amd64"
	imageTag := "v1.4.0"

	federationDNSProvider := "aws-route53"
	federationDNSProviderConfig := ""

	// TODO: define exactly what these do...
	serviceCIDR := "10.10.0.0/24"
	federationAdmissionControl := "NamespaceLifecycle"

	data := make(map[string]string)
	data["FEDERATION_NAMESPACE"] = namespace
	data["FEDERATION_NAME"] = name

	data["FEDERATION_APISERVER_DEPLOYMENT_NAME"] = apiserverDeploymentName
	data["FEDERATION_CONTROLLER_MANAGER_DEPLOYMENT_NAME"] = controllerDeploymentName

	data["FEDERATION_APISERVER_IMAGE_REPO"] = imageRepo
	data["FEDERATION_APISERVER_IMAGE_TAG"] = imageTag
	data["FEDERATION_CONTROLLER_MANAGER_IMAGE_REPO"] = imageRepo
	data["FEDERATION_CONTROLLER_MANAGER_IMAGE_TAG"] = imageTag

	data["FEDERATION_SERVICE_CIDR"] = serviceCIDR
	data["EXTERNAL_HOSTNAME"] = apiserverHostname
	data["FEDERATION_ADMISSION_CONTROL"] = federationAdmissionControl

	data["FEDERATION_DNS_PROVIDER"] = federationDNSProvider
	data["FEDERATION_DNS_PROVIDER_CONFIG"] = federationDNSProviderConfig

	data["DNS_ZONE_NAME"] = dnsZoneName

	return data
}

func (o *ApplyFederationOperation) executeTemplate(key string, templateDefinition string) (string, error) {
	data := o.buildTemplateData()

	t := template.New(key)

	funcMap := make(template.FuncMap)
	//funcMap["Args"] = func() []string {
	//	return args
	//}
	//funcMap["RenderResource"] = func(resourceName string, args []string) (string, error) {
	//	return l.renderResource(resourceName, args)
	//}
	//for k, fn := range l.TemplateFunctions {
	//	funcMap[k] = fn
	//}
	t.Funcs(funcMap)

	t.Option("missingkey=zero")

	_, err := t.Parse(templateDefinition)
	if err != nil {
		return "", fmt.Errorf("error parsing template %q: %v", key, err)
	}

	var buffer bytes.Buffer
	err = t.ExecuteTemplate(&buffer, key, data)
	if err != nil {
		return "", fmt.Errorf("error executing template %q: %v", key, err)
	}

	return buffer.String(), nil
}

func (o *ApplyFederationOperation) EnsureNamespace(c *fi.Context) error {
	k8s := c.Target.(*kubernetestarget.KubernetesTarget).KubernetesClient

	ns, err := k8s.CoreV1().Namespaces().Get(o.namespace, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			ns = nil
		} else {
			return fmt.Errorf("error reading namespace: %v", err)
		}
	}
	if ns == nil {
		ns = &v1.Namespace{}
		ns.Name = o.namespace
		ns, err = k8s.CoreV1().Namespaces().Create(ns)
		if err != nil {
			return fmt.Errorf("error creating namespace: %v", err)
		}
	}

	return nil
}

func (o *ApplyFederationOperation) ensureFederationNamespace(k8s federation_clientset.Interface, name string) (*k8sapiv1.Namespace, error) {
	return mutateNamespace(k8s, name, func(n *k8sapiv1.Namespace) (*k8sapiv1.Namespace, error) {
		if n == nil {
			n = &k8sapiv1.Namespace{}
			n.Name = name
		}
		return n, nil
	})
}

func EnsureCASecret(keystore fi.Keystore) (*fi.Certificate, *fi.PrivateKey, error) {
	id := fi.CertificateId_CA
	caCert, caPrivateKey, err := keystore.FindKeypair(id)
	if err != nil {
		return nil, nil, err
	}
	if caPrivateKey == nil {
		template := fi.BuildCAX509Template()
		caRsaKey, err := rsa.GenerateKey(crypto_rand.Reader, 2048)
		if err != nil {
			return nil, nil, fmt.Errorf("error generating RSA private key: %v", err)
		}

		caPrivateKey = &fi.PrivateKey{Key: caRsaKey}

		caCert, err = fi.SignNewCertificate(caPrivateKey, template, nil, nil)
		if err != nil {
			return nil, nil, err
		}

		err = keystore.StoreKeypair(id, caCert, caPrivateKey)
		if err != nil {
			return nil, nil, err
		}
	}
	return caCert, caPrivateKey, nil
}
