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

/*
The addon package contains tools opinionated for managing Kubernetes cluster addons.
The declarative.DeclarativeObject must be castable to a addonsv1alpha1.CommonObject
in order for this pattern to be used.

What is an Addon Object?

An Addon Object is an instance of a type defined as a CustomResourceDefinition that
implements the addonsv1alpha1.CommonObject interface. The object represents the intent
to deploy an instance of a specific Addon in the cluster. This pattern manages a
Kubernetes deployment for the specific addon based on the Addon Object.

Writing an Addon Operator

Follow the dashboard walkthrough to stand up an addon operator[1]. Then dig into the
declarative and addon patterns to extend it.

[1] https://github.com/kubernetes-sigs/kubebuilder-declarative-pattern/tree/master/docs/addon/walkthrough
*/
package addon
