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

package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"runtime/debug"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/kops/cmd/kops-controller/pkg/config"
	"k8s.io/kops/cmd/kops-controller/pkg/controllerclientset"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/bootstrap"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/pkg/rbac"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/secrets"
	"k8s.io/kops/util/pkg/vfs"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Server struct {
	opt         *config.Options
	certNames   sets.Set[string]
	keypairIDs  map[string]string
	server      *http.Server
	verifier    bootstrap.Verifier
	keystore    *keystore
	secretStore fi.SecretStore

	clientset simple.Clientset

	// configBase is the base of the configuration storage.
	configBase vfs.Path

	// uncachedClient is an uncached client for the kube apiserver
	uncachedClient client.Client

	// challengeClient performs our callback-challenge into the node
	challengeClient *bootstrap.ChallengeClient
}

var _ manager.LeaderElectionRunnable = &Server{}

func NewServer(vfsContext *vfs.VFSContext, opt *config.Options, verifier bootstrap.Verifier, uncachedClient client.Client) (*Server, error) {
	server := &http.Server{
		Addr: opt.Server.Listen,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	s := &Server{
		opt:            opt,
		certNames:      sets.New(opt.Server.CertNames...),
		server:         server,
		verifier:       verifier,
		uncachedClient: uncachedClient,
	}

	configBase, err := vfsContext.BuildVfsPath(opt.ConfigBase)
	if err != nil {
		return nil, fmt.Errorf("cannot parse ConfigBase %q: %w", opt.ConfigBase, err)
	}
	s.configBase = configBase

	s.keystore, s.keypairIDs, err = newKeystore(opt.Server.CABasePath, opt.Server.SigningCAs)
	if err != nil {
		return nil, err
	}

	p, err := vfsContext.BuildVfsPath(opt.SecretStore)
	if err != nil {
		return nil, fmt.Errorf("cannot parse SecretStore %q: %w", opt.SecretStore, err)
	}
	s.secretStore = secrets.NewVFSSecretStore(nil, p)

	clientset, err := controllerclientset.New(vfsContext, configBase, opt.ClusterName, s.keystore, s.secretStore)
	if err != nil {
		return nil, fmt.Errorf("building controller clientset: %w", err)
	}
	s.clientset = clientset

	challengeClient, err := bootstrap.NewChallengeClient(s.keystore)
	if err != nil {
		return nil, err
	}
	s.challengeClient = challengeClient

	r := http.NewServeMux()
	r.Handle("/bootstrap", http.HandlerFunc(s.bootstrap))
	server.Handler = recovery(r)

	return s, nil
}

func (s *Server) NeedLeaderElection() bool {
	return false
}

func (s *Server) Start(ctx context.Context) error {
	go func() {
		<-ctx.Done()

		shutdownContext, cleanup := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanup()

		if err := s.server.Shutdown(shutdownContext); err != nil {
			klog.Warningf("error during HTTP server shutdown: %v", err)
		}

		if err := s.server.Close(); err != nil {
			klog.Warningf("error from HTTP server close: %v", err)
		}
	}()

	klog.Infof("kops-controller listening on %s", s.opt.Server.Listen)
	return s.server.ListenAndServeTLS(s.opt.Server.ServerCertificatePath, s.opt.Server.ServerKeyPath)
}

func (s *Server) bootstrap(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		klog.Infof("bootstrap %s no body", r.RemoteAddr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		klog.Infof("bootstrap %s read err: %v", r.RemoteAddr, err)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(fmt.Sprintf("bootstrap %s failed to read body: %v", r.RemoteAddr, err)))
		return
	}

	ctx := r.Context()

	id, err := s.verifier.VerifyToken(ctx, r, r.Header.Get("Authorization"), body)
	if err != nil {
		// means that we should exit nodeup gracefully
		if err == bootstrap.ErrAlreadyExists {
			w.WriteHeader(http.StatusConflict)
			klog.Infof("%s: %v", r.RemoteAddr, err)
			return
		}
		klog.Infof("bootstrap %s verify err: %v", r.RemoteAddr, err)
		w.WriteHeader(http.StatusForbidden)
		// don't return the error; this allows us to have richer errors without security implications
		_, _ = w.Write([]byte("failed to verify token"))
		return
	}

	// Once the node is registered, we don't allow further registrations, this protects against a pod or escaped workload attempting to impersonate the node.
	{
		node := &corev1.Node{}
		err := s.uncachedClient.Get(ctx, types.NamespacedName{Name: id.NodeName}, node)
		if err == nil {
			for _, condition := range node.Status.Conditions {
				if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
					klog.Infof("bootstrap %s node %q already exists; denying to avoid node-impersonation attacks", r.RemoteAddr, id.NodeName)
					w.WriteHeader(http.StatusConflict)
					_, _ = w.Write([]byte("node already registered"))
					return
				}
			}
		}
		if err != nil && !errors.IsNotFound(err) {
			klog.Infof("bootstrap %s error querying for node %q: %v", r.RemoteAddr, id.NodeName, err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("internal error"))
			return
		}
	}

	req := &nodeup.BootstrapRequest{}
	if err := json.Unmarshal(body, req); err != nil {
		klog.Infof("bootstrap %s decode err: %v", r.RemoteAddr, err)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(fmt.Sprintf("failed to decode: %v", err)))
		return
	}

	if req.APIVersion != nodeup.BootstrapAPIVersion {
		klog.Infof("bootstrap %s wrong APIVersion", r.RemoteAddr)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("unexpected APIVersion"))
		return
	}

	if model.UseChallengeCallback(kops.CloudProviderID(s.opt.Cloud)) {
		if id.ChallengeEndpoint == "" {
			klog.Infof("cannot determine endpoint for bootstrap callback challenge from %q", r.RemoteAddr)
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("callback failed"))
			return
		}
		if err := s.challengeClient.DoCallbackChallenge(ctx, s.opt.ClusterName, id.ChallengeEndpoint, req); err != nil {
			klog.Infof("bootstrap %s callback challenge failed: %v", r.RemoteAddr, err)
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("callback failed"))
			return
		}

		klog.Infof("performed successful callback challenge with %s; identified as %s", id.ChallengeEndpoint, id.NodeName)
	}

	resp := &nodeup.BootstrapResponse{
		Certs: map[string]string{},
	}

	// Support for nodes that have no access to the state store
	if req.IncludeNodeConfig {
		nodeConfig, err := s.getNodeConfig(r.Context(), req, id)
		if err != nil {
			klog.Infof("bootstrap failed to build node config: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("failed to build node config"))
			return
		}
		resp.NodeConfig = nodeConfig
	}

	// Skew the certificate lifetime by up to 30 days based on information about the requesting node.
	// This is so that different nodes created at the same time have the certificates they generated
	// expire at different times, but all certificates on a given node expire around the same time.
	hash := fnv.New32()
	_, _ = hash.Write([]byte(r.RemoteAddr))
	validHours := (455 * 24) + (hash.Sum32() % (30 * 24))

	for name, pubKey := range req.Certs {
		cert, err := s.issueCert(ctx, name, pubKey, id, validHours, req.KeypairIDs)
		if err != nil {
			klog.Infof("bootstrap %s cert %q issue err: %v", r.RemoteAddr, name, err)
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(fmt.Sprintf("failed to issue %q: %v", name, err)))
			return
		}
		resp.Certs[name] = cert
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
	klog.Infof("bootstrap %s %s success", r.RemoteAddr, id.NodeName)
}

func (s *Server) issueCert(ctx context.Context, name string, pubKey string, id *bootstrap.VerifyResult, validHours uint32, keypairIDs map[string]string) (string, error) {
	block, _ := pem.Decode([]byte(pubKey))
	if block.Type != "RSA PUBLIC KEY" {
		return "", fmt.Errorf("unexpected key type %q", block.Type)
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("parsing key: %v", err)
	}

	issueReq := &pki.IssueCertRequest{
		Signer:    fi.CertificateIDCA,
		Type:      "client",
		PublicKey: key,
		Validity:  time.Hour * time.Duration(validHours),
	}

	if !s.certNames.Has(name) {
		return "", fmt.Errorf("key name not enabled")
	}
	switch name {
	case "etcd-client-cilium":
		issueReq.Signer = "etcd-clients-ca-cilium"
		issueReq.Subject = pkix.Name{
			CommonName: "cilium",
		}
	case "kubelet":
		issueReq.Subject = pkix.Name{
			CommonName:   fmt.Sprintf("system:node:%s", id.NodeName),
			Organization: []string{rbac.NodesGroup},
		}
	case "kubelet-server":
		issueReq.Subject = pkix.Name{
			CommonName: id.NodeName,
		}
		issueReq.AlternateNames = id.CertificateNames
		issueReq.Type = "server"
	case "kube-proxy":
		issueReq.Subject = pkix.Name{
			CommonName: rbac.KubeProxy,
		}
	case "kube-router":
		issueReq.Subject = pkix.Name{
			CommonName: rbac.KubeRouter,
		}
	default:
		return "", fmt.Errorf("unexpected key name")
	}

	// This field was added to the protocol in kOps 1.22.
	if len(keypairIDs) > 0 {
		if keypairIDs[issueReq.Signer] != s.keypairIDs[issueReq.Signer] {
			return "", fmt.Errorf("request's keypair ID %q for %s didn't match server's %q", keypairIDs[issueReq.Signer], issueReq.Signer, s.keypairIDs[issueReq.Signer])
		}
	}

	cert, _, _, err := pki.IssueCert(ctx, issueReq, s.keystore)
	if err != nil {
		return "", fmt.Errorf("issuing certificate: %v", err)
	}

	return cert.AsString()
}

// recovery is responsible for ensuring we don't exit on a panic.
func recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)

				klog.Errorf("failed to handle request: threw exception: %v: %s", err, debug.Stack())
			}
		}()

		next.ServeHTTP(w, req)
	})
}
