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
	"bytes"
	"context"
	"crypto/x509/pkix"
	"fmt"
	"hash/fnv"
	"io"
	"net"
	"path/filepath"
	"sort"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
)

// PKIXName is a simplified form of pkix.Name, for better golden test output
type PKIXName struct {
	fi.NodeupNotADependency
	CommonName   string
	Organization []string `json:",omitempty"`
}

func (n *PKIXName) toPKIXName() pkix.Name {
	return pkix.Name{
		CommonName:   n.CommonName,
		Organization: n.Organization,
	}
}

type IssueCert struct {
	Name string

	Signer         string   `json:"signer"`
	KeypairID      string   `json:"keypairID"`
	Type           string   `json:"type"`
	Subject        PKIXName `json:"subject"`
	AlternateNames []string `json:"alternateNames,omitempty"`

	// IncludeRootCertificate will force the certificate data to include the full chain, not just the leaf
	IncludeRootCertificate bool `json:"includeRootCertificate,omitempty"`

	cert *fi.NodeupTaskDependentResource
	key  *fi.NodeupTaskDependentResource
	ca   *fi.NodeupTaskDependentResource
}

var (
	_ fi.NodeupTask = &IssueCert{}
	_ fi.HasName    = &IssueCert{}
)

func (i *IssueCert) GetName() *string {
	return &i.Name
}

// String returns a string representation, implementing the Stringer interface
func (i *IssueCert) String() string {
	return fmt.Sprintf("IssueCert: %s", i.Name)
}

func (i *IssueCert) GetResources() (certResource, keyResource, caResource *fi.NodeupTaskDependentResource) {
	if i.cert == nil {
		i.cert = &fi.NodeupTaskDependentResource{Task: i}
		i.key = &fi.NodeupTaskDependentResource{Task: i}
		i.ca = &fi.NodeupTaskDependentResource{Task: i}
	}
	return i.cert, i.key, i.ca
}

// AddFileTasks creates the directory, certificates and keys.
// For more control, prefer calling AddCertificateFileTasks and AddKeyFileTasks directly.
func (i *IssueCert) AddFileTasks(c *fi.NodeupModelBuilderContext, dir string, name string, caName string, owner *string) error {
	c.EnsureTask(&File{
		Path: dir,
		Type: FileType_Directory,
		Mode: fi.PtrTo("0755"),
	})

	if err := i.AddCertificateFileTasks(c, dir, name, caName, owner); err != nil {
		return err
	}
	if err := i.AddKeyFileTasks(c, dir, name, owner); err != nil {
		return err
	}
	return nil
}

func (i *IssueCert) AddKeyFileTasks(c *fi.NodeupModelBuilderContext, dir string, name string, owner *string) error {
	_, keyResource, _ := i.GetResources()

	c.AddTask(&File{
		Path:     filepath.Join(dir, name+".key"),
		Contents: keyResource,
		Type:     FileType_File,
		Mode:     fi.PtrTo("0600"),
		Owner:    owner,
	})

	return nil
}

func (i *IssueCert) AddCertificateFileTasks(c *fi.NodeupModelBuilderContext, dir string, name string, caName string, owner *string) error {
	certResource, _, caResource := i.GetResources()

	c.AddTask(&File{
		Path:     filepath.Join(dir, name+".crt"),
		Contents: certResource,
		Type:     FileType_File,
		Mode:     fi.PtrTo("0644"),
		Owner:    owner,
	})

	if caName != "" {
		c.EnsureTask(&File{
			Path:     filepath.Join(dir, caName+".crt"),
			Contents: caResource,
			Type:     FileType_File,
			Mode:     fi.PtrTo("0644"),
			Owner:    owner,
		})
	}

	return nil
}

func (e *IssueCert) Run(c *fi.NodeupContext) error {
	ctx := c.Context()

	// Skew the certificate lifetime by up to 30 days based on information about the generating node.
	// This is so that different nodes created at the same time have the certificates they generated
	// expire at different times, but all certificates on a given node expire around the same time.
	hash := fnv.New32()
	addrs, err := net.InterfaceAddrs()
	sort.Slice(addrs, func(i, j int) bool {
		return addrs[i].String() < addrs[j].String()
	})
	if err == nil {
		for _, addr := range addrs {
			_, _ = hash.Write([]byte(addr.String()))
		}
	} else {
		klog.Warningf("cannot skew certificate lifetime: failed to get interface addresses: %v", err)
	}
	validHours := (455 * 24) + (hash.Sum32() % (30 * 24))

	req := &pki.IssueCertRequest{
		Signer:         e.Signer,
		Type:           e.Type,
		Subject:        e.Subject.toPKIXName(),
		AlternateNames: e.AlternateNames,
		Validity:       time.Hour * time.Duration(validHours),
	}

	keystore, err := newStaticKeystore(ctx, e.Signer, e.KeypairID, c.T.Keystore)
	if err != nil {
		return err
	}

	klog.Infof("signing certificate for %q", e.Name)
	certificate, privateKey, caCertificate, err := pki.IssueCert(ctx, req, keystore)
	if err != nil {
		return err
	}

	certResource, keyResource, caResource := e.GetResources()
	certResource.Resource = &asBytesResource{certificate}
	keyResource.Resource = &asBytesResource{privateKey}
	caResource.Resource = &asBytesResource{caCertificate}

	if e.IncludeRootCertificate {
		var b bytes.Buffer
		if _, err := certificate.WriteTo(&b); err != nil {
			return err
		}
		b.WriteString("\n")
		if _, err := caCertificate.WriteTo(&b); err != nil {
			return err
		}
		certResource.Resource = fi.NewBytesResource(b.Bytes())
	}

	return nil
}

type hasAsBytes interface {
	AsBytes() ([]byte, error)
}

type asBytesResource struct {
	hasAsBytes
}

func (a asBytesResource) Open() (io.Reader, error) {
	data, err := a.AsBytes()
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

type staticKeystore struct {
	keyset      string
	certificate *pki.Certificate
	key         *pki.PrivateKey
}

// FindPrimaryKeypair implements pki.Keystore
func (s staticKeystore) FindPrimaryKeypair(ctx context.Context, name string) (*pki.Certificate, *pki.PrivateKey, error) {
	if name != s.keyset {
		return nil, nil, fmt.Errorf("wrong signer: expected %q got %q", s.keyset, name)
	}
	return s.certificate, s.key, nil
}

func newStaticKeystore(ctx context.Context, signer string, keypairID string, keystore fi.KeystoreReader) (pki.Keystore, error) {
	if signer == "" {
		return nil, nil
	}

	if keypairID == "" {
		return nil, fmt.Errorf("missing keypairID for %s", signer)
	}

	keyset, err := keystore.FindKeyset(ctx, signer)
	if err != nil {
		return nil, fmt.Errorf("reading keyset for %s: %v", signer, err)
	}
	if keyset == nil {
		return nil, fmt.Errorf("keyset %q not found", signer)
	}

	item := keyset.Items[keypairID]
	if item == nil {
		return nil, fmt.Errorf("no keypair with id %s for %s", keypairID, signer)
	}

	return &staticKeystore{
		keyset:      signer,
		certificate: item.Certificate,
		key:         item.PrivateKey,
	}, nil
}
