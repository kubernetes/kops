/*
Copyright 2020 The Kubernetes Authors.

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

package helpers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/pkg/rbac"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubectl/pkg/util/i18n"
)

var (
	kubectlAuthShort = i18n.T(`kubectl authentication plugin`)
)

// HelperKubectlAuthOptions holds the options for generating an authentication token
type HelperKubectlAuthOptions struct {
	// ClusterName is the name of the cluster we are targeting
	ClusterName string

	// Lifetime specifies the desired duration of the credential
	Lifetime time.Duration

	// APIVersion specifies the version of the client.authentication.k8s.io schema in use
	APIVersion string
}

// InitDefaults populates the default values of options
func (o *HelperKubectlAuthOptions) InitDefaults() {
	o.Lifetime = 1 * time.Hour
	o.APIVersion = "v1beta1"
}

// NewCmdHelperKubectlAuth builds a cobra command for the kubectl-auth command
func NewCmdHelperKubectlAuth(f *util.Factory, out io.Writer) *cobra.Command {
	options := &HelperKubectlAuthOptions{}
	options.InitDefaults()

	cmd := &cobra.Command{
		Use:   "kubectl-auth",
		Short: kubectlAuthShort,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.TODO()

			err := RunKubectlAuthHelper(ctx, f, out, options)
			if err != nil {
				commandutils.ExitWithError(err)
			}
		},
	}

	cmd.Flags().StringVar(&options.APIVersion, "api-version", options.APIVersion, "version of client.authentication.k8s.io schema in use")
	cmd.Flags().StringVar(&options.ClusterName, "cluster", options.ClusterName, "cluster to target")
	cmd.Flags().DurationVar(&options.Lifetime, "lifetime", options.Lifetime, "lifetime of the credential to issue")

	return cmd
}

// RunKubectlAuthHelper implements the kubectl auth helper, which creates an authentication token
func RunKubectlAuthHelper(ctx context.Context, f *util.Factory, out io.Writer, options *HelperKubectlAuthOptions) error {
	if options.ClusterName == "" {
		return fmt.Errorf("ClusterName is required")
	}

	execCredential := &ExecCredential{
		Kind: "ExecCredential",
	}

	switch options.APIVersion {
	case "":
		return fmt.Errorf("api-version must be specified")
	case "v1alpha1":
		execCredential.APIVersion = "client.authentication.k8s.io/v1alpha1"
	case "v1beta1":
		execCredential.APIVersion = "client.authentication.k8s.io/v1beta1"

	default:
		return fmt.Errorf("api-version %q is not supported", options.APIVersion)
	}

	cacheFilePath := cacheFilePath(f.KopsStateStore(), options.ClusterName)
	cached, err := loadCachedExecCredential(cacheFilePath)
	if err != nil {
		klog.Infof("cached credential %q was not valid: %v", cacheFilePath, err)
		cached = nil
	}

	if cached != nil && cached.APIVersion != execCredential.APIVersion {
		klog.Infof("cached credential had wrong api version")
		cached = nil
	}

	isCached := false
	if cached != nil {
		execCredential = cached
		isCached = true
	} else {
		status, err := buildCredentials(ctx, f, options)
		if err != nil {
			return err
		}
		execCredential.Status = *status
	}

	b, err := json.MarshalIndent(execCredential, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling json: %v", err)
	}
	_, err = out.Write(b)
	if err != nil {
		return fmt.Errorf("error writing to stdout: %v", err)
	}

	if !isCached {
		if err := os.MkdirAll(filepath.Dir(cacheFilePath), 0755); err != nil {
			klog.Warningf("failed to make cache directory for %q: %v", cacheFilePath, err)
		}
		if err := ioutil.WriteFile(cacheFilePath, b, 0600); err != nil {
			klog.Warningf("failed to write cache file %q: %v", cacheFilePath, err)
		}
	}

	return nil
}

// ExecCredential specifies the client.authentication.k8s.io ExecCredential object
type ExecCredential struct {
	APIVersion string               `json:"apiVersion,omitempty"`
	Kind       string               `json:"kind,omitempty"`
	Status     ExecCredentialStatus `json:"status"`
}

// ExecCredentialStatus specifies the status of the client.authentication.k8s.io ExecCredential object
type ExecCredentialStatus struct {
	ClientCertificateData string    `json:"clientCertificateData,omitempty"`
	ClientKeyData         string    `json:"clientKeyData,omitempty"`
	ExpirationTimestamp   time.Time `json:"expirationTimestamp,omitempty"`
}

func cacheFilePath(kopsStateStore string, clusterName string) string {
	var b bytes.Buffer
	b.WriteString(kopsStateStore)
	b.WriteByte(0)
	b.WriteString(clusterName)
	b.WriteByte(0)

	hash := fmt.Sprintf("%x", sha256.New().Sum(b.Bytes()))
	sanitizedName := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		default:
			return '_'
		}
	}, clusterName)
	return filepath.Join(homedir.HomeDir(), ".kube", "cache", "kops-authentication", sanitizedName+"_"+hash)
}

func loadCachedExecCredential(cacheFilePath string) (*ExecCredential, error) {
	b, err := ioutil.ReadFile(cacheFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// expected - a cache miss
			return nil, nil
		} else {
			return nil, err
		}
	}

	execCredential := &ExecCredential{}
	if err := json.Unmarshal(b, execCredential); err != nil {
		return nil, fmt.Errorf("error parsing: %v", err)
	}

	if execCredential.Status.ExpirationTimestamp.Before(time.Now()) {
		return nil, nil
	}

	if execCredential.Status.ClientCertificateData == "" || execCredential.Status.ClientKeyData == "" {
		return nil, fmt.Errorf("no credentials in cached file")
	}

	return execCredential, nil
}

func buildCredentials(ctx context.Context, f *util.Factory, options *HelperKubectlAuthOptions) (*ExecCredentialStatus, error) {
	clientset, err := f.Clientset()
	if err != nil {
		return nil, err
	}

	cluster, err := clientset.GetCluster(ctx, options.ClusterName)
	if err != nil {
		return nil, err
	}

	if cluster == nil {
		return nil, fmt.Errorf("cluster not found %q", options.ClusterName)
	}

	keyStore, err := clientset.KeyStore(cluster)
	if err != nil {
		return nil, fmt.Errorf("unable to get cluster keystore: %v", err)
	}

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
			CommonName: cn,

			Organization: []string{rbac.SystemPrivilegedGroup},
		},
		Validity: options.Lifetime,
	}
	cert, privateKey, _, err := pki.IssueCert(&req, keyStore)
	if err != nil {
		return nil, fmt.Errorf("unable to issue certificate: %v", err)
	}

	status := &ExecCredentialStatus{}
	status.ClientCertificateData, err = cert.AsString()
	if err != nil {
		return nil, err
	}
	status.ClientKeyData, err = privateKey.AsString()
	if err != nil {
		return nil, err
	}

	// Subtract a few minutes from the validity for clock skew
	status.ExpirationTimestamp = cert.Certificate.NotAfter.Add(-5 * time.Minute)

	return status, nil
}
