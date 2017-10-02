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
	. "github.com/kubernetes-incubator/apiserver-builder/example/pkg/client/clientset_generated/clientset/typed/miskatonic/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("University", func() {
	var instance University
	var expected University
	var client UniversityInterface

	BeforeEach(func() {
		instance = University{}
		instance.Name = "miskatonic-university"
		instance.Spec.FacultySize = 7
		instance.Spec.ServiceSpec = corev1.ServiceSpec{}
		instance.Spec.ServiceSpec.ClusterIP = "1.1.1.1"

		expected = instance
		val := 15
		expected.Spec.MaxStudents = &val
		expected.Spec.ServiceSpec = corev1.ServiceSpec{}
		expected.Spec.ServiceSpec.ClusterIP = "1.1.1.1"
	})

	AfterEach(func() {
		client.Delete(instance.Name, &metav1.DeleteOptions{})
	})

	Describe("when sending a storage request", func() {
		Context("for a valid config", func() {
			It("should provide CRUD access to the object", func() {
				client = cs.MiskatonicV1beta1Client.Universities("university-test-valid")

				By("returning success from the create request")
				actual, err := client.Create(&instance)
				Expect(err).ShouldNot(HaveOccurred())

				By("defaulting the expected fields")
				Expect(actual.Spec).To(Equal(expected.Spec))

				By("returning the item for list requests")
				result, err := client.List(metav1.ListOptions{})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(result.Items).To(HaveLen(1))
				Expect(result.Items[0].Spec).To(Equal(expected.Spec))

				By("returning the item for get requests")
				actual, err = client.Get(instance.Name, metav1.GetOptions{})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(actual.Spec).To(Equal(expected.Spec))

				By("deleting the item for delete requests")
				err = client.Delete(instance.Name, &metav1.DeleteOptions{})
				Expect(err).ShouldNot(HaveOccurred())
				result, err = client.List(metav1.ListOptions{})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(result.Items).To(HaveLen(0))
			})
		})
		Context("for an invalid config", func() {
			It("should fail if there are too many students", func() {
				client = cs.MiskatonicV1beta1Client.Universities("university-test-too-many")
				val := 151
				instance.Spec.MaxStudents = &val
				_, err := client.Create(&instance)
				Expect(err).Should(HaveOccurred())
			})

			It("should fail if there are not enough students", func() {
				client = cs.MiskatonicV1beta1Client.Universities("university-test-not-enough")
				val := 0
				instance.Spec.MaxStudents = &val
				_, err := client.Create(&instance)
				Expect(err).Should(HaveOccurred())
			})
		})
	})

	Describe("when sending a scale request", func() {
		It("should set the faculty count", func() {
			client = cs.MiskatonicV1beta1Client.Universities("university-test-scale")
			_, err := client.Create(&instance)
			Expect(err).ShouldNot(HaveOccurred())

			scale := &Scale{
				Faculty: 30,
			}
			scale.Name = instance.Name
			restClient := cs.MiskatonicV1beta1Client.RESTClient()
			err = restClient.Post().Namespace("university-test-scale").
				Name(instance.Name).
				Resource("universities").
				SubResource("scale").
				Body(scale).Do().Error()
			Expect(err).ShouldNot(HaveOccurred())

			expected.Spec.FacultySize = 30
			actual, err := client.Get(instance.Name, metav1.GetOptions{})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(actual.Spec).Should(Equal(expected.Spec))
		})
	})
})
