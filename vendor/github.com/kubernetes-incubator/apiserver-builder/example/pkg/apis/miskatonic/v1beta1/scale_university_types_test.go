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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/client-go/pkg/api/v1"

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
		//instance.Spec.ServiceSpec = v1.ServiceSpec{}
		//instance.Spec.ServiceSpec.ClusterIP = "1.1.1.1"

		expected = instance
		val := 15
		expected.Spec.MaxStudents = &val
		//expected.Spec.ServiceSpec = v1.ServiceSpec{}
		//expected.Spec.ServiceSpec.ClusterIP = "1.1.1.1"

	})

	AfterEach(func() {
		client.Delete(instance.Name, &metav1.DeleteOptions{})
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
