/*
Copyright 2026 The Kubernetes Authors.

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
	"bytes"
	"context"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/util/pkg/awsinterfaces"
)

func TestUpdateClusterKubeconfigFlags(t *testing.T) {
	for _, tt := range []struct {
		args       []string
		wantExport bool
	}{
		{nil, true},
		{[]string{"--use-kubeconfig"}, false},
		{[]string{"--use-kubeconfig=false"}, true},
		{[]string{"--create-kube-config=false"}, false},
		{[]string{"--use-kubeconfig", "--create-kube-config=true"}, true},
		{[]string{"--use-kubeconfig", "--create-kube-config=false"}, false},
		{[]string{"--use-kubeconfig", "--api-server=https://override.example.com"}, false},
		{[]string{"--use-kubeconfig", "--admin"}, true},
		{[]string{"--use-kubeconfig", "--user=custom"}, true},
		{[]string{"--use-kubeconfig", "--internal"}, true},
		{[]string{"--use-kubeconfig", "--admin=0s"}, false},
		{[]string{"--use-kubeconfig", "--internal=false"}, false},
		{[]string{"--use-kubeconfig", "--create-kube-config=false", "--admin"}, true},
		{[]string{"--use-kubeconfig", "--create-kube-config=false", "--user=custom"}, true},
		{[]string{"--use-kubeconfig", "--create-kube-config=false", "--internal"}, true},
	} {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			cmd := NewCmdUpdateCluster(nil, io.Discard)
			require.NoError(t, cmd.ParseFlags(tt.args))
			require.NoError(t, cmd.PreRunE(cmd, nil))
			// Run option normalization, stopping before any state or cloud access.
			err := cmd.RunE(cmd, nil)
			require.ErrorContains(t, err, "Cluster name is required")
			got, err := cmd.Flags().GetBool("create-kube-config")
			require.NoError(t, err)
			require.Equal(t, tt.wantExport, got)
		})
	}
}

func setupUpdateConnectionTest(t *testing.T) (*util.Factory, *kops.Cluster) {
	t.Helper()
	h := testutils.NewIntegrationTestHarness(t)
	t.Cleanup(h.Close)
	h.SetupMockAWS()
	f := util.NewFactory(&util.FactoryOptions{RegistryPath: "memfs://tests"})
	clientset, err := f.KopsClient()
	require.NoError(t, err)
	cluster, err := clientset.CreateCluster(context.Background(), testutils.BuildMinimalClusterAWS("connection.example.com"))
	require.NoError(t, err)
	keyStore, err := clientset.KeyStore(cluster)
	require.NoError(t, err)
	cert, key, _, err := pki.IssueCert(context.Background(), &pki.IssueCertRequest{Type: "ca"}, nil)
	require.NoError(t, err)
	keyset, err := fi.NewKeyset(cert, key)
	require.NoError(t, err)
	require.NoError(t, keyStore.StoreKeyset(context.Background(), fi.CertificateIDCA, keyset))
	t.Setenv("KUBECONFIG", filepath.Join(t.TempDir(), "config"))
	return f, cluster
}

func writeUpdateTestKubeconfig(t *testing.T, clusterName, server string, ca []byte) *clientcmdapi.Config {
	t.Helper()
	config := clientcmdapi.NewConfig()
	config.CurrentContext = "unrelated"
	config.Contexts[clusterName] = &clientcmdapi.Context{Cluster: "custom-cluster", AuthInfo: "custom-user"}
	config.Clusters["custom-cluster"] = &clientcmdapi.Cluster{Server: server, TLSServerName: "example.com", CertificateAuthorityData: ca}
	config.AuthInfos["custom-user"] = &clientcmdapi.AuthInfo{Token: "test-token"}
	require.NoError(t, clientcmd.WriteToFile(*config, os.Getenv("KUBECONFIG")))
	return config
}

func TestUpdateClusterConnectionOptions(t *testing.T) {
	for _, command := range []string{"update", "reconcile-dry-run", "reconcile"} {
		for _, reuse := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/reuse=%t", command, reuse), func(t *testing.T) {
				f, cluster := setupUpdateConnectionTest(t)
				var requests atomic.Int32
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests.Add(1)
					assert.Equal(t, "/api/v1/nodes", r.URL.Path)
					assert.Equal(t, "node-role.kubernetes.io/control-plane", r.URL.Query().Get("labelSelector"))
					if reuse {
						assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
						assert.Equal(t, "example.com", r.TLS.ServerName)
					}
					w.Header().Set("Content-Type", "application/json")
					io.WriteString(w, `{"apiVersion":"v1","kind":"NodeList","items":[]}`)
				})
				var server *httptest.Server
				if reuse {
					server = httptest.NewTLSServer(handler)
					ca := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
					writeUpdateTestKubeconfig(t, cluster.Name, "https://unused.invalid", ca)
				} else {
					// Generated credentials and CA come from the state store, with no local kubeconfig.
					server = httptest.NewServer(handler)
				}
				defer server.Close()
				connection := kubeconfig.CreateKubecfgOptions{UseKubeconfig: reuse, OverrideAPIServer: server.URL}
				var err error
				if command != "update" {
					options := &ReconcileClusterOptions{CreateKubecfgOptions: connection}
					options.InitDefaults()
					options.ClusterName = cluster.Name
					options.Yes = command == "reconcile"
					err = RunReconcileCluster(context.Background(), f, io.Discard, options)
				} else {
					options := &UpdateClusterOptions{}
					options.InitDefaults()
					options.ClusterName = cluster.Name
					options.CreateKubecfgOptions = connection
					_, err = RunUpdateCluster(context.Background(), f, io.Discard, options)
				}
				// This incomplete cluster intentionally stops at ApplyCluster validation.
				require.ErrorContains(t, err, "kubernetes upgrade is required")
				require.EqualValues(t, 1, requests.Load(), "version check must use the requested endpoint before applying")
			})
		}
	}
}

func TestCheckControlPlaneRunningVersion(t *testing.T) {
	f, cluster := setupUpdateConnectionTest(t)
	cluster.Spec.KubernetesVersion = "1.35.0"
	for _, tt := range []struct {
		name     string
		response string
		status   int
		want     string
	}{
		{"minimum version", `{"apiVersion":"v1","kind":"NodeList","items":[{"status":{"nodeInfo":{"kubeletVersion":"v1.34.3"}}},{"status":{"nodeInfo":{"kubeletVersion":"v1.33.7"}}}]}`, http.StatusOK, "1.33.7"},
		{"no nodes", `{"apiVersion":"v1","kind":"NodeList","items":[]}`, http.StatusOK, "1.35.0"},
		{"API unavailable", "unavailable", http.StatusServiceUnavailable, "1.35.0"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				io.WriteString(w, tt.response)
			}))
			defer server.Close()
			version, err := checkControlPlaneRunningVersion(context.Background(), f, cluster, kubeconfig.CreateKubecfgOptions{OverrideAPIServer: server.URL})
			if tt.status == http.StatusOK {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			require.Equal(t, tt.want, version)
		})
	}
	// Reuse requires an existing context; failure must retain the configured version.
	version, err := checkControlPlaneRunningVersion(context.Background(), f, cluster, kubeconfig.CreateKubecfgOptions{UseKubeconfig: true})
	require.Error(t, err)
	require.Equal(t, cluster.Spec.KubernetesVersion, version)
}

func TestUpdateClusterKubeconfigExport(t *testing.T) {
	for _, tt := range []struct {
		name       string
		args       []string
		wantExport bool
	}{
		{"generated defaults", nil, true},
		{"reuse", []string{"--use-kubeconfig"}, false},
		{"explicit export", []string{"--use-kubeconfig", "--create-kube-config=true"}, true},
		{"explicit no export", []string{"--use-kubeconfig", "--create-kube-config=false"}, false},
		{"user implies export", []string{"--use-kubeconfig", "--user=custom-user"}, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("KOPS_RUN_TOO_NEW_VERSION", "1")
			h := testutils.NewIntegrationTestHarness(t)
			defer h.Close()
			h.MockKopsVersion(integrationTestKopsVersion)
			h.SetupMockAWS()
			ctx := context.Background()
			i := newIntegrationTest("minimal.example.com", filepath.Join(updateClusterTestBase, "minimal-1.35"))
			f := i.setupCluster(t, ctx, "in-v1alpha2.yaml", bytes.Buffer{})
			t.Setenv("KUBECONFIG", filepath.Join(t.TempDir(), "config"))
			config := writeUpdateTestKubeconfig(t, i.clusterName, "https://custom.example.com", []byte("custom-ca"))
			config.Clusters["custom-cluster"].ProxyURL = "http://proxy.example.com:8080"
			config.AuthInfos["custom-user"].Exec = &clientcmdapi.ExecConfig{Command: "custom-auth", APIVersion: "client.authentication.k8s.io/v1beta1"}
			require.NoError(t, clientcmd.WriteToFile(*config, os.Getenv("KUBECONFIG")))
			before, err := os.ReadFile(os.Getenv("KUBECONFIG"))
			require.NoError(t, err)

			originalClusterName := rootCommand.clusterName
			t.Cleanup(func() { rootCommand.clusterName = originalClusterName })
			rootCommand.clusterName = ""
			cmd := NewCmdUpdateCluster(f, io.Discard)
			cmd.SetContext(ctx)
			// Apply to mock AWS so export checks do not depend on concurrent Terraform file rendering.
			args := append([]string{"--yes", "--ignore-kubelet-version-skew", "--out=" + t.TempDir()}, tt.args...)
			require.NoError(t, cmd.ParseFlags(args))
			require.NoError(t, cmd.Args(cmd, []string{i.clusterName}))
			require.NoError(t, cmd.PreRunE(cmd, nil))
			require.NoError(t, cmd.RunE(cmd, nil))

			after, err := os.ReadFile(os.Getenv("KUBECONFIG"))
			require.NoError(t, err)
			if !tt.wantExport {
				require.Equal(t, before, after, "reuse must preserve the complete kubeconfig")
			} else {
				require.NotEqual(t, before, after)
				exported, err := clientcmd.Load(after)
				require.NoError(t, err)
				require.Equal(t, i.clusterName, exported.Contexts[i.clusterName].Cluster)
				require.Equal(t, "https://api."+i.clusterName, exported.Clusters[i.clusterName].Server)
				require.NotEmpty(t, exported.Clusters[i.clusterName].CertificateAuthorityData)
			}
		})
	}
}

func TestUpdateClusterRefreshesEndpoint(t *testing.T) {
	for _, reuse := range []bool{false, true} {
		t.Run(fmt.Sprintf("reuse=%t", reuse), func(t *testing.T) {
			t.Setenv("KOPS_RUN_TOO_NEW_VERSION", "1")
			h := testutils.NewIntegrationTestHarness(t)
			defer h.Close()
			h.MockKopsVersion(integrationTestKopsVersion)
			mockCloud := h.SetupMockAWS()
			mockCloud.MockELBV2 = &readyUpdateLoadBalancerAPI{mockCloud.MockELBV2}
			ctx := context.Background()
			i := newIntegrationTest("minimal.example.com", filepath.Join(updateClusterTestBase, "minimal-dns-none"))
			f := i.setupCluster(t, ctx, "in-v1alpha2.yaml", bytes.Buffer{})
			t.Setenv("KUBECONFIG", filepath.Join(t.TempDir(), "config"))
			cluster, err := GetCluster(ctx, f, i.clusterName)
			require.NoError(t, err)
			// The integration fixture's historical CA key is too small to sign
			// new client certificates with current Go versions.
			clientset, err := f.KopsClient()
			require.NoError(t, err)
			keyStore, err := clientset.KeyStore(cluster)
			require.NoError(t, err)
			cert, key, _, err := pki.IssueCert(ctx, &pki.IssueCertRequest{Type: "ca"}, nil)
			require.NoError(t, err)
			keyset, err := fi.NewKeyset(cert, key)
			require.NoError(t, err)
			require.NoError(t, keyStore.StoreKeyset(ctx, fi.CertificateIDCA, keyset))
			options := kubeconfig.CreateKubecfgOptions{UseKubeconfig: reuse}
			if reuse {
				writeUpdateTestKubeconfig(t, cluster.Name, "https://custom.example.com", nil)
			}

			before, err := f.RESTConfig(ctx, cluster, options)
			require.NoError(t, err)
			if !reuse {
				require.Equal(t, "https://api."+cluster.Name, before.Host, "no NLB exists yet")
			}
			// A failed pre-apply version check leaves its connection in the factory.
			// Cancel the request to avoid contacting the unavailable endpoint.
			probeCtx, cancel := context.WithCancel(ctx)
			cancel()
			_, err = checkControlPlaneRunningVersion(probeCtx, f, cluster, options)
			require.Error(t, err)

			update := &CoreUpdateClusterOptions{}
			update.InitDefaults()
			update.ClusterName = cluster.Name
			update.OutDir = t.TempDir()
			update.IgnoreKubeletVersionSkew = true // The version check was exercised above.
			_, err = RunCoreUpdateCluster(ctx, f, io.Discard, update, options)
			require.NoError(t, err)
			afterDryRun, err := f.RESTConfig(ctx, cluster, options)
			require.NoError(t, err)
			require.Same(t, before, afterDryRun, "dry runs do not change infrastructure")

			// This is the update stage used by reconcile before validation.
			update.Yes = true
			_, err = RunCoreUpdateCluster(ctx, f, io.Discard, update, options)
			require.NoError(t, err)
			cloud, err := cloudup.BuildCloud(cluster)
			require.NoError(t, err)
			ingresses, err := cloud.GetApiIngressStatus(cluster)
			require.NoError(t, err)
			require.Len(t, ingresses, 1, "the update must create the API load balancer")
			require.NotEmpty(t, ingresses[0].Hostname)
			after, err := f.RESTConfig(ctx, cluster, options)
			require.NoError(t, err)
			if reuse {
				require.Same(t, before, after, "keep the user's configured connection")
			} else {
				require.Equal(t, "https://"+ingresses[0].Hostname, after.Host, "validation must discover the new NLB")
				require.NotEqual(t, before.Host, after.Host)
				require.Equal(t, cluster.APIInternalName(), after.ServerName)
				require.NotEmpty(t, after.CertData)
			}
		})
	}
}

// The shared AWS mock does not set NLB readiness. Mark its load balancers ready
// so a direct apply can finish without waiting for real AWS provisioning.
type readyUpdateLoadBalancerAPI struct {
	awsinterfaces.ELBV2API
}

func (m *readyUpdateLoadBalancerAPI) DescribeLoadBalancers(ctx context.Context, input *elbv2.DescribeLoadBalancersInput, options ...func(*elbv2.Options)) (*elbv2.DescribeLoadBalancersOutput, error) {
	output, err := m.ELBV2API.DescribeLoadBalancers(ctx, input, options...)
	if err == nil {
		for i := range output.LoadBalancers {
			output.LoadBalancers[i].State = &elbv2types.LoadBalancerState{Code: elbv2types.LoadBalancerStateEnumActive}
		}
	}
	return output, err
}

func TestUpdateConnectionFactory(t *testing.T) {
	f, cluster := setupUpdateConnectionTest(t)
	config := writeUpdateTestKubeconfig(t, cluster.Name, "https://custom.example.com", []byte("custom-ca"))
	config.Clusters["custom-cluster"].ProxyURL = "http://proxy.example.com:8080"
	config.AuthInfos["custom-user"] = &clientcmdapi.AuthInfo{Exec: &clientcmdapi.ExecConfig{
		Command: "custom-auth", Args: []string{"login"}, APIVersion: "client.authentication.k8s.io/v1",
		InteractiveMode: clientcmdapi.NeverExecInteractiveMode,
	}}
	require.NoError(t, clientcmd.WriteToFile(*config, os.Getenv("KUBECONFIG")))
	before, err := os.ReadFile(os.Getenv("KUBECONFIG"))
	require.NoError(t, err)
	ctx := context.Background()

	generated, err := f.RESTConfig(ctx, cluster, kubeconfig.CreateKubecfgOptions{})
	require.NoError(t, err)
	require.Equal(t, "https://api."+cluster.Name, generated.Host)
	require.NotEmpty(t, generated.CertData)
	require.NotEmpty(t, generated.KeyData)

	options := kubeconfig.CreateKubecfgOptions{UseKubeconfig: true}
	reused, err := f.RESTConfig(ctx, cluster, options)
	require.NoError(t, err)
	require.Equal(t, "https://custom.example.com", reused.Host)
	require.Equal(t, "example.com", reused.ServerName)
	require.Equal(t, []byte("custom-ca"), reused.CAData)
	require.Equal(t, config.AuthInfos["custom-user"].Exec, reused.ExecProvider)
	require.NotNil(t, reused.Proxy)
	proxy, err := reused.Proxy(&http.Request{})
	require.NoError(t, err)
	require.Equal(t, "http://proxy.example.com:8080", proxy.String())

	options.OverrideAPIServer = "https://override.example.com"
	overridden, err := f.RESTConfig(ctx, cluster, options)
	require.NoError(t, err)
	require.Equal(t, options.OverrideAPIServer, overridden.Host)
	require.Equal(t, reused.ServerName, overridden.ServerName)
	require.Equal(t, reused.ExecProvider, overridden.ExecProvider)
	options.OverrideAPIServer = ""
	again, err := f.RESTConfig(ctx, cluster, options)
	require.NoError(t, err)
	require.Equal(t, reused.Host, again.Host)
	after, err := os.ReadFile(os.Getenv("KUBECONFIG"))
	require.NoError(t, err)
	require.Equal(t, before, after, "building connections must preserve kubeconfig")
}

func TestUpdateClusterTargetValidation(t *testing.T) {
	const partialTerraformError = "cannot use --target=terraform with --instance-group or --instance-group-roles; generate the full configuration and use terraform apply -target instead"
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "terraform with instance groups",
			args:    []string{"--target=terraform", "--instance-group=nodes-a,nodes-b"},
			wantErr: partialTerraformError,
		},
		{
			name:    "terraform with instance group roles",
			args:    []string{"--target=terraform", "--instance-group-roles=control-plane,apiserver"},
			wantErr: partialTerraformError,
		},
		{
			name:    "terraform with both filters",
			args:    []string{"--target=terraform", "--instance-group=nodes", "--instance-group-roles=node"},
			wantErr: partialTerraformError,
		},
		{
			name: "terraform without filters",
			args: []string{"--target=terraform"},
		},
		{
			name: "terraform with empty filters",
			args: []string{"--target=terraform", "--instance-group=", "--instance-group-roles="},
		},
		{
			name: "direct with instance groups",
			args: []string{"--target=direct", "--instance-group=nodes", "--yes"},
		},
		{
			name: "direct with instance group roles",
			args: []string{"--target=direct", "--instance-group-roles=node", "--yes"},
		},
		{
			name: "dry run with instance groups",
			args: []string{"--target=dryrun", "--instance-group=nodes"},
		},
		{
			name: "default target with instance group roles",
			args: []string{"--instance-group-roles=node"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCmdUpdateCluster(nil, io.Discard)
			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Fatalf("parsing flags: %v", err)
			}
			var err error
			if cmd.PreRunE != nil {
				err = cmd.PreRunE(cmd, nil)
			}
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected validation error: %v", err)
				}
			} else if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("validation error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}
