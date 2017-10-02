/*
Copyright YEAR The Kubernetes Authors.

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

package sharedinformers

import "log"

// SetupKubernetesTypes registers the config for watching Kubernetes types
func (si *SharedInformers) SetupKubernetesTypes() bool {
	return true
}

// StartAdditionalInformers starts watching Deployments
func (si *SharedInformers) StartAdditionalInformers(shutdown <-chan struct{}) {
	log.Printf("Listen for Deployments")
	go si.KubernetesFactory.Extensions().V1beta1().Deployments().Informer().Run(shutdown)
}
