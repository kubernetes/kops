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

package v1beta1_test

import (
	. "github.com/kubernetes-incubator/apiserver-builder/example/pkg/apis/miskatonic/v1beta1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Student", func() {
	var instance Student
	var expected Student

	BeforeEach(func() {
		instance = Student{}
		instance.Name = "joe"
		instance.Spec.ID = 3

		expected = instance
	})

	Describe("when sending a storage request", func() {
		It("should return the instance with an incremented the ID", func() {
			client := cs.MiskatonicV1beta1Client.Students("test-create-delete-students")
			actual, err := client.Create(&instance)
			Expect(err).NotTo(HaveOccurred())
			expected.Spec.ID = instance.Spec.ID + 1
			Expect(actual.Spec).To(Equal(expected.Spec))
		})
	})
})
