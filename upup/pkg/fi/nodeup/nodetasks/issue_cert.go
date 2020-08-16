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
	fi.NotADependency
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
	Type           string   `json:"type"`
	Subject        PKIXName `json:"subject"`
	AlternateNames []string `json:"alternateNames,omitempty"`

	cert *fi.TaskDependentResource
	key  *fi.TaskDependentResource
	ca   *fi.TaskDependentResource
}

var _ fi.Task = &IssueCert{}
var _ fi.HasName = &IssueCert{}

func (i *IssueCert) GetName() *string {
	return &i.Name
}

// String returns a string representation, implementing the Stringer interface
func (i *IssueCert) String() string {
	return fmt.Sprintf("IssueCert: %s", i.Name)
}

func (i *IssueCert) GetResources() (certResource, keyResource, caResource *fi.TaskDependentResource) {
	if i.cert == nil {
		i.cert = &fi.TaskDependentResource{Task: i}
		i.key = &fi.TaskDependentResource{Task: i}
		i.ca = &fi.TaskDependentResource{Task: i}
	}
	return i.cert, i.key, i.ca
}

func (i *IssueCert) AddFileTasks(c *fi.ModelBuilderContext, dir string, name string, caName string, owner *string) error {
	certResource, keyResource, caResource := i.GetResources()
	err := c.EnsureTask(&File{
		Path: dir,
		Type: FileType_Directory,
		Mode: fi.String("0755"),
	})
	if err != nil {
		return err
	}

	c.AddTask(&File{
		Path:     filepath.Join(dir, name+".crt"),
		Contents: certResource,
		Type:     FileType_File,
		Mode:     fi.String("0644"),
		Owner:    owner,
	})

	c.AddTask(&File{
		Path:     filepath.Join(dir, name+".key"),
		Contents: keyResource,
		Type:     FileType_File,
		Mode:     fi.String("0600"),
		Owner:    owner,
	})

	if caName != "" {
		err = c.EnsureTask(&File{
			Path:     filepath.Join(dir, caName+".crt"),
			Contents: caResource,
			Type:     FileType_File,
			Mode:     fi.String("0644"),
			Owner:    owner,
		})
		if err != nil {
			return nil
		}
	}

	return nil
}

func (e *IssueCert) Run(c *fi.Context) error {
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

	klog.Infof("signing certificate for %q", e.Name)
	certificate, privateKey, caCertificate, err := pki.IssueCert(req, c.Keystore)
	if err != nil {
		return err
	}

	certResource, keyResource, caResource := e.GetResources()
	certResource.Resource = &asBytesResource{certificate}
	keyResource.Resource = &asBytesResource{privateKey}
	caResource.Resource = &asBytesResource{caCertificate}

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
