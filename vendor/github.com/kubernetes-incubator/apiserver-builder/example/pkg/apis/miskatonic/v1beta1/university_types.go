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

package v1beta1

import (
	"log"

	"github.com/kubernetes-incubator/apiserver-builder/example/pkg/apis/miskatonic"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/endpoints/request"
)

// Generating code from university_types.go file will generate storage and status REST endpoints for
// University.

// +genclient=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +k8s:openapi-gen=true
// +resource:path=universities,strategy=UniversityStrategy
// +subresource:request=Scale,path=scale,rest=ScaleUniversityREST
type University struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UniversitySpec   `json:"spec,omitempty"`
	Status UniversityStatus `json:"status,omitempty"`
}

// UniversitySpec defines the desired state of University
type UniversitySpec struct {
	// faculty_size defines the desired faculty size of the university.  Defaults to 15.
	FacultySize int `json:"faculty_size,omitempty"`

	// max_students defines the maximum number of enrolled students.  Defaults to 300.
	// +optional
	MaxStudents *int `json:"max_students,omitempty"`

	// The unversioned struct definition for this field must be manually defined in the group package
	Manual ManualCreateUnversionedType

	// The unversioned struct definition for this field is automatically generated in the group package
	Automatic AutomaticCreateUnversionedType

	Template *corev1.PodSpec `json:"template,omitempty"`

	ServiceSpec corev1.ServiceSpec `json:"service_spec,omitempty"`

	Rollout []extensionsv1beta1.Deployment `json:"rollout,omitempty"`
}

// Require that the unversioned struct is manually created.  This is *NOT* the default behavior for
// structs appearing as fields in a resource that are defined in the same package as that resource,
// but is explicitly configured through the +genregister comment.
// +genregister:unversioned=false
type ManualCreateUnversionedType struct {
	A string
	B bool
}

// Automatically create an unversioned copy of this struct by copying its definition
// This is the default behavior for structs appearing as fields in a resource and that are defined in the
// same package as that resource.
type AutomaticCreateUnversionedType struct {
	A string
	B bool
}

// UniversityStatus defines the observed state of University
type UniversityStatus struct {
	// enrolled_students is the number of currently enrolled students
	EnrolledStudents []string `json:"enrolled_students,omitempty"`

	// statusfield provides status information about University
	FacultyEmployed []string `json:"faculty_employed,omitempty"`
}

// Resource Validation
func (UniversityStrategy) Validate(ctx request.Context, obj runtime.Object) field.ErrorList {
	university := obj.(*miskatonic.University)
	log.Printf("Validating University %s\n", university.Name)
	errors := field.ErrorList{}
	if university.Spec.MaxStudents == nil || *university.Spec.MaxStudents < 1 || *university.Spec.MaxStudents > 150 {
		errors = append(errors, field.Invalid(
			field.NewPath("spec", "MaxStudents"),
			*university.Spec.MaxStudents,
			"Must be between 1 and 150"))
	}
	return errors
}

// GetDefaultingFunctions returns functions for defaulting v1beta1.University values
func (UniversitySchemeFns) DefaultingFunction(o interface{}) {
	obj := o.(*University)
	log.Printf("Defaulting University %s\n", obj.Name)
	if obj.Spec.MaxStudents == nil {
		n := 15
		obj.Spec.MaxStudents = &n
	}
}

// GetConversionFunctions returns functions for converting resource versions to override the
// conversion functions
func (UniversitySchemeFns) GetConversionFunctions() []interface{} {
	return []interface{}{}
}
