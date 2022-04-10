/*
Copyright 2020 The cert-manager Authors.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	cmacme "github.com/cert-manager/cert-manager/pkg/apis/acme/v1"
)

type GenericIssuer interface {
	runtime.Object
	metav1.Object

	GetObjectMeta() *metav1.ObjectMeta
	GetSpec() *IssuerSpec
	GetStatus() *IssuerStatus
}

var _ GenericIssuer = &Issuer{}
var _ GenericIssuer = &ClusterIssuer{}

func (c *ClusterIssuer) GetObjectMeta() *metav1.ObjectMeta {
	return &c.ObjectMeta
}
func (c *ClusterIssuer) GetSpec() *IssuerSpec {
	return &c.Spec
}
func (c *ClusterIssuer) GetStatus() *IssuerStatus {
	return &c.Status
}
func (c *ClusterIssuer) SetSpec(spec IssuerSpec) {
	c.Spec = spec
}
func (c *ClusterIssuer) SetStatus(status IssuerStatus) {
	c.Status = status
}
func (c *ClusterIssuer) Copy() GenericIssuer {
	return c.DeepCopy()
}
func (c *Issuer) GetObjectMeta() *metav1.ObjectMeta {
	return &c.ObjectMeta
}
func (c *Issuer) GetSpec() *IssuerSpec {
	return &c.Spec
}
func (c *Issuer) GetStatus() *IssuerStatus {
	return &c.Status
}
func (c *Issuer) SetSpec(spec IssuerSpec) {
	c.Spec = spec
}
func (c *Issuer) SetStatus(status IssuerStatus) {
	c.Status = status
}
func (c *Issuer) Copy() GenericIssuer {
	return c.DeepCopy()
}

// TODO: refactor these functions away
func (i *IssuerStatus) ACMEStatus() *cmacme.ACMEIssuerStatus {
	// this is an edge case, but this will prevent panics
	if i == nil {
		return &cmacme.ACMEIssuerStatus{}
	}
	if i.ACME == nil {
		i.ACME = &cmacme.ACMEIssuerStatus{}
	}
	return i.ACME
}
