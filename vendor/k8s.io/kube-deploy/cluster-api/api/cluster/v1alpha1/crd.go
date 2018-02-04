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

package v1alpha1

import (
	"fmt"
	"reflect"
	"time"

	extensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	CRDGroup   = "cluster.k8s.io"
	CRDVersion = "v1alpha1"

	ClustersCRDPlural = "clusters"
	ClustersCRDName   = ClustersCRDPlural + "." + CRDGroup
	MachinesCRDPlural = "machines"
	MachinesCRDName   = MachinesCRDPlural + "." + CRDGroup
)

var SchemeGroupVersion = schema.GroupVersion{Group: CRDGroup, Version: CRDVersion}

func CreateClustersCRD(clientset apiextensionsclient.Interface) (*extensionsv1.CustomResourceDefinition, error) {
	crd := &extensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: ClustersCRDName,
		},
		Spec: extensionsv1.CustomResourceDefinitionSpec{
			Group:   SchemeGroupVersion.Group,
			Version: SchemeGroupVersion.Version,
			Scope:   extensionsv1.ClusterScoped,
			Names: extensionsv1.CustomResourceDefinitionNames{
				Plural: ClustersCRDPlural,
				Kind:   reflect.TypeOf(Cluster{}).Name(),
			},
		},
	}
	return crd, createCRD(clientset, crd)
}

func CreateMachinesCRD(clientset apiextensionsclient.Interface) (*extensionsv1.CustomResourceDefinition, error) {
	crd := &extensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: MachinesCRDName,
		},
		Spec: extensionsv1.CustomResourceDefinitionSpec{
			Group:   SchemeGroupVersion.Group,
			Version: SchemeGroupVersion.Version,
			Scope:   extensionsv1.ClusterScoped,
			Names: extensionsv1.CustomResourceDefinitionNames{
				Plural: MachinesCRDPlural,
				Kind:   reflect.TypeOf(Machine{}).Name(),
			},
		},
	}
	return crd, createCRD(clientset, crd)
}

func createCRD(clientset apiextensionsclient.Interface, crd *extensionsv1.CustomResourceDefinition) error {
	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err != nil {
		return err
	}

	// wait for CRD being established
	err = wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		crd, err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crd.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case extensionsv1.Established:
				if cond.Status == extensionsv1.ConditionTrue {
					return true, err
				}
			case extensionsv1.NamesAccepted:
				if cond.Status == extensionsv1.ConditionFalse {
					fmt.Printf("Name conflict: %v\n", cond.Reason)
				}
			}
		}
		return false, err
	})
	if err != nil {
		deleteErr := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(crd.ObjectMeta.Name, nil)
		if deleteErr != nil {
			return errors.NewAggregate([]error{err, deleteErr})
		}
		return err
	}
	return nil
}
