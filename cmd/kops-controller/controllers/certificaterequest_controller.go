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

package controllers

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/jetstack/cert-manager/pkg/api/util"
	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type CertificateRequestReconciler struct {
	client   client.Client
	log      logr.Logger
	keystore pki.Keystore
}

func NewCertificateRequestReconciler(mgr manager.Manager, keystore pki.Keystore) (*CertificateRequestReconciler, error) {
	r := &CertificateRequestReconciler{
		client:   mgr.GetClient(),
		log:      ctrl.Log.WithName("controllers").WithName("CertificateRequest"),
		keystore: keystore,
	}

	return r, nil
}

func (r *CertificateRequestReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	_ = r.log.WithValues("certificaterequestcontroller", req.NamespacedName)

	cr := &cmapi.CertificateRequest{}
	if err := r.client.Get(ctx, req.NamespacedName, cr); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	r.log.V(8).Info("reconciling ", cr.ObjectMeta.Name)

	if cr.Spec.IssuerRef.Group != "kops.k8s.io" {
		r.log.V(8).Info("resource is now owned by kops ", cr.Spec.IssuerRef.Group)
		return ctrl.Result{}, nil
	}

	if len(cr.Status.Certificate) > 0 {
		r.log.V(4).Info("existing certificate data found in status, skipping already completed CertificateRequest")
		return ctrl.Result{}, nil
	}

	// Kops doesn't sign CAs
	if cr.Spec.IsCA {
		return ctrl.Result{}, nil
	}

	err := signCSR(cr, r.keystore)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, r.setStatus(ctx, cr, cmmeta.ConditionTrue, cmapi.CertificateRequestReasonIssued, "Certificate issued")
}

func signCSR(cr *cmapi.CertificateRequest, keystore pki.Keystore) error {

	csrBytes := cr.Spec.Request

	block, _ := pem.Decode(csrBytes)

	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return fmt.Errorf("could not parse pem block: %v", err)
	}

	client := false
	server := false
	for _, u := range cr.Spec.Usages {
		if u == cmapi.UsageServerAuth {
			server = true
		}
		if u == cmapi.UsageClientAuth {
			client = true
		}
	}

	certType := ""

	switch true {
	case client && server:
		certType = "clientServer"
	case client:
		certType = "client"
	case server:
		certType = "server"
	}

	issueReq := &pki.IssueCertRequest{
		Signer:         fi.CertificateIDCA,
		Type:           certType,
		AlternateNames: csr.DNSNames,
		PublicKey:      csr.PublicKey,
		Subject:        csr.Subject,
	}

	signedCert, _, caCertificate, err := pki.IssueCert(issueReq, keystore)
	if err != nil {
		return fmt.Errorf("failed to issue cert: %v", err)
	}

	signedBytes, err := signedCert.AsBytes()
	if err != nil {
		return fmt.Errorf("failed to encode signed cert: %v", err)
	}

	caBytes, err := caCertificate.AsBytes()
	if err != nil {
		return fmt.Errorf("failed to encode caCert: %v", err)
	}

	cr.Status.Certificate = signedBytes
	cr.Status.CA = caBytes
	return nil

}

func (r *CertificateRequestReconciler) setStatus(ctx context.Context, cr *cmapi.CertificateRequest, status cmmeta.ConditionStatus, reason, message string, args ...interface{}) error {
	completeMessage := fmt.Sprintf(message, args...)
	util.SetCertificateRequestCondition(cr, cmapi.CertificateRequestConditionReady, status, reason, completeMessage)

	return r.client.Status().Update(ctx, cr)
}

func (r *CertificateRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cmapi.CertificateRequest{}).
		Complete(r)
}
