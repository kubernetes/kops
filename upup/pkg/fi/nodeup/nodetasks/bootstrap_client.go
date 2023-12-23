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

package nodetasks

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strconv"

	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/bootstrap"
	"k8s.io/kops/pkg/kopscontrollerclient"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
)

type BootstrapClientTask struct {
	// Certs are the requested certificates.
	Certs map[string]*BootstrapCert
	// KeypairIDs are the keypair IDs of the CAs to use for issuing certificates.
	KeypairIDs map[string]string

	// Client holds the client wrapper for the kops-bootstrap protocol
	Client *kopscontrollerclient.Client

	// UseChallengeCallback is true if we should run a challenge responder during the request.
	UseChallengeCallback bool

	// ClusterName is the name of the cluster
	ClusterName string

	keys map[string]*pki.PrivateKey
}

type BootstrapCert struct {
	Cert *fi.NodeupTaskDependentResource
	Key  *fi.NodeupTaskDependentResource
}

var (
	_ fi.NodeupTask            = &BootstrapClientTask{}
	_ fi.HasName               = &BootstrapClientTask{}
	_ fi.NodeupHasDependencies = &BootstrapClientTask{}
)

func (b *BootstrapClientTask) GetDependencies(tasks map[string]fi.NodeupTask) []fi.NodeupTask {
	// BootstrapClient depends on the protokube service to ensure gossip DNS
	var deps []fi.NodeupTask
	for _, v := range tasks {
		// BootstrapClient depends on the protokube service to ensure gossip DNS
		if svc, ok := v.(*Service); ok && svc.Name == protokubeService {
			deps = append(deps, v)
		}
	}
	return deps
}

func (b *BootstrapClientTask) GetName() *string {
	name := "BootstrapClient"
	return &name
}

func (b *BootstrapClientTask) String() string {
	return "BootstrapClientTask"
}

func (b *BootstrapClientTask) Run(c *fi.NodeupContext) error {
	ctx := c.Context()

	req := nodeup.BootstrapRequest{
		APIVersion: nodeup.BootstrapAPIVersion,
		Certs:      map[string]string{},
		KeypairIDs: b.KeypairIDs,
	}

	var challengeServer *bootstrap.ChallengeServer
	if b.UseChallengeCallback {
		s, err := bootstrap.NewChallengeServer(b.ClusterName, b.Client.CAs)
		if err != nil {
			return err
		}
		challengeServer = s
		listen := ":" + strconv.Itoa(wellknownports.NodeupChallenge)

		listener, err := challengeServer.NewListener(ctx, listen)
		if err != nil {
			return fmt.Errorf("error starting challenge listener: %w", err)
		}
		defer listener.Stop()

		req.Challenge = listener.CreateChallenge()
	}

	if b.keys == nil {
		b.keys = map[string]*pki.PrivateKey{}
	}

	for name, certRequest := range b.Certs {
		key, ok := b.keys[name]
		if !ok {
			var err error
			key, err = pki.GeneratePrivateKey()
			if err != nil {
				return fmt.Errorf("generating private key: %v", err)
			}

			certRequest.Key.Resource = &asBytesResource{key}
			b.keys[name] = key
		}

		pkData, err := x509.MarshalPKIXPublicKey(key.Key.Public())
		if err != nil {
			return fmt.Errorf("marshalling public key: %v", err)
		}
		// TODO perhaps send a CSR instead to prove we own the private key?
		req.Certs[name] = string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: pkData}))
	}

	var resp nodeup.BootstrapResponse
	err := b.Client.Query(ctx, &req, &resp)
	if err != nil {
		return err
	}

	for name, certRequest := range b.Certs {
		cert, ok := resp.Certs[name]
		if !ok {
			return fmt.Errorf("kops-controller did not return a %q certificate", name)
		}
		certificate, err := pki.ParsePEMCertificate([]byte(cert))
		if err != nil {
			return fmt.Errorf("parsing %q certificate: %v", name, err)
		}
		certRequest.Cert.Resource = asBytesResource{certificate}
	}

	return nil
}
