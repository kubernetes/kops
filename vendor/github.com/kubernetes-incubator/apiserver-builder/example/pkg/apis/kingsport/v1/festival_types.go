/*
Copyright 2017 The Kubernetes Authors.

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
	"log"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/endpoints/request"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/kubernetes-incubator/apiserver-builder/example/pkg/apis/kingsport"
)

// +genclient=true
// +nonNamespaced=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Festival
// +k8s:openapi-gen=true
// +resource:path=festivals,strategy=FestivalStrategy
type Festival struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FestivalSpec   `json:"spec,omitempty"`
	Status FestivalStatus `json:"status,omitempty"`
}

// FestivalSpec defines the desired state of Festival
type FestivalSpec struct {
	// Year when the festival was held, may be negative (BC)
	Year int `json:"year,omitempty"`
	// Invited holds the number of invited attendees
	Invited uint `json:"invited,omitempty"`
}

// FestivalStatus defines the observed state of Festival
type FestivalStatus struct {
	// Attended holds the actual number of attendees
	Attended uint `json:"attended,omitempty"`
}

// Validate checks that an instance of Festival is well formed
func (FestivalStrategy) Validate(ctx request.Context, obj runtime.Object) field.ErrorList {
	o := obj.(*kingsport.Festival)
	log.Printf("Validating fields for Festival %s\n", o.Name)
	errors := field.ErrorList{}
	// perform validation here and add to errors using field.Invalid
	return errors
}

func (FestivalStrategy) NamespaceScoped() bool { return false }

func (FestivalStatusStrategy) NamespaceScoped() bool { return false }

// DefaultingFunction sets default Festival field values
func (FestivalSchemeFns) DefaultingFunction(o interface{}) {
	obj := o.(*Festival)
	// set default field values here
	log.Printf("Defaulting fields for Festival %s\n", obj.Name)
}
