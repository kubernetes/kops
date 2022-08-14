/*
Copyright 2019 The Kubernetes Authors.

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

package channels

import (
	"context"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"net/url"

	"go.uber.org/multierr"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/util/pkg/vfs"

	certmanager "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/kops/channels/pkg/api"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Applier interface {
	Apply(ctx context.Context, data []byte) error
}

// Addon is a wrapper around a single version of an addon
type Addon struct {
	Name            string
	ChannelName     string
	ChannelLocation url.URL
	Spec            *api.AddonSpec
}

// AddonUpdate holds data about a proposed update to an addon
type AddonUpdate struct {
	Name            string
	ExistingVersion *ChannelVersion
	NewVersion      *ChannelVersion
	InstallPKI      bool
}

// AddonMenu is a collection of addons, with helpers for computing the latest versions
type AddonMenu struct {
	Addons map[string]*Addon
}

func NewAddonMenu() *AddonMenu {
	return &AddonMenu{
		Addons: make(map[string]*Addon),
	}
}

func (m *AddonMenu) MergeAddons(o *AddonMenu) {
	for k, v := range o.Addons {
		existing := m.Addons[k]
		if existing == nil {
			m.Addons[k] = v
		} else {
			if v.ChannelVersion().replaces(k, existing.ChannelVersion()) {
				m.Addons[k] = v
			}
		}
	}
}

func (a *Addon) ChannelVersion() *ChannelVersion {
	return &ChannelVersion{
		Channel:          &a.ChannelName,
		Id:               a.Spec.Id,
		ManifestHash:     a.Spec.ManifestHash,
		SystemGeneration: CurrentSystemGeneration,
	}
}

func (a *Addon) buildChannel() *Channel {
	channel := &Channel{
		Namespace: a.GetNamespace(),
		Name:      a.Name,
	}
	return channel
}

func (a *Addon) GetNamespace() string {
	namespace := "kube-system"
	if a.Spec.Namespace != nil {
		namespace = *a.Spec.Namespace
	}
	return namespace
}

func (a *Addon) GetRequiredUpdates(ctx context.Context, k8sClient kubernetes.Interface, cmClient certmanager.Interface, existingVersion *ChannelVersion) (*AddonUpdate, error) {
	newVersion := a.ChannelVersion()

	channel := a.buildChannel()

	pkiInstalled := true

	if a.Spec.NeedsPKI {
		needsPKI, err := channel.IsPKIInstalled(ctx, k8sClient, cmClient)
		if err != nil {
			return nil, err
		}
		pkiInstalled = needsPKI
	}

	if existingVersion != nil && !newVersion.replaces(a.Name, existingVersion) {
		newVersion = nil
	}

	if pkiInstalled && newVersion == nil {
		return nil, nil
	}

	return &AddonUpdate{
		Name:            a.Name,
		ExistingVersion: existingVersion,
		NewVersion:      newVersion,
		InstallPKI:      !pkiInstalled,
	}, nil
}

func (a *Addon) GetManifestFullUrl() (*url.URL, error) {
	if a.Spec.Manifest == nil || *a.Spec.Manifest == "" {
		return nil, field.Required(field.NewPath("spec", "manifest"), "")
	}

	manifest := *a.Spec.Manifest
	manifestURL, err := url.Parse(manifest)
	if err != nil {
		return nil, field.Invalid(field.NewPath("spec", "manifest"), manifest, "Not a valid URL")
	}
	if !manifestURL.IsAbs() {
		manifestURL = a.ChannelLocation.ResolveReference(manifestURL)
	}
	return manifestURL, nil
}

func (a *Addon) EnsureUpdated(ctx context.Context, k8sClient kubernetes.Interface, cmClient certmanager.Interface, pruner *Pruner, applier Applier, existingVersion *ChannelVersion) (*AddonUpdate, error) {
	required, err := a.GetRequiredUpdates(ctx, k8sClient, cmClient, existingVersion)
	if err != nil {
		return nil, err
	}
	if required == nil {
		return nil, nil
	}

	var merr error

	if required.NewVersion != nil {
		err := a.updateAddon(ctx, k8sClient, pruner, applier, required)
		if err != nil {
			merr = multierr.Append(merr, err)
		}
	}
	if required.InstallPKI {
		err := a.installPKI(ctx, k8sClient, cmClient)
		if err != nil {
			merr = multierr.Append(merr, err)
		}
	}
	return required, merr
}

func (a *Addon) updateAddon(ctx context.Context, k8sClient kubernetes.Interface, pruner *Pruner, applier Applier, required *AddonUpdate) error {
	manifestURL, err := a.GetManifestFullUrl()
	if err != nil {
		return err
	}

	klog.Infof("Applying update from %q", manifestURL)

	// We copy the manifest to a temp file because it is likely e.g. an s3 URL, which kubectl can't read
	data, err := vfs.Context.ReadFile(manifestURL.String())
	if err != nil {
		return fmt.Errorf("error reading manifest: %w", err)
	}

	if err := applier.Apply(ctx, data); err != nil {
		return fmt.Errorf("error applying update from %q: %w", manifestURL, err)
	}

	if err := pruner.Prune(ctx, data, a.Spec.Prune); err != nil {
		return fmt.Errorf("error pruning manifest from %q: %w", manifestURL, err)
	}

	if err := a.AddNeedsUpdateLabel(ctx, k8sClient, required); err != nil {
		return fmt.Errorf("error adding needs-update label: %v", err)
	}

	channel := a.buildChannel()
	err = channel.SetInstalledVersion(ctx, k8sClient, a.ChannelVersion())
	if err != nil {
		return fmt.Errorf("error applying annotation to record addon installation: %v", err)
	}
	return nil
}

func (a *Addon) AddNeedsUpdateLabel(ctx context.Context, k8sClient kubernetes.Interface, required *AddonUpdate) error {
	if required.ExistingVersion != nil {
		if a.Spec.NeedsRollingUpdate != "" {
			err := a.patchNeedsUpdateLabel(ctx, k8sClient)
			if err != nil {
				return fmt.Errorf("error patching needs-update label: %v", err)
			}
		}
	}
	return nil
}

func (a *Addon) patchNeedsUpdateLabel(ctx context.Context, k8sClient kubernetes.Interface) error {
	klog.Infof("addon %v wants to update %v nodes", a.Name, a.Spec.NeedsRollingUpdate)
	selector := ""
	switch a.Spec.NeedsRollingUpdate {
	case "control-plane":
		selector = "node-role.kubernetes.io/master="
	case "worker":
		selector = "node-role.kubernetes.io/node="
	}

	annotationPatch := &annotationPatch{Metadata: annotationPatchMetadata{Annotations: map[string]string{
		"kops.k8s.io/needs-update": "",
	}}}
	annotationPatchJSON, err := json.Marshal(annotationPatch)
	if err != nil {
		return err
	}

	nodeInterface := k8sClient.CoreV1().Nodes()
	nodes, err := nodeInterface.List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return err
	}
	for _, node := range nodes.Items {
		_, err = nodeInterface.Patch(ctx, node.Name, types.StrategicMergePatchType, annotationPatchJSON, metav1.PatchOptions{})

		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Addon) installPKI(ctx context.Context, k8sClient kubernetes.Interface, cmClient certmanager.Interface) error {
	klog.Infof("installing PKI for %q", a.Name)
	req := &pki.IssueCertRequest{
		Type: "ca",
		Subject: pkix.Name{
			CommonName: a.Name,
		},
		AlternateNames: []string{
			a.Name,
		},
	}
	cert, privateKey, _, err := pki.IssueCert(req, nil)
	if err != nil {
		return err
	}

	secretName := a.Name + "-ca"

	certString, err := cert.AsString()
	if err != nil {
		return err
	}
	keyString, err := privateKey.AsString()
	if err != nil {
		return err
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: "kube-system",
		},
		StringData: map[string]string{
			"tls.crt": certString,
			"tls.key": keyString,
		},
		Type: "kubernetes.io/tls",
	}
	_, err = k8sClient.CoreV1().Secrets("kube-system").Create(ctx, secret, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	issuer := &cmv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.Name,
			Namespace: "kube-system",
		},
		Spec: cmv1.IssuerSpec{
			IssuerConfig: cmv1.IssuerConfig{
				CA: &cmv1.CAIssuer{
					SecretName: secretName,
				},
			},
		},
	}

	_, err = cmClient.CertmanagerV1().Issuers("kube-system").Create(ctx, issuer, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}
