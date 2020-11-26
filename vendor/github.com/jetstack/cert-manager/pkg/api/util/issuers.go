/*
Copyright 2019 The Jetstack cert-manager contributors.

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

package util

import (
	"fmt"

	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
)

const (
	// IssuerACME is the name of the ACME issuer
	IssuerACME string = "acme"
	// IssuerCA is the name of the simple issuer
	IssuerCA string = "ca"
	// IssuerVault is the name of the Vault issuer
	IssuerVault string = "vault"
	// IssuerSelfSigned is a self signing issuer
	IssuerSelfSigned string = "selfsigned"
	// IssuerVenafi uses Venafi Trust Protection Platform and Venafi Cloud
	IssuerVenafi string = "venafi"
)

// NameForIssuer determines the name of the Issuer implementation given an
// Issuer resource.
func NameForIssuer(i cmapi.GenericIssuer) (string, error) {
	switch {
	case i.GetSpec().ACME != nil:
		return IssuerACME, nil
	case i.GetSpec().CA != nil:
		return IssuerCA, nil
	case i.GetSpec().Vault != nil:
		return IssuerVault, nil
	case i.GetSpec().SelfSigned != nil:
		return IssuerSelfSigned, nil
	case i.GetSpec().Venafi != nil:
		return IssuerVenafi, nil
	}
	return "", fmt.Errorf("no issuer specified for Issuer '%s/%s'", i.GetObjectMeta().Namespace, i.GetObjectMeta().Name)
}

// issuerKind returns the kind of issuer for a certificate
func IssuerKind(ref cmmeta.ObjectReference) string {
	if ref.Kind == "" {
		return cmapi.IssuerKind
	}
	return ref.Kind
}
