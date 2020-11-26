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

package api

import (
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kscheme "k8s.io/client-go/kubernetes/scheme"
	apireg "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1beta1"

	whapi "github.com/jetstack/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	cmacmev1 "github.com/jetstack/cert-manager/pkg/apis/acme/v1"
	cmacmev1alpha2 "github.com/jetstack/cert-manager/pkg/apis/acme/v1alpha2"
	cmacmev1alpha3 "github.com/jetstack/cert-manager/pkg/apis/acme/v1alpha3"
	cmacmev1beta1 "github.com/jetstack/cert-manager/pkg/apis/acme/v1beta1"
	cmapiv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	cmapiv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	cmapiv1alpha3 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha3"
	cmapiv1beta1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1beta1"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
)

// This package defines a Scheme and Codec that has the *external* API types
// registered.
// This means that the scheme will *not* perform defaulting or conversions for
// cert-manager API resources.
// This is to ensure a clean separation between API semantics and controllers.
// Only the webhook should utilise a scheme with conversions and defaults
// registered in order to ensure all controllers have a consistent view of
// resource types in the apiserver.

var Scheme = runtime.NewScheme()
var Codecs = serializer.NewCodecFactory(Scheme)
var ParameterCodec = runtime.NewParameterCodec(Scheme)
var localSchemeBuilder = runtime.SchemeBuilder{
	cmapiv1alpha2.AddToScheme,
	cmapiv1alpha3.AddToScheme,
	cmapiv1beta1.AddToScheme,
	cmapiv1.AddToScheme,
	cmacmev1alpha2.AddToScheme,
	cmacmev1alpha3.AddToScheme,
	cmacmev1beta1.AddToScheme,
	cmacmev1.AddToScheme,
	cmmeta.AddToScheme,
	whapi.AddToScheme,
	kscheme.AddToScheme,
	apireg.AddToScheme,
	apiext.AddToScheme,
}

// AddToScheme adds all types of this clientset into the given scheme. This allows composition
// of clientsets, like in:
//
//   import (
//     "k8s.io/client-go/kubernetes"
//     clientsetscheme "k8s.io/client-go/kubernetes/scheme"
//     aggregatorclientsetscheme "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset/scheme"
//   )
//
//   kclientset, _ := kubernetes.NewForConfig(c)
//   _ = aggregatorclientsetscheme.AddToScheme(clientsetscheme.Scheme)
//
// After this, RawExtensions in Kubernetes types will serialize kube-aggregator types
// correctly.
var AddToScheme = localSchemeBuilder.AddToScheme

func init() {
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})
	utilruntime.Must(AddToScheme(Scheme))
}
