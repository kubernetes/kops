/*
Copyright 2021 The Kubernetes Authors.

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

package main

import (
	"bufio"
	"context"
	"crypto/x509/pkix"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	rotateCALong = templates.LongDesc(i18n.T(`
	rotates the cluster CA.`))

	rotateCAExample = templates.Examples(i18n.T(`
	# Rotate the cluster CA
	kops rotate ca --name k8s-cluster.example.com
	`))

	rotateCAShort = i18n.T(`Rotates the cluster CA.`)
)

func NewCmdRotateCA(f *util.Factory, out io.Writer) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "ca",
		Long:    rotateCALong,
		Short:   rotateCAShort,
		Example: rotateCAExample,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			err := rootCommand.ProcessArgs(args)
			if err != nil {
				exitWithError(err)
			}

			clusterName := rootCommand.ClusterName()

			if err := RunRotateCA(ctx, f, clusterName, out); err != nil {
				exitWithError(err)
			}
		},
	}

	return cmd
}

func RunRotateCA(ctx context.Context, f *util.Factory, clusterName string, out io.Writer) error {

	cluster, err := rootCommand.Cluster(ctx)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(os.Stdin)

	contextName := cluster.ObjectMeta.Name
	clientConfig, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()

	if err != nil {
		return fmt.Errorf("failed to delete secondary key: %w", err)
	}

	if !model.UseKopsControllerForNodeBootstrap(cluster) {
		return fmt.Errorf("only clusters using kops-controller for boostrapping nodes are supported")
	}

	exportAdmin := clientConfig.Contexts[contextName].AuthInfo == contextName

	fmt.Println("This comamnd will rotate the cluster CA. It is largely safe, but be aware of the following:")
	fmt.Println(" * exporting the admin TLS credentials before this command has succeeded will break")
	fmt.Println("   your client credentials and the only way to recover is to `kops rolling update --cloudonly --yes --force`")
	fmt.Println(" * This command will rotate all nodes multiple times. This will take a while.")
	fmt.Println(" * It is safe to restart this command provided you did not export credentials manually")
	fmt.Println("")
	fmt.Println("Your cluster should be fully updated and rotated before starting this procedure.")
	fmt.Println("")
	if exportAdmin {
		fmt.Println("The admin TLS credentials was detected. We will export the admin credentials after rotating the CA")
	} else {
		fmt.Println("Could not detect any admin credentials. Assuming admin credentials are not in use.")
	}

	fmt.Println("If you understand the above, type 'yes'. Anything else will abort.")
	scanner.Scan()
	err = scanner.Err()
	if err != nil {
		exitWithError(fmt.Errorf("unable to interpret input: %w", err))
	}
	val := scanner.Text()
	val = strings.TrimSpace(val)
	val = strings.ToLower(val)
	if val != "yes" {
		exitWithError(fmt.Errorf("Aborting"))
	}

	clientset, err := f.Clientset()
	if err != nil {
		return err
	}

	keyStore, err := clientset.KeyStore(cluster)
	if err != nil {
		return err
	}

	pool, err := keyStore.FindCertificatePool(fi.CertificateIDCA)
	if err != nil {
		return fmt.Errorf("could not fetch the CA pool: %w", err)
	}

	if len(pool.Secondary) > 0 {
		klog.Info("Secondary CA cert already in the pool. Not issuing a new CA")
	} else {
		err := rotateCAIssueCert(keyStore)
		if err != nil {
			return fmt.Errorf("could not issue new CA: %w", err)
		}

		//Update the pool
		pool, err = keyStore.FindCertificatePool(fi.CertificateIDCA)
		if err != nil {
			return fmt.Errorf("could not fetch the CA pool: %w", err)
		}
	}

	// Update the cluster to trust both CAs
	err = rotateCAUpdateCluster(ctx, cluster, pool, f, out, false)
	if err != nil {
		return fmt.Errorf("failed to update the cluster: %w", err)
	}

	//Delete the old key
	klog.Info("deleting the old CA")

	keyId := pool.Secondary[0].Certificate.SerialNumber.String()
	keyset, err := keyStore.FindCertificateKeyset(fi.CertificateIDCA)
	if err != nil {
		return fmt.Errorf("failed to load keyset: %w", err)
	}

	err = keyStore.DeleteKeysetItem(keyset, keyId)
	if err != nil {
		return fmt.Errorf("failed to delete secondary key: %w", err)
	}

	if exportAdmin {
		klog.Info("Detected the admin TLS user. Will also export a new admin certificate")
	} else {
		klog.Info("Could not detect the admin TLS user. Assuming existing credentials will continue to work")
	}

	// Update the cluster one last time to trust only the new CA
	err = rotateCAUpdateCluster(ctx, cluster, pool, f, out, exportAdmin)
	if err != nil {
		return fmt.Errorf("failed to update the cluster: %w", err)
	}

	return nil
}

func rotateCAIssueCert(keyStore fi.CAStore) error {

	klog.Infof("Issuing new certificate")

	serial := pki.BuildPKISerial(time.Now().UnixNano())

	subjectPkix := &pkix.Name{
		CommonName: "kubernetes",
	}

	req := pki.IssueCertRequest{
		Signer:         fi.CertificateIDCA,
		Type:           "ca",
		Subject:        *subjectPkix,
		AlternateNames: []string{},
		Serial:         serial,
	}
	cert, privateKey, _, err := pki.IssueCert(&req, keyStore)
	if err != nil {
		return err
	}
	err = keyStore.StoreKeypair(fi.CertificateIDCA, cert, privateKey)
	if err != nil {
		return err
	}

	return nil
}

func rotateCAUpdateServiceAccounts(ctx context.Context, cluster *kops.Cluster, caBundle []byte) error {
	klog.Info("updating ServiceAccounts with a new CA bundle")

	caBundle64 := base64.StdEncoding.EncodeToString(caBundle)
	caBundle64Bytes := []byte(caBundle64)

	contextName := cluster.ObjectMeta.Name
	configLoadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		configLoadingRules,
		&clientcmd.ConfigOverrides{CurrentContext: contextName}).ClientConfig()
	if err != nil {
		return fmt.Errorf("cannot load kubecfg settings for %q: %v", contextName, err)
	}

	k8sClient, err := kubernetes.NewForConfig(config)

	if err != nil {
		return fmt.Errorf("cannot build kubernetes api client for %q: %v", contextName, err)
	}

	secretClient := k8sClient.CoreV1().Secrets("")

	secrets, err := secretClient.List(ctx, v1.ListOptions{
		FieldSelector: "type=kubernetes.io/service-account-token",
	})
	if err != nil {
		return err
	}

	for _, secret := range secrets.Items {
		secret.Data["ca.crt"] = caBundle64Bytes
		secretClient.Update(ctx, &secret, v1.UpdateOptions{})
	}

	return nil
}

func rotateCAUpdateCluster(ctx context.Context, cluster *kops.Cluster, pool *fi.CertificatePool, f *util.Factory, out io.Writer, exportAdmin bool) error {
	caBundle, err := pool.AsBytes()
	if err != nil {
		return fmt.Errorf("failed to encode ca bundle: %w", err)
	}

	//Update service accounts to trust old and new CA
	err = rotateCAUpdateServiceAccounts(ctx, cluster, caBundle)
	if err != nil {
		return fmt.Errorf("error updating ServiceAccounts: %v", err)
	}

	adminTTL, _ := time.ParseDuration("0")
	if exportAdmin {
		adminTTL, _ = time.ParseDuration("18h")
	}

	//New kubeconfig with bundled CA so we trust both new and old api servers
	RunExportKubecfg(ctx, f, out, &ExportKubecfgOptions{admin: adminTTL}, []string{})

	klog.Info("rotating all nodes")

	//Update nodes first. This will make kubelet trust new and old CA.
	ruo := &RollingUpdateOptions{}
	ruo.InitDefaults()
	ruo.Yes = true
	ruo.ClusterName = cluster.ObjectMeta.Name
	ruo.Force = true
	ruo.InstanceGroupRoles = []string{"node"}

	err = RunRollingUpdateCluster(ctx, f, out, ruo)
	if err != nil {
		return fmt.Errorf("failed to rotate cluster: %v", err)
	}

	klog.Info("rotating the control plane")

	//Update masters. This will issue new certs for k8s using the new CA.
	//New nodes, service accounts etc will use new CA
	ruo.InstanceGroupRoles = []string{"master", "apiserver"}

	err = RunRollingUpdateCluster(ctx, f, out, ruo)
	if err != nil {
		return fmt.Errorf("failed to rotate cluster: %v", err)
	}
	return nil

}
