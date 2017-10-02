/*
Copyright 2016 The Kubernetes Authors.

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

// Package install installs the kops API group, making it available as
// an option to all of the API encoding/decoding machinery.
package install

import (
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/apimachinery/announced"
	"k8s.io/apimachinery/pkg/apimachinery/registered"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/v1alpha1"
	"k8s.io/kops/pkg/apis/kops/v1alpha2"
)

// Install registers the API group and adds types to a scheme
func Install(groupFactoryRegistry announced.APIGroupFactoryRegistry, registry *registered.APIRegistrationManager, scheme *runtime.Scheme) {
	err := announced.NewGroupMetaFactory(
		&announced.GroupMetaFactoryArgs{
			GroupName: kops.GroupName,
			VersionPreferenceOrder: []string{
				v1alpha2.SchemeGroupVersion.Version,
				v1alpha1.SchemeGroupVersion.Version,
			},
			// RootScopedKinds are resources that are not namespaced.
			RootScopedKinds: sets.NewString(),
			//ImportPrefix:               "k8s.io/kops/pkg/apis/kops",
			AddInternalObjectsToScheme: kops.AddToScheme,
		},
		announced.VersionToSchemeFunc{
			v1alpha1.SchemeGroupVersion.Version: v1alpha1.AddToScheme,
			v1alpha2.SchemeGroupVersion.Version: v1alpha2.AddToScheme,
		},
	).Announce(groupFactoryRegistry).RegisterAndEnable(registry, scheme)
	if err != nil {
		glog.Fatalf("error registering kops schema: %v", err)
	}
}
