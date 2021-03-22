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
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/kops/cmd/kops-controller/pkg/config"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/deprecations"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/pkg/rbac"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

type Server struct {
	opt       *config.Options
	certNames sets.String
	server    *http.Server
	verifier  fi.Verifier
	keystore  pki.Keystore

	// configBase is the base of the configuration storage.
	configBase vfs.Path
}

func NewServer(opt *config.Options, verifier fi.Verifier) (*Server, error) {
	server := &http.Server{
		Addr: opt.Server.Listen,
		TLSConfig: &tls.Config{
			MinVersion:               tls.VersionTLS12,
			PreferServerCipherSuites: true,
		},
	}

	s := &Server{
		opt:       opt,
		certNames: sets.NewString(opt.Server.CertNames...),
		server:    server,
		verifier:  verifier,
	}

	configBase, err := vfs.Context.BuildVfsPath(opt.ConfigBase)
	if err != nil {
		return nil, fmt.Errorf("cannot parse ConfigBase %q: %v", opt.ConfigBase, err)
	}
	s.configBase = configBase

	r := http.NewServeMux()
	r.Handle("/bootstrap", http.HandlerFunc(s.bootstrap))
	server.Handler = recovery(r)

	return s, nil
}

func (s *Server) Start() error {
	var err error
	s.keystore, err = newKeystore(s.opt.Server.CABasePath, s.opt.Server.SigningCAs)
	if err != nil {
		return err
	}

	return s.server.ListenAndServeTLS(s.opt.Server.ServerCertificatePath, s.opt.Server.ServerKeyPath)
}

func (s *Server) bootstrap(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		klog.Infof("bootstrap %s no body", r.RemoteAddr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		klog.Infof("bootstrap %s read err: %v", r.RemoteAddr, err)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(fmt.Sprintf("bootstrap %s failed to read body: %v", r.RemoteAddr, err)))
		return
	}

	id, err := s.verifier.VerifyToken(r.Header.Get("Authorization"), body)
	if err != nil {
		klog.Infof("bootstrap %s verify err: %v", r.RemoteAddr, err)
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(fmt.Sprintf("failed to verify token: %v", err)))
		return
	}

	req := &nodeup.BootstrapRequest{}
	err = json.Unmarshal(body, req)
	if err != nil {
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
	validity := time.Hour * time.Duration(validHours)

	for name, csr := range req.CertificateSigningRequests {
		cert, err := s.issueCSR(name, csr, id, validity)
		if err != nil {
			klog.Infof("bootstrap %s cert %q issueCSR failed: %v", r.RemoteAddr, name, err)
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(fmt.Sprintf("failed to issue %q: %v", name, err)))
			return
		}
		resp.Certs[name] = cert
	}

	// @deprecated:ShouldIssueWithCSRs
	for name, pubKey := range req.Certs {
		if _, done := resp.Certs[name]; done {
			continue
		}
		deprecations.ShouldIssueWithCSRs.Use()

		cert, err := s.issueCert(name, pubKey, id, validity)
		if err != nil {
			klog.Infof("bootstrap %s cert %q issueCert failed: %v", r.RemoteAddr, name, err)
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

// @deprecated:ShouldIssueWithCSRs
func (s *Server) issueWellKnownCertificate(name string, publicKey crypto.PublicKey, id *fi.VerifyResult, validity time.Duration) (*pki.Certificate, error) {
	issueReq := &pki.IssueCertRequest{
		Signer:    fi.CertificateIDCA,
		Type:      "client",
		PublicKey: publicKey,
		Validity:  validity,
	}

	if !s.certNames.Has(name) {
		return nil, fmt.Errorf("key name %q not enabled", name)
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
		issueReq.AlternateNames = []string{id.NodeName}
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
		return nil, fmt.Errorf("unexpected key name %q", name)
	}

	cert, _, _, err := pki.IssueCert(issueReq, s.keystore)
	if err != nil {
		return nil, fmt.Errorf("error issuing certificate: %w", err)
	}

	return cert, nil
}

// deprecated:ShouldIssueWithCSRs
func (s *Server) issueCert(name string, pubKey string, id *fi.VerifyResult, validity time.Duration) (string, error) {
	block, _ := pem.Decode([]byte(pubKey))
	if block.Type != "RSA PUBLIC KEY" {
		return "", fmt.Errorf("unexpected key type %q", block.Type)
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("error parsing public key: %w", err)
	}

	publicKey, ok := key.(crypto.PublicKey)
	if !ok {
		return "", fmt.Errorf("unexpected type for PublicKey: %T", key)
	}

	cert, err := s.issueWellKnownCertificate(name, publicKey, id, validity)
	if err != nil {
		return "", err
	}
	return cert.AsString()
}

func (s *Server) issueCSR(name string, request *nodeup.CertificateSigningRequest, id *fi.VerifyResult, validity time.Duration) (string, error) {
	block, _ := pem.Decode([]byte(request.PEMData))
	if block.Type != "CERTIFICATE REQUEST" {
		return "", fmt.Errorf("unexpected PEM block type %q", block.Type)
	}
	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("error parsing CSR: %w", err)
	}

	// Ensure that the CSR is correctly signed, indicating that the caller controls the public key.
	if err := csr.CheckSignature(); err != nil {
		return "", fmt.Errorf("error validating CSR: %w", err)
	}

	publicKey, ok := csr.PublicKey.(crypto.PublicKey)
	if !ok {
		return "", fmt.Errorf("unexpected type for PublicKey: %T", csr.PublicKey)
	}

	cert, err := s.issueWellKnownCertificate(name, publicKey, id, validity)
	if err != nil {
		return "", err
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
