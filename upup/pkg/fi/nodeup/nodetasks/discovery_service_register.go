/*
Copyright 2025 The Kubernetes Authors.

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
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	discoveryapi "k8s.io/kops/discovery/apis/discovery.kops.k8s.io/v1alpha1"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
)

// DiscoveryServiceRegisterTask is responsible for registering with the discovery service.
type DiscoveryServiceRegisterTask struct {
	// Name is a reference for our task
	Name string

	// DiscoveryService is the discovery service to register with (including the universe ID prefix)
	DiscoveryService string

	// RegisterNamespace is the namespace to use for registration with the discovery service
	RegisterNamespace string

	// RegisterName is the name to use for registration with the discovery service
	RegisterName string

	// ClientCert is the client certificate to present when registering
	ClientCert fi.Resource

	// ClientKey is the client key to use when registering
	ClientKey fi.Resource

	// ClientCA is the CA certificate to use when registering,
	// we include it in the bundle presented to the server,
	// as it is likely self-signed.
	ClientCA fi.Resource

	// JWKS is the set of public keys to advertise through the discovery service.
	JWKS []JSONWebKey
}

// JSONWebKey wraps discoveryapi.JSONWebKey, to implement dependency discovery.
type JSONWebKey struct {
	discoveryapi.JSONWebKey
}

var _ fi.NodeupHasDependencies = (*JSONWebKey)(nil)

// GetDependencies returns the dependencies for the JSONWebKey; there are none.
func (j *JSONWebKey) GetDependencies(tasks map[string]fi.NodeupTask) []fi.NodeupTask {
	return nil
}

var _ fi.NodeupTask = (*UpdateEtcHostsTask)(nil)

func (e *DiscoveryServiceRegisterTask) String() string {
	return fmt.Sprintf("DiscoveryServiceRegisterTask: %s", e.Name)
}

var _ fi.HasName = (*DiscoveryServiceRegisterTask)(nil)

func (f *DiscoveryServiceRegisterTask) GetName() *string {
	return &f.Name
}

func (e *DiscoveryServiceRegisterTask) Find(c *fi.NodeupContext) (*DiscoveryServiceRegisterTask, error) {
	// We always register with the service.
	return nil, nil
}

func (e *DiscoveryServiceRegisterTask) Run(c *fi.NodeupContext) error {
	return fi.NodeupDefaultDeltaRunMethod(e, c)
}

func (_ *DiscoveryServiceRegisterTask) CheckChanges(a, e, changes *DiscoveryServiceRegisterTask) error {
	return nil
}

func (_ *DiscoveryServiceRegisterTask) RenderLocal(t *local.LocalTarget, a, e, changes *DiscoveryServiceRegisterTask) error {
	ctx := context.TODO()

	log := klog.FromContext(ctx)

	clientCert, err := fi.ResourceAsBytes(e.ClientCert)
	if err != nil {
		return err
	}
	clientKey, err := fi.ResourceAsBytes(e.ClientKey)
	if err != nil {
		return err
	}
	clientCA, err := fi.ResourceAsBytes(e.ClientCA)
	if err != nil {
		return err
	}

	clientCertBundle := []byte{}
	clientCertBundle = append(clientCertBundle, clientCert...)
	clientCertBundle = append(clientCertBundle, clientCA...)

	config := &rest.Config{
		Host: e.DiscoveryService,
		TLSClientConfig: rest.TLSClientConfig{
			CertData: clientCertBundle,
			KeyData:  clientKey,
		},
	}
	kubeClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("error creating dynamic client: %w", err)
	}

	spec := discoveryapi.DiscoveryEndpointSpec{}

	spec.OIDC = &discoveryapi.OIDCSpec{}

	for _, jwk := range e.JWKS {
		spec.OIDC.Keys = append(spec.OIDC.Keys, jwk.JSONWebKey)
	}

	ep := &discoveryapi.DiscoveryEndpoint{
		Spec: spec,
	}

	ep.Kind = "DiscoveryEndpoint"
	ep.APIVersion = "discovery.kops.k8s.io/v1alpha1"

	ep.Name = e.RegisterName
	ep.Namespace = e.RegisterNamespace

	// Convert to Unstructured
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(ep)
	if err != nil {
		return fmt.Errorf("failed to convert to unstructured: %w", err)
	}
	u := &unstructured.Unstructured{Object: obj}
	gvr := discoveryapi.DiscoveryEndpointGVR

	// Use Server-Side Apply to Create/Update
	created, err := kubeClient.Resource(gvr).Namespace(u.GetNamespace()).Apply(ctx, u.GetName(), u, metav1.ApplyOptions{FieldManager: "nodeup-register"})
	if err != nil {
		return fmt.Errorf("failed to register with discovery service: %w", err)
	}

	var result discoveryapi.DiscoveryEndpoint
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(created.Object, &result); err != nil {
		return fmt.Errorf("failed to convert from unstructured: %w", err)
	}
	log.Info("registered with discovery service", "result", result)

	return nil
}
